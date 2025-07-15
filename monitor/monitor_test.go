package monitor

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
	t.Run("nonexistent file", func(t *testing.T) {
		config, err := loadConfig("nonexistent.json")
		assert.NoError(t, err)
		assert.Equal(t, "30s", config.Interval)
		assert.False(t, config.Verbose)
	})

	t.Run("valid config file", func(t *testing.T) {
		// Create temporary config file
		configData := Config{
			Rooms: []RoomConfig{
				{Platform: "bilibili", RoomID: "123", Enabled: true},
				{Platform: "bilibili", RoomID: "456", Enabled: false},
			},
			Interval: "60s",
			Verbose:  true,
		}

		data, err := json.Marshal(configData)
		require.NoError(t, err)

		tmpFile, err := ioutil.TempFile("", "test-config-*.json")
		require.NoError(t, err)
		defer os.Remove(tmpFile.Name())

		_, err = tmpFile.Write(data)
		require.NoError(t, err)
		tmpFile.Close()

		// Load config
		config, err := loadConfig(tmpFile.Name())
		assert.NoError(t, err)
		assert.Equal(t, "60s", config.Interval)
		assert.True(t, config.Verbose)
		assert.Len(t, config.Rooms, 2)
		assert.Equal(t, "bilibili", config.Rooms[0].Platform)
		assert.Equal(t, "123", config.Rooms[0].RoomID)
		assert.True(t, config.Rooms[0].Enabled)
	})

	t.Run("invalid JSON", func(t *testing.T) {
		tmpFile, err := ioutil.TempFile("", "test-config-*.json")
		require.NoError(t, err)
		defer os.Remove(tmpFile.Name())

		_, err = tmpFile.WriteString("invalid json")
		require.NoError(t, err)
		tmpFile.Close()

		_, err = loadConfig(tmpFile.Name())
		assert.Error(t, err)
	})

	t.Run("empty string config file", func(t *testing.T) {
		config, err := loadConfig("")
		assert.NoError(t, err)
		assert.Equal(t, "30s", config.Interval)
		assert.False(t, config.Verbose)
	})
}

func TestNewMonitor(t *testing.T) {
	t.Run("with valid config", func(t *testing.T) {
		// Create temporary config file
		configData := Config{
			Rooms: []RoomConfig{
				{Platform: "bilibili", RoomID: "123", Enabled: true},
				{Platform: "unknown", RoomID: "456", Enabled: true}, // This should be skipped
			},
			Interval: "30s",
			Verbose:  false,
		}

		data, err := json.Marshal(configData)
		require.NoError(t, err)

		tmpFile, err := ioutil.TempFile("", "test-config-*.json")
		require.NoError(t, err)
		defer os.Remove(tmpFile.Name())

		_, err = tmpFile.Write(data)
		require.NoError(t, err)
		tmpFile.Close()

		monitor, err := NewMonitor(tmpFile.Name())
		assert.NoError(t, err)
		assert.NotNil(t, monitor)
		assert.Len(t, monitor.sources, 1) // Only bilibili source should be created
	})

	t.Run("with no enabled rooms", func(t *testing.T) {
		configData := Config{
			Rooms: []RoomConfig{
				{Platform: "bilibili", RoomID: "123", Enabled: false},
			},
			Interval: "30s",
			Verbose:  false,
		}

		data, err := json.Marshal(configData)
		require.NoError(t, err)

		tmpFile, err := ioutil.TempFile("", "test-config-*.json")
		require.NoError(t, err)
		defer os.Remove(tmpFile.Name())

		_, err = tmpFile.Write(data)
		require.NoError(t, err)
		tmpFile.Close()

		monitor, err := NewMonitor(tmpFile.Name())
		assert.NoError(t, err)
		assert.NotNil(t, monitor)
		assert.Len(t, monitor.sources, 0)
	})
}

func TestMonitor_RunAndStop(t *testing.T) {
	// Create a monitor with minimal config
	configData := Config{
		Rooms: []RoomConfig{
			{Platform: "bilibili", RoomID: "76", Enabled: true},
		},
		Interval: "100ms", // Short interval for testing
		Verbose:  false,
	}

	data, err := json.Marshal(configData)
	require.NoError(t, err)

	tmpFile, err := ioutil.TempFile("", "test-config-*.json")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.Write(data)
	require.NoError(t, err)
	tmpFile.Close()

	monitor, err := NewMonitor(tmpFile.Name())
	require.NoError(t, err)

	// Start monitor in a goroutine
	done := make(chan error, 1)
	go func() {
		done <- monitor.Run()
	}()

	// Let it run for a short time
	time.Sleep(200 * time.Millisecond)

	// Stop the monitor
	monitor.Stop()

	// Wait for it to finish
	select {
	case err := <-done:
		assert.NoError(t, err)
	case <-time.After(time.Second):
		t.Fatal("Monitor did not stop within timeout")
	}
}

func TestMonitor_RunWithNoSources(t *testing.T) {
	configData := Config{
		Rooms:    []RoomConfig{}, // No rooms
		Interval: "30s",
		Verbose:  false,
	}

	data, err := json.Marshal(configData)
	require.NoError(t, err)

	tmpFile, err := ioutil.TempFile("", "test-config-*.json")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.Write(data)
	require.NoError(t, err)
	tmpFile.Close()

	monitor, err := NewMonitor(tmpFile.Name())
	require.NoError(t, err)

	err = monitor.Run()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no valid stream sources")
}