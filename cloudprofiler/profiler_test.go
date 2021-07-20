package cloudprofiler

import (
	"os"
	"testing"

	"cloud.google.com/go/profiler"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/api/option"
	"gotest.tools/v3/assert"
)

func TestStart(t *testing.T) {
	input := Config{
		Enabled:        true,
		MutexProfiling: true,
		AllocForceGC:   true,
	}
	expected := profiler.Config{
		ProjectID:      "project",
		Service:        "service",
		ServiceVersion: "version",
		MutexProfiling: true,
		AllocForceGC:   true,
	}
	setEnv(t, "GOOGLE_CLOUD_PROJECT", "project")
	setEnv(t, "K_SERVICE", "service")
	setEnv(t, "SERVICE_VERSION", "version")
	var actual profiler.Config
	withProfilerStart(t, func(config profiler.Config, _ ...option.ClientOption) error {
		actual = config
		return nil
	})
	err := Start(input)
	assert.NilError(t, err)
	assert.DeepEqual(t, expected, actual, cmpopts.IgnoreUnexported(profiler.Config{}))
}

func withProfilerStart(t *testing.T, fn func(profiler.Config, ...option.ClientOption) error) {
	prev := profilerStart
	profilerStart = fn
	t.Cleanup(func() {
		profilerStart = prev
	})
}

// setEnv will be available in the standard library from Go 1.17 as t.SetEnv.
func setEnv(t *testing.T, key, value string) {
	prevValue, ok := os.LookupEnv(key)
	if err := os.Setenv(key, value); err != nil {
		t.Fatalf("cannot set environment variable: %v", err)
	}
	if ok {
		t.Cleanup(func() {
			_ = os.Setenv(key, prevValue)
		})
	} else {
		t.Cleanup(func() {
			_ = os.Unsetenv(key)
		})
	}
}
