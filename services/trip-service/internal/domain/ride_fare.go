package domain

import (
	"time"

	tripGrpc "ride-sharing/shared/proto/trip"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RideFareModel struct {
	ID                primitive.ObjectID `json:"_id,omitempty"`
	UserID            string             `json:"userId"`
	PackageSlug       string             `json:"packageSlug"` // ex: van, luxury, sedan
	TotalPriceInCents float64            `json:"totalPriceInCents"`
	ExpiresAt         time.Time          `json:"expiresAt"`
}

func (r *RideFareModel) ToProto() *tripGrpc.RideFare {
	return &tripGrpc.RideFare{
		Id:                r.ID.Hex(),
		UserId:            r.UserID,
		PackageSlug:       r.PackageSlug,
		TotalPriceInCents: r.TotalPriceInCents,
	}
}

func ToRideFaresProto(fares []*RideFareModel) []*tripGrpc.RideFare {
	if len(fares) == 0 {
		return nil
	}

	protoFares := make([]*tripGrpc.RideFare, 0, len(fares))
	for _, fare := range fares {
		protoFares = append(protoFares, fare.ToProto())
	}

	return protoFares
}
