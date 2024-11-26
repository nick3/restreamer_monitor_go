package cli

import (
	"log"

	"github.com/spf13/cobra"
)

func init() {
	var monitorCmd = &cobra.Command{
		Use:   "monitor",
		Short: "监控直播间开播与下播状态，实时发送通知到指定的 Telegram 聊天中",
		Run: func(cmd *cobra.Command, args []string) {
			// m := monitor.NewMonitor(cfgFile)
			// m.Run()
			log.Println("monitor")
		},
	}

	monitorCmd.Flags().StringP("interval", "i", "30s", "监控检查间隔")
	monitorCmd.Flags().BoolP("verbose", "v", false, "启用详细日志")

	rootCmd.AddCommand(monitorCmd)
}