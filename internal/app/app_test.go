package app_test

import (
	"errors"
	"os"
	"testing"

	"github.com/javiyt/tweettgram/internal/app"
	"github.com/javiyt/tweettgram/internal/bot"
	"github.com/javiyt/tweettgram/internal/handlers"
	"github.com/stretchr/testify/require"

	mockBot "github.com/javiyt/tweettgram/mocks/bot"
)

func TestInitializeConfiguration(t *testing.T) {
	envFile := []byte("BOT_TOKEN=asdfg")
	envTestFile := []byte("BOT_TOKEN=qwert")

	t.Run("it should load configuration from environment file when not in test env", func(t *testing.T) {
		e := app.InitializeConfiguration(false, envFile, envTestFile)

		require.NoError(t, e)
		require.Equal(t, "asdfg", os.Getenv("BOT_TOKEN"))
		_ = os.Unsetenv("BOT_TOKEN")
	})

	t.Run("it should load test configuration from environment file when in test env", func(t *testing.T) {
		e := app.InitializeConfiguration(true, envFile, envTestFile)

		require.NoError(t, e)
		require.Equal(t, "qwert", os.Getenv("BOT_TOKEN"))
		_ = os.Unsetenv("BOT_TOKEN")
	})

	t.Run("it should fail when not a valid env file", func(t *testing.T) {
		e := app.InitializeConfiguration(true, []byte("BOT_TOKEN"), envTestFile)

		require.EqualError(t, e, "error loading env file: line `BOT_TOKEN` doesn't match format")
	})

	t.Run("it should fail when not a valid env.test file", func(t *testing.T) {
		e := app.InitializeConfiguration(true, envFile, []byte("BOT_TOKEN"))

		require.EqualError(t, e, "error loading env.test file: line `BOT_TOKEN` doesn't match format")
		_ = os.Unsetenv("BOT_TOKEN")
	})
}

func TestStart(t *testing.T) {
	t.Run("it should fail when getting bot instance", func(t *testing.T) {
		mbp := func() (bot.AppBot, error) {
			return nil, errors.New("bot instance not ready")
		}

		a := app.NewApp(mbp, handlers.NewHandlersManager())
		e := a.Start()

		require.EqualError(t, e, "error getting bot instance: bot instance not ready")
	})

	t.Run("it should fail when starting bot instance", func(t *testing.T) {
		mb := new(mockBot.AppBot)
		mbp := func() (bot.AppBot, error) {
			return mb, nil
		}
		mb.On("Start").Once().Return(errors.New("could not start"))

		a := app.NewApp(mbp, handlers.NewHandlersManager())
		e := a.Start()

		require.EqualError(t, e, "error starting bot: could not start")
		mb.AssertExpectations(t)
	})

	t.Run("it should start bot instance", func(t *testing.T) {
		mb := new(mockBot.AppBot)
		mbp := func() (bot.AppBot, error) {
			return mb, nil
		}
		mb.On("Start").Once().Return(nil)

		a := app.NewApp(mbp, handlers.NewHandlersManager())
		e := a.Start()

		require.NoError(t, e)
		mb.AssertExpectations(t)
	})
}

func TestRun(t *testing.T) {
	mb := new(mockBot.AppBot)
	mbp := func() (bot.AppBot, error) {
		return mb, nil
	}
	mb.On("Start").Once().Return(nil)
	mb.On("Run").Once()

	a := app.NewApp(mbp, handlers.NewHandlersManager())
	_ = a.Start()
	a.Run()

	mb.AssertExpectations(t)
}

func TestStop(t *testing.T) {
	mb := new(mockBot.AppBot)
	mbp := func() (bot.AppBot, error) {
		return mb, nil
	}
	mb.On("Start").Once().Return(nil)
	mb.On("Stop").Once()

	a := app.NewApp(mbp, handlers.NewHandlersManager())
	e := a.Start()
	a.Stop()

	require.NoError(t, e)
	mb.AssertExpectations(t)
}
