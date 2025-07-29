package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"ride-sharing/services/trip-service/internal/infrastructure/events"
	"ride-sharing/services/trip-service/internal/infrastructure/repository"
	"ride-sharing/services/trip-service/internal/service"
	"ride-sharing/shared/env"
	"ride-sharing/shared/messaging"
	"ride-sharing/shared/tracing"
	"syscall"

	grpcHandlers "ride-sharing/services/trip-service/internal/infrastructure/grpc"

	"google.golang.org/grpc"
)

var GrpcAddress = ":9093"

func main() {

	tracerCfg := tracing.Config{
		ServiceName:    "api-gateway",
		Environment:    env.GetString("ENVIRONMENT", "development"),
		JaegerEndpoint: env.GetString("JAEGER_ENDPOINT", "http://localhost:14268/api/traces"),
	}

	tracingShutdown, err := tracing.InitTracer(tracerCfg)

	if err != nil {
		log.Fatalf("Failed to initialize tracer: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer tracingShutdown(ctx)

	// mongoClient, err := db.NewMongoClient(ctx, db.NewMongoDefaultConfig())
	// if err != nil {
	// 	log.Fatalf("Failed to create MongoDB client: %v", err)
	// }
	// defer mongoClient.Disconnect(ctx)

	// mongoDb := db.GetDatabase(mongoClient, db.NewMongoDefaultConfig())

	rabbitMQURI := env.GetString("RABBITMQ_URI", "amqp://guest:guest@localhost:5672/")

	// tripRepo := repository.NewMongoRepository(mongoDb)
	inmemRepo := repository.NewInmemRepository()

	svc := service.NewTripService(inmemRepo)

	go func() {
		signalChan := make(chan os.Signal, 1)
		signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
		<-signalChan
		cancel()
	}()

	lis, err := net.Listen("tcp", GrpcAddress)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// rabbitmq connection
	rabbitmq, err := messaging.NewRabbitMQ(rabbitMQURI)
	if err != nil {
		log.Fatalf("failed to connect to RabbitMQ: %v", err)
	}
	defer rabbitmq.Close()

	publisher := events.NewTripEventPublisher(rabbitmq)

	// Driver consumer
	driverConsumer := events.NewDriverConsumer(rabbitmq, svc)
	go driverConsumer.Listen()

	// Payment consumer
	paymentConsumer := events.NewPaymentConsumer(rabbitmq, svc)
	go paymentConsumer.Listen()

	// Starting gRPC server
	grpcServer := grpc.NewServer(tracing.WithTracingInterceptors()...)

	grpcHandlers.NewGRPCHandler(grpcServer, svc, publisher)

	log.Printf("Trip service is running on %s", lis.Addr().String())

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("Shutting down trip service...")
	grpcServer.GracefulStop()

}
