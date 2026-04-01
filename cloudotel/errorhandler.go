package cloudotel

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel"
)

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
	slog.WarnContext(ctx, "otel error", slog.Any("error", err))
}
