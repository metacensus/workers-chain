# Fabric Worker Configuration
# This chart deploys a Temporal worker for managing Hyperledger Fabric resources

# Organization identifier (e.g., "org1", "acme-corp")
# Used to derive namespace names and service account names
org: "{{.Org}}"

# Container image configuration
image:
  repository: {{.Image.Repository}}
  tag: "{{.Image.Tag}}"
  pullPolicy: {{.Image.PullPolicy}}

# Number of replicas (typically 1 per organization)
replicas: {{.Replicas}}

# Temporal configuration
temporal:
  address: "{{.Temporal.Address}}"
  namespace: "{{.Temporal.Namespace}}"
  taskQueue: "{{.Temporal.TaskQueue}}"

# Vault configuration for secret management
vault:
  address: "{{.Vault.Address}}"
  role: "{{.Vault.Role}}"

# Logging configuration
logLevel: "{{.LogLevel}}"

# Health check configuration
healthCheck:
  port: {{.HealthCheck.Port}}

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
  path: {{.LivenessProbe.Path}}
  initialDelaySeconds: {{.LivenessProbe.InitialDelaySeconds}}
  periodSeconds: {{.LivenessProbe.PeriodSeconds}}

# Readiness probe configuration
readinessProbe:
  path: {{.ReadinessProbe.Path}}
  initialDelaySeconds: {{.ReadinessProbe.InitialDelaySeconds}}
  periodSeconds: {{.ReadinessProbe.PeriodSeconds}}

# Common labels applied to all resources
commonLabels:
  managedBy: {{.CommonLabels.ManagedBy}}
  partOf: {{.CommonLabels.PartOf}}



