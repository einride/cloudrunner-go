package cloudclient

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Middleware struct{}

// GRPCUnaryClientInterceptor adds standard middleware for gRPC clients.
func (l *Middleware) GRPCUnaryClientInterceptor(
	ctx context.Context,
	fullMethod string,
	request interface{},
	response interface{},
	cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker,
	opts ...grpc.CallOption,
) error {
	return handleHTTPResponseToGRPCRequest(invoker(ctx, fullMethod, request, response, cc, opts...))
}

func handleHTTPResponseToGRPCRequest(errInput error) error {
	if errInput == nil {
		return nil
	}
	// When grpc-go encounters http content type text/html it will not
	// forward any actionable fields for us to look at. Instead we resort
	// to matching strings.
	//
	// These strings are coming from:
	// * "google.golang.org/grpc/internal/transport/http_util.go".
	// * "google.golang.org/grpc/internal/transport/http2_client.go".
	errStatus := status.Convert(errInput)
	if errStatus.Code() != codes.Unknown {
		return errInput
	}
	errorMessage := errStatus.Message()
	if !isContentTypeHTML(errorMessage) {
		return errInput
	}
	if isStatusCode(errorMessage, http.StatusForbidden) {
		// This happens when the gRPC request got rejected due to missing IAM permissions.
		// The request gets rejected at the HTTP level and a gRPC error will not be available.
		return status.Errorf(
			codes.PermissionDenied,
			"the gRPC request failed with a HTTP 403 error "+
				"(on Google Cloud this happens when the client service account does not have IAM permissions "+
				"to call the remote service - "+
				"on Cloud Run, the client service account must have roles/run.invoker on the remote service): %v",
			errorMessage,
		)
	}
	// Other HTTP responses to gRPC requests are assumed to be transient.
	return status.Error(codes.Unavailable, errorMessage)
}

func isContentTypeHTML(msg string) bool {
	return strings.Contains(msg, "transport") && strings.Contains(msg, `content-type "text/html`)
}

func isStatusCode(msg string, statusCode int) bool {
	return strings.Contains(msg, strconv.Itoa(statusCode))
}
