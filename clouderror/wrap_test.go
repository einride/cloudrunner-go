package clouderror

import (
	"context"
	"fmt"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gotest.tools/v3/assert"
)

func Test_WrapTransient(t *testing.T) {
	for _, tt := range []struct {
		name         string
		err          error
		expectedCode codes.Code
	}{
		{
			name:         "nil",
			err:          nil,
			expectedCode: codes.Internal,
		},
		{
			name:         "codes.DeadlineExceeded",
			err:          status.Error(codes.DeadlineExceeded, "transient"),
			expectedCode: codes.DeadlineExceeded,
		},
		{
			name:         "codes.Canceled",
			err:          status.Error(codes.Canceled, "transient"),
			expectedCode: codes.Canceled,
		},
		{
			name:         "codes.Unavailable",
			err:          status.Error(codes.Unavailable, "transient"),
			expectedCode: codes.Unavailable,
		},
		{
			name: "wrapped transient",
			err: Wrap(
				fmt.Errorf("network unavailable"),
				status.New(codes.Unavailable, "bad"),
			),
			expectedCode: codes.Unavailable,
		},
		{
			name:         "context.DeadlineExceeded",
			err:          context.DeadlineExceeded,
			expectedCode: codes.DeadlineExceeded,
		},
		{
			name:         "context.Canceled",
			err:          context.Canceled,
			expectedCode: codes.Canceled,
		},
		{
			name:         "wrapped context.Canceled",
			err:          fmt.Errorf("bad: %w", context.Canceled),
			expectedCode: codes.Canceled,
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := WrapTransient(tt.err, "boom")
			assert.Equal(t, tt.expectedCode, status.Code(got), got)
		})
	}
}
