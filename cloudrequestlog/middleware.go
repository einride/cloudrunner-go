package cloudrequestlog

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"reflect"
	"runtime"
	"time"

	"go.einride.tech/cloudrunner/cloudstream"
	"go.einride.tech/cloudrunner/cloudzap"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
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
	// MessageTransformer is an optional transform applied to proto.Message request and responses.
	MessageTransformer func(proto.Message) proto.Message
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
	code := status.Code(err)
	checkedEntry := l.logger(ctx).Check(
		l.codeToLevel(code),
		grpcServerLogMessage(code, info.FullMethod),
	)
	if checkedEntry == nil {
		return response, err
	}
	grpcRequest := cloudzap.HTTPRequestObject{
		Protocol: "gRPC",
		Latency:  time.Since(startTime),
	}
	grpcRequest.RequestSize = proto.Size(requestClone)
	if protoResponse, ok := response.(proto.Message); ok {
		grpcRequest.ResponseSize = proto.Size(protoResponse)
	}
	fields := []zapcore.Field{
		zap.Stringer("code", code),
		zap.Object("httpRequest", &grpcRequest),
		l.messageField("request", requestClone),
		l.messageField("response", response),
		zap.Error(err),
		ErrorDetails(err),
	}
	fields = appendFullMethodFields(info.FullMethod, fields)
	if additionalFields, ok := GetAdditionalFields(ctx); ok {
		fields = additionalFields.AppendTo(fields)
	}
	var errCaller interface {
		Caller() (pc uintptr, file string, line int, ok bool)
	}
	if errors.As(err, &errCaller) {
		checkedEntry.Caller = zapcore.NewEntryCaller(errCaller.Caller())
		checkedEntry.Entry.Caller = checkedEntry.Caller
	}
	fields = append(fields, cloudzap.SourceLocationForCaller(checkedEntry.Caller))
	checkedEntry.Write(fields...)
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
	code := status.Code(err)
	checkedEntry := l.logger(ctx).Check(
		l.codeToLevel(code),
		grpcServerLogMessage(code, info.FullMethod),
	)
	if checkedEntry == nil {
		return err
	}
	grpcRequest := cloudzap.HTTPRequestObject{
		Protocol: "gRPC",
		Latency:  time.Since(startTime),
	}
	fields := []zapcore.Field{
		zap.Stringer("code", code),
		zap.Object("httpRequest", &grpcRequest),
		zap.Error(err),
		ErrorDetails(err),
	}
	fields = appendFullMethodFields(info.FullMethod, fields)
	if additionalFields, ok := GetAdditionalFields(ctx); ok {
		fields = additionalFields.AppendTo(fields)
	}
	var errCaller interface {
		Caller() (pc uintptr, file string, line int, ok bool)
	}
	if errors.As(err, &errCaller) {
		checkedEntry.Caller = zapcore.NewEntryCaller(errCaller.Caller())
		checkedEntry.Entry.Caller = checkedEntry.Caller
	}
	fields = append(fields, cloudzap.SourceLocationForCaller(checkedEntry.Caller))
	checkedEntry.Write(fields...)
	return err
}

func (l *Middleware) logger(ctx context.Context) *zap.Logger {
	logger, ok := cloudzap.GetLogger(ctx)
	if !ok {
		panic("cloudrequestlog.Middleware requires a logger in the context")
	}
	return logger
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
	code := status.Code(err)
	checkedEntry := l.logger(ctx).Check(
		l.codeToLevel(code),
		grpcClientLogMessage(code, fullMethod),
	)
	if checkedEntry == nil {
		return err
	}
	grpcRequest := cloudzap.HTTPRequestObject{
		Protocol: "gRPC",
		Latency:  time.Since(startTime),
	}
	grpcRequest.RequestSize = proto.Size(requestClone)
	if protoResponse, ok := response.(proto.Message); ok {
		grpcRequest.ResponseSize = proto.Size(protoResponse)
	}
	// assuming this middleware is first in the chain, the caller of the client method is 4 stack frames up
	checkedEntry.Caller = zapcore.NewEntryCaller(runtime.Caller(4))
	checkedEntry.Entry.Caller = checkedEntry.Caller
	fields := []zap.Field{
		zap.Stringer("code", code),
		zap.Object("httpRequest", &grpcRequest),
		l.messageField("request", requestClone),
		l.messageField("response", response),
		zap.Error(err),
		ErrorDetails(err),
		cloudzap.SourceLocationForCaller(checkedEntry.Caller),
	}
	fields = appendFullMethodFields(fullMethod, fields)
	checkedEntry.Write(fields...)
	return err
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
		if !logger.Enabled(ctx, levelToSlog(level)) {
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
		logger.LogAttrs(ctx, levelToSlog(level), logMessage, attrs...)
	})
}

func (l *Middleware) messageField(key string, message interface{}) zap.Field {
	protoMessage, ok := message.(proto.Message)
	if !ok || protoMessage == nil || reflect.ValueOf(protoMessage).IsNil() {
		return zap.Skip()
	}
	if l.Config.MessageSizeLimit > 0 {
		size := proto.Size(protoMessage)
		if size > l.Config.MessageSizeLimit {
			return zap.Object(key, truncatedMessageField{size: size, sizeLimit: l.Config.MessageSizeLimit})
		}
	}
	return cloudzap.ProtoMessage(key, l.applyMessageTransform(protoMessage))
}

func (l *Middleware) applyMessageTransform(message proto.Message) proto.Message {
	if l.MessageTransformer == nil {
		return message
	}
	return l.MessageTransformer(message)
}

func (l *Middleware) codeToLevel(code codes.Code) zapcore.Level {
	if level, ok := l.Config.CodeToLevel[code]; ok {
		return level
	}
	return CodeToLevel(code)
}

func (l *Middleware) statusToLevel(status int) zapcore.Level {
	if level, ok := l.Config.StatusToLevel[status]; ok {
		return level
	}
	switch {
	case status < http.StatusBadRequest:
		return zap.InfoLevel
	case http.StatusBadRequest <= status && status < http.StatusInternalServerError:
		return zap.WarnLevel
	case status == http.StatusGatewayTimeout || status == http.StatusServiceUnavailable:
		// special case for 503 (unavailable) and 504 (timeout) to match severity for gRPC status codes
		return zap.WarnLevel
	default:
		return zap.ErrorLevel
	}
}

func appendFullMethodFields(fullMethod string, dst []zap.Field) []zap.Field {
	service, method, ok := splitFullMethod(fullMethod)
	if !ok {
		return dst
	}
	return append(
		dst,
		zap.String("service", service),
		zap.String("method", method),
	)
}
