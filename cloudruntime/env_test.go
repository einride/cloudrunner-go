package cloudruntime

import (
	"os"
	"strconv"
	"testing"

	"gotest.tools/v3/assert"
)

func TestService(t *testing.T) {
	t.Run("from env", func(t *testing.T) {
		const expected = "foo"
		setEnv(t, "K_SERVICE", expected)
		actual, ok := Service()
		assert.Assert(t, ok)
		assert.Equal(t, expected, actual)
	})

	t.Run("defaults to name of binary", func(t *testing.T) {
		actual, ok := Service()
		assert.Assert(t, ok)
		assert.Assert(t, actual != "") // don't assume the name of the test binary
	})
}

func TestRevision(t *testing.T) {
	t.Run("from env", func(t *testing.T) {
		const expected = "foo"
		setEnv(t, "K_REVISION", expected)
		actual, ok := Revision()
		assert.Assert(t, ok)
		assert.Equal(t, expected, actual)
	})

	t.Run("undefined", func(t *testing.T) {
		actual, ok := Revision()
		assert.Assert(t, !ok)
		assert.Equal(t, "", actual)
	})
}

func TestConfiguration(t *testing.T) {
	t.Run("from env", func(t *testing.T) {
		const expected = "foo"
		setEnv(t, "K_CONFIGURATION", expected)
		actual, ok := Configuration()
		assert.Assert(t, ok)
		assert.Equal(t, expected, actual)
	})

	t.Run("undefined", func(t *testing.T) {
		actual, ok := Configuration()
		assert.Assert(t, !ok)
		assert.Equal(t, "", actual)
	})
}

func TestPort(t *testing.T) {
	t.Run("from env", func(t *testing.T) {
		const expected = 42
		setEnv(t, "PORT", strconv.Itoa(expected))
		actual, ok := Port()
		assert.Assert(t, ok)
		assert.Equal(t, expected, actual)
	})

	t.Run("invalid", func(t *testing.T) {
		setEnv(t, "PORT", "invalid")
		actual, ok := Port()
		assert.Assert(t, !ok)
		assert.Equal(t, 0, actual)
	})

	t.Run("undefined", func(t *testing.T) {
		actual, ok := Port()
		assert.Assert(t, !ok)
		assert.Equal(t, 0, actual)
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
