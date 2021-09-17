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
	mockedBot.On("Handle", tb.OnPhoto, mock.Anything).Once().Return(nil, nil)
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
