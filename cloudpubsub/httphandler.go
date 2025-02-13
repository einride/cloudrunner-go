package cloudpubsub

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"cloud.google.com/go/pubsub/apiv1/pubsubpb"
	"go.einride.tech/cloudrunner/cloudrequestlog"
	"go.einride.tech/cloudrunner/cloudstatus"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// HTTPHandler creates a new HTTP handler for Cloud Pub/Sub push messages.
// See: https://cloud.google.com/pubsub/docs/push
func HTTPHandler(fn func(context.Context, *pubsubpb.PubsubMessage) error) http.Handler {
	return httpHandlerFn(fn)
}

type httpHandlerFn func(ctx context.Context, message *pubsubpb.PubsubMessage) error

func (fn httpHandlerFn) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	var payload struct {
		Subscription string `json:"subscription"`
		Message      struct {
			Attributes  map[string]string `json:"attributes"`
			Data        []byte            `json:"data"`
			MessageID   string            `json:"messageId"`
			PublishTime time.Time         `json:"publishTime"`
			OrderingKey string            `json:"orderingKey"`
		} `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		if fields, ok := cloudrequestlog.GetAdditionalFields(r.Context()); ok {
			fields.Add(slog.Any("error", err))
		}
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	pubsubMessage := pubsubpb.PubsubMessage{
		Data:        payload.Message.Data,
		Attributes:  payload.Message.Attributes,
		MessageId:   payload.Message.MessageID,
		PublishTime: timestamppb.New(payload.Message.PublishTime),
		OrderingKey: payload.Message.OrderingKey,
	}
	if fields, ok := cloudrequestlog.GetAdditionalFields(r.Context()); ok {
		fields.Add(slog.Any("pubsubMessage", &pubsubMessage))
	}
	ctx := withSubscription(r.Context(), payload.Subscription)
	if err := fn(ctx, &pubsubMessage); err != nil {
		if fields, ok := cloudrequestlog.GetAdditionalFields(r.Context()); ok {
			fields.Add(slog.Any("error", err))
		}
		code := status.Code(err)
		httpStatus := cloudstatus.ToHTTP(code)
		http.Error(w, http.StatusText(httpStatus), httpStatus)
		return
	}
}
