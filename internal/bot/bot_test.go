package bot_test

import (
	"testing"

	"github.com/javiyt/tweetgram/internal/bot"
	mb "github.com/javiyt/tweetgram/mocks/bot"
	mq "github.com/javiyt/tweetgram/mocks/pubsub"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	tb "gopkg.in/tucnak/telebot.v2"
)

type settingCommandError struct{}

func (m settingCommandError) Error() string {
	return "not setting commands"
}

func TestStart(t *testing.T) {
	mockedBot := new(mb.TelegramBot)
	mockedTwitter := new(mb.TwitterClient)

	cmds := []bot.TelegramBotCommand{
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
		mockedBot.On("SetCommands", cmds).Once().Return(settingCommandError{})

		require.EqualError(t, b.Start(nil), "not setting commands")

		mockedBot.AssertExpectations(t)
		mockedBot.AssertNotCalled(t, "Handle", "/start", mock.Anything)
		mockedBot.AssertNotCalled(t, "Handle", "/help", mock.Anything)
		mockedBot.AssertNotCalled(t, "Handle", tb.OnPhoto, mock.Anything)
		mockedBot.AssertNotCalled(t, "Handle", tb.OnText, mock.Anything)
		mockedBot.AssertNotCalled(t, "Start")
	})

	t.Run("it should start the bot successfully", func(t *testing.T) {
		mockedBot.On("SetCommands", cmds).Once().Return(nil)
		mockedBot.On("Handle", "/start", mock.Anything).Once().Return(nil, nil)
		mockedBot.On("Handle", "/help", mock.Anything).Once().Return(nil, nil)
		mockedBot.On("Handle", "/stop", mock.Anything).Once().Return(nil, nil)
		mockedBot.On("Handle", tb.OnPhoto, mock.Anything).Once().Return(nil, nil)
		mockedBot.On("Handle", tb.OnText, mock.Anything).Once().Return(nil, nil)

		require.Nil(t, b.Start(nil))

		mockedBot.AssertExpectations(t)
	})
}

func TestRun(t *testing.T) {
	mockedBot := new(mb.TelegramBot)
	mockedBot.On("Start").Once()

	bot.NewBot(bot.WithTelegramBot(mockedBot)).Run()

	mockedBot.AssertExpectations(t)
}

func TestStop(t *testing.T) {
	mockedBot := new(mb.TelegramBot)
	mockedBot.On("Stop").Once()

	mockedQueue := new(mq.Queue)
	mockedQueue.On("Close").Once().Return(nil)

	bot.NewBot(bot.WithTelegramBot(mockedBot), bot.WithQueue(mockedQueue)).Stop()

	mockedBot.AssertExpectations(t)
	mockedQueue.AssertExpectations(t)
}
