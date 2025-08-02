package main

import (
	"encoding/json"
	"log"
	"net/http"
	"ride-sharing/services/api-gateway/grpc_clients"
	"ride-sharing/shared/contracts"
	"ride-sharing/shared/messaging"

	driverGrpc "ride-sharing/shared/proto/driver"
)

var (
	connManager = messaging.NewConnectionManager()
)

func handleRidersWebSocket(w http.ResponseWriter, r *http.Request, rb *messaging.RabbitMQ) {
	conn, err := connManager.Upgrade(w, r)

	if err != nil {
		log.Printf("Websocket upgrade failed: %v", err)
	}
	defer conn.Close()

	userID := r.URL.Query().Get("userID")

	if userID == "" {
		log.Println("User ID is required for WebSocket connection")
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	connManager.Add(userID, conn)
	defer connManager.Remove(userID)

	// initialize queue consumers
	queues := []string{
		messaging.NotifyDriverNoDriversFoundQueue,
		messaging.NotifyDriverAssignQueue,
		messaging.NotifyPaymentSessionCreatedQueue,
	}

	for _, q := range queues {
		consumer := messaging.NewQueueConsumer(rb, connManager, q)
		if err := consumer.Start(); err != nil {
			log.Printf("Error starting consumer for queue %s: %v", q, err)
		}
	}

	// reading messages from the rider from its ws connection. there is nothing they do now. they actually send http requests to the API gateway.
	for {
		_, message, err := conn.ReadMessage()

		if err != nil {
			log.Printf("Error reading message: %v", err)
			break
		}

		log.Printf("Received message from rider %s: %s", userID, message)
	}
}

func handleDriversWebSocket(w http.ResponseWriter, r *http.Request, rb *messaging.RabbitMQ) {
	conn, err := connManager.Upgrade(w, r)

	if err != nil {
		log.Printf("Websocket upgrade failed: %v", err)
	}

	defer conn.Close()

	userID := r.URL.Query().Get("userID")

	if userID == "" {
		log.Println("User ID is required for WebSocket connection")
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	packageSlug := r.URL.Query().Get("packageSlug")

	if packageSlug == "" {
		log.Println("Package slug is required for WebSocket connection")
		http.Error(w, "Package slug is required", http.StatusBadRequest)
		return
	}

	connManager.Add(userID, conn)

	ctx := r.Context()

	driverService, err := grpc_clients.NewDriverServiceClient()
	if err != nil {
		log.Printf("Error creating driver service client: %v", err)
		return
	}

	defer func() {
		connManager.Remove(userID)
		driverService.Client.UnregisterDriver(ctx, &driverGrpc.RegisterDriverRequest{
			DriverID:    userID,
			PackageSlug: packageSlug,
		})

		log.Printf("Driver %s unregistered successfully", userID)

		driverService.Close()
	}()

	driverData, err := driverService.Client.RegisterDriver(ctx, &driverGrpc.RegisterDriverRequest{
		DriverID: userID, PackageSlug: packageSlug,
	})

	if err != nil {
		log.Printf("Error registering driver: %v", err)
		return
	}

	msg := contracts.WSMessage{
		Type: contracts.DriverCmdRegister,
		Data: driverData.Driver,
	}

	if err := connManager.SendMessage(userID, msg); err != nil {
		log.Printf("Error sending message to driver %s: %v", userID, err)
		return
	}

	// initialize queue consumers
	queues := []string{
		messaging.DriverCmdTripRequestQueue,
	}

	// start queue consumers for the driver
	for _, q := range queues {
		consumer := messaging.NewQueueConsumer(rb, connManager, q)
		if err := consumer.Start(); err != nil {
			log.Printf("Error starting consumer for queue %s: %v", q, err)
		}
	}

	// reading messages from the driver from its ws connection.
	for {
		_, message, err := conn.ReadMessage()

		if err != nil {
			log.Printf("Error reading message: %v", err)
			break
		}

		type driverMessage struct {
			Type string          `json:"type"`
			Data json.RawMessage `json:"data"`
		}

		var driverMsg driverMessage

		if err := json.Unmarshal(message, &driverMsg); err != nil {
			log.Printf("Error unmarshalling driver message: %v", err)
			continue
		}

		switch driverMsg.Type {
		case contracts.DriverCmdLocation:
			continue
		case contracts.DriverCmdTripAccept, contracts.DriverCmdTripDecline:
			// forward the message to queue
			if err := rb.PublishMessage(ctx, driverMsg.Type, &contracts.AmqpMessage{
				OwnerID: userID,
				Data:    driverMsg.Data,
			}); err != nil {
				log.Printf("Error publishing message to RabbitMQ: %v", err)
			}
		default:
			log.Printf("Unknown driver command type: %s", driverMsg.Type)
		}

		log.Printf("Received message from rider %s: %s", userID, message)
	}
}
