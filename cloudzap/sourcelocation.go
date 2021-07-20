package cloudzap

import (
	"runtime"
	"strconv"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const sourceLocationKey = "logging.googleapis.com/sourceLocation"

// SourceLocationForCaller returns a structured logging field for the source location of the provided caller.
func SourceLocationForCaller(caller zapcore.EntryCaller) zapcore.Field {
	return SourceLocation(caller.PC, caller.File, caller.Line, caller.Defined)
}

// SourceLocation returns a structured logging field for the provided source location.
func SourceLocation(pc uintptr, file string, line int, ok bool) zapcore.Field {
	if !ok {
		return zap.Skip()
	}
	return zap.Object(sourceLocationKey, sourceLocation{
		file:     file,
		line:     line,
		function: runtime.FuncForPC(pc).Name(),
	})
}

type sourceLocation struct {
	file     string
	line     int
	function string
}

func (s sourceLocation) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	encoder.AddString("file", s.file)
	encoder.AddString("line", strconv.Itoa(s.line))
	encoder.AddString("function", s.function)
	return nil
}
