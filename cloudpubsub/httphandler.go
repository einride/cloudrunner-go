package cloudpubsub

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"go.einride.tech/cloudrunner/cloudrequestlog"
	"go.einride.tech/cloudrunner/cloudstatus"
	"go.einride.tech/cloudrunner/cloudzap"
	"go.uber.org/zap"
	"google.golang.org/genproto/googleapis/pubsub/v1"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// HTTPHandler creates a new HTTP handler for Cloud Pub/Sub push messages.
// See: https://cloud.google.com/pubsub/docs/push
func HTTPHandler(fn func(context.Context, *pubsub.PubsubMessage) error) http.Handler {
	return httpHandlerFn(fn)
}

type httpHandlerFn func(ctx context.Context, message *pubsub.PubsubMessage) error

func (fn httpHandlerFn) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Subscription string `json:"subscription"`
		Message      struct {
			Attributes  map[string]string `json:"attributes"`
			Data        []byte            `json:"data"`
			MessageID   string            `json:"messageId"`
			PublishTime time.Time         `json:"publishTime"`
		} `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		if fields, ok := cloudrequestlog.GetAdditionalFields(r.Context()); ok {
			fields.Add(zap.Error(err))
		}
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	pubsubMessage := pubsub.PubsubMessage{
		Data:        payload.Message.Data,
		Attributes:  payload.Message.Attributes,
		MessageId:   payload.Message.MessageID,
		PublishTime: timestamppb.New(payload.Message.PublishTime),
	}
	if fields, ok := cloudrequestlog.GetAdditionalFields(r.Context()); ok {
		fields.Add(cloudzap.ProtoMessage("pubsubMessage", &pubsubMessage))
	}
	if err := fn(r.Context(), &pubsubMessage); err != nil {
		if fields, ok := cloudrequestlog.GetAdditionalFields(r.Context()); ok {
			fields.Add(zap.Error(err))
		}
		code := status.Code(err)
		httpStatus := cloudstatus.ToHTTP(code)
		http.Error(w, http.StatusText(httpStatus), httpStatus)
		return
	}
}
