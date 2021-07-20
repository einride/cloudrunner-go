package cloudrequestlog

import "go.uber.org/zap/zapcore"

type truncatedMessageField struct {
	size      int
	sizeLimit int
}

func (t truncatedMessageField) MarshalLogObject(e zapcore.ObjectEncoder) error {
	e.AddString("message", "truncated due to size")
	e.AddInt("size", t.size)
	e.AddInt("sizeLimit", t.sizeLimit)
	return nil
}
