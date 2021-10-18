package handlers_twitter_test

import (
	"context"
	"errors"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	ht "github.com/javiyt/tweetgram/internal/handlers/twitter"
	"github.com/javiyt/tweetgram/internal/pubsub"
	mb "github.com/javiyt/tweetgram/mocks/bot"
	mq "github.com/javiyt/tweetgram/mocks/pubsub"
	"github.com/mailru/easyjson"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestTwitter_ExecuteHandlers(t *testing.T) {
	t.Run("it should fail getting channel for text and photo notifications", func(t *testing.T) {
		mockedTwitter := new(mb.TwitterClient)
		mockedQueue := new(mq.Queue)

		th := ht.NewTwitter(ht.WithTwitterClient(mockedTwitter), ht.WithQueue(mockedQueue))

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
		mockedTwitter := new(mb.TwitterClient)
		mockedQueue := new(mq.Queue)

		th := ht.NewTwitter(ht.WithTwitterClient(mockedTwitter), ht.WithQueue(mockedQueue))

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
			<-newMessage.Acked()
			return true
		}, time.Second, time.Millisecond)
		mockedQueue.AssertExpectations(t)
	})

	t.Run("it should fail sending text message to twitter", func(t *testing.T) {
		mockedTwitter := new(mb.TwitterClient)
		mockedQueue := new(mq.Queue)

		th := ht.NewTwitter(ht.WithTwitterClient(mockedTwitter), ht.WithQueue(mockedQueue))

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
		mockedQueue.On(
			"Publish",
			pubsub.ErrorTopic.String(),
			mock.MatchedBy(func(m *message.Message) bool {
				return string(m.Payload) == "{\"error\":\"couldn't send message to twitter\"}"
			}),
		).Once().
			Return(nil)
		mockedTwitter.On("SendUpdate", "testing message").Once().Return(errors.New("couldn't send message to twitter"))

		th.ExecuteHandlers()
		newMessage := message.NewMessage(watermill.NewUUID(), []byte("{\"text\":\"testing message\"}"))
		textChannel <- newMessage

		require.Eventually(t, func() bool {
			<-newMessage.Nacked()
			return true
		}, time.Second, time.Millisecond)
		mockedQueue.AssertExpectations(t)
		mockedTwitter.AssertExpectations(t)
	})

	t.Run("it should send text message to twitter", func(t *testing.T) {
		mockedTwitter := new(mb.TwitterClient)
		mockedQueue := new(mq.Queue)

		th := ht.NewTwitter(ht.WithTwitterClient(mockedTwitter), ht.WithQueue(mockedQueue))

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
		mockedTwitter.On("SendUpdate", "testing message").Once().Return(nil)

		th.ExecuteHandlers()
		newMessage := message.NewMessage(watermill.NewUUID(), []byte("{\"text\":\"testing message\"}"))
		textChannel <- newMessage

		require.Eventually(t, func() bool {
			<-newMessage.Acked()
			return true
		}, time.Second, time.Millisecond)
		mockedQueue.AssertExpectations(t)
		mockedTwitter.AssertExpectations(t)
	})

	t.Run("it should fail unmarshaling photo event", func(t *testing.T) {
		mockedTwitter := new(mb.TwitterClient)
		mockedQueue := new(mq.Queue)

		th := ht.NewTwitter(ht.WithTwitterClient(mockedTwitter), ht.WithQueue(mockedQueue))

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
			<-newMessage.Acked()
			return true
		}, time.Second, time.Millisecond)
		mockedQueue.AssertExpectations(t)
	})

	t.Run("it should fail sending photo to twitter", func(t *testing.T) {
		mockedTwitter := new(mb.TwitterClient)
		mockedQueue := new(mq.Queue)

		th := ht.NewTwitter(ht.WithTwitterClient(mockedTwitter), ht.WithQueue(mockedQueue))

		textChannel := make(chan *message.Message)
		photoChannel := make(chan *message.Message)
		photoContent := []byte("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAIAAACQd1PeAAAAEElEQVR4nGKaks0ECAAA//8CoAEEsZgdLgAAAABJRU5ErkJggg==")
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
		mockedQueue.On(
			"Publish",
			pubsub.ErrorTopic.String(),
			mock.MatchedBy(func(m *message.Message) bool {
				return string(m.Payload) == "{\"error\":\"couldn't send message to twitter\"}"
			}),
		).Once().
			Return(nil)
		mockedTwitter.On(
			"SendUpdateWithPhoto",
			"testing caption",
			photoContent,
		).
			Once().
			Return(errors.New("couldn't send message to twitter"))

		th.ExecuteHandlers()

		bytes, _ := easyjson.Marshal(pubsub.PhotoEvent{
			Caption:     "testing caption",
			FileID:      "123456789asdfg",
			FileURL:     "http://photo.url",
			FileSize:    12345,
			FileContent: photoContent,
		})
		newMessage := message.NewMessage(watermill.NewUUID(), bytes)
		photoChannel <- newMessage

		require.Eventually(t, func() bool {
			<-newMessage.Nacked()
			return true
		}, time.Second, time.Millisecond)
		mockedQueue.AssertExpectations(t)
		mockedTwitter.AssertExpectations(t)
	})

	t.Run("it should send photo to twitter", func(t *testing.T) {
		mockedTwitter := new(mb.TwitterClient)
		mockedQueue := new(mq.Queue)

		th := ht.NewTwitter(ht.WithTwitterClient(mockedTwitter), ht.WithQueue(mockedQueue))

		textChannel := make(chan *message.Message)
		photoChannel := make(chan *message.Message)
		photoContent := []byte("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAIAAACQd1PeAAAAEElEQVR4nGKaks0ECAAA//8CoAEEsZgdLgAAAABJRU5ErkJggg==")
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
		mockedTwitter.On(
			"SendUpdateWithPhoto",
			"testing caption",
			photoContent,
		).Once().Return(nil)

		th.ExecuteHandlers()

		bytes, _ := easyjson.Marshal(pubsub.PhotoEvent{
			Caption:     "testing caption",
			FileID:      "123456789asdfg",
			FileURL:     "http://photo.url",
			FileSize:    12345,
			FileContent: photoContent,
		})
		newMessage := message.NewMessage(watermill.NewUUID(), bytes)
		photoChannel <- newMessage

		require.Eventually(t, func() bool {
			<-newMessage.Acked()
			return true
		}, time.Second, time.Millisecond)
		mockedQueue.AssertExpectations(t)
		mockedTwitter.AssertExpectations(t)
	})
}
