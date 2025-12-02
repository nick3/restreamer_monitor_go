package control

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/nick3/restreamer_monitor_go/logger"
	"github.com/nick3/restreamer_monitor_go/monitor"
	"github.com/nick3/restreamer_monitor_go/notification"
	"github.com/nick3/restreamer_monitor_go/relay"
	"github.com/nick3/restreamer_monitor_go/telegram"
	"github.com/sirupsen/logrus"
)

// ServiceController manages all services and provides Telegram bot control
type ServiceController struct {
	config          monitor.Config
	monitorService  *monitor.Monitor
	relayManager    *relay.RelayManager
	notificationMgr *notification.NotificationManager
	telegramBot     *telegram.Bot
	ctx             context.Context
	cancel          context.CancelFunc
	mu              sync.RWMutex
	startTime       time.Time
	status          ServiceStatus
	logger          *logrus.Entry
}

// ServiceStatus represents the status of all services
type ServiceStatus struct {
	Monitor ServiceInfo `json:"monitor"`
	Relay   ServiceInfo `json:"relay"`
	Bot     ServiceInfo `json:"bot"`
	System  SystemInfo  `json:"system"`
}

// ServiceInfo represents individual service information
type ServiceInfo struct {
	Running   bool      `json:"running"`
	StartTime time.Time `json:"start_time"`
	Uptime    string    `json:"uptime"`
	Error     string    `json:"error,omitempty"`
}

// SystemInfo represents system information
type SystemInfo struct {
	CPUUsage    float64 `json:"cpu_usage"`
	MemoryUsage float64 `json:"memory_usage"`
	Uptime      string  `json:"uptime"`
	GoRoutines  int     `json:"goroutines"`
}

// NewServiceController creates a new service controller
func NewServiceController(configFile string) (*ServiceController, error) {
	config, err := loadConfig(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	sc := &ServiceController{
		config:    config,
		ctx:       ctx,
		cancel:    cancel,
		startTime: time.Now(),
		logger: logger.GetLogger(map[string]interface{}{
			"component": "control",
			"module":    "controller",
		}),
	}

	// Initialize notification manager
	if config.Telegram.Enabled {
		// Convert monitor.Config to notification.Config
		nmConfig := notification.Config{
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
		sc.notificationMgr, err = notification.NewNotificationManager(nmConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create notification manager: %w", err)
		}
		sc.telegramBot = sc.notificationMgr.GetTelegramBot()
	}

	return sc, nil
}

// Start starts all enabled services
func (sc *ServiceController) Start() error {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	sc.logger.Info("Starting service controller...")

	// Start notification manager first
	if sc.notificationMgr != nil {
		if err := sc.notificationMgr.Start(); err != nil {
			return fmt.Errorf("failed to start notification manager: %w", err)
		}
		
		// Setup bot command handlers
		sc.setupBotHandlers()
		
		sc.status.Bot.Running = true
		sc.status.Bot.StartTime = time.Now()
	}

	// Send startup notification
	if sc.notificationMgr != nil {
		sc.notificationMgr.SendSystemNotification("系统启动中...")
	}

	// Initialize and start monitor service if there are rooms configured
	if len(sc.config.Rooms) > 0 {
		var err error
		sc.monitorService, err = monitor.NewMonitor("")
		if err != nil {
			sc.logger.WithError(err).Error("Failed to create monitor service")
			sc.status.Monitor.Error = err.Error()
		} else {
			go func() {
				sc.status.Monitor.Running = true
				sc.status.Monitor.StartTime = time.Now()
				
				if err := sc.monitorService.Run(); err != nil {
					sc.logger.WithError(err).Error("Monitor service error")
					sc.status.Monitor.Error = err.Error()
					sc.status.Monitor.Running = false
					
					if sc.notificationMgr != nil {
						sc.notificationMgr.SendErrorNotification("监控服务错误", err.Error())
					}
				}
			}()
		}
	}

	// Initialize and start relay manager if there are relays configured
	if len(sc.config.Relays) > 0 {
		var err error
		sc.relayManager, err = relay.NewRelayManager("")
		if err != nil {
			sc.logger.WithError(err).Error("Failed to create relay manager")
			sc.status.Relay.Error = err.Error()
		} else {
			go func() {
				sc.status.Relay.Running = true
				sc.status.Relay.StartTime = time.Now()

				if err := sc.relayManager.Run(); err != nil {
					sc.logger.WithError(err).Error("Relay manager error")
					sc.status.Relay.Error = err.Error()
					sc.status.Relay.Running = false

					if sc.notificationMgr != nil {
						sc.notificationMgr.SendErrorNotification("转播服务错误", err.Error())
					}
				}
			}()
		}
	}

	// Start status update routine
	go sc.updateSystemStatus()

	// Send startup complete notification
	if sc.notificationMgr != nil {
		sc.notificationMgr.SendSystemNotification("🚀 系统启动完成")
	}

	sc.logger.Info("Service controller started successfully")
	return nil
}

// Stop stops all services
func (sc *ServiceController) Stop() {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	sc.logger.Info("Stopping service controller...")

	// Send shutdown notification
	if sc.notificationMgr != nil {
		sc.notificationMgr.SendSystemNotification("🛑 系统关闭中...")
	}

	// Stop services
	if sc.monitorService != nil {
		sc.monitorService.Stop()
		sc.status.Monitor.Running = false
	}

	if sc.relayManager != nil {
		sc.relayManager.Stop()
		sc.status.Relay.Running = false
	}

	if sc.notificationMgr != nil {
		sc.notificationMgr.Stop()
		sc.status.Bot.Running = false
	}

	sc.cancel()
	sc.logger.Info("Service controller stopped")
}

// setupBotHandlers sets up Telegram bot command handlers
func (sc *ServiceController) setupBotHandlers() {
	if sc.telegramBot == nil {
		return
	}

	// Add custom command handlers for service control
	sc.telegramBot.AddNotificationListener("command", func(event telegram.NotificationEvent) {
		if data, ok := event.Data["command"].(string); ok {
			sc.handleBotCommand(data, event.Data)
		}
	})
}

// handleBotCommand handles bot commands for service control
func (sc *ServiceController) handleBotCommand(command string, data map[string]interface{}) {
	switch command {
	case "status":
		sc.sendStatusUpdate()
	case "stop_monitor":
		sc.stopMonitorService()
	case "start_monitor":
		sc.startMonitorService()
	case "stop_relay":
		sc.stopRelayService()
	case "start_relay":
		sc.startRelayService()
	case "restart_system":
		sc.restartSystem()
	}
}

// sendStatusUpdate sends current system status to Telegram
func (sc *ServiceController) sendStatusUpdate() {
	if sc.notificationMgr == nil {
		return
	}

	status := sc.GetStatus()
	
	message := fmt.Sprintf(`📊 *系统状态报告*

🖥️ *系统信息*
• 运行时间: %s
• CPU使用率: %.1f%%
• 内存使用: %.1f MB
• Go协程数: %d

📺 *监控服务*
• 状态: %s
• 运行时间: %s
%s

🔄 *转播服务*
• 状态: %s
• 运行时间: %s
%s

🤖 *Telegram Bot*
• 状态: %s
• 运行时间: %s`,
		status.System.Uptime,
		status.System.CPUUsage,
		status.System.MemoryUsage,
		status.System.GoRoutines,
		sc.getStatusEmoji(status.Monitor.Running), status.Monitor.Uptime,
		sc.getErrorText(status.Monitor.Error),
		sc.getStatusEmoji(status.Relay.Running), status.Relay.Uptime,
		sc.getErrorText(status.Relay.Error),
		sc.getStatusEmoji(status.Bot.Running), status.Bot.Uptime)

	sc.notificationMgr.SendSystemNotification(message)
}

// getStatusEmoji returns appropriate emoji for service status
func (sc *ServiceController) getStatusEmoji(running bool) string {
	if running {
		return "🟢 运行中"
	}
	return "🔴 已停止"
}

// getErrorText returns error text if present
func (sc *ServiceController) getErrorText(error string) string {
	if error != "" {
		return fmt.Sprintf("• 错误: %s", error)
	}
	return ""
}

// stopMonitorService stops the monitor service
func (sc *ServiceController) stopMonitorService() {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	if sc.monitorService != nil {
		sc.monitorService.Stop()
		sc.status.Monitor.Running = false
		
		if sc.notificationMgr != nil {
			sc.notificationMgr.SendSystemNotification("🛑 监控服务已停止")
		}
	}
}

// startMonitorService starts the monitor service
func (sc *ServiceController) startMonitorService() {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	if sc.monitorService == nil && len(sc.config.Rooms) > 0 {
		var err error
		sc.monitorService, err = monitor.NewMonitor("")
		if err != nil {
			if sc.notificationMgr != nil {
				sc.notificationMgr.SendErrorNotification("启动监控服务失败", err.Error())
			}
			return
		}
	}

	if sc.monitorService != nil && !sc.status.Monitor.Running {
		go func() {
			sc.status.Monitor.Running = true
			sc.status.Monitor.StartTime = time.Now()
			sc.status.Monitor.Error = ""
			
			if sc.notificationMgr != nil {
				sc.notificationMgr.SendSystemNotification("🟢 监控服务已启动")
			}
			
			if err := sc.monitorService.Run(); err != nil {
				sc.status.Monitor.Error = err.Error()
				sc.status.Monitor.Running = false
				
				if sc.notificationMgr != nil {
					sc.notificationMgr.SendErrorNotification("监控服务错误", err.Error())
				}
			}
		}()
	}
}

// stopRelayService stops the relay service
func (sc *ServiceController) stopRelayService() {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	if sc.relayManager != nil {
		sc.relayManager.Stop()
		sc.status.Relay.Running = false
		
		if sc.notificationMgr != nil {
			sc.notificationMgr.SendSystemNotification("🛑 转播服务已停止")
		}
	}
}

// startRelayService starts the relay service
func (sc *ServiceController) startRelayService() {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	if sc.relayManager == nil && len(sc.config.Relays) > 0 {
		var err error
		sc.relayManager, err = relay.NewRelayManager("")
		if err != nil {
			if sc.notificationMgr != nil {
				sc.notificationMgr.SendErrorNotification("启动转播服务失败", err.Error())
			}
			return
		}
	}

	if sc.relayManager != nil && !sc.status.Relay.Running {
		go func() {
			sc.status.Relay.Running = true
			sc.status.Relay.StartTime = time.Now()
			sc.status.Relay.Error = ""
			
			if sc.notificationMgr != nil {
				sc.notificationMgr.SendSystemNotification("🟢 转播服务已启动")
			}
			
			if err := sc.relayManager.Run(); err != nil {
				sc.status.Relay.Error = err.Error()
				sc.status.Relay.Running = false
				
				if sc.notificationMgr != nil {
					sc.notificationMgr.SendErrorNotification("转播服务错误", err.Error())
				}
			}
		}()
	}
}

// restartSystem restarts the entire system
func (sc *ServiceController) restartSystem() {
	if sc.notificationMgr != nil {
		sc.notificationMgr.SendSystemNotification("🔄 系统重启中...")
	}

	// Stop all services
	sc.Stop()

	// Wait a moment
	time.Sleep(2 * time.Second)

	// Restart all services
	if err := sc.Start(); err != nil {
		sc.logger.WithError(err).Error("Failed to restart system")
		if sc.notificationMgr != nil {
			sc.notificationMgr.SendErrorNotification("系统重启失败", err.Error())
		}
	}
}

// updateSystemStatus updates system status periodically
func (sc *ServiceController) updateSystemStatus() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-sc.ctx.Done():
			return
		case <-ticker.C:
			sc.updateStatus()
		}
	}
}

// updateStatus updates the current status
func (sc *ServiceController) updateStatus() {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	now := time.Now()

	// Update service uptimes
	if sc.status.Monitor.Running {
		sc.status.Monitor.Uptime = formatDuration(now.Sub(sc.status.Monitor.StartTime))
	}
	if sc.status.Relay.Running {
		sc.status.Relay.Uptime = formatDuration(now.Sub(sc.status.Relay.StartTime))
	}
	if sc.status.Bot.Running {
		sc.status.Bot.Uptime = formatDuration(now.Sub(sc.status.Bot.StartTime))
	}

	// Update system info
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	sc.status.System.Uptime = formatDuration(now.Sub(sc.startTime))
	sc.status.System.MemoryUsage = float64(m.Alloc) / 1024 / 1024 // MB
	sc.status.System.GoRoutines = runtime.NumGoroutine()
	// Note: CPU usage would require additional implementation
	sc.status.System.CPUUsage = 0.0
}

// GetStatus returns current service status
func (sc *ServiceController) GetStatus() ServiceStatus {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.status
}

// formatDuration formats duration in a human-readable format
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.0f秒", d.Seconds())
	} else if d < time.Hour {
		return fmt.Sprintf("%.0f分钟", d.Minutes())
	} else if d < 24*time.Hour {
		hours := int(d.Hours())
		minutes := int(d.Minutes()) % 60
		return fmt.Sprintf("%d小时%d分钟", hours, minutes)
	} else {
		days := int(d.Hours()) / 24
		hours := int(d.Hours()) % 24
		return fmt.Sprintf("%d天%d小时", days, hours)
	}
}

// loadConfig loads configuration from file (placeholder)
func loadConfig(configFile string) (monitor.Config, error) {
	// This should use the same loadConfig function from monitor package
	// For now, return empty config
	return monitor.Config{}, nil
}