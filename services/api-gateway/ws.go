package main

import (
	"log"
	"net/http"
	"ride-sharing/services/api-gateway/grpc_clients"
	"ride-sharing/shared/contracts"

	driverGrpc "ride-sharing/shared/proto/driver"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func handleRidersWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Printf("Websocket upgrade failed: %v", err)
	}
	defer conn.Close()

	userId := r.URL.Query().Get("userID")

	if userId == "" {
		log.Println("User ID is required for WebSocket connection")
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	for {
		_, message, err := conn.ReadMessage()

		if err != nil {
			log.Printf("Error reading message: %v", err)
			break
		}

		log.Printf("Received message from rider %s: %s", userId, message)
	}
}

func handleDriversWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Printf("Websocket upgrade failed: %v", err)
	}

	defer conn.Close()

	userId := r.URL.Query().Get("userID")

	if userId == "" {
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

	ctx := r.Context()

	driverService, err := grpc_clients.NewDriverServiceClient()
	if err != nil {
		log.Printf("Error creating driver service client: %v", err)
		return
	}

	defer func() {
		driverService.Client.UnregisterDriver(ctx, &driverGrpc.RegisterDriverRequest{
			DriverID:    userId,
			PackageSlug: packageSlug,
		})

		log.Printf("Driver %s unregistered successfully", userId)

		driverService.Close()
	}()

	driverData, err := driverService.Client.RegisterDriver(ctx, &driverGrpc.RegisterDriverRequest{
		DriverID: userId, PackageSlug: packageSlug,
	})

	if err != nil {
		log.Printf("Error registering driver: %v", err)
		return
	}

	msg := contracts.WSMessage{
		Type: "driver.cmd.register",
		Data: driverData.Driver,
	}

	if err := conn.WriteJSON(msg); err != nil {
		log.Printf("Error sending message to driver %s: %v", userId, err)
		return
	}

	for {
		_, message, err := conn.ReadMessage()

		if err != nil {
			log.Printf("Error reading message: %v", err)
			break
		}

		log.Printf("Received message from rider %s: %s", userId, message)
	}
}
