package commands

import (
	"errors"

	"github.com/AlecAivazis/survey/v2"
	"github.com/dbason/opni-supportagent/pkg/publish"
	"github.com/spf13/cobra"
)

var (
	password     string
	distribution string
)

type Distribution string

func BuildPublishCommand() *cobra.Command {
	command := &cobra.Command{
		Use:     "publish cluster-type",
		Short:   "publish support bundle to remote opni cluster",
		PreRunE: getPassword,
		RunE:    publishLogs,
	}

	return command
}

func publishLogs(cmd *cobra.Command, args []string) error {
	var err error
	caseNumber, err := cmd.Flags().GetString("case-number")
	if err != nil {
		return err
	}
	endpoint, err := cmd.Flags().GetString("endpoint")
	if err != nil {
		return err
	}
	nodeName, err := cmd.Flags().GetString("node-name")
	if err != nil {
		return err
	}
	username, err := cmd.Flags().GetString("username")
	if err != nil {
		return err
	}

	switch Distribution(args[0]) {
	case RKE:
		err = publish.ShipRKEControlPlane(
			cmd.Context(),
			endpoint,
			caseNumber,
			nodeName,
			username,
			password,
		)
	case K3S:
		err = publish.ShipK3SControlPlane(
			cmd.Context(),
			endpoint,
			caseNumber,
			nodeName,
			username,
			password,
		)
	case RKE2:
		err = publish.ShipRKE2ControlPlane(
			cmd.Context(),
			endpoint,
			caseNumber,
			nodeName,
			username,
			password,
		)
	default:
		err = errors.New("distribution must be one of rke, rke2, k3s")
	}
	return err
}

func getPassword(cmd *cobra.Command, args []string) error {
	var err error
	if len(args) != 1 {
		return errors.New("publish requires exactly 1 arguments; the distribution")
	}
	password, err = cmd.Flags().GetString("password")
	if err != nil {
		return err
	}

	if password != "" {
		return nil
	}

	return survey.AskOne(
		&survey.Password{
			Message: "please enter the opensearch password",
		},
		&password,
		survey.WithValidator(survey.Required),
	)
}
