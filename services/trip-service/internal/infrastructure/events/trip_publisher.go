package events

import (
	"context"
	"fmt"
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

func (p *TripEventPublisher) PublishTripCreated(ctx context.Context) error {
	err := p.rabbitmq.PublishMessage(ctx, "trip_queue", []byte("Trip Created Event"))
	if err != nil {
		return fmt.Errorf("failed to publish trip created event: %w", err)
	}
	return nil
}
