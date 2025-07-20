package grpc

import (
	"context"
	"ride-sharing/services/trip-service/internal/domain"
	tripGrpc "ride-sharing/shared/proto/trip"
	"ride-sharing/shared/types"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type gRPCHandler struct {
	tripGrpc.UnimplementedTripServiceServer
	service domain.TripService
}

func NewGRPCHandler(server *grpc.Server, service domain.TripService) *gRPCHandler {
	handler := &gRPCHandler{
		service: service,
	}

	tripGrpc.RegisterTripServiceServer(server, handler)
	return handler
}

func (h *gRPCHandler) PreviewTrip(ctx context.Context, req *tripGrpc.PreviewTripRequest) (*tripGrpc.PreviewTripResponse, error) {
	pickup := req.GetStartLocation()
	destination := req.GetEndLocation()

	pickupCoord := &types.Coordinate{
		Latitude:  pickup.GetLatitude(),
		Longitude: pickup.GetLongitude(),
	}
	destinationCoord := &types.Coordinate{
		Latitude:  destination.GetLatitude(),
		Longitude: destination.GetLongitude(),
	}

	t, err := h.service.GetRoute(ctx, pickupCoord, destinationCoord)

	if err != nil {
		return nil, status.Errorf(codes.Aborted, "failed to get route: %v", err)
	}

	return &tripGrpc.PreviewTripResponse{
		Route:     t.ToProto(),
		RideFares: []*tripGrpc.RideFare{},
	}, nil
}
