package bot

import (
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

func (b *Bot) onlyAdmins(f func(*tb.Message)) func(*tb.Message) {
	return func(m *tb.Message) {
		if !b.cfg.IsAdmin(m.Sender.ID) {
			return
		}

		f(m)
	}
}
