package pubsub

import (
	"context"

	"github.com/ThreeDotsLabs/watermill/message"
)

type TopicName int

const (
	ErrorTopic TopicName = iota
	PhotoTopic
	TextTopic
)

type Queue interface {
	Publish(topic string, messages ...*message.Message) error
	Subscribe(ctx context.Context, topic string) (<-chan *message.Message, error)
	Close() error
}

//easyjson:json
type ErrorEvent struct {
	Err string `json:"error"`
}

//easyjson:json
type PhotoEvent struct {
	Caption     string `json:"caption"`
	FileID      string `json:"fileId"`
	FileURL     string `json:"fileUrl"`
	FileSize    int    `json:"fileSize"`
	FileContent []byte `json:"fileContent"`
}

//easyjson:json
type TextEvent struct {
	Text string `json:"text"`
}
