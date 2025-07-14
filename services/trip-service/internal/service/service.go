package service

import (
	"context"
	"ride-sharing/services/trip-service/internal/domain"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type service struct {
	repo domain.TripRepository
}

func NewTripService(r domain.TripRepository) *service {
	return &service{
		repo: r,
	}
}

func (s *service) CreateTrip(ctx context.Context, fare *domain.RideFareModel) (*domain.TripModel, error) {

	trip := domain.TripModel{
		ID: primitive.NewObjectID(),
		UserID: fare.UserID,
		Status: "pending",
		RideFare: fare,
	}

	return s.repo.CreateTrip(ctx, &trip)
}
