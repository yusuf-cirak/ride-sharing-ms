package domain

import (
	"context"

	"ride-sharing/services/payment-service/pkg/types"
)

type Service interface {
	CreatePaymentSession(ctx context.Context, tripID, userID, driverID string, amount int64) (*types.PaymentIntent, error)
}

type PaymentProcessor interface {
	CreatePaymentSession(ctx context.Context, amount int64, metadata map[string]string) (string, error)
}
