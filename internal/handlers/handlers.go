package handlers

import (
	"context"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/javiyt/tweetgram/internal/pubsub"
	"github.com/mailru/easyjson"
)

type EventHandler interface {
	ID() string
	ExecuteHandlers(context.Context)
	StopNotifications()
}

type Manager struct {
	q  pubsub.Queue
	hs []EventHandler
}

func NewHandlersManager(q pubsub.Queue, hs ...EventHandler) *Manager {
	return &Manager{q: q, hs: hs}
}

func (hm *Manager) StartHandlers(ctx context.Context) {
	for _, v := range hm.hs {
		v.ExecuteHandlers(ctx)
	}

	hm.StopNotifications(ctx)
}

func (hm *Manager) StopNotifications(ctx context.Context) {
	messages, err := hm.q.Subscribe(ctx, pubsub.CommandTopic.String())
	if err != nil {
		return
	}

	go func() {
		for msg := range messages {
			var m pubsub.CommandEvent
			if err := easyjson.Unmarshal(msg.Payload, &m); err != nil {
				SendError(hm.q, err)
				msg.Ack()

				continue
			}

			if m.Command == pubsub.StopCommand {
				for i := range hm.hs {
					if m.Handler == hm.hs[i].ID() || m.Handler == "" {
						hm.hs[i].StopNotifications()
					}
				}
			}

			msg.Ack()
		}
	}()
}

func SendError(q pubsub.Queue, err error) {
	eb, _ := easyjson.Marshal(pubsub.ErrorEvent{Err: err.Error()})
	_ = q.Publish(pubsub.ErrorTopic.String(), message.NewMessage(watermill.NewUUID(), eb))
}
