package main

import (
	"context"
	pb "ride-sharing/shared/proto/driver"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type grpcHandler struct {
	pb.UnimplementedDriverServiceServer
	Service *Service
}

func NewGrpcHandler(s *grpc.Server, service *Service) *grpcHandler {
	grpcHandler := &grpcHandler{
		Service: service,
	}

	pb.RegisterDriverServiceServer(s, grpcHandler)

	return grpcHandler
}

func (h *grpcHandler) RegisterDriver(ctx context.Context, req *pb.RegisterDriverRequest) (*pb.RegisterDriverResponse, error) {
	driver, err := h.Service.RegisterDriver(req.GetDriverID(), req.GetPackageSlug())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to register driver: %v", err)
	}
	return &pb.RegisterDriverResponse{
		Driver: driver,
	}, nil
}
func (h *grpcHandler) UnregisterDriver(ctx context.Context, req *pb.RegisterDriverRequest) (*pb.RegisterDriverResponse, error) {
	if req.GetDriverID() == "" {
		return nil, status.Error(codes.InvalidArgument, "driver ID is required")
	}

	h.Service.UnregisterDriver(req.GetDriverID())
	return &pb.RegisterDriverResponse{}, nil
}
