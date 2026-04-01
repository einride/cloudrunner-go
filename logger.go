package cloudrunner

import (
	"context"

	"go.einride.tech/cloudrunner/cloudrequestlog"
	"go.einride.tech/cloudrunner/cloudslog"
)

// WithLoggerFields attaches structured fields to the returned child context.
func WithLoggerFields(ctx context.Context, args ...any) context.Context {
	return cloudslog.With(ctx, args...)
}

// AddRequestLogFields adds fields to the current request log, and is safe to call concurrently.
func AddRequestLogFields(ctx context.Context, args ...any) {
	requestLogFields, ok := cloudrequestlog.GetAdditionalFields(ctx)
	if !ok {
		panic("cloudrunner.AddRequestLogFields must be called with a context from cloudrequestlog.Middleware")
	}
	requestLogFields.Add(args...)
}

// AddRequestLogFieldsToArray appends objects to an array field in the request log and is safe to call concurrently.
func AddRequestLogFieldsToArray(ctx context.Context, key string, objects ...any) {
	additionalFields, ok := cloudrequestlog.GetAdditionalFields(ctx)
	if !ok {
		panic("cloudrunner.AddRequestLogFieldsToArray must be called with a context from cloudrequestlog.Middleware")
	}
	additionalFields.AddToArray(key, objects...)
}
