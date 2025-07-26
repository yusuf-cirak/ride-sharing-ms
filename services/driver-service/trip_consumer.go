package main

import (
	"context"
	"encoding/json"
	"log"
	"ride-sharing/shared/contracts"
	"ride-sharing/shared/messaging"

	"github.com/rabbitmq/amqp091-go"
)

type tripConsumer struct {
	rabbitMQ *messaging.RabbitMQ
	service  *Service
}

func NewTripConsumer(rabbitMQ *messaging.RabbitMQ, service *Service) *tripConsumer {
	return &tripConsumer{
		rabbitMQ: rabbitMQ,
		service:  service,
	}
}

func (c *tripConsumer) Listen() error {
	return c.rabbitMQ.ConsumeMessages(messaging.FindAvailableDriversQueue, func(ctx context.Context, msg amqp091.Delivery) error {
		// Handle the incoming message
		var tripEvent contracts.AmqpMessage
		if err := json.Unmarshal(msg.Body, &tripEvent); err != nil {
			log.Printf("failed to unmarshal message: %v", err)
			return err
		}

		var payload messaging.TripEventData
		if err := json.Unmarshal(tripEvent.Data, &payload); err != nil {
			log.Printf("failed to unmarshal message: %v", err)
			return err
		}

		log.Printf("Driver received message: %+v", payload)

		switch msg.RoutingKey {
		case contracts.TripEventCreated, contracts.TripEventDriverNotInterested:
			return c.handleFindAndNotifyDrivers(ctx, &payload)
		}

		log.Printf("Unhandled routing key: %s", msg.RoutingKey)
		return nil
	})
}

func (c *tripConsumer) handleFindAndNotifyDrivers(ctx context.Context, payload *messaging.TripEventData) error {

	suitableIDs := c.service.FindAvailableDrivers(payload.Trip.RideFare.PackageSlug)

	if suitableIDs == nil || len(suitableIDs) == 0 {
		log.Printf("No suitable drivers found for trip %s", payload.Trip.Id)

		if err := c.rabbitMQ.PublishMessage(ctx, contracts.TripEventNoDriversFound, &contracts.AmqpMessage{OwnerID: payload.Trip.UserID}); err != nil {
			log.Printf("failed to publish message: %v", err)
			return err
		}
		return nil
	}

	suitableDriverID := (suitableIDs)[0]

	marshaledData, err := json.Marshal(suitableDriverID)

	if err != nil {
		log.Printf("failed to marshal driver ID: %v", err)
		return err
	}

	log.Printf("Found suitable driver %s for trip %s", suitableDriverID, payload.Trip.Id)

	if err := c.rabbitMQ.PublishMessage(ctx, contracts.DriverCmdTripRequest, &contracts.AmqpMessage{
		OwnerID: payload.Trip.UserID,
		Data:    marshaledData,
	}); err != nil {
		log.Printf("failed to publish message: %v", err)
		return err
	}

	return nil
}
