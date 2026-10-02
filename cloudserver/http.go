package cloudserver

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"
	"slices"

	"go.einride.tech/cloudrunner/cloudrequestlog"
)

// HTTPServer provides HTTP server middleware.
func (i *Middleware) HTTPServer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				writer.WriteHeader(http.StatusInternalServerError)
				if fields, ok := cloudrequestlog.GetAdditionalFields(request.Context()); ok {
					fields.Add(
						slog.String("stack", string(debug.Stack())),
						slog.Any("error", fmt.Errorf("recovered panic: %v", r)),
					)
				}
			}
		}()
		if i.Config.Timeout <= 0 {
			next.ServeHTTP(writer, request)
			return
		}
		ctx, cancel := context.WithTimeout(request.Context(), i.Config.Timeout)
		defer cancel()
		next.ServeHTTP(writer, request.WithContext(ctx))
	})
}

// HTTPMiddleware is a HTTP middleware.
type HTTPMiddleware = func(http.Handler) http.Handler

// ChainHTTPMiddleware chains the HTTP handler middleware to execute from left to right.
func ChainHTTPMiddleware(next http.Handler, middlewares ...HTTPMiddleware) http.Handler {
	if len(middlewares) == 0 {
		return next
	}
	wrapped := next
	// loop in reverse to preserve middleware order
	for _, middleware := range slices.Backward(middlewares) {
		wrapped = middleware(wrapped)
	}
	return wrapped
}
