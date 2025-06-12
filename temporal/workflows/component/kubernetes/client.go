package kubernetes

import (
	"fmt"

	"go.temporal.io/sdk/temporal"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
)

type Wrapper struct {
	dynamicClient   dynamic.Interface
	discoveryClient discovery.DiscoveryInterface
}

func NewWrapper(config *rest.Config) (*Wrapper, error) {

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	discoveryClient, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return nil, err
	}

	return &Wrapper{
		dynamicClient:   dynamicClient,
		discoveryClient: discoveryClient,
	}, nil
}

func (w *Wrapper) GetResourceInterface(obj *unstructured.Unstructured) (dynamic.ResourceInterface, error) {
	gvk := obj.GroupVersionKind()

	mapping, err := w.getMapping(gvk)
	if err != nil {
		return nil, err
	}

	// Get the resource interface
	var resourceInterface dynamic.ResourceInterface
	if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
		// Namespaced resource
		ns := obj.GetNamespace()
		resourceInterface = w.dynamicClient.Resource(mapping.Resource).Namespace(ns)
	} else {
		// Cluster-scoped resource
		resourceInterface = w.dynamicClient.Resource(mapping.Resource)
	}

	return resourceInterface, nil
}

func (w *Wrapper) getMapping(gvk schema.GroupVersionKind) (*meta.RESTMapping, error) {
	// Create REST mapper for resource discovery
	gr, err := restmapper.GetAPIGroupResources(w.discoveryClient)
	if err != nil {
		return nil, temporal.NewApplicationError("failed to get API resources", "K8S_DISCOVERY_ERROR", err)
	}

	mapper := restmapper.NewDiscoveryRESTMapper(gr)

	mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return nil, fmt.Errorf("failed to get resource mapping: %w", err)
	}

	return mapping, nil
}
