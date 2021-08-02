package cloudzap

import (
	"context"
	"net/http"

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
