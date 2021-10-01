package handlers_error_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	hse "github.com/javiyt/tweetgram/internal/handlers/error"
	"github.com/javiyt/tweetgram/internal/pubsub"
	mq "github.com/javiyt/tweetgram/mocks/pubsub"
	"github.com/sirupsen/logrus"
	logrusTest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
)

func TestError_ExecuteHandlers(t *testing.T) {
	t.Run("it should fail getting channel for text notifications", func(t *testing.T) {
		mockedLogger, _ := logrusTest.NewNullLogger()
		mockedQueue := new(mq.Queue)

		th := hse.NewErrorHandler(mockedLogger, mockedQueue)

		mockedQueue.On("Subscribe", context.Background(), pubsub.ErrorTopic.String()).
			Once().
			Return(nil, errors.New("error getting channel error"))

		th.ExecuteHandlers()

		mockedQueue.AssertExpectations(t)
	})

	t.Run("it should fail unmarshaling text event", func(t *testing.T) {
		mockedLogger, hook := logrusTest.NewNullLogger()
		mockedQueue := new(mq.Queue)

		th := hse.NewErrorHandler(mockedLogger, mockedQueue)

		errorChannel := make(chan *message.Message)
		mockedQueue.On("Subscribe", context.Background(), pubsub.ErrorTopic.String()).
			Once().
			Return(func(context.Context, string) <-chan *message.Message {
				return errorChannel
			}, nil)

		th.ExecuteHandlers()
		newMessage := message.NewMessage(watermill.NewUUID(), []byte("{\"asd\":\"qwer"))
		errorChannel <- newMessage

		require.Eventually(t, func() bool {
			<-newMessage.Acked()
			return true
		}, time.Second, time.Millisecond)
		require.Equal(t, 1, len(hook.Entries))
		require.Equal(t, logrus.ErrorLevel, hook.LastEntry().Level)
		require.Equal(t, "parse error: unterminated string literal near offset 12 of '{\"asd\":\"qwer'", hook.LastEntry().Message)
	  
		hook.Reset()

		mockedQueue.AssertExpectations(t)
	})

	t.Run("it should log error message", func(t *testing.T) {
		mockedLogger, hook := logrusTest.NewNullLogger()
		mockedQueue := new(mq.Queue)

		th := hse.NewErrorHandler(mockedLogger, mockedQueue)

		errorChannel := make(chan *message.Message)
		mockedQueue.On("Subscribe", context.Background(), pubsub.ErrorTopic.String()).
			Once().
			Return(func(context.Context, string) <-chan *message.Message {
				return errorChannel
			}, nil)

		th.ExecuteHandlers()
		newMessage := message.NewMessage(watermill.NewUUID(), []byte("{\"error\":\"an error message\"}"))
		errorChannel <- newMessage

		require.Eventually(t, func() bool {
			<-newMessage.Acked()
			return true
		}, time.Second, time.Millisecond)
		require.Equal(t, 1, len(hook.Entries))
		require.Equal(t, logrus.ErrorLevel, hook.LastEntry().Level)
		require.Equal(t, "an error message", hook.LastEntry().Message)
	  
		hook.Reset()

		mockedQueue.AssertExpectations(t)
	})
}