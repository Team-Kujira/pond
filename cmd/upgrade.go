package cmd

import (
	"pond/pond"

	"github.com/spf13/cobra"
)

// upgradeCmd represents the upgrade command
var upgradeCmd = &cobra.Command{
	Use:   "start",
	Short: "Upgrade kujira-1",
	// Long: ``,
	Run: func(cmd *cobra.Command, args []string) {
		pond, _ := pond.NewPond(LogLevel)
		pond.Upgrade()
	},
}

func init() {
	rootCmd.AddCommand(upgradeCmd)
}
