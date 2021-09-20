package config_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/javiyt/tweettgram/internal/config"
	"github.com/stretchr/testify/require"
)

func TestNewConfig(t *testing.T) {
	original := map[string]string{}
	mocked := map[string]string{
		"BOT_TOKEN":         "asdfg",
		"ADMINS":            "12345",
		"BROADCAST_CHANNEL": "9876543",
		"TWITTER_API_KEY": "asdfg1234",
		"TWITTER_API_SECRET": "poiuyt",
		"TWITTER_BEARER_TOKEN": "qwertyui",
		"TWITTER_ACCESS_TOKEN": "zxcvbnm",
		"TWITTER_ACCESS_SECRET": "lkjhgfd",
	}

	for k, v := range mocked {
		original[k] = os.Getenv(k)
		os.Setenv(k, v)
	}

	t.Run("it should get whole configuration", func(t *testing.T) {
		c, err := config.NewEnvConfig()

		require.NoError(t, err)
		require.Equal(t, config.EnvConfig{
			BotToken:         "asdfg",
			Admins:           []int{12345},
			BroadcastChannel: 9876543,
			TwitterApiKey: "asdfg1234",
			TwitterApiSecret: "poiuyt",
			TwitterBearerToken: "qwertyui",
			TwitterAccessToken: "zxcvbnm",
			TwitterAccessSecret: "lkjhgfd",
	
		}, c)	
	})

	for k, mv := range mocked {
		t.Run(fmt.Sprintf("it should fail when %s not present", k), func(t *testing.T) {
			os.Unsetenv(k)
	
			_, err := config.NewEnvConfig()
	
			require.EqualError(t, err, fmt.Sprintf("required key %s missing value", k))
	
			os.Setenv(k, mv)
		})			
	}


	for k, v := range original {
		os.Setenv(k, v)
	}
}

func TestIsAdmin(t *testing.T) {
	original := map[string]string{}
	mocked := map[string]string{
		"BOT_TOKEN":         "asdfg",
		"ADMINS":            "12345",
		"BROADCAST_CHANNEL": "9876543",
		"TWITTER_API_KEY": "asdfg1234",
		"TWITTER_API_SECRET": "poiuyt",
		"TWITTER_BEARER_TOKEN": "qwertyui",
		"TWITTER_ACCESS_TOKEN": "zxcvbnm",
		"TWITTER_ACCESS_SECRET": "lkjhgfd",	
	}

	for k, v := range mocked {
		original[k] = os.Getenv(k)
		os.Setenv(k, v)
	}

	c, _ := config.NewEnvConfig()

	t.Run("it should return true when user is admin", func(t *testing.T) {
		require.True(t, c.IsAdmin(12345))
	})

	t.Run("it should return false when user isn't admin", func(t *testing.T) {
		require.False(t, c.IsAdmin(6547))
	})

	for k, v := range original {
		os.Setenv(k, v)
	}
}