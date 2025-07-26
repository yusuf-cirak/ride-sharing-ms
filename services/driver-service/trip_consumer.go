package main

import (
	"context"
	"ride-sharing/shared/messaging"

	"github.com/rabbitmq/amqp091-go"
)

type tripConsumer struct {
	rabbitMQ *messaging.RabbitMQ
}

func NewTripConsumer(rabbitMQ *messaging.RabbitMQ) *tripConsumer {
	return &tripConsumer{
		rabbitMQ: rabbitMQ,
	}
}

func (c *tripConsumer) Listen() error {
	return c.rabbitMQ.ConsumeMessages("hello", func(ctx context.Context, msg amqp091.Delivery) error {
		// Handle the incoming message
		return nil
	})
}
