package config_test

import (
	"os"
	"testing"

	"github.com/javiyt/tweettgram/internal/config"
	"github.com/stretchr/testify/require"
)

func TestNewConfig(t *testing.T) {
	original := map[string]string{}
	mocked := map[string]string{
		"BOT_TOKEN": "asdfg",
		"ADMINS": "12345",
	}

	for k, v := range mocked {
		original[k] = os.Getenv(k)
		os.Setenv(k, v)
	}

	c := config.NewEnvConfig()

	require.Equal(t, config.EnvConfig{
		BotToken: "asdfg",
		Admins: []int{12345},
	}, c)

	for k, v := range original {
		os.Setenv(k, v)
	}

}