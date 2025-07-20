package main

import (
	"log"
	"net/http"
	"ride-sharing/shared/contracts"
	"ride-sharing/shared/util"

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

	type Driver struct {
		Id             string `json:"id"`
		Name           string `json:"name"`
		ProfilePicture string `json:"profilePicture"`
		CarPlate       string `json:"carPlate"`
		PackageSlug    string `json:"packageSlug"`
	}

	msg := contracts.WSMessage{
		Type: "driver.cmd.register",
		Data: Driver{
			Id:             userId,
			Name:           "Tiago",
			ProfilePicture: util.GetRandomAvatar(1),
			PackageSlug:    packageSlug,
			CarPlate:       "ABC-1234",
		},
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
