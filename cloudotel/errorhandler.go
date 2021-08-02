package cloudotel

import (
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
