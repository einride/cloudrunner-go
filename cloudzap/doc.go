// Package cloudzap provides primitives for structured logging with go.uber.org/zap.
//
// Deprecated: Use log/slog with the cloudslog package instead. The cloudslog.Handler
// automatically handles trace correlation, error reporting, and Cloud Logging field
// formatting without requiring explicit middleware or field injection.
package cloudzap
