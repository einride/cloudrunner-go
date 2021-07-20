package cloudzap

import (
	"strings"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"
	"google.golang.org/genproto/googleapis/example/library/v1"
	"gotest.tools/v3/assert"
)

func TestProtoMessage(t *testing.T) {
	var buffer zaptest.Buffer
	encoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	logger := zap.New(zapcore.NewCore(encoder, &buffer, zap.DebugLevel))
	logger.Info("test", ProtoMessage("protoMessage", &library.Book{
		Name:   "name",
		Author: "author",
		Title:  "title",
		Read:   true,
	}))
	assert.Assert(
		t,
		strings.Contains(
			buffer.Stripped(),
			`"protoMessage":{"name":"name","author":"author","title":"title","read":true}`,
		),
	)
}
