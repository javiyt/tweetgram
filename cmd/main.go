package main

import (
	"flag"
	"log"
	"os"
	"os/signal"

	_ "embed"

	"github.com/javiyt/tweettgram/internal/app"
)

//go:embed env
var envFile []byte

//go:embed env.test
var envTestFile []byte

func main() {
	testBot := flag.Bool("test", false, "Should execute test bot")
	flag.Parse()

	botApp := app.NewApp(nil)
	if err := botApp.InitializeConfiguration(*testBot, envFile, envTestFile); err != nil {
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
	}()
}
