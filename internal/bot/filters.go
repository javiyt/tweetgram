package bot

import (
	"strconv"
)

type filterFunc func(f TelegramHandler) TelegramHandler

func (b *Bot) onlyPrivate(f TelegramHandler) TelegramHandler {
	return func(m *TelegramMessage) error {
		if !m.IsPrivate {
			return nil
		}

		return f(m)
	}
}

func (b *Bot) onlyAdmins(f TelegramHandler) TelegramHandler {
	return func(m *TelegramMessage) error {
		senderID, err := strconv.Atoi(m.SenderID)
		if err != nil {
			return err
		}

		if !b.cfg.IsAdmin(senderID) {
			return nil
		}

		return f(m)
	}
}
