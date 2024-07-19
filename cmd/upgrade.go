package cmd

import (
	"pond/pond"

	"github.com/spf13/cobra"
)

// upgradeCmd represents the upgrade command
var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade kujira-1",
	// Long: ``,
	Run: func(cmd *cobra.Command, args []string) {
		pond, _ := pond.NewPond(LogLevel)
		err := pond.Upgrade(Version, Binary)
		check(err)
	},
}

func init() {
	rootCmd.AddCommand(upgradeCmd)

	upgradeCmd.PersistentFlags().StringVar(&Binary, "binary", "", "Path to new local Kujira binary")
	upgradeCmd.PersistentFlags().StringVar(&Version, "version", "", "New Kujira version")
}
