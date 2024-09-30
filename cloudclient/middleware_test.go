package cloudclient

import (
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gotest.tools/v3/assert"
)

func TestHandleHTTPResponseToGRPCRequest(t *testing.T) {
	for _, tt := range []struct {
		name     string
		err      error
		expected error
	}{
		{
			name: "gRPC NotFound",
			err: status.Error(
				codes.NotFound,
				"not found",
			),
			expected: status.Error(
				codes.NotFound,
				"not found",
			),
		},

		{
			name: "HTTP 200 OK",
			err: status.Error(
				codes.Unknown,
				"OK: HTTP status code 200; transport: received the unexpected content-type \"text/html\"",
			),
			expected: status.Error(
				codes.Unavailable,
				"OK: HTTP status code 200; transport: received the unexpected content-type \"text/html\"",
			),
		},

		{
			name: "HTTP 500 Internal Server Error",
			err: status.Error(
				codes.Unknown,
				"rpc error: code = Unknown desc = Internal Server Error: HTTP status code 500; "+
					"transport: received the unexpected content-type \"text/html; charset=UTF-8\"",
			),
			expected: status.Error(
				codes.Unavailable,
				"rpc error: code = Unknown desc = Internal Server Error: HTTP status code 500; "+
					"transport: received the unexpected content-type \"text/html; charset=UTF-8\"",
			),
		},
		{
			name: "HTTP 500 Internal Server Error",
			err: status.Error(
				codes.Unknown,
				"unexpected HTTP status code received from server: 500 (Internal Server Error); "+
					"transport: received unexpected content-type \"text/html; charset=UTF-8\"",
			),
			expected: status.Error(
				codes.Unavailable,
				"unexpected HTTP status code received from server: 500 (Internal Server Error); "+
					"transport: received unexpected content-type \"text/html; charset=UTF-8\"",
			),
		},

		{
			name: "HTTP 403 Forbidden",
			err: status.Error(
				codes.Unknown,
				"Forbidden: HTTP status code 403; transport: received the unexpected content-type \"text/html\"",
			),
			expected: status.Error(
				codes.PermissionDenied,
				"the gRPC request failed with a HTTP 403 error (on Google Cloud this happens when the client "+
					"service account does not have IAM permissions to call the remote service - on Cloud Run, "+
					"the client service account must have roles/run.invoker on the remote service): "+
					"Forbidden: HTTP status code 403; transport: received the unexpected content-type \"text/html\"",
			),
		},
		{
			name: "HTTP 403 Forbidden",
			err: status.Error(
				codes.Unknown,
				"unexpected HTTP status code received from server: 403 (Forbidden); "+
					"transport: received unexpected content-type \"text/html; charset=UTF-8\"",
			),
			expected: status.Error(
				codes.PermissionDenied,
				"the gRPC request failed with a HTTP 403 error (on Google Cloud this happens when the client "+
					"service account does not have IAM permissions to call the remote service - on Cloud Run, "+
					"the client service account must have roles/run.invoker on the remote service): "+
					"unexpected HTTP status code received from server: 403 (Forbidden); "+
					"transport: received unexpected content-type \"text/html; charset=UTF-8\"",
			),
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			actual := handleHTTPResponseToGRPCRequest(tt.err)
			assert.Equal(t, status.Code(actual), status.Code(tt.expected))
			assert.Equal(t, status.Convert(actual).Message(), status.Convert(tt.expected).Message())
		})
	}
}
