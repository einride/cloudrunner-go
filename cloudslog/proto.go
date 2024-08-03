package cloudslog

import (
	"log/slog"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func newProtoValue(m proto.Message, sizeLimit int) protoValue {
	return protoValue{Message: m, sizeLimit: sizeLimit}
}

type protoValue struct {
	proto.Message
	sizeLimit int
}

func (v protoValue) LogValue() slog.Value {
	if size := proto.Size(v.Message); size > v.sizeLimit {
		return slog.GroupValue(
			slog.String("message", "truncated due to size"),
			slog.Int("size", size),
			slog.Int("limit", v.sizeLimit),
		)
	}
	return slog.AnyValue(jsonProtoValue{Message: v.Message})
}

type jsonProtoValue struct {
	proto.Message
}

func (v jsonProtoValue) MarshalJSON() ([]byte, error) {
	return protojson.MarshalOptions{}.Marshal(v.Message)
}
