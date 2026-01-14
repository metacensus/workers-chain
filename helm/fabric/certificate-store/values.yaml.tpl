# Certificate Store Configuration
# This chart deploys PostgreSQL for Hyperledger Fabric certificate storage

# Organization identifier (e.g., "org1", "acme-corp")
org: "{{.Org}}"

# Container image configuration
image:
  repository: {{.Image.Repository}}
  tag: "{{.Image.Tag}}"
  pullPolicy: {{.Image.PullPolicy}}

# Number of replicas (1 for now, can scale later)
replicas: {{.Replicas}}

# Database configuration
database:
  name: {{.Database.Name}}

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



