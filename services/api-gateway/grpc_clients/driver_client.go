package grpc_clients

import (
	"os"
	driverGrpc "ride-sharing/shared/proto/driver"
	"ride-sharing/shared/tracing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type driverServiceClient struct {
	Client driverGrpc.DriverServiceClient
	conn   *grpc.ClientConn
}

func NewDriverServiceClient() (*driverServiceClient, error) {
	driverServiceUrl := os.Getenv("DRIVER_SERVICE_URL")
	if driverServiceUrl == "" {
		driverServiceUrl = "driver-service:9092"
	}

	dialOptions := append(tracing.DialOptionsWithTracing(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	conn, err := grpc.NewClient(driverServiceUrl, dialOptions...)

	if err != nil {
		return nil, err
	}

	client := driverGrpc.NewDriverServiceClient(conn)

	return &driverServiceClient{
		Client: client,
		conn:   conn,
	}, nil

}

func (c *driverServiceClient) Close() {
	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			// Handle error if needed
		}
	}
}
