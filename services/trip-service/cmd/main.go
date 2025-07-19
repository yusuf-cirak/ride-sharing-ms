package main

import (
	"log"
	"net/http"
	h "ride-sharing/services/trip-service/internal/infrastructure/http"
	"ride-sharing/services/trip-service/internal/infrastructure/repository"
	"ride-sharing/services/trip-service/internal/service"
)

func main() {
	inmemRepo := repository.NewInmemRepository()

	svc := service.NewTripService(inmemRepo)

	httpHandler := h.HttpHandler{
		Service: svc,
	}

	mux := http.NewServeMux()

	mux.HandleFunc("POST /trip/preview", httpHandler.HandleTripReview)


	server:= &http.Server{
		Addr:    ":8083",
		Handler: mux,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}

}