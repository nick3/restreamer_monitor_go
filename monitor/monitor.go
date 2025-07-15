package monitor

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"
)

// Config represents the application configuration
type Config struct {
	Rooms    []RoomConfig  `json:"rooms"`
	Relays   []RelayConfig `json:"relays,omitempty"`
	Telegram TelegramConfig `json:"telegram,omitempty"`
	Interval string        `json:"interval"`
	Verbose  bool          `json:"verbose"`
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

// Monitor manages multiple stream sources and Telegram notifications
type Monitor struct {
	config            Config
	sources           map[string]StreamSource
	notificationMgr   interface{} // Will be *notification.NotificationManager when imported
	ctx               context.Context
	cancel            context.CancelFunc
	lastStatus        map[string]bool // Track last status for notifications
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
			log.Printf("Unsupported platform: %s", room.Platform)
			continue
		}

		if err != nil {
			log.Printf("Failed to create source for room %s: %v", room.RoomID, err)
			continue
		}

		key := fmt.Sprintf("%s:%s", room.Platform, room.RoomID)
		monitor.sources[key] = source
	}

	return monitor, nil
}

// loadConfig loads configuration from JSON file
func loadConfig(configFile string) (Config, error) {
	var config Config
	
	// Set default values
	config.Interval = "30s"
	config.Verbose = false

	if configFile == "" {
		log.Println("No config file specified, using default configuration")
		return config, nil
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("Config file %s not found, using default configuration", configFile)
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
		log.Printf("Invalid interval %s, using default 30s", m.config.Interval)
		interval = 30 * time.Second
	}

	log.Printf("Starting monitor with %d sources, checking every %v", len(m.sources), interval)

	// Start message listeners
	for key, source := range m.sources {
		if m.config.Verbose {
			log.Printf("Starting message listener for %s", key)
		}
		source.StartMsgListener()
	}

	// Main monitoring loop
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			log.Println("Monitor stopping...")
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
}

// checkAllSources checks the status of all configured sources
func (m *Monitor) checkAllSources() {
	for key, source := range m.sources {
		if m.config.Verbose {
			log.Printf("Checking status for %s", key)
		}

		status := source.GetStatus()
		roomInfo := source.GetRoomInfo()

		// Check if status changed
		lastStatus, exists := m.lastStatus[key]
		if !exists || status != lastStatus {
			// Status changed, send notification
			if m.notificationMgr != nil {
				// Note: This will be properly typed when notification package is imported
				// nm := m.notificationMgr.(*notification.NotificationManager)
				// nm.SendLiveStatusNotification(roomInfo.RoomID, roomInfo.Platform, status, roomInfo)
			}
			m.lastStatus[key] = status
		}

		if m.config.Verbose || status {
			statusStr := "offline"
			if status {
				statusStr = "live"
			}
			log.Printf("Room %s (%s): %s", roomInfo.RoomID, roomInfo.Platform, statusStr)
		}

		if status {
			playURL := source.GetPlayURL()
			if playURL != "" && m.config.Verbose {
				log.Printf("Room %s play URL: %s", roomInfo.RoomID, playURL)
			}
		}
	}
}

// cleanup performs cleanup operations when stopping
func (m *Monitor) cleanup() {
	log.Println("Cleaning up monitor resources...")
	for key, source := range m.sources {
		if m.config.Verbose {
			log.Printf("Closing message listener for %s", key)
		}
		source.CloseMsgListener()
	}
}