package bot

import (
	"sort"
	"strings"

	"github.com/javiyt/tweettgram/internal/config"
	tb "gopkg.in/tucnak/telebot.v2"
)

type TelegramBot interface {
	Start()
	Stop()
	SetCommands(cmds []tb.Command) error
	Handle(endpoint interface{}, handler interface{})
	Send(to tb.Recipient, what interface{}, options ...interface{}) (*tb.Message, error)
}

type AppBot interface {
	Start() error
	Stop()
}

type TwitterClient interface{
	SendUpdate(string) error
}

type Bot struct {
	bot TelegramBot
	tc  TwitterClient
	cfg config.EnvConfig
}

type BotOption func(b *Bot)

type botHandler struct {
	handlerFunc func(*tb.Message)
	help        string
	filters     []filterFunc
}

func WithTelegramBot(tb TelegramBot) BotOption {
	return func(b *Bot) {
		b.bot = tb
	}
}

func WithConfig(cfg config.EnvConfig) BotOption {
	return func(b *Bot) {
		b.cfg = cfg
	}
}

func WithTwitterClient(tc TwitterClient) BotOption {
	return func(b *Bot) {
		b.tc = tc
	}
}

func NewBot(options ...BotOption) AppBot {
	b := &Bot{}

	for _, o := range options {
		o(b)
	}

	return b
}

func (b *Bot) Start() error {
	if err := b.setCommandList(); err != nil {
		return err
	}

	b.setUpHandlers()
	b.bot.Start()

	return nil
}

func (b *Bot) Stop() {
	b.bot.Stop()
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

func (b *Bot) getCommands() []tb.Command {
	var cmds []tb.Command
	for c, h := range b.getHandlers() {
		if strings.TrimSpace(h.help) != "" {
			cmds = append(cmds, tb.Command{
				Text:        strings.Replace(c, "/", "", 1),
				Description: h.help,
			})
		}
	}

	sort.Slice(cmds, func(i, j int) bool {
		return cmds[i].Text < cmds[j].Text
	})

	return cmds
}

func (b *Bot) setCommandList() error {
	return b.bot.SetCommands(b.getCommands())
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
