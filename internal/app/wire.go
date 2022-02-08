//go:build wireinject
// +build wireinject

package app

import (
	"net/http"
	"os"
	"time"

	"github.com/javiyt/tweetgram/internal/telegram"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	"github.com/javiyt/tweetgram/internal/handlers"
	hse "github.com/javiyt/tweetgram/internal/handlers/error"
	hstl "github.com/javiyt/tweetgram/internal/handlers/telegram"
	hstw "github.com/javiyt/tweetgram/internal/handlers/twitter"
	"github.com/javiyt/tweetgram/internal/pubsub"
	"github.com/sirupsen/logrus"

	"github.com/dghubble/oauth1"
	"github.com/google/wire"
	"github.com/javiyt/tweetgram/internal/bot"
	"github.com/javiyt/tweetgram/internal/config"
	"github.com/javiyt/tweetgram/internal/twitter"

	gt "github.com/javiyt/go-twitter/twitter"
	tb "gopkg.in/telebot.v3"
)

type customHandlerGenerator func() []handlers.EventHandler

var (
	queueInstance *gochannel.GoChannel
	twitterClient = wire.NewSet(
		provideTwitterHttpClient,
		provideTwitterClient,
		wire.Bind(new(bot.TwitterClient), new(*twitter.Client)),
	)
	queue        = wire.NewSet(provideGoChannelQueue, wire.Bind(new(pubsub.Queue), new(*gochannel.GoChannel)))
	telegramDeps = wire.NewSet(provideConfiguration, provideTBot, queue)
	twitterDeps  = wire.NewSet(provideConfiguration, twitterClient, queue)
	errorDeps    = wire.NewSet(provideConfiguration, queue, provideLogger)
)

func ProvideApp() (*App, func(), error) {
	panic(wire.Build(
		provideBotProvider,
		initializeCustomHandlers,
		provideHandlers,
		wire.NewSet(queue, provideHandlerManager),
		NewApp,
	))
}

func provideBotProvider() botProvider {
	return provideBot
}

func provideBot() (bot.AppBot, error) {
	panic(wire.Build(
		provideConfiguration,
		provideTBot,
		twitterClient,
		queue,
		provideBotOptions,
		bot.NewBot,
	))
}

func provideConfiguration() (config.AppConfig, error) {
	panic(wire.Build(config.NewAppConfig))
}

func provideTBot() (bot.TelegramBot, error) {
	panic(wire.Build(provideConfiguration, provideTBotSettings, tb.NewBot, telegram.NewBot))
}

func provideTBotSettings(cfg config.AppConfig) tb.Settings {
	return tb.Settings{
		Token:  cfg.BotToken,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	}
}

func provideTwitterClient(*http.Client) *twitter.Client {
	wire.Build(gt.NewClient, twitter.NewTwitterClient)

	return &twitter.Client{}
}

func provideTwitterHttpClient(cfg config.AppConfig) *http.Client {
	return oauth1.NewConfig(cfg.TwitterAPIKey, cfg.TwitterAPISecret).
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

func provideBotOptions(b bot.TelegramBot, cfg config.AppConfig, tc bot.TwitterClient, gq pubsub.Queue) []bot.Option {
	return []bot.Option{
		bot.WithTelegramBot(b),
		bot.WithConfig(cfg),
		bot.WithTwitterClient(tc),
		bot.WithQueue(gq),
	}
}

func provideLogger(cfg config.AppConfig) (*logrus.Logger, func()) {
	var (
		file *os.File
		err  error
	)

	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		DisableColors: true,
		FullTimestamp: true,
	})
	logger.SetReportCaller(true)

	lvl := logrus.DebugLevel
	if cfg.IsProd() {
		lvl = logrus.ErrorLevel

		if cfg.LogFile != "" {
			file, err = os.OpenFile(cfg.LogFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o755)
			if err != nil {
				logger.Fatal(err)
			}
			logger.SetOutput(file)
		}
	}

	logger.SetLevel(lvl)

	return logger, func() {
		logger.Exit(0)
		_ = file.Close()
	}
}

func provideTelegramOptions(cfg config.AppConfig, tb bot.TelegramBot, pq pubsub.Queue) []hstl.Option {
	return []hstl.Option{
		hstl.WithAppConfig(cfg),
		hstl.WithTelegramBot(tb),
		hstl.WithQueue(pq),
	}
}

func provideTelegramHandler() (*hstl.Telegram, error) {
	panic(wire.Build(telegramDeps, provideTelegramOptions, hstl.NewTelegram))
}

func provideTwitterOptions(tc bot.TwitterClient, pq pubsub.Queue) []hstw.Option {
	return []hstw.Option{
		hstw.WithTwitterClient(tc),
		hstw.WithQueue(pq),
	}
}

func provideTwitterHandler() (*hstw.Twitter, error) {
	panic(wire.Build(twitterDeps, provideTwitterOptions, hstw.NewTwitter))
}

func provideErrorHandler() (*hse.ErrorHandler, func(), error) {
	panic(wire.Build(errorDeps, hse.NewErrorHandler))
}

func provideHandlers(customHandlers customHandlerGenerator) ([]handlers.EventHandler, func(), error) {
	telegramHandler, err := provideTelegramHandler()
	if err != nil {
		return nil, nil, err
	}
	twitterHandler, err := provideTwitterHandler()
	if err != nil {
		return nil, nil, err
	}
	errorHandler, cleanup, err := provideErrorHandler()
	if err != nil {
		return nil, nil, err
	}

	return append(customHandlers(),
		telegramHandler,
		twitterHandler,
		errorHandler,
	), cleanup, nil
}

func provideHandlerManager(q pubsub.Queue, h []handlers.EventHandler) *handlers.Manager {
	return handlers.NewHandlersManager(q, h...)
}

func initializeCustomHandlers() customHandlerGenerator {
	return func() []handlers.EventHandler {
		return nil
	}
}
