package commands

import (
	"errors"

	"github.com/dbason/opni-supportagent/pkg/rke"
	"github.com/spf13/cobra"
)

func BuildRKECommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "rke",
		Short: "publish RKE support bundle",
		RunE:  publishRKE,
	}
	return command
}

func publishRKE(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return errors.New("rke accepts the endpoint as a single argument")
	}
	return rke.ShipRKEControlPlane(args[0])
}
