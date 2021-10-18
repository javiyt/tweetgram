package handlers_twitter

import (
	"context"

	"github.com/javiyt/tweetgram/internal/bot"
	"github.com/javiyt/tweetgram/internal/handlers"
	"github.com/javiyt/tweetgram/internal/pubsub"
	"github.com/mailru/easyjson"
)

type Twitter struct {
	tc bot.TwitterClient
	q  pubsub.Queue
}

type Option func(b *Twitter)

func WithTwitterClient(tc bot.TwitterClient) Option {
	return func(t *Twitter) {
		t.tc = tc
	}
}

func WithQueue(q pubsub.Queue) Option {
	return func(t *Twitter) {
		t.q = q
	}
}

func NewTwitter(options ...Option) *Twitter {
	t := &Twitter{}

	for _, o := range options {
		o(t)
	}

	return t
}

func (t *Twitter) ExecuteHandlers() {
	t.handleText()
	t.handlePhoto()
}

func (t *Twitter) handleText() {
	messages, err := t.q.Subscribe(context.Background(), pubsub.TextTopic.String())
	if err != nil {
		handlers.SendError(t.q, err)
	}

	go func() {
		for msg := range messages {
			var m pubsub.TextEvent
			if err := easyjson.Unmarshal(msg.Payload, &m); err != nil {
				handlers.SendError(t.q, err)
				msg.Ack()
				continue
			}

			if err := t.tc.SendUpdate(m.Text); err != nil {
				handlers.SendError(t.q, err)
				msg.Nack()
				continue
			}

			msg.Ack()
		}
	}()
}

func (t *Twitter) handlePhoto() {
	messages, err := t.q.Subscribe(context.Background(), pubsub.PhotoTopic.String())
	if err != nil {
		handlers.SendError(t.q, err)
	}

	go func() {
		for msg := range messages {
			var m pubsub.PhotoEvent
			if err := easyjson.Unmarshal(msg.Payload, &m); err != nil {
				handlers.SendError(t.q, err)
				msg.Ack()
				continue
			}

			if err := t.tc.SendUpdateWithPhoto(m.Caption, m.FileContent); err != nil {
				handlers.SendError(t.q, err)
				msg.Nack()
				continue
			}

			msg.Ack()
		}
	}()
}
