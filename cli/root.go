package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	cfgFile string
	rootCmd = &cobra.Command{
		Use:   "RestreamerMonitor",
		Short: "Restreamer Monitor 是一个多平台直播间监测与转播工具",
	}
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	// 全局配置文件标志
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "../config.json", "指定 JSON 配置文件路径")
}