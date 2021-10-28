package handlerserror_test

import (
	"context"
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

type gettingChannelError struct{}

func (m gettingChannelError) Error() string {
	return "error getting channel error"
}

func TestError_ExecuteHandlers(t *testing.T) {
	t.Run("it should fail getting channel for text notifications", func(t *testing.T) {
		_, mockedQueue, th, _ := generateMocksAndErrorChannel()

		mockedQueue.On("Subscribe", context.Background(), pubsub.ErrorTopic.String()).
			Once().
			Return(nil, gettingChannelError{})

		th.ExecuteHandlers()

		mockedQueue.AssertExpectations(t)
	})

	t.Run("it should fail unmarshaling text event", func(t *testing.T) {
		hook, mockedQueue, th, errorChannel := generateMocksAndErrorChannel()

		mockedQueue.On("Subscribe", context.Background(), pubsub.ErrorTopic.String()).
			Once().
			Return(func(context.Context, string) <-chan *message.Message {
				return errorChannel
			}, nil)

		th.ExecuteHandlers()
		sendMessageToChannel(t, errorChannel, []byte("{\"asd\":\"qwer"))

		assertLogMessage(
			t,
			hook,
			"parse error: unterminated string literal near offset 12 of '{\"asd\":\"qwer'",
		)
		mockedQueue.AssertExpectations(t)
	})

	t.Run("it should log error message", func(t *testing.T) {
		hook, mockedQueue, th, errorChannel := generateMocksAndErrorChannel()
		mockedQueue.On("Subscribe", context.Background(), pubsub.ErrorTopic.String()).
			Once().
			Return(func(context.Context, string) <-chan *message.Message {
				return errorChannel
			}, nil)

		th.ExecuteHandlers()
		sendMessageToChannel(t, errorChannel, []byte("{\"error\":\"an error message\"}"))

		assertLogMessage(t, hook, "an error message")
		mockedQueue.AssertExpectations(t)
	})
}

func generateMocksAndErrorChannel() (*logrusTest.Hook, *mq.Queue, *hse.ErrorHandler, chan *message.Message) {
	mockedLogger, hook := logrusTest.NewNullLogger()
	mockedQueue := new(mq.Queue)

	th := hse.NewErrorHandler(mockedLogger, mockedQueue)

	errorChannel := make(chan *message.Message)

	return hook, mockedQueue, th, errorChannel
}

func sendMessageToChannel(t *testing.T, errorChannel chan *message.Message, errMsg []byte) {
	newMessage := message.NewMessage(watermill.NewUUID(), errMsg)
	errorChannel <- newMessage

	require.Eventually(t, func() bool {
		<-newMessage.Acked()
		return true
	}, time.Second, time.Millisecond)
}

func assertLogMessage(t *testing.T, hook *logrusTest.Hook, logMsg string) {
	require.Equal(t, 1, len(hook.Entries))
	require.Equal(t, logrus.ErrorLevel, hook.LastEntry().Level)
	require.Equal(
		t,
		logMsg,
		hook.LastEntry().Message,
	)

	hook.Reset()
}
