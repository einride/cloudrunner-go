package cloudpubsub

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"cloud.google.com/go/pubsub/apiv1/pubsubpb"
	"go.einride.tech/cloudrunner/cloudrequestlog"
	"go.einride.tech/cloudrunner/cloudstatus"
	"go.einride.tech/cloudrunner/cloudzap"
	"go.uber.org/zap"
	"google.golang.org/api/idtoken"
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
			fields.Add(zap.Error(err))
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
		fields.Add(cloudzap.ProtoMessage("pubsubMessage", &pubsubMessage))
	}
	ctx := withSubscription(r.Context(), payload.Subscription)
	if err := fn(ctx, &pubsubMessage); err != nil {
		if fields, ok := cloudrequestlog.GetAdditionalFields(r.Context()); ok {
			fields.Add(zap.Error(err))
		}
		code := status.Code(err)
		httpStatus := cloudstatus.ToHTTP(code)
		http.Error(w, http.StatusText(httpStatus), httpStatus)
		return
	}
}

// AuthenticatedHTTPHandler creates a new HTTP handler for authenticated Cloud Pub/Sub push messages, and verifies the
// token passed in the Authorization header.
// See: https://cloud.google.com/pubsub/docs/authenticate-push-subscriptions
//
// The audience parameter is optional and only verified against the token audience claim if non-empty.
// If audience isn't configured in the push subscription  configuration, it defaults to the push endpoint URL.
// See: https://cloud.google.com/pubsub/docs/reference/rest/v1/projects.subscriptions#oidctoken
//
// The allowedEmails list is optional, and if non-empty will verify that it contains the token email claim.
func AuthenticatedHTTPHandler(
	fn func(context.Context, *pubsubpb.PubsubMessage) error,
	audience string,
	allowedEmails ...string,
) http.Handler {
	emails := make(map[string]struct{}, len(allowedEmails))
	for _, email := range allowedEmails {
		emails[email] = struct{}{}
	}

	handler := httpHandlerFn(fn)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Inspired by https://cloud.google.com/pubsub/docs/authenticate-push-subscriptions#go,
		// but always return HTTP 401 for all authentication errors.
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		headerParts := strings.Split(authHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		token := headerParts[1]
		payload, err := idtoken.Validate(r.Context(), token, audience)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		switch payload.Issuer {
		case "accounts.google.com", "https://accounts.google.com":
		default:
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		email, ok := payload.Claims["email"].(string)
		if !ok || payload.Claims["email_verified"] != true {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		if _, found := emails[email]; len(emails) > 0 && !found {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		// Authenticated, pass along to handler.
		handler.ServeHTTP(w, r)
	})
}
