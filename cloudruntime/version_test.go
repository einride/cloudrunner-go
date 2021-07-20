package cloudruntime

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestServiceVersion(t *testing.T) {
	t.Run("SERVICE_VERSION highest priority", func(t *testing.T) {
		setEnv(t, "SERVICE_VERSION", "foo")
		setServiceVersion(t, "bar")
		setEnv(t, "K_REVISION", "baz")
		actual, ok := ServiceVersion()
		assert.Assert(t, ok)
		assert.Equal(t, "foo", actual)
	})

	t.Run("global variable second highest priority", func(t *testing.T) {
		setServiceVersion(t, "bar")
		setEnv(t, "K_REVISION", "baz")
		actual, ok := ServiceVersion()
		assert.Assert(t, ok)
		assert.Equal(t, "bar", actual)
	})

	t.Run("K_REVISION third highest priority", func(t *testing.T) {
		setEnv(t, "K_REVISION", "baz")
		actual, ok := ServiceVersion()
		assert.Assert(t, ok)
		assert.Equal(t, "baz", actual)
	})

	t.Run("no version", func(t *testing.T) {
		actual, ok := ServiceVersion()
		assert.Assert(t, !ok)
		assert.Equal(t, "", actual)
	})
}

func setServiceVersion(t *testing.T, value string) {
	prev := serviceVersion
	serviceVersion = value
	t.Cleanup(func() {
		serviceVersion = prev
	})
}
