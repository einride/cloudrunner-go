package cloudclient

import (
	"context"
	"crypto/x509"
	"fmt"
	"strconv"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"google.golang.org/api/idtoken"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/oauth"
	"google.golang.org/grpc/keepalive"
)

// DialService dials another Cloud Run gRPC service with the default service account's RPC credentials.
func DialService(ctx context.Context, target string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	tokenSource, err := newTokenSource(ctx, target)
	if err != nil {
		return nil, err
	}
	systemCertPool, err := x509.SystemCertPool()
	if err != nil {
		return nil, fmt.Errorf("dial %s: %w", target, err)
	}
	defaultOpts := []grpc.DialOption{
		grpc.WithPerRPCCredentials(&oauth.TokenSource{TokenSource: tokenSource}),
		grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(systemCertPool, "")),
		// Enable connection keepalive to mitigate "connection reset by peer".
		// https://cloud.google.com/run/docs/troubleshooting
		// For details on keepalive settings, see:
		// https://github.com/grpc/grpc-go/blob/master/Documentation/keepalive.md
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                1 * time.Minute,
			Timeout:             10 * time.Second,
			PermitWithoutStream: true,
		}),
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

func newTokenSource(ctx context.Context, target string) (_ oauth2.TokenSource, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("new token source: %w", err)
		}
	}()
	audience := "https://" + trimPort(target)
	idTokenSource, err := idtoken.NewTokenSource(ctx, audience, option.WithAudiences(audience))
	return idTokenSource, err
}
