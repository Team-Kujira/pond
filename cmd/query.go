package cmd

import (
	"fmt"
	"os"

	"pond/pond"

	"github.com/spf13/cobra"
)

// queryCmd represents the query command
var queryCmd = &cobra.Command{
	Use:     "query",
	Aliases: []string{"q"},
	Short:   "Query subcommands",
	Run: func(cmd *cobra.Command, args []string) {
		chain := "kujira-1"
		queryArgs := []string{}

		for i := 0; i < len(args); i++ {
			switch args[i] {
			case "--chain-id":
				if i+1 < len(args) {
					chain = args[i+1]
					i++
				}
			default:
				queryArgs = append(queryArgs, args[i])
			}
		}

		pond, err := pond.NewPond(LogLevel)
		check(err)

		output, err := pond.Query(chain, queryArgs)
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
	queryCmd.DisableFlagParsing = true
	rootCmd.AddCommand(queryCmd)
}
