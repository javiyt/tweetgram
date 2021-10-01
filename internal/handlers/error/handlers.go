package handlers_error

import (
	"context"

	"github.com/javiyt/tweetgram/internal/pubsub"
	"github.com/mailru/easyjson"
	"github.com/sirupsen/logrus"
)

type ErrorHandler struct {
	log *logrus.Logger
	q   pubsub.Queue
}

func NewErrorHandler(log *logrus.Logger, q pubsub.Queue) *ErrorHandler {
	return &ErrorHandler{log: log, q: q}
}

func (eh *ErrorHandler) ExecuteHandlers() {
	messages, err := eh.q.Subscribe(context.Background(), pubsub.ErrorTopic.String())
	if err != nil {
		eh.log.Error(err)
	}

	go func() {
		for msg := range messages {
			var m pubsub.ErrorEvent
			if err := easyjson.Unmarshal(msg.Payload, &m); err != nil {
				eh.log.Error(err)
				msg.Ack()
				continue
			}

			eh.log.Error(m.Err)
			msg.Ack()
		}
	}()
}