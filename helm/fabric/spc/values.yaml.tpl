# Vault CSI SecretProviderClass Configuration
# This chart creates Kubernetes CRDs that configure how the Vault CSI driver
# mounts Fabric certificates into pod filesystems

# Vault server address (must be accessible from inside the cluster)
vault:
  address: "{{.Vault.Address}}"

# Organization name (e.g., "org1", "metacensus")
# Used in Vault path construction: secret/fabric/{organization}/msp/{identity}
# REQUIRED: Must be provided by workflow
organization: "{{.Organization}}"

# Identity name (e.g., "peer0", "orderer0", "admin")
# Used in Vault path construction and resource naming
# REQUIRED: Must be provided by workflow
identity: "{{.Identity}}"

# Kubernetes namespace where the SecretProviderClass will be created
namespace: {{.Namespace}}

# Optional: Custom resource name
# If empty, defaults to "{organization}-{identity}-spc"
resourceName: "{{.ResourceName}}"

# Common labels applied to all resources
commonLabels:
  managedBy: {{.CommonLabels.ManagedBy}}
  partOf: {{.CommonLabels.PartOf}}

