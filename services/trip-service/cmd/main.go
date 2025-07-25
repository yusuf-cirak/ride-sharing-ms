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
	"syscall"

	grpcHandlers "ride-sharing/services/trip-service/internal/infrastructure/grpc"

	"google.golang.org/grpc"
)

var GrpcAddress = ":9093"

func main() {
	rabbitMQURI := env.GetString("RABBITMQ_URI", "amqp://guest:guest@localhost:5672/")
	inmemRepo := repository.NewInmemRepository()

	svc := service.NewTripService(inmemRepo)

	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()

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

	// Starting gRPC server
	grpcServer := grpc.NewServer()

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
