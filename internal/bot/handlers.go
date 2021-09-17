package bot

import (
	"log"
	"strings"

	tb "gopkg.in/tucnak/telebot.v2"
)

func (b *Bot) handleStartCommand(m *tb.Message) {
	_, _ = b.bot.Send(m.Sender, "Thanks for using the bot! You can type /help command to know what can I do")
}

func (b *Bot) handleHelpCommand(m *tb.Message) {
	var helpText string
	for _, h := range b.getCommands() {
		helpText += "/" + h.Text + " - " + h.Description + "\n"
	}

	_, _ = b.bot.Send(m.Sender, helpText)
}

func (b *Bot) handlePhoto(m *tb.Message) {
	log.Printf("%+v", m)
	caption := strings.TrimSpace(m.Caption)
	if m.Caption == "" {
		return
	}

	b.bot.Send(m.Sender, caption)
}
