package handlers

import (
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/javiyt/tweettgram/internal/pubsub"
	"github.com/mailru/easyjson"
)

type EventHandler interface {
	ExecuteHandlers()
}

type Manager struct {
	hs []EventHandler
}

func NewHandlersManager(hs ...EventHandler) *Manager {
	return &Manager{hs: hs}
}

func (hm *Manager) StartHandlers() {
	for _, v := range hm.hs {
		v.ExecuteHandlers()
	}
}

func SendError(q pubsub.Queue, err error) {
	eb, _ := easyjson.Marshal(pubsub.ErrorEvent{Err: err.Error()})
	_ = q.Publish(pubsub.ErrorTopic.String(), message.NewMessage(watermill.NewUUID(), eb))
}
