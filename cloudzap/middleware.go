package cloudzap

import (
	"context"
	"net/http"

	"go.einride.tech/cloudrunner/cloudstream"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type Middleware struct {
	Logger *zap.Logger
}

// HTTPServer implements HTTP server middleware to add a logger to the request context.
func (l *Middleware) HTTPServer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		next.ServeHTTP(writer, request.WithContext(WithLogger(request.Context(), l.Logger)))
	})
}

// GRPCUnaryServerInterceptor implements grpc.UnaryServerInterceptor to add a logger to the request context.
func (l *Middleware) GRPCUnaryServerInterceptor(
	ctx context.Context,
	request interface{},
	_ *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	return handler(WithLogger(ctx, l.Logger), request)
}

// GRPCStreamServerInterceptor adds a zap logger to the server stream context.
func (l *Middleware) GRPCStreamServerInterceptor(
	srv interface{},
	ss grpc.ServerStream,
	_ *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) (err error) {
	return handler(srv, cloudstream.NewContextualServerStream(WithLogger(ss.Context(), l.Logger), ss))
}
