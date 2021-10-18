package cluster

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/dbason/opni-supportagent/pkg/manifests"
	"github.com/dbason/opni-supportagent/pkg/util"
	opniv1beta1 "github.com/rancher/opni/apis/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/dynamic"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	opniNamespace = "opni"
)

var (
	certmanagerLabels = map[string]string{
		"app.kubernetes.io/instance": "cert-manager",
	}
	namespace = corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: opniNamespace,
		},
	}

	controlplaneModel = opniv1beta1.PretrainedModel{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "controlplane",
			Namespace: opniNamespace,
		},
		Spec: opniv1beta1.PretrainedModelSpec{
			ModelSource: opniv1beta1.ModelSource{
				HTTP: &opniv1beta1.HTTPSource{
					URL: "https://opni-public.s3.us-east-2.amazonaws.com/pretrain-models/control-plane-model-v0.1.2.zip",
				},
			},
			Hyperparameters: map[string]intstr.IntOrString{
				"modelThreshold": intstr.FromString("0.6"),
				"minLogTokens":   intstr.FromInt(4),
				"isControlPlane": intstr.FromString("true"),
			},
		},
	}

	localPassword = corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "local-password",
			Namespace: opniNamespace,
		},
		Data: map[string][]byte{
			"password": []byte("local"),
		},
	}

	opniCluster = opniv1beta1.OpniCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cluster",
			Namespace: opniNamespace,
		},
		Spec: opniv1beta1.OpniClusterSpec{
			Version:            "v0.2.0",
			DeployLogCollector: pointer.BoolPtr(false),
			Services: opniv1beta1.ServicesSpec{
				GPUController: opniv1beta1.GPUControllerServiceSpec{
					Enabled: pointer.BoolPtr(false),
				},
				Preprocessing: opniv1beta1.PreprocessingServiceSpec{
					ImageSpec: opniv1beta1.ImageSpec{
						Image: pointer.StringPtr("quay.io/dbason/opni-preprocessing-service:dev"),
					},
				},
				Metrics: opniv1beta1.MetricsServiceSpec{
					Enabled: pointer.BoolPtr(false),
				},
				Insights: opniv1beta1.InsightsServiceSpec{
					Enabled: pointer.BoolPtr(false),
				},
				UI: opniv1beta1.UIServiceSpec{
					Enabled: pointer.BoolPtr(false),
				},
			},
			Elastic: opniv1beta1.ElasticSpec{
				Version: "1.13.2",
				AdminPasswordFrom: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "local-password",
					},
					Key: "password",
				},
			},
			S3: opniv1beta1.S3Spec{
				Internal: &opniv1beta1.InternalSpec{},
			},
		},
	}
)

func DeployCertManager(ctx context.Context) error {
	util.Log.Info("Installing cert manager as prerequisite")
	msgs := util.ForEachResource(
		util.RestConfig,
		manifests.CertManagerYaml,
		func(dr dynamic.ResourceInterface, obj *unstructured.Unstructured) error {
			_, err := dr.Create(
				ctx,
				obj,
				metav1.CreateOptions{},
			)
			return err
		},
	)
	if len(msgs) > 0 {
		for _, msg := range msgs {
			util.Log.Error(msg)
		}
		return errors.New("failed to install cert manager")
	}

	// Wait for the cert manager deployments to become ready
	deployments := appsv1.DeploymentList{}
	err := util.K8sClient.List(ctx, &deployments, client.MatchingLabels(certmanagerLabels))
	if err != nil {
		return err
	}
	wg := &sync.WaitGroup{}

	util.Log.Info("waiting for cert manager to become ready")
	for _, deployment := range deployments.Items {
		wg.Add(1)
		go func(deployment appsv1.Deployment) {
			defer wg.Done()
			ready := false
			for !ready {
				time.Sleep(5 * time.Second)
				util.K8sClient.Get(ctx, client.ObjectKeyFromObject(&deployment), &deployment)
				ready = deployment.Status.AvailableReplicas > 0
			}
		}(deployment)
	}
	wg.Wait()

	return nil
}

func DeployOpniController(ctx context.Context) error {
	util.Log.Info("Installing opni controller")
	msgs := util.ForEachResource(
		util.RestConfig,
		manifests.OperatorYaml,
		func(dr dynamic.ResourceInterface, obj *unstructured.Unstructured) error {
			_, err := dr.Create(
				ctx,
				obj,
				metav1.CreateOptions{},
			)
			return err
		},
	)
	if len(msgs) > 0 {
		for _, msg := range msgs {
			util.Log.Error(msg)
		}
		return errors.New("failed to install opni controller")
	}

	util.Log.Info("waiting for opni controller to become ready")
	ready := false
	for !ready {
		time.Sleep(5 * time.Second)
		deployment := appsv1.Deployment{}
		err := util.K8sClient.Get(ctx, types.NamespacedName{
			Name:      "opni-controller-manager",
			Namespace: "opni-system",
		}, &deployment)
		if err != nil {
			util.Log.Error(err)
		}
		ready = deployment.Status.AvailableReplicas > 0
	}
	return nil
}

func DeployOpni(ctx context.Context) error {
	ready := false
	err := util.K8sClient.Create(ctx, &namespace)
	if err != nil {
		if !k8serrors.IsAlreadyExists(err) {
			return err
		}
	}

	err = util.K8sClient.Create(ctx, &localPassword)
	if err != nil {
		if !k8serrors.IsAlreadyExists(err) {
			return err
		}
	}

	err = util.K8sClient.Create(ctx, &controlplaneModel)
	if err != nil {
		if !k8serrors.IsAlreadyExists(err) {
			return err
		}
	}

	// Check for existing opnicluster in the namespace
	opniclusters := &opniv1beta1.OpniClusterList{}
	util.K8sClient.List(ctx, opniclusters, client.InNamespace(opniNamespace))
	if len(opniclusters.Items) == 0 {
		err = util.K8sClient.Create(ctx, &opniCluster)
		if err != nil {
			return err

		}
	} else {
		util.Log.Warn("already an opnicluster in the namespace, skipping create")
	}

	util.Log.Info("waiting for opnicluster to become ready")
	for !ready {
		time.Sleep(5 * time.Second)

		cluster := &opniv1beta1.OpniCluster{}
		err = util.K8sClient.Get(ctx, client.ObjectKeyFromObject(&opniCluster), cluster)
		if err != nil {
			util.Log.Error(err)
		}
		ready = cluster.Status.State == "Ready" && cluster.Status.IndexState == "Ready"
	}

	return nil
}
