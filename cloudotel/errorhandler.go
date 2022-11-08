package cloudotel

import (
	"context"

	"go.einride.tech/cloudrunner/cloudzap"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewErrorLogger returns a new otel.ErrorHandler that logs errors using the provided logger, level and message.
func NewErrorLogger(logger *zap.Logger, level zapcore.Level, message string) otel.ErrorHandler {
	return errorHandler{logger: logger, level: level, message: message}
}

type errorHandler struct {
	logger  *zap.Logger
	level   zapcore.Level
	message string
}

// Handle implements otel.ErrorHandler.
func (e errorHandler) Handle(err error) {
	e.logger.Check(e.level, e.message).Write(zap.Error(err))
}

// RegisterErrorHandler registers a global OpenTelemetry error handler.
func RegisterErrorHandler(ctx context.Context) {
	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(err error) {
		handleError(ctx, err)
	}))
}

func handleError(ctx context.Context, err error) {
	if isUnsupportedSamplerErr(err) {
		// The OpenCensus bridge does not support all features from OpenCensus,
		// for example custom samplers which is used in some libraries.
		// The bridge presumably falls back to the configured sampler, so
		// this error can be ignored.
		//
		// See
		// https://pkg.go.dev/go.opentelemetry.io/otel/bridge/opencensus
		return
	}
	if logger, ok := cloudzap.GetLogger(ctx); ok {
		logger.Warn("otel error", zap.Error(err))
	}
}
