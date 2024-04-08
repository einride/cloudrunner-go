package cloudclient

import (
	"context"
	"fmt"
	"net/url"

	"cloud.google.com/go/compute/metadata"
	"golang.org/x/oauth2"
	"google.golang.org/api/idtoken"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// DialServiceInsecure establishes an insecure connection to another service that must be on the local host.
// Only works outside of GCE, and fails if attempting to dial any other host than localhost.
// Should never be used in production code, only for debugging and local development.
func DialServiceInsecure(ctx context.Context, target string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	if metadata.OnGCE() {
		return nil, fmt.Errorf("dial insecure: forbidden on GCE")
	}
	parsedTarget, err := url.Parse(target)
	if err != nil {
		return nil, fmt.Errorf("dial insecure '%s': %w", target, err)
	}
	if parsedTarget.Hostname() != "localhost" {
		return nil, fmt.Errorf("dial insecure '%s': only allowed for localhost", target)
	}
	const audience = "http://localhost"
	idTokenSource, err := idtoken.NewTokenSource(ctx, audience, option.WithAudiences(audience))
	if err != nil {
		return nil, fmt.Errorf("dial insecure '%s': %w", target, err)
	}
	defaultOpts := []grpc.DialOption{
		grpc.WithPerRPCCredentials(insecureTokenSource{TokenSource: idTokenSource}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	conn, err := grpc.NewClient(parsedTarget.Host, append(defaultOpts, opts...)...)
	if err != nil {
		return nil, fmt.Errorf("dial insecure '%s': %w", target, err)
	}
	return conn, nil
}

type insecureTokenSource struct {
	oauth2.TokenSource
}

func (ts insecureTokenSource) GetRequestMetadata(context.Context, ...string) (map[string]string, error) {
	token, err := ts.Token()
	if err != nil {
		return nil, err
	}
	return map[string]string{"authorization": token.Type() + " " + token.AccessToken}, nil
}

func (insecureTokenSource) RequireTransportSecurity() bool {
	return false
}
