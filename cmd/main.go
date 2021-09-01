package main

import (
	"bytes"
	"flag"
	"log"
	"time"

	_ "embed"

	"github.com/javiyt/tweettgram/internal/config"
	"github.com/subosito/gotenv"
	tb "gopkg.in/tucnak/telebot.v2"
)

//go:embed env
var envFile []byte

//go:embed env.test
var envTestFile []byte

func main() {
	err := gotenv.Apply(bytes.NewReader(envFile))
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	testBot := flag.Bool("test", false, "Should execute test bot")
	flag.Parse()
	if *testBot {
		err := gotenv.OverApply(bytes.NewReader(envTestFile))
		if err != nil {
			log.Fatal("Error loading .env file")
		}
	}

	cfg := config.NewEnvConfig()

	b, err := tb.NewBot(tb.Settings{
		Token:  cfg.BotToken,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})

	if err != nil {
		log.Fatal(err)
		return
	}

	b.Handle("/hello", func(m *tb.Message) {
		b.Send(m.Sender, "Hello World!")
	})

	b.Start()
}
