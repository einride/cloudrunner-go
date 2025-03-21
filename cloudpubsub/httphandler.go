package cloudpubsub

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"cloud.google.com/go/pubsub/apiv1/pubsubpb"
	"go.einride.tech/cloudrunner/cloudrequestlog"
	"go.einride.tech/cloudrunner/cloudstatus"
	"google.golang.org/grpc/status"
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
	var payload Payload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		if fields, ok := cloudrequestlog.GetAdditionalFields(r.Context()); ok {
			fields.Add(slog.Any("error", err))
		}
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	pubsubMessage := payload.BuildPubSubMessage()
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
