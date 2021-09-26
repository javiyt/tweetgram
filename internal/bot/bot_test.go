package bot_test

import (
	"errors"
	"testing"

	"github.com/javiyt/tweettgram/internal/bot"
	mb "github.com/javiyt/tweettgram/mocks/bot"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	tb "gopkg.in/tucnak/telebot.v2"
)

func TestStart(t *testing.T) {
	mockedBot := new(mb.TelegramBot)
	mockedTwitter := new(mb.TwitterClient)

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

	b := bot.NewBot(bot.WithTelegramBot(mockedBot), bot.WithTwitterClient(mockedTwitter))

	t.Run("it should fail setting up the commands", func(t *testing.T) {
		mockedBot.On("SetCommands", cmds).Once().Return(errors.New("not setting commands"))

		require.EqualError(t, b.Start(), "not setting commands")

		mockedBot.AssertExpectations(t)
		mockedBot.AssertNotCalled(t, "Handle", "/start", mock.Anything)
		mockedBot.AssertNotCalled(t,"Handle", "/help", mock.Anything)
		mockedBot.AssertNotCalled(t,"Handle", tb.OnPhoto, mock.Anything)
		mockedBot.AssertNotCalled(t,"Handle", tb.OnText, mock.Anything)
		mockedBot.AssertNotCalled(t,"Start")
	})

	t.Run("it should start the bot successfully", func(t *testing.T) {
		mockedBot.On("SetCommands", cmds).Once().Return(nil)
		mockedBot.On("Handle", "/start", mock.Anything).Once().Return(nil, nil)
		mockedBot.On("Handle", "/help", mock.Anything).Once().Return(nil, nil)
		mockedBot.On("Handle", tb.OnPhoto, mock.Anything).Once().Return(nil, nil)
		mockedBot.On("Handle", tb.OnText, mock.Anything).Once().Return(nil, nil)
		mockedBot.On("Start").Once()

		require.Nil(t, b.Start())

		mockedBot.AssertExpectations(t)
	})
}

func TestStop(t *testing.T) {
	mockedBot := new(mb.TelegramBot)
	mockedBot.On("Stop").Once()

	bot.NewBot(bot.WithTelegramBot(mockedBot)).Stop()

	mockedBot.AssertExpectations(t)
}
