package domain

import (
	"time"

	tripTypes "ride-sharing/services/trip-service/pkg/types"
	tripGrpc "ride-sharing/shared/proto/trip"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RideFareModel struct {
	ID                primitive.ObjectID         `bson:"id"`
	UserID            string                     `bson:"userId"`
	PackageSlug       string                     `bson:"packageSlug"` // ex: van, luxury, sedan
	TotalPriceInCents float64                    `bson:"totalPriceInCents"`
	ExpiresAt         time.Time                  `bson:"expiresAt"`
	Route             *tripTypes.OsrmApiResponse `bson:"route"`
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
