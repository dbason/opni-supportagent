package cluster

import (
	"context"
	"fmt"

	"github.com/dbason/opni-supportagent/pkg/util"
	"github.com/phayes/freeport"
	"github.com/rancher/k3d/v5/pkg/client"
	"github.com/rancher/k3d/v5/pkg/config"
	k3dv1alpha3 "github.com/rancher/k3d/v5/pkg/config/v1alpha3"
	"github.com/rancher/k3d/v5/pkg/runtimes"
)

const (
	clusterName = "opni-support"
)

func CreateCluster(ctx context.Context) error {
	freePort, err := freeport.GetFreePort()
	if err != nil {
		return err
	}

	simpleConfig := k3dv1alpha3.SimpleConfig{
		Name:  clusterName,
		Image: "rancher/k3s:latest",
		ExposeAPI: k3dv1alpha3.SimpleExposureOpts{
			HostIP:   "127.0.0.1",
			HostPort: fmt.Sprint(freePort),
		},
		Servers: 1,
		Agents:  1,
		Volumes: []k3dv1alpha3.VolumeWithNodeFilters{
			{
				Volume: "/etc/os-release:/etc/os-release",
			},
		},
		Options: k3dv1alpha3.SimpleConfigOptions{
			K3dOptions: k3dv1alpha3.SimpleConfigOptionsK3d{
				DisableLoadbalancer: true,
			},
		},
	}

	conf, err := config.TransformSimpleToClusterConfig(ctx, runtimes.Docker, simpleConfig)
	if err != nil {
		return err
	}
	conf, err = config.ProcessClusterConfig(*conf)
	if err != nil {
		return err
	}

	err = config.ValidateClusterConfig(ctx, runtimes.Docker, *conf)
	if err != nil {
		return err
	}

	if _, err := client.ClusterGet(ctx, runtimes.SelectedRuntime, &conf.Cluster); err == nil {
		util.Log.Info("k3d cluster exists, skipping create")
	} else {
		err = client.ClusterRun(ctx, runtimes.Docker, conf)
		if err != nil {
			return err
		}
	}

	kubeconfig, err := client.KubeconfigGet(ctx, runtimes.Docker, &conf.Cluster)
	if err != nil {
		return err
	}

	util.LoadProvidedConfig(kubeconfig)

	return nil
}
