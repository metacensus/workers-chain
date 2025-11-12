# Hyperledger Fabric Peer Node Configuration
# This chart deploys a SINGLE peer instance with attached CouchDB state database

# Organization context
organization:
  id: "{{.OrgID}}"
  mspId: "{{.MSPMID}}"

# Peer component and instance number
# Used to construct unique identity: peer-{number}
# Examples: peer-0, peer-1, peer-2
component: "peer"
number: {{.Number}}

# Peer container image
image:
  repository: {{.ImageRepository}}
  tag: "{{.ImageTag}}"
  pullPolicy: {{.ImagePullPolicy}}
  pullSecrets: {{.ImagePullSecretsYAML}}

# Peer service configuration
service:
  type: {{.ServiceType}}
  port: {{.ServicePort}}

# Peer resource requests and limits
resources:
  requests:
    cpu: {{.ResourcesRequestsCPU}}
    memory: {{.ResourcesRequestsMemory}}
  limits:
    cpu: {{.ResourcesLimitsCPU}}
    memory: {{.ResourcesLimitsMemory}}

# Peer-specific configuration
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

# CouchDB state database (paired with this peer)
couchdb:
  enabled: {{.CouchdbEnabled}}
  image:
    repository: {{.CouchdbImageRepository}}
    tag: "{{.CouchdbImageTag}}"
  service:
    type: {{.CouchdbServiceType}}
    port: {{.CouchdbServicePort}}
  # NOTE: Credentials MUST be generated and stored in Vault
  # These are fetched from Vault at runtime, do NOT set here
  credentials:
    username: "{{.CouchdbUsername}}"
    password: "{{.CouchdbPassword}}"
  resources:
    requests:
      cpu: {{.CouchdbResourcesRequestsCPU}}
      memory: {{.CouchdbResourcesRequestsMemory}}
    limits:
      cpu: {{.CouchdbResourcesLimitsCPU}}
      memory: {{.CouchdbResourcesLimitsMemory}}
  readinessProbe:
    httpGet:
      path: {{.CouchdbReadinessProbeHTTPGetPath}}
      port: {{.CouchdbReadinessProbeHTTPGetPort}}
    initialDelaySeconds: {{.CouchdbReadinessProbeInitialDelaySeconds}}
    periodSeconds: {{.CouchdbReadinessProbePeriodSeconds}}

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

