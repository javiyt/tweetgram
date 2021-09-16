package bot

import (
	"encoding/json"
	"log"

	tb "gopkg.in/tucnak/telebot.v2"
)

type filterFunc func(f func(*tb.Message)) func(*tb.Message)

func (b *Bot) onlyPrivate(f func(*tb.Message)) func(*tb.Message) {
	return func(m *tb.Message) {
		if !m.Private() {
			return
		}

		f(m)
	}
}

func (b *Bot) validChannel(f func(*tb.Message)) func(*tb.Message) {
	return func(m *tb.Message) {
		jm, _ := json.Marshal(m)
		log.Printf("message: %s", string(jm))
		
		if !b.cfg.IsValidChannel(m.Chat.ID) {
			return
		}

		f(m)
	}
}
