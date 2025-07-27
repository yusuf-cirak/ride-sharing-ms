package events

import (
	"context"
	"encoding/json"
	"log"
	"ride-sharing/services/trip-service/internal/domain"
	"ride-sharing/shared/contracts"
	"ride-sharing/shared/messaging"

	"github.com/rabbitmq/amqp091-go"
)

type driverConsumer struct {
	rabbitMQ *messaging.RabbitMQ
	service  domain.TripService
}

func NewDriverConsumer(rabbitMQ *messaging.RabbitMQ, service domain.TripService) *driverConsumer {
	return &driverConsumer{
		rabbitMQ: rabbitMQ,
		service:  service,
	}
}

func (c *driverConsumer) Listen() error {
	return c.rabbitMQ.ConsumeMessages(messaging.DriverTripResponseQueue, func(ctx context.Context, msg amqp091.Delivery) error {
		// Handle the incoming message
		var tripEvent contracts.AmqpMessage
		if err := json.Unmarshal(msg.Body, &tripEvent); err != nil {
			log.Printf("failed to unmarshal message: %v", err)
			return err
		}

		var payload messaging.DriverTripResponseData
		if err := json.Unmarshal(tripEvent.Data, &payload); err != nil {
			log.Printf("failed to unmarshal message: %v", err)
			return err
		}

		switch msg.RoutingKey {
		case contracts.DriverCmdTripAccept:
			if err := c.handleTripAccepted(ctx, &payload); err != nil {
				log.Printf("failed to handle trip accepted: %v", err)
				return err
			}
		case contracts.DriverCmdTripDecline:
			log.Printf("Trip declined by driver: %s", payload.Driver.Id)
			return nil
		}

		log.Printf("Unhandled routing key: %s", msg.RoutingKey)
		return nil
	})
}

func (c *driverConsumer) handleTripAccepted(ctx context.Context, payload *messaging.DriverTripResponseData) error {

	trip, err := c.service.GetTripByID(ctx, payload.TripID)
	if err != nil || trip == nil {
		log.Printf("failed to get trip by ID: %v", err)
		return err
	}

	if err := c.service.UpdateTrip(ctx, payload.TripID, "accepted", &payload.Driver); err != nil {
		log.Printf("failed to update trip status: %v", err)
		return err
	}

	trip, err = c.service.GetTripByID(ctx, payload.TripID)
	if err != nil || trip == nil {
		log.Printf("failed to get trip by ID: %v", err)
		return err
	}

	marshalTrip, err := json.Marshal(trip)

	if err != nil {
		log.Printf("failed to marshal trip: %v", err)
		return err
	}

	// notify the reider that driver has been assigned
	if err := c.rabbitMQ.PublishMessage(ctx, contracts.TripEventDriverAssigned, &contracts.AmqpMessage{
		OwnerID: trip.UserID,
		Data:    marshalTrip,
	}); err != nil {
		log.Printf("failed to publish trip response: %v", err)
		return err
	}

	// todo: notify the payment service to start a payment link

	return nil
}
