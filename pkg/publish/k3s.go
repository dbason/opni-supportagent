package publish

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/dbason/opni-supportagent/pkg/input"
	"github.com/dbason/opni-supportagent/pkg/util"
)

func ShipK3SControlPlane(
	ctx context.Context,
	endpoint string,
	clusterName string,
	nodeName string,
	username string,
	password string,
) error {
	var (
		dateFile *os.File
		timezone string
		year     string
		err      error
	)
	// Extract timezone and year from the date output
	if _, err = os.Stat("systeminfo/date"); err == nil {
		dateFile, err = os.Open("systeminfo/date")
		if err != nil {
			return err
		}
		defer dateFile.Close()
	}

	scanner := bufio.NewScanner(dateFile)
	scanner.Scan()
	dateline := scanner.Text()
	re := regexp.MustCompile(`^[A-Z][a-z]{2} [A-Z][a-z]{2} \d{1,2} \d{2}:\d{2}:\d{2} ([A-Z]{3}) (\d{4})`)
	matches := re.FindStringSubmatch(dateline)
	if len(matches) != 0 {
		timezone = matches[1]
		year = matches[2]
	}

	if _, err := os.Stat("journald/k3s"); err != nil {
		return err
	}
	opensearch, err := input.NewOpensearchInput(ctx, endpoint, username, password, input.OpensearchConfig{
		ClusterID: clusterName,
		Paths:     []string{"journald/k3s"},
		Component: "k3s",
		NodeName:  nodeName,
	})
	if err != nil {
		return err
	}

	journaldParser := input.NewDateZoneParser(timezone, year, input.JournaldRegex, input.JournaldLayout)

	_, _, err = opensearch.Publish(journaldParser, input.LogTypeControlplane)
	if err != nil {
		return err
	}

	files, err := filepath.Glob("k3s/podlogs/cattle-system-rancher-*")
	if err != nil {
		util.Log.Errorf("unable to list rancher files: %s", err)
		return nil
	}

	parser := &input.MultipleParser{
		Dateformats: []input.Dateformat{
			{
				DateRegex: input.RancherRegex,
				Layout:    input.RancherLayout,
			},
			{
				DateRegex:  input.KlogRegex,
				Layout:     input.KlogLayout,
				DateSuffix: fmt.Sprintf(" %s %s", timezone, year),
			},
		},
	}

	rancher, err := input.NewOpensearchInput(ctx, endpoint, username, password, input.OpensearchConfig{
		ClusterID: clusterName,
		NodeName:  nodeName,
		Component: "",
		Paths:     files,
	})
	if err != nil {
		return err
	}

	util.Log.Info("publishing rancher server logs")
	_, _, err = rancher.Publish(parser, input.LogTypeRancher)
	return err

}
