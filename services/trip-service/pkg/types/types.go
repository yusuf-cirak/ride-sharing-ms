package types

import tripGrpc "ride-sharing/shared/proto/trip"

type OsrmApiResponse struct {
	Routes []struct {
		Distance float64 `json:"distance"`
		Duration float64 `json:"duration"`
		Geometry struct {
			Coordinates [][]float64 `json:"coordinates"`
		} `json:"geometry"`
	} `json:"routes"`
}

func (r *OsrmApiResponse) ToProto() *tripGrpc.Route {
	if len(r.Routes) == 0 {
		return &tripGrpc.Route{
			Geometry: []*tripGrpc.Geometry{},
			Distance: 0,
			Duration: 0,
		}
	}

	route := r.Routes[0]
	geometry := route.Geometry.Coordinates

	coordinates := make([]*tripGrpc.Coordinate, len(geometry))

	for i, coord := range geometry {
		coordinates[i] = &tripGrpc.Coordinate{
			Latitude:  coord[0],
			Longitude: coord[1],
		}
	}

	return &tripGrpc.Route{
		Geometry: []*tripGrpc.Geometry{
			{
				Coordinates: coordinates,
			},
		},
		Distance: route.Distance,
		Duration: route.Duration,
	}
}

type PricingConfig struct {
	PricePerUnitOfDistance float64
	PricingPerMinute       float64
}

func DefaultPricingConfig() *PricingConfig {
	return &PricingConfig{
		PricePerUnitOfDistance: 1.5,
		PricingPerMinute:       0.25, // Example value
	}
}
