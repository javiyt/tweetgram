package bot

import tb "gopkg.in/tucnak/telebot.v2"

type Bot struct {
	bot *tb.Bot
}

type BotOption func(b *Bot)

func WithTelegramBot(tb *tb.Bot) BotOption {
	return func(b *Bot) {
		b.bot = tb
	}
}

func NewBot(options ...BotOption) *Bot {
	b := &Bot{}

	for _, o := range options {
		o(b)
	}

	return b
}

func (b *Bot) Start() {
	b.setCommands()
	b.setUpHandlers()

	b.bot.Start()
}

func (b *Bot) Stop() {
	b.bot.Stop()
}