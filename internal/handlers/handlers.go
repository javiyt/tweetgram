package handlers

import (
	"context"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/javiyt/tweetgram/internal/pubsub"
	"github.com/mailru/easyjson"
)

type EventHandler interface {
	ExecuteHandlers(context.Context)
}

type Manager struct {
	hs []EventHandler
}

func NewHandlersManager(hs ...EventHandler) *Manager {
	return &Manager{hs: hs}
}

func (hm *Manager) StartHandlers(ctx context.Context) {
	for _, v := range hm.hs {
		v.ExecuteHandlers(ctx)
	}
}

func SendError(q pubsub.Queue, err error) {
	eb, _ := easyjson.Marshal(pubsub.ErrorEvent{Err: err.Error()})
	_ = q.Publish(pubsub.ErrorTopic.String(), message.NewMessage(watermill.NewUUID(), eb))
}
