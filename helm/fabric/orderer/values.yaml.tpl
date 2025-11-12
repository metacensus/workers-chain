# Hyperledger Fabric Orderer Node Configuration
# This chart deploys a SINGLE orderer instance

# Organization context
organization:
  id: "{{.OrgID}}"
  mspId: "{{.MSPMID}}"

# Orderer component and instance number
# Used to construct unique identity: orderer-{number}
# Examples: orderer-0, orderer-1, orderer-2
component: "orderer"
number: {{.Number}}

# Orderer container image
image:
  repository: {{.ImageRepository}}
  tag: "{{.ImageTag}}"
  pullPolicy: {{.ImagePullPolicy}}
  pullSecrets: {{.ImagePullSecretsYAML}}

# Orderer service configuration
service:
  type: {{.ServiceType}}
  port: {{.ServicePort}}

# Orderer resource requests and limits
resources:
  requests:
    cpu: {{.ResourcesRequestsCPU}}
    memory: {{.ResourcesRequestsMemory}}
  limits:
    cpu: {{.ResourcesLimitsCPU}}
    memory: {{.ResourcesLimitsMemory}}

# Orderer-specific configuration
# MSP ID for the orderer (typically different from organization)
mspId: "{{.OrdererMSPID}}"
loggingSpec: "{{.LoggingSpec}}"
tls:
  enabled: {{.TLSEnabled}}
  clientAuthRequired: {{.TLSClientAuthRequired}}
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

