//go:build wireinject
// +build wireinject

package app

import (
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	"github.com/javiyt/tweettgram/internal/pubsub"
	"net/http"
	"time"

	"github.com/dghubble/oauth1"
	"github.com/google/wire"
	"github.com/javiyt/tweettgram/internal/bot"
	"github.com/javiyt/tweettgram/internal/config"
	"github.com/javiyt/tweettgram/internal/twitter"

	gt "github.com/javiyt/go-twitter/twitter"
	tb "gopkg.in/tucnak/telebot.v2"
)

var (
	twitterClient = wire.NewSet(provideTwitterHttpClient, provideTwitterClient)
	telegramBot   = wire.NewSet(provideTBot)
	queue         = wire.NewSet(provideGoChannelQueue, wire.Bind(new(pubsub.Queue), new(*gochannel.GoChannel)))
)

func ProvideBot() (bot.AppBot, error) {
	panic(wire.Build(
		provideConfiguration,
		telegramBot,
		twitterClient,
		queue,
		provideBotOptions,
		bot.NewBot,
	))
}

func provideConfiguration() (config.EnvConfig, error) {
	panic(wire.Build(config.NewEnvConfig))
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

func provideTwitterClient(*http.Client) *twitter.Client {
	wire.Build(gt.NewClient, twitter.NewTwitterClient)

	return &twitter.Client{}
}

func provideTwitterHttpClient(cfg config.EnvConfig) *http.Client {
	return oauth1.NewConfig(cfg.TwitterApiKey, cfg.TwitterApiSecret).
		Client(oauth1.NoContext, oauth1.NewToken(cfg.TwitterAccessToken, cfg.TwitterAccessSecret))
}

func provideGoChannelQueue() *gochannel.GoChannel {
	return gochannel.NewGoChannel(
		gochannel.Config{},
		watermill.NewStdLogger(false, false),
	)
}

func provideBotOptions(b *tb.Bot, cfg config.EnvConfig, tc *twitter.Client, gq pubsub.Queue) []bot.Option {
	return []bot.Option{
		bot.WithTelegramBot(b),
		bot.WithConfig(cfg),
		bot.WithTwitterClient(tc),
		bot.WithQueue(gq),
	}
}
