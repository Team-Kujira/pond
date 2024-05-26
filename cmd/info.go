package cmd

import (
	"fmt"
	"os"

	"pond/pond"

	"github.com/spf13/cobra"
)

var Account string

// infoCmd represents the info command
var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Print information",
	// Run: func(cmd *cobra.Command, args []string) {}
}

var accountsCmd = &cobra.Command{
	Use:   "accounts",
	Short: "List all accounts",
	Run: func(cmd *cobra.Command, args []string) {
		info, err := pond.LoadInfo()
		if err != nil {
			fmt.Fprint(os.Stderr, err)
			os.Exit(1)
		}

		info.ListAccounts()
	},
}

var seedsCmd = &cobra.Command{
	Use:   "seed [account]",
	Short: "List seed phrase",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		info, err := pond.LoadInfo()
		if err != nil {
			fmt.Fprint(os.Stderr, err)
			os.Exit(1)
		}

		info.ListSeed(args[0])
	},
}

var codesCmd = &cobra.Command{
	Use:   "codes",
	Short: "List all codes",
	Run: func(cmd *cobra.Command, args []string) {
		info, err := pond.LoadInfo()
		if err != nil {
			fmt.Fprint(os.Stderr, err)
			os.Exit(1)
		}

		err = info.ListCodes()
		check(err)
	},
}

var contractsCmd = &cobra.Command{
	Use:   "contracts",
	Short: "List all contracts",
	Run: func(cmd *cobra.Command, args []string) {
		info, err := pond.LoadInfo()
		if err != nil {
			fmt.Fprint(os.Stderr, err)
			os.Exit(1)
		}

		err = info.ListContracts()
		check(err)
	},
}

var infoUrlsCmd = &cobra.Command{
	Use:   "urls",
	Short: "List all urls",
	Run: func(cmd *cobra.Command, args []string) {
		info, err := pond.LoadInfo()
		if err != nil {
			fmt.Fprint(os.Stderr, err)
			os.Exit(1)
		}

		err = info.ListUrls()
		check(err)
	},
}

func init() {
	infoCmd.AddCommand(accountsCmd)
	infoCmd.AddCommand(seedsCmd)
	infoCmd.AddCommand(codesCmd)
	infoCmd.AddCommand(contractsCmd)
	infoCmd.AddCommand(infoUrlsCmd)

	rootCmd.AddCommand(infoCmd)
}
