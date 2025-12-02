package notification

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/nick3/restreamer_monitor_go/logger"
	"github.com/nick3/restreamer_monitor_go/models"
	"github.com/nick3/restreamer_monitor_go/telegram"
	"github.com/sirupsen/logrus"
)

// NotificationConfig represents the notification settings
type NotificationConfig struct {
	SystemEvents  bool `json:"system_events"`
	MonitorEvents bool `json:"monitor_events"`
	RelayEvents   bool `json:"relay_events"`
	ErrorEvents   bool `json:"error_events"`
}

// Config represents the notification configuration
type Config struct {
	Telegram      telegram.Config
	Notifications NotificationConfig
}

// NotificationManager manages all notifications
type NotificationManager struct {
	telegramBot *telegram.Bot
	config      Config
	ctx         context.Context
	cancel      context.CancelFunc
	mu          sync.RWMutex
	logger      *logrus.Entry
}

// NewNotificationManager creates a new notification manager
func NewNotificationManager(config Config) (*NotificationManager, error) {
	ctx, cancel := context.WithCancel(context.Background())

	nm := &NotificationManager{
		config: config,
		ctx:    ctx,
		cancel: cancel,
		logger: logger.GetLogger(map[string]interface{}{
			"component": "notification",
			"module":    "manager",
		}),
	}

	// Initialize Telegram bot if enabled
	if config.Telegram.Enabled {
		bot, err := telegram.NewBot(config.Telegram)
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
		nm.logger.Info("Telegram bot started successfully")
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

// SendSystemNotification sends a system notification (to admins only)
func (nm *NotificationManager) SendSystemNotification(message string) {
	if !nm.config.Notifications.SystemEvents {
		return
	}

	if nm.telegramBot != nil {
		event := telegram.NewSystemNotification(message)
		nm.telegramBot.SendNotificationToAdmins(event)
	}
}

// SendMonitorNotification sends a monitor notification
func (nm *NotificationManager) SendMonitorNotification(message string, roomID string, platform string) {
	if !nm.config.Notifications.MonitorEvents {
		return
	}

	if nm.telegramBot != nil {
		event := telegram.NewMonitorNotification(message, roomID, platform)
		nm.telegramBot.SendNotification(event)
	}
}

// SendRelayNotification sends a relay notification (to admins only)
func (nm *NotificationManager) SendRelayNotification(message string, relayName string, status string) {
	if !nm.config.Notifications.RelayEvents {
		return
	}

	if nm.telegramBot != nil {
		event := telegram.NewRelayNotification(message, relayName, status)
		nm.telegramBot.SendNotificationToAdmins(event)
	}
}

// SendErrorNotification sends an error notification (to admins only)
func (nm *NotificationManager) SendErrorNotification(message string, error string) {
	if !nm.config.Notifications.ErrorEvents {
		return
	}

	if nm.telegramBot != nil {
		event := telegram.NewErrorNotification(message, error)
		nm.telegramBot.SendNotificationToAdmins(event)
	}
}

// SendLiveStatusNotification sends a live status change notification
func (nm *NotificationManager) SendLiveStatusNotification(roomID string, platform string, isLive bool, roomInfo interface{}) {
	if !nm.config.Notifications.MonitorEvents {
		return
	}

	if nm.telegramBot == nil {
		return
	}

	// Try to cast roomInfo to models.RoomInfo if possible
	if info, ok := roomInfo.(models.RoomInfo); ok {
		if isLive {
			// Use rich notification with photo for live start
			message, photoURL := telegram.FormatLiveStartNotification(info)
			event := telegram.NotificationEvent{
				Type:    "monitor",
				Message: message,
				Data: map[string]interface{}{
					"room_id":  roomID,
					"platform": platform,
					"is_live":  isLive,
					"room_info": info,
				},
				Timestamp: time.Now(),
			}

			// Send notification with photo
			// Prefer user_cover, fall back to keyframe
			if photoURL == "" && info.Keyframe != "" {
				photoURL = info.Keyframe
			}
			nm.telegramBot.SendNotificationWithPhoto(event, photoURL)
		} else {
			// Use rich notification for live end
			message := telegram.FormatLiveEndNotification(info)
			event := telegram.NotificationEvent{
				Type:    "monitor",
				Message: message,
				Data: map[string]interface{}{
					"room_id":  roomID,
					"platform": platform,
					"is_live":  isLive,
					"room_info": info,
				},
				Timestamp: time.Now(),
			}

			// Send notification with photo for live end as well
			// Use keyframe as the image
			photoURL := info.Keyframe
			if photoURL != "" {
				nm.telegramBot.SendNotificationWithPhoto(event, photoURL)
			} else {
				// Fallback to text-only if no keyframe available
				nm.telegramBot.SendNotification(event)
			}
		}
	} else {
		// Fallback to simple notification if roomInfo is not available
		var message string
		var emoji string

		if isLive {
			emoji = "🟢"
			message = fmt.Sprintf("直播间 %s 开始直播", roomID)
		} else {
			emoji = "🔴"
			message = fmt.Sprintf("直播间 %s 停止直播", roomID)
		}

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
	if !nm.config.Notifications.RelayEvents {
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
		nm.telegramBot.SendNotificationToAdmins(event)
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
func (nm *NotificationManager) GetConfig() Config {
	return nm.config
}