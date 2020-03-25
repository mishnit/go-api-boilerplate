package grpc

import (
	"context"
	"fmt"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

	"github.com/vardius/go-api-boilerplate/pkg/grpc/middleware"
	"github.com/vardius/go-api-boilerplate/pkg/grpc/middleware/firewall"
	"github.com/vardius/go-api-boilerplate/pkg/log"
)

// ConnectionConfig provides values for gRPC connection configuration
type ConnectionConfig struct {
	ConnTime    time.Duration
	ConnTimeout time.Duration
}

// NewConnection provides new grpc connection
func NewConnection(ctx context.Context, host string, port int, cfg ConnectionConfig, logger *log.Logger) *grpc.ClientConn {
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                cfg.ConnTime,    // send pings every 10 seconds if there is no activity
			Timeout:             cfg.ConnTimeout, // wait 20 second for ping ack before considering the connection dead
			PermitWithoutStream: true,            // send pings even without active streams
		}),
		grpc.WithChainUnaryInterceptor(
			middleware.AppendMetadataToOutgoingUnaryContext(),
			middleware.LogOutgoingUnaryRequest(logger),
			firewall.AppendIdentityToOutgoingUnaryContext(),
		),
		grpc.WithChainStreamInterceptor(
			middleware.AppendMetadataToOutgoingStreamContext(),
			middleware.LogOutgoingStreamRequest(logger),
			firewall.AppendIdentityToOutgoingStreamContext(),
		),
	}
	conn, err := grpc.DialContext(ctx, fmt.Sprintf("%s:%d", host, port), opts...)
	if err != nil {
		logger.Critical(ctx, "[gRPC|Client] auth conn dial error: %v\n", err)
		os.Exit(1)
	}

	return conn
}