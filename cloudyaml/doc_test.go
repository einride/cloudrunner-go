package cloudyaml

import (
	"context"
	"testing"

	"gotest.tools/v3/assert"
)

func TestResolveEnvFromFile(t *testing.T) {
	actual, err := ResolveEnvFromFile(context.Background(), "-", "testdata/example.yaml")
	assert.NilError(t, err)
	assert.DeepEqual(t, []string{"K_SERVICE=example", "FOO=baz", "BAR=3"}, actual)
}
