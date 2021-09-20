package bot

import (
	"encoding/json"
	"log"
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
	if m.Caption == "" && m.AlbumID == "" {
		log.Print("not album not caption")
		return
	}

	js, _ := json.Marshal(m)
	log.Print(string(js))
	if m.AlbumID != "" {
		log.Print("message to channel")
		b.albumChan <- m
		return
	}

	b.bot.Send(tb.ChatID(b.cfg.BroadcastChannel), &tb.Photo{
		Caption: caption,
		File: tb.File{
			FileID:   m.Photo.FileID,
			FileURL:  m.Photo.FileURL,
			FileSize: m.Photo.FileSize,
		},
	})
}

func (b *Bot) handleText(m *tb.Message) {
	msg := strings.TrimSpace(m.Text)
	if msg == "" {
		return
	}

	b.bot.Send(tb.ChatID(b.cfg.BroadcastChannel), msg)
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
