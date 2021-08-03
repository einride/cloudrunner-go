package cloudrunner

import (
	"context"

	"go.einride.tech/cloudrunner/cloudrequestlog"
	"go.einride.tech/cloudrunner/cloudzap"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger returns the logger for the current context.
func Logger(ctx context.Context) *zap.Logger {
	logger, ok := cloudzap.GetLogger(ctx)
	if !ok {
		panic("cloudrunner.Logger must be called with a context from cloudrunner.Run")
	}
	return logger
}

// WithLoggerFields attaches structured fields to a new logger in the returned child context.
func WithLoggerFields(ctx context.Context, fields ...zap.Field) context.Context {
	logger, ok := cloudzap.GetLogger(ctx)
	if !ok {
		panic("cloudrunner.WithLoggerFields must be called with a context from cloudrunner.Run")
	}
	return cloudzap.WithLogger(ctx, logger.With(fields...))
}

// AddRequestLogFields adds fields to the current request log, and is safe to call concurrently.
func AddRequestLogFields(ctx context.Context, fields ...zap.Field) {
	requestLogFields, ok := cloudrequestlog.GetAdditionalFields(ctx)
	if !ok {
		panic("cloudrunner.AddRequestLogFields must be called with a context from cloudrequestlog.Middleware")
	}
	requestLogFields.Add(fields...)
}

// AddRequestLogFieldsToArray appends objects to an array field in the request log and is safe to call concurrently.
func AddRequestLogFieldsToArray(ctx context.Context, key string, objects ...zapcore.ObjectMarshaler) {
	additionalFields, ok := cloudrequestlog.GetAdditionalFields(ctx)
	if !ok {
		panic("cloudrunner.AddRequestLogFieldsToArray must be called with a context from cloudrequestlog.Middleware")
	}
	additionalFields.AddToArray(key, objects...)
}
