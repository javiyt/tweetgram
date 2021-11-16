package handlerstwitter_test

import (
	"context"
	"testing"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	ht "github.com/javiyt/tweetgram/internal/handlers/twitter"
	"github.com/javiyt/tweetgram/internal/pubsub"
	mb "github.com/javiyt/tweetgram/mocks/bot"
	mq "github.com/javiyt/tweetgram/mocks/pubsub"
	"github.com/mailru/easyjson"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type messageNotSendError struct{}

func (m messageNotSendError) Error() string {
	return "couldn't send message to twitter"
}

type channelError struct{}

func (c channelError) Error() string {
	return "error getting channel error"
}

func TestTwitter_ExecuteHandlers(t *testing.T) {
	t.Run("it should fail getting channel for text and photo notifications", func(t *testing.T) {
		ctx := context.Background()

		th, mockedQueue, _, _, _ := getTwitterHandlerAndMocks(ctx, false)

		mockedQueue.On("Subscribe", context.Background(), pubsub.TextTopic.String()).
			Once().
			Return(nil, channelError{})
		mockedQueue.On("Subscribe", context.Background(), pubsub.PhotoTopic.String()).
			Once().
			Return(nil, channelError{})
		mockedQueue.On("Publish", pubsub.ErrorTopic.String(), mock.MatchedBy(func(m *message.Message) bool {
			return string(m.Payload) == "{\"error\":\"error getting channel error\"}"
		})).Times(2).
			Return(nil)

		th.ExecuteHandlers(ctx)

		mockedQueue.AssertExpectations(t)
	})
}

func TestTwitter_ExecuteHandlersText(t *testing.T) {
	ctx := context.Background()

	t.Run("it should fail unmarshaling text event", func(t *testing.T) {
		th, mockedQueue, _, textChannel, _ := getTwitterHandlerAndMocks(ctx, true)

		mockedQueue.On("Publish", pubsub.ErrorTopic.String(), mock.MatchedBy(func(m *message.Message) bool {
			return string(m.Payload) ==
				"{\"error\":\"parse error: unterminated string literal near offset 12 of '{\\\"asd\\\":\\\"qwer'\"}"
		})).Once().
			Return(nil)

		th.ExecuteHandlers(ctx)

		sendMessageToChannel(t, textChannel, []byte("{\"asd\":\"qwer"), true)

		mockedQueue.AssertExpectations(t)
	})

	t.Run("it should fail sending text message to twitter", func(t *testing.T) {
		th, mockedQueue, mockedTwitter, textChannel, _ := getTwitterHandlerAndMocks(ctx, true)

		mockedQueue.On(
			"Publish",
			pubsub.ErrorTopic.String(),
			mock.MatchedBy(func(m *message.Message) bool {
				return string(m.Payload) == "{\"error\":\"couldn't send message to twitter\"}"
			}),
		).Once().
			Return(nil)
		mockedTwitter.On("SendUpdate", "testing message").
			Once().
			Return(messageNotSendError{})

		th.ExecuteHandlers(ctx)

		sendMessageToChannel(t, textChannel, []byte("{\"text\":\"testing message\"}"), false)

		mockedQueue.AssertExpectations(t)
		mockedTwitter.AssertExpectations(t)
	})

	t.Run("it should send text message to twitter", func(t *testing.T) {
		th, mockedQueue, mockedTwitter, textChannel, _ := getTwitterHandlerAndMocks(ctx, true)

		mockedTwitter.On("SendUpdate", "testing message").Once().Return(nil)

		th.ExecuteHandlers(ctx)

		sendMessageToChannel(t, textChannel, []byte("{\"text\":\"testing message\"}"), true)

		mockedQueue.AssertExpectations(t)
		mockedTwitter.AssertExpectations(t)
	})
}

func TestTwitter_ExecuteHandlersPhoto(t *testing.T) {
	photoContent := []byte("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAIAAACQd1PeAAAAEElEQVR4nGKaks0ECAAA//" +
		"8CoAEEsZgdLgAAAABJRU5ErkJggg==")

	bytes, _ := easyjson.Marshal(pubsub.PhotoEvent{
		Caption:     "testing caption",
		FileID:      "123456789asdfg",
		FileURL:     "http://photo.url",
		FileSize:    12345,
		FileContent: photoContent,
	})

	t.Run("it should fail unmarshaling photo event", func(t *testing.T) {
		th, mockedQueue, _, _, photoChannel := getTwitterHandlerAndMocks(context.Background(), true)

		mockedQueue.On("Publish", pubsub.ErrorTopic.String(), mock.MatchedBy(func(m *message.Message) bool {
			return string(m.Payload) ==
				"{\"error\":\"parse error: unterminated string literal near offset 12 of '{\\\"asd\\\":\\\"qwer'\"}"
		})).Once().Return(nil)

		th.ExecuteHandlers(context.Background())

		sendMessageToChannel(t, photoChannel, []byte("{\"asd\":\"qwer"), true)

		mockedQueue.AssertExpectations(t)
	})

	t.Run("it should fail sending photo to twitter", func(t *testing.T) {
		th, mockedQueue, mockedTwitter, _, photoChannel := getTwitterHandlerAndMocks(context.Background(), true)

		mockedQueue.On(
			"Publish",
			pubsub.ErrorTopic.String(),
			mock.MatchedBy(func(m *message.Message) bool {
				return string(m.Payload) == "{\"error\":\"couldn't send message to twitter\"}"
			}),
		).Once().Return(nil)
		mockedTwitter.On("SendUpdateWithPhoto", "testing caption", photoContent).
			Once().Return(messageNotSendError{})

		th.ExecuteHandlers(context.Background())

		sendMessageToChannel(t, photoChannel, bytes, false)

		mockedQueue.AssertExpectations(t)
		mockedTwitter.AssertExpectations(t)
	})

	t.Run("it should send photo to twitter", func(t *testing.T) {
		th, mockedQueue, mockedTwitter, _, photoChannel := getTwitterHandlerAndMocks(context.Background(), true)

		mockedTwitter.On("SendUpdateWithPhoto", "testing caption", photoContent).
			Once().Return(nil)

		th.ExecuteHandlers(context.Background())

		sendMessageToChannel(t, photoChannel, bytes, true)

		mockedQueue.AssertExpectations(t)
		mockedTwitter.AssertExpectations(t)
	})
}

func getTwitterHandlerAndMocks(ctx context.Context, returnChannels bool) (
	*ht.Twitter,
	*mq.Queue,
	*mb.TwitterClient,
	chan *message.Message,
	chan *message.Message,
) {
	mockedTwitter := new(mb.TwitterClient)
	mockedQueue := new(mq.Queue)

	th := ht.NewTwitter(ht.WithTwitterClient(mockedTwitter), ht.WithQueue(mockedQueue))

	textChannel := make(chan *message.Message)
	photoChannel := make(chan *message.Message)

	if returnChannels {
		mockedQueue.On("Subscribe", ctx, pubsub.TextTopic.String()).
			Once().
			Return(func(context.Context, string) <-chan *message.Message {
				return textChannel
			}, nil)
		mockedQueue.On("Subscribe", ctx, pubsub.PhotoTopic.String()).
			Once().
			Return(func(context.Context, string) <-chan *message.Message {
				return photoChannel
			}, nil)
	}

	return th, mockedQueue, mockedTwitter, textChannel, photoChannel
}

func sendMessageToChannel(t *testing.T, channel chan *message.Message, eventMsg []byte, acked bool) {
	newMessage := message.NewMessage(watermill.NewUUID(), eventMsg)
	channel <- newMessage

	require.Eventually(t, func() bool {
		if acked {
			<-newMessage.Acked()
		} else {
			<-newMessage.Nacked()
		}

		return true
	}, time.Second, time.Millisecond)
}
