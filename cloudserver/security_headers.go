package cloudserver

import (
	"net/http"
)

// SecurityHeadersMiddleware adds security headers to responses.
type SecurityHeadersMiddleware struct{}

// HTTPServer provides HTTP server middleware.
func (i *SecurityHeadersMiddleware) HTTPServer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		next.ServeHTTP(w, r)
	})
}
