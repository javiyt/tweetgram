package app

import (
	"bytes"
	"fmt"

	"github.com/javiyt/tweettgram/internal/bot"
	"github.com/subosito/gotenv"
)

type botProvider func() (bot.AppBot, error)

type App struct {
	bp botProvider
	tb bot.AppBot
}

func NewApp(bp botProvider) *App {
	if bp == nil {
		bp = ProvideBot
	}

	return &App{bp: bp}
}

func (a *App) InitializeConfiguration(testBot bool, envFile []byte, envTestFile []byte) error {
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

func (a *App) Start() error {
	tBot, err := a.bp()
	if err != nil {
		return fmt.Errorf("error getting bot instance: %w", err)
	}

	if err := tBot.Start(); err != nil {
		return fmt.Errorf("error starting bot: %w", err)
	}

	a.tb = tBot

	return nil
}

func (a *App) Stop() {
	a.tb.Stop()
}
