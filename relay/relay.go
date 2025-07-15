package relay

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/nick3/restreamer_monitor_go/monitor"
)

// RelayManager manages multiple stream relays
type RelayManager struct {
	config   monitor.Config
	relays   map[string]*StreamRelay
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
	mu       sync.RWMutex
}

// StreamRelay represents a single stream relay instance
type StreamRelay struct {
	config      monitor.RelayConfig
	source      monitor.StreamSource
	processes   map[string]*exec.Cmd
	ctx         context.Context
	cancel      context.CancelFunc
	mu          sync.RWMutex
	isRunning   bool
	lastError   error
	startTime   time.Time
	restartCount int
}

// NewRelayManager creates a new relay manager
func NewRelayManager(configFile string) (*RelayManager, error) {
	config, err := loadConfig(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	
	manager := &RelayManager{
		config: config,
		relays: make(map[string]*StreamRelay),
		ctx:    ctx,
		cancel: cancel,
	}

	// Initialize relay instances
	for _, relayConfig := range config.Relays {
		if !relayConfig.Enabled {
			continue
		}

		relay, err := NewStreamRelay(relayConfig, ctx)
		if err != nil {
			log.Printf("Failed to create relay %s: %v", relayConfig.Name, err)
			continue
		}

		manager.relays[relayConfig.Name] = relay
	}

	return manager, nil
}

// NewStreamRelay creates a new stream relay instance
func NewStreamRelay(config monitor.RelayConfig, parentCtx context.Context) (*StreamRelay, error) {
	// Create stream source based on platform
	var source monitor.StreamSource
	var err error

	switch config.Source.Platform {
	case "bilibili":
		source, err = monitor.NewBilibiliStreamSource(config.Source.RoomID)
	default:
		return nil, fmt.Errorf("unsupported platform: %s", config.Source.Platform)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create stream source: %w", err)
	}

	ctx, cancel := context.WithCancel(parentCtx)
	
	return &StreamRelay{
		config:    config,
		source:    source,
		processes: make(map[string]*exec.Cmd),
		ctx:       ctx,
		cancel:    cancel,
	}, nil
}

// Run starts the relay manager
func (rm *RelayManager) Run() error {
	if len(rm.relays) == 0 {
		return fmt.Errorf("no relay configurations found")
	}

	log.Printf("Starting relay manager with %d relays", len(rm.relays))

	// Start all relays
	for name, relay := range rm.relays {
		rm.wg.Add(1)
		go func(name string, relay *StreamRelay) {
			defer rm.wg.Done()
			if err := relay.Start(); err != nil {
				log.Printf("Relay %s failed: %v", name, err)
			}
		}(name, relay)
	}

	// Wait for context cancellation
	<-rm.ctx.Done()
	
	// Stop all relays
	rm.Stop()
	
	return nil
}

// Stop stops all relays
func (rm *RelayManager) Stop() {
	log.Println("Stopping relay manager...")
	
	rm.mu.Lock()
	defer rm.mu.Unlock()
	
	for name, relay := range rm.relays {
		log.Printf("Stopping relay %s", name)
		relay.Stop()
	}
	
	rm.cancel()
	rm.wg.Wait()
}

// Start starts the stream relay
func (sr *StreamRelay) Start() error {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	
	if sr.isRunning {
		return nil
	}
	
	log.Printf("Starting relay %s", sr.config.Name)
	sr.startTime = time.Now()
	sr.isRunning = true
	
	// Main relay loop
	for {
		select {
		case <-sr.ctx.Done():
			return nil
		default:
			if err := sr.runRelay(); err != nil {
				log.Printf("Relay %s error: %v", sr.config.Name, err)
				sr.lastError = err
				sr.restartCount++
				
				// Wait before restart
				select {
				case <-sr.ctx.Done():
					return nil
				case <-time.After(5 * time.Second):
					// Continue to retry
				}
			}
		}
	}
}

// runRelay runs the actual relay process
func (sr *StreamRelay) runRelay() error {
	// Check if source is live
	if !sr.source.GetStatus() {
		log.Printf("Relay %s: source is not live, waiting...", sr.config.Name)
		time.Sleep(10 * time.Second)
		return nil
	}
	
	// Get source stream URL
	sourceURL := sr.source.GetPlayURL()
	if sourceURL == "" {
		return fmt.Errorf("failed to get source stream URL")
	}
	
	log.Printf("Relay %s: got source URL: %s", sr.config.Name, sourceURL)
	
	// Start relay processes for each destination
	var wg sync.WaitGroup
	errChan := make(chan error, len(sr.config.Destinations))
	
	for _, dest := range sr.config.Destinations {
		wg.Add(1)
		go func(dest monitor.Destination) {
			defer wg.Done()
			if err := sr.startRelayProcess(sourceURL, dest); err != nil {
				errChan <- fmt.Errorf("destination %s failed: %w", dest.Name, err)
			}
		}(dest)
	}
	
	// Wait for all processes to complete or context cancellation
	go func() {
		wg.Wait()
		close(errChan)
	}()
	
	// Wait for first error or context cancellation
	select {
	case <-sr.ctx.Done():
		sr.stopAllProcesses()
		return nil
	case err := <-errChan:
		sr.stopAllProcesses()
		if err != nil {
			return err
		}
		return nil
	}
}

// startRelayProcess starts a single relay process to a destination
func (sr *StreamRelay) startRelayProcess(sourceURL string, dest monitor.Destination) error {
	// Build FFmpeg command
	args := sr.buildFFmpegArgs(sourceURL, dest)
	
	cmd := exec.CommandContext(sr.ctx, "ffmpeg", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	log.Printf("Starting relay to %s: ffmpeg %s", dest.Name, strings.Join(args, " "))
	
	// Store process for cleanup
	sr.mu.Lock()
	sr.processes[dest.Name] = cmd
	sr.mu.Unlock()
	
	// Start process
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start ffmpeg: %w", err)
	}
	
	// Wait for process to complete
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("ffmpeg process failed: %w", err)
	}
	
	return nil
}

// buildFFmpegArgs builds FFmpeg command arguments
func (sr *StreamRelay) buildFFmpegArgs(sourceURL string, dest monitor.Destination) []string {
	args := []string{
		"-i", sourceURL,
		"-c", "copy", // Copy streams without re-encoding
		"-f", "flv",  // Output format
	}
	
	// Add quality options if specified
	if sr.config.Quality != "" {
		switch sr.config.Quality {
		case "best":
			// Use best quality available
		case "worst":
			args = append(args, "-b:v", "500k")
		case "720p":
			args = append(args, "-s", "1280x720", "-b:v", "2000k")
		case "480p":
			args = append(args, "-s", "854x480", "-b:v", "1000k")
		}
	}
	
	// Add destination-specific options
	for key, value := range dest.Options {
		args = append(args, "-"+key, value)
	}
	
	// Add destination URL
	args = append(args, dest.URL)
	
	return args
}

// stopAllProcesses stops all running processes
func (sr *StreamRelay) stopAllProcesses() {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	
	for name, cmd := range sr.processes {
		if cmd != nil && cmd.Process != nil {
			log.Printf("Stopping relay process for %s", name)
			cmd.Process.Kill()
		}
	}
	
	// Clear processes map
	sr.processes = make(map[string]*exec.Cmd)
}

// Stop stops the stream relay
func (sr *StreamRelay) Stop() {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	
	if !sr.isRunning {
		return
	}
	
	log.Printf("Stopping relay %s", sr.config.Name)
	sr.isRunning = false
	sr.cancel()
	sr.stopAllProcesses()
}

// GetStatus returns the relay status
func (sr *StreamRelay) GetStatus() RelayStatus {
	sr.mu.RLock()
	defer sr.mu.RUnlock()
	
	return RelayStatus{
		Name:         sr.config.Name,
		IsRunning:    sr.isRunning,
		StartTime:    sr.startTime,
		LastError:    sr.lastError,
		RestartCount: sr.restartCount,
		ProcessCount: len(sr.processes),
	}
}

// RelayStatus represents the status of a relay
type RelayStatus struct {
	Name         string
	IsRunning    bool
	StartTime    time.Time
	LastError    error
	RestartCount int
	ProcessCount int
}

// loadConfig loads configuration from JSON file
func loadConfig(configFile string) (monitor.Config, error) {
	var config monitor.Config
	
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