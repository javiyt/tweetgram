package bot_test

import (
	"os"
	"testing"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/javiyt/tweetgram/internal/pubsub"
	mq "github.com/javiyt/tweetgram/mocks/pubsub"

	"github.com/javiyt/tweetgram/internal/bot"
	"github.com/javiyt/tweetgram/internal/config"
	mb "github.com/javiyt/tweetgram/mocks/bot"
	"github.com/stretchr/testify/mock"

	tb "gopkg.in/tucnak/telebot.v2"
)

const (
	adminID          = 12345
	broadcastChannel = int64(987654)
	imagePayload     = "{\"caption\":\"testing\"," +
		"\"file_id\":\"blablabla\"," +
		"\"file_url\":\"http://myimage.com/test.jpg\"," +
		"\"file_size\":1234," +
		"\"file_content\":\"iVBORw0KGgoAAAANSUhEUgAAAAQAAAAECAIAAAAmkwkpAAAAQklEQVR4nGJWTd9ZaWdyOfW69Y8z" +
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
		handler, mockedBot, _ := generateHandlerAndMockedBot(commands[i].command, config.EnvConfig{})

		t.Run("it should do nothing when not in private conversation", func(t *testing.T) {
			handler(&tb.Message{
				Chat: &tb.Chat{
					Type: tb.ChatGroup,
				},
			})

			mockedBot.AssertExpectations(t)
			mockedBot.AssertNotCalled(t, "Send", mock.Anything, mock.Anything)
		})

		t.Run("it should send message when in private conversation", func(t *testing.T) {
			m := &tb.Message{
				Chat: &tb.Chat{
					Type: tb.ChatPrivate,
				},
				Sender: &tb.User{
					ID: 1234,
				},
			}
			mockedBot.On("Send", m.Sender, commands[i].expected).Once().Return(nil, nil)

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
		handler, mockedBot, _ := generateHandlerAndMockedBot(commands[i], config.EnvConfig{
			Admins:           []int{adminID},
			BroadcastChannel: broadcastChannel,
		})

		testCases := []struct {
			name string
			m    *tb.Message
		}{
			{
				name: "it should do nothing when not in private conversation",
				m: &tb.Message{
					Chat:   &tb.Chat{Type: tb.ChatGroup},
					Sender: &tb.User{ID: 1234},
				},
			},
			{
				name: "it should do nothing when in private conversation but not admin",
				m: &tb.Message{
					Chat:   &tb.Chat{Type: tb.ChatPrivate},
					Sender: &tb.User{ID: 54321},
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
	handler, mockedBot, mockedQueue := generateHandlerAndMockedBot(tb.OnPhoto, config.EnvConfig{
		Admins:           []int{adminID},
		BroadcastChannel: broadcastChannel,
	})

	successPhoto := &tb.Message{
		Chat:    &tb.Chat{Type: tb.ChatPrivate},
		Sender:  &tb.User{ID: adminID},
		Caption: "testing",
		Photo: &tb.Photo{
			Caption: "testing",
			File: tb.File{
				FileID:   "blablabla",
				FileURL:  "http://myimage.com/test.jpg",
				FileSize: 1234,
			},
		},
	}

	t.Run("it should do nothing when caption no present", func(t *testing.T) {
		handler(&tb.Message{
			Chat:    &tb.Chat{Type: tb.ChatPrivate},
			Sender:  &tb.User{ID: adminID},
			Caption: "",
		})

		mockedBot.AssertExpectations(t)
		mockedQueue.AssertNotCalled(t, "Publish", mock.Anything, mock.Anything)
	})

	t.Run("it should do nothing when error getting image", func(t *testing.T) {
		mockedBot.On("GetFile", &successPhoto.Photo.File).Once().
			Return(nil, downloadImageError{})

		handler(successPhoto)

		mockedBot.AssertExpectations(t)
		mockedQueue.AssertNotCalled(t, "Publish", mock.Anything, mock.Anything)
	})

	t.Run("it should send photo when caption is present and image could be downloaded", func(t *testing.T) {
		file, _ := os.Open("testdata/test.png")
		defer func() { _ = file.Close() }()

		mockedBot.On("GetFile", &successPhoto.Photo.File).Once().Return(file, nil)
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
	handler, mockedBot, mockedQueue := generateHandlerAndMockedBot(tb.OnText, config.EnvConfig{
		Admins:           []int{adminID},
		BroadcastChannel: broadcastChannel,
	})

	t.Run("it should do nothing when text no present", func(t *testing.T) {
		handler(&tb.Message{
			Chat:   &tb.Chat{Type: tb.ChatPrivate},
			Sender: &tb.User{ID: adminID},
			Text:   "",
		})

		mockedBot.AssertExpectations(t)
		mockedQueue.AssertNotCalled(t, "Send", mock.Anything, mock.Anything)
	})

	t.Run("it should send text when present", func(t *testing.T) {
		m := &tb.Message{
			Chat:   &tb.Chat{Type: tb.ChatPrivate},
			Sender: &tb.User{ID: adminID},
			Text:   "testing",
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

func generateHandlerAndMockedBot(toHandle string, cfg config.EnvConfig) (
	func(*tb.Message),
	*mb.TelegramBot,
	*mq.Queue,
) {
	allHandlers := []string{"/start", "/help", tb.OnPhoto, tb.OnText}

	var handler func(*tb.Message)

	mockedQueue := new(mq.Queue)

	mockedBot := new(mb.TelegramBot)
	mockedBot.On("SetCommands", mock.Anything).Once().Return(nil)

	for _, v := range allHandlers {
		if v == toHandle {
			mockedBot.On("Handle", toHandle, mock.Anything).
				Once().
				Return(nil, nil).
				Run(func(args mock.Arguments) {
					handler = args.Get(1).(func(*tb.Message))
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
