package cloudslog

import (
	"context"

	ltype "google.golang.org/genproto/googleapis/logging/type"
)

type httpRequestContextKey struct{}

// WithHTTPRequest stores an [ltype.HttpRequest] on the context for use in error reporting.
// This should be called by request middleware before handling the request, so that
// error reports logged during request handling can include HTTP request context.
func WithHTTPRequest(parent context.Context, req *ltype.HttpRequest) context.Context {
	return context.WithValue(parent, httpRequestContextKey{}, req)
}

func httpRequestFromContext(ctx context.Context) *ltype.HttpRequest {
	req, _ := ctx.Value(httpRequestContextKey{}).(*ltype.HttpRequest)
	return req
}
