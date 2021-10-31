package publish

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"

	"github.com/dbason/opni-supportagent/pkg/input"
)

type rke2Shipper struct {
	endpoint    string
	clusterName string
	timezone    string
	year        string
}

func ShipRKE2ControlPlane(endpoint string, clusterName string) error {
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

	return nil
}

func (r *rke2Shipper) shipEtcd() error {
	parser := input.RKE2EtcdParser{}
	files, err := filepath.Glob("rke2/podlogs/kube-system-etcd-*")
	if err != nil {
		return err
	}
	for _, file := range files {
		fileInput := input.NewFileInput("etcd", file, r.clusterName)
		err = fileInput.Publish(r.endpoint, parser)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *rke2Shipper) shipKubelet() error {
	parser := input.NewDateZoneParser(r.timezone, r.year, input.DatetimeRegexK8s, input.LayoutK8s)
	fileInput := input.NewFileInput("kubelet", "rke2/agent-logs/kubelet.log", r.clusterName)
	return fileInput.Publish(r.endpoint, parser)
}

func (r *rke2Shipper) shipKubeApiServer() error {
	parser := input.NewDateZoneParser(r.timezone, r.year, input.DatetimeRegexK8s, input.LayoutK8s)
	files, err := filepath.Glob("rke2/podlogs/kube-system-kube-apiserver-*")
	if err != nil {
		return err
	}
	for _, file := range files {
		fileInput := input.NewFileInput("kube-apiserver", file, r.clusterName)
		err = fileInput.Publish(r.endpoint, parser)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *rke2Shipper) shipKubeControllerManager() error {
	parser := input.NewDateZoneParser(r.timezone, r.year, input.DatetimeRegexK8s, input.LayoutK8s)
	files, err := filepath.Glob("rke2/podlogs/kube-system-kube-controller-manager-*")
	if err != nil {
		return err
	}
	for _, file := range files {
		fileInput := input.NewFileInput("kube-controller-manager", file, r.clusterName)
		err = fileInput.Publish(r.endpoint, parser)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *rke2Shipper) shipKubeScheduler() error {
	parser := input.NewDateZoneParser(r.timezone, r.year, input.DatetimeRegexK8s, input.LayoutK8s)
	files, err := filepath.Glob("rke2/podlogs/kube-system-kube-scheduler-*")
	if err != nil {
		return err
	}
	for _, file := range files {
		fileInput := input.NewFileInput("kube-scheduler", file, r.clusterName)
		err = fileInput.Publish(r.endpoint, parser)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *rke2Shipper) shipKubeProxy() error {
	parser := input.NewDateZoneParser(r.timezone, r.year, input.DatetimeRegexK8s, input.LayoutK8s)
	files, err := filepath.Glob("rke2/podlogs/kube-system-kube-proxy-*")
	if err != nil {
		return err
	}
	for _, file := range files {
		fileInput := input.NewFileInput("kube-proxy", file, r.clusterName)
		err = fileInput.Publish(r.endpoint, parser)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *rke2Shipper) shipRKE2JournalD() error {
	parser := input.NewDateZoneParser(r.timezone, r.year, input.DatetimeRegexJournalD, input.LayoutJournalD)
	fileInput := input.NewFileInput("rke2", "journald/rke2-server", r.clusterName)
	return fileInput.Publish(r.endpoint, parser)
}
