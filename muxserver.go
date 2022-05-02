package cloudrunner

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"go.einride.tech/cloudrunner/cloudmux"
	"google.golang.org/grpc"
)

// ListenGRPCHTTP binds a listener on the configured port and listens for gRPC and HTTP requests.
func ListenGRPCHTTP(ctx context.Context, grpcServer *grpc.Server, httpServer *http.Server) error {
	l, err := (&net.ListenConfig{}).Listen(ctx, "tcp", fmt.Sprintf(":%d", Runtime(ctx).Port))
	if err != nil {
		return fmt.Errorf("serve gRPC and HTTP: %w", err)
	}
	if err := cloudmux.ServeGRPCHTTP(ctx, l, grpcServer, httpServer); err != nil {
		return fmt.Errorf("serve gRPC and HTTP: %w", err)
	}
	return nil
}
