package cloudrunner_test

import (
	"os"
	"path/filepath"
	"testing"

	"go.einride.tech/cloudrunner"
	"gotest.tools/v3/assert"
)

type validateTestConfig struct {
	Topic  string `required:"true"`
	Count  int
	APIKey string `secret:"true" required:"true"`
}

func TestValidateConfig(t *testing.T) {
	for _, tt := range []struct {
		name    string
		env     string
		wantErr string
	}{
		{
			name: "valid",
			env: `
            - name: TOPIC
              value: events
            - name: COUNT
              value: "3"
            - name: APIKEY
              valueFrom:
                secretKeyRef:
                  name: api-key
                  key: latest`,
		},
		{
			name: "missing required value",
			env: `
            - name: COUNT
              value: "3"`,
			wantErr: "required key TOPIC missing value",
		},
		{
			name: "invalid value",
			env: `
            - name: TOPIC
              value: events
            - name: COUNT
              value: many`,
			wantErr: "converting 'many' to type int",
		},
		{
			name: "invalid cloudrunner value",
			env: `
            - name: TOPIC
              value: events
            - name: SERVER_TIMEOUT
              value: soon`,
			wantErr: "converting 'soon' to type time.Duration",
		},
		{
			name: "secret without secret tag",
			env: `
            - name: TOPIC
              valueFrom:
                secretKeyRef:
                  name: topic
                  key: latest`,
			wantErr: "field Topic does not have the correct secret tag",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var config validateTestConfig
			err := cloudrunner.ValidateConfig(writeServiceSpec(t, tt.env), cloudrunner.WithConfig("test", &config))
			if tt.wantErr == "" {
				assert.NilError(t, err)
				return
			}
			assert.ErrorContains(t, err, tt.wantErr)
		})
	}
}

func TestValidateConfig_ignoresProcessEnvironment(t *testing.T) {
	t.Setenv("TOPIC", "from-process")
	spec := writeServiceSpec(t, `
            - name: COUNT
              value: "3"`)
	var config validateTestConfig
	err := cloudrunner.ValidateConfig(spec, cloudrunner.WithConfig("test", &config))
	assert.ErrorContains(t, err, "required key TOPIC missing value")
}

func TestValidateConfig_leavesProcessEnvironmentAlone(t *testing.T) {
	spec := writeServiceSpec(t, `
            - name: TOPIC
              value: events`)
	var config validateTestConfig
	assert.NilError(t, cloudrunner.ValidateConfig(spec, cloudrunner.WithConfig("test", &config)))
	for _, key := range []string{"TOPIC", "K_SERVICE"} {
		_, ok := os.LookupEnv(key)
		assert.Assert(t, !ok, "%s set in the process environment", key)
	}
}

// writeServiceSpec writes a Cloud Run service specification with the given container env list.
func writeServiceSpec(t *testing.T, env string) string {
	t.Helper()
	spec := `apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: test-service
spec:
  template:
    spec:
      containers:
        - image: test
          env:` + env + "\n"
	path := filepath.Join(t.TempDir(), "service.yaml")
	assert.NilError(t, os.WriteFile(path, []byte(spec), 0o600))
	return path
}
