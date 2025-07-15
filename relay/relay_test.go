package relay

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"

	"github.com/nick3/restreamer_monitor_go/monitor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
	t.Run("valid config with relays", func(t *testing.T) {
		// Create temporary config file
		configData := monitor.Config{
			Relays: []monitor.RelayConfig{
				{
					Name: "test-relay",
					Source: monitor.Source{
						Platform: "bilibili",
						RoomID:   "76",
					},
					Destinations: []monitor.Destination{
						{
							Name:     "youtube",
							URL:      "rtmp://a.rtmp.youtube.com/live2/TEST_KEY",
							Protocol: "rtmp",
							Options: map[string]string{
								"bufsize": "3000k",
								"maxrate": "3000k",
							},
						},
					},
					Enabled: true,
					Quality: "720p",
				},
			},
			Interval: "30s",
			Verbose:  true,
		}

		data, err := json.Marshal(configData)
		require.NoError(t, err)

		tmpFile, err := ioutil.TempFile("", "test-relay-config-*.json")
		require.NoError(t, err)
		defer os.Remove(tmpFile.Name())

		_, err = tmpFile.Write(data)
		require.NoError(t, err)
		tmpFile.Close()

		// Load config
		config, err := loadConfig(tmpFile.Name())
		assert.NoError(t, err)
		assert.Equal(t, "30s", config.Interval)
		assert.True(t, config.Verbose)
		assert.Len(t, config.Relays, 1)
		assert.Equal(t, "test-relay", config.Relays[0].Name)
		assert.Equal(t, "bilibili", config.Relays[0].Source.Platform)
		assert.Equal(t, "76", config.Relays[0].Source.RoomID)
		assert.Len(t, config.Relays[0].Destinations, 1)
		assert.Equal(t, "youtube", config.Relays[0].Destinations[0].Name)
		assert.Equal(t, "720p", config.Relays[0].Quality)
	})

	t.Run("nonexistent file", func(t *testing.T) {
		config, err := loadConfig("nonexistent.json")
		assert.NoError(t, err)
		assert.Equal(t, "30s", config.Interval)
		assert.False(t, config.Verbose)
		assert.Len(t, config.Relays, 0)
	})
}

func TestNewStreamRelay(t *testing.T) {
	t.Run("valid bilibili relay", func(t *testing.T) {
		config := monitor.RelayConfig{
			Name: "test-relay",
			Source: monitor.Source{
				Platform: "bilibili",
				RoomID:   "76",
			},
			Destinations: []monitor.Destination{
				{
					Name:     "test-dest",
					URL:      "rtmp://test.example.com/live/test",
					Protocol: "rtmp",
				},
			},
			Enabled: true,
			Quality: "720p",
		}

		ctx := context.Background()
		relay, err := NewStreamRelay(config, ctx)
		assert.NoError(t, err)
		assert.NotNil(t, relay)
		assert.Equal(t, "test-relay", relay.config.Name)
		assert.NotNil(t, relay.source)
		assert.Equal(t, "720p", relay.config.Quality)
	})

	t.Run("unsupported platform", func(t *testing.T) {
		config := monitor.RelayConfig{
			Name: "test-relay",
			Source: monitor.Source{
				Platform: "unsupported",
				RoomID:   "123",
			},
			Destinations: []monitor.Destination{
				{
					Name:     "test-dest",
					URL:      "rtmp://test.example.com/live/test",
					Protocol: "rtmp",
				},
			},
			Enabled: true,
		}

		ctx := context.Background()
		relay, err := NewStreamRelay(config, ctx)
		assert.Error(t, err)
		assert.Nil(t, relay)
		assert.Contains(t, err.Error(), "unsupported platform")
	})

	t.Run("invalid room ID", func(t *testing.T) {
		config := monitor.RelayConfig{
			Name: "test-relay",
			Source: monitor.Source{
				Platform: "bilibili",
				RoomID:   "invalid",
			},
			Destinations: []monitor.Destination{
				{
					Name:     "test-dest",
					URL:      "rtmp://test.example.com/live/test",
					Protocol: "rtmp",
				},
			},
			Enabled: true,
		}

		ctx := context.Background()
		relay, err := NewStreamRelay(config, ctx)
		assert.Error(t, err)
		assert.Nil(t, relay)
	})
}

func TestNewRelayManager(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		// Create temporary config file
		configData := monitor.Config{
			Relays: []monitor.RelayConfig{
				{
					Name: "test-relay",
					Source: monitor.Source{
						Platform: "bilibili",
						RoomID:   "76",
					},
					Destinations: []monitor.Destination{
						{
							Name:     "test-dest",
							URL:      "rtmp://test.example.com/live/test",
							Protocol: "rtmp",
						},
					},
					Enabled: true,
				},
			},
			Interval: "30s",
			Verbose:  false,
		}

		data, err := json.Marshal(configData)
		require.NoError(t, err)

		tmpFile, err := ioutil.TempFile("", "test-relay-config-*.json")
		require.NoError(t, err)
		defer os.Remove(tmpFile.Name())

		_, err = tmpFile.Write(data)
		require.NoError(t, err)
		tmpFile.Close()

		manager, err := NewRelayManager(tmpFile.Name())
		assert.NoError(t, err)
		assert.NotNil(t, manager)
		assert.Len(t, manager.relays, 1)
		assert.Contains(t, manager.relays, "test-relay")
	})

	t.Run("no enabled relays", func(t *testing.T) {
		configData := monitor.Config{
			Relays: []monitor.RelayConfig{
				{
					Name: "test-relay",
					Source: monitor.Source{
						Platform: "bilibili",
						RoomID:   "76",
					},
					Destinations: []monitor.Destination{
						{
							Name:     "test-dest",
							URL:      "rtmp://test.example.com/live/test",
							Protocol: "rtmp",
						},
					},
					Enabled: false, // Disabled
				},
			},
			Interval: "30s",
			Verbose:  false,
		}

		data, err := json.Marshal(configData)
		require.NoError(t, err)

		tmpFile, err := ioutil.TempFile("", "test-relay-config-*.json")
		require.NoError(t, err)
		defer os.Remove(tmpFile.Name())

		_, err = tmpFile.Write(data)
		require.NoError(t, err)
		tmpFile.Close()

		manager, err := NewRelayManager(tmpFile.Name())
		assert.NoError(t, err)
		assert.NotNil(t, manager)
		assert.Len(t, manager.relays, 0)
	})
}

func TestStreamRelay_BuildFFmpegArgs(t *testing.T) {
	config := monitor.RelayConfig{
		Name: "test-relay",
		Source: monitor.Source{
			Platform: "bilibili",
			RoomID:   "76",
		},
		Quality: "720p",
	}

	ctx := context.Background()
	relay, err := NewStreamRelay(config, ctx)
	require.NoError(t, err)

	t.Run("basic args", func(t *testing.T) {
		dest := monitor.Destination{
			Name:     "test-dest",
			URL:      "rtmp://test.example.com/live/test",
			Protocol: "rtmp",
		}

		args := relay.buildFFmpegArgs("http://test.m3u8", dest)
		
		assert.Contains(t, args, "-i")
		assert.Contains(t, args, "http://test.m3u8")
		assert.Contains(t, args, "-c")
		assert.Contains(t, args, "copy")
		assert.Contains(t, args, "-f")
		assert.Contains(t, args, "flv")
		assert.Contains(t, args, "rtmp://test.example.com/live/test")
	})

	t.Run("with quality settings", func(t *testing.T) {
		dest := monitor.Destination{
			Name:     "test-dest",
			URL:      "rtmp://test.example.com/live/test",
			Protocol: "rtmp",
		}

		args := relay.buildFFmpegArgs("http://test.m3u8", dest)
		
		// Should contain 720p settings
		assert.Contains(t, args, "-s")
		assert.Contains(t, args, "1280x720")
		assert.Contains(t, args, "-b:v")
		assert.Contains(t, args, "2000k")
	})

	t.Run("with custom options", func(t *testing.T) {
		dest := monitor.Destination{
			Name:     "test-dest",
			URL:      "rtmp://test.example.com/live/test",
			Protocol: "rtmp",
			Options: map[string]string{
				"bufsize": "3000k",
				"maxrate": "3000k",
			},
		}

		args := relay.buildFFmpegArgs("http://test.m3u8", dest)
		
		assert.Contains(t, args, "-bufsize")
		assert.Contains(t, args, "3000k")
		assert.Contains(t, args, "-maxrate")
		assert.Contains(t, args, "3000k")
	})
}

func TestStreamRelay_Status(t *testing.T) {
	config := monitor.RelayConfig{
		Name: "test-relay",
		Source: monitor.Source{
			Platform: "bilibili",
			RoomID:   "76",
		},
		Destinations: []monitor.Destination{
			{
				Name:     "test-dest",
				URL:      "rtmp://test.example.com/live/test",
				Protocol: "rtmp",
			},
		},
		Enabled: true,
	}

	ctx := context.Background()
	relay, err := NewStreamRelay(config, ctx)
	require.NoError(t, err)

	// Test initial status
	status := relay.GetStatus()
	assert.Equal(t, "test-relay", status.Name)
	assert.False(t, status.IsRunning)
	assert.Equal(t, 0, status.RestartCount)
	assert.Equal(t, 0, status.ProcessCount)

	// Test stop when not running
	relay.Stop()
	status = relay.GetStatus()
	assert.False(t, status.IsRunning)
}

func TestRelayManager_RunWithNoRelays(t *testing.T) {
	configData := monitor.Config{
		Relays:   []monitor.RelayConfig{}, // No relays
		Interval: "30s",
		Verbose:  false,
	}

	data, err := json.Marshal(configData)
	require.NoError(t, err)

	tmpFile, err := ioutil.TempFile("", "test-relay-config-*.json")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.Write(data)
	require.NoError(t, err)
	tmpFile.Close()

	manager, err := NewRelayManager(tmpFile.Name())
	require.NoError(t, err)

	err = manager.Run()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no relay configurations found")
}