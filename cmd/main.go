package main

import (
	"flag"
	"log"
	"os"
	"os/signal"

	_ "embed"

	"github.com/javiyt/tweetgram/internal/app"
)

//nolint:gochecknoglobals
//go:embed env
var envFile []byte

//nolint:gochecknoglobals
//go:embed env.test
var envTestFile []byte

func main() {
	testBot := flag.Bool("test", false, "Should execute test bot")
	flag.Parse()

	if err := app.InitializeConfiguration(*testBot, envFile, envTestFile); err != nil {
		log.Fatal(err)
	}

	botApp, cleanup, err := app.ProvideApp()
	if err != nil {
		log.Fatal(err)
	}

	if err := botApp.Start(); err != nil {
		log.Fatal(err)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go func() {
		defer close(c)
		<-c
		botApp.Stop()
		cleanup()
	}()

	botApp.Run()
}
