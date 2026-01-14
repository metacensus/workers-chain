# CouchDB Configuration
# This chart deploys CouchDB for Hyperledger Fabric peer state storage

# Organization identifier (e.g., "org1", "acme-corp")
org: "{{.Org}}"

# Peer identifier (e.g., "peer0", "peer1")
peerId: "{{.PeerId}}"

# Container image configuration
image:
  repository: {{.Image.Repository}}
  tag: "{{.Image.Tag}}"
  pullPolicy: {{.Image.PullPolicy}}

# Number of replicas (1 for single instance)
replicas: {{.Replicas}}

# Vault configuration for CSI driver
vault:
  roleName: "{{.Vault.RoleName}}"
  secretPath: "{{.Vault.SecretPath}}"

# Service configuration
service:
  type: {{.Service.Type}}
  port: {{.Service.Port}}

# Persistence configuration
persistence:
  enabled: {{.Persistence.Enabled}}
  storageClass: "{{.Persistence.StorageClass}}"
  accessModes:
  {{- range .Persistence.AccessModes }}
  - {{.}}
  {{- end }}
  size: {{.Persistence.Size}}
  mountPath: {{.Persistence.MountPath}}

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

