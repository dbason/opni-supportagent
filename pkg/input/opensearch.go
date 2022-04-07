package input

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/dbason/opni-supportagent/pkg/util"
	"github.com/opensearch-project/opensearch-go"
	"github.com/opensearch-project/opensearch-go/opensearchtransport"
	"github.com/opensearch-project/opensearch-go/opensearchutil"
)

type OpensearchInput struct {
	ctx    context.Context
	config OpensearchConfig
	*opensearch.Client
}

type OpensearchConfig struct {
	ClusterID string
	NodeName  string
	Paths     []string
	Component string
}

func NewOpensearchInput(
	ctx context.Context,
	opensearchURL string,
	username string,
	password string,
	config OpensearchConfig,
) (*OpensearchInput, error) {
	// Set sane transport timeouts
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.Dial = (&net.Dialer{
		Timeout: 5 * time.Second,
	}).Dial
	transport.TLSHandshakeTimeout = 5 * time.Second

	retryBackoff := backoff.NewExponentialBackOff()

	osCfg := opensearch.Config{
		Addresses: []string{
			opensearchURL,
		},
		Username:             username,
		Password:             password,
		UseResponseCheckOnly: true,
		Transport:            transport,
		RetryOnStatus:        []int{502, 503, 504, 429},
		RetryBackoff: func(i int) time.Duration {
			if i == 1 {
				retryBackoff.Reset()
			}
			return retryBackoff.NextBackOff()
		},
		Logger: &opensearchtransport.ColorLogger{Output: os.Stdout},
	}

	osClient, err := opensearch.NewClient(osCfg)
	if err != nil {
		return nil, err
	}

	return &OpensearchInput{
		ctx:    ctx,
		config: config,
		Client: osClient,
	}, nil
}

func (i *OpensearchInput) ComponentName() string {
	return i.config.Component
}

func (i *OpensearchInput) Publish(parser DateParser) (time.Time, time.Time, error) {
	var start, end time.Time
	indexer, err := opensearchutil.NewBulkIndexer(opensearchutil.BulkIndexerConfig{
		Index:  "logs",
		Client: i.Client,
	})
	if err != nil {
		return start, end, err
	}
	defer i.finalizeIndexing(indexer)

	for _, path := range i.config.Paths {
		// Read the file
		file, err := os.Open(path)
		if err != nil {
			return start, end, err
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		continueScan := scanner.Scan()
		var previousLog LogMessage
		for continueScan {
			line := scanner.Text()

			datetime, valid := parser.ParseTimestamp(line)
			if err != nil {
				util.Log.Errorf("error parsing date in log: %s", line)
				return start, end, err
			}
			if valid {
				if start.IsZero() || datetime.Before(start) {
					start = datetime
				}

				if end.IsZero() || datetime.After(end) {
					end = datetime
				}

				if (LogMessage{}) != previousLog {
					data, err := json.Marshal(previousLog)
					if err != nil {
						util.Log.Error("could not encode log to json")
					} else {
						err = indexer.Add(
							i.ctx,
							opensearchutil.BulkIndexerItem{
								Action: "index",
								Body:   bytes.NewReader(data),
								OnFailure: func(ctx context.Context, item opensearchutil.BulkIndexerItem, res opensearchutil.BulkIndexerResponseItem, err error) {
									if err != nil {
										util.Log.Errorf("%s", err)
									} else {
										util.Log.Errorf("%s: %s", res.Error.Type, res.Error.Reason)
									}
								},
							},
						)
						// Failing to add item to the bulk indexer is unrecoverable
						if err != nil {
							return start, end, err
						}
					}
				}
				previousLog = LogMessage{
					Time:           datetime,
					Timestamp:      datetime,
					Log:            line,
					Agent:          "support",
					IsControlPlane: true,
					Component:      i.config.Component,
					ClusterID:      i.config.ClusterID,
					NodeName:       i.config.NodeName,
				}
			} else {
				// if it's not a valid datetime add the log to the previous string
				if (LogMessage{}) != previousLog {
					previousLog.Log = previousLog.Log + line
				}
			}
			continueScan = scanner.Scan()
		}
		err = scanner.Err()
		if err != nil {
			util.Log.Errorf("error reading file %s: %s", path, err)
		}
	}

	return start, end, nil
}

func (i *OpensearchInput) finalizeIndexing(indexer opensearchutil.BulkIndexer) {
	indexer.Close(i.ctx)
	stats := indexer.Stats()
	util.Log.Infof("%s logs: %d flushed, %d failed", i.config.Component, stats.NumFlushed, stats.NumFailed)
}
