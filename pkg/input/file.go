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
	datetimeRegex     = `^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{9}Z`
	headerContentType = "Content-Type"
	jsonContentHeader = "application/json"
	batchSize         = 20
)

type LogMessage struct {
	Time           time.Time `json:"time,omitempty"`
	Log            string    `json:"log,omitempty"`
	Agent          bool      `json:"agent,omitempty"`
	IsControlPlane bool      `json:"is_control_plane_log,omitempty"`
	Component      string    `json:"kubernetes_component,omitempty"`
}

type FileInput struct {
	component string
	path      string
	client    *http.Client
}

func NewFileInput(component string, path string) *FileInput {
	return &FileInput{
		component: component,
		path:      path,
		client:    &http.Client{},
	}
}

func (f *FileInput) Publish(endpoint string) error {
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
		lineCounter += 1
		line := scanner.Text()

		re := regexp.MustCompile(datetimeRegex)
		datestring := re.FindString(line)
		datetime, err := time.Parse(time.RFC3339Nano, datestring)
		if err != nil {
			util.Log.Errorf("error parsing date in log: %s", line)
			return err
		}

		log := LogMessage{
			Time:           datetime,
			Log:            line,
			Agent:          true,
			IsControlPlane: true,
			Component:      f.component,
		}
		batchedMessages = append(batchedMessages, log)

		// If we have reached the batch limit, or the end of the file send the batch
		continueScan = scanner.Scan()
		if lineCounter >= batchSize || !continueScan {
			err = f.postBody(endpoint, batchedMessages)
			lineCounter = 0
			batchedMessages = []LogMessage{}
			if err != nil {
				return err
			}
		}
	}
	return scanner.Err()
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
	res.Body.Close()

	if res.StatusCode >= 400 {
		return fmt.Errorf("publish failed: %s", res.Status)
	}
	return nil
}
