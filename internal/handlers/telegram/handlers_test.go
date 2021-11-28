package handlerstelegram_test

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/javiyt/tweetgram/internal/bot"
	"github.com/javiyt/tweetgram/internal/config"
	ht "github.com/javiyt/tweetgram/internal/handlers/telegram"
	"github.com/javiyt/tweetgram/internal/pubsub"
	mb "github.com/javiyt/tweetgram/mocks/bot"
	mq "github.com/javiyt/tweetgram/mocks/pubsub"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type messageNotSendError struct{}

func (m messageNotSendError) Error() string {
	return "couldn't send message to telegram"
}

type gettingChannelError struct{}

func (m gettingChannelError) Error() string {
	return "error getting channel error"
}

func TestTelegram_ExecuteHandlers(t *testing.T) {
	cfg := config.AppConfig{
		BroadcastChannel: 1234,
	}
	ctx := context.Background()

	t.Run("it should fail getting channel for text and photo notifications", func(t *testing.T) {
		th, mockedQueue, _, _, _ := generateHandlerAndMocks(ctx, cfg, false)

		mockedQueue.On("Subscribe", ctx, pubsub.TextTopic.String()).
			Once().
			Return(nil, gettingChannelError{})
		mockedQueue.On("Subscribe", ctx, pubsub.PhotoTopic.String()).
			Once().
			Return(nil, gettingChannelError{})
		mockedQueue.On("Publish", pubsub.ErrorTopic.String(), mock.MatchedBy(func(m *message.Message) bool {
			return string(m.Payload) == "{\"error\":\"error getting channel error\"}"
		})).Times(2).
			Return(nil)

		th.ExecuteHandlers(ctx)

		mockedQueue.AssertExpectations(t)
	})
}

func TestTelegram_ExecuteHandlersText(t *testing.T) {
	cfg := config.AppConfig{
		BroadcastChannel: 1234,
	}
	ctx := context.Background()

	t.Run("it should fail unmarshaling text event", func(t *testing.T) {
		th, mockedQueue, _, textChannel, _ := generateHandlerAndMocks(ctx, cfg, true)

		mockedQueue.On("Publish", pubsub.ErrorTopic.String(), mock.MatchedBy(func(m *message.Message) bool {
			return string(m.Payload) ==
				"{\"error\":\"parse error: unterminated string literal near offset 12 of '{\\\"asd\\\":\\\"qwer'\"}"
		})).Once().
			Return(nil)

		th.ExecuteHandlers(ctx)

		sendMessageToChannel(t, textChannel, []byte("{\"asd\":\"qwer"), true)

		mockedQueue.AssertExpectations(t)
	})

	t.Run("it should fail sending text message to telegram", func(t *testing.T) {
		th, mockedQueue, mockedBot, textChannel, _ := generateHandlerAndMocks(ctx, cfg, true)

		mockedQueue.On("Publish", pubsub.ErrorTopic.String(), mock.MatchedBy(func(m *message.Message) bool {
			return string(m.Payload) == "{\"error\":\"couldn't send message to telegram\"}"
		})).Once().
			Return(nil)
		mockedBot.On("Send", strconv.Itoa(int(cfg.BroadcastChannel)), "failing message").
			Once().
			Return(messageNotSendError{})

		th.ExecuteHandlers(ctx)
		sendMessageToChannel(t, textChannel, []byte("{\"text\":\"failing message\"}"), false)

		mockedQueue.AssertExpectations(t)
		mockedBot.AssertExpectations(t)
	})

	t.Run("it should send text message to telegram", func(t *testing.T) {
		th, mockedQueue, mockedBot, textChannel, _ := generateHandlerAndMocks(ctx, cfg, true)

		mockedBot.On("Send", strconv.Itoa(int(cfg.BroadcastChannel)), "testing message").
			Once().
			Return(nil, nil)

		th.ExecuteHandlers(ctx)

		sendMessageToChannel(t, textChannel, []byte("{\"text\":\"testing message\"}"), true)

		mockedQueue.AssertExpectations(t)
		mockedBot.AssertExpectations(t)
	})
}

func TestTelegram_ExecuteHandlersPhoto(t *testing.T) {
	cfg := config.AppConfig{
		BroadcastChannel: 1234,
	}
	eventMsg := []byte("{\"caption\":\"testing message\",\"fileId\":\"blablabla\",\"fileUrl\":\"http://photo.url\"," +
		"\"fileSize\":1234}")
	ctx := context.Background()

	t.Run("it should fail unmarshaling photo event", func(t *testing.T) {
		th, mockedQueue, _, _, photoChannel := generateHandlerAndMocks(ctx, cfg, true)

		mockedQueue.On("Publish", pubsub.ErrorTopic.String(), mock.MatchedBy(func(m *message.Message) bool {
			return string(m.Payload) ==
				"{\"error\":\"parse error: unterminated string literal near offset 12 of '{\\\"asd\\\":\\\"qwer'\"}"
		})).Once().
			Return(nil)

		th.ExecuteHandlers(ctx)

		sendMessageToChannel(t, photoChannel, []byte("{\"asd\":\"qwer"), true)

		mockedQueue.AssertExpectations(t)
	})

	t.Run("it should fail sending photo message to telegram", func(t *testing.T) {
		th, mockedQueue, mockedBot, _, photoChannel := generateHandlerAndMocks(ctx, cfg, true)

		mockedQueue.On("Publish", pubsub.ErrorTopic.String(), mock.MatchedBy(func(m *message.Message) bool {
			return string(m.Payload) == "{\"error\":\"couldn't send message to telegram\"}"
		})).Once().
			Return(nil)
		mockedBot.On("Send", strconv.Itoa(int(cfg.BroadcastChannel)), mock.MatchedBy(matchTelegramPhoto())).
			Once().Return(messageNotSendError{})

		th.ExecuteHandlers(ctx)

		sendMessageToChannel(t, photoChannel, eventMsg, false)

		mockedQueue.AssertExpectations(t)
		mockedBot.AssertExpectations(t)
	})

	t.Run("it should send photo message to telegram", func(t *testing.T) {
		th, mockedQueue, mockedBot, _, photoChannel := generateHandlerAndMocks(ctx, cfg, true)

		mockedBot.On("Send", strconv.Itoa(int(cfg.BroadcastChannel)), mock.MatchedBy(matchTelegramPhoto())).
			Once().Return(nil)

		th.ExecuteHandlers(ctx)
		sendMessageToChannel(t, photoChannel, eventMsg, true)

		mockedQueue.AssertExpectations(t)
		mockedBot.AssertExpectations(t)
	})
}

func generateHandlerAndMocks(
	ctx context.Context,
	cfg config.AppConfig,
	returnChannels bool,
) (*ht.Telegram, *mq.Queue, *mb.TelegramBot, chan *message.Message, chan *message.Message) {
	mockedBot := new(mb.TelegramBot)
	mockedQueue := new(mq.Queue)

	th := ht.NewTelegram(ht.WithAppConfig(cfg), ht.WithTelegramBot(mockedBot), ht.WithQueue(mockedQueue))

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

	return th, mockedQueue, mockedBot, textChannel, photoChannel
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

func matchTelegramPhoto() func(m interface{}) bool {
	return func(m interface{}) bool {
		var (
			p  *bot.TelegramPhoto
			ok bool
		)

		if p, ok = m.(*bot.TelegramPhoto); !ok {
			return false
		}

		return p.Caption == "testing message" &&
			p.FileID == "blablabla" &&
			p.FileURL == "http://photo.url" &&
			p.FileSize == 1234
	}
}
