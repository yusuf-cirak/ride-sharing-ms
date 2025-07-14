package main

import (
	"ride-sharing/shared/types"
)

type previewTripRequest struct {
	UserID string `json:"userId"`
	Pickup types.Coordinate `json:"pickup"`
	Destination types.Coordinate `json:"destination"`
}