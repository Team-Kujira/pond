package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	Version string
	Commit  string
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print binary version",
	// Long: ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("version:", Version)
		fmt.Println("commit:", Commit)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
