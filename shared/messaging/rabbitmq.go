package messaging

import (
	"context"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQ struct {
	conn    *amqp.Connection
	Channel *amqp.Channel
}

func NewRabbitMQ(uri string) (*RabbitMQ, error) {
	conn, err := amqp.Dial(uri)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}

	rabbitmq := &RabbitMQ{
		conn:    conn,
		Channel: channel,
	}

	if err := rabbitmq.setupExchangesAndQueues(); err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to set up exchanges and queues: %w", err)
	}

	return rabbitmq, nil
}

func (r *RabbitMQ) setupExchangesAndQueues() error {
	_, err := r.Channel.QueueDeclare(
		"trip_queue", // name
		true,         // durable
		false,        // delete when unused
		false,        // exclusive
		false,        // no-wait
		nil,          // arguments
	)

	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	return nil
}

type MessageHandler func(ctx context.Context, msg amqp.Delivery) error

func (r *RabbitMQ) ConsumeMessages(queueName string, handler MessageHandler) error {
	msgs, err := r.Channel.Consume(
		queueName, // queue
		"",        // consumer
		false,     // auto-ack
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)
	if err != nil {
		fmt.Printf("Failed to register a consumer: %v\n", err)
		return fmt.Errorf("failed to register a consumer: %w", err)
	}

	go func() {
		for msg := range msgs {
			ctx := context.Background()
			if err := handler(ctx, msg); err != nil {
				err = msg.Nack(false, false) // nack the message if handling fails

				if err != nil {
					// Log the error if nack fails
					fmt.Printf("Failed to nack message: %v\n", err)
				}
				fmt.Printf("Error handling message: %v\n", err)

				continue
			}

			msg.Ack(false) // ack the message if handling succeeds
			fmt.Printf("Message processed: %s\n", msg.Body)
		}
	}()

	return nil
}

func (r *RabbitMQ) PublishMessage(ctx context.Context, routingKey string, message any) error {
	jsonMessage, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	err = r.Channel.PublishWithContext(ctx,
		"",         // exchange
		routingKey, // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         jsonMessage,
			DeliveryMode: amqp.Persistent, // make message persistent
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}
	return nil
}

func (r *RabbitMQ) Close() error {
	if r.conn != nil {
		return r.conn.Close()
	}

	if r.Channel != nil {
		if err := r.Channel.Close(); err != nil {
			return fmt.Errorf("failed to close channel: %w", err)
		}
	}
	return nil
}
