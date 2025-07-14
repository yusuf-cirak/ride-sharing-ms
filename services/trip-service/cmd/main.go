package main

import (
	"context"
	"log"
	"ride-sharing/services/trip-service/internal/domain"
	"ride-sharing/services/trip-service/internal/infrastructure/repository"
	"ride-sharing/services/trip-service/internal/service"
	"time"
)

func main() {
	ctx:= context.Background();
	inmemRepo := repository.NewInmemRepository()

	svc:= service.NewTripService(inmemRepo);

	fare:= &domain.RideFareModel{
		UserID: "123",
		PackageSlug: "van",
		TotalPriceInCents: 1000,
	}

	t,err := svc.CreateTrip(ctx,fare)

	if err != nil {
		log.Fatal(err)
	}

	log.Println(t)

	for  {
		time.Sleep(time.Second)
	}
}