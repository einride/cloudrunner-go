package cloudzap

import (
	"fmt"

	"go.einride.tech/cloudrunner/cloudruntime"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// LoggerConfig configures the application logger.
type LoggerConfig struct {
	// Development indicates if the logger should output human-readable output for development.
	Development bool `default:"true" onGCE:"false"`
	// Level indicates which log level the logger should output at.
	Level zapcore.Level `default:"debug" onGCE:"info"`
	// ReportErrors indicates if error reports should be logged for errors.
	ReportErrors bool `onGCE:"true"`
}

// NewLogger creates a new Logger.
func NewLogger(config LoggerConfig) (*zap.Logger, error) {
	if config.Development {
		zapConfig := zap.NewDevelopmentConfig()
		zapConfig.EncoderConfig.EncodeLevel = zapcore.LowercaseColorLevelEncoder
		zapConfig.Level = zap.NewAtomicLevelAt(config.Level)
		return zapConfig.Build(
			zap.AddCaller(),
			zap.AddStacktrace(zap.FatalLevel), // add stacktraces manually where needed
		)
	}
	zapConfig := zap.NewProductionConfig()
	zapConfig.EncoderConfig = NewEncoderConfig()
	zapConfig.Level = zap.NewAtomicLevelAt(config.Level)
	zapOptions := []zap.Option{
		zap.AddCaller(),
		zap.AddStacktrace(zap.FatalLevel), // add stacktraces manually where needed
		zap.WrapCore(func(core zapcore.Core) zapcore.Core {
			return sourceLocationCore{Core: core}
		}),
	}
	if config.ReportErrors {
		if service, ok := cloudruntime.Service(); ok {
			if serviceVersion, ok := cloudruntime.ServiceVersion(); ok {
				zapOptions = append(zapOptions, zap.WrapCore(func(core zapcore.Core) zapcore.Core {
					return errorReportingCore{
						Core:           core,
						serviceName:    service,
						serviceVersion: serviceVersion,
					}
				}))
			}
		}
	}
	logger, err := zapConfig.Build(zapOptions...)
	if err != nil {
		return nil, fmt.Errorf("init logger: %w", err)
	}
	return logger, nil
}

type sourceLocationCore struct {
	zapcore.Core
}

// Check implements zapcore.Core.
func (c sourceLocationCore) Check(entry zapcore.Entry, checked *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if !c.Enabled(entry.Level) {
		return checked
	}
	return checked.AddCore(entry, c)
}

// Write implements zapcore.Core.
func (c sourceLocationCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	if entry.Caller.Defined {
		fields = appendIfNotExists(fields, SourceLocationForCaller(entry.Caller))
	}
	return c.Core.Write(entry, fields)
}

type errorReportingCore struct {
	zapcore.Core
	serviceName    string
	serviceVersion string
}

// Check implements zapcore.Core.
func (c errorReportingCore) Check(entry zapcore.Entry, checked *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if !c.Enabled(entry.Level) {
		return checked
	}
	return checked.AddCore(entry, c)
}

// Write implements zapcore.Core.
func (c errorReportingCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	if entry.Caller.Defined && zap.ErrorLevel.Enabled(entry.Level) {
		fields = appendIfNotExists(fields, ErrorReportContextForCaller(entry.Caller))
		fields = appendIfNotExists(fields, ErrorReportServiceContext(c.serviceName, c.serviceVersion))
	}
	return c.Core.Write(entry, fields)
}

func appendIfNotExists(fields []zapcore.Field, field zap.Field) []zapcore.Field {
	for _, existing := range fields {
		if existing.Key == field.Key {
			return fields
		}
	}
	return append(fields, field)
}
