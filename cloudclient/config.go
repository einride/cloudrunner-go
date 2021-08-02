package cloudclient

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"google.golang.org/grpc/codes"
)

// Config configures a gRPC client's default timeout and retry behavior.
// See: https://github.com/grpc/grpc-proto/blob/master/grpc/service_config/service_config.proto
type Config struct {
	// The timeout of outgoing gRPC method calls. Set to zero to disable.
	Timeout time.Duration `default:"10s"`
	// Retry config.
	Retry RetryConfig
}

// RetryConfig configures default retry behavior for outgoing gRPC client calls.
// See: https://github.com/grpc/grpc-proto/blob/master/grpc/service_config/service_config.proto
type RetryConfig struct {
	// Enabled indicates if retries are enabled.
	Enabled bool `default:"true"`
	// InitialBackoff is the initial exponential backoff duration.
	//
	// The initial retry attempt will occur at:
	//   random(0, initial_backoff).
	//
	// In general, the nth attempt will occur at:
	//   random(0, min(initial_backoff*backoff_multiplier**(n-1), max_backoff)).
	//
	// Must be greater than zero.
	InitialBackoff time.Duration `default:"200ms"`
	// MaxBackoff is the maximum duration between retries.
	MaxBackoff time.Duration `default:"60s"`
	// MaxAttempts is the max number of backoff attempts retried.
	MaxAttempts int `default:"5"`
	// BackoffMultiplier is the exponential backoff multiplier.
	BackoffMultiplier float64 `default:"1.3"`
	// RetryableStatusCodes is the set of status codes which may be retried.
	RetryableStatusCodes []codes.Code `default:"Unavailable"`
}

// AsServiceConfigJSON returns the default method call config as a valid gRPC service JSON config.
func (c *Config) AsServiceConfigJSON() string {
	type methodNameJSON struct {
		Service string `json:"service"`
		Method  string `json:"method"`
	}
	type retryPolicyJSON struct {
		MaxAttempts          int      `json:"maxAttempts"`
		MaxBackoff           string   `json:"maxBackoff"`
		InitialBackoff       string   `json:"initialBackoff"`
		BackoffMultiplier    float64  `json:"backoffMultiplier"`
		RetryableStatusCodes []string `json:"retryableStatusCodes"`
	}
	type methodConfigJSON struct {
		Name        []methodNameJSON `json:"name"`
		Timeout     *string          `json:"timeout,omitempty"`
		RetryPolicy *retryPolicyJSON `json:"retryPolicy,omitempty"`
	}
	type serviceConfigJSON struct {
		MethodConfig []methodConfigJSON `json:"methodConfig"`
	}
	var s strings.Builder
	methodConfig := methodConfigJSON{
		Name: []methodNameJSON{
			{}, // no service or method specified means the config applies to all methods, all services.
		},
	}
	if c.Timeout > 0 {
		methodConfig.Timeout = new(string)
		*methodConfig.Timeout = fmt.Sprintf("%gs", c.Timeout.Seconds())
	}
	if c.Retry.Enabled {
		methodConfig.RetryPolicy = &retryPolicyJSON{
			MaxAttempts:          c.Retry.MaxAttempts,
			InitialBackoff:       fmt.Sprintf("%gs", c.Retry.InitialBackoff.Seconds()),
			MaxBackoff:           fmt.Sprintf("%gs", c.Retry.MaxBackoff.Seconds()),
			BackoffMultiplier:    c.Retry.BackoffMultiplier,
			RetryableStatusCodes: make([]string, 0, len(c.Retry.RetryableStatusCodes)),
		}
		for _, code := range c.Retry.RetryableStatusCodes {
			methodConfig.RetryPolicy.RetryableStatusCodes = append(
				methodConfig.RetryPolicy.RetryableStatusCodes, strings.ToUpper(code.String()),
			)
		}
	}
	_ = json.NewEncoder(&s).Encode(serviceConfigJSON{
		MethodConfig: []methodConfigJSON{methodConfig},
	})
	return strings.TrimSpace(s.String())
}
