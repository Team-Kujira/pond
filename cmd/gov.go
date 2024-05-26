package cmd

import (
	"fmt"

	"pond/pond"

	"github.com/spf13/cobra"
)

var VoteOption string

// govCmd represents the gov command
var govCmd = &cobra.Command{
	Use:   "gov",
	Short: "Handle gov proposals",
}

// govSubmitCmd represents the gov submit command
var govSubmitCmd = &cobra.Command{
	Use:   "submit [file]",
	Short: "Submit gov proposal json",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		switch VoteOption {
		case "yes", "no", "abstain", "veto", "":
			break
		default:
			err := fmt.Errorf(
				"vote option must be one of: yes, no, abstain, veto",
			)
			check(err)
		}

		pond, err := pond.NewPond(LogLevel)
		check(err)

		err = pond.SubmitProposal(args[0], VoteOption)
		check(err)
	},
}

func init() {
	govSubmitCmd.PersistentFlags().StringVar(&VoteOption, "vote", "", "vote option")

	govCmd.AddCommand(govSubmitCmd)
	rootCmd.AddCommand(govCmd)
}
