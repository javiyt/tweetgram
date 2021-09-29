package handlers_twitter

import (
	"context"

	"github.com/javiyt/tweettgram/internal/bot"
	"github.com/javiyt/tweettgram/internal/handlers"
	"github.com/javiyt/tweettgram/internal/pubsub"
	"github.com/mailru/easyjson"
)

type Twitter struct {
	tc bot.TwitterClient
	q   pubsub.Queue
}

func NewTwitter(tc bot.TwitterClient, q pubsub.Queue) *Twitter {
	return &Twitter{tc: tc, q: q}
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

			if err := t.tc.SendUpdateWithPhoto(m.Caption, m.FileContent); err !=nil {
				handlers.SendError(t.q, err)
				msg.Nack()
				continue
			}

			msg.Ack()
		}
	}()
}


