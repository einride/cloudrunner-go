package cloudrunner

import (
	"context"
	"net/http"

	"go.einride.tech/cloudrunner/cloudpubsub"
	"google.golang.org/genproto/googleapis/pubsub/v1"
)

// PubsubHTTPHandler creates a new HTTP handler for Cloud Pub/Sub push messages.
// See: https://cloud.google.com/pubsub/docs/push
func PubsubHTTPHandler(fn func(context.Context, *pubsub.PubsubMessage) error) http.Handler {
	return cloudpubsub.HTTPHandler(fn)
}
