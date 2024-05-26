package cmd

import (
	"pond/pond"

	"github.com/spf13/cobra"
)

// deployCmd represents the deploy command
var deployCmd = &cobra.Command{
	Use:   "deploy [file]",
	Short: "Deploy plan or wasm files",
	Args:  cobra.RangeArgs(1, 99),
	Run: func(cmd *cobra.Command, args []string) {
		pond, err := pond.NewPond(LogLevel)
		check(err)

		// err = pond.DeployPlanfile(args[0])
		err = pond.Deploy(args)
		check(err)
	},
}

func init() {
	rootCmd.AddCommand(deployCmd)
}
