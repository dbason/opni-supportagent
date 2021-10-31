package publish

import (
	"os"
	"reflect"

	"github.com/dbason/opni-supportagent/pkg/input"
	"github.com/dbason/opni-supportagent/pkg/util"
)

func ShipRKEControlPlane(endpoint string, clusterName string) error {
	for _, component := range []*input.FileInput{
		createETCDInput(clusterName),
		createKubeAPIInput(clusterName),
		createKubeletInput(clusterName),
		createKubeControllerManagerInput(clusterName),
		createKubeSchedulerInput(clusterName),
		createKubeProxyInput(clusterName),
	} {
		if !reflect.ValueOf(component).IsNil() && component != nil {
			util.Log.Infof("publishing %s logs", component.ComponentName())
			err := component.Publish(endpoint, component)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func createETCDInput(clusterName string) *input.FileInput {
	if _, err := os.Stat("k8s/containerlogs/etcd"); err == nil {
		return input.NewFileInput("etcd", "k8s/containerlogs/etcd", clusterName)
	}
	util.Log.Info("etcd log is missing, skipping")
	return nil
}

func createKubeAPIInput(clusterName string) *input.FileInput {
	if _, err := os.Stat("k8s/containerlogs/kube-apiserver"); err == nil {
		return input.NewFileInput("kube-apiserver", "k8s/containerlogs/kube-apiserver", clusterName)
	}
	util.Log.Info("kube-apiserver log is missing, skipping")
	return nil
}

func createKubeletInput(clusterName string) *input.FileInput {
	if _, err := os.Stat("k8s/containerlogs/kubelet"); err == nil {
		return input.NewFileInput("kubelet", "k8s/containerlogs/kubelet", clusterName)
	}
	util.Log.Info("kubelet log is missing, skipping")
	return nil
}

func createKubeControllerManagerInput(clusterName string) *input.FileInput {
	if _, err := os.Stat("k8s/containerlogs/kube-controller-manager"); err == nil {
		return input.NewFileInput("kube-controller-manager", "k8s/containerlogs/kube-controller-manager", clusterName)
	}
	util.Log.Info("kube-controller-manager log is missing, skipping")
	return nil
}

func createKubeSchedulerInput(clusterName string) *input.FileInput {
	if _, err := os.Stat("k8s/containerlogs/kube-scheduler"); err == nil {
		return input.NewFileInput("kube-scheduler", "k8s/containerlogs/kube-scheduler", clusterName)
	}
	util.Log.Info("kube-scheduler log is missing, skipping")
	return nil
}

func createKubeProxyInput(clusterName string) *input.FileInput {
	if _, err := os.Stat("k8s/containerlogs/kube-proxy"); err == nil {
		return input.NewFileInput("kube-proxy", "k8s/containerlogs/kube-proxy", clusterName)
	}
	util.Log.Info("kube-proxy log is missing, skipping")
	return nil
}
