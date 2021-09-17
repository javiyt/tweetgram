package main

import (
	"bytes"
	"flag"
	"log"
	"os"
	"os/signal"
	"time"

	_ "embed"

	"github.com/javiyt/tweettgram/internal/bot"
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

	tBot := bot.NewBot(
		bot.WithTelegramBot(b),
		bot.WithConfig(cfg),
	)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go func() {
		<-c
		tBot.Stop()
	}()

	if err := tBot.Start(); err != nil {
		log.Fatal(err)
		return
	}
}
