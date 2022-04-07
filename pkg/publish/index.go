package publish

import (
	"context"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/dbason/opni-supportagent/pkg/util"
	"github.com/opensearch-project/opensearch-go"
	"github.com/opensearch-project/opensearch-go/opensearchapi"
	"github.com/opensearch-project/opensearch-go/opensearchtransport"
	"github.com/opensearch-project/opensearch-go/opensearchutil"
)

type SupportFetcherDoc struct {
	Start time.Time `json:"start,omitempty"`
	End   time.Time `json:"end,omitempty"`
	Case  string    `json:"case,omitempty"`
}

const (
	supportDocIndexName = "pending-cases"
)

func indexFetcherDoc(
	ctx context.Context,
	opensearchURL string,
	username string,
	password string,
	doc SupportFetcherDoc,
) error {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.Dial = (&net.Dialer{
		Timeout: 5 * time.Second,
	}).Dial
	transport.TLSHandshakeTimeout = 5 * time.Second

	osCfg := opensearch.Config{
		Addresses: []string{
			opensearchURL,
		},
		Username:             username,
		Password:             password,
		UseResponseCheckOnly: true,
		Transport:            transport,
		Logger:               &opensearchtransport.ColorLogger{Output: os.Stdout},
	}

	osClient, err := opensearch.NewClient(osCfg)
	if err != nil {
		return err
	}

	req := opensearchapi.IndexRequest{
		Index: supportDocIndexName,
		Body:  opensearchutil.NewJSONReader(doc),
	}

	resp, err := req.Do(ctx, osClient)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.IsError() {
		util.Log.Errorf("failed to add pending cases doc: %s", resp.String())
	}

	return nil
}
