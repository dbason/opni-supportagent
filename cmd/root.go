package cmd

import (
	"context"
	"os"

	"github.com/dbason/opni-supportagent/cmd/commands"
	"github.com/spf13/cobra"
)

func BuildRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "opni-support",
		Short: "Rancher Support agent for opni",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	rootCmd.AddCommand(commands.BuildPublishCommand())
	rootCmd.AddCommand(commands.BuildDeleteCommand())

	rootCmd.PersistentFlags().String("case-number", "", "case number to store the logs under")
	rootCmd.PersistentFlags().String("endpoint", "https://support-opensearch.danbason.dev", "Opensearch endpoint to publish logs to")
	rootCmd.PersistentFlags().String("node-name", "default-node", "node name to attach to the logs")
	rootCmd.PersistentFlags().String("username", "index-user", "username for Opensearch")
	rootCmd.PersistentFlags().String("password", "", "password for Opensearch")

	rootCmd.MarkFlagRequired("case-number")

	return rootCmd
}

func Execute() {
	if err := BuildRootCmd().ExecuteContext(context.Background()); err != nil {
		os.Exit(1)
	}
}
