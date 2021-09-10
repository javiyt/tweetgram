package bot

import (
	"strings"

	tb "gopkg.in/tucnak/telebot.v2"
)

type botHandler struct {
	handlerFunc func(*tb.Message)
	help        string
}

func (b *Bot) getHandlers() map[string]botHandler {
	return map[string]botHandler{
		"/hello": {
			handlerFunc: func(m *tb.Message) {
				b.bot.Send(m.Sender, "Hello World!")
			},
			help: "Show a hello world message",
		},
	}
}

func (b *Bot) setCommands() {
	var cmds []tb.Command
	for c, h := range b.getHandlers() {
		cmds = append(cmds, tb.Command{
			Text:        strings.Replace(c, "/", "", 1),
			Description: h.help,
		})
	}

	b.bot.SetCommands(cmds)
}

func (b *Bot) setUpHandlers() {
	for c, h := range b.getHandlers() {
		b.bot.Handle(c, h.handlerFunc)
	}
}
