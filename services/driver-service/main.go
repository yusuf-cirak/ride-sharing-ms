package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"ride-sharing/shared/env"
	"ride-sharing/shared/messaging"
	"ride-sharing/shared/tracing"
	"syscall"

	"google.golang.org/grpc"
)

var GrpcAddress = ":9092"

func main() {

	// init tracing

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

	rabbitMQURI := env.GetString("RABBITMQ_URI", "amqp://guest:guest@localhost:5672/")

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

	service := NewService()
	// Starting gRPC server
	grpcServer := grpc.NewServer(tracing.WithTracingInterceptors()...)
	NewGrpcHandler(grpcServer, service)

	consumer := NewTripConsumer(rabbitmq, service)
	go func() {
		if err := consumer.Listen(); err != nil {
			log.Fatalf("failed to listen for messages: %v", err)
		}
	}()

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
