package cloudzap

import (
	"encoding/json"

	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// ProtoMessage constructs a zap.Field with the given key and proto message encoded as JSON.
func ProtoMessage(key string, message proto.Message) zap.Field {
	return zap.Reflect(key, reflectProtoMessage{message: message})
}

type reflectProtoMessage struct {
	message proto.Message
}

var _ json.Marshaler = reflectProtoMessage{}

func (p reflectProtoMessage) MarshalJSON() ([]byte, error) {
	return protojson.Marshal(p.message)
}
