package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RideFareModel struct {
	ID primitive.ObjectID `json:"_id,omitempty"`
	UserID string `json:"userId"`
	PackageSlug string `json:"packageSlug"` // ex: van, luxury, sedan
	TotalPriceInCents float64 `json:"totalPriceInCents"`
	ExpiresAt time.Time `json:"expiresAt"`
}