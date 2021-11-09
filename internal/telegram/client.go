package telegram

import (
	"errors"
	"io"
	"strconv"
	"strings"

	"github.com/javiyt/tweetgram/internal/bot"
	tb "gopkg.in/tucnak/telebot.v2"
)

const telegramMessageLength = 4096

type Bot struct {
	b *tb.Bot
}

func NewBot(b *tb.Bot) bot.TelegramBot {
	return &Bot{b: b}
}

func (b *Bot) Start() {
	b.b.Start()
}

func (b *Bot) Stop() {
	b.b.Stop()
}

func (b *Bot) SetCommands(commands []bot.TelegramBotCommand) error {
	var cmd []tb.Command
	for i := range commands {
		cmd = append(cmd, tb.Command{
			Text:        commands[i].Text,
			Description: commands[i].Description,
		})
	}

	return b.b.SetCommands(cmd)
}

func (b *Bot) Handle(endpoint string, handler bot.TelegramHandler) {
	b.b.Handle(endpoint, func(m *tb.Message) {
		var p bot.TelegramPhoto
		if m.Photo != nil && strings.TrimSpace(m.Photo.FileID) != "" {
			p = bot.TelegramPhoto{
				Caption:  m.Caption,
				FileID:   m.Photo.FileID,
				FileURL:  m.Photo.FileURL,
				FileSize: m.Photo.FileSize,
			}
		}
		message := bot.TelegramMessage{
			SenderID:  m.Sender.Recipient(),
			Text:      m.Text,
			Photo:     p,
			IsPrivate: m.Private(),
		}

		handler(&message)
	})
}

func (b *Bot) Send(to string, what interface{}, options ...interface{}) error {
	toInt, err := strconv.ParseFloat(to, 0)
	if err != nil {
		return err
	}

	var whatTB interface{}
	switch v := what.(type) {
	case string:
		var replyTo *tb.Message
		for _, ts := range b.chunks(v, telegramMessageLength) {
			options = append(options, &tb.SendOptions{ReplyTo: replyTo})
			replyTo, err = b.b.Send(tb.ChatID(toInt), ts, options...)
			if err != nil {
				return err
			}
		}

		return nil
	case bot.TelegramPhoto:
		whatTB = &tb.Photo{
			Caption: v.Caption,
			File: tb.File{
				FileID:   v.FileID,
				FileURL:  v.FileURL,
				FileSize: v.FileSize,
			},
		}
	default:
		return errors.New("unsupported type")
	}

	_, err = b.b.Send(tb.ChatID(toInt), whatTB, options...)
	return err
}

func (b *Bot) GetFile(fileID string) (io.ReadCloser, error) {
	return b.b.GetFile(&tb.File{FileID: fileID})
}

func (b *Bot) chunks(s string, chunkSize int) []string {
	if chunkSize >= len(s) {
		return []string{s}
	}

	chunks := make([]string, 0, (len(s)-1)/chunkSize+1)
	currentLen := 0
	currentStart := 0
	for i := range s {
		if currentLen == chunkSize {
			chunks = append(chunks, s[currentStart:i])
			currentLen = 0
			currentStart = i
		}
		currentLen++
	}
	chunks = append(chunks, s[currentStart:])

	return chunks
}
