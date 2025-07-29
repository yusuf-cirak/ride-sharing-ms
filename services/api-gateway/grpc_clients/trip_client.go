package grpc_clients

import (
	"os"
	tripGrpc "ride-sharing/shared/proto/trip"
	"ride-sharing/shared/tracing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type tripServiceClient struct {
	Client tripGrpc.TripServiceClient
	conn   *grpc.ClientConn
}

func NewTripServiceClient() (*tripServiceClient, error) {
	tripServiceUrl := os.Getenv("TRIP_SERVICE_URL")
	if tripServiceUrl == "" {
		tripServiceUrl = "trip-service:9093"
	}

	dialOptions := append(tracing.DialOptionsWithTracing(), grpc.WithTransportCredentials(insecure.NewCredentials()))

	conn, err := grpc.NewClient(tripServiceUrl, dialOptions...)

	if err != nil {
		return nil, err
	}

	client := tripGrpc.NewTripServiceClient(conn)

	return &tripServiceClient{
		Client: client,
		conn:   conn,
	}, nil

}

func (c *tripServiceClient) Close() {
	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			// Handle error if needed
		}
	}
}
