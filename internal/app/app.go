package app

import (
	"bytes"
	"context"
	"fmt"

	"github.com/javiyt/tweetgram/internal/bot"
	"github.com/javiyt/tweetgram/internal/handlers"
	"github.com/subosito/gotenv"
)

type botProvider func() (bot.AppBot, error)

type App struct {
	bp botProvider
	tb bot.AppBot
	hm *handlers.Manager
}

func InitializeConfiguration(testBot bool, envFile []byte, envTestFile []byte) error {
	err := gotenv.OverApply(bytes.NewReader(envFile))
	if err != nil {
		return fmt.Errorf("error loading env file: %w", err)
	}

	if testBot {
		err := gotenv.OverApply(bytes.NewReader(envTestFile))
		if err != nil {
			return fmt.Errorf("error loading env.test file: %w", err)
		}
	}

	return nil
}

func NewApp(bp botProvider, hm *handlers.Manager) *App {
	if bp == nil {
		bp = provideBot
	}

	return &App{bp: bp, hm: hm}
}

func (a *App) Start(ctx context.Context) error {
	tBot, err := a.bp()
	if err != nil {
		return fmt.Errorf("error getting bot instance: %w", err)
	}

	if err := tBot.Start(ctx); err != nil {
		return fmt.Errorf("error starting bot: %w", err)
	}

	a.hm.StartHandlers(ctx)
	a.tb = tBot

	return nil
}

func (a *App) Run() {
	a.tb.Run()
}

func (a *App) Stop() {
	a.tb.Stop()
}
