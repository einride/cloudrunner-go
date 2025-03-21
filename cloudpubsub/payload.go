package cloudpubsub

import (
	"time"

	"cloud.google.com/go/pubsub/apiv1/pubsubpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type PubSubMessage struct {
	Attributes  map[string]string `json:"attributes"`
	Data        []byte            `json:"data"`
	MessageID   string            `json:"messageId"`
	PublishTime time.Time         `json:"publishTime"`
	OrderingKey string            `json:"orderingKey"`
}

type Payload struct {
	Subscription string        `json:"subscription"`
	Message      PubSubMessage `json:"message"`
}

func (p Payload) IsValid() bool {
	return p.Message.MessageID != ""
}

func (p Payload) BuildPubSubMessage() pubsubpb.PubsubMessage {
	return pubsubpb.PubsubMessage{
		Data:        p.Message.Data,
		Attributes:  p.Message.Attributes,
		MessageId:   p.Message.MessageID,
		PublishTime: timestamppb.New(p.Message.PublishTime),
		OrderingKey: p.Message.OrderingKey,
	}
}
