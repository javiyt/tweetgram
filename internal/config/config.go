package config

import (
	"github.com/kelseyhightower/envconfig"
)

type EnvConfig struct {
	BotToken            string `required:"true" split_words:"true"`
	Admins              []int  `required:"true" split_words:"true"`
	BroadcastChannel    int64  `required:"true" split_words:"true"`
	TwitterAPIKey       string `required:"true" split_words:"true"`
	TwitterAPISecret    string `required:"true" split_words:"true"`
	TwitterBearerToken  string `required:"true" split_words:"true"`
	TwitterAccessToken  string `required:"true" split_words:"true"`
	TwitterAccessSecret string `required:"true" split_words:"true"`
	Environment         string `required:"true" split_words:"true"`
	LogFile             string `split_words:"true"`
}

func NewEnvConfig() (EnvConfig, error) {
	var e EnvConfig

	err := envconfig.Process("", &e)
	if err != nil {
		return EnvConfig{}, err
	}

	return e, nil
}

func (ec EnvConfig) IsAdmin(userID int) bool {
	for _, v := range ec.Admins {
		if v == userID {
			return true
		}
	}

	return false
}

func (ec EnvConfig) IsProd() bool {
	return ec.Environment == "PROD"
}
