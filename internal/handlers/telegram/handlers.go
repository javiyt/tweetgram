package handlers_telegram

import (
	"context"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/javiyt/tweettgram/internal/bot"
	"github.com/javiyt/tweettgram/internal/config"
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
		t.sendError(err)
	}

	go func() {
		for msg := range messages {
			var m pubsub.TextEvent
			if err := easyjson.Unmarshal(msg.Payload, &m); err != nil {
				t.sendError(err)
				msg.Nack()
				continue
			}

			if _, err := t.bot.Send(tb.ChatID(t.cfg.BroadcastChannel), &tb.Message{Text: m.Text}); err != nil {
				t.sendError(err)
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
		t.sendError(err)
	}

	go func() {
		for msg := range messages {
			var m pubsub.PhotoEvent
			if err := easyjson.Unmarshal(msg.Payload, &m); err != nil {
				t.sendError(err)
				msg.Nack()
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
				t.sendError(err)
				msg.Nack()
				continue
			}

			msg.Ack()
		}
	}()
}

func (t *Telegram) sendError(err error) {
	eb, _ := easyjson.Marshal(pubsub.ErrorEvent{Err: err.Error()})
	_ = t.q.Publish(pubsub.ErrorTopic.String(), message.NewMessage(watermill.NewUUID(), eb))
}
