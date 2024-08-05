package cloudrequestlog

import (
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protopath"
	"google.golang.org/protobuf/reflect/protorange"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

// redact sensitive fields from the input message.
func redact(input proto.Message) proto.Message {
	var hasSensitiveFields bool
	_ = protorange.Range(input.ProtoReflect(), func(values protopath.Values) error {
		last := values.Index(-1)
		if _, ok := last.Value.Interface().(string); !ok {
			return nil
		}
		if last.Step.Kind() != protopath.FieldAccessStep {
			return nil
		}
		if !last.Step.FieldDescriptor().Options().(*descriptorpb.FieldOptions).GetDebugRedact() {
			return nil
		}
		hasSensitiveFields = true
		return protorange.Terminate
	})
	if !hasSensitiveFields {
		return input
	}
	output := proto.Clone(input)
	_ = protorange.Range(output.ProtoReflect(), func(values protopath.Values) error {
		last := values.Index(-1)
		if _, ok := last.Value.Interface().(string); !ok {
			return nil
		}
		if last.Step.Kind() != protopath.FieldAccessStep {
			return nil
		}
		if !last.Step.FieldDescriptor().Options().(*descriptorpb.FieldOptions).GetDebugRedact() {
			return nil
		}
		values.Index(-2).Value.Message().Set(last.Step.FieldDescriptor(), protoreflect.ValueOfString("<redacted>"))
		return nil
	})
	return output
}
