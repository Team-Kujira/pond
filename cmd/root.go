package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	LogLevel string
	ChainId  string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "pond",
	Short: "Set up your local Kujira development environment",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&ChainId, "chain-id", "kujira-1", "Set chain-id")
	rootCmd.PersistentFlags().StringVar(&LogLevel, "log-level", "info", "Set log level")
}
