package cloudrequestlog

import (
	"fmt"
	"net/http"
	"strings"

	"google.golang.org/grpc/codes"
)

func grpcServerLogMessage(code codes.Code, fullMethod string) string {
	methodName := fullMethod
	if _, method, ok := splitFullMethod(fullMethod); ok {
		methodName = method
	}
	return fmt.Sprintf("gRPCServer %s %s", code.String(), methodName)
}

func grpcClientLogMessage(code codes.Code, fullMethod string) string {
	methodName := fullMethod
	if _, method, ok := splitFullMethod(fullMethod); ok {
		methodName = method
	}
	return fmt.Sprintf("gRPCClient %s %s", code.String(), methodName)
}

func httpServerLogMessage(res *httpResponseWriter, req *http.Request) string {
	return fmt.Sprintf(
		"HTTPServer %d %s/%s",
		res.Status(),
		req.Host,
		strings.TrimPrefix(req.RequestURI, "/"),
	)
}
