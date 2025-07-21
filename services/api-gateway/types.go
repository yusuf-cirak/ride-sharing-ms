package main

import (
	tripGrpc "ride-sharing/shared/proto/trip"
	"ride-sharing/shared/types"
)

type previewTripRequest struct {
	UserID      string           `json:"userId"`
	Pickup      types.Coordinate `json:"pickup"`
	Destination types.Coordinate `json:"destination"`
}

func (p *previewTripRequest) toProto() *tripGrpc.PreviewTripRequest {
	return &tripGrpc.PreviewTripRequest{
		UserID: p.UserID,
		StartLocation: &tripGrpc.Coordinate{
			Latitude:  p.Pickup.Latitude,
			Longitude: p.Pickup.Longitude,
		},
		EndLocation: &tripGrpc.Coordinate{
			Latitude:  p.Destination.Latitude,
			Longitude: p.Destination.Longitude,
		},
	}
}

type startTripRequest struct {
	RideFareID string `json:"rideFareID"`
	UserID     string `json:"userID"`
}

func (s *startTripRequest) toProto() *tripGrpc.CreateTripRequest {
	return &tripGrpc.CreateTripRequest{
		RideFareID: s.RideFareID,
		UserID:     s.UserID,
	}
}
