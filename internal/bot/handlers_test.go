package bot_test

import (
	"os"
	"strconv"
	"testing"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/javiyt/tweetgram/internal/bot"
	"github.com/javiyt/tweetgram/internal/config"
	"github.com/javiyt/tweetgram/internal/pubsub"
	mb "github.com/javiyt/tweetgram/mocks/bot"
	mq "github.com/javiyt/tweetgram/mocks/pubsub"
	"github.com/stretchr/testify/mock"

	tb "gopkg.in/tucnak/telebot.v2"
)

const (
	adminID          = 12345
	broadcastChannel = int64(987654)
	imagePayload     = "{\"caption\":\"testing\"," +
		"\"fileId\":\"blablabla\"," +
		"\"fileUrl\":\"http://myimage.com/test.jpg\"," +
		"\"fileSize\":1234," +
		"\"fileContent\":\"iVBORw0KGgoAAAANSUhEUgAAAAQAAAAECAIAAAAmkwkpAAAAQklEQVR4nGJWTd9ZaWdyOfW69Y8z" +
		"DF5sfALun5c7SL+8ysQUqp7euSxThUtU5v9FJg2PoueTrrw5Vyt36AYgAAD//yOnFnjB+cHEAAAAAElFTkSuQmCC\"}"
)

type downloadImageError struct{}

func (m downloadImageError) Error() string {
	return "error downloading image"
}

func TestHandlerStartAndHelpCommand(t *testing.T) {
	commands := []struct {
		command  string
		expected string
	}{
		{
			command:  "/start",
			expected: "Thanks for using the bot! You can type /help command to know what can I do",
		},
		{
			command:  "/help",
			expected: "/help - Show help\n/start - Start a conversation with the bot\n",
		},
	}
	for i := range commands {
		i := i
		handler, mockedBot, _ := generateHandlerAndMockedBot(t, commands[i].command, config.EnvConfig{})

		t.Run("it should do nothing when not in private conversation", func(t *testing.T) {
			handler(&bot.TelegramMessage{
				IsPrivate: false,
			})

			mockedBot.AssertExpectations(t)
			mockedBot.AssertNotCalled(t, "Send", mock.Anything, mock.Anything)
		})

		t.Run("it should message when in private conversation", func(t *testing.T) {
			m := &bot.TelegramMessage{
				IsPrivate: true,
				SenderID:  "1234",
			}
			mockedBot.On(
				"Send",
				m.SenderID,
				commands[i].expected,
			).Once().Return(nil, nil)

			handler(m)

			mockedBot.AssertExpectations(t)
		})
	}
}

func TestHandlersFilters(t *testing.T) {
	commands := []string{tb.OnPhoto, tb.OnText}
	for i := range commands {
		i := i
		adminID := 12345
		broadcastChannel := int64(987654)

		mockedQueue := new(mq.Queue)
		handler, mockedBot, _ := generateHandlerAndMockedBot(t, commands[i], config.EnvConfig{
			Admins:           []int{adminID},
			BroadcastChannel: broadcastChannel,
		})

		testCases := []struct {
			name string
			m    *bot.TelegramMessage
		}{
			{
				name: "it should do nothing when not in private conversation",
				m: &bot.TelegramMessage{
					IsPrivate: false,
					SenderID:  "1234",
				},
			},
			{
				name: "it should do nothing when in private conversation but not admin",
				m: &bot.TelegramMessage{
					IsPrivate: true,
					SenderID:  "54321",
				},
			},
			{
				name: "it should fail when in private conversation but sender can't be converted to int",
				m: &bot.TelegramMessage{
					IsPrivate: true,
					SenderID:  "asdfg",
				},
			},
		}

		for i := range testCases {
			i := i
			t.Run(testCases[i].name, func(t *testing.T) {
				handler(testCases[i].m)

				mockedBot.AssertExpectations(t)
				mockedQueue.AssertNotCalled(t, "Publish", mock.Anything, mock.Anything)
			})
		}
	}
}

func TestHandlerPhoto(t *testing.T) {
	handler, mockedBot, mockedQueue := generateHandlerAndMockedBot(t, tb.OnPhoto, config.EnvConfig{
		Admins:           []int{adminID},
		BroadcastChannel: broadcastChannel,
	})

	successPhoto := &bot.TelegramMessage{
		IsPrivate: true,
		SenderID:  strconv.Itoa(adminID),
		Photo: bot.TelegramPhoto{
			Caption:  "testing",
			FileID:   "blablabla",
			FileURL:  "http://myimage.com/test.jpg",
			FileSize: 1234,
		},
	}

	t.Run("it should do nothing when caption no present", func(t *testing.T) {
		handler(&bot.TelegramMessage{
			IsPrivate: true,
			SenderID:  strconv.Itoa(adminID),
			Photo:     bot.TelegramPhoto{Caption: ""},
		})

		mockedBot.AssertExpectations(t)
		mockedQueue.AssertNotCalled(t, "Publish", mock.Anything, mock.Anything)
	})

	t.Run("it should do nothing when error getting image", func(t *testing.T) {
		mockedBot.On("GetFile", successPhoto.Photo.FileID).Once().
			Return(nil, downloadImageError{})

		handler(successPhoto)

		mockedBot.AssertExpectations(t)
		mockedQueue.AssertNotCalled(t, "Publish", mock.Anything, mock.Anything)
	})

	t.Run("it should send photo when caption is present and image could be downloaded", func(t *testing.T) {
		file, _ := os.Open("testdata/test.png")
		defer func() { _ = file.Close() }()

		mockedBot.On("GetFile", successPhoto.Photo.FileID).Once().Return(file, nil)
		mockedQueue.On(
			"Publish",
			pubsub.PhotoTopic.String(),
			mock.MatchedBy(func(message *message.Message) bool {
				return string(message.Payload) == imagePayload
			}),
		).Once().Return(nil)

		handler(successPhoto)

		mockedBot.AssertExpectations(t)
		mockedQueue.AssertExpectations(t)
	})
}

func TestHandlerText(t *testing.T) {
	handler, mockedBot, mockedQueue := generateHandlerAndMockedBot(t, tb.OnText, config.EnvConfig{
		Admins:           []int{adminID},
		BroadcastChannel: broadcastChannel,
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

func generateHandlerAndMockedBot(
	t *testing.T,
	toHandle string,
	cfg config.EnvConfig,
) (bot.TelegramHandler, *mb.TelegramBot, *mq.Queue) {
	allHandlers := []string{"/start", "/help", tb.OnPhoto, tb.OnText}

	var (
		handler bot.TelegramHandler
		ok      bool
	)

	mockedQueue := new(mq.Queue)

	mockedBot := new(mb.TelegramBot)
	mockedBot.On("SetCommands", mock.Anything).Once().Return(nil)

	for _, v := range allHandlers {
		if v == toHandle {
			mockedBot.On("Handle", toHandle, mock.Anything).
				Once().
				Return(nil, nil).
				Run(func(args mock.Arguments) {
					handler, ok = args.Get(1).(bot.TelegramHandler)
					if !ok {
						t.Fatal("given handler is not valid")
					}
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

	return handler, mockedBot, mockedQueue
}
