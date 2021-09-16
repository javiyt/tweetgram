package bot

import (
	"strings"

	"github.com/javiyt/tweettgram/internal/config"
	tb "gopkg.in/tucnak/telebot.v2"
)

type Bot struct {
	bot *tb.Bot
	cfg config.EnvConfig
}

type BotOption func(b *Bot)

type botHandler struct {
	handlerFunc func(*tb.Message)
	help        string
	filters     []filterFunc
}

func WithTelegramBot(tb *tb.Bot) BotOption {
	return func(b *Bot) {
		b.bot = tb
	}
}

func WithConfig(cfg config.EnvConfig) BotOption {
	return func(b *Bot) {
		b.cfg = cfg
	}
}

func NewBot(options ...BotOption) *Bot {
	b := &Bot{}

	for _, o := range options {
		o(b)
	}

	return b
}

func (b *Bot) Start() error {
	if err := b.setCommands(); err != nil {
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
				b.validChannel,
			},
		},
	}
}

func (b *Bot) setCommands() error {
	var cmds []tb.Command
	for c, h := range b.getHandlers() {
		if strings.TrimSpace(h.help) != "" {
			cmds = append(cmds, tb.Command{
				Text:        strings.Replace(c, "/", "", 1),
				Description: h.help,
			})
		}
	}

	return b.bot.SetCommands(cmds)
}

func (b *Bot) setUpHandlers() {
	for c, h := range b.getHandlers() {
		exec := h.handlerFunc

		for _, v := range h.filters {
			exec = v(exec)
		}

		b.bot.Handle(c, h.handlerFunc)
	}
}
