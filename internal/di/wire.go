//+build wireinject

package di

import (
	"time"

	"github.com/google/wire"
	"github.com/javiyt/tweettgram/internal/bot"
	"github.com/javiyt/tweettgram/internal/config"

	tb "gopkg.in/tucnak/telebot.v2"
)

func ProvideBot() (*bot.Bot, error) {
	panic(wire.Build(
		provideConfiguration,
		provideTBot,
		provideBotOptions,
		bot.NewBot,
	))
}

func provideConfiguration() config.EnvConfig {
	wire.Build(config.NewEnvConfig)

	return config.EnvConfig{}
}

func provideTBot() (*tb.Bot, error) {
	panic(wire.Build(provideConfiguration, provideTBotSettings, tb.NewBot))
}

func provideTBotSettings(cfg config.EnvConfig) tb.Settings {
	return tb.Settings{
		Token:  cfg.BotToken,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	}
}

func provideBotOptions(b *tb.Bot, cfg config.EnvConfig) []bot.BotOption {
	return []bot.BotOption{
		bot.WithTelegramBot(b),
		bot.WithConfig(cfg),
	}
}
