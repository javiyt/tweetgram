package bot

import (
	"bytes"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/javiyt/tweetgram/internal/pubsub"
	"github.com/mailru/easyjson"
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
	caption := strings.TrimSpace(m.Caption)
	if caption == "" {
		return
	}

	fileReader, err := b.bot.GetFile(m.Photo.MediaFile())
	if err != nil {
		return
	}

	fileContent := new(bytes.Buffer)
	_, _ = fileContent.ReadFrom(fileReader)

	mb, _ := easyjson.Marshal(pubsub.PhotoEvent{
		Caption:  caption,
		FileID:   m.Photo.FileID,
		FileURL:  m.Photo.FileURL,
		FileSize: m.Photo.FileSize,
		FileContent: fileContent.Bytes(),
	})

	_ = b.q.Publish(pubsub.PhotoTopic.String(), message.NewMessage(watermill.NewUUID(), mb))
}

func (b *Bot) handleText(m *tb.Message) {
	msg := strings.TrimSpace(m.Text)
	if msg == "" {
		return
	}

	mb, _ := easyjson.Marshal(pubsub.TextEvent{
		Text: msg,
	})

	_ = b.q.Publish(pubsub.TextTopic.String(), message.NewMessage(watermill.NewUUID(), mb))
}
