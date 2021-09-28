package handlers_telegram

import (
	"context"

	"github.com/javiyt/tweettgram/internal/bot"
	"github.com/javiyt/tweettgram/internal/config"
	"github.com/javiyt/tweettgram/internal/handlers"
	"github.com/javiyt/tweettgram/internal/pubsub"
	"github.com/mailru/easyjson"
	tb "gopkg.in/tucnak/telebot.v2"
)

type Telegram struct {
	bot bot.TelegramBot
	cfg config.EnvConfig
	q   pubsub.Queue
}

func NewTelegram(cfg config.EnvConfig, bot bot.TelegramBot, q pubsub.Queue) *Telegram {
	return &Telegram{bot: bot, cfg: cfg, q: q}
}

func (t *Telegram) ExecuteHandlers() {
	t.handleText()
	t.handlePhoto()
}

func (t *Telegram) handleText() {
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

			if _, err := t.bot.Send(tb.ChatID(t.cfg.BroadcastChannel), m.Text); err != nil {
				handlers.SendError(t.q, err)
				msg.Nack()
				continue
			}

			msg.Ack()
		}
	}()
}

func (t *Telegram) handlePhoto() {
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

			if _, err := t.bot.Send(tb.ChatID(t.cfg.BroadcastChannel), &tb.Photo{
				Caption: m.Caption,
				File: tb.File{
					FileID:   m.FileID,
					FileURL:  m.FileURL,
					FileSize: m.FileSize,
				},
			}); err != nil {
				handlers.SendError(t.q, err)
				msg.Nack()
				continue
			}

			msg.Ack()
		}
	}()
}
