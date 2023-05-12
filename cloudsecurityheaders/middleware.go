package cloudsecurityheaders

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type Middleware struct{}

func (m *Middleware) GRPCUnaryServerInterceptor(
	ctx context.Context,
	request interface{},
	_ *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (resp interface{}, err error) {
	headers := metadata.Pairs(
		"X-Content-Type-Options", "nosniff",
		"Strict-Transport-Security", "max-age=15724800; includeSubDomains",
		"Referrer-Policy", "strict-origin-when-cross-origin",
	)
	// ignore errors here
	_ = grpc.SetHeader(ctx, headers)
	return handler(ctx, request)
}
