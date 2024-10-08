package cloudmux

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/soheilhy/cmux"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

// ServeGRPCHTTP serves both a gRPC and an HTTP server on listener l.
// When the context is canceled, the servers will be gracefully shutdown and
// then the function will return.
func ServeGRPCHTTP(
	ctx context.Context,
	l net.Listener,
	grpcServer *grpc.Server,
	httpServer *http.Server,
) error {
	m := cmux.New(l)
	grpcL := m.MatchWithWriters(
		cmux.HTTP2MatchHeaderFieldSendSettings("content-type", "application/grpc"),
		cmux.HTTP2MatchHeaderFieldSendSettings("content-type", "application/grpc+proto"),
	)
	httpL := m.Match(cmux.Any())
	var g errgroup.Group
	// wait for context to be canceled and gracefully stop all servers.
	g.Go(func() error {
		<-ctx.Done()
		slog.DebugContext(ctx, "stopping cmux server")
		m.Close()
		slog.DebugContext(ctx, "stopping HTTP server")
		// use a new context because the parent ctx is already canceled.
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		if err := httpServer.Shutdown(ctx); err != nil && !isClosedErr(err) {
			slog.WarnContext(ctx, "stopping http server", slog.Any("error", err))
		}
		slog.DebugContext(ctx, "stopping gRPC server")
		grpcServer.GracefulStop()
		slog.DebugContext(ctx, "stopped both http and grpc server")
		return nil
	})

	g.Go(func() error {
		slog.DebugContext(ctx, "serving gRPC")
		if err := grpcServer.Serve(grpcL); err != nil && !isClosedErr(err) {
			return fmt.Errorf("serve gRPC: %w", err)
		}
		slog.DebugContext(ctx, "stopped serving gRPC")
		return nil
	})

	g.Go(func() error {
		slog.DebugContext(ctx, "serving HTTP")
		if err := httpServer.Serve(httpL); err != nil && !isClosedErr(err) {
			return fmt.Errorf("serve HTTP: %w", err)
		}
		slog.DebugContext(ctx, "stopped serving HTTP")
		return nil
	})

	if err := m.Serve(); err != nil && !isClosedErr(err) {
		slog.ErrorContext(ctx, "oops", slog.Any("error", err))
		return fmt.Errorf("serve cmux: %w", err)
	}
	return g.Wait()
}

func isClosedErr(err error) bool {
	return isClosedConnErr(err) ||
		errors.Is(err, http.ErrServerClosed) ||
		errors.Is(err, cmux.ErrServerClosed) ||
		errors.Is(err, grpc.ErrServerStopped)
}

func isClosedConnErr(err error) bool {
	return strings.Contains(err.Error(), "use of closed network connection")
}
