# Fabric Peer Configuration
# This chart deploys a Hyperledger Fabric peer node

# Organization identifier (e.g., "org1", "acme-corp")
org: "{{.Org}}"

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
vault:
  roleName: "{{.Vault.RoleName}}"
  mspSigningPath: "{{.Vault.MspSigningPath}}"
  mspRootCaCertPath: "{{.Vault.MspRootCaCertPath}}"
  mspIntCaCertPath: "{{.Vault.MspIntCaCertPath}}"
  tlsPath: "{{.Vault.TlsPath}}"
  tlsRootCaCertPath: "{{.Vault.TlsRootCaCertPath}}"
  tlsIntCaCertPath: "{{.Vault.TlsIntCaCertPath}}"
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

# Liveness probe configuration
livenessProbe:
  initialDelaySeconds: {{.LivenessProbe.InitialDelaySeconds}}
  periodSeconds: {{.LivenessProbe.PeriodSeconds}}
  timeoutSeconds: {{.LivenessProbe.TimeoutSeconds}}

# Readiness probe configuration
readinessProbe:
  initialDelaySeconds: {{.ReadinessProbe.InitialDelaySeconds}}
  periodSeconds: {{.ReadinessProbe.PeriodSeconds}}
  timeoutSeconds: {{.ReadinessProbe.TimeoutSeconds}}

# Common labels applied to all resources
commonLabels:
  managedBy: {{.CommonLabels.ManagedBy}}
  partOf: {{.CommonLabels.PartOf}}

