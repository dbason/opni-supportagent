package rke

import (
	"os"
	"reflect"

	"github.com/dbason/opni-supportagent/pkg/input"
	"github.com/dbason/opni-supportagent/pkg/util"
)

func ShipRKEControlPlane(endpoint string) error {
	for _, component := range []input.ComponentInput{
		createETCDInput(),
		createKubeAPIInput(),
		createKubeletInput(),
		createKubeControllerManagerInput(),
		createKubeSchedulerInput(),
		createKubeProxyInput(),
	} {
		if !reflect.ValueOf(component).IsNil() && component != nil {
			err := component.Publish(endpoint)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func createETCDInput() *input.FileInput {
	if _, err := os.Stat("k8s/containerlogs/etcd"); err == nil {
		return input.NewFileInput("etcd", "k8s/containerlogs/etcd")
	}
	util.Log.Info("etcd log is missing, skipping")
	return nil
}

func createKubeAPIInput() *input.FileInput {
	if _, err := os.Stat("k8s/containerlogs/kube-apiserver"); err == nil {
		return input.NewFileInput("kube-apiserver", "k8s/containerlogs/kube-apiserver")
	}
	util.Log.Info("kube-apiserver log is missing, skipping")
	return nil
}

func createKubeletInput() *input.FileInput {
	if _, err := os.Stat("k8s/containerlogs/kubelet"); err == nil {
		return input.NewFileInput("kubelet", "k8s/containerlogs/kubelet")
	}
	util.Log.Info("kubelet log is missing, skipping")
	return nil
}

func createKubeControllerManagerInput() *input.FileInput {
	if _, err := os.Stat("k8s/containerlogs/kube-controller-manager"); err == nil {
		return input.NewFileInput("kube-controller-manager", "k8s/containerlogs/kube-controller-manager")
	}
	util.Log.Info("kube-controller-manager log is missing, skipping")
	return nil
}

func createKubeSchedulerInput() *input.FileInput {
	if _, err := os.Stat("k8s/containerlogs/kube-scheduler"); err == nil {
		return input.NewFileInput("kube-scheduler", "k8s/containerlogs/kube-scheduler")
	}
	util.Log.Info("kube-scheduler log is missing, skipping")
	return nil
}

func createKubeProxyInput() *input.FileInput {
	if _, err := os.Stat("k8s/containerlogs/kube-proxy"); err == nil {
		return input.NewFileInput("kube-proxy", "k8s/containerlogs/kube-proxy")
	}
	util.Log.Info("kube-proxy log is missing, skipping")
	return nil
}
