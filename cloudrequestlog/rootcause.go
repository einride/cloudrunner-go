package cloudrequestlog

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"log/slog"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
)

const (
	// rootCauseKey is the request log field holding the root cause fingerprint of a failed request.
	rootCauseKey = "rootCause"
	// rootCauseTrailerKey is the gRPC response trailer carrying the root cause fingerprint to the caller.
	rootCauseTrailerKey = "cloudrunner-root-cause"
	// rootCauseFingerprintBytes is the number of hash bytes kept in a root cause fingerprint.
	rootCauseFingerprintBytes = 8
)

// rootCauseFingerprint returns an opaque fingerprint identifying a request that failed with the provided code.
// The fingerprint is sent to callers, which may be external, so it must not expose internal details such as method
// names. Equal failures have equal fingerprints, which is all that is needed to correlate them.
func rootCauseFingerprint(kind string, code codes.Code, fullMethod string) string {
	sum := sha256.Sum256([]byte(kind + " " + code.String() + " " + fullMethod))
	return hex.EncodeToString(sum[:rootCauseFingerprintBytes])
}

// isRootCauseFingerprint reports whether s is a well-formed root cause fingerprint. Trailers received from
// downstream services may not be trusted, so anything else is ignored.
func isRootCauseFingerprint(s string) bool {
	if len(s) != hex.EncodedLen(rootCauseFingerprintBytes) {
		return false
	}
	_, err := hex.DecodeString(s)
	return err == nil
}

// clientRootCause resolves the root cause of a failed outgoing gRPC call and records it on the current request, so
// that the server request log and the response trailer to the caller carry the same root cause as the client
// request log. A root cause reported by the downstream service takes precedence. Otherwise the call itself is the
// root cause when it failed at error level. Returns an empty string when the call has no root cause.
func (l *Middleware) clientRootCause(
	ctx context.Context,
	err error,
	level slog.Level,
	code codes.Code,
	fullMethod string,
	trailer metadata.MD,
) string {
	if !l.Config.PropagateRootCause || err == nil {
		return ""
	}
	var rootCause string
	if values := trailer.Get(rootCauseTrailerKey); len(values) > 0 && isRootCauseFingerprint(values[0]) {
		rootCause = values[0]
	}
	if rootCause == "" {
		if level < slog.LevelError {
			return ""
		}
		rootCause = rootCauseFingerprint("gRPCClient", code, fullMethod)
	}
	if additionalFields, ok := GetAdditionalFields(ctx); ok {
		return additionalFields.recordRootCause(rootCause)
	}
	return rootCause
}

// serverRootCause resolves the root cause of a gRPC request that failed at error level: the root cause recorded from
// a failed downstream call, or otherwise the request itself. Returns an empty string when the request has no root
// cause.
func (l *Middleware) serverRootCause(ctx context.Context, level slog.Level, code codes.Code, fullMethod string) string {
	if !l.Config.PropagateRootCause || level < slog.LevelError {
		return ""
	}
	if additionalFields, ok := GetAdditionalFields(ctx); ok {
		if rootCause := additionalFields.getRootCause(); rootCause != "" {
			return rootCause
		}
	}
	return rootCauseFingerprint("gRPCServer", code, fullMethod)
}

// recordedRootCause returns the root cause recorded on the current request, if any.
func (l *Middleware) recordedRootCause(ctx context.Context) string {
	if !l.Config.PropagateRootCause {
		return ""
	}
	if additionalFields, ok := GetAdditionalFields(ctx); ok {
		return additionalFields.getRootCause()
	}
	return ""
}

func rootCauseTrailer(rootCause string) metadata.MD {
	return metadata.Pairs(rootCauseTrailerKey, rootCause)
}

func appendRootCauseAttr(rootCause string, dst []slog.Attr) []slog.Attr {
	if rootCause == "" {
		return dst
	}
	return append(dst, slog.String(rootCauseKey, rootCause))
}
