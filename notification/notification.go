package notification

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/nick3/restreamer_monitor_go/monitor"
	"github.com/nick3/restreamer_monitor_go/telegram"
)

// NotificationManager manages all notifications
type NotificationManager struct {
	telegramBot *telegram.Bot
	config      monitor.Config
	ctx         context.Context
	cancel      context.CancelFunc
	mu          sync.RWMutex
}

// NewNotificationManager creates a new notification manager
func NewNotificationManager(config monitor.Config) (*NotificationManager, error) {
	ctx, cancel := context.WithCancel(context.Background())
	
	nm := &NotificationManager{
		config: config,
		ctx:    ctx,
		cancel: cancel,
	}

	// Initialize Telegram bot if enabled
	if config.Telegram.Enabled {
		botConfig := telegram.Config{
			BotToken:        config.Telegram.BotToken,
			ChatIDs:         config.Telegram.ChatIDs,
			AdminIDs:        config.Telegram.AdminIDs,
			Enabled:         config.Telegram.Enabled,
			EnabledCommands: config.Telegram.EnabledCommands,
		}

		bot, err := telegram.NewBot(botConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create Telegram bot: %w", err)
		}

		nm.telegramBot = bot
	}

	return nm, nil
}

// Start starts the notification manager
func (nm *NotificationManager) Start() error {
	if nm.telegramBot != nil {
		if err := nm.telegramBot.Start(); err != nil {
			return fmt.Errorf("failed to start Telegram bot: %w", err)
		}
		log.Println("Telegram bot started successfully")
	}

	return nil
}

// Stop stops the notification manager
func (nm *NotificationManager) Stop() {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	if nm.telegramBot != nil {
		nm.telegramBot.Stop()
	}

	if nm.cancel != nil {
		nm.cancel()
	}
}

// SendSystemNotification sends a system notification
func (nm *NotificationManager) SendSystemNotification(message string) {
	if !nm.config.Telegram.Notifications.SystemEvents {
		return
	}

	if nm.telegramBot != nil {
		event := telegram.NewSystemNotification(message)
		nm.telegramBot.SendNotification(event)
	}
}

// SendMonitorNotification sends a monitor notification
func (nm *NotificationManager) SendMonitorNotification(message string, roomID string, platform string) {
	if !nm.config.Telegram.Notifications.MonitorEvents {
		return
	}

	if nm.telegramBot != nil {
		event := telegram.NewMonitorNotification(message, roomID, platform)
		nm.telegramBot.SendNotification(event)
	}
}

// SendRelayNotification sends a relay notification
func (nm *NotificationManager) SendRelayNotification(message string, relayName string, status string) {
	if !nm.config.Telegram.Notifications.RelayEvents {
		return
	}

	if nm.telegramBot != nil {
		event := telegram.NewRelayNotification(message, relayName, status)
		nm.telegramBot.SendNotification(event)
	}
}

// SendErrorNotification sends an error notification
func (nm *NotificationManager) SendErrorNotification(message string, error string) {
	if !nm.config.Telegram.Notifications.ErrorEvents {
		return
	}

	if nm.telegramBot != nil {
		event := telegram.NewErrorNotification(message, error)
		nm.telegramBot.SendNotification(event)
	}
}

// SendLiveStatusNotification sends a live status change notification
func (nm *NotificationManager) SendLiveStatusNotification(roomID string, platform string, isLive bool, roomInfo interface{}) {
	if !nm.config.Telegram.Notifications.MonitorEvents {
		return
	}

	var message string
	var emoji string

	if isLive {
		emoji = "🟢"
		message = fmt.Sprintf("直播间 %s 开始直播", roomID)
	} else {
		emoji = "🔴"
		message = fmt.Sprintf("直播间 %s 停止直播", roomID)
	}

	if nm.telegramBot != nil {
		event := telegram.NotificationEvent{
			Type:    "monitor",
			Message: emoji + " " + message,
			Data: map[string]interface{}{
				"room_id":  roomID,
				"platform": platform,
				"is_live":  isLive,
				"room_info": roomInfo,
			},
			Timestamp: time.Now(),
		}
		nm.telegramBot.SendNotification(event)
	}
}

// SendRelayStatusNotification sends a relay status change notification
func (nm *NotificationManager) SendRelayStatusNotification(relayName string, status string, details map[string]interface{}) {
	if !nm.config.Telegram.Notifications.RelayEvents {
		return
	}

	var message string
	var emoji string

	switch status {
	case "started":
		emoji = "🟢"
		message = fmt.Sprintf("转播 %s 已启动", relayName)
	case "stopped":
		emoji = "🔴"
		message = fmt.Sprintf("转播 %s 已停止", relayName)
	case "error":
		emoji = "❌"
		message = fmt.Sprintf("转播 %s 发生错误", relayName)
	case "restarted":
		emoji = "🔄"
		message = fmt.Sprintf("转播 %s 已重启", relayName)
	default:
		emoji = "ℹ️"
		message = fmt.Sprintf("转播 %s 状态更新: %s", relayName, status)
	}

	if nm.telegramBot != nil {
		event := telegram.NotificationEvent{
			Type:    "relay",
			Message: emoji + " " + message,
			Data: map[string]interface{}{
				"relay_name": relayName,
				"status":     status,
				"details":    details,
			},
			Timestamp: time.Now(),
		}
		nm.telegramBot.SendNotification(event)
	}
}

// GetTelegramBot returns the Telegram bot instance
func (nm *NotificationManager) GetTelegramBot() *telegram.Bot {
	return nm.telegramBot
}

// IsEnabled returns whether notifications are enabled
func (nm *NotificationManager) IsEnabled() bool {
	return nm.config.Telegram.Enabled
}

// GetConfig returns the notification configuration
func (nm *NotificationManager) GetConfig() monitor.Config {
	return nm.config
}