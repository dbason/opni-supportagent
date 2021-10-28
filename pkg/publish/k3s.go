package publish

import (
	"bufio"
	"os"
	"regexp"

	"github.com/dbason/opni-supportagent/pkg/input"
)

func ShipK3SControlPlane(endpoint string) error {
	// Extract timezone and year from the date output
	file, err := os.Open("systeminfo/date")
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Scan()
	dateline := scanner.Text()
	re := regexp.MustCompile(`^[A-Z][a-z]{2} [A-Z][a-z]{2} \d{1,2} \d{2}:\d{2}:\d{2} ([A-Z]{3}) (\d{4})`)
	matches := re.FindStringSubmatch(dateline)

	journaldFile := input.NewJournalDFileInput("k3s", "journald/k3s", matches[1], matches[2])
	return journaldFile.Publish(endpoint)
}
