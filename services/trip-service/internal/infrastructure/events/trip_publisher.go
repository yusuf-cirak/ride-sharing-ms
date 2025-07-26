package events

import (
	"context"
	"encoding/json"
	"fmt"
	"ride-sharing/services/trip-service/internal/domain"
	"ride-sharing/shared/contracts"
	"ride-sharing/shared/messaging"
)

type TripEventPublisher struct {
	rabbitmq *messaging.RabbitMQ
}

func NewTripEventPublisher(rabbitmq *messaging.RabbitMQ) *TripEventPublisher {
	return &TripEventPublisher{
		rabbitmq: rabbitmq,
	}
}

func (p *TripEventPublisher) PublishTripCreated(ctx context.Context, trip *domain.TripModel) error {

	msg := messaging.TripEventData{
		Trip: trip.ToProto(),
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal trip: %w", err)
	}
	err = p.rabbitmq.PublishMessage(ctx, contracts.TripEventCreated, &contracts.AmqpMessage{
		OwnerID: trip.UserID,
		Data:    data,
	})
	if err != nil {
		return fmt.Errorf("failed to publish trip created event: %w", err)
	}
	return nil
}
