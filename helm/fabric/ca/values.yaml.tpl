# Hyperledger Fabric Certificate Authority Configuration
# This chart deploys a SINGLE Fabric CA instance

# Organization context
organization:
  id: "{{.OrgID}}"
  mspId: "{{.MSPMID}}"

# CA component and instance number
# Used to construct unique identity: ca-{number}
# Examples: ca-0, ca-1
component: "ca"
number: {{.Number}}

# CA container image
image:
  repository: {{.ImageRepository}}
  tag: "{{.ImageTag}}"
  pullPolicy: {{.ImagePullPolicy}}
  pullSecrets: {{.ImagePullSecretsYAML}}

# CA service configuration
service:
  type: {{.ServiceType}}
  port: {{.ServicePort}}

# CA resource requests and limits
resources:
  requests:
    cpu: {{.ResourcesRequestsCPU}}
    memory: {{.ResourcesRequestsMemory}}
  limits:
    cpu: {{.ResourcesLimitsCPU}}
    memory: {{.ResourcesLimitsMemory}}

# CA-specific configuration
loggingSpec: "{{.LoggingSpec}}"
readinessProbe:
  httpGet:
    path: {{.ReadinessProbeHTTPGetPath}}
    port: {{.ReadinessProbeHTTPGetPort}}
  initialDelaySeconds: {{.ReadinessProbeInitialDelaySeconds}}
  periodSeconds: {{.ReadinessProbePeriodSeconds}}

# MSP (Membership Service Provider) configuration
msp:
  basePath: {{.MSPBasePath}}
  paths:
    signcerts: {{.MSPPathsSigncerts}}
    keystore: {{.MSPPathsKeystore}}
    cacerts: {{.MSPPathsCacerts}}
  
  # CSI (Vault) volume mounting for certificates
  csi:
    enabled: {{.MSPCSIEnabled}}
    secretProviderClassName: "{{.MSPCSISecretProviderClassName}}"

# Namespace configuration
namespace:
  create: {{.NamespaceCreate}}
  name: "{{.NamespaceName}}"

# Common labels applied to all resources
commonLabels:
  app.kubernetes.io/managed-by: {{.CommonLabelsAppManagedBy}}
  app.kubernetes.io/part-of: {{.CommonLabelsAppPartOf}}

