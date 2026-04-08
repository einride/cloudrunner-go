package cloudrequestlog

import (
	"encoding/json"
	"log/slog"

	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// ErrorDetails creates a slog.Attr that logs the gRPC error details of the provided error.
func ErrorDetails(err error) slog.Attr {
	if err == nil {
		return slog.Attr{}
	}
	s, ok := status.FromError(err)
	if !ok {
		return slog.Attr{}
	}
	protoDetails := s.Proto().GetDetails()
	if len(protoDetails) == 0 {
		return slog.Attr{}
	}
	details := make([]reflectProtoMessage, len(protoDetails))
	for i, detail := range protoDetails {
		details[i] = reflectProtoMessage{message: detail}
	}
	return slog.Any("errorDetails", details)
}

type reflectProtoMessage struct {
	message proto.Message
}

var _ json.Marshaler = reflectProtoMessage{}

func (p reflectProtoMessage) MarshalJSON() ([]byte, error) {
	return protojson.Marshal(p.message)
}
