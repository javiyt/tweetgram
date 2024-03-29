package bot

import (
	"context"
	"io"
	"sort"
	"strings"

	"github.com/javiyt/tweetgram/internal/pubsub"

	"github.com/javiyt/tweetgram/internal/config"
	tb "gopkg.in/telebot.v3"
)

type TelegramBot interface {
	Start()
	Stop()
	SetCommands([]TelegramBotCommand) error
	Handle(string, TelegramHandler)
	Send(string, interface{}, ...interface{}) error
	GetFile(string) (io.ReadCloser, error)
}

type TelegramHandler func(TelegramMessage) error

type TelegramBotCommand struct {
	Text        string
	Description string
	IsAdmin     bool
}

type TelegramMessage struct {
	SenderID  string
	Text      string
	Payload   string
	Photo     TelegramPhoto
	IsPrivate bool
}

type TelegramPhoto struct {
	Caption  string
	FileID   string
	FileURL  string
	FileSize int64
}

type AppBot interface {
	Start(ctx context.Context) error
	Run()
	Stop()
}

type TwitterClient interface {
	SendUpdate(string) error
	SendUpdateWithPhoto(string, []byte) error
}

type Bot struct {
	bot TelegramBot
	tc  TwitterClient
	cfg config.AppConfig
	q   pubsub.Queue
}

type Option func(b *Bot)

type botHandler struct {
	handlerFunc TelegramHandler
	help        string
	isAdmin     bool
	filters     []filterFunc
}

func WithTelegramBot(tb TelegramBot) Option {
	return func(b *Bot) {
		b.bot = tb
	}
}

func WithConfig(cfg config.AppConfig) Option {
	return func(b *Bot) {
		b.cfg = cfg
	}
}

func WithTwitterClient(tc TwitterClient) Option {
	return func(b *Bot) {
		b.tc = tc
	}
}

func WithQueue(q pubsub.Queue) Option {
	return func(b *Bot) {
		b.q = q
	}
}

func NewBot(options ...Option) AppBot {
	b := &Bot{}

	for _, o := range options {
		o(b)
	}

	return b
}

func (b *Bot) Start(context.Context) error {
	if err := b.setCommandList(); err != nil {
		return err
	}

	b.setUpHandlers()

	return nil
}

func (b *Bot) Run() {
	b.bot.Start()
}

func (b *Bot) Stop() {
	b.bot.Stop()
	_ = b.q.Close()
}

func (b *Bot) getHandlers() map[string]botHandler {
	return map[string]botHandler{
		"/start": {
			handlerFunc: b.handleStartCommand,
			help:        "Start a conversation with the bot",
			filters: []filterFunc{
				b.onlyPrivate,
			},
		},
		"/help": {
			handlerFunc: b.handleHelpCommand,
			help:        "Show help",
			filters: []filterFunc{
				b.onlyPrivate,
			},
		},
		"/stop": {
			handlerFunc: b.handleStopNotificationsCommand,
			help:        "Stop notifications for all handlers or specific handler",
			filters: []filterFunc{
				b.onlyPrivate,
				b.onlyAdmins,
			},
			isAdmin: true,
		},
		tb.OnPhoto: {
			handlerFunc: b.handlePhoto,
			filters: []filterFunc{
				b.onlyPrivate,
				b.onlyAdmins,
			},
		},
		tb.OnText: {
			handlerFunc: b.handleText,
			filters: []filterFunc{
				b.onlyPrivate,
				b.onlyAdmins,
			},
		},
	}
}

func (b *Bot) getCommands(includeAdmin bool) []TelegramBotCommand {
	var cmd []TelegramBotCommand

	for c, h := range b.getHandlers() {
		if !includeAdmin && h.isAdmin {
			continue
		}

		if strings.TrimSpace(h.help) != "" {
			cmd = append(cmd, TelegramBotCommand{
				Text:        strings.Replace(c, "/", "", 1),
				Description: h.help,
			})
		}
	}

	sort.Slice(cmd, func(i, j int) bool {
		return cmd[i].Text < cmd[j].Text
	})

	return cmd
}

func (b *Bot) setCommandList() error {
	return b.bot.SetCommands(b.getCommands(false))
}

func (b *Bot) setUpHandlers() {
	for c, h := range b.getHandlers() {
		exec := h.handlerFunc

		for _, v := range h.filters {
			exec = v(exec)
		}

		b.bot.Handle(c, exec)
	}
}
