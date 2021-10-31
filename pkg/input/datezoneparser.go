package input

import (
	"fmt"
	"regexp"
	"time"

	"github.com/dbason/opni-supportagent/pkg/util"
)

const (
	DatetimeRegexJournalD = `^[A-Z][a-z]{2} \d{1,2} \d{2}:\d{2}:\d{2}`
	DatetimeRegexK8s      = `\d{4} \d{2}:\d{2}:\d{2}.\d{6}`
	LayoutJournalD        = "Jan 02 15:04:05 MST 2006"
	LayoutK8s             = "0102 15:04:05.999999 MST 2006"
)

type DateZoneParser struct {
	datetimeRegex string
	layout        string
	timezone      string
	year          string
}

func NewDateZoneParser(timezone string, year string, datetimeRegex string, layout string) *DateZoneParser {
	if timezone == "" {
		timezone = "UTC"
	}
	if year == "" {
		year = fmt.Sprint((time.Now().Year()))
	}
	return &DateZoneParser{
		datetimeRegex: datetimeRegex,
		layout:        layout,
		timezone:      timezone,
		year:          year,
	}
}

func (d *DateZoneParser) ParseTimestamp(log string) (time.Time, bool) {
	re := regexp.MustCompile(d.datetimeRegex)
	datestring := re.FindString(log)
	if len(datestring) == 0 {
		return time.Now(), false
	}
	datetime, err := time.Parse(d.layout, fmt.Sprintf("%s %s %s", datestring, d.timezone, d.year))
	if err != nil {
		util.Log.Panic(err)
	}
	return datetime, true
}
