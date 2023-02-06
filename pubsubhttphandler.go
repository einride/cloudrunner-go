package cloudrunner

import (
	"context"
	"net/http"

	"cloud.google.com/go/pubsub/apiv1/pubsubpb"
	"go.einride.tech/cloudrunner/cloudpubsub"
)

// PubsubHTTPHandler creates a new HTTP handler for Cloud Pub/Sub push messages.
// See: https://cloud.google.com/pubsub/docs/push
func PubsubHTTPHandler(fn func(context.Context, *pubsubpb.PubsubMessage) error) http.Handler {
	return cloudpubsub.HTTPHandler(fn)
}
