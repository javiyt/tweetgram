package config

import (
	"log"

	"github.com/kelseyhightower/envconfig"
)

type EnvConfig struct {
	BotToken         string `required:"true" split_words:"true"`
	Admins           []int  `required:"true" split_words:"true"`
	BroadcastChannel int64  `required:"true" split_words:"true"`
}

func NewEnvConfig() EnvConfig {
	var e EnvConfig
	err := envconfig.Process("", &e)
	if err != nil {
		log.Fatal(err.Error())
	}

	return e
}

func (ec EnvConfig) IsAdmin(userID int) bool {
	for _, v := range ec.Admins {
		if v == userID {
			return true
		}
	}

	return false
}
