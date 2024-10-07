package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"go.einride.tech/cloudrunner"
)

func main() {
	if err := cloudrunner.Run(func(ctx context.Context) error {
		cloudrunner.Logger(ctx).Info("hello world")
		httpServer := cloudrunner.NewHTTPServer(ctx, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			slog.InfoContext(ctx, "hello from handler")
			cloudrunner.AddRequestLogFields(r.Context(), "foo", "bar")
			w.Header().Set("Content-Type", "text/plain")
			_, _ = w.Write([]byte("hello world"))
		}))
		return cloudrunner.ListenHTTP(ctx, httpServer)
	}); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
