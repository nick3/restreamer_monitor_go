package cli

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// executeCommand executes a command and captures its output
func executeCommand(root *cobra.Command, args ...string) (output string, err error) {
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)
	
	err = root.Execute()
	return buf.String(), err
}

func TestRootCommand(t *testing.T) {
	t.Run("basic properties", func(t *testing.T) {
		assert.Equal(t, "RestreamerMonitor", rootCmd.Use)
		assert.Contains(t, rootCmd.Short, "多平台直播间监测与转播工具")
	})

	t.Run("help output", func(t *testing.T) {
		output, err := executeCommand(rootCmd, "--help")
		assert.NoError(t, err)
		assert.Contains(t, output, "--config")
		assert.Contains(t, output, "-c")
	})

	t.Run("config flag", func(t *testing.T) {
		flag := rootCmd.PersistentFlags().Lookup("config")
		assert.NotNil(t, flag)
		assert.Equal(t, "c", flag.Shorthand)
		assert.Equal(t, "../config.json", flag.DefValue)
	})

	t.Run("has subcommands", func(t *testing.T) {
		commands := rootCmd.Commands()
		assert.NotEmpty(t, commands)
		
		var hasMonitor, hasRelay bool
		for _, cmd := range commands {
			if cmd.Use == "monitor" {
				hasMonitor = true
			}
			if cmd.Use == "relay" {
				hasRelay = true
			}
		}
		
		assert.True(t, hasMonitor, "Should have monitor command")
		assert.True(t, hasRelay, "Should have relay command")
	})
}

func TestMonitorCommand(t *testing.T) {
	monitorCmd := findCommand(rootCmd, "monitor")
	assert.NotNil(t, monitorCmd)
	
	t.Run("has correct flags", func(t *testing.T) {
		intervalFlag := monitorCmd.Flags().Lookup("interval")
		assert.NotNil(t, intervalFlag)
		assert.Equal(t, "i", intervalFlag.Shorthand)
		
		verboseFlag := monitorCmd.Flags().Lookup("verbose")
		assert.NotNil(t, verboseFlag)
		assert.Equal(t, "v", verboseFlag.Shorthand)
	})
}

func TestRelayCommand(t *testing.T) {
	relayCmd := findCommand(rootCmd, "relay")
	assert.NotNil(t, relayCmd)
	assert.Equal(t, "relay", relayCmd.Use)
}

// Helper function to find a command by name
func findCommand(root *cobra.Command, name string) *cobra.Command {
	for _, cmd := range root.Commands() {
		if cmd.Use == name {
			return cmd
		}
	}
	return nil
}

func TestExecute(t *testing.T) {
	// Test that Execute function exists and can be called
	// We can't test the actual execution since it might exit the process
	assert.NotNil(t, Execute)
}