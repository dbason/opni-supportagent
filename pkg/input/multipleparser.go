package input

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/dbason/opni-supportagent/pkg/util"
)

const (
	leadingDateRegex = `^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{9}Z`
)

type MultipleParser struct {
	Dateformats      []Dateformat
	StripLeadingDate bool
}

type Dateformat struct {
	DateRegex  string
	Layout     string
	DateSuffix string
}

func (p *MultipleParser) ParseTimestamp(log string) (time.Time, string, bool) {
	var datetime time.Time
	var err error
	if p.StripLeadingDate {
		re := regexp.MustCompile(leadingDateRegex)
		datestring := re.FindString(log)
		datetime, err = time.Parse(time.RFC3339Nano, datestring)
		if err != nil {
			util.Log.Panic(err)
		}
		log = strings.TrimSpace(re.ReplaceAllString(log, ""))
	}

	for _, dateFormat := range p.Dateformats {
		re := regexp.MustCompile(dateFormat.DateRegex)
		datestring := re.FindString(log)
		if len(datestring) == 0 {
			continue
		}
		if !p.StripLeadingDate {
			datetime, err = time.Parse(dateFormat.Layout, fmt.Sprintf("%s%s", datestring, dateFormat.DateSuffix))
			if err != nil {
				util.Log.Panic(err)
			}
		}
		return datetime, log, true
	}

	return time.Now(), log, false
}
