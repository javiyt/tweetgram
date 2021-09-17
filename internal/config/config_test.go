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
		"BOT_TOKEN":         "asdfg",
		"ADMINS":            "12345",
		"BROADCAST_CHANNEL": "9876543",
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
		}, c)	
	})

	t.Run("it should fail when bot token not present", func(t *testing.T) {
		os.Unsetenv("BOT_TOKEN")

		_, err := config.NewEnvConfig()

		require.EqualError(t, err, "required key BOT_TOKEN missing value")

		os.Setenv("BOT_TOKEN", mocked["BOT_TOKEN"])
	})

	t.Run("it should fail when admins not present", func(t *testing.T) {
		os.Unsetenv("ADMINS")

		_, err := config.NewEnvConfig()

		require.EqualError(t, err, "required key ADMINS missing value")

		os.Setenv("ADMINS", mocked["ADMINS"])
	})

	t.Run("it should fail when broadcast channel not present", func(t *testing.T) {
		os.Unsetenv("BROADCAST_CHANNEL")

		_, err := config.NewEnvConfig()

		require.EqualError(t, err, "required key BROADCAST_CHANNEL missing value")

		os.Setenv("BROADCAST_CHANNEL", mocked["BROADCAST_CHANNEL"])
	})

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