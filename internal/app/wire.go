//go:build wireinject
// +build wireinject

package app

import (
	"net/http"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	"github.com/javiyt/tweettgram/internal/handlers"
	hse "github.com/javiyt/tweettgram/internal/handlers/error"
	hstl "github.com/javiyt/tweettgram/internal/handlers/telegram"
	hstw "github.com/javiyt/tweettgram/internal/handlers/twitter"
	"github.com/javiyt/tweettgram/internal/pubsub"
	"github.com/sirupsen/logrus"

	"github.com/dghubble/oauth1"
	"github.com/google/wire"
	"github.com/javiyt/tweettgram/internal/bot"
	"github.com/javiyt/tweettgram/internal/config"
	"github.com/javiyt/tweettgram/internal/twitter"

	gt "github.com/javiyt/go-twitter/twitter"
	tb "gopkg.in/tucnak/telebot.v2"
)

var (
	twitterClient = wire.NewSet(provideTwitterHttpClient, provideTwitterClient, wire.Bind(new(bot.TwitterClient), new(*twitter.Client)))
	telegramBot   = wire.NewSet(provideTBot, wire.Bind(new(bot.TelegramBot), new(*tb.Bot)))
	queue         = wire.NewSet(provideGoChannelQueue, wire.Bind(new(pubsub.Queue), new(*gochannel.GoChannel)))
	queueInstance *gochannel.GoChannel
)

func ProvideApp() (*App, func(), error) {
	panic(wire.Build(
		provideBotProvider,
		twitterClient,
		provideConfiguration,
		telegramBot,
		queue,
		provideLogger,
		provideTelegramHandler,
		provideTwitterHandler,
		provideErrorHandler,
		provideHandlerManager,
		NewApp,
	))
}

func provideBotProvider() botProvider {
	return provideBot
}

func provideBot() (bot.AppBot, error) {
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
	if queueInstance == nil {
		queueInstance = gochannel.NewGoChannel(
			gochannel.Config{},
			watermill.NewStdLogger(true, true),
		)
	}
	return queueInstance
}

func provideBotOptions(b bot.TelegramBot, cfg config.EnvConfig, tc bot.TwitterClient, gq pubsub.Queue) []bot.Option {
	return []bot.Option{
		bot.WithTelegramBot(b),
		bot.WithConfig(cfg),
		bot.WithTwitterClient(tc),
		bot.WithQueue(gq),
	}
}

func provideLogger() (*logrus.Logger, func()) {
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		DisableColors: true,
		FullTimestamp: true,
	})
	logger.SetReportCaller(true)
	logger.SetLevel(logrus.DebugLevel)

	return logger, func() {
		logger.Exit(0)
	}
}

func provideTelegramHandler(config.EnvConfig, bot.TelegramBot, pubsub.Queue) *hstl.Telegram {
	wire.Build(hstl.NewTelegram)
	return &hstl.Telegram{}
}

func provideTwitterHandler(bot.TwitterClient, pubsub.Queue) *hstw.Twitter {
	wire.Build(hstw.NewTwitter)
	return &hstw.Twitter{}
}

func provideErrorHandler(pubsub.Queue, *logrus.Logger) *hse.ErrorHandler {
	wire.Build(hse.NewErrorHandler)
	return &hse.ErrorHandler{}
}

func provideHandlerManager(tlh *hstl.Telegram, twh *hstw.Twitter, eh *hse.ErrorHandler) *handlers.Manager {
	return handlers.NewHandlersManager(tlh, twh, eh)
}
