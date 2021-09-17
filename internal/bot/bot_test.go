package bot_test

import (
	"testing"

	"github.com/javiyt/tweettgram/internal/bot"
	mb "github.com/javiyt/tweettgram/mocks/bot"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	tb "gopkg.in/tucnak/telebot.v2"
)

func TestStart(t *testing.T) {
	mockedBot := new(mb.TelegramBot)

	cmds := []tb.Command{
		{
			Text:        "help",
			Description: "Show help",
		},
		{
			Text:        "start",
			Description: "Start a conversation with the bot",
		},
	}
	mockedBot.On("SetCommands", cmds).Once().Return(nil)
	mockedBot.On("Handle", "/start", mock.Anything).Once().Return(nil, nil)
	mockedBot.On("Handle", "/help", mock.Anything).Once().Return(nil, nil)
	mockedBot.On("Start").Once()

	require.Nil(t, bot.NewBot(bot.WithTelegramBot(mockedBot)).Start())

	mockedBot.AssertExpectations(t)
}

func TestStop(t *testing.T) {
	mockedBot := new(mb.TelegramBot)
	mockedBot.On("Stop").Once()

	bot.NewBot(bot.WithTelegramBot(mockedBot)).Stop()

	mockedBot.AssertExpectations(t)
}

func TestHandlerStartCommand(t *testing.T) {
	var handler func(*tb.Message)
	mockedBot := new(mb.TelegramBot)
	mockedBot.On("SetCommands", mock.Anything).Once().Return(nil)
	mockedBot.On("Handle", "/start", mock.Anything).
		Once().
		Return(nil, nil).
		Run(func(args mock.Arguments) {
			handler = args.Get(1).(func(*tb.Message))
		})
	mockedBot.On("Handle", "/help", mock.Anything).Once().Return(nil, nil)
	mockedBot.On("Start").Once()

	bot.NewBot(bot.WithTelegramBot(mockedBot)).Start()

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
	var handler func(*tb.Message)
	mockedBot := new(mb.TelegramBot)
	mockedBot.On("SetCommands", mock.Anything).Once().Return(nil)
	mockedBot.On("Handle", "/help", mock.Anything).
		Once().
		Return(nil, nil).
		Run(func(args mock.Arguments) {
			handler = args.Get(1).(func(*tb.Message))
		})
	mockedBot.On("Handle", "/start", mock.Anything).Once().Return(nil, nil)
	mockedBot.On("Start").Once()

	bot.NewBot(bot.WithTelegramBot(mockedBot)).Start()

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
