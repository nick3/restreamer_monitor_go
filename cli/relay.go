package cli

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/nick3/restreamer_monitor_go/relay"
	"github.com/spf13/cobra"
)

func init() {
	var relayCmd = &cobra.Command{
		Use:   "relay",
		Short: "Relay live streams to multiple destinations",
		Long:  "Relay live streams from source platforms to multiple destination URLs using FFmpeg.",
		Run: func(cmd *cobra.Command, args []string) {
			verbose, _ := cmd.Flags().GetBool("verbose")
			
			// Create relay manager
			manager, err := relay.NewRelayManager(cfgFile)
			if err != nil {
				log.Fatalf("Failed to create relay manager: %v", err)
			}
			
			if verbose {
				log.Printf("Starting relay with config file: %s", cfgFile)
			}
			
			// Handle graceful shutdown
			signalChan := make(chan os.Signal, 1)
			signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
			
			// Start relay manager in a goroutine
			go func() {
				if err := manager.Run(); err != nil {
					log.Printf("Relay manager error: %v", err)
				}
			}()
			
			// Wait for shutdown signal
			<-signalChan
			log.Println("Shutdown signal received")
			manager.Stop()
		},
	}

	relayCmd.Flags().BoolP("verbose", "v", false, "Enable verbose logging")
	relayCmd.Flags().StringP("quality", "q", "", "Stream quality (best, worst, 720p, 480p)")

	rootCmd.AddCommand(relayCmd)
}