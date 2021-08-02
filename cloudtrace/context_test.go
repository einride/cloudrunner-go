package cloudtrace

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestContext_UnmarshalString(t *testing.T) {
	for _, tt := range []struct {
		name          string
		input         string
		expected      Context
		errorContains string
	}{
		{
			name:  "ok",
			input: "105445aa7843bc8bf206b12000100000/1;o=1",
			expected: Context{
				TraceID: "105445aa7843bc8bf206b12000100000",
				SpanID:  "1",
				Sampled: true,
			},
		},

		{
			name:          "empty",
			input:         "",
			errorContains: "empty x-cloud-trace-context",
		},

		{
			name:          "invalid",
			input:         "foo",
			errorContains: "invalid x-cloud-trace-context 'foo': trace ID is not a 32-character hex value",
		},

		{
			name:  "only trace ID",
			input: "105445aa7843bc8bf206b12000100000",
			expected: Context{
				TraceID: "105445aa7843bc8bf206b12000100000",
			},
		},

		{
			name:  "missing sampled",
			input: "105445aa7843bc8bf206b12000100000/1",
			expected: Context{
				TraceID: "105445aa7843bc8bf206b12000100000",
				SpanID:  "1",
			},
		},

		{
			name:          "malformed sampled",
			input:         "105445aa7843bc8bf206b12000100000/1;foo",
			errorContains: "invalid x-cloud-trace-context '105445aa7843bc8bf206b12000100000/1;foo'",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var actual Context
			err := actual.UnmarshalString(tt.input)
			if tt.errorContains != "" {
				assert.ErrorContains(t, err, tt.errorContains)
			} else {
				assert.NilError(t, err)
				assert.Equal(t, tt.expected, actual)
			}
		})
	}
}
