package publish

import (
	"context"
	"os"
	"reflect"

	"github.com/dbason/opni-supportagent/pkg/input"
	"github.com/dbason/opni-supportagent/pkg/util"
)

type rkeShipper struct {
	ctx         context.Context
	endpoint    string
	clusterName string
	nodeName    string
	username    string
	password    string
}

func ShipRKEControlPlane(
	ctx context.Context,
	endpoint string,
	clusterName string,
	nodeName string,
	username string,
	password string,
) error {
	shipper := rkeShipper{
		ctx:         ctx,
		endpoint:    endpoint,
		clusterName: clusterName,
		nodeName:    nodeName,
		username:    username,
		password:    password,
	}

	for _, component := range []*input.OpensearchInput{
		shipper.createETCDInput(),
		shipper.createKubeAPIInput(),
		shipper.createKubeletInput(),
		shipper.createKubeControllerManagerInput(),
		shipper.createKubeSchedulerInput(),
		shipper.createKubeProxyInput(),
	} {
		if !reflect.ValueOf(component).IsNil() && component != nil {
			util.Log.Infof("publishing %s logs", component.ComponentName())
			_, _, err := component.Publish(&input.DefaultParser{}, input.LogTypeControlplane)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (s rkeShipper) createETCDInput() *input.OpensearchInput {
	if _, err := os.Stat("k8s/containerlogs/etcd"); err == nil {
		os, err := input.NewOpensearchInput(s.ctx, s.endpoint, s.username, s.password, input.OpensearchConfig{
			ClusterID: s.clusterName,
			NodeName:  s.nodeName,
			Component: "etcd",
			Paths:     []string{"k8s/containerlogs/etcd"},
		})
		if err != nil {
			util.Log.Errorf("unable to create etcd shipper: %s", err)
			return nil
		}
		return os
	}
	util.Log.Info("etcd log is missing, skipping")
	return nil
}

func (s rkeShipper) createKubeAPIInput() *input.OpensearchInput {
	if _, err := os.Stat("k8s/containerlogs/kube-apiserver"); err == nil {
		os, err := input.NewOpensearchInput(s.ctx, s.endpoint, s.username, s.password, input.OpensearchConfig{
			ClusterID: s.clusterName,
			NodeName:  s.nodeName,
			Component: "kube-apiserver",
			Paths:     []string{"k8s/containerlogs/kube-apiserver"},
		})
		if err != nil {
			util.Log.Errorf("unable to create api shipper: %s", err)
			return nil
		}
		return os
	}
	util.Log.Info("kube-apiserver log is missing, skipping")
	return nil
}

func (s rkeShipper) createKubeletInput() *input.OpensearchInput {
	if _, err := os.Stat("k8s/containerlogs/kubelet"); err == nil {
		os, err := input.NewOpensearchInput(s.ctx, s.endpoint, s.username, s.password, input.OpensearchConfig{
			ClusterID: s.clusterName,
			NodeName:  s.nodeName,
			Component: "kubelet",
			Paths:     []string{"k8s/containerlogs/kubelet"},
		})
		if err != nil {
			util.Log.Errorf("unable to create kubelet shipper: %s", err)
			return nil
		}
		return os
	}
	util.Log.Info("kubelet log is missing, skipping")
	return nil
}

func (s rkeShipper) createKubeControllerManagerInput() *input.OpensearchInput {
	if _, err := os.Stat("k8s/containerlogs/kube-controller-manager"); err == nil {
		os, err := input.NewOpensearchInput(s.ctx, s.endpoint, s.username, s.password, input.OpensearchConfig{
			ClusterID: s.clusterName,
			NodeName:  s.nodeName,
			Component: "kube-controller-manager",
			Paths:     []string{"k8s/containerlogs/kube-controller-manager"},
		})
		if err != nil {
			util.Log.Errorf("unable to create controller manager shipper: %s", err)
			return nil
		}
		return os
	}
	util.Log.Info("kube-controller-manager log is missing, skipping")
	return nil
}

func (s rkeShipper) createKubeSchedulerInput() *input.OpensearchInput {
	if _, err := os.Stat("k8s/containerlogs/kube-scheduler"); err == nil {
		os, err := input.NewOpensearchInput(s.ctx, s.endpoint, s.username, s.password, input.OpensearchConfig{
			ClusterID: s.clusterName,
			NodeName:  s.nodeName,
			Component: "kube-scheduler",
			Paths:     []string{"k8s/containerlogs/kube-scheduler"},
		})
		if err != nil {
			util.Log.Errorf("unable to create kube-scheduler shipper: %s", err)
			return nil
		}
		return os
	}
	util.Log.Info("kube-scheduler log is missing, skipping")
	return nil
}

func (s rkeShipper) createKubeProxyInput() *input.OpensearchInput {
	if _, err := os.Stat("k8s/containerlogs/kube-proxy"); err == nil {
		os, err := input.NewOpensearchInput(s.ctx, s.endpoint, s.username, s.password, input.OpensearchConfig{
			ClusterID: s.clusterName,
			NodeName:  s.nodeName,
			Component: "kube-proxy",
			Paths:     []string{"k8s/containerlogs/kube-proxy"},
		})
		if err != nil {
			util.Log.Errorf("unable to create kube-proxy shipper: %s", err)
			return nil
		}
		return os
	}
	util.Log.Info("kube-proxy log is missing, skipping")
	return nil
}
