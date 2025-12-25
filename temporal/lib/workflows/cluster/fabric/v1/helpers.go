package cluster

import "fmt"

func ServiceAccountNameFromOrg(org string) string {
	return fmt.Sprintf("fabric-worker-%s", org)
}

func RoleNameFromOrg(org string) string {
	return fmt.Sprintf("fabric-worker-%s", org)
}

func NamespaceNameFromOrg(org string) string {
	return fmt.Sprintf("fabric-%s", org)
}

