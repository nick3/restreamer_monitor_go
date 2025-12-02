package notification

import (
	"testing"

	"github.com/nick3/restreamer_monitor_go/telegram"
	"github.com/stretchr/testify/assert"
)

func TestNewNotificationManager(t *testing.T) {
	t.Run("telegram disabled", func(t *testing.T) {
		config := Config{
			Telegram: telegram.Config{
				Enabled: false,
			},
		}

		nm, err := NewNotificationManager(config)
		assert.NoError(t, err)
		assert.NotNil(t, nm)
		assert.Nil(t, nm.telegramBot)
	})

	t.Run("telegram enabled with invalid token", func(t *testing.T) {
		config := Config{
			Telegram: telegram.Config{
				Enabled:  true,
				BotToken: "invalid_token",
				ChatIDs:  []int64{123456789},
			},
		}

		nm, err := NewNotificationManager(config)
		assert.Error(t, err)
		assert.Nil(t, nm)
	})
}

func TestNotificationManager_Notifications(t *testing.T) {
	config := Config{
		Telegram: telegram.Config{
			Enabled: false, // Disabled to avoid actual API calls
		},
		Notifications: NotificationConfig{
			SystemEvents:  true,
			MonitorEvents: true,
			RelayEvents:   true,
			ErrorEvents:   true,
		},
	}

	nm, err := NewNotificationManager(config)
	assert.NoError(t, err)

	// These should not panic even with nil telegramBot
	t.Run("system notification", func(t *testing.T) {
		assert.NotPanics(t, func() {
			nm.SendSystemNotification("Test system message")
		})
	})

	t.Run("monitor notification", func(t *testing.T) {
		assert.NotPanics(t, func() {
			nm.SendMonitorNotification("Test monitor message", "123", "bilibili")
		})
	})

	t.Run("relay notification", func(t *testing.T) {
		assert.NotPanics(t, func() {
			nm.SendRelayNotification("Test relay message", "test-relay", "running")
		})
	})

	t.Run("error notification", func(t *testing.T) {
		assert.NotPanics(t, func() {
			nm.SendErrorNotification("Test error message", "test error")
		})
	})

	t.Run("live status notification", func(t *testing.T) {
		assert.NotPanics(t, func() {
			nm.SendLiveStatusNotification("123", "bilibili", true, nil)
		})
	})

	t.Run("relay status notification", func(t *testing.T) {
		assert.NotPanics(t, func() {
			nm.SendRelayStatusNotification("test-relay", "started", map[string]interface{}{
				"quality": "720p",
			})
		})
	})
}

func TestNotificationManager_Config(t *testing.T) {
	config := Config{
		Telegram: telegram.Config{
			Enabled: false,
		},
		Notifications: NotificationConfig{
			SystemEvents:  false,
			MonitorEvents: true,
			RelayEvents:   false,
			ErrorEvents:   true,
		},
	}

	nm, err := NewNotificationManager(config)
	assert.NoError(t, err)

	// Test that disabled notifications are not sent
	t.Run("disabled system notification", func(t *testing.T) {
		assert.NotPanics(t, func() {
			nm.SendSystemNotification("Should not be sent")
		})
	})

	t.Run("enabled monitor notification", func(t *testing.T) {
		assert.NotPanics(t, func() {
			nm.SendMonitorNotification("Should be sent", "123", "bilibili")
		})
	})

	t.Run("disabled relay notification", func(t *testing.T) {
		assert.NotPanics(t, func() {
			nm.SendRelayNotification("Should not be sent", "test-relay", "running")
		})
	})

	t.Run("enabled error notification", func(t *testing.T) {
		assert.NotPanics(t, func() {
			nm.SendErrorNotification("Should be sent", "test error")
		})
	})
}

func TestNotificationManager_Methods(t *testing.T) {
	config := Config{
		Telegram: telegram.Config{
			Enabled: false,
		},
	}

	nm, err := NewNotificationManager(config)
	assert.NoError(t, err)

	t.Run("get telegram bot", func(t *testing.T) {
		bot := nm.GetTelegramBot()
		assert.Nil(t, bot) // Should be nil when disabled
	})

	t.Run("is enabled", func(t *testing.T) {
		assert.False(t, nm.IsEnabled())
	})

	t.Run("get config", func(t *testing.T) {
		resultConfig := nm.GetConfig()
		assert.Equal(t, config, resultConfig)
	})

	t.Run("start and stop", func(t *testing.T) {
		assert.NotPanics(t, func() {
			err := nm.Start()
			assert.NoError(t, err)
		})

		assert.NotPanics(t, func() {
			nm.Stop()
		})
	})
}