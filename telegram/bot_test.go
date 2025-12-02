package telegram

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewBot(t *testing.T) {
	t.Run("disabled bot", func(t *testing.T) {
		config := Config{
			Enabled: false,
		}
		
		bot, err := NewBot(config)
		assert.Error(t, err)
		assert.Nil(t, bot)
		assert.Contains(t, err.Error(), "disabled")
	})

	t.Run("missing bot token", func(t *testing.T) {
		config := Config{
			Enabled:  true,
			BotToken: "",
		}
		
		bot, err := NewBot(config)
		assert.Error(t, err)
		assert.Nil(t, bot)
		assert.Contains(t, err.Error(), "bot token is required")
	})

	t.Run("invalid bot token", func(t *testing.T) {
		config := Config{
			Enabled:  true,
			BotToken: "invalid_token",
		}
		
		bot, err := NewBot(config)
		assert.Error(t, err)
		assert.Nil(t, bot)
	})
}

func TestBot_FormatNotification(t *testing.T) {
	config := Config{
		Enabled:  true,
		BotToken: "test_token",
		ChatIDs:  []int64{123456789},
	}

	// Create bot without actually initializing API
	bot := &Bot{
		config: config,
	}

	t.Run("basic notification", func(t *testing.T) {
		event := NotificationEvent{
			Type:      "system",
			Message:   "Test message",
			Timestamp: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
		}

		result := bot.formatNotification(event)
		assert.Contains(t, result, "2023-01-01 12:00:00")
		assert.Contains(t, result, "Test message")
	})

	t.Run("notification with data", func(t *testing.T) {
		event := NotificationEvent{
			Type:    "monitor",
			Message: "Room status changed",
			Data: map[string]interface{}{
				"room_id":  "123",
				"platform": "bilibili",
				"status":   "live",
			},
			Timestamp: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
		}

		result := bot.formatNotification(event)
		assert.Contains(t, result, "Room status changed")
		// Note: Data fields are no longer included in formatted output to avoid Markdown parsing errors
		// The message content should contain all necessary information for the user
	})
}

func TestBot_IsAuthorized(t *testing.T) {
	config := Config{
		Enabled:  true,
		BotToken: "test_token",
		AdminIDs: []int64{123456789, 987654321},
	}

	bot := &Bot{
		config: config,
	}

	t.Run("authorized user", func(t *testing.T) {
		assert.True(t, bot.isAuthorized(123456789))
		assert.True(t, bot.isAuthorized(987654321))
	})

	t.Run("unauthorized user", func(t *testing.T) {
		assert.False(t, bot.isAuthorized(111111111))
	})
}

func TestBot_IsCommandEnabled(t *testing.T) {
	t.Run("all commands enabled by default", func(t *testing.T) {
		config := Config{
			Enabled:         true,
			BotToken:        "test_token",
			EnabledCommands: []string{},
		}

		bot := &Bot{
			config: config,
		}

		assert.True(t, bot.isCommandEnabled("status"))
		assert.True(t, bot.isCommandEnabled("help"))
		assert.True(t, bot.isCommandEnabled("stop"))
	})

	t.Run("specific commands enabled", func(t *testing.T) {
		config := Config{
			Enabled:         true,
			BotToken:        "test_token",
			EnabledCommands: []string{"status", "help"},
		}

		bot := &Bot{
			config: config,
		}

		assert.True(t, bot.isCommandEnabled("status"))
		assert.True(t, bot.isCommandEnabled("help"))
		assert.False(t, bot.isCommandEnabled("stop"))
	})
}

func TestBot_NotificationListeners(t *testing.T) {
	config := Config{
		Enabled:  true,
		BotToken: "test_token",
	}

	bot := &Bot{
		config:    config,
		listeners: make(map[string][]NotificationListener),
	}

	t.Run("add and trigger listener", func(t *testing.T) {
		triggered := false
		listener := func(event NotificationEvent) {
			triggered = true
			assert.Equal(t, "test", event.Type)
		}

		bot.AddNotificationListener("test", listener)
		assert.Len(t, bot.listeners["test"], 1)

		// Manually trigger listener
		if listeners, exists := bot.listeners["test"]; exists {
			for _, l := range listeners {
				l(NotificationEvent{Type: "test"})
			}
		}

		assert.True(t, triggered)
	})

	t.Run("remove listener", func(t *testing.T) {
		bot.RemoveNotificationListener("test")
		assert.Len(t, bot.listeners["test"], 0)
	})
}

func TestNotificationCreators(t *testing.T) {
	t.Run("system notification", func(t *testing.T) {
		event := NewSystemNotification("System started")
		assert.Equal(t, "system", event.Type)
		assert.Contains(t, event.Message, "🖥️")
		assert.Contains(t, event.Message, "System started")
		assert.WithinDuration(t, time.Now(), event.Timestamp, time.Second)
	})

	t.Run("monitor notification", func(t *testing.T) {
		event := NewMonitorNotification("Room online", "123", "bilibili")
		assert.Equal(t, "monitor", event.Type)
		assert.Contains(t, event.Message, "👁️")
		assert.Contains(t, event.Message, "Room online")
		assert.Equal(t, "123", event.Data["room_id"])
		assert.Equal(t, "bilibili", event.Data["platform"])
	})

	t.Run("relay notification", func(t *testing.T) {
		event := NewRelayNotification("Relay started", "test-relay", "running")
		assert.Equal(t, "relay", event.Type)
		assert.Contains(t, event.Message, "🔄")
		assert.Contains(t, event.Message, "Relay started")
		assert.Equal(t, "test-relay", event.Data["relay_name"])
		assert.Equal(t, "running", event.Data["status"])
	})

	t.Run("error notification", func(t *testing.T) {
		event := NewErrorNotification("Connection failed", "timeout error")
		assert.Equal(t, "error", event.Type)
		assert.Contains(t, event.Message, "❌")
		assert.Contains(t, event.Message, "Connection failed")
		assert.Equal(t, "timeout error", event.Data["error"])
	})
}