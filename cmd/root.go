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
	rootCmd.AddCommand(commands.BuildLocalCommand())

	return rootCmd
}

func Execute() {
	if err := BuildRootCmd().ExecuteContext(context.Background()); err != nil {
		os.Exit(1)
	}
}
