package input

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/dbason/opni-supportagent/pkg/util"
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

func (d *DateZoneParser) ParseTimestamp(log string) (time.Time, string, bool) {
	re := regexp.MustCompile(d.datetimeRegex)
	datestring := re.FindString(log)
	if len(datestring) == 0 {
		return time.Now(), log, false
	}
	datetime, err := time.Parse(d.layout, fmt.Sprintf("%s %s %s", datestring, d.timezone, d.year))
	if err != nil {
		util.Log.Panic(err)
	}

	retLog := log
	valid := true

	if d.datetimeRegex != KlogRegex {
		cleaned := strings.TrimSpace(re.ReplaceAllString(log, ""))
		retLog = cleaned

		re = regexp.MustCompile(KlogRegex)
		datestring = re.FindString(cleaned)
		valid = len(datestring) > 0
	}

	return datetime, retLog, valid
}
