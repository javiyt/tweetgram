package bot

import (
	"encoding/json"
	"log"
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
	if caption == "" && m.AlbumID == "" {
		return
	}

	js, _ := json.Marshal(m)
	log.Print(string(js))
	if m.AlbumID != "" {
		log.Print("message to channel")
		b.albumChan <- m
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

func (b *Bot) handleAlbum() {
	chanAlbums := make(chan []*tb.Photo)
	albums := map[string][]*tb.Photo{}
	var lastAlbumID string

	go func() {
		log.Print("Run album listener")
		defer close(chanAlbums)

		for m := range b.albumChan {
			log.Print("Getting photo from album")
			if m.AlbumID == "" {
				continue
			}

			if _, ok := albums[m.AlbumID]; !ok {
				albums[m.AlbumID] = []*tb.Photo{
					{
						Caption: strings.TrimSpace(m.Caption),
						File: tb.File{
							FileID:   m.Photo.FileID,
							FileURL:  m.Photo.FileURL,
							FileSize: m.Photo.FileSize,
						},
					},
				}
			} else {
				albums[m.AlbumID] = append(
					albums[m.AlbumID],
					&tb.Photo{
						Caption: strings.TrimSpace(m.Caption),
						File: tb.File{
							FileID:   m.Photo.FileID,
							FileURL:  m.Photo.FileURL,
							FileSize: m.Photo.FileSize,
						},
					},
				)
			}

			if lastAlbumID != "" && lastAlbumID != m.AlbumID {
				chanAlbums <- albums[m.AlbumID]
				lastAlbumID = m.AlbumID
				delete(albums, m.AlbumID)
			}
		}
	}()

	go func() {
		for m := range chanAlbums {
			log.Print("Sending album")
			a := tb.Album{}
			for _, v := range m {
				a = append(a, v)
			}
			if _, err := b.bot.SendAlbum(tb.ChatID(b.cfg.BroadcastChannel), a); err != nil {
				log.Print(err.Error())
			}
		}
	}()
}
