package telegram

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/javiyt/tweetgram/internal/bot"
	tb "gopkg.in/telebot.v3"
)

const telegramMessageLength = 4096

type TbBot interface {
	Start()
	Stop()
	SetCommands(opts ...interface{}) error
	Handle(endpoint interface{}, h tb.HandlerFunc, m ...tb.MiddlewareFunc)
	Send(to tb.Recipient, what interface{}, opts ...interface{}) (*tb.Message, error)
	File(file *tb.File) (io.ReadCloser, error)
	FileByID(fileID string) (tb.File, error)
}

type Bot struct {
	b TbBot
}

func NewBot(b TbBot) bot.TelegramBot {
	return &Bot{b: b}
}

func (b *Bot) Start() {
	b.b.Start()
}

func (b *Bot) Stop() {
	b.b.Stop()
}

func (b *Bot) SetCommands(commands []bot.TelegramBotCommand) error {
	cmd := make([]tb.Command, 0, len(commands))

	for i := range commands {
		cmd = append(cmd, tb.Command{
			Text:        commands[i].Text,
			Description: commands[i].Description,
		})
	}

	return b.b.SetCommands(cmd)
}

func (b *Bot) Handle(endpoint string, handler bot.TelegramHandler) {
	b.b.Handle(endpoint, func(m tb.Context) error {
		var p bot.TelegramPhoto
		if m.Message().Photo != nil && strings.TrimSpace(m.Message().Photo.FileID) != "" {
			p = bot.TelegramPhoto{
				Caption:  m.Message().Caption,
				FileID:   m.Message().Photo.FileID,
				FileURL:  m.Message().Photo.FileURL,
				FileSize: m.Message().Photo.FileSize,
			}
		}

		return handler(bot.TelegramMessage{
			SenderID:  fmt.Sprintf("%v", m.Sender().ID),
			Text:      m.Text(),
			Payload:   m.Message().Payload,
			Photo:     p,
			IsPrivate: m.Chat().Private,
		})
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
	fileByID, err := b.b.FileByID(fileID)
	if err != nil {
		return nil, err
	}

	return b.b.File(&fileByID)
}

func (b *Bot) chunks(s string, chunkSize int) []string {
	if chunkSize >= len(s) {
		return []string{s}
	}

	chunks := make([]string, 0, (len(s)-1)/chunkSize+1)
	currentLen, currentStart := 0, 0

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
