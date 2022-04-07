package publish

import (
	"bufio"
	"context"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/dbason/opni-supportagent/pkg/input"
)

type rke2Shipper struct {
	ctx         context.Context
	username    string
	password    string
	endpoint    string
	clusterName string
	nodeName    string
	timezone    string
	year        string
	start       time.Time
	end         time.Time
}

func ShipRKE2ControlPlane(
	ctx context.Context,
	endpoint string,
	clusterName string,
	nodeName string,
	username string,
	password string,
) error {
	var (
		dateFile *os.File
		timezone string
		year     string
		err      error
	)
	// Extract timezone and year from the date output
	if _, err = os.Stat("systeminfo/date"); err == nil {
		dateFile, err = os.Open("systeminfo/date")
		if err != nil {
			return err
		}
		defer dateFile.Close()
	}

	scanner := bufio.NewScanner(dateFile)
	scanner.Scan()
	dateline := scanner.Text()
	re := regexp.MustCompile(`^[A-Z][a-z]{2} [A-Z][a-z]{2} \d{1,2} \d{2}:\d{2}:\d{2} ([A-Z]{3}) (\d{4})`)
	matches := re.FindStringSubmatch(dateline)
	if len(matches) != 0 {
		timezone = matches[1]
		year = matches[2]
	}

	shipper := rke2Shipper{
		endpoint:    endpoint,
		clusterName: clusterName,
		timezone:    timezone,
		year:        year,
	}

	err = shipper.shipEtcd()
	if err != nil {
		return err
	}

	err = shipper.shipKubelet()
	if err != nil {
		return err
	}

	err = shipper.shipKubeApiServer()
	if err != nil {
		return err
	}

	err = shipper.shipKubeControllerManager()
	if err != nil {
		return err
	}

	err = shipper.shipKubeScheduler()
	if err != nil {
		return err
	}

	err = shipper.shipKubeProxy()
	if err != nil {
		return err
	}

	err = shipper.shipRKE2JournalD()
	if err != nil {
		return err
	}

	doc := SupportFetcherDoc{
		Start: shipper.start,
		End:   shipper.end,
		Case:  clusterName,
	}

	return indexFetcherDoc(
		ctx,
		endpoint,
		username,
		password,
		doc,
	)
}

func (r *rke2Shipper) shipEtcd() error {
	parser := input.RKE2EtcdParser{}
	files, err := filepath.Glob("rke2/podlogs/kube-system-etcd-*")
	if err != nil {
		return err
	}
	os, err := input.NewOpensearchInput(r.ctx, r.endpoint, r.username, r.password, input.OpensearchConfig{
		ClusterID: r.clusterName,
		NodeName:  r.nodeName,
		Component: "etcd",
		Paths:     files,
	})
	if err != nil {
		return err
	}
	start, end, err := os.Publish(parser)
	if err != nil {
		return err
	}
	if r.start.IsZero() || start.Before(r.start) {
		r.start = start
	}
	if r.end.IsZero() || end.After(r.end) {
		r.end = end
	}
	return nil
}

func (r *rke2Shipper) shipKubelet() error {
	parser := input.NewDateZoneParser(r.timezone, r.year, input.DatetimeRegexK8s, input.LayoutK8s)
	os, err := input.NewOpensearchInput(r.ctx, r.endpoint, r.username, r.password, input.OpensearchConfig{
		ClusterID: r.clusterName,
		NodeName:  r.nodeName,
		Component: "kubelet",
		Paths:     []string{"rke2/agent-logs/kubelet.log"},
	})
	if err != nil {
		return err
	}
	start, end, err := os.Publish(parser)
	if err != nil {
		return err
	}
	if r.start.IsZero() || start.Before(r.start) {
		r.start = start
	}
	if r.end.IsZero() || end.After(r.end) {
		r.end = end
	}
	return nil
}

func (r *rke2Shipper) shipKubeApiServer() error {
	parser := input.NewDateZoneParser(r.timezone, r.year, input.DatetimeRegexK8s, input.LayoutK8s)
	files, err := filepath.Glob("rke2/podlogs/kube-system-kube-apiserver-*")
	if err != nil {
		return err
	}
	os, err := input.NewOpensearchInput(r.ctx, r.endpoint, r.username, r.password, input.OpensearchConfig{
		ClusterID: r.clusterName,
		NodeName:  r.nodeName,
		Component: "kube-apiserver",
		Paths:     files,
	})
	if err != nil {
		return err
	}
	start, end, err := os.Publish(parser)
	if err != nil {
		return err
	}
	if r.start.IsZero() || start.Before(r.start) {
		r.start = start
	}
	if r.end.IsZero() || end.After(r.end) {
		r.end = end
	}
	return nil
}

func (r *rke2Shipper) shipKubeControllerManager() error {
	parser := input.NewDateZoneParser(r.timezone, r.year, input.DatetimeRegexK8s, input.LayoutK8s)
	files, err := filepath.Glob("rke2/podlogs/kube-system-kube-controller-manager-*")
	if err != nil {
		return err
	}
	os, err := input.NewOpensearchInput(r.ctx, r.endpoint, r.username, r.password, input.OpensearchConfig{
		ClusterID: r.clusterName,
		NodeName:  r.nodeName,
		Component: "kube-controller-manager",
		Paths:     files,
	})
	if err != nil {
		return err
	}
	start, end, err := os.Publish(parser)
	if err != nil {
		return err
	}
	if r.start.IsZero() || start.Before(r.start) {
		r.start = start
	}
	if r.end.IsZero() || end.After(r.end) {
		r.end = end
	}
	return nil
}

func (r *rke2Shipper) shipKubeScheduler() error {
	parser := input.NewDateZoneParser(r.timezone, r.year, input.DatetimeRegexK8s, input.LayoutK8s)
	files, err := filepath.Glob("rke2/podlogs/kube-system-kube-scheduler-*")
	if err != nil {
		return err
	}
	os, err := input.NewOpensearchInput(r.ctx, r.endpoint, r.username, r.password, input.OpensearchConfig{
		ClusterID: r.clusterName,
		NodeName:  r.nodeName,
		Component: "kube-scheduler",
		Paths:     files,
	})
	if err != nil {
		return err
	}
	start, end, err := os.Publish(parser)
	if err != nil {
		return err
	}
	if r.start.IsZero() || start.Before(r.start) {
		r.start = start
	}
	if r.end.IsZero() || end.After(r.end) {
		r.end = end
	}
	return nil
}

func (r *rke2Shipper) shipKubeProxy() error {
	parser := input.NewDateZoneParser(r.timezone, r.year, input.DatetimeRegexK8s, input.LayoutK8s)
	files, err := filepath.Glob("rke2/podlogs/kube-system-kube-proxy-*")
	if err != nil {
		return err
	}
	os, err := input.NewOpensearchInput(r.ctx, r.endpoint, r.username, r.password, input.OpensearchConfig{
		ClusterID: r.clusterName,
		NodeName:  r.nodeName,
		Component: "kube-proxy",
		Paths:     files,
	})
	if err != nil {
		return err
	}
	start, end, err := os.Publish(parser)
	if err != nil {
		return err
	}
	if r.start.IsZero() || start.Before(r.start) {
		r.start = start
	}
	if r.end.IsZero() || end.After(r.end) {
		r.end = end
	}
	return nil
}

func (r *rke2Shipper) shipRKE2JournalD() error {
	parser := input.NewDateZoneParser(r.timezone, r.year, input.DatetimeRegexJournalD, input.LayoutJournalD)
	os, err := input.NewOpensearchInput(r.ctx, r.endpoint, r.username, r.password, input.OpensearchConfig{
		ClusterID: r.clusterName,
		NodeName:  r.nodeName,
		Component: "rke2",
		Paths:     []string{"journald/rke2-server"},
	})
	if err != nil {
		return err
	}
	start, end, err := os.Publish(parser)
	if err != nil {
		return err
	}
	if r.start.IsZero() || start.Before(r.start) {
		r.start = start
	}
	if r.end.IsZero() || end.After(r.end) {
		r.end = end
	}
	return nil
}
