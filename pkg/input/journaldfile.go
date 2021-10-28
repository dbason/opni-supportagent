package input

import (
	"bufio"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"time"

	"github.com/dbason/opni-supportagent/pkg/util"
	"github.com/elastic/go-elasticsearch/v7/esutil"
)

type JournalDFileInput struct {
	component string
	path      string
	client    *http.Client
	timezone  string
	year      string
}

func NewJournalDFileInput(component string, path string, timezone string, year string) *JournalDFileInput {
	return &JournalDFileInput{
		component: component,
		path:      path,
		client: &http.Client{
			Transport: &http.Transport{
				MaxConnsPerHost:   50,
				DisableKeepAlives: true,
			},
		},
		timezone: timezone,
		year:     year,
	}
}

func (j *JournalDFileInput) ComponentName() string {
	return j.component
}

func (j *JournalDFileInput) Publish(endpoint string) error {
	layout := "Jan 02 15:04:05 MST 2006"

	file, err := os.Open(j.path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	continueScan := scanner.Scan()
	lineCounter := 0
	var batchedMessages []LogMessage
	for continueScan {
		line := scanner.Text()

		re := regexp.MustCompile(datetimeRegexJournalD)
		datestring := re.FindString(line)
		if len(datestring) > 0 {
			// If the log is a valid log, and we previously reached the batch level sent it off
			if lineCounter >= batchSize {
				err = j.postBody(endpoint, batchedMessages)
				lineCounter = 0
				batchedMessages = []LogMessage{}
				if err != nil {
					return err
				}
			}
			datetime, err := time.Parse(layout, fmt.Sprintf("%s %s %s", datestring, j.timezone, j.year))
			if err != nil {
				util.Log.Errorf("error parsing date in log: %s", line)
				return err
			}

			log := LogMessage{
				Time:           datetime,
				Log:            line,
				Agent:          "support",
				IsControlPlane: true,
				Component:      j.component,
			}
			batchedMessages = append(batchedMessages, log)
		} else {
			util.Log.Debugf("log line has no date: %s", line)
			// If the log line has no date string append it to the previous entry
			if len(batchedMessages) > 0 {
				batchedMessages[lineCounter-1].Log = batchedMessages[lineCounter-1].Log + line
			}
		}

		// If we have reached the end of the file send the batch
		continueScan = scanner.Scan()
		if !continueScan {
			err = j.postBody(endpoint, batchedMessages)
			lineCounter = 0
			batchedMessages = []LogMessage{}
			if err != nil {
				return err
			}
		}
		lineCounter += 1
	}
	j.client.CloseIdleConnections()
	return scanner.Err()
}

func (j *JournalDFileInput) postBody(endpoint string, messages []LogMessage) error {
	url, err := url.ParseRequestURI(endpoint)
	if err != nil {
		return err
	}

	body := esutil.NewJSONReader(messages)

	req, err := http.NewRequest("POST", url.String(), body)
	if err != nil {
		return err
	}
	req.Header.Add(headerContentType, jsonContentHeader)

	res, err := j.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode >= 400 {
		return fmt.Errorf("publish failed: %s", res.Status)
	}
	return nil
}
