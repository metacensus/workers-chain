# Fabric Worker Configuration
# This chart deploys a Temporal worker for managing Hyperledger Fabric resources.
# It serves both the shared fabric worker (operator-owned) and per-org shard workers.

# Operator identifier (e.g., "metacensus")
# The operator owns shared infrastructure like TLS certificate authorities
operator: "{{.Operator}}"

# Owner identifier used for resource naming
# (org ID for shard workers, operator ID for the shared fabric worker)
org: "{{.Org}}"

# Organization identifier passed as ORG_ID to the worker container.
# Empty for the shared fabric worker, which owns no organization.
orgId: "{{.OrgId}}"

# Kubernetes namespace the worker is deployed into
# ("fabric" for the shared fabric worker, "fabric-shard-{org}" for shard workers)
namespace: "{{.Namespace}}"

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
  tls:
    secretName: "{{.Temporal.TLS.SecretName}}"
    caPath: "{{.Temporal.TLS.CAPath}}"
    certPath: "{{.Temporal.TLS.CertPath}}"
    keyPath: "{{.Temporal.TLS.KeyPath}}"
    serverName: "{{.Temporal.TLS.ServerName}}"

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
