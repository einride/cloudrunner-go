package cloudmux

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/soheilhy/cmux"
	"go.einride.tech/cloudrunner/cloudzap"
	"go.uber.org/zap"
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
	grpcL := m.MatchWithWriters(cmux.HTTP2MatchHeaderFieldSendSettings("content-type", "application/grpc"))
	httpL := m.Match(cmux.Any())
	logger, ok := cloudzap.GetLogger(ctx)
	if !ok {
		logger = zap.NewNop()
	}

	var g errgroup.Group

	// wait for context to be canceled and gracefully stop all servers.
	g.Go(func() error {
		<-ctx.Done()

		logger.Debug("stopping cmux server")
		m.Close()

		logger.Debug("stopping HTTP server")
		// use a new context because the parent ctx is already canceled.
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		if err := httpServer.Shutdown(ctx); err != nil && !isClosedErr(err) {
			logger.Warn("stopping http server", zap.Error(err))
		}

		logger.Debug("stopping gRPC server")
		grpcServer.GracefulStop()
		logger.Debug("stopped both http and grpc server")
		return nil
	})

	g.Go(func() error {
		logger.Debug("serving gRPC")
		if err := grpcServer.Serve(grpcL); err != nil && !isClosedErr(err) {
			return fmt.Errorf("serve gRPC: %w", err)
		}
		logger.Debug("stopped serving gRPC")
		return nil
	})

	g.Go(func() error {
		logger.Debug("serving HTTP")
		if err := httpServer.Serve(httpL); err != nil && !isClosedErr(err) {
			return fmt.Errorf("serve HTTP: %w", err)
		}
		logger.Debug("stopped serving HTTP")
		return nil
	})

	if err := m.Serve(); err != nil && !isClosedErr(err) {
		logger.Error("oops", zap.Error(err))
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
