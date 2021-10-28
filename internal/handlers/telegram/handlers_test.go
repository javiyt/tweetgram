package handlerstelegram_test

import (
	"context"
	"testing"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/javiyt/tweetgram/internal/config"
	ht "github.com/javiyt/tweetgram/internal/handlers/telegram"
	"github.com/javiyt/tweetgram/internal/pubsub"
	mb "github.com/javiyt/tweetgram/mocks/bot"
	mq "github.com/javiyt/tweetgram/mocks/pubsub"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	tb "gopkg.in/tucnak/telebot.v2"
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
	cfg := config.EnvConfig{
		BroadcastChannel: 1234,
	}

	t.Run("it should fail getting channel for text and photo notifications", func(t *testing.T) {
		th, mockedQueue, _, _, _ := generateHandlerAndMocks(cfg, false)

		mockedQueue.On("Subscribe", context.Background(), pubsub.TextTopic.String()).
			Once().
			Return(nil, gettingChannelError{})
		mockedQueue.On("Subscribe", context.Background(), pubsub.PhotoTopic.String()).
			Once().
			Return(nil, gettingChannelError{})
		mockedQueue.On("Publish", pubsub.ErrorTopic.String(), mock.MatchedBy(func(m *message.Message) bool {
			return string(m.Payload) == "{\"error\":\"error getting channel error\"}"
		})).Times(2).
			Return(nil)

		th.ExecuteHandlers()

		mockedQueue.AssertExpectations(t)
	})
}

func TestTelegram_ExecuteHandlersText(t *testing.T) {
	cfg := config.EnvConfig{
		BroadcastChannel: 1234,
	}

	t.Run("it should fail unmarshaling text event", func(t *testing.T) {
		th, mockedQueue, _, textChannel, _ := generateHandlerAndMocks(cfg, true)

		mockedQueue.On("Publish", pubsub.ErrorTopic.String(), mock.MatchedBy(func(m *message.Message) bool {
			return string(m.Payload) ==
				"{\"error\":\"parse error: unterminated string literal near offset 12 of '{\\\"asd\\\":\\\"qwer'\"}"
		})).Once().
			Return(nil)

		th.ExecuteHandlers()

		sendMessageToChannel(t, textChannel, []byte("{\"asd\":\"qwer"), true)

		mockedQueue.AssertExpectations(t)
	})

	t.Run("it should fail sending text message to telegram", func(t *testing.T) {
		th, mockedQueue, mockedBot, textChannel, _ := generateHandlerAndMocks(cfg, true)

		mockedQueue.On("Publish", pubsub.ErrorTopic.String(), mock.MatchedBy(func(m *message.Message) bool {
			return string(m.Payload) == "{\"error\":\"couldn't send message to telegram\"}"
		})).Once().
			Return(nil)
		mockedBot.On("Send", tb.ChatID(cfg.BroadcastChannel), "testing message").
			Once().
			Return(nil, messageNotSendError{})

		th.ExecuteHandlers()
		sendMessageToChannel(t, textChannel, []byte("{\"text\":\"testing message\"}"), false)

		mockedQueue.AssertExpectations(t)
		mockedBot.AssertExpectations(t)
	})

	t.Run("it should send text message to telegram", func(t *testing.T) {
		th, mockedQueue, mockedBot, textChannel, _ := generateHandlerAndMocks(cfg, true)

		mockedBot.On("Send", tb.ChatID(cfg.BroadcastChannel), "testing message").
			Once().
			Return(nil, nil)

		th.ExecuteHandlers()

		sendMessageToChannel(t, textChannel, []byte("{\"text\":\"testing message\"}"), true)

		mockedQueue.AssertExpectations(t)
		mockedBot.AssertExpectations(t)
	})
}

func TestTelegram_ExecuteHandlersPhoto(t *testing.T) {
	cfg := config.EnvConfig{
		BroadcastChannel: 1234,
	}
	eventMsg := []byte("{\"caption\":\"testing message\",\"file_id\":\"blablabla\",\"file_url\":\"http://photo.url\"," +
		"\"file_size\":1234}")

	t.Run("it should fail unmarshaling photo event", func(t *testing.T) {
		th, mockedQueue, _, _, photoChannel := generateHandlerAndMocks(cfg, true)

		mockedQueue.On("Publish", pubsub.ErrorTopic.String(), mock.MatchedBy(func(m *message.Message) bool {
			return string(m.Payload) ==
				"{\"error\":\"parse error: unterminated string literal near offset 12 of '{\\\"asd\\\":\\\"qwer'\"}"
		})).Once().
			Return(nil)

		th.ExecuteHandlers()

		sendMessageToChannel(t, photoChannel, []byte("{\"asd\":\"qwer"), true)

		mockedQueue.AssertExpectations(t)
	})

	t.Run("it should fail sending photo message to telegram", func(t *testing.T) {
		th, mockedQueue, mockedBot, _, photoChannel := generateHandlerAndMocks(cfg, true)

		mockedQueue.On("Publish", pubsub.ErrorTopic.String(), mock.MatchedBy(func(m *message.Message) bool {
			return string(m.Payload) == "{\"error\":\"couldn't send message to telegram\"}"
		})).Once().
			Return(nil)
		mockedBot.On("Send", tb.ChatID(cfg.BroadcastChannel), mock.MatchedBy(matchTelegramPhoto())).
			Once().Return(nil, messageNotSendError{})

		th.ExecuteHandlers()

		sendMessageToChannel(t, photoChannel, eventMsg, false)

		mockedQueue.AssertExpectations(t)
		mockedBot.AssertExpectations(t)
	})

	t.Run("it should send photo message to telegram", func(t *testing.T) {
		th, mockedQueue, mockedBot, _, photoChannel := generateHandlerAndMocks(cfg, true)

		mockedBot.On("Send", tb.ChatID(cfg.BroadcastChannel), mock.MatchedBy(matchTelegramPhoto())).
			Once().Return(nil, nil)

		th.ExecuteHandlers()
		sendMessageToChannel(t, photoChannel, eventMsg, true)

		mockedQueue.AssertExpectations(t)
		mockedBot.AssertExpectations(t)
	})
}

func generateHandlerAndMocks(cfg config.EnvConfig, returnChannels bool) (
	*ht.Telegram,
	*mq.Queue,
	*mb.TelegramBot,
	chan *message.Message,
	chan *message.Message,
) {
	mockedBot := new(mb.TelegramBot)
	mockedQueue := new(mq.Queue)

	th := ht.NewTelegram(ht.WithConfig(cfg), ht.WithTelegramBot(mockedBot), ht.WithQueue(mockedQueue))

	textChannel := make(chan *message.Message)
	photoChannel := make(chan *message.Message)

	if returnChannels {
		mockedQueue.On("Subscribe", context.Background(), pubsub.TextTopic.String()).
			Once().
			Return(func(context.Context, string) <-chan *message.Message {
				return textChannel
			}, nil)
		mockedQueue.On("Subscribe", context.Background(), pubsub.PhotoTopic.String()).
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
			p  *tb.Photo
			ok bool
		)

		if p, ok = m.(*tb.Photo); !ok {
			return false
		}

		return p.Caption == "testing message" &&
			p.FileID == "blablabla" &&
			p.FileURL == "http://photo.url" &&
			p.FileSize == 1234
	}
}
