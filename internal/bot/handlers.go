package bot

import (
	tb "gopkg.in/tucnak/telebot.v2"
)

func (b *Bot) handleStartCommand(m *tb.Message) {
	b.bot.Send(m.Sender, "Thanks for using the bot! You can type /help command to know what can I do")
}

func (b *Bot) handleHelpCommand(m *tb.Message) {
	var helpText string
	for c, h := range b.getHandlers() {
		helpText += c + " - " + h.help + "\n"
	}

	b.bot.Send(m.Sender, helpText)
}
