package handlerstelegram

import (
	"context"
	"strconv"

	"github.com/javiyt/tweetgram/internal/bot"
	"github.com/javiyt/tweetgram/internal/config"
	"github.com/javiyt/tweetgram/internal/handlers"
	"github.com/javiyt/tweetgram/internal/pubsub"
	"github.com/mailru/easyjson"
)

type Telegram struct {
	bot bot.TelegramBot
	cfg config.EnvConfig
	q   pubsub.Queue
}

type Option func(b *Telegram)

func WithTelegramBot(tb bot.TelegramBot) Option {
	return func(b *Telegram) {
		b.bot = tb
	}
}

func WithConfig(cfg config.EnvConfig) Option {
	return func(b *Telegram) {
		b.cfg = cfg
	}
}

func WithQueue(q pubsub.Queue) Option {
	return func(b *Telegram) {
		b.q = q
	}
}

func NewTelegram(options ...Option) *Telegram {
	t := &Telegram{}

	for _, o := range options {
		o(t)
	}

	return t
}

func (t *Telegram) ExecuteHandlers(ctx context.Context) {
	t.handleText(ctx)
	t.handlePhoto(ctx)
}

func (t *Telegram) handleText(ctx context.Context) {
	messages, err := t.q.Subscribe(ctx, pubsub.TextTopic.String())
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

			if err := t.bot.Send(strconv.Itoa(int(t.cfg.BroadcastChannel)), m.Text); err != nil {
				handlers.SendError(t.q, err)
				msg.Nack()

				continue
			}

			msg.Ack()
		}
	}()
}

func (t *Telegram) handlePhoto(ctx context.Context) {
	messages, err := t.q.Subscribe(ctx, pubsub.PhotoTopic.String())
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

			if err := t.bot.Send(strconv.Itoa(int(t.cfg.BroadcastChannel)), &bot.TelegramPhoto{
				Caption:  m.Caption,
				FileID:   m.FileID,
				FileURL:  m.FileURL,
				FileSize: m.FileSize,
			}); err != nil {
				handlers.SendError(t.q, err)
				msg.Nack()

				continue
			}

			msg.Ack()
		}
	}()
}
