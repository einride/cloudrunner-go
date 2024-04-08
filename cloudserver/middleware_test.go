package cloudserver_test

import (
	"context"
	"log"
	"net"
	"testing"
	"time"

	testproto "github.com/grpc-ecosystem/go-grpc-middleware/testing/testproto"
	"go.einride.tech/cloudrunner/cloudserver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"gotest.tools/v3/assert"
)

const bufSize = 1024 * 1024

type Server struct {
	panicOnRequest    bool
	deadlineExceeeded bool
}

// Ping implements mwitkow_testproto.TestServiceServer.
func (s *Server) Ping(context.Context, *testproto.PingRequest) (*testproto.PingResponse, error) {
	if s.panicOnRequest {
		panic("boom!")
	}
	if s.deadlineExceeeded {
		return nil, context.DeadlineExceeded
	}
	return &testproto.PingResponse{}, nil
}

// PingEmpty implements mwitkow_testproto.TestServiceServer.
func (*Server) PingEmpty(context.Context, *testproto.Empty) (*testproto.PingResponse, error) {
	panic("unimplemented")
}

// PingError implements mwitkow_testproto.TestServiceServer.
func (*Server) PingError(context.Context, *testproto.PingRequest) (*testproto.Empty, error) {
	panic("unimplemented")
}

// PingList implements mwitkow_testproto.TestServiceServer.
func (*Server) PingList(*testproto.PingRequest, testproto.TestService_PingListServer) error {
	panic("unimplemented")
}

// PingStream implements mwitkow_testproto.TestServiceServer.
func (s *Server) PingStream(out testproto.TestService_PingStreamServer) error {
	if s.panicOnRequest {
		panic("boom!")
	}
	if s.deadlineExceeeded {
		return context.DeadlineExceeded
	}
	return out.Send(&testproto.PingResponse{})
}

var _ testproto.TestServiceServer = &Server{}

func bufDialer(lis *bufconn.Listener) func(context.Context, string) (net.Conn, error) {
	return func(context.Context, string) (net.Conn, error) { return lis.Dial() }
}

func TestGRPCUnary_ContextTimeoutWithDeadlineExceededErr(t *testing.T) {
	ctx := context.Background()
	server, client := grpcSetup(t)
	server.deadlineExceeeded = true

	_, err := client.Ping(ctx, &testproto.PingRequest{})
	assert.ErrorIs(t, err, status.Error(codes.DeadlineExceeded, "context deadline exceeded"))
}

func TestGRPCUnary_RescuePanicsWithStatusInternalError(t *testing.T) {
	ctx := context.Background()
	server, client := grpcSetup(t)
	server.panicOnRequest = true

	_, err := client.Ping(ctx, &testproto.PingRequest{})
	assert.ErrorIs(t, err, status.Error(codes.Internal, "internal error"))
}

func TestGRPCStream_ContextTimeoutWithDeadlineExceededErr(t *testing.T) {
	ctx := context.Background()
	server, client := grpcSetup(t)
	server.deadlineExceeeded = true

	stream, err := client.PingStream(ctx)
	assert.NilError(t, err) // while it looks strange, this is setting up the stream
	_, err = stream.Recv()
	assert.ErrorIs(t, err, status.Error(codes.DeadlineExceeded, "context deadline exceeded"))
}

func TestGRPCStream_RescuePanicsWithStatusInternalError(t *testing.T) {
	ctx := context.Background()
	server, client := grpcSetup(t)
	server.panicOnRequest = true

	stream, err := client.PingStream(ctx)
	assert.NilError(t, err) // while it looks strange, this is setting up the stream

	_, err = stream.Recv()
	assert.ErrorIs(t, err, status.Error(codes.Internal, "internal error"))
}

func TestGRPCUnary_NoRequestError(t *testing.T) {
	ctx := context.Background()
	_, client := grpcSetup(t)

	_, err := client.Ping(ctx, &testproto.PingRequest{})
	assert.NilError(t, err)
}

func TestGRPCStream_NoRequestError(t *testing.T) {
	ctx := context.Background()
	_, client := grpcSetup(t)

	stream, err := client.PingStream(ctx)
	assert.NilError(t, err) // while it looks strange, this is setting up the stream

	_, err = stream.Recv()
	assert.NilError(t, err)
	_, err = stream.Recv()
	assert.Error(t, err, "EOF")
}

func grpcSetup(t *testing.T) (*Server, testproto.TestServiceClient) {
	lis := bufconn.Listen(bufSize)
	middleware := cloudserver.Middleware{Config: cloudserver.Config{Timeout: time.Second * 5}}
	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(middleware.GRPCUnaryServerInterceptor),
		grpc.ChainStreamInterceptor(middleware.GRPCStreamServerInterceptor),
	)
	testServer := &Server{}
	testproto.RegisterTestServiceServer(server, testServer)
	go func() {
		if err := server.Serve(lis); err != nil {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()
	conn, err := grpc.NewClient(
		"passthrough://bufnet",
		grpc.WithContextDialer(bufDialer(lis)),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	assert.NilError(t, err)
	client := testproto.NewTestServiceClient(conn)
	return testServer, client
}
