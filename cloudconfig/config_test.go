package cloudconfig

import (
	"os"
	"path/filepath"
	"testing"

	"gotest.tools/v3/assert"
)

type loadTestConfig struct {
	Topic string
	Count int
}

func TestConfig_Load_yamlServiceSpecificationFile(t *testing.T) {
	spec := filepath.Join(t.TempDir(), "service.yaml")
	assert.NilError(t, os.WriteFile(spec, []byte(`apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: test-service
spec:
  template:
    spec:
      containers:
        - image: test
          env:
            - name: TOPIC
              value: from-file
            - name: COUNT
              value: "3"
`), 0o600))

	t.Run("sets the file's env in the process environment", func(t *testing.T) {
		t.Setenv("TOPIC", "from-process")
		unsetEnv(t, "COUNT")
		unsetEnv(t, "K_SERVICE")
		var config loadTestConfig
		c, err := New("test", &config, WithYAMLServiceSpecificationFile(spec))
		assert.NilError(t, err)
		assert.NilError(t, c.Load())
		assert.Equal(t, config, loadTestConfig{Topic: "from-process", Count: 3})
		assert.Equal(t, os.Getenv("COUNT"), "3")
		assert.Equal(t, os.Getenv("K_SERVICE"), "test-service")
	})

	t.Run("layers the file's env under a custom lookup", func(t *testing.T) {
		unsetEnv(t, "COUNT")
		var config loadTestConfig
		c, err := New(
			"test",
			&config,
			WithYAMLServiceSpecificationFile(spec),
			WithLookupEnv(func(key string) (string, bool) {
				if key == "TOPIC" {
					return "from-lookup", true
				}
				return "", false
			}),
		)
		assert.NilError(t, err)
		assert.NilError(t, c.Load())
		assert.Equal(t, config, loadTestConfig{Topic: "from-lookup", Count: 3})
		_, ok := os.LookupEnv("COUNT")
		assert.Assert(t, !ok, "COUNT set in the process environment")
	})
}

// unsetEnv unsets key for the duration of the test.
func unsetEnv(t *testing.T, key string) {
	t.Helper()
	t.Setenv(key, "") // Restores the original value when the test ends.
	assert.NilError(t, os.Unsetenv(key))
}
