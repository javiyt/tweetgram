package bot_test

import (
	"testing"

	"github.com/javiyt/tweettgram/internal/bot"
	"github.com/javiyt/tweettgram/internal/config"
	mb "github.com/javiyt/tweettgram/mocks/bot"
	"github.com/stretchr/testify/mock"

	tb "gopkg.in/tucnak/telebot.v2"
)

func TestHandlerStartCommand(t *testing.T) {
	handler, mockedBot := generateHandlerAndMockedBot("/start", config.EnvConfig{})

	t.Run("it should do nothing when not in private conversation", func(t *testing.T) {
		handler(&tb.Message{
			Chat: &tb.Chat{
				Type: tb.ChatGroup,
			},
		})

		mockedBot.AssertExpectations(t)
		mockedBot.AssertNotCalled(t, "Send", mock.Anything, mock.Anything)
	})

	t.Run("it should send welcome message when in private conversation", func(t *testing.T) {
		m := &tb.Message{
			Chat: &tb.Chat{
				Type: tb.ChatPrivate,
			},
			Sender: &tb.User{
				ID: 1234,
			},
		}
		mockedBot.On(
			"Send",
			m.Sender,
			"Thanks for using the bot! You can type /help command to know what can I do",
		).Once().Return(nil, nil)

		handler(m)

		mockedBot.AssertExpectations(t)
	})
}

func TestHandlerHelpCommand(t *testing.T) {
	handler, mockedBot := generateHandlerAndMockedBot("/help", config.EnvConfig{})

	t.Run("it should do nothing when not in private conversation", func(t *testing.T) {
		handler(&tb.Message{
			Chat: &tb.Chat{
				Type: tb.ChatGroup,
			},
		})

		mockedBot.AssertExpectations(t)
		mockedBot.AssertNotCalled(t, "Send", mock.Anything, mock.Anything)
	})

	t.Run("it should send welcome message when in private conversation", func(t *testing.T) {
		m := &tb.Message{
			Chat: &tb.Chat{
				Type: tb.ChatPrivate,
			},
			Sender: &tb.User{
				ID: 1234,
			},
		}
		mockedBot.On(
			"Send",
			m.Sender,
			"/help - Show help\n/start - Start a conversation with the bot\n",
		).Once().Return(nil, nil)

		handler(m)

		mockedBot.AssertExpectations(t)
	})
}

func TestHandlerPhoto(t *testing.T) {
	adminID := 12345
	broadcastChannel := int64(987654)

	handler, mockedBot := generateHandlerAndMockedBot(
		tb.OnPhoto,
		config.EnvConfig{
			Admins:           []int{adminID},
			BroadcastChannel: broadcastChannel,
		},
	)

	t.Run("it should do nothing when not in private conversation", func(t *testing.T) {
		handler(&tb.Message{
			Chat: &tb.Chat{
				Type: tb.ChatGroup,
			},
			Sender: &tb.User{
				ID: 1234,
			},
		})

		mockedBot.AssertExpectations(t)
		mockedBot.AssertNotCalled(t, "Send", mock.Anything, mock.Anything)
	})

	t.Run("it should do nothing when in private conversation but not admin", func(t *testing.T) {
		handler(&tb.Message{
			Chat: &tb.Chat{
				Type: tb.ChatPrivate,
			},
			Sender: &tb.User{
				ID: 54321,
			},
		})

		mockedBot.AssertExpectations(t)
		mockedBot.AssertNotCalled(t, "Send", mock.Anything, mock.Anything)
	})

	t.Run("it should do nothing when caption no present", func(t *testing.T) {
		handler(&tb.Message{
			Chat: &tb.Chat{
				Type: tb.ChatPrivate,
			},
			Sender: &tb.User{
				ID: adminID,
			},
			Caption: "",
		})

		mockedBot.AssertExpectations(t)
		mockedBot.AssertNotCalled(t, "Send", mock.Anything, mock.Anything)
	})

	t.Run("it should send caption when present", func(t *testing.T) {
		m := &tb.Message{
			Chat: &tb.Chat{
				Type: tb.ChatPrivate,
			},
			Sender: &tb.User{
				ID: adminID,
			},
			Caption: "testing",
		}
		mockedBot.On(
			"Send",
			tb.ChatID(broadcastChannel),
			"testing",
		).Once().Return(nil, nil)

		handler(m)

		mockedBot.AssertExpectations(t)
	})
}

func TestHandlerText(t *testing.T) {
	adminID := 12345
	broadcastChannel := int64(987654)

	handler, mockedBot := generateHandlerAndMockedBot(
		tb.OnText,
		config.EnvConfig{
			Admins:           []int{adminID},
			BroadcastChannel: broadcastChannel,
		},
	)

	t.Run("it should do nothing when not in private conversation", func(t *testing.T) {
		handler(&tb.Message{
			Chat: &tb.Chat{
				Type: tb.ChatGroup,
			},
			Sender: &tb.User{
				ID: 1234,
			},
		})

		mockedBot.AssertExpectations(t)
		mockedBot.AssertNotCalled(t, "Send", mock.Anything, mock.Anything)
	})

	t.Run("it should do nothing when in private conversation but not admin", func(t *testing.T) {
		handler(&tb.Message{
			Chat: &tb.Chat{
				Type: tb.ChatPrivate,
			},
			Sender: &tb.User{
				ID: 54321,
			},
		})

		mockedBot.AssertExpectations(t)
		mockedBot.AssertNotCalled(t, "Send", mock.Anything, mock.Anything)
	})

	t.Run("it should do nothing when text no present", func(t *testing.T) {
		handler(&tb.Message{
			Chat: &tb.Chat{
				Type: tb.ChatPrivate,
			},
			Sender: &tb.User{
				ID: adminID,
			},
			Text: "",
		})

		mockedBot.AssertExpectations(t)
		mockedBot.AssertNotCalled(t, "Send", mock.Anything, mock.Anything)
	})

	t.Run("it should send text when present", func(t *testing.T) {
		m := &tb.Message{
			Chat: &tb.Chat{
				Type: tb.ChatPrivate,
			},
			Sender: &tb.User{
				ID: adminID,
			},
			Text: "testing",
		}
		mockedBot.On(
			"Send",
			tb.ChatID(broadcastChannel),
			"testing",
		).Once().Return(nil, nil)

		handler(m)

		mockedBot.AssertExpectations(t)
	})
}

func generateHandlerAndMockedBot(toHandle string, cfg config.EnvConfig) (func(*tb.Message), *mb.TelegramBot) {
	allHandlers := []string{"/start", "/help", tb.OnPhoto, tb.OnText}
	var handler func(*tb.Message)

	mockedBot := new(mb.TelegramBot)
	mockedBot.On("SetCommands", mock.Anything).Once().Return(nil)
	mockedBot.On("Start").Once()

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

	bot.NewBot(
		bot.WithTelegramBot(mockedBot),
		bot.WithConfig(cfg),
	).Start()

	return handler, mockedBot
}
