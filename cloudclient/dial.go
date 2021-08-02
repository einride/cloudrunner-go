package cloudclient

import (
	"context"
	"crypto/x509"
	"fmt"
	"strconv"
	"strings"

	"google.golang.org/api/idtoken"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/oauth"
)

// DialService dials another Cloud Run gRPC service with the default service account's RPC credentials.
func DialService(ctx context.Context, target string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	audience := "https://" + trimPort(target)
	idTokenSource, err := idtoken.NewTokenSource(ctx, audience, option.WithAudiences(audience))
	if err != nil {
		return nil, fmt.Errorf("dial %s: %w", target, err)
	}
	systemCertPool, err := x509.SystemCertPool()
	if err != nil {
		return nil, fmt.Errorf("dial %s: %w", target, err)
	}
	defaultOpts := []grpc.DialOption{
		grpc.WithPerRPCCredentials(
			&oauth.TokenSource{TokenSource: idTokenSource},
		),
		grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(systemCertPool, "")),
	}
	conn, err := grpc.DialContext(ctx, withDefaultPort(target, 443), append(defaultOpts, opts...)...)
	if err != nil {
		return nil, fmt.Errorf("dial %s: %w", target, err)
	}
	return conn, nil
}

func trimPort(target string) string {
	parts := strings.Split(target, ":")
	if len(parts) == 1 {
		return target
	}
	return strings.Join(parts[:len(parts)-1], ":")
}

func withDefaultPort(target string, port int) string {
	parts := strings.Split(target, ":")
	if len(parts) == 1 {
		return target + ":" + strconv.Itoa(port)
	}
	return target
}
