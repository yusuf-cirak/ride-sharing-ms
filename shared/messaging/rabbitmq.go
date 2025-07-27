package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"ride-sharing/shared/contracts"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	TripExchange = "trip"
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
	err := r.Channel.ExchangeDeclare(
		TripExchange, // name
		"topic",      // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare exchange: %s: %v", TripExchange, err)
	}

	if err := r.declareAndBindQueue(
		FindAvailableDriversQueue,
		[]string{
			contracts.TripEventCreated, contracts.TripEventDriverNotInterested,
		},
		TripExchange,
	); err != nil {
		return err
	}

	if err := r.declareAndBindQueue(
		DriverCmdTripRequestQueue,
		[]string{contracts.DriverCmdTripRequest},
		TripExchange,
	); err != nil {
		return err
	}

	if err := r.declareAndBindQueue(
		DriverTripResponseQueue,
		[]string{contracts.DriverCmdTripAccept, contracts.DriverCmdTripDecline},
		TripExchange,
	); err != nil {
		return err
	}

	if err := r.declareAndBindQueue(
		NotifyDriverNoDriversFoundQueue,
		[]string{contracts.TripEventNoDriversFound},
		TripExchange,
	); err != nil {
		return err
	}

	if err := r.declareAndBindQueue(
		NotifyDriverAssignQueue,
		[]string{contracts.TripEventDriverAssigned},
		TripExchange,
	); err != nil {
		return err
	}

	if err := r.declareAndBindQueue(
		PaymentTripResponseQueue,
		[]string{contracts.PaymentCmdCreateSession},
		TripExchange,
	); err != nil {
		return err
	}

	if err := r.declareAndBindQueue(
		NotifyPaymentSessionCreatedQueue,
		[]string{contracts.PaymentEventSessionCreated},
		TripExchange,
	); err != nil {
		return err
	}

	if err := r.declareAndBindQueue(
		NotifyPaymentSuccessQueue,
		[]string{contracts.PaymentEventSuccess},
		TripExchange,
	); err != nil {
		return err
	}

	return nil
}

func (r *RabbitMQ) declareAndBindQueue(queueName string, messageTypes []string, exchange string) error {
	_, err := r.Channel.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)

	if err != nil {
		return fmt.Errorf("failed to declare queue %s: %w", queueName, err)
	}

	for _, messageType := range messageTypes {
		if err := r.Channel.QueueBind(
			queueName,   // queue name
			messageType, // routing key
			exchange,    // exchange
			false,       // no-wait
			nil,         // arguments
		); err != nil {
			return fmt.Errorf("failed to bind queue %s to exchange %s: %w", queueName, exchange, err)
		}
	}

	return nil
}

type MessageHandler func(ctx context.Context, msg amqp.Delivery) error

func (r *RabbitMQ) ConsumeMessages(queueName string, handler MessageHandler) error {

	// Set prefetch count to 1 for fair dispatch
	// This tells RabbitMQ not to give more than one message to a service at a time.
	// The worker will only get the next message after it has acknowledged the previous one.
	err := r.Channel.Qos(
		1,     // prefetchCount: Limit to 1 unacknowledged message per consumer
		0,     // prefetchSize: No specific limit on message size
		false, // global: Apply prefetchCount to each consumer individually
	)
	if err != nil {
		return fmt.Errorf("failed to set QoS: %v", err)
	}

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

func (r *RabbitMQ) PublishMessage(ctx context.Context, routingKey string, message *contracts.AmqpMessage) error {
	jsonMessage, err := json.Marshal(message)

	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	err = r.Channel.PublishWithContext(ctx,
		TripExchange, // exchange
		routingKey,   // routing key
		false,        // mandatory
		false,        // immediate
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
