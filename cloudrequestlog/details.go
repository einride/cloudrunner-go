package cloudrequestlog

import (
	"encoding/json"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

// ErrorDetails creates a zap.Field that logs the gRPC error details of the provided error.
func ErrorDetails(err error) zap.Field {
	if err == nil {
		return zap.Skip()
	}
	s, ok := status.FromError(err)
	if !ok {
		return zap.Skip()
	}
	protoDetails := s.Proto().GetDetails()
	if len(protoDetails) == 0 {
		return zap.Skip()
	}
	return zap.Array("errorDetails", errorDetailsMarshaler(protoDetails))
}

type errorDetailsMarshaler []*anypb.Any

var _ zapcore.ArrayMarshaler = errorDetailsMarshaler{}

// MarshalLogArray implements zapcore.ArrayMarshaler.
func (d errorDetailsMarshaler) MarshalLogArray(encoder zapcore.ArrayEncoder) error {
	for _, detail := range d {
		if err := encoder.AppendReflected(reflectProtoMessage{message: detail}); err != nil {
			return err
		}
	}
	return nil
}

type reflectProtoMessage struct {
	message proto.Message
}

var _ json.Marshaler = reflectProtoMessage{}

func (p reflectProtoMessage) MarshalJSON() ([]byte, error) {
	return protojson.Marshal(p.message)
}
