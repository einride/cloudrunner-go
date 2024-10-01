package cloudslog

import (
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protopath"
	"google.golang.org/protobuf/reflect/protorange"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

func needsRedact(input proto.Message) bool {
	var result bool
	_ = protorange.Range(input.ProtoReflect(), func(values protopath.Values) error {
		last := values.Index(-1)
		if _, ok := last.Value.Interface().(string); !ok {
			return nil
		}
		if last.Step.Kind() != protopath.FieldAccessStep {
			return nil
		}
		if last.Step.FieldDescriptor().Options().(*descriptorpb.FieldOptions).GetDebugRedact() {
			result = true
			return protorange.Terminate
		}
		return nil
	})
	return result
}

func redact(input proto.Message) {
	_ = protorange.Range(input.ProtoReflect(), func(values protopath.Values) error {
		last := values.Index(-1)
		if _, ok := last.Value.Interface().(string); !ok {
			return nil
		}
		if last.Step.Kind() != protopath.FieldAccessStep {
			return nil
		}
		if last.Step.FieldDescriptor().Options().(*descriptorpb.FieldOptions).GetDebugRedact() {
			values.Index(-2).Value.Message().Set(last.Step.FieldDescriptor(), protoreflect.ValueOfString("<redacted>"))
			return nil
		}
		values.Index(-2).Value.Message().Set(last.Step.FieldDescriptor(), protoreflect.ValueOfString("<redacted>"))
		return nil
	})
}
