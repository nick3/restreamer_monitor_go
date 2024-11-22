package cli

import (
	"log"

	"github.com/spf13/cobra"
)

func init() {
	var relayCmd = &cobra.Command{
		Use:   "relay",
		Short: "转播直播间",
		Run: func(cmd *cobra.Command, args []string) {
			// r := relay.NewRelay(cfgFile)
			// r.Run()

			log.Println("relay")
		},
	}

	rootCmd.AddCommand(relayCmd)
}