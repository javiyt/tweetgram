package main

import (
	"bytes"
	"flag"
	"log"
	"os"
	"os/signal"

	_ "embed"

	"github.com/javiyt/tweettgram/internal/di"
	"github.com/subosito/gotenv"
)

//go:embed env
var envFile []byte

//go:embed env.test
var envTestFile []byte

func main() {
	initializeConfiguration()

	tBot, err := di.ProvideBot()
	if err != nil {
		log.Fatal(err)
		return
	}

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

func initializeConfiguration() {
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
}
