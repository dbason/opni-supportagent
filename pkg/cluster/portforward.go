package cluster

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/dbason/opni-supportagent/pkg/util"
	"github.com/phayes/freeport"
	opniv1beta1 "github.com/rancher/opni/apis/v1beta1"
	"github.com/rancher/opni/pkg/resources"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	payloadReceiverPort = 80
	kibanaPort          = 5601
)

var (
	PayloadReceiverForwardedPort int
	KibanaForwardedPort          int
)

func PayloadReceiverPort(ctx context.Context) (err error) {
	serviceName := opniv1beta1.PayloadReceiverService.ServiceName()
	svc := &corev1.Service{}
	err = util.K8sClient.Get(ctx, types.NamespacedName{
		Name:      serviceName,
		Namespace: opniNamespace,
	}, svc)
	if err != nil {
		return
	}

	pods := corev1.PodList{}
	err = util.K8sClient.List(ctx, &pods, client.MatchingLabels(svc.Spec.Selector))
	if err != nil {
		return
	}

	util.Log.Info("creating payload receiver port forward")
	forwarder, readyCh, forwardedPort, err := newForwarder(ctx, pods, payloadReceiverPort)
	if err != nil {
		return
	}
	PayloadReceiverForwardedPort = forwardedPort

	errChan := make(chan error)
	go func() {
		errChan <- forwarder.ForwardPorts()
	}()

	util.Log.Info("waiting for port forward to be ready")
	for {
		select {
		case err = <-errChan:
			if err != nil {
				return
			}
		case <-readyCh:
			return
		default:
			time.Sleep(5 * time.Second)
		}
	}
}

func KibanaPort(ctx context.Context) (err error) {
	labels := resources.NewElasticLabels().
		WithRole(opniv1beta1.ElasticKibanaRole)
	pods := corev1.PodList{}
	err = util.K8sClient.List(ctx, &pods, client.MatchingLabels(labels))
	if err != nil {
		return
	}

	util.Log.Info("creating kibana port forward")
	forwarder, readyCh, forwardedPort, err := newForwarder(ctx, pods, kibanaPort)
	if err != nil {
		return
	}
	KibanaForwardedPort = forwardedPort

	errChan := make(chan error)
	go func() {
		errChan <- forwarder.ForwardPorts()
	}()

	util.Log.Info("waiting for port forward to be ready")
	for {
		select {
		case err = <-errChan:
			if err != nil {
				return
			}
		case <-readyCh:
			return
		default:
			time.Sleep(5 * time.Second)
		}
	}
}

func newForwarder(
	ctx context.Context,
	pods corev1.PodList,
	podPort int,
) (
	forwarder *portforward.PortForwarder,
	readyCh chan struct{},
	forwardedPort int,
	err error,
) {
	forwardedPort, err = freeport.GetFreePort()
	if err != nil {
		return
	}

	readyCh = make(chan struct{})
	transport, upgrader, err := spdy.RoundTripperFor(util.RestConfig)
	if err != nil {
		return
	}
	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, http.MethodPost,
		&url.URL{
			Scheme: "https",
			Path: fmt.Sprintf("/api/v1/namespaces/%s/pods/%s/portforward",
				pods.Items[0].Namespace, pods.Items[0].Name),
			Host: strings.TrimLeft(util.RestConfig.Host, "htps:/"),
		})
	forwarder, err = portforward.New(dialer, []string{
		fmt.Sprintf("%d:%d", forwardedPort, podPort),
	}, ctx.Done(), readyCh, nil, io.Discard)
	if err != nil {
		return
	}
	return
}
