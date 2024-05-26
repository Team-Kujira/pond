package cmd

import (
	"os"

	"pond/pond"

	"github.com/spf13/cobra"
)

var (
	NewRegistryName   string
	NewRegistrySource string
)

// registryCmd represents the registry command
var registryCmd = &cobra.Command{
	Use:   "registry",
	Short: "Manage code registry",
}

var listRegistryCmd = &cobra.Command{
	Use:   "list",
	Short: "List all registered codes",
	Run: func(cmd *cobra.Command, args []string) {
		pond, err := pond.NewPond(LogLevel)
		check(err)

		err = pond.ListRegistry()
		check(err)
	},
}

var updateRegistryCmd = &cobra.Command{
	Use:   "update [name]",
	Short: "Update registry item",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		updates := map[string]string{}

		if NewRegistryName != "" {
			updates["name"] = NewRegistryName
		}

		if NewRegistrySource != "" {
			updates["source"] = NewRegistrySource
		}

		pond, err := pond.NewPond(LogLevel)
		check(err)

		err = pond.UpdateRegistry(args[0], updates)
		if err != nil {
			os.Exit(1)
		}
	},
}

var exportRegistryCmd = &cobra.Command{
	Use:   "export [file]",
	Short: "Export registry to json",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		pond, err := pond.NewPond(LogLevel)
		check(err)

		err = pond.ExportRegistry(args[0])
		if err != nil {
			os.Exit(1)
		}
	},
}

var importRegistryCmd = &cobra.Command{
	Use:   "import [file]",
	Short: "Import registry from json",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		pond, err := pond.NewPond(LogLevel)
		check(err)

		err = pond.ImportRegistry(args[0])
		if err != nil {
			os.Exit(1)
		}
	},
}

func init() {
	registryCmd.AddCommand(listRegistryCmd)

	updateRegistryCmd.PersistentFlags().StringVar(&NewRegistryName, "name", "", "Update name")
	updateRegistryCmd.PersistentFlags().StringVar(&NewRegistrySource, "source", "", "Update source")
	registryCmd.AddCommand(updateRegistryCmd)
	registryCmd.AddCommand(exportRegistryCmd)
	registryCmd.AddCommand(importRegistryCmd)

	rootCmd.AddCommand(registryCmd)
}
