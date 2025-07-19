package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"ride-sharing/services/trip-service/internal/domain"
	"ride-sharing/shared/types"

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


func (s *service) GetRoute(ctx context.Context, pickup,destination *types.Coordinate) (*types.OsrmApiResponse,error){
	const baseUrl = "http://router.project-osrm.org"
	
	url:= fmt.Sprintf("%s/route/v1/driving/%f,%f;%f,%f?overview=full&geometries=geojson",baseUrl,pickup.Latitude,pickup.Longitude,destination.Latitude,destination.Longitude)


	response,err :=http.Get(url)

	if err != nil {
		return nil, fmt.Errorf("failed to fetch from OSRM API: %w", err)
	}

	defer response.Body.Close()

	body,err:= io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var route types.OsrmApiResponse

	if err:= json.Unmarshal(body,&route); err!=nil{
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &route, nil
}