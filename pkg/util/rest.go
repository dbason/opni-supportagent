package util

import (
	"bytes"
	"log"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	yamlutil "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
)

// ForEachResource will call the given callback function for each
// Kubernetes resource in the embedded string. See the manifests package for
// more details.
// This function will not abort if the callback returns an error, rather it
// will collect all errors that have been returned and return them all at
// once.
func ForEachResource(
	clientConfig *rest.Config,
	embeddedString string,
	callback func(dynamic.ResourceInterface, *unstructured.Unstructured) error,
) (errors []string) {
	errors = []string{}

	decodingSerializer := yaml.NewDecodingSerializer(
		unstructured.UnstructuredJSONScheme)
	decoder := yamlutil.NewYAMLOrJSONDecoder(
		bytes.NewReader([]byte(embeddedString)), 32)
	dynamicClient := dynamic.NewForConfigOrDie(clientConfig)

	dc, err := discovery.NewDiscoveryClientForConfig(clientConfig)
	if err != nil {
		log.Fatal(err)
	}
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(
		memory.NewMemCacheClient(dc))

	for {
		var rawObj runtime.RawExtension
		if err = decoder.Decode(&rawObj); err != nil {
			break
		}

		if len(rawObj.Raw) == 0 {
			continue
		}

		obj := &unstructured.Unstructured{}
		_, gvk, err := decodingSerializer.Decode(rawObj.Raw, nil, obj)
		if err != nil {
			log.Fatal(err)
		}

		mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
		if err != nil {
			errors = append(errors, err.Error())
			continue
		}

		var dr dynamic.ResourceInterface
		if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
			// namespaced resources should specify the namespace
			dr = dynamicClient.Resource(mapping.Resource).
				Namespace(obj.GetNamespace())
		} else {
			// for cluster-wide resources
			dr = dynamicClient.Resource(mapping.Resource)
		}

		if err := callback(dr, obj); err != nil {
			if !k8serrors.IsAlreadyExists(err) {
				errors = append(errors, err.Error())
			}
		}
	}
	return
}
