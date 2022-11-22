package cloudserver

import (
	"context"
	"fmt"
	"net/http"

	"go.einride.tech/cloudrunner/cloudrequestlog"
	"go.uber.org/zap"
)

// HTTPServer provides HTTP server middleware.
func (i *Middleware) HTTPServer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				writer.WriteHeader(http.StatusInternalServerError)
				if fields, ok := cloudrequestlog.GetAdditionalFields(request.Context()); ok {
					fields.Add(
						zap.Stack("stack"),
						zap.Error(fmt.Errorf("recovered panic: %v", r)),
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
	for i := len(middlewares) - 1; i >= 0; i-- {
		wrapped = middlewares[i](wrapped)
	}
	return wrapped
}
