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
