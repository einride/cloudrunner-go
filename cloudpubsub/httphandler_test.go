package cloudpubsub

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"cloud.google.com/go/pubsub/apiv1/pubsubpb"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gotest.tools/v3/assert"
)

func TestNewHTTPHandler(t *testing.T) {
	ctx := context.Background()
	// From: https://cloud.google.com/pubsub/docs/push#receiving_messages
	const example = `
		{
			"message": {
				"attributes": {
					"key": "value"
				},
				"data": "SGVsbG8gQ2xvdWQgUHViL1N1YiEgSGVyZSBpcyBteSBtZXNzYWdlIQ==",
				"messageId": "2070443601311540",
				"message_id": "2070443601311540",
				"publishTime": "2021-02-26T19:13:55.749Z",
				"publish_time": "2021-02-26T19:13:55.749Z"
			},
		   "subscription": "projects/myproject/subscriptions/mysubscription"
		}
	`
	expectedMessage := &pubsubpb.PubsubMessage{
		Data:       []byte("Hello Cloud Pub/Sub! Here is my message!"),
		Attributes: map[string]string{"key": "value"},
		MessageId:  "2070443601311540",
		PublishTime: &timestamppb.Timestamp{
			Seconds: 1614366835,
			Nanos:   749000000,
		},
	}
	var actualMessage *pubsubpb.PubsubMessage
	var subscription string
	var subscriptionOk bool
	fn := func(ctx context.Context, message *pubsubpb.PubsubMessage) error {
		actualMessage = message
		subscription, subscriptionOk = GetSubscription(ctx)
		return nil
	}
	server := httptest.NewServer(HTTPHandler(fn))
	defer server.Close()
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, server.URL, strings.NewReader(example))
	assert.NilError(t, err)
	response, err := http.DefaultClient.Do(request)
	assert.NilError(t, err)
	t.Cleanup(func() {
		assert.NilError(t, response.Body.Close())
	})
	assert.Equal(t, http.StatusOK, response.StatusCode)
	assert.DeepEqual(t, expectedMessage, actualMessage, protocmp.Transform())
	assert.Assert(t, subscriptionOk)
	assert.Equal(t, subscription, "projects/myproject/subscriptions/mysubscription")
}
