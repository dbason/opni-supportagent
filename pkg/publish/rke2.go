package publish

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/dbason/opni-supportagent/pkg/input"
	"github.com/dbason/opni-supportagent/pkg/util"
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
		ctx:         ctx,
		endpoint:    endpoint,
		clusterName: clusterName,
		timezone:    timezone,
		year:        year,
		username:    username,
		password:    password,
	}

	err = shipper.shipEtcd()
	if err != nil {
		return err
	}

	err = shipper.shipKubelet()
	if err != nil {
		return err
	}

	err = shipper.shipKubeAPIServer()
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

	err = shipper.shipRancher()
	return err

}

func (r *rke2Shipper) shipEtcd() error {
	parser := &input.RKE2EtcdParser{}
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
	start, end, err := os.Publish(parser, input.LogTypeControlplane)
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
	parser := input.NewDateZoneParser(r.timezone, r.year, input.KlogRegex, input.KlogLayout)
	os, err := input.NewOpensearchInput(r.ctx, r.endpoint, r.username, r.password, input.OpensearchConfig{
		ClusterID: r.clusterName,
		NodeName:  r.nodeName,
		Component: "kubelet",
		Paths:     []string{"rke2/agent-logs/kubelet.log"},
	})
	if err != nil {
		return err
	}
	start, end, err := os.Publish(parser, input.LogTypeControlplane)
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

func (r *rke2Shipper) shipKubeAPIServer() error {
	parser := input.NewDateZoneParser(r.timezone, r.year, input.KlogRegex, input.KlogLayout)
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
	start, end, err := os.Publish(parser, input.LogTypeControlplane)
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
	parser := input.NewDateZoneParser(r.timezone, r.year, input.KlogRegex, input.KlogLayout)
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
	start, end, err := os.Publish(parser, input.LogTypeControlplane)
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
	parser := input.NewDateZoneParser(r.timezone, r.year, input.KlogRegex, input.KlogLayout)
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
	start, end, err := os.Publish(parser, input.LogTypeControlplane)
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
	parser := input.NewDateZoneParser(r.timezone, r.year, input.KlogRegex, input.KlogLayout)
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
	start, end, err := os.Publish(parser, input.LogTypeControlplane)
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
	parser := input.NewDateZoneParser(r.timezone, r.year, input.JournaldRegex, input.JournaldLayout)
	os, err := input.NewOpensearchInput(r.ctx, r.endpoint, r.username, r.password, input.OpensearchConfig{
		ClusterID: r.clusterName,
		NodeName:  r.nodeName,
		Component: "rke2",
		Paths:     []string{"journald/rke2-server"},
	})
	if err != nil {
		return err
	}
	start, end, err := os.Publish(parser, input.LogTypeControlplane)
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

func (r *rke2Shipper) shipRancher() error {
	files, err := filepath.Glob("rke2/podlogs/cattle-system-rancher-*")
	if err != nil {
		util.Log.Errorf("unable to list rancher files: %s", err)
		return err
	}
	parser := &input.MultipleParser{
		Dateformats: []input.Dateformat{
			{
				DateRegex: input.RancherRegex,
				Layout:    input.RancherLayout,
			},
			{
				DateRegex:  input.KlogRegex,
				Layout:     input.KlogLayout,
				DateSuffix: fmt.Sprintf(" %s %s", r.timezone, r.year),
			},
		},
	}

	rancher, err := input.NewOpensearchInput(r.ctx, r.endpoint, r.username, r.password, input.OpensearchConfig{
		ClusterID: r.clusterName,
		NodeName:  r.nodeName,
		Component: "",
		Paths:     files,
	})
	if err != nil {
		return err
	}

	util.Log.Info("publishing rancher server logs")
	_, _, err = rancher.Publish(parser, input.LogTypeRancher)
	return err
}
