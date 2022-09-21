package cloudclient

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"go.einride.tech/cloudrunner/cloudzap"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
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
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                30 * time.Second,
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
	idTokenSource, errIDTokenSource := idtoken.NewTokenSource(ctx, audience, option.WithAudiences(audience))
	if errIDTokenSource == nil {
		return idTokenSource, nil
	}
	// Google's idtoken package does not support credential type other than `service_account`.
	// This blocks local development with using `impersonated_service_account` type credentials. If that happens,
	// we work it around by using our Application Default Credentials (which is impersonated already) to fetch
	// an id_token on the fly.
	// This however still blocks `authorized_user` type of credentials passing through.
	// Related issue page: https://github.com/googleapis/google-api-go-client/issues/873
	defaultCredentials, err := google.FindDefaultCredentials(ctx)
	if err != nil {
		return nil, err
	}
	var defaultCredentialsJSON struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(defaultCredentials.JSON, &defaultCredentialsJSON); err != nil {
		return nil, err
	}
	if defaultCredentialsJSON.Type != "impersonated_service_account" {
		// We only patch the case where type of "impersonated_service_account" is used
		// if not return original error.
		return nil, err
	}
	defaultTokenSource, err := google.DefaultTokenSource(ctx)
	if err != nil {
		return nil, err
	}
	if logger, ok := cloudzap.GetLogger(ctx); ok {
		logger.Warn("using default token source - this should not happen in production", zap.Error(errIDTokenSource))
	}
	return oauth2.ReuseTokenSource(nil, &idTokenSourceWrapper{TokenSource: defaultTokenSource}), nil
}

// idTokenSourceWrapper is an oauth2.TokenSource wrapper used for getting id_token for local development using
// `authorized_user` type credentials
// It takes the id_token from TokenSource and passes that on as a bearer token.
type idTokenSourceWrapper struct {
	TokenSource oauth2.TokenSource
}

func (s *idTokenSourceWrapper) Token() (*oauth2.Token, error) {
	token, err := s.TokenSource.Token()
	if err != nil {
		return nil, err
	}
	idToken, ok := token.Extra("id_token").(string)
	if !ok {
		return nil, fmt.Errorf("token did not contain an id_token")
	}
	return &oauth2.Token{
		AccessToken: idToken,
		TokenType:   "Bearer",
		Expiry:      token.Expiry,
	}, nil
}
