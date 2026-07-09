# Fabric Peer Configuration
# This chart deploys a Hyperledger Fabric peer node

# Organization identifier (e.g., "org1", "acme-corp")
org: "{{.Org}}"

# Kubernetes namespace for this component
# ("fabric" for operator-owned components, "fabric-shard-{org}" for org-owned)
namespace: "{{.Namespace}}"

# Peer identifier (e.g., "peer0", "peer1")
peerId: "{{.PeerId}}"

# Container image configuration
image:
  repository: {{.Image.Repository}}
  tag: "{{.Image.Tag}}"
  pullPolicy: {{.Image.PullPolicy}}

# Number of replicas (1 per StatefulSet - scale by adding peers)
replicas: {{.Replicas}}

# MSP configuration
msp:
  id: "{{.Msp.Id}}"

# Vault configuration for CSI driver
# Keys (private) and certs (public) live at separate paths.
# CA trust-store paths point at root CA certs; TLS CA paths resolve to the
# operator's shared Vault tree ("fabric/{operator}/..."), org CA paths to the
# org's shard tree ("fabric/shard/{org}/...").
vault:
  roleName: "{{.Vault.RoleName}}"
  mspSigningKeyPath: "{{.Vault.MspSigningKeyPath}}"
  mspSigningCertPath: "{{.Vault.MspSigningCertPath}}"
  mspRootCaCertPath: "{{.Vault.MspRootCaCertPath}}"
  tlsKeyPath: "{{.Vault.TlsKeyPath}}"
  tlsCertPath: "{{.Vault.TlsCertPath}}"
  tlsRootCaCertPath: "{{.Vault.TlsRootCaCertPath}}"
  couchdbPath: "{{.Vault.CouchdbPath}}"

# Service configuration
service:
  type: {{.Service.Type}}
  port: {{.Service.Port}}
  chaincodePort: {{.Service.ChaincodePort}}
  operationsPort: {{.Service.OperationsPort}}

# Gossip protocol configuration
gossip:
  bootstrap: "{{.Gossip.Bootstrap}}"
  externalEndpoint: "{{.Gossip.ExternalEndpoint}}"

# CouchDB state database
couchdb:
  address: "{{.Couchdb.Address}}"

# Persistence configuration
persistence:
  storageClass: "{{.Persistence.StorageClass}}"
  accessModes:
  {{- range .Persistence.AccessModes }}
  - {{.}}
  {{- end }}
  size: {{.Persistence.Size}}

# Logging configuration
logging:
  level: {{.Logging.Level}}

# Resource requests and limits
resources:
  requests:
    cpu: {{.Resources.Requests.CPU}}
    memory: {{.Resources.Requests.Memory}}
  limits:
    cpu: {{.Resources.Limits.CPU}}
    memory: {{.Resources.Limits.Memory}}

# Startup probe configuration - runs first, gives container time to initialize
# Once startup probe succeeds, liveness probe takes over
startupProbe:
  initialDelaySeconds: {{.StartupProbe.InitialDelaySeconds}}
  periodSeconds: {{.StartupProbe.PeriodSeconds}}
  timeoutSeconds: {{.StartupProbe.TimeoutSeconds}}
  failureThreshold: {{.StartupProbe.FailureThreshold}}

# Liveness probe configuration - only runs after startup probe succeeds
livenessProbe:
  initialDelaySeconds: {{.LivenessProbe.InitialDelaySeconds}}
  periodSeconds: {{.LivenessProbe.PeriodSeconds}}
  timeoutSeconds: {{.LivenessProbe.TimeoutSeconds}}
  failureThreshold: {{.LivenessProbe.FailureThreshold}}

# Readiness probe configuration
readinessProbe:
  initialDelaySeconds: {{.ReadinessProbe.InitialDelaySeconds}}
  periodSeconds: {{.ReadinessProbe.PeriodSeconds}}
  timeoutSeconds: {{.ReadinessProbe.TimeoutSeconds}}
  failureThreshold: {{.ReadinessProbe.FailureThreshold}}

# Common labels applied to all resources
commonLabels:
  managedBy: {{.CommonLabels.ManagedBy}}
  partOf: {{.CommonLabels.PartOf}}

