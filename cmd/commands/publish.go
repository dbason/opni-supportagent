package commands

import (
	"errors"

	"github.com/dbason/opni-supportagent/pkg/publish"
	"github.com/spf13/cobra"
)

type Distribution string

func BuildPublishCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "publish",
		Short: "publish support bundle to remote opni cluster",
		RunE:  publishLogs,
	}

	return command
}

func publishLogs(cmd *cobra.Command, args []string) error {
	var err error
	if len(args) != 2 {
		return errors.New("publish requires exactly 2 arguments, the distribution and the endpoint")
	}
	switch Distribution(args[0]) {
	case RKE:
		err = publish.ShipRKEControlPlane(args[1])
	case K3S:
		err = publish.ShipK3SControlPlane(args[1])
	case RKE2:
		err = errors.New("distribution not currently supported")
	default:
		err = errors.New("distribution must be one of rke, rke2, k3s")
	}
	return err
}
