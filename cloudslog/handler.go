package cloudslog

import (
	"context"
	"io"
	"log/slog"
	"os"
	"runtime/debug"

	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/trace"
	ltype "google.golang.org/genproto/googleapis/logging/type"
	"google.golang.org/protobuf/proto"
)

// LoggerConfig configures the application logger.
type LoggerConfig struct {
	// Development indicates if the logger should output human-readable output for development.
	Development bool `default:"true" onGCE:"false"`
	// Level indicates which log level the logger should output at.
	Level slog.Level `default:"debug" onGCE:"info"`
	// MessageSizeLimit is the maximum size, in bytes, of requests and responses to log.
	// Messages large than the limit will be truncated.
	// Default value, 0, means that no messages will be truncated.
	MessageSizeLimit int `onGCE:"1024"`
}

func NewHandler(config LoggerConfig) slog.Handler {
	return newHandler(os.Stdout, config)
}

func newHandler(w io.Writer, config LoggerConfig) slog.Handler {
	var result slog.Handler
	if config.Development {
		result = slog.NewTextHandler(w, &slog.HandlerOptions{
			Level:       config.Level,
			ReplaceAttr: replaceAttr,
		})
	} else {
		result = slog.NewJSONHandler(w, &slog.HandlerOptions{
			AddSource:   true,
			Level:       config.Level,
			ReplaceAttr: replaceAttr,
		})
	}
	result = &handler{Handler: result}
	return result
}

type handler struct {
	slog.Handler
}

var _ slog.Handler = &handler{}

// Handle adds attributes from the span context to the [slog.Record].
func (t *handler) Handle(ctx context.Context, record slog.Record) error {
	if s := trace.SpanContextFromContext(ctx); s.IsValid() {
		// See: https://cloud.google.com/logging/docs/structured-logging#special-payload-fields
		record.AddAttrs(slog.Any("logging.googleapis.com/trace", s.TraceID()))
		record.AddAttrs(slog.Any("logging.googleapis.com/spanId", s.SpanID()))
		record.AddAttrs(slog.Bool("logging.googleapis.com/trace_sampled", s.TraceFlags().IsSampled()))
	}
	if value, ok := ctx.Value(contextKey{}).(*contextValue); ok {
		value.mu.Lock()
		record.AddAttrs(value.attrs...)
		value.mu.Unlock()
	}
	return t.Handler.Handle(ctx, record)
}

func replaceAttr(_ []string, attr slog.Attr) slog.Attr {
	switch attr.Key {
	case slog.LevelKey:
		// See: https://cloud.google.com/logging/docs/reference/v2/rest/v2/LogEntry#LogSeverity
		attr.Key = "severity"
		if level := attr.Value.Any().(slog.Level); level == slog.LevelWarn {
			attr.Value = slog.StringValue("WARNING")
		}
	case slog.TimeKey:
		attr.Key = "timestamp"
	case slog.MessageKey:
		attr.Key = "message"
	case slog.SourceKey:
		attr.Key = "logging.googleapis.com/sourceLocation"
	default:
		if attr.Value.Kind() == slog.KindAny {
			switch value := attr.Value.Any().(type) {
			case *resource.Resource:
				attr.Value = slog.AnyValue(newResourceValue(value))
			case *debug.BuildInfo:
				attr.Value = slog.AnyValue(newBuildInfoValue(value))
			case *ltype.HttpRequest:
				attr.Value = slog.AnyValue(newProtoValue(fixHTTPRequest(value)))
			case proto.Message:
				attr.Value = slog.AnyValue(newProtoValue(value))
			}
		}
	}
	return attr
}
