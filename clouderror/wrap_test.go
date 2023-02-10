package clouderror

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gotest.tools/v3/assert"
)

func Test_WrapTransient(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name         string
		err          error
		expectedCode codes.Code
	}{
		{
			name:         "non gRPC error",
			err:          errors.New("err"),
			expectedCode: codes.Internal,
		},
		{
			name:         "codes.Canceled",
			err:          status.Error(codes.Canceled, "transient"),
			expectedCode: codes.Canceled,
		},
		{
			name:         "codes.Unknown",
			err:          status.Error(codes.Unknown, "transient"),
			expectedCode: codes.Unknown,
		},
		{
			name:         "codes.InvalidArgument",
			err:          status.Error(codes.InvalidArgument, "transient"),
			expectedCode: codes.InvalidArgument,
		},
		{
			name:         "codes.DeadlineExceeded",
			err:          status.Error(codes.DeadlineExceeded, "transient"),
			expectedCode: codes.DeadlineExceeded,
		},
		{
			name:         "codes.NotFound",
			err:          status.Error(codes.NotFound, "transient"),
			expectedCode: codes.NotFound,
		},
		{
			name:         "codes.AlreadyExists",
			err:          status.Error(codes.AlreadyExists, "transient"),
			expectedCode: codes.AlreadyExists,
		},
		{
			name:         "codes.PermissionDenied",
			err:          status.Error(codes.PermissionDenied, "transient"),
			expectedCode: codes.PermissionDenied,
		},
		{
			name:         "codes.ResourceExhausted",
			err:          status.Error(codes.ResourceExhausted, "transient"),
			expectedCode: codes.ResourceExhausted,
		},
		{
			name:         "codes.FailedPrecondition",
			err:          status.Error(codes.FailedPrecondition, "transient"),
			expectedCode: codes.FailedPrecondition,
		},
		{
			name:         "codes.Aborted",
			err:          status.Error(codes.Aborted, "transient"),
			expectedCode: codes.Aborted,
		},
		{
			name:         "codes.OutOfRange",
			err:          status.Error(codes.OutOfRange, "transient"),
			expectedCode: codes.OutOfRange,
		},
		{
			name:         "codes.Unimplemented",
			err:          status.Error(codes.Unimplemented, "transient"),
			expectedCode: codes.Unimplemented,
		},
		{
			name:         "codes.Unavailable",
			err:          status.Error(codes.Unavailable, "transient"),
			expectedCode: codes.Unavailable,
		},
		{
			name:         "codes.DataLoss",
			err:          status.Error(codes.DataLoss, "transient"),
			expectedCode: codes.DataLoss,
		},
		{
			name:         "codes.Unauthenticated",
			err:          status.Error(codes.Unauthenticated, "transient"),
			expectedCode: codes.Unauthenticated,
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
