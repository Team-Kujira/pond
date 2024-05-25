package cmd

import (
	"pond/pond"

	"github.com/spf13/cobra"
)

// stopCmd represents the stop command
var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop pond environment",
	// Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		pond, _ := pond.NewPond(LogLevel)
		pond.Stop()
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
}
