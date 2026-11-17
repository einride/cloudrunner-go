package cloudrequestlog

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/emptypb"
	"gotest.tools/v3/assert"
)

// Tests in this file replace the default logger and must not run in parallel.

func TestRootCause_LeafFailure(t *testing.T) {
	logs := captureLogs(t)
	middleware := &Middleware{Config: Config{PropagateErrorRootCause: true}}
	leaf := dial(t, middleware, startChainServer(t, middleware, "Leaf", func(context.Context) error {
		return status.Error(codes.Internal, "boom")
	}))
	middle := dial(t, middleware, startChainServer(t, middleware, "Middle", forward(leaf, "Leaf")))
	edge := dial(t, middleware, startChainServer(t, middleware, "Edge", forward(middle, "Middle")))

	err := invoke(context.Background(), edge, "Edge")
	s, _ := status.FromError(err)
	assert.Equal(t, s.Message(), "boom")
	assert.Equal(t, s.Code(), codes.Internal)

	rootCause := rootCauseFingerprint("gRPCServer", codes.Internal, "/test.ChainService/Leaf")
	assert.DeepEqual(t, logs.errorRootCauses(), map[string]string{
		"gRPCServer Internal Leaf":   rootCause,
		"gRPCClient Internal Leaf":   rootCause,
		"gRPCServer Internal Middle": rootCause,
		"gRPCClient Internal Middle": rootCause,
		"gRPCServer Internal Edge":   rootCause,
		"gRPCClient Internal Edge":   rootCause,
	})
}

func TestRootCause_MiddleFailsOnLeafWithoutRootCause(t *testing.T) {
	logs := captureLogs(t)
	middleware := &Middleware{Config: Config{PropagateErrorRootCause: true}}
	middlewareNoRootCause := &Middleware{Config: Config{PropagateErrorRootCause: false}}
	leaf := dial(t, middleware, startChainServer(t, middlewareNoRootCause, "Leaf", func(context.Context) error {
		return status.Error(codes.Internal, "boom")
	}))
	middle := dial(t, middleware, startChainServer(t, middleware, "Middle", forward(leaf, "Leaf")))
	edge := dial(t, middleware, startChainServer(t, middleware, "Edge", forward(middle, "Middle")))

	err := invoke(context.Background(), edge, "Edge")
	assert.Equal(t, status.Code(err), codes.Internal)

	rootCause := rootCauseFingerprint("gRPCClient", codes.Internal, "/test.ChainService/Leaf")
	assert.DeepEqual(t, logs.errorRootCauses(), map[string]string{
		// The Leaf server does not propagate root causes.
		"gRPCServer Internal Leaf":   "",
		"gRPCClient Internal Leaf":   rootCause,
		"gRPCServer Internal Middle": rootCause,
		"gRPCClient Internal Middle": rootCause,
		"gRPCServer Internal Edge":   rootCause,
		"gRPCClient Internal Edge":   rootCause,
	})
}

func TestRootCause_WarnLevelLeafFailureIsNotRootCause(t *testing.T) {
	logs := captureLogs(t)
	middleware := &Middleware{Config: Config{PropagateErrorRootCause: true}}
	leaf := dial(t, middleware, startChainServer(t, middleware, "Leaf", func(context.Context) error {
		return status.Error(codes.NotFound, "not found")
	}))
	middle := dial(t, middleware, startChainServer(t, middleware, "Middle", func(ctx context.Context) error {
		if err := invoke(ctx, leaf, "Leaf"); err != nil {
			// Escalate the warn level NotFound to an error level failure of Middle itself.
			return status.Error(codes.Internal, err.Error())
		}
		return nil
	}))
	edge := dial(t, middleware, startChainServer(t, middleware, "Edge", forward(middle, "Middle")))

	err := invoke(context.Background(), edge, "Edge")
	assert.Equal(t, status.Code(err), codes.Internal)

	rootCause := rootCauseFingerprint("gRPCServer", codes.Internal, "/test.ChainService/Middle")
	assert.DeepEqual(t, logs.errorRootCauses(), map[string]string{
		"gRPCServer Internal Middle": rootCause,
		"gRPCClient Internal Middle": rootCause,
		"gRPCServer Internal Edge":   rootCause,
		"gRPCClient Internal Edge":   rootCause,
	})
}

func TestRootCause_InvalidTrailerIsIgnored(t *testing.T) {
	logs := captureLogs(t)
	middleware := &Middleware{Config: Config{PropagateErrorRootCause: true}}
	middlewareNoRootCause := &Middleware{Config: Config{PropagateErrorRootCause: false}}
	leaf := dial(t, middleware, startChainServer(t, middlewareNoRootCause, "Leaf", func(ctx context.Context) error {
		assert.NilError(t, grpc.SetTrailer(ctx, rootCauseTrailer("internal details")))
		return status.Error(codes.Internal, "boom")
	}))
	edge := dial(t, middleware, startChainServer(t, middleware, "Edge", forward(leaf, "Leaf")))

	err := invoke(context.Background(), edge, "Edge")
	assert.Equal(t, status.Code(err), codes.Internal)

	rootCause := rootCauseFingerprint("gRPCClient", codes.Internal, "/test.ChainService/Leaf")
	assert.DeepEqual(t, logs.errorRootCauses(), map[string]string{
		"gRPCServer Internal Leaf": "",
		"gRPCClient Internal Leaf": rootCause,
		"gRPCServer Internal Edge": rootCause,
		"gRPCClient Internal Edge": rootCause,
	})
}

func TestRootCause_Disabled(t *testing.T) {
	logs := captureLogs(t)
	middleware := &Middleware{}
	leaf := dial(t, middleware, startChainServer(t, middleware, "Leaf", func(context.Context) error {
		return status.Error(codes.Internal, "boom")
	}))
	edge := dial(t, middleware, startChainServer(t, middleware, "Edge", forward(leaf, "Leaf")))

	err := invoke(context.Background(), edge, "Edge")
	assert.Equal(t, status.Code(err), codes.Internal)

	assert.DeepEqual(t, logs.errorRootCauses(), map[string]string{
		"gRPCServer Internal Leaf": "",
		"gRPCClient Internal Leaf": "",
		"gRPCServer Internal Edge": "",
		"gRPCClient Internal Edge": "",
	})
}

func TestRootCause_HTTPServerLogsDownstreamRootCause(t *testing.T) {
	logs := captureLogs(t)
	middleware := &Middleware{Config: Config{PropagateErrorRootCause: true}}
	leaf := dial(t, middleware, startChainServer(t, middleware, "Leaf", func(context.Context) error {
		return status.Error(codes.Internal, "boom")
	}))
	handler := middleware.HTTPServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := invoke(r.Context(), leaf, "Leaf"); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))

	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/edge", nil))

	rootCause := rootCauseFingerprint("gRPCServer", codes.Internal, "/test.ChainService/Leaf")
	assert.DeepEqual(t, logs.errorRootCauses(), map[string]string{
		"gRPCServer Internal Leaf":        rootCause,
		"gRPCClient Internal Leaf":        rootCause,
		"HTTPServer 500 example.com/edge": rootCause,
	})
}

func TestRootCause_HTTPServerWithoutDownstreamFailureHasNoRootCause(t *testing.T) {
	logs := captureLogs(t)
	middleware := &Middleware{Config: Config{PropagateErrorRootCause: true}}
	handler := middleware.HTTPServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))

	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/edge", nil))

	assert.DeepEqual(t, logs.errorRootCauses(), map[string]string{
		"HTTPServer 500 example.com/edge": "",
	})
}

const chainServiceName = "test.ChainService"

// startChainServer starts an in-memory gRPC server with request logging, serving a single method of the chain
// service that runs handle, and returns the listener to dial it with.
func startChainServer(
	t *testing.T,
	middleware *Middleware,
	method string,
	handle func(context.Context) error,
) *bufconn.Listener {
	t.Helper()
	listener := bufconn.Listen(1024 * 1024)
	server := grpc.NewServer(grpc.ChainUnaryInterceptor(middleware.GRPCUnaryServerInterceptor))
	server.RegisterService(&grpc.ServiceDesc{
		ServiceName: chainServiceName,
		HandlerType: (*any)(nil),
		Methods: []grpc.MethodDesc{
			{
				MethodName: method,
				Handler: func(
					_ any,
					ctx context.Context,
					decode func(any) error,
					interceptor grpc.UnaryServerInterceptor,
				) (any, error) {
					request := &emptypb.Empty{}
					if err := decode(request); err != nil {
						return nil, err
					}
					handler := func(ctx context.Context, _ any) (any, error) {
						if err := handle(ctx); err != nil {
							return nil, err
						}
						return &emptypb.Empty{}, nil
					}
					info := &grpc.UnaryServerInfo{FullMethod: "/" + chainServiceName + "/" + method}
					return interceptor(ctx, request, info, handler)
				},
			},
		},
	}, struct{}{})
	go func() {
		_ = server.Serve(listener)
	}()
	t.Cleanup(server.Stop)
	return listener
}

// dial returns a client connection to the listener, with request logging by the caller's middleware.
func dial(t *testing.T, middleware *Middleware, listener *bufconn.Listener) *grpc.ClientConn {
	t.Helper()
	conn, err := grpc.NewClient(
		"passthrough:///bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return listener.Dial() }),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(middleware.GRPCUnaryClientInterceptor),
	)
	assert.NilError(t, err)
	t.Cleanup(func() { _ = conn.Close() })
	return conn
}

func forward(conn *grpc.ClientConn, method string) func(context.Context) error {
	return func(ctx context.Context) error {
		return invoke(ctx, conn, method)
	}
}

func invoke(ctx context.Context, conn *grpc.ClientConn, method string) error {
	return conn.Invoke(ctx, "/"+chainServiceName+"/"+method, &emptypb.Empty{}, &emptypb.Empty{})
}

type capturedLogs struct {
	mu      sync.Mutex
	records []slog.Record
}

// captureLogs replaces the default logger with one capturing all records for the duration of the test.
func captureLogs(t *testing.T) *capturedLogs {
	t.Helper()
	logs := &capturedLogs{}
	previous := slog.Default()
	slog.SetDefault(slog.New(logs))
	t.Cleanup(func() { slog.SetDefault(previous) })
	return logs
}

func (c *capturedLogs) Enabled(context.Context, slog.Level) bool { return true }

func (c *capturedLogs) Handle(_ context.Context, record slog.Record) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.records = append(c.records, record.Clone())
	return nil
}

func (c *capturedLogs) WithAttrs([]slog.Attr) slog.Handler { return c }

func (c *capturedLogs) WithGroup(string) slog.Handler { return c }

// errorRootCauses returns the rootCause field of each error level log record, keyed by log message.
func (c *capturedLogs) errorRootCauses() map[string]string {
	c.mu.Lock()
	defer c.mu.Unlock()
	result := map[string]string{}
	for _, record := range c.records {
		if record.Level < slog.LevelError {
			continue
		}
		result[record.Message] = ""
		record.Attrs(func(attr slog.Attr) bool {
			if attr.Key == rootCauseKey {
				result[record.Message] = attr.Value.String()
				return false
			}
			return true
		})
	}
	return result
}
