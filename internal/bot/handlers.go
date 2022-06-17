package bot

import (
	"bytes"
	"strconv"
	"strings"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/javiyt/tweetgram/internal/pubsub"
	"github.com/mailru/easyjson"
)

func (b *Bot) handleStartCommand(m TelegramMessage) error {
	return b.bot.Send(m.SenderID, "Thanks for using the bot! You can type /help command to know what can I do")
}

func (b *Bot) handleHelpCommand(m TelegramMessage) error {
	user, err := strconv.Atoi(m.SenderID)
	if err != nil {
		return err
	}

	var helpText string
	for _, h := range b.getCommands(b.cfg.IsAdmin(user)) {
		helpText += "/" + h.Text + " - " + h.Description + "\n"
	}

	return b.bot.Send(m.SenderID, helpText)
}

func (b *Bot) handleStopNotificationsCommand(m TelegramMessage) error {
	ce := pubsub.CommandEvent{Command: pubsub.StopCommand}
	if m.Payload != "" {
		ce.Handler = m.Payload
	}

	marshal, _ := easyjson.Marshal(ce)

	return b.q.Publish(pubsub.CommandTopic.String(), message.NewMessage(watermill.NewUUID(), marshal))
}

func (b *Bot) handlePhoto(m TelegramMessage) error {
	caption := strings.TrimSpace(m.Photo.Caption)
	if caption == "" {
		return nil
	}

	fileReader, err := b.bot.GetFile(m.Photo.FileID)
	if err != nil {
		return err
	}

	fileContent := new(bytes.Buffer)
	_, _ = fileContent.ReadFrom(fileReader)

	mb, _ := easyjson.Marshal(pubsub.PhotoEvent{
		Caption:     caption,
		FileID:      m.Photo.FileID,
		FileURL:     m.Photo.FileURL,
		FileSize:    m.Photo.FileSize,
		FileContent: fileContent.Bytes(),
	})

	return b.q.Publish(pubsub.PhotoTopic.String(), message.NewMessage(watermill.NewUUID(), mb))
}

func (b *Bot) handleText(m TelegramMessage) error {
	msg := strings.TrimSpace(m.Text)
	if msg == "" {
		return nil
	}

	mb, _ := easyjson.Marshal(pubsub.TextEvent{Text: msg})

	return b.q.Publish(pubsub.TextTopic.String(), message.NewMessage(watermill.NewUUID(), mb))
}
