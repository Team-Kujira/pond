package cmd

import (
	"fmt"
	"os"

	"pond/pond"

	"github.com/spf13/cobra"
)

// txCmd represents the tx command
var txCmd = &cobra.Command{
	Use:   "tx",
	Short: "Transaction subcommands",
	Run: func(cmd *cobra.Command, args []string) {
		chain := "kujira-1"
		txArgs := []string{}

		for i := 0; i < len(args); i++ {
			switch args[i] {
			case "--chain-id":
				if i+1 < len(args) {
					chain = args[i+1]
					i++
				}
			default:
				txArgs = append(txArgs, args[i])
			}
		}

		pond, err := pond.NewPond(LogLevel)
		check(err)

		output, err := pond.Tx(chain, txArgs)
		if err != nil {
			if output == nil {
				// print error and exit
				check(err)
			}
			fmt.Fprint(os.Stderr, string(output))
			os.Exit(1)
		}
		fmt.Fprint(os.Stdout, string(output))
	},
}

func init() {
	txCmd.DisableFlagParsing = true
	rootCmd.AddCommand(txCmd)
}
