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

const (
	datetimeRegexISO8601 = `^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{9}Z`
)

type FileInput struct {
	component   string
	path        string
	client      *http.Client
	clusterName string
}

func NewFileInput(component string, path string, clusterName string) *FileInput {
	return &FileInput{
		component:   component,
		path:        path,
		clusterName: clusterName,
		client: &http.Client{
			Transport: &http.Transport{
				MaxConnsPerHost:   50,
				DisableKeepAlives: true,
			},
		},
	}
}

func (f *FileInput) Publish(endpoint string, parser DateParser) error {
	file, err := os.Open(f.path)
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

		datetime, valid := parser.ParseTimestamp(line)
		if err != nil {
			util.Log.Errorf("error parsing date in log: %s", line)
			return err
		}
		if valid {
			// If the log is a valid log, and we previously reached the batch level sent it off
			if lineCounter >= batchSize {
				err = f.postBody(endpoint, batchedMessages)
				lineCounter = 0
				batchedMessages = []LogMessage{}
				if err != nil {
					return err
				}
			}

			log := LogMessage{
				Time:           datetime,
				Log:            line,
				Agent:          "support",
				IsControlPlane: true,
				Component:      f.component,
				ClusterName:    f.clusterName,
			}
			batchedMessages = append(batchedMessages, log)
		} else {
			util.Log.Debugf("log line has no date: %s", line)
			// If the log line has no date string append it to the previous entry
			if len(batchedMessages) > 0 {
				lineCounter -= 1
				batchedMessages[lineCounter].Log = batchedMessages[lineCounter].Log + line
			}
		}

		// If we have reached the end of the file send the batch
		continueScan = scanner.Scan()
		if !continueScan {
			err = f.postBody(endpoint, batchedMessages)
			lineCounter = 0
			batchedMessages = []LogMessage{}
			if err != nil {
				return err
			}
		}
		lineCounter += 1
	}
	f.client.CloseIdleConnections()
	return scanner.Err()
}

func (f *FileInput) ComponentName() string {
	return f.component
}

func (f *FileInput) ParseTimestamp(log string) (time.Time, bool) {
	re := regexp.MustCompile(datetimeRegexISO8601)
	datestring := re.FindString(log)
	datetime, err := time.Parse(time.RFC3339Nano, datestring)
	if err != nil {
		util.Log.Panic(err)
	}
	return datetime, true
}

func (f *FileInput) postBody(endpoint string, messages []LogMessage) error {
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

	res, err := f.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode >= 400 {
		return fmt.Errorf("publish failed: %s", res.Status)
	}
	return nil
}
