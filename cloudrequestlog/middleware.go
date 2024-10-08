package cloudrequestlog

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"runtime"
	"time"

	"go.einride.tech/cloudrunner/cloudstream"
	ltype "google.golang.org/genproto/googleapis/logging/type"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
)

// Middleware for request logging.
type Middleware struct {
	// Config for the request logger middleware.
	Config Config
}

// GRPCUnaryServerInterceptor implements request logging as a grpc.UnaryServerInterceptor.
func (l *Middleware) GRPCUnaryServerInterceptor(
	ctx context.Context,
	request interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	startTime := time.Now()
	ctx = WithAdditionalFields(ctx)
	// Clone request to ensure not using a mutated one later
	requestClone := proto.Clone(request.(proto.Message))
	response, err := handler(ctx, request)
	responseStatus := status.Convert(err)
	level := l.codeToLevel(responseStatus.Code())
	logger := slog.Default()
	if !logger.Enabled(ctx, level) {
		return response, err
	}
	grpcRequest := &ltype.HttpRequest{
		Protocol: "gRPC",
		Latency:  durationpb.New(time.Since(startTime)),
	}
	grpcRequest.RequestSize = int64(proto.Size(requestClone))
	if protoResponse, ok := response.(proto.Message); ok {
		grpcRequest.ResponseSize = int64(proto.Size(protoResponse))
	}
	attrs := []slog.Attr{
		slog.String("code", responseStatus.Code().String()),
		slog.Any("status", responseStatus),
		slog.Any("httpRequest", grpcRequest),
		slog.Any("request", requestClone),
		slog.Any("response", response),
	}
	if err != nil {
		attrs = append(attrs, slog.Any("error", err))
	}
	attrs = appendFullMethodAttrs(info.FullMethod, attrs)
	if additionalFields, ok := GetAdditionalFields(ctx); ok {
		attrs = additionalFields.appendTo(attrs)
	}
	var errCaller interface {
		Caller() (pc uintptr, file string, line int, ok bool)
	}
	if errors.As(err, &errCaller) {
		attrs = append(attrs, newSourceAttr(errCaller.Caller()))
	}
	logger.LogAttrs(ctx, level, grpcServerLogMessage(responseStatus.Code(), info.FullMethod), attrs...)
	return response, err
}

// GRPCStreamServerInterceptor implements request logging as a grpc.UnaryServerInterceptor.
// This middleware differs from the unary one in that it does not log request or response payload.
// The reason for this is that this info is not readily available in the middleware layer.
func (l *Middleware) GRPCStreamServerInterceptor(
	srv interface{},
	ss grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {
	startTime := time.Now()
	ctx := WithAdditionalFields(ss.Context())
	ss = cloudstream.NewContextualServerStream(ctx, ss)
	err := handler(srv, ss)
	responseStatus := status.Convert(err)
	level := l.codeToLevel(responseStatus.Code())
	logger := slog.Default()
	if !logger.Enabled(ctx, level) {
		return err
	}
	grpcRequest := &ltype.HttpRequest{
		Protocol: "gRPC",
		Latency:  durationpb.New(time.Since(startTime)),
	}
	attrs := []slog.Attr{
		slog.String("code", responseStatus.Code().String()),
		slog.Any("status", responseStatus),
		slog.Any("httpRequest", grpcRequest),
	}
	if err != nil {
		attrs = append(attrs, slog.Any("error", err))
	}
	attrs = appendFullMethodAttrs(info.FullMethod, attrs)
	if additionalFields, ok := GetAdditionalFields(ctx); ok {
		attrs = additionalFields.appendTo(attrs)
	}
	var errCaller interface {
		Caller() (pc uintptr, file string, line int, ok bool)
	}
	if errors.As(err, &errCaller) {
		attrs = append(attrs, newSourceAttr(errCaller.Caller()))
	}
	logger.LogAttrs(ctx, level, grpcServerLogMessage(responseStatus.Code(), info.FullMethod), attrs...)
	return err
}

// GRPCUnaryClientInterceptor provides request logging as a grpc.UnaryClientInterceptor.
func (l *Middleware) GRPCUnaryClientInterceptor(
	ctx context.Context,
	fullMethod string,
	request interface{},
	response interface{},
	cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker,
	opts ...grpc.CallOption,
) error {
	startTime := time.Now()
	// Clone request to ensure not using a mutated one later
	requestClone := proto.Clone(request.(proto.Message))
	err := invoker(ctx, fullMethod, request, response, cc, opts...)
	responseStatus := status.Convert(err)
	level := l.codeToLevel(responseStatus.Code())
	logger := slog.Default()
	if !logger.Enabled(ctx, level) {
		return err
	}
	grpcRequest := &ltype.HttpRequest{
		Protocol: "gRPC",
		Latency:  durationpb.New(time.Since(startTime)),
	}
	grpcRequest.RequestSize = int64(proto.Size(requestClone))
	if protoResponse, ok := response.(proto.Message); ok {
		grpcRequest.ResponseSize = int64(proto.Size(protoResponse))
	}
	attrs := []slog.Attr{
		slog.String("code", responseStatus.Code().String()),
		slog.Any("status", responseStatus),
		slog.Any("httpRequest", grpcRequest),
		slog.Any("request", requestClone),
		slog.Any("response", response),
		// assuming this middleware is first in the chain, the caller of the client method is 4 stack frames up
		newSourceAttr(runtime.Caller(4)),
	}
	if err != nil {
		attrs = append(attrs, slog.Any("error", err))
	}
	attrs = appendFullMethodAttrs(fullMethod, attrs)
	logger.LogAttrs(ctx, level, grpcClientLogMessage(responseStatus.Code(), fullMethod), attrs...)
	return err
}

func newSourceAttr(pc uintptr, file string, line int, _ bool) slog.Attr {
	return slog.Any(
		slog.SourceKey,
		&slog.Source{
			Function: runtime.FuncForPC(pc).Name(),
			File:     file,
			Line:     line,
		},
	)
}

func measureHeaderSize(h http.Header) int {
	var result int
	for k, vs := range h {
		result += len(k)
		for _, v := range vs {
			result += len(v)
		}
	}
	return result
}

// HTTPServer provides request logging for HTTP servers.
func (l *Middleware) HTTPServer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		ctx := WithAdditionalFields(r.Context())
		r = r.WithContext(ctx)
		responseWriter := &httpResponseWriter{ResponseWriter: w}
		next.ServeHTTP(responseWriter, r)
		level := l.statusToLevel(responseWriter.Status())
		logger := slog.Default()
		if !logger.Enabled(ctx, level) {
			return
		}
		logMessage := httpServerLogMessage(responseWriter, r)
		httpRequest := &ltype.HttpRequest{
			RequestMethod: r.Method,
			Status:        int32(responseWriter.Status()),
			ResponseSize:  int64(responseWriter.size + measureHeaderSize(w.Header())),
			UserAgent:     r.UserAgent(),
			RemoteIp:      r.RemoteAddr,
			Referer:       r.Referer(),
			Latency:       durationpb.New(time.Since(startTime)),
			Protocol:      r.Proto,
		}
		if r.URL != nil {
			httpRequest.RequestUrl = r.URL.String()
		}
		attrs := []slog.Attr{
			slog.Any("httpRequest", &httpRequest),
		}
		if additionalFields, ok := GetAdditionalFields(ctx); ok {
			attrs = additionalFields.appendTo(attrs)
		}
		logger.LogAttrs(ctx, level, logMessage, attrs...)
	})
}

func (l *Middleware) codeToLevel(code codes.Code) slog.Level {
	if level, ok := l.Config.CodeToLevel[code]; ok {
		return level
	}
	return codeToLevel(code)
}

func (l *Middleware) statusToLevel(status int) slog.Level {
	if level, ok := l.Config.StatusToLevel[status]; ok {
		return level
	}
	switch {
	case status < http.StatusBadRequest:
		return slog.LevelInfo
	case http.StatusBadRequest <= status && status < http.StatusInternalServerError:
		return slog.LevelWarn
	case status == http.StatusGatewayTimeout || status == http.StatusServiceUnavailable:
		// special case for 503 (unavailable) and 504 (timeout) to match severity for gRPC status codes
		return slog.LevelWarn
	default:
		return slog.LevelError
	}
}

func appendFullMethodAttrs(fullMethod string, dst []slog.Attr) []slog.Attr {
	service, method, ok := splitFullMethod(fullMethod)
	if !ok {
		return dst
	}
	return append(
		dst,
		slog.String("service", service),
		slog.String("method", method),
	)
}
