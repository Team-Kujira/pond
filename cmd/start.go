package cmd

import (
	"pond/pond"

	"github.com/spf13/cobra"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start pond environment",
	// Long: ``,
	Run: func(cmd *cobra.Command, args []string) {
		pond, _ := pond.NewPond(LogLevel)
		pond.Start()
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}
