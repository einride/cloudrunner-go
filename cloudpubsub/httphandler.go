package cloudpubsub

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"cloud.google.com/go/pubsub/v2/apiv1/pubsubpb"
	"go.einride.tech/cloudrunner/cloudrequestlog"
	"go.einride.tech/cloudrunner/cloudstatus"
	"google.golang.org/api/idtoken"
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

		if payload.Issuer != "accounts.google.com" && payload.Issuer != "https://accounts.google.com" {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		if len(allowedEmails) > 0 {
			email, ok := payload.Claims["email"].(string)
			if !ok || payload.Claims["email_verified"] != true {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}
			if _, found := emails[email]; !found {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}
		}

		// Authenticated, pass along to handler.
		handler.ServeHTTP(w, r)
	})
}
