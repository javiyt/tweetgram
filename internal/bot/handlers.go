package bot

import (
	tb "gopkg.in/tucnak/telebot.v2"
)

func (b *Bot) helloHandler(m *tb.Message) {
	b.bot.Send(m.Sender, "Hello World!")
}
