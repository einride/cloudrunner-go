package cloudclient

import (
	"testing"
	"time"

	"google.golang.org/grpc/codes"
	"gotest.tools/v3/assert"
)

func TestClientConfig_AsServiceConfigJSON(t *testing.T) {
	input := Config{
		Timeout: 5 * time.Second,
		Retry: RetryConfig{
			Enabled:           true,
			InitialBackoff:    200 * time.Millisecond,
			MaxBackoff:        3 * time.Second,
			MaxAttempts:       5,
			BackoffMultiplier: 2,
			RetryableStatusCodes: []codes.Code{
				codes.Unavailable,
				codes.Unknown,
			},
		},
	}
	const expected = `{"methodConfig":[{"name":[{"service":"","method":""}],"timeout":"5s",` +
		`"retryPolicy":{"maxAttempts":5,"maxBackoff":"3s","initialBackoff":"0.2s","backoffMultiplier":2,` +
		`"retryableStatusCodes":["UNAVAILABLE","UNKNOWN"]}}]}`
	assert.Equal(t, expected, input.AsServiceConfigJSON())
}
