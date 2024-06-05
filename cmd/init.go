package cmd

import (
	"fmt"
	"strings"

	"pond/pond"
	"pond/pond/templates"
	"pond/utils"

	"github.com/spf13/cobra"
)

var (
	Namespace     string
	Nodes         uint
	Chains        []string
	Contracts     []string
	UnbondingTime uint
	ListenAddress string
	NoContracts   bool
	ApiUrl        string
	RpcUrl        string
	KujiraVersion string
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new pond environment",
	// Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		if NoContracts {
			Contracts = []string{}
		}

		chains, err := templates.GetChains()
		check(err)

		missing := utils.ListDiff(chains, Chains)
		if len(missing) > 0 {
			err := fmt.Errorf(
				"chain(s) not supported: %s", strings.Join(missing, ", "),
			)
			check(err)
		}

		plans, err := templates.GetPlans()
		check(err)

		missing = utils.ListDiff(plans, Contracts)
		if len(missing) > 0 {
			err := fmt.Errorf(
				"contract(s) not supported: %s", strings.Join(missing, ", "),
			)
			check(err)
		}

		unbondingTime := fmt.Sprintf("%ds", UnbondingTime)

		options := map[string]string{
			"app_state/staking/params/unbonding_time": unbondingTime,
		}

		pond, _ := pond.NewPond(LogLevel)
		pond.Init(
			"docker", Namespace, ListenAddress, ApiUrl, RpcUrl, KujiraVersion,
			Chains, Contracts, Nodes, options,
		)
	},
}

func init() {
	rootCmd.AddCommand(initCmd)

	initCmd.PersistentFlags().UintVar(&Nodes, "nodes", 1, "Set number of validator nodes")
	initCmd.PersistentFlags().UintVar(&UnbondingTime, "unbonding-time", 1209600, "Set unbonding time in seconds")
	initCmd.PersistentFlags().StringVar(&Namespace, "namespace", "teamkujira", "Set docker.io namespace")
	initCmd.PersistentFlags().StringVar(&ListenAddress, "listen", "127.0.0.1", "Set listen address")
	initCmd.PersistentFlags().StringVar(&ApiUrl, "api-url", "https://rest.cosmos.directory/kujira", "Set API URL")
	initCmd.PersistentFlags().StringVar(&RpcUrl, "rpc-url", "https://rpc.cosmos.directory/kujira", "Set RPC URL")
	initCmd.PersistentFlags().StringVar(&KujiraVersion, "kujira-version", "", "Set Kujira version")
	initCmd.PersistentFlags().BoolVar(&NoContracts, "no-contracts", false, "Don't deploy contracts on first start")

	chains, err := templates.GetChains()
	check(err)

	initCmd.PersistentFlags().StringSliceVar(
		&Chains, "chains", []string{}, fmt.Sprintf(
			"Set extra chains (default [])\nOptions: %s",
			strings.Join(chains, ", "),
		),
	)

	plans, err := templates.GetPlans()
	check(err)

	initCmd.PersistentFlags().StringSliceVar(
		&Contracts, "contracts", []string{"kujira"}, fmt.Sprintf(
			"Create contracts on first start (default [kujira])\nOptions: %s",
			strings.Join(plans, ", "),
		),
	)
}
