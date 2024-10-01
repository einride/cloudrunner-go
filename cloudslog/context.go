package cloudslog

import (
	"context"
	"log/slog"
	"slices"
)

type contextKey struct{}

// With appends log attributes to the current parent context.
// Arguments are converted to attributes as if by [slog.Logger.Log].
func With(parent context.Context, args ...any) context.Context {
	if v, ok := parent.Value(contextKey{}).([]slog.Attr); ok {
		return context.WithValue(parent, contextKey{}, append(slices.Clip(v), argsToAttrSlice(args)...))
	}
	return context.WithValue(parent, contextKey{}, argsToAttrSlice(args))
}

func attributesFromContext(ctx context.Context) []slog.Attr {
	if v, ok := ctx.Value(contextKey{}).([]slog.Attr); ok {
		return v
	}
	return nil
}

// argsToAttrSlice is copied from the slog stdlib.
func argsToAttrSlice(args []any) []slog.Attr {
	var attr slog.Attr
	attrs := make([]slog.Attr, 0, len(args))
	for len(args) > 0 {
		attr, args = argsToAttr(args)
		attrs = append(attrs, attr)
	}
	return attrs
}

// argsToAttr is copied from the slog stdlib.
func argsToAttr(args []any) (slog.Attr, []any) {
	const badKey = "!BADKEY"
	switch x := args[0].(type) {
	case string:
		if len(args) == 1 {
			return slog.String(badKey, x), nil
		}
		return slog.Any(x, args[1]), args[2:]
	case slog.Attr:
		return x, args[1:]
	default:
		return slog.Any(badKey, x), args[1:]
	}
}
