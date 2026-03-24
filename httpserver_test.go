package cloudrunner

import (
	"net/http"
	"net/url"
	"testing"

	"gotest.tools/v3/assert"
)

func TestHTTPSpanName(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name     string
		request  *http.Request
		wantName string
	}{
		{
			name: "pattern with method prefix",
			request: &http.Request{
				Method:  http.MethodPost,
				Pattern: "POST /pubsub/events",
				URL:     &url.URL{Path: "/pubsub/events"},
			},
			wantName: "POST /pubsub/events",
		},
		{
			name: "pattern without method prefix",
			request: &http.Request{
				Method:  http.MethodPost,
				Pattern: "/pubsub/events",
				URL:     &url.URL{Path: "/pubsub/events"},
			},
			wantName: "POST /pubsub/events",
		},
		{
			name: "HEAD request matching GET pattern uses actual method",
			request: &http.Request{
				Method:  http.MethodHead,
				Pattern: "GET /healthz",
				URL:     &url.URL{Path: "/healthz"},
			},
			wantName: "HEAD /healthz",
		},
		{
			name: "no pattern falls back to method and path",
			request: &http.Request{
				Method: http.MethodGet,
				URL:    &url.URL{Path: "/healthz"},
			},
			wantName: "GET /healthz",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := httpSpanName("", tt.request)
			assert.Equal(t, tt.wantName, got)
		})
	}
}
