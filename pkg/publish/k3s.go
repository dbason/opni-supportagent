package publish

import (
	"bufio"
	"os"
	"regexp"

	"github.com/dbason/opni-supportagent/pkg/input"
)

func ShipK3SControlPlane(endpoint string, clusterName string) error {
	var (
		dateFile  *os.File
		timezone  string
		year      string
		err       error
		fileInput input.ComponentInput
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
	fileInput = input.NewFileInput("k3s", "journald/k3s", clusterName)
	journaldParser := input.NewDateZoneParser(timezone, year, input.DatetimeRegexJournalD, input.LayoutJournalD)

	return fileInput.Publish(endpoint, journaldParser)
}
