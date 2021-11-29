package config_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/javiyt/tweetgram/internal/config"
	"github.com/stretchr/testify/require"
)

func TestNewEnvConfig(t *testing.T) {
	original := map[string]string{}
	mocked := map[string]string{
		"BOT_TOKEN":             "asdfg",
		"ADMINS":                "12345",
		"BROADCAST_CHANNEL":     "9876543",
		"TWITTER_API_KEY":       "asdfg1234",
		"TWITTER_API_SECRET":    "poiuyt",
		"TWITTER_BEARER_TOKEN":  "qwertyui",
		"TWITTER_ACCESS_TOKEN":  "zxcvbnm",
		"TWITTER_ACCESS_SECRET": "lkjhgfd",
		"ENVIRONMENT":           "testing",
	}

	for k, v := range mocked {
		original[k] = os.Getenv(k)
		_ = os.Setenv(k, v)
	}

	t.Run("it should get whole configuration", func(t *testing.T) {
		c, err := config.NewAppConfig()

		require.NoError(t, err)
		require.Equal(t, config.AppConfig{
			BotToken:            "asdfg",
			Admins:              []int{12345},
			BroadcastChannel:    9876543,
			TwitterAPIKey:       "asdfg1234",
			TwitterAPISecret:    "poiuyt",
			TwitterBearerToken:  "qwertyui",
			TwitterAccessToken:  "zxcvbnm",
			TwitterAccessSecret: "lkjhgfd",
			Environment:         "testing",
			LogFile:             "",
		}, c)
	})

	for k := range mocked {
		k := k
		t.Run(fmt.Sprintf("it should fail when %s not present", k), func(t *testing.T) {
			_ = os.Unsetenv(k)

			_, err := config.NewAppConfig()

			require.EqualError(t, err, fmt.Sprintf("required key %s missing value", k))

			_ = os.Setenv(k, mocked[k])
		})
	}

	for k, v := range original {
		_ = os.Setenv(k, v)
	}
}

func TestEnvConfig_IsAdmin(t *testing.T) {
	original := map[string]string{}
	mocked := map[string]string{
		"BOT_TOKEN":             "asdfg",
		"ADMINS":                "12345",
		"BROADCAST_CHANNEL":     "9876543",
		"TWITTER_API_KEY":       "asdfg1234",
		"TWITTER_API_SECRET":    "poiuyt",
		"TWITTER_BEARER_TOKEN":  "qwertyui",
		"TWITTER_ACCESS_TOKEN":  "zxcvbnm",
		"TWITTER_ACCESS_SECRET": "lkjhgfd",
		"ENVIRONMENT":           "testing",
	}

	for k, v := range mocked {
		original[k] = os.Getenv(k)
		_ = os.Setenv(k, v)
	}

	c, _ := config.NewAppConfig()

	t.Run("it should return true when user is admin", func(t *testing.T) {
		require.True(t, c.IsAdmin(12345))
	})

	t.Run("it should return false when user isn't admin", func(t *testing.T) {
		require.False(t, c.IsAdmin(6547))
	})

	for k, v := range original {
		_ = os.Setenv(k, v)
	}
}

func TestEnvConfig_IsProd(t *testing.T) {
	original := map[string]string{}
	mocked := map[string]string{
		"BOT_TOKEN":             "asdfg",
		"ADMINS":                "12345",
		"BROADCAST_CHANNEL":     "9876543",
		"TWITTER_API_KEY":       "asdfg1234",
		"TWITTER_API_SECRET":    "poiuyt",
		"TWITTER_BEARER_TOKEN":  "qwertyui",
		"TWITTER_ACCESS_TOKEN":  "zxcvbnm",
		"TWITTER_ACCESS_SECRET": "lkjhgfd",
		"ENVIRONMENT":           "testing",
	}

	for k, v := range mocked {
		original[k] = os.Getenv(k)
		_ = os.Setenv(k, v)
	}

	t.Run("it should return true when environment is prod", func(t *testing.T) {
		_ = os.Setenv("ENVIRONMENT", "PROD")
		c, _ := config.NewAppConfig()
		require.True(t, c.IsProd())
	})

	t.Run("it should return false when environment is not prod", func(t *testing.T) {
		_ = os.Setenv("ENVIRONMENT", "testing")
		c, _ := config.NewAppConfig()
		require.False(t, c.IsProd())
	})

	for k, v := range original {
		_ = os.Setenv(k, v)
	}
}
