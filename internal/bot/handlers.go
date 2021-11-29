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

func (b *Bot) handleStartCommand(m *TelegramMessage) {
	_ = b.bot.Send(m.SenderID, "Thanks for using the bot! You can type /help command to know what can I do")
}

func (b *Bot) handleHelpCommand(m *TelegramMessage) {
	user, err := strconv.Atoi(m.SenderID)
	if err != nil {
		return
	}

	var helpText string
	for _, h := range b.getCommands(b.cfg.IsAdmin(user)) {
		helpText += "/" + h.Text + " - " + h.Description + "\n"
	}

	_ = b.bot.Send(m.SenderID, helpText)
}

func (b *Bot) handleStopNotificationsCommand(m *TelegramMessage) {
	ce := pubsub.CommandEvent{Command: pubsub.StopCommand}
	if m.Payload != "" {
		ce.Handler = m.Payload
	}

	marshal, _ := easyjson.Marshal(ce)

	_ = b.q.Publish(pubsub.CommandTopic.String(), message.NewMessage(watermill.NewUUID(), marshal))
}

func (b *Bot) handlePhoto(m *TelegramMessage) {
	caption := strings.TrimSpace(m.Photo.Caption)
	if caption == "" {
		return
	}

	fileReader, err := b.bot.GetFile(m.Photo.FileID)
	if err != nil {
		return
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

	_ = b.q.Publish(pubsub.PhotoTopic.String(), message.NewMessage(watermill.NewUUID(), mb))
}

func (b *Bot) handleText(m *TelegramMessage) {
	msg := strings.TrimSpace(m.Text)
	if msg == "" {
		return
	}

	mb, _ := easyjson.Marshal(pubsub.TextEvent{Text: msg})

	_ = b.q.Publish(pubsub.TextTopic.String(), message.NewMessage(watermill.NewUUID(), mb))
}
