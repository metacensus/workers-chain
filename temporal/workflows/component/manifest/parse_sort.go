package manifest

import (
	"bytes"
	"fmt"
	"io"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/yaml"
)

// parseManifests parses YAML manifests into unstructured objects
func parseManifests(data []byte) ([]*unstructured.Unstructured, error) {
	var resources []*unstructured.Unstructured
	decoder := yaml.NewYAMLOrJSONDecoder(bytes.NewReader(data), 4096)

	for {
		var rawObj runtime.RawExtension
		if err := decoder.Decode(&rawObj); err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("failed to decode manifest: %w", err)
		}

		if len(rawObj.Raw) == 0 {
			continue // Skip empty documents
		}

		obj := &unstructured.Unstructured{}
		if err := yaml.Unmarshal(rawObj.Raw, obj); err != nil {
			return nil, fmt.Errorf("failed to unmarshal manifest: %w", err)
		}

		// Skip empty objects
		if obj.GetKind() == "" {
			continue
		}

		resources = append(resources, obj)
	}

	return resources, nil
}

// sortResourcesByKind sorts resources to apply in the correct order
func sortResourcesByKind(resources []*unstructured.Unstructured) []*unstructured.Unstructured {
	// Priority order for resource types
	priority := map[string]int{
		"Namespace":                1,
		"ResourceQuota":            2,
		"LimitRange":               3,
		"PodSecurityPolicy":        4,
		"NetworkPolicy":            5,
		"ServiceAccount":           6,
		"Secret":                   7,
		"ConfigMap":                8,
		"StorageClass":             9,
		"PersistentVolume":         10,
		"PersistentVolumeClaim":    11,
		"CustomResourceDefinition": 12,
		"ClusterRole":              13,
		"ClusterRoleBinding":       14,
		"Role":                     15,
		"RoleBinding":              16,
		"Service":                  17,
		"DaemonSet":                18,
		"Pod":                      19,
		"ReplicaSet":               20,
		"Deployment":               21,
		"StatefulSet":              22,
		"Job":                      23,
		"CronJob":                  24,
		"Ingress":                  25,
	}

	// Sort resources
	sorted := make([]*unstructured.Unstructured, len(resources))
	copy(sorted, resources)

	// Stable sort to maintain order within same priority
	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			p1, ok1 := priority[sorted[i].GetKind()]
			if !ok1 {
				p1 = 100 // Default priority for unknown kinds
			}
			p2, ok2 := priority[sorted[j].GetKind()]
			if !ok2 {
				p2 = 100
			}
			if p1 > p2 {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	return sorted
}
