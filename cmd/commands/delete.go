package commands

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/dbason/opni-supportagent/pkg/util"
	"github.com/opensearch-project/opensearch-go"
	"github.com/opensearch-project/opensearch-go/opensearchtransport"
	"github.com/spf13/cobra"
	"github.com/tidwall/sjson"
)

func BuildDeleteCommand() *cobra.Command {
	command := &cobra.Command{
		Use:     "delete",
		Short:   "delete all logs for specified case",
		PreRunE: getDeletePassword,
		RunE:    deleteLogs,
	}

	return command
}

func deleteLogs(cmd *cobra.Command, args []string) error {
	caseNumber, err := cmd.Flags().GetString("case-number")
	if err != nil {
		return err
	}
	endpoint, err := cmd.Flags().GetString("endpoint")
	if err != nil {
		return err
	}

	username, err := cmd.Flags().GetString("username")
	if err != nil {
		return err
	}

	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.Dial = (&net.Dialer{
		Timeout: 5 * time.Second,
	}).Dial
	transport.TLSHandshakeTimeout = 5 * time.Second

	osCfg := opensearch.Config{
		Addresses: []string{
			endpoint,
		},
		Username:             username,
		Password:             password,
		UseResponseCheckOnly: true,
		Transport:            transport,
		Logger:               &opensearchtransport.ColorLogger{Output: os.Stdout},
	}

	osClient, err := opensearch.NewClient(osCfg)
	if err != nil {
		return err
	}

	query, _ := sjson.Set("", `query.term.cluster_id\.keyword`, caseNumber)
	util.Log.Infof("query is %s", query)

	resp, err := osClient.DeleteByQuery(
		[]string{
			"logs",
		},
		strings.NewReader(query),
		osClient.DeleteByQuery.WithWaitForCompletion(false),
	)
	if err != nil {
		return err
	}
	if resp.IsError() {
		resp.Body.Close()
		return fmt.Errorf("failed to queue delete: %s", resp.String())
	}

	resp.Body.Close()
	query, _ = sjson.Set("", `query.term.case\.keyword`, caseNumber)
	resp, err = osClient.DeleteByQuery(
		[]string{
			"pending-cases",
		},
		strings.NewReader(query),
		osClient.DeleteByQuery.WithWaitForCompletion(false),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.IsError() {
		return fmt.Errorf("failed to queue delete: %s", resp.String())
	}

	util.Log.Infof("logs for %s scheduled to be deleted in the background", caseNumber)

	return nil
}

func getDeletePassword(cmd *cobra.Command, args []string) error {
	var err error
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
