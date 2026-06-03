package client

import (
	"context"

	ft12v1 "github.com/dimbo1324/ttron-ttr20-time-r/internal/api/grpc/ft12/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func DialEmulator(ctx context.Context, addr string) (ft12v1.EmulatorServiceClient, *grpc.ClientConn, error) {
	if err := ctx.Err(); err != nil {
		return nil, nil, err
	}
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, err
	}
	conn.Connect()
	return ft12v1.NewEmulatorServiceClient(conn), conn, nil
}
