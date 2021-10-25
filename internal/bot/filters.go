package bot

import (
	"strconv"
)

type filterFunc func(f TelegramHandler) TelegramHandler

func (b *Bot) onlyPrivate(f TelegramHandler) TelegramHandler {
	return func(m *TelegramMessage) {
		if !m.IsPrivate {
			return
		}

		f(m)
	}
}

func (b *Bot) onlyAdmins(f TelegramHandler) TelegramHandler {
	return func(m *TelegramMessage) {
		senderID, err := strconv.Atoi(m.SenderID)
		if err != nil {
			return
		}

		if !b.cfg.IsAdmin(senderID) {
			return
		}

		f(m)
	}
}
