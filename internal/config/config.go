package config

import (
	"github.com/kelseyhightower/envconfig"
)

type AppConfig struct {
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

func NewAppConfig() (AppConfig, error) {
	var e AppConfig

	err := envconfig.Process("", &e)
	if err != nil {
		return AppConfig{}, err
	}

	return e, nil
}

func (ec AppConfig) IsAdmin(userID int) bool {
	for _, v := range ec.Admins {
		if v == userID {
			return true
		}
	}

	return false
}

func (ec AppConfig) IsProd() bool {
	return ec.Environment == "PROD"
}
