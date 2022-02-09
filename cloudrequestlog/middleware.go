package cloudrequestlog

import (
	"context"
	"errors"
	"net/http"
	"reflect"
	"runtime"
	"time"

	"go.einride.tech/cloudrunner/cloudzap"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
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
		responseWriter := &httpResponseWriter{ResponseWriter: w}
		ctx := WithAdditionalFields(r.Context())
		r = r.WithContext(ctx)
		startTime := time.Now()
		next.ServeHTTP(responseWriter, r)
		checkedEntry := l.logger(ctx).Check(
			l.statusToLevel(responseWriter.Status()),
			httpServerLogMessage(responseWriter, r),
		)
		if checkedEntry == nil {
			return
		}
		httpRequest := cloudzap.HTTPRequestObject{
			RequestMethod: r.Method,
			Status:        responseWriter.statusCode,
			ResponseSize:  responseWriter.size + measureHeaderSize(w.Header()),
			UserAgent:     r.UserAgent(),
			RemoteIP:      r.RemoteAddr,
			Referer:       r.Referer(),
			Latency:       time.Since(startTime),
			Protocol:      r.Proto,
		}
		if r.URL != nil {
			httpRequest.RequestURL = r.URL.String()
		}
		fields := []zapcore.Field{
			cloudzap.HTTPRequest(&httpRequest),
		}
		if additionalFields, ok := GetAdditionalFields(ctx); ok {
			fields = additionalFields.AppendTo(fields)
		}
		checkedEntry.Write(fields...)
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
	switch code {
	case codes.OK:
		return zap.DebugLevel
	case
		codes.NotFound,
		codes.InvalidArgument,
		codes.AlreadyExists,
		codes.FailedPrecondition,
		codes.Unauthenticated,
		codes.PermissionDenied,
		codes.DeadlineExceeded,
		codes.OutOfRange,
		codes.Canceled,
		codes.Aborted:
		return zap.WarnLevel
	case
		codes.Unknown, codes.Internal, codes.DataLoss:
		return zap.ErrorLevel
	default:
		return zap.ErrorLevel
	}
}

func (l *Middleware) statusToLevel(status int) zapcore.Level {
	if level, ok := l.Config.StatusToLevel[status]; ok {
		return level
	}
	switch {
	case status < http.StatusBadRequest:
		return zap.DebugLevel
	case http.StatusBadRequest <= status && status < http.StatusInternalServerError:
		return zap.WarnLevel
	case status == http.StatusGatewayTimeout:
		// special case for 504 (timeout) to match severity for gRPC status codes
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
