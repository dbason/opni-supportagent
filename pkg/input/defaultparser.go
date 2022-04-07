package input

import (
	"regexp"
	"time"

	"github.com/dbason/opni-supportagent/pkg/util"
)

const (
	datetimeRegexISO8601 = `^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{9}Z`
)

type DefaultParser struct {
}

func (p *DefaultParser) ParseTimestamp(log string) (time.Time, bool) {
	re := regexp.MustCompile(datetimeRegexISO8601)
	datestring := re.FindString(log)
	datetime, err := time.Parse(time.RFC3339Nano, datestring)
	if err != nil {
		util.Log.Panic(err)
	}
	return datetime, true
}
