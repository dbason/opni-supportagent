package input

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/dbason/opni-supportagent/pkg/util"
)

const (
	EtcdTimestampRegex  = `^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}.\d{6}`
	EtcdTimestampLayout = "2006-01-02 15:04:05.999999Z07:00"
	RFC3339Milli        = "2006-01-02T15:04:05.999Z07:00"
)

type RKE2EtcdParser struct{}

type EtcdJSONLog struct {
	LogLevel  string `json:"level,omitempty"`
	Timestamp string `json:"ts,omitempty"`
	Message   string `json:"msg,omitempty"`
}

func (r RKE2EtcdParser) ParseTimestamp(log string) (time.Time, bool) {
	if strings.HasPrefix(log, "{") {
		jsonLog := &EtcdJSONLog{}
		json.Unmarshal([]byte(log), jsonLog)
		datetime, err := time.Parse(RFC3339Milli, jsonLog.Timestamp)
		if err != nil {
			util.Log.Panic(err)
		}
		return datetime, true
	}
	re := regexp.MustCompile(EtcdTimestampRegex)
	datestring := re.FindString(log)
	if len(datestring) == 0 {
		util.Log.Warnf("no date found in log: %s", log)
		return time.Now(), true
	}
	datetime, err := time.Parse(EtcdTimestampLayout, fmt.Sprintf("%sZ", datestring))
	if err != nil {
		util.Log.Panic(err)
	}
	return datetime, true
}
