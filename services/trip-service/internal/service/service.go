package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"ride-sharing/services/trip-service/internal/domain"
	tripTypes "ride-sharing/services/trip-service/pkg/types"
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
		ID:       primitive.NewObjectID(),
		UserID:   fare.UserID,
		Status:   "pending",
		RideFare: fare,
	}

	return s.repo.CreateTrip(ctx, &trip)
}

func (s *service) GetRoute(ctx context.Context, pickup, destination *types.Coordinate) (*tripTypes.OsrmApiResponse, error) {
	const baseUrl = "http://router.project-osrm.org"

	url := fmt.Sprintf("%s/route/v1/driving/%f,%f;%f,%f?overview=full&geometries=geojson", baseUrl, pickup.Latitude, pickup.Longitude, destination.Latitude, destination.Longitude)

	response, err := http.Get(url)

	if err != nil {
		return nil, fmt.Errorf("failed to fetch from OSRM API: %w", err)
	}

	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var route tripTypes.OsrmApiResponse

	if err := json.Unmarshal(body, &route); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &route, nil
}

func (s *service) EstimatePackagesPriceWithRoute(route *tripTypes.OsrmApiResponse) []*domain.RideFareModel {
	baseFares := getBaseFares()

	estimatedFares := make([]*domain.RideFareModel, len(baseFares))

	for i, fare := range baseFares {
		estimatedFares[i] = estimateFareRoute(fare, route)
	}

	return estimatedFares
}

func estimateFareRoute(f *domain.RideFareModel, r *tripTypes.OsrmApiResponse) *domain.RideFareModel {
	pricingConfig := tripTypes.DefaultPricingConfig()

	carPackagePrice := f.TotalPriceInCents

	distanceKm := r.Routes[0].Distance
	durationInMinutes := r.Routes[0].Duration

	distanceFare := distanceKm * pricingConfig.PricePerUnitOfDistance
	timeFare := durationInMinutes * pricingConfig.PricingPerMinute
	totalPrice := distanceFare + timeFare + carPackagePrice

	return &domain.RideFareModel{
		TotalPriceInCents: totalPrice,
		PackageSlug:       f.PackageSlug,
	}
}
func (s *service) GenerateTripFares(ctx context.Context, f []*domain.RideFareModel, userID string) ([]*domain.RideFareModel, error) {
	for _, fare := range f {
		fare.ID = primitive.NewObjectID()
		fare.UserID = userID

		if err := s.repo.SaveRideFare(ctx, fare); err != nil {
			return nil, fmt.Errorf("failed to save ride fare: %w", err)
		}
	}

	return f, nil
}

func getBaseFares() []*domain.RideFareModel {
	return []*domain.RideFareModel{
		{
			PackageSlug:       "suv",
			TotalPriceInCents: 200,
		},
		{
			PackageSlug:       "sedan",
			TotalPriceInCents: 350,
		},
		{
			PackageSlug:       "van",
			TotalPriceInCents: 400,
		},
		{
			PackageSlug:       "luxury",
			TotalPriceInCents: 1000,
		},
	}
}
