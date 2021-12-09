package handlerstwitter

import (
	"context"

	"github.com/javiyt/tweetgram/internal/bot"
	"github.com/javiyt/tweetgram/internal/handlers"
	"github.com/javiyt/tweetgram/internal/pubsub"
	"github.com/mailru/easyjson"
)

type Twitter struct {
	tc           bot.TwitterClient
	q            pubsub.Queue
	shouldNotify bool
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
	t := &Twitter{shouldNotify: true}

	for _, o := range options {
		o(t)
	}

	return t
}

func (t *Twitter) ID() string {
	return "twitter"
}

func (t *Twitter) ExecuteHandlers(ctx context.Context) {
	t.handleText(ctx)
	t.handlePhoto(ctx)
}

func (t *Twitter) StopNotifications() {
	t.shouldNotify = false
}

func (t *Twitter) handleText(ctx context.Context) {
	messages, err := t.q.Subscribe(ctx, pubsub.TextTopic.String())
	if err != nil {
		handlers.SendError(t.q, err)
	}

	go func() {
		for msg := range messages {
			if !t.shouldNotify {
				msg.Ack()

				continue
			}

			var m pubsub.TextEvent
			if err := easyjson.Unmarshal(msg.Payload, &m); err != nil {
				handlers.SendError(t.q, err)
				msg.Ack()

				continue
			}

			if err := t.tc.SendUpdate(m.Text); err != nil {
				handlers.SendError(t.q, err)
			}

			msg.Ack()
		}
	}()
}

func (t *Twitter) handlePhoto(ctx context.Context) {
	messages, err := t.q.Subscribe(ctx, pubsub.PhotoTopic.String())
	if err != nil {
		handlers.SendError(t.q, err)
	}

	go func() {
		for msg := range messages {
			if !t.shouldNotify {
				msg.Ack()

				continue
			}

			var m pubsub.PhotoEvent
			if err := easyjson.Unmarshal(msg.Payload, &m); err != nil {
				handlers.SendError(t.q, err)
				msg.Ack()

				continue
			}

			if err := t.tc.SendUpdateWithPhoto(m.Caption, m.FileContent); err != nil {
				handlers.SendError(t.q, err)
			}

			msg.Ack()
		}
	}()
}
