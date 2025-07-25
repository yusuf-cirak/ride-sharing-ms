package main

import (
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

func handleRidersWebSocket(w http.ResponseWriter, r *http.Request) {
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

	for {
		_, message, err := conn.ReadMessage()

		if err != nil {
			log.Printf("Error reading message: %v", err)
			break
		}

		log.Printf("Received message from rider %s: %s", userID, message)
	}
}

func handleDriversWebSocket(w http.ResponseWriter, r *http.Request) {
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

	for {
		_, message, err := conn.ReadMessage()

		if err != nil {
			log.Printf("Error reading message: %v", err)
			break
		}

		log.Printf("Received message from rider %s: %s", userID, message)
	}
}
