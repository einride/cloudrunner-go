package cloudzap

import (
	"runtime"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	errorReportContextKey        = "context"
	errorReportServiceContextKey = "serviceContext"
)

// ErrorReportContextForCaller returns a structured logging field for error report context for the provided caller.
func ErrorReportContextForCaller(caller zapcore.EntryCaller) zapcore.Field {
	return ErrorReportContextForSourceLocation(caller.PC, caller.File, caller.Line, caller.Defined)
}

// ErrorReportContextForSourceLocation returns an error report context structured logging field for a source location.
func ErrorReportContextForSourceLocation(pc uintptr, file string, line int, ok bool) zapcore.Field {
	if !ok {
		return zap.Skip()
	}
	return ErrorReportContext(file, line, runtime.FuncForPC(pc).Name())
}

// ErrorReportContext returns a structured logging field for error report context for the provided caller.
func ErrorReportContext(file string, line int, function string) zapcore.Field {
	return zap.Object(errorReportContextKey, errorReportContext{
		reportLocation: errorReportLocation{
			filePath:     file,
			line:         line,
			functionName: function,
		},
	})
}

// ErrorReportServiceContext returns a structured logging field for error report context for the provided caller.
func ErrorReportServiceContext(serviceName, serviceVersion string) zapcore.Field {
	return zap.Object(errorReportServiceContextKey, errorReportServiceContext{
		name:    serviceName,
		version: serviceVersion,
	})
}

type errorReportServiceContext struct {
	name    string
	version string
}

func (s errorReportServiceContext) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	encoder.AddString("name", s.name)
	encoder.AddString("version", s.version)
	return nil
}

type errorReportContext struct {
	reportLocation errorReportLocation
}

func (c errorReportContext) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	return encoder.AddObject("reportLocation", c.reportLocation)
}

type errorReportLocation struct {
	filePath     string
	line         int
	functionName string
}

func (l errorReportLocation) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("filePath", l.filePath)
	enc.AddInt("lineNumber", l.line)
	enc.AddString("functionName", l.functionName)
	return nil
}
