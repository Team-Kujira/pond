package cmd

import (
	"fmt"
	"os"
	"strings"

	"pond/pond"
	"pond/pond/chain"
	"pond/pond/globals"
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
	Empty         bool
	ApiUrl        string
	RpcUrl        string
	KujiraVersion string
	Binary        string
	Horcrux       bool
	Overrides     string
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new pond environment",
	// Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		if NoContracts || Empty {
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

		overrides := []byte("{}")
		if Overrides != "" {
			overrides, err = os.ReadFile(Overrides)
			check(err)
		}

		// unbondingTime := fmt.Sprintf("%ds", UnbondingTime)

		overrides, err = utils.JsonMerge(overrides, []byte(fmt.Sprintf(`
		{
			"app_state": {
				"staking": {
					"params": {
						"unbonding_time": "%ds"
					}
				}
			}
		}`, UnbondingTime)))

		check(err)

		// options := map[string]string{
		// 	"app_state/staking/params/unbonding_time": unbondingTime,
		// }

		signers := make([]string, Nodes)
		for i := range signers {
			if i == 0 && Horcrux {
				signers[i] = "horcrux"
			} else {
				signers[i] = "local"
			}
		}

		config := pond.Config{
			Command:   "docker",
			Binary:    Binary,
			Namespace: Namespace,
			Address:   ListenAddress,
			ApiUrl:    ApiUrl,
			RpcUrl:    RpcUrl,
			Plans:     Contracts,
			Chains: []chain.Config{{
				Type:    "kujira",
				TypeNum: 1,
				Nodes:   Nodes,
				Signers: signers,
			}},
			Versions: globals.Versions,
		}

		if KujiraVersion != "" {
			config.Versions["kujira"] = KujiraVersion
		}

		pond, _ := pond.NewPond(LogLevel)
		pond.Init(
			config, Chains, overrides,
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
	initCmd.PersistentFlags().StringVar(&Binary, "binary", "", "Path to local Kujira binary")
	initCmd.PersistentFlags().StringVar(&Overrides, "overrides", "", "Path to genesis overrides")
	initCmd.PersistentFlags().BoolVar(&NoContracts, "no-contracts", false, "Don't deploy contracts on first start")
	initCmd.PersistentFlags().BoolVar(&Empty, "empty", false, "Don't deploy contracts on first start")
	initCmd.PersistentFlags().BoolVar(&Horcrux, "horcrux", false, "Use horcrux remote signers")

	initCmd.PersistentFlags().MarkDeprecated("no-contracts", "please use '--empty instead'")

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
