package bot_test

import (
	"errors"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/javiyt/tweetgram/internal/pubsub"
	mq "github.com/javiyt/tweetgram/mocks/pubsub"
	"os"
	"strconv"
	"testing"

	"github.com/javiyt/tweetgram/internal/bot"
	"github.com/javiyt/tweetgram/internal/config"
	mb "github.com/javiyt/tweetgram/mocks/bot"
	"github.com/stretchr/testify/mock"

	tb "gopkg.in/tucnak/telebot.v2"
)

func TestHandlerStartCommand(t *testing.T) {
	handler, mockedBot := generateHandlerAndMockedBot(
		"/start",
		config.EnvConfig{},
		new(mq.Queue),
	)

	t.Run("it should do nothing when not in private conversation", func(t *testing.T) {
		handler(&bot.TelegramMessage{
			IsPrivate: false,
		})

		mockedBot.AssertExpectations(t)
		mockedBot.AssertNotCalled(t, "Send", mock.Anything, mock.Anything)
	})

	t.Run("it should send welcome message when in private conversation", func(t *testing.T) {
		m := &bot.TelegramMessage{
			IsPrivate: true,
			SenderID:  "1234",
		}
		mockedBot.On(
			"Send",
			m.SenderID,
			"Thanks for using the bot! You can type /help command to know what can I do",
		).Once().Return(nil, nil)

		handler(m)

		mockedBot.AssertExpectations(t)
	})
}

func TestHandlerHelpCommand(t *testing.T) {
	handler, mockedBot := generateHandlerAndMockedBot(
		"/help",
		config.EnvConfig{},
		new(mq.Queue),
	)

	t.Run("it should do nothing when not in private conversation", func(t *testing.T) {
		handler(&bot.TelegramMessage{
			IsPrivate: false,
		})

		mockedBot.AssertExpectations(t)
		mockedBot.AssertNotCalled(t, "Send", mock.Anything, mock.Anything)
	})

	t.Run("it should send welcome message when in private conversation", func(t *testing.T) {
		m := &bot.TelegramMessage{
			IsPrivate: true,
			SenderID:  "1234",
		}
		mockedBot.On(
			"Send",
			m.SenderID,
			"/help - Show help\n/start - Start a conversation with the bot\n",
		).Once().Return(nil, nil)

		handler(m)

		mockedBot.AssertExpectations(t)
	})
}

func TestHandlerPhoto(t *testing.T) {
	adminID := 12345
	broadcastChannel := int64(987654)

	mockedQueue := new(mq.Queue)
	handler, mockedBot := generateHandlerAndMockedBot(
		tb.OnPhoto,
		config.EnvConfig{
			Admins:           []int{adminID},
			BroadcastChannel: broadcastChannel,
		},
		mockedQueue,
	)

	t.Run("it should do nothing when not in private conversation", func(t *testing.T) {
		handler(&bot.TelegramMessage{
			IsPrivate: false,
			SenderID:  "1234",
		})

		mockedBot.AssertExpectations(t)
		mockedQueue.AssertNotCalled(t, "Publish", mock.Anything, mock.Anything)
	})

	t.Run("it should do nothing when in private conversation but not admin", func(t *testing.T) {
		handler(&bot.TelegramMessage{
			IsPrivate: true,
			SenderID:  "1234",
		})

		mockedBot.AssertExpectations(t)
		mockedQueue.AssertNotCalled(t, "Publish", mock.Anything, mock.Anything)
	})

	t.Run("it should do nothing when caption no present", func(t *testing.T) {
		handler(&bot.TelegramMessage{
			IsPrivate: true,
			SenderID:  "1234",
			Photo:     bot.TelegramPhoto{Caption: ""},
		})

		mockedBot.AssertExpectations(t)
		mockedQueue.AssertNotCalled(t, "Publish", mock.Anything, mock.Anything)
	})

	t.Run("it should do nothing when error getting image", func(t *testing.T) {
		m := &bot.TelegramMessage{
			IsPrivate: true,
			SenderID:  strconv.Itoa(adminID),
			Photo: bot.TelegramPhoto{
				Caption:  "testing",
				FileID:   "blablabla",
				FileURL:  "http://myimage.com/test.jpg",
				FileSize: 1234,
			},
		}
		mockedBot.On("GetFile", m.Photo.FileID).
			Once().
			Return(nil, errors.New("error downloading image"))

		handler(m)

		mockedBot.AssertExpectations(t)
		mockedQueue.AssertNotCalled(t, "Publish", mock.Anything, mock.Anything)
	})

	t.Run("it should send photo when caption is present and image could be downloaded", func(t *testing.T) {
		m := &bot.TelegramMessage{
			IsPrivate: true,
			SenderID:  strconv.Itoa(adminID),
			Photo: bot.TelegramPhoto{
				Caption:  "testing",
				FileID:   "blablabla",
				FileURL:  "http://myimage.com/test.jpg",
				FileSize: 1234,
			},
		}
		file, _ := os.Open("testdata/test.png")
		defer func() { _ = file.Close() }()
		mockedBot.On("GetFile", m.Photo.FileID).
			Once().
			Return(file, nil)
		mockedQueue.On(
			"Publish",
			pubsub.PhotoTopic.String(),
			mock.MatchedBy(func(message *message.Message) bool {
				return string(message.Payload) == "{\"caption\":\"testing\","+
					"\"file_id\":\"blablabla\","+
					"\"file_url\":\"http://myimage.com/test.jpg\","+
					"\"file_size\":1234,"+
					"\"file_content\":\"iVBORw0KGgoAAAANSUhEUgAAAAQAAAAECAIAAAAmkwkpAAAAQklEQVR4nGJWTd9ZaWdyOfW69Y8zDF5sfALun5c7SL+8ysQUqp7euSxThUtU5v9FJg2PoueTrrw5Vyt36AYgAAD//yOnFnjB+cHEAAAAAElFTkSuQmCC\"}"
			}),
		).Once().Return(nil)

		handler(m)

		mockedBot.AssertExpectations(t)
		mockedQueue.AssertExpectations(t)
	})
}

func TestHandlerText(t *testing.T) {
	adminID := 12345
	broadcastChannel := int64(987654)

	mockedQueue := new(mq.Queue)
	handler, mockedBot := generateHandlerAndMockedBot(
		tb.OnText,
		config.EnvConfig{
			Admins:           []int{adminID},
			BroadcastChannel: broadcastChannel,
		},
		mockedQueue,
	)

	t.Run("it should do nothing when not in private conversation", func(t *testing.T) {
		handler(&bot.TelegramMessage{
			IsPrivate: false,
			SenderID:  "1234",
		})

		mockedBot.AssertExpectations(t)
		mockedQueue.AssertNotCalled(t, "Publish", mock.Anything, mock.Anything)
	})

	t.Run("it should do nothing when in private conversation but not admin", func(t *testing.T) {
		handler(&bot.TelegramMessage{
			IsPrivate: true,
			SenderID:  "54321",
		})

		mockedBot.AssertExpectations(t)
		mockedQueue.AssertNotCalled(t, "Publish", mock.Anything, mock.Anything)
	})

	t.Run("it should do nothing when text no present", func(t *testing.T) {
		handler(&bot.TelegramMessage{
			IsPrivate: true,
			SenderID:  strconv.Itoa(adminID),
			Text:      "",
		})

		mockedBot.AssertExpectations(t)
		mockedQueue.AssertNotCalled(t, "Send", mock.Anything, mock.Anything)
	})

	t.Run("it should send text when present", func(t *testing.T) {
		m := &bot.TelegramMessage{
			IsPrivate: true,
			SenderID:  strconv.Itoa(adminID),
			Text:      "testing",
		}
		mockedQueue.On(
			"Publish",
			pubsub.TextTopic.String(),
			mock.MatchedBy(func(message *message.Message) bool {
				return string(message.Payload) == "{\"text\":\"testing\"}"
			}),
		).Once().Return(nil)

		handler(m)

		mockedBot.AssertExpectations(t)
		mockedQueue.AssertExpectations(t)
	})
}

func generateHandlerAndMockedBot(toHandle string, cfg config.EnvConfig, mockedQueue *mq.Queue) (func(*bot.TelegramMessage), *mb.TelegramBot) {
	allHandlers := []string{"/start", "/help", tb.OnPhoto, tb.OnText}
	var handler bot.TelegramHandler

	mockedBot := new(mb.TelegramBot)
	mockedBot.On("SetCommands", mock.Anything).Once().Return(nil)

	for _, v := range allHandlers {
		if v == toHandle {
			mockedBot.On("Handle", toHandle, mock.Anything).
				Once().
				Return(nil, nil).
				Run(func(args mock.Arguments) {
					handler = args.Get(1).(bot.TelegramHandler)
				})
		} else {
			mockedBot.On("Handle", v, mock.Anything).Once().Return(nil, nil)
		}
	}

	_ = bot.NewBot(
		bot.WithTelegramBot(mockedBot),
		bot.WithConfig(cfg),
		bot.WithQueue(mockedQueue),
	).Start()

	return handler, mockedBot
}
