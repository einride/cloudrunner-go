package cloudotel

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"go.opentelemetry.io/otel/propagation"
	"gotest.tools/v3/assert"
)

func TestPropagatePubSubTracing(t *testing.T) {
	t.Run("valid pubsub payload", func(t *testing.T) {
		// arrange
		ctx := context.Background()
		payload := `{
			"subscription": "projects/test-project/subscriptions/test-sub",
			"message": {
				"attributes": {
					"googclient_traceparent": "00-161d8104e1e09a3e3a4d80129acbfe30-75a8fe4fee0c65b8-00"
				},
				"data": "data",
				"messageId": "12345",
				"publishTime": "2025-03-24T12:34:56Z"
			},
			"deliveryAttempt": "5"
		}`
		req := &http.Request{
			Method: http.MethodPost,
			Body:   newReadCloser(payload),
		}

		// act
		ctx = propagatePubsubTracing(ctx, req)

		// assert
		actualTraceContext := extractTraceContext(ctx)
		expectedTraceContext := `{"traceparent":"00-161d8104e1e09a3e3a4d80129acbfe30-75a8fe4fee0c65b8-00"}`
		assert.Equal(t, expectedTraceContext, actualTraceContext)

		actualPayload, err := io.ReadAll(req.Body)
		assert.NilError(t, err)
		assert.DeepEqual(t, string(actualPayload), payload)
	})

	t.Run("non pubsub payload", func(t *testing.T) {
		// arrange
		ctx := context.Background()
		payload := `{"user": "test-user", "action": "login"}`
		req := &http.Request{
			Method: http.MethodPost,
			Body:   newReadCloser(payload),
		}

		// act
		ctx = propagatePubsubTracing(ctx, req)

		// assert
		actualTraceContext := extractTraceContext(ctx)
		assert.Equal(t, "{}", actualTraceContext)

		actualPayload, err := io.ReadAll(req.Body)
		assert.NilError(t, err)
		assert.DeepEqual(t, string(actualPayload), payload)
	})

	t.Run("invalid payload", func(t *testing.T) {
		// arrange
		ctx := context.Background()
		payload := `"message":{"messageId":12345,data":"data","attributes":{"key""value"}}`
		req := &http.Request{
			Method: http.MethodPost,
			Body:   newReadCloser(payload),
		}

		// act
		ctx = propagatePubsubTracing(ctx, req)

		// assert
		actualTraceContext := extractTraceContext(ctx)
		assert.Equal(t, "{}", actualTraceContext)

		actualPayload, err := io.ReadAll(req.Body)
		assert.NilError(t, err)
		assert.DeepEqual(t, string(actualPayload), payload)
	})
}

func TestPubsubTraceExtractor(t *testing.T) {
	t.Parallel()
	pubsubPayload := `{
		"subscription": "projects/test-project/subscriptions/test-sub",
		"message": {
			"attributes": {
				"googclient_traceparent": "00-161d8104e1e09a3e3a4d80129acbfe30-75a8fe4fee0c65b8-00"
			},
			"data": "data",
			"messageId": "12345",
			"publishTime": "2025-03-24T12:34:56Z"
		}
	}`
	nonPubsubPayload := `{"action": "login"}`

	tests := []struct {
		name                  string
		enablePubsubTracing   bool
		body                  string
		wantTraceContext      string
		wantTraceparentHeader string
		wantBodyPassedThrough bool
	}{
		{
			name:                  "pubsub message extracts trace context and sets traceparent header",
			enablePubsubTracing:   true,
			body:                  pubsubPayload,
			wantTraceContext:      `{"traceparent":"00-161d8104e1e09a3e3a4d80129acbfe30-75a8fe4fee0c65b8-00"}`,
			wantTraceparentHeader: "00-161d8104e1e09a3e3a4d80129acbfe30-75a8fe4fee0c65b8-00",
		},
		{
			name:                  "no-op when disabled",
			enablePubsubTracing:   false,
			body:                  pubsubPayload,
			wantTraceContext:      "{}",
			wantTraceparentHeader: "",
		},
		{
			name:                  "non-pubsub request passes through unchanged",
			enablePubsubTracing:   true,
			body:                  nonPubsubPayload,
			wantTraceContext:      "{}",
			wantTraceparentHeader: "",
			wantBodyPassedThrough: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			// arrange
			middleware := TraceMiddleware{EnablePubsubTracing: tc.enablePubsubTracing}
			var gotTraceContext, gotTraceparentHeader, gotBody string
			inner := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
				gotTraceContext = extractTraceContext(r.Context())
				gotTraceparentHeader = r.Header.Get("Traceparent")
				if tc.wantBodyPassedThrough {
					b, _ := io.ReadAll(r.Body)
					gotBody = string(b)
				}
			})
			handler := middleware.PubsubTraceExtractor(inner)
			req, err := http.NewRequestWithContext(
				context.Background(),
				http.MethodPost,
				"/pubsub/handler",
				newReadCloser(tc.body),
			)
			assert.NilError(t, err)

			// act
			handler.ServeHTTP(nil, req)

			// assert
			assert.Equal(t, tc.wantTraceContext, gotTraceContext)
			assert.Equal(t, tc.wantTraceparentHeader, gotTraceparentHeader)
			if tc.wantBodyPassedThrough {
				assert.Equal(t, tc.body, gotBody)
			}
		})
	}
}

func extractTraceContext(ctx context.Context) string {
	propagator := propagation.TraceContext{}
	carrier := make(propagation.MapCarrier)
	propagator.Inject(ctx, carrier)
	data, err := json.Marshal(carrier)
	if err != nil {
		return ""
	}
	return string(data)
}

func newReadCloser(content string) io.ReadCloser {
	return io.NopCloser(bytes.NewReader([]byte(content)))
}
