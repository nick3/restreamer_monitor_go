package cli

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// 测试辅助函数：执行命令并捕获输出
func executeCommand(root *cobra.Command, args ...string) (output string, err error) {
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)
	
	err = root.Execute()
	return buf.String(), err
}

func TestRootCommand(t *testing.T) {
	// 测试默认配置文件路径
	assert.Equal(t, "../config.json", cfgFile)

	// 测试命令基本信息
	assert.Equal(t, "RestreamerMonitor", rootCmd.Use)
	assert.Equal(t, "Restreamer Monitor 是一个多平台直播间监测与转播工具", rootCmd.Short)

	// 测试配置文件标志
	output, err := executeCommand(rootCmd, "--help")
	assert.NoError(t, err)
	assert.Contains(t, output, "--config")
	assert.Contains(t, output, "-c")

	// 测试自定义配置文件路径
	_, err = executeCommand(rootCmd, "--config", "custom_config.json")
	assert.NoError(t, err)
	assert.Equal(t, "custom_config.json", cfgFile)

	// 测试自定义配置文件路径
	_, err = executeCommand(rootCmd, "--config", "custom_config.json")
	assert.NoError(t, err)
	assert.Equal(t, "custom_config.json", cfgFile)
}

// 测试 Execute 函数本身
func TestExecute(t *testing.T) {
	// 由于 Execute() 在错误时会调用 os.Exit(1)，这里只测试正常执行的情况
	oldRoot := rootCmd
	defer func() { rootCmd = oldRoot }()
	
	testCmd := &cobra.Command{Use: "test"}
	rootCmd = testCmd
	
	Execute()
}