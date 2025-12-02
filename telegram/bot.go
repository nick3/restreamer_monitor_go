package telegram

import (
	"context"
	"fmt"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/nick3/restreamer_monitor_go/logger"
	"github.com/sirupsen/logrus"
)

// Bot represents a Telegram bot instance
type Bot struct {
	api       *tgbotapi.BotAPI
	config    Config
	ctx       context.Context
	cancel    context.CancelFunc
	listeners map[string][]NotificationListener
	logger    *logrus.Entry
}

// Config represents Telegram bot configuration
type Config struct {
	BotToken    string   `json:"bot_token"`
	ChatIDs     []int64  `json:"chat_ids"`
	AdminIDs    []int64  `json:"admin_ids"`
	Enabled     bool     `json:"enabled"`
	EnabledCommands []string `json:"enabled_commands,omitempty"`
}

// NotificationListener represents a callback for handling notifications
type NotificationListener func(event NotificationEvent)

// NotificationEvent represents a notification event
type NotificationEvent struct {
	Type      string                 `json:"type"`
	Message   string                 `json:"message"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
}

// NewBot creates a new Telegram bot instance
func NewBot(config Config) (*Bot, error) {
	if !config.Enabled {
		return nil, fmt.Errorf("telegram bot is disabled")
	}

	if config.BotToken == "" {
		return nil, fmt.Errorf("bot token is required")
	}

	api, err := tgbotapi.NewBotAPI(config.BotToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot API: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	bot := &Bot{
		api:       api,
		config:    config,
		ctx:       ctx,
		cancel:    cancel,
		listeners: make(map[string][]NotificationListener),
		logger:    logger.GetLogger(map[string]interface{}{"component": "telegram", "module": "bot"}),
	}

	bot.logger.Infof("Telegram bot authorized on account %s", api.Self.UserName)

	return bot, nil
}

// Start starts the bot
func (b *Bot) Start() error {
	if !b.config.Enabled {
		return fmt.Errorf("telegram bot is disabled")
	}

	b.logger.Info("Starting Telegram bot...")

	// Send startup notification
	b.SendNotification(NotificationEvent{
		Type:      "system",
		Message:   "🚀 Restreamer Monitor started successfully",
		Timestamp: time.Now(),
	})

	// Start command handling
	go b.handleCommands()

	return nil
}

// Stop stops the bot
func (b *Bot) Stop() {
	if b.cancel != nil {
		b.logger.Info("Stopping Telegram bot...")
		
		// Send shutdown notification
		b.SendNotification(NotificationEvent{
			Type:      "system",
			Message:   "🛑 Restreamer Monitor stopping...",
			Timestamp: time.Now(),
		})
		
		b.cancel()
	}
}

// SendNotification sends a notification to all configured chat IDs
func (b *Bot) SendNotification(event NotificationEvent) {
	if !b.config.Enabled {
		return
	}

	message := b.formatNotification(event)

	for _, chatID := range b.config.ChatIDs {
		if len(message) > 200 {
		} else {
		}

		msg := tgbotapi.NewMessage(chatID, message)
		msg.ParseMode = tgbotapi.ModeMarkdown

		if _, err := b.api.Send(msg); err != nil {
			b.logger.WithError(err).WithFields(logrus.Fields{
				"chat_id": chatID,
				"failed_method": "Send(markdown)",
			}).Error("Failed to send notification")

			// Try without Markdown if it fails
			msg.ParseMode = ""
			if _, err := b.api.Send(msg); err != nil {
				b.logger.WithError(err).WithFields(logrus.Fields{
					"chat_id": chatID,
					"failed_method": "Send(plain)",
				}).Error("Failed to send notification without markdown")
			}
		} else {
		}
	}

	// Notify listeners
	if listeners, exists := b.listeners[event.Type]; exists {
		for _, listener := range listeners {
			listener(event)
		}
	}
}

// SendNotificationWithPhoto sends a notification with a photo to all configured chat IDs
func (b *Bot) SendNotificationWithPhoto(event NotificationEvent, photoURL string) {
	if !b.config.Enabled {
		return
	}

	if photoURL == "" {
		// Fallback to text-only notification if no photo URL
		b.SendNotification(event)
		return
	}

	for _, chatID := range b.config.ChatIDs {
		if len(event.Message) > 200 {
		} else {
		}

		// Create photo message with caption
		msg := tgbotapi.NewPhoto(chatID, tgbotapi.FileURL(photoURL))
		msg.Caption = event.Message
		msg.ParseMode = tgbotapi.ModeMarkdown

		if _, err := b.api.Send(msg); err != nil {
			b.logger.WithError(err).WithFields(logrus.Fields{
				"chat_id": chatID,
				"failed_method": "SendPhoto",
			}).Error("Failed to send photo notification")

			// Fallback to text-only notification
			textMsg := tgbotapi.NewMessage(chatID, event.Message)
			textMsg.ParseMode = tgbotapi.ModeMarkdown
			if _, err := b.api.Send(textMsg); err != nil {
				b.logger.WithError(err).WithFields(logrus.Fields{
					"chat_id": chatID,
					"failed_method": "Send(text_fallback)",
				}).Error("Failed to send fallback text notification")
			}
		} else {
		}
	}

	// Notify listeners
	if listeners, exists := b.listeners[event.Type]; exists {
		for _, listener := range listeners {
			listener(event)
		}
	}
}

// SendNotificationToAdmins sends a notification to all configured admin IDs
func (b *Bot) SendNotificationToAdmins(event NotificationEvent) {
	if !b.config.Enabled {
		return
	}

	message := b.formatNotification(event)

	for _, chatID := range b.config.AdminIDs {
		msg := tgbotapi.NewMessage(chatID, message)
		msg.ParseMode = tgbotapi.ModeMarkdown

		if _, err := b.api.Send(msg); err != nil {
			b.logger.WithError(err).WithFields(logrus.Fields{
				"admin_id": chatID,
				"failed_method": "Send(markdown)",
			}).Error("Failed to send notification to admin")

			// Try without Markdown if it fails
			msg.ParseMode = ""
			if _, err := b.api.Send(msg); err != nil {
				b.logger.WithError(err).WithFields(logrus.Fields{
					"admin_id": chatID,
					"failed_method": "Send(plain)",
				}).Error("Failed to send notification without markdown to admin")
			}
		}
	}

	// Notify listeners
	if listeners, exists := b.listeners[event.Type]; exists {
		for _, listener := range listeners {
			listener(event)
		}
	}
}

// SendNotificationWithPhotoToAdmins sends a notification with a photo to all configured admin IDs
func (b *Bot) SendNotificationWithPhotoToAdmins(event NotificationEvent, photoURL string) {
	if !b.config.Enabled {
		return
	}

	if photoURL == "" {
		// Fallback to text-only notification if no photo URL
		b.SendNotificationToAdmins(event)
		return
	}

	for _, chatID := range b.config.AdminIDs {
		// Create photo message with caption
		msg := tgbotapi.NewPhoto(chatID, tgbotapi.FileURL(photoURL))
		msg.Caption = event.Message
		msg.ParseMode = tgbotapi.ModeMarkdown

		if _, err := b.api.Send(msg); err != nil {
			b.logger.WithError(err).WithFields(logrus.Fields{
				"admin_id": chatID,
				"failed_method": "SendPhoto",
			}).Error("Failed to send photo notification to admin")

			// Fallback to text-only notification
			textMsg := tgbotapi.NewMessage(chatID, event.Message)
			textMsg.ParseMode = tgbotapi.ModeMarkdown
			if _, err := b.api.Send(textMsg); err != nil {
				b.logger.WithError(err).WithFields(logrus.Fields{
					"admin_id": chatID,
					"failed_method": "Send(text_fallback)",
				}).Error("Failed to send fallback text notification to admin")
			}
		}
	}

	// Notify listeners
	if listeners, exists := b.listeners[event.Type]; exists {
		for _, listener := range listeners {
			listener(event)
		}
	}
}

// formatNotification formats a notification event into a readable message
func (b *Bot) formatNotification(event NotificationEvent) string {
	var message strings.Builder

	// Add timestamp
	message.WriteString(fmt.Sprintf("*%s*\n\n", event.Timestamp.Format("2006-01-02 15:04:05")))

	// Add message (main content already includes all key information)
	message.WriteString(event.Message)

	return message.String()
}

// handleCommands handles incoming commands
func (b *Bot) handleCommands() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)

	for {
		select {
		case <-b.ctx.Done():
			return
		case update := <-updates:
			if update.Message == nil {
				continue
			}

			// Only process commands, ignore regular messages
			if !update.Message.IsCommand() {
				continue
			}

			// Check if user is authorized
			if !b.isAuthorized(update.Message.From.ID) {
				b.sendMessage(update.Message.Chat.ID, "❌ 您没有权限使用此功能")
				continue
			}

			// Handle command
			b.handleCommand(update.Message)
		}
	}
}

// isAuthorized checks if user is authorized to use the bot
func (b *Bot) isAuthorized(userID int64) bool {
	for _, adminID := range b.config.AdminIDs {
		if adminID == userID {
			return true
		}
	}
	return false
}

// handleCommand handles a specific command
func (b *Bot) handleCommand(message *tgbotapi.Message) {
	if !message.IsCommand() {
		return
	}

	command := message.Command()
	args := strings.Fields(message.CommandArguments())

	// Check if command is enabled
	if !b.isCommandEnabled(command) {
		b.sendMessage(message.Chat.ID, "❌ 此命令已禁用")
		return
	}

	switch command {
	case "start":
		b.handleStartCommand(message)
	case "help":
		b.handleHelpCommand(message)
	case "status":
		b.handleStatusCommand(message)
	case "rooms":
		b.handleRoomsCommand(message)
	case "relays":
		b.handleRelaysCommand(message)
	case "stop":
		b.handleStopCommand(message, args)
	case "restart":
		b.handleRestartCommand(message, args)
	default:
		b.sendMessage(message.Chat.ID, "❌ 未知命令。使用 /help 查看可用命令")
	}
}

// isCommandEnabled checks if a command is enabled
func (b *Bot) isCommandEnabled(command string) bool {
	if len(b.config.EnabledCommands) == 0 {
		return true // All commands enabled by default
	}
	
	for _, enabledCmd := range b.config.EnabledCommands {
		if enabledCmd == command {
			return true
		}
	}
	return false
}

// sendMessage sends a message to a chat
func (b *Bot) sendMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = tgbotapi.ModeMarkdown

	if _, err := b.api.Send(msg); err != nil {
		b.logger.WithError(err).WithField("chat_id", chatID).Error("Failed to send message")
	}
}

// Command handlers
func (b *Bot) handleStartCommand(message *tgbotapi.Message) {
	response := `🤖 *Restreamer Monitor Bot*

欢迎使用直播监控与转播机器人！

*可用命令:*
/help - 显示帮助信息
/status - 查看系统状态
/rooms - 查看监控房间列表
/relays - 查看转播状态
/stop - 停止服务
/restart - 重启服务

使用 /help 获取更多信息。`

	b.sendMessage(message.Chat.ID, response)
}

func (b *Bot) handleHelpCommand(message *tgbotapi.Message) {
	response := `📚 *帮助信息*

*监控命令:*
/status - 查看系统运行状态
/rooms - 查看正在监控的房间列表
/relays - 查看转播服务状态

*控制命令:*
/stop [service] - 停止指定服务 (monitor/relay)
/restart [service] - 重启指定服务

*示例:*
/stop monitor - 停止监控服务
/restart relay - 重启转播服务

*注意:* 只有管理员才能使用控制命令。`

	b.sendMessage(message.Chat.ID, response)
}

// handleStatusCommand handles the status command with real data
func (b *Bot) handleStatusCommand(message *tgbotapi.Message) {
	// Trigger status update via notification system
	b.SendNotification(NotificationEvent{
		Type:    "command",
		Message: "status_requested",
		Data: map[string]interface{}{
			"command": "status",
			"chat_id": message.Chat.ID,
			"user_id": message.From.ID,
		},
		Timestamp: time.Now(),
	})
}

// handleRoomsCommand handles the rooms command with real data
func (b *Bot) handleRoomsCommand(message *tgbotapi.Message) {
	// Trigger rooms status update via notification system
	b.SendNotification(NotificationEvent{
		Type:    "command",
		Message: "rooms_requested",
		Data: map[string]interface{}{
			"command": "rooms",
			"chat_id": message.Chat.ID,
			"user_id": message.From.ID,
		},
		Timestamp: time.Now(),
	})
}

// handleRelaysCommand handles the relays command with real data
func (b *Bot) handleRelaysCommand(message *tgbotapi.Message) {
	// Trigger relays status update via notification system
	b.SendNotification(NotificationEvent{
		Type:    "command",
		Message: "relays_requested",
		Data: map[string]interface{}{
			"command": "relays",
			"chat_id": message.Chat.ID,
			"user_id": message.From.ID,
		},
		Timestamp: time.Now(),
	})
}

// handleStopCommand handles the stop command with real functionality
func (b *Bot) handleStopCommand(message *tgbotapi.Message, args []string) {
	if len(args) == 0 {
		b.sendMessage(message.Chat.ID, "❌ 请指定要停止的服务: monitor 或 relay")
		return
	}

	service := args[0]
	switch service {
	case "monitor":
		b.SendNotification(NotificationEvent{
			Type:    "command",
			Message: "stop_monitor_requested",
			Data: map[string]interface{}{
				"command": "stop_monitor",
				"chat_id": message.Chat.ID,
				"user_id": message.From.ID,
			},
			Timestamp: time.Now(),
		})
	case "relay":
		b.SendNotification(NotificationEvent{
			Type:    "command",
			Message: "stop_relay_requested",
			Data: map[string]interface{}{
				"command": "stop_relay",
				"chat_id": message.Chat.ID,
				"user_id": message.From.ID,
			},
			Timestamp: time.Now(),
		})
	default:
		b.sendMessage(message.Chat.ID, "❌ 未知服务。可用服务: monitor, relay")
	}
}

// handleRestartCommand handles the restart command with real functionality
func (b *Bot) handleRestartCommand(message *tgbotapi.Message, args []string) {
	if len(args) == 0 {
		b.sendMessage(message.Chat.ID, "❌ 请指定要重启的服务: monitor, relay 或 system")
		return
	}

	service := args[0]
	switch service {
	case "monitor":
		b.SendNotification(NotificationEvent{
			Type:    "command",
			Message: "restart_monitor_requested",
			Data: map[string]interface{}{
				"command": "start_monitor",
				"chat_id": message.Chat.ID,
				"user_id": message.From.ID,
			},
			Timestamp: time.Now(),
		})
	case "relay":
		b.SendNotification(NotificationEvent{
			Type:    "command",
			Message: "restart_relay_requested",
			Data: map[string]interface{}{
				"command": "start_relay",
				"chat_id": message.Chat.ID,
				"user_id": message.From.ID,
			},
			Timestamp: time.Now(),
		})
	case "system":
		b.sendMessage(message.Chat.ID, "🔄 正在重启整个系统...")
		b.SendNotification(NotificationEvent{
			Type:    "command",
			Message: "restart_system_requested",
			Data: map[string]interface{}{
				"command": "restart_system",
				"chat_id": message.Chat.ID,
				"user_id": message.From.ID,
			},
			Timestamp: time.Now(),
		})
	default:
		b.sendMessage(message.Chat.ID, "❌ 未知服务。可用服务: monitor, relay, system")
	}
}

// AddNotificationListener adds a notification listener
func (b *Bot) AddNotificationListener(eventType string, listener NotificationListener) {
	if b.listeners[eventType] == nil {
		b.listeners[eventType] = make([]NotificationListener, 0)
	}
	b.listeners[eventType] = append(b.listeners[eventType], listener)
}

// RemoveNotificationListener removes a notification listener
func (b *Bot) RemoveNotificationListener(eventType string) {
	delete(b.listeners, eventType)
}

// GetBotInfo returns bot information
func (b *Bot) GetBotInfo() (string, error) {
	return fmt.Sprintf("Bot: @%s (ID: %d)", b.api.Self.UserName, b.api.Self.ID), nil
}

// Utility functions for creating notifications
func NewSystemNotification(message string) NotificationEvent {
	return NotificationEvent{
		Type:      "system",
		Message:   "🖥️ " + message,
		Timestamp: time.Now(),
	}
}

func NewMonitorNotification(message string, roomID string, platform string) NotificationEvent {
	return NotificationEvent{
		Type:    "monitor",
		Message: "👁️ " + message,
		Data: map[string]interface{}{
			"room_id":  roomID,
			"platform": platform,
		},
		Timestamp: time.Now(),
	}
}

func NewRelayNotification(message string, relayName string, status string) NotificationEvent {
	return NotificationEvent{
		Type:    "relay",
		Message: "🔄 " + message,
		Data: map[string]interface{}{
			"relay_name": relayName,
			"status":     status,
		},
		Timestamp: time.Now(),
	}
}

func NewErrorNotification(message string, error string) NotificationEvent {
	return NotificationEvent{
		Type:    "error",
		Message: "❌ " + message,
		Data: map[string]interface{}{
			"error": error,
		},
		Timestamp: time.Now(),
	}
}