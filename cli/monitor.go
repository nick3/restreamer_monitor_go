package cli

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/nick3/restreamer_monitor_go/logger"
	"github.com/nick3/restreamer_monitor_go/monitor"
	"github.com/spf13/cobra"
)

func init() {
	var monitorCmd = &cobra.Command{
		Use:   "monitor",
		Short: "Monitor live room status and send notifications",
		Long:  "Monitor live room status for multiple platforms and send real-time notifications when streams go live or offline.",
		Run: func(cmd *cobra.Command, args []string) {
			interval, _ := cmd.Flags().GetString("interval")
			verbose, _ := cmd.Flags().GetBool("verbose")

			// Create monitor instance first to load config
			m, err := monitor.NewMonitor(cfgFile)
			if err != nil {
				log.Fatalf("Failed to create monitor: %v", err)
			}

			// Initialize logger with config
			loggerCfg := m.GetConfig().Logger
			if verbose {
				loggerCfg.Level = "debug"
			}
			if err := logger.InitLogger(&loggerCfg); err != nil {
				log.Fatalf("Failed to initialize logger: %v", err)
			}
			logger.InitCompatLogger()

			logger.Entry.Info("Monitor service initialized successfully")

			// Update config from command line flags if needed
			_ = interval // Flag is available for future use
			
			// Handle graceful shutdown
			signalChan := make(chan os.Signal, 1)
			signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
			
			// Start monitoring in a goroutine
			go func() {
				if err := m.Run(); err != nil {
					log.Printf("Monitor error: %v", err)
				}
			}()
			
			// Wait for shutdown signal
			<-signalChan
			log.Println("Shutdown signal received")
			m.Stop()
		},
	}

	monitorCmd.Flags().StringP("interval", "i", "30s", "Monitoring check interval (e.g., 30s, 1m)")
	monitorCmd.Flags().BoolP("verbose", "v", false, "Enable verbose logging")

	rootCmd.AddCommand(monitorCmd)
}