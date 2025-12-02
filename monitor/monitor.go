package monitor

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/nick3/restreamer_monitor_go/logger"
	"github.com/nick3/restreamer_monitor_go/notification"
	"github.com/nick3/restreamer_monitor_go/telegram"
	"github.com/sirupsen/logrus"
)

// Config represents the application configuration
type Config struct {
	Rooms    []RoomConfig  `json:"rooms"`
	Relays   []RelayConfig `json:"relays,omitempty"`
	Telegram TelegramConfig `json:"telegram,omitempty"`
	Interval string        `json:"interval"`
	Verbose  bool          `json:"verbose"`
	Logger   LoggerConfig  `json:"logger"`
}

// TelegramConfig represents Telegram bot configuration
type TelegramConfig struct {
	BotToken        string   `json:"bot_token"`
	ChatIDs         []int64  `json:"chat_ids"`
	AdminIDs        []int64  `json:"admin_ids"`
	Enabled         bool     `json:"enabled"`
	EnabledCommands []string `json:"enabled_commands,omitempty"`
	Notifications   NotificationConfig `json:"notifications,omitempty"`
}

// NotificationConfig represents notification settings
type NotificationConfig struct {
	SystemEvents bool `json:"system_events"`
	MonitorEvents bool `json:"monitor_events"`
	RelayEvents   bool `json:"relay_events"`
	ErrorEvents   bool `json:"error_events"`
}

// ToNotificationConfig converts TelegramConfig to notification.Config
// This method centralizes the configuration conversion logic and avoids
// manual field copying in controllers or other components.
func (tc TelegramConfig) ToNotificationConfig() notification.Config {
	return notification.Config{
		Telegram: telegram.Config{
			BotToken:        tc.BotToken,
			ChatIDs:         tc.ChatIDs,
			AdminIDs:        tc.AdminIDs,
			Enabled:         tc.Enabled,
			EnabledCommands: tc.EnabledCommands,
		},
		Notifications: notification.NotificationConfig{
			SystemEvents:  tc.Notifications.SystemEvents,
			MonitorEvents: tc.Notifications.MonitorEvents,
			RelayEvents:   tc.Notifications.RelayEvents,
			ErrorEvents:   tc.Notifications.ErrorEvents,
		},
	}
}

// RoomConfig represents a single room configuration for monitoring
type RoomConfig struct {
	Platform string `json:"platform"`
	RoomID   string `json:"room_id"`
	Enabled  bool   `json:"enabled"`
}

// RelayConfig represents a relay configuration for streaming
type RelayConfig struct {
	Name         string `json:"name"`
	Source       Source `json:"source"`
	Destinations []Destination `json:"destinations"`
	Enabled      bool   `json:"enabled"`
	Quality      string `json:"quality,omitempty"` // e.g., "best", "worst", "720p"
}

// Source represents the source stream configuration
type Source struct {
	Platform string `json:"platform"`
	RoomID   string `json:"room_id"`
}

// Destination represents the destination stream configuration
type Destination struct {
	Name     string            `json:"name"`
	URL      string            `json:"url"`
	Protocol string            `json:"protocol"` // rtmp, rtmps, etc.
	Options  map[string]string `json:"options,omitempty"`
}

// LoggerConfig is a type alias for logger.Config
type LoggerConfig = logger.Config

// Monitor manages multiple stream sources and Telegram notifications
type Monitor struct {
	config            Config
	sources           map[string]StreamSource
	notificationMgr   *notification.NotificationManager
	ctx               context.Context
	cancel            context.CancelFunc
	lastStatus        map[string]bool // Track last status for notifications
	logger            *logrus.Entry
}

// NewMonitor creates a new monitor instance
func NewMonitor(configFile string) (*Monitor, error) {
	config, err := loadConfig(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	monitor := &Monitor{
		config:     config,
		sources:    make(map[string]StreamSource),
		ctx:        ctx,
		cancel:     cancel,
		lastStatus: make(map[string]bool),
		logger:     logger.GetLogger(map[string]interface{}{"component": "monitor", "module": "main"}),
	}

	// Initialize stream sources
	for _, room := range config.Rooms {
		if !room.Enabled {
			continue
		}

		var source StreamSource
		var err error

		switch room.Platform {
		case "bilibili":
			source, err = NewBilibiliStreamSource(room.RoomID)
		default:
			monitor.logger.Warnf("Unsupported platform: %s", room.Platform)
			continue
		}

		if err != nil {
			monitor.logger.WithError(err).Errorf("Failed to create source for room %s", room.RoomID)
			continue
		}

		key := fmt.Sprintf("%s:%s", room.Platform, room.RoomID)
		monitor.sources[key] = source
	}

	// Initialize notification manager
	notificationConfig := notification.Config{
		Telegram: telegram.Config{
			BotToken:        config.Telegram.BotToken,
			ChatIDs:         config.Telegram.ChatIDs,
			AdminIDs:        config.Telegram.AdminIDs,
			Enabled:         config.Telegram.Enabled,
			EnabledCommands: config.Telegram.EnabledCommands,
		},
		Notifications: notification.NotificationConfig{
			SystemEvents:  config.Telegram.Notifications.SystemEvents,
			MonitorEvents: config.Telegram.Notifications.MonitorEvents,
			RelayEvents:   config.Telegram.Notifications.RelayEvents,
			ErrorEvents:   config.Telegram.Notifications.ErrorEvents,
		},
	}
	notificationMgr, err := notification.NewNotificationManager(notificationConfig)
	if err != nil {
		monitor.logger.WithError(err).Warn("Failed to create notification manager, continuing without notifications")
		// Continue without notifications
	} else {
		monitor.notificationMgr = notificationMgr
	}

	return monitor, nil
}

// loadConfig loads configuration from JSON file
func loadConfig(configFile string) (Config, error) {
	var config Config

	// Set default values
	config.Interval = "30s"
	config.Verbose = false
	config.Logger = logger.DefaultConfig()

	if configFile == "" {
		logger.DefaultWrapper.Println("No config file specified, using default configuration")
		return config, nil
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		if os.IsNotExist(err) {
			logger.DefaultWrapper.Printf("Config file %s not found, using default configuration", configFile)
			return config, nil
		}
		return config, err
	}

	if err := json.Unmarshal(data, &config); err != nil {
		return config, fmt.Errorf("failed to parse config file: %w", err)
	}

	return config, nil
}

// Run starts the monitoring process
func (m *Monitor) Run() error {
	if len(m.sources) == 0 {
		return fmt.Errorf("no valid stream sources configured")
	}

	interval, err := time.ParseDuration(m.config.Interval)
	if err != nil {
		m.logger.Warnf("Invalid interval %s, using default 30s", m.config.Interval)
		interval = 30 * time.Second
	}

	m.logger.Infof("Starting monitor with %d sources, checking every %v", len(m.sources), interval)

	// Start notification manager if available
	if m.notificationMgr != nil {
		if err := m.notificationMgr.Start(); err != nil {
			m.logger.WithError(err).Warn("Failed to start notification manager")
		}
	}

	// Start message listeners
	for key, source := range m.sources {
		if m.config.Verbose {
			m.logger.Debugf("Starting message listener for %s", key)
		}
		source.StartMsgListener()
	}

	// Main monitoring loop
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			m.logger.Info("Monitor stopping...")
			m.cleanup()
			return nil
		case <-ticker.C:
			m.checkAllSources()
		}
	}
}

// Stop stops the monitoring process
func (m *Monitor) Stop() {
	if m.cancel != nil {
		m.cancel()
	}

	// Stop notification manager if available
	if m.notificationMgr != nil {
		m.notificationMgr.Stop()
	}
}

// GetConfig returns the monitor configuration
func (m *Monitor) GetConfig() Config {
	return m.config
}

// checkAllSources checks the status of all configured sources
func (m *Monitor) checkAllSources() {
	for key, source := range m.sources {
		// Check if context is cancelled before processing each source
		select {
		case <-m.ctx.Done():
			return
		default:
		}

		if m.config.Verbose {
			m.logger.Debugf("Checking status for %s", key)
		}

		status := source.GetStatus()
		roomInfo := source.GetRoomInfo()

		// Check if status changed
		lastStatus, exists := m.lastStatus[key]
		if !exists || status != lastStatus {
			// Status changed, record end time if going from live to offline
			if exists && lastStatus && !status {
				// From live to offline, record the end time
				roomInfo.EndTime = time.Now()
			}

			// Status changed, send notification
			if m.notificationMgr != nil {
				m.notificationMgr.SendLiveStatusNotification(roomInfo.RoomID, roomInfo.Platform, status, roomInfo)
			}
			m.lastStatus[key] = status
		}

		if m.config.Verbose || status {
			statusStr := "offline"
			if status {
				statusStr = "live"
			}
			m.logger.WithFields(logrus.Fields{
				"room_id":  roomInfo.RoomID,
				"platform": roomInfo.Platform,
				"status":   statusStr,
			}).Info("Room status update")
		}

		if status {
			playURL := source.GetPlayURL()
			if playURL != "" && m.config.Verbose {
				m.logger.WithFields(logrus.Fields{
					"room_id": roomInfo.RoomID,
					"play_url": playURL,
				}).Debug("Room play URL retrieved")
			}
		}
	}
}

// cleanup performs cleanup operations when stopping
func (m *Monitor) cleanup() {
	m.logger.Info("Cleaning up monitor resources...")
	for key, source := range m.sources {
		if m.config.Verbose {
			m.logger.Debugf("Closing message listener for %s", key)
		}
		source.CloseMsgListener()
	}
}