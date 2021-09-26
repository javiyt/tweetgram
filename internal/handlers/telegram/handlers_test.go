package handlers_telegram

import (
	"context"
	"errors"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/javiyt/tweettgram/internal/config"
	"github.com/javiyt/tweettgram/internal/pubsub"
	mb "github.com/javiyt/tweettgram/mocks/bot"
	mq "github.com/javiyt/tweettgram/mocks/pubsub"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	tb "gopkg.in/tucnak/telebot.v2"
	"testing"
	"time"
)

func TestTelegram_ExecuteHandlers(t *testing.T) {
	cfg := config.EnvConfig{
		BroadcastChannel: 1234,
	}

	t.Run("it should fail getting channel for text and photo notifications", func(t *testing.T) {
		mockedBot := new(mb.TelegramBot)
		mockedQueue := new(mq.Queue)

		th := NewTelegram(cfg, mockedBot, mockedQueue)

		mockedQueue.On("Subscribe", context.Background(), pubsub.TextTopic.String()).
			Once().
			Return(nil, errors.New("error getting channel error"))
		mockedQueue.On("Subscribe", context.Background(), pubsub.PhotoTopic.String()).
			Once().
			Return(nil, errors.New("error getting channel error"))
		mockedQueue.On("Publish", pubsub.ErrorTopic.String(), mock.MatchedBy(func(m *message.Message) bool {
			return string(m.Payload) == "{\"error\":\"error getting channel error\"}"
		})).Times(2).
			Return(nil)

		th.ExecuteHandlers()

		mockedQueue.AssertExpectations(t)
	})

	t.Run("it should fail unmarshaling text event", func(t *testing.T) {
		mockedBot := new(mb.TelegramBot)
		mockedQueue := new(mq.Queue)

		th := NewTelegram(cfg, mockedBot, mockedQueue)

		textChannel := make(chan *message.Message)
		photoChannel := make(chan *message.Message)
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
		mockedQueue.On("Publish", pubsub.ErrorTopic.String(), mock.MatchedBy(func(m *message.Message) bool {
			return string(m.Payload) == "{\"error\":\"parse error: unterminated string literal near offset 12 of '{\\\"asd\\\":\\\"qwer'\"}"
		})).Once().
			Return(nil)

		th.ExecuteHandlers()
		newMessage := message.NewMessage(watermill.NewUUID(), []byte("{\"asd\":\"qwer"))
		textChannel <- newMessage

		require.Eventually(t, func() bool {
			<-newMessage.Nacked()
			return true
		}, time.Second, time.Millisecond)
		mockedQueue.AssertExpectations(t)
	})

	t.Run("it should fail sending text message to telegram", func(t *testing.T) {
		mockedBot := new(mb.TelegramBot)
		mockedQueue := new(mq.Queue)

		th := NewTelegram(cfg, mockedBot, mockedQueue)

		textChannel := make(chan *message.Message)
		photoChannel := make(chan *message.Message)
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
		mockedQueue.On("Publish", pubsub.ErrorTopic.String(), mock.MatchedBy(func(m *message.Message) bool {
			return string(m.Payload) == "{\"error\":\"couldn't send message to telegram\"}"
		})).Once().
			Return(nil)
		mockedBot.On("Send", tb.ChatID(cfg.BroadcastChannel), mock.MatchedBy(func(m interface{}) bool {
			return m.(*tb.Message).Text == "testing message"
		})).Once().Return(nil, errors.New("couldn't send message to telegram"))

		th.ExecuteHandlers()
		newMessage := message.NewMessage(watermill.NewUUID(), []byte("{\"text\":\"testing message\"}"))
		textChannel <- newMessage

		require.Eventually(t, func() bool {
			<-newMessage.Nacked()
			return true
		}, time.Second, time.Millisecond)
		mockedQueue.AssertExpectations(t)
		mockedBot.AssertExpectations(t)
	})

	t.Run("it should send text message to telegram", func(t *testing.T) {
		mockedBot := new(mb.TelegramBot)
		mockedQueue := new(mq.Queue)

		th := NewTelegram(cfg, mockedBot, mockedQueue)

		textChannel := make(chan *message.Message)
		photoChannel := make(chan *message.Message)
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
		mockedBot.On("Send", tb.ChatID(cfg.BroadcastChannel), mock.MatchedBy(func(m interface{}) bool {
			return m.(*tb.Message).Text == "testing message"
		})).Once().Return(nil, nil)

		th.ExecuteHandlers()
		newMessage := message.NewMessage(watermill.NewUUID(), []byte("{\"text\":\"testing message\"}"))
		textChannel <- newMessage

		require.Eventually(t, func() bool {
			<-newMessage.Acked()
			return true
		}, time.Second, time.Millisecond)
		mockedQueue.AssertExpectations(t)
		mockedBot.AssertExpectations(t)
	})

	t.Run("it should fail unmarshaling photo event", func(t *testing.T) {
		mockedBot := new(mb.TelegramBot)
		mockedQueue := new(mq.Queue)

		th := NewTelegram(cfg, mockedBot, mockedQueue)

		textChannel := make(chan *message.Message)
		photoChannel := make(chan *message.Message)
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
		mockedQueue.On("Publish", pubsub.ErrorTopic.String(), mock.MatchedBy(func(m *message.Message) bool {
			return string(m.Payload) == "{\"error\":\"parse error: unterminated string literal near offset 12 of '{\\\"asd\\\":\\\"qwer'\"}"
		})).Once().
			Return(nil)

		th.ExecuteHandlers()
		newMessage := message.NewMessage(watermill.NewUUID(), []byte("{\"asd\":\"qwer"))
		photoChannel <- newMessage

		require.Eventually(t, func() bool {
			<-newMessage.Nacked()
			return true
		}, time.Second, time.Millisecond)
		mockedQueue.AssertExpectations(t)
	})

	t.Run("it should fail sending photo message to telegram", func(t *testing.T) {
		mockedBot := new(mb.TelegramBot)
		mockedQueue := new(mq.Queue)

		th := NewTelegram(cfg, mockedBot, mockedQueue)

		textChannel := make(chan *message.Message)
		photoChannel := make(chan *message.Message)
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
		mockedQueue.On("Publish", pubsub.ErrorTopic.String(), mock.MatchedBy(func(m *message.Message) bool {
			return string(m.Payload) == "{\"error\":\"couldn't send message to telegram\"}"
		})).Once().
			Return(nil)
		mockedBot.On("Send", tb.ChatID(cfg.BroadcastChannel), mock.MatchedBy(func(m interface{}) bool {
			var p *tb.Photo
			var ok bool
			if p, ok = m.(*tb.Photo); !ok {
				return false
			}

			return p.Caption == "testing message" &&
				p.FileID == "blablabla" &&
				p.FileURL == "http://photo.url" &&
				p.FileSize == 1234
		})).Once().Return(nil, errors.New("couldn't send message to telegram"))

		th.ExecuteHandlers()
		newMessage := message.NewMessage(watermill.NewUUID(), []byte("{\"caption\":\"testing message\",\"file_id\":\"blablabla\",\"file_url\":\"http://photo.url\",\"file_size\":1234}"))
		photoChannel <- newMessage

		require.Eventually(t, func() bool {
			<-newMessage.Nacked()
			return true
		}, time.Second, time.Millisecond)
		mockedQueue.AssertExpectations(t)
		mockedBot.AssertExpectations(t)
	})

	t.Run("it should send photo message to telegram", func(t *testing.T) {
		mockedBot := new(mb.TelegramBot)
		mockedQueue := new(mq.Queue)

		th := NewTelegram(cfg, mockedBot, mockedQueue)

		textChannel := make(chan *message.Message)
		photoChannel := make(chan *message.Message)
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
		mockedBot.On("Send", tb.ChatID(cfg.BroadcastChannel), mock.MatchedBy(func(m interface{}) bool {
			var p *tb.Photo
			var ok bool
			if p, ok = m.(*tb.Photo); !ok {
				return false
			}

			return p.Caption == "testing message" &&
				p.FileID == "blablabla" &&
				p.FileURL == "http://photo.url" &&
				p.FileSize == 1234
		})).Once().Return(nil, nil)

		th.ExecuteHandlers()
		newMessage := message.NewMessage(watermill.NewUUID(), []byte("{\"caption\":\"testing message\",\"file_id\":\"blablabla\",\"file_url\":\"http://photo.url\",\"file_size\":1234}"))
		photoChannel <- newMessage

		require.Eventually(t, func() bool {
			<-newMessage.Acked()
			return true
		}, time.Second, time.Millisecond)
		mockedQueue.AssertExpectations(t)
		mockedBot.AssertExpectations(t)
	})
}
