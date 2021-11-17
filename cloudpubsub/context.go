package cloudpubsub

import "context"

type subscriptionContextKey struct{}

func withSubscription(ctx context.Context, subscription string) context.Context {
	return context.WithValue(ctx, subscriptionContextKey{}, subscription)
}

// GetSubscription gets the pubsub subscription from the current context.
func GetSubscription(ctx context.Context) (string, bool) {
	result, ok := ctx.Value(subscriptionContextKey{}).(string)
	return result, ok
}
