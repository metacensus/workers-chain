# Certificate Authority Configuration
# This chart deploys Hyperledger Fabric CA server

# Organization identifier (e.g., "org1", "acme-corp")
org: "{{.Org}}"

# Certificate type and CA ID for naming
certType: "{{.CertType}}"
caId: "{{.CaId}}"

# Container image configuration
image:
  repository: {{.Image.Repository}}
  tag: "{{.Image.Tag}}"
  pullPolicy: {{.Image.PullPolicy}}

# Number of replicas (typically 1 for root CA, can scale for intermediate)
replicas: {{.Replicas}}

# Database configuration (connection to certificate store)
database:
  host: "{{.Database.Host}}"
  port: {{.Database.Port}}
  name: "{{.Database.Name}}"

# Vault configuration for CSI driver
vault:
  roleName: "{{.Vault.RoleName}}"
  # Paths for different secret types
  # Signing path: CA's signing keypair + certificate (for issuing certs)
  signingPath: "{{.Vault.SigningPath}}"
  # TLS path: CA's TLS keypair + certificate (for HTTPS endpoint)
  tlsPath: "{{.Vault.TlsPath}}"
  # Bootstrap admin credentials (username/password)
  adminCredentialsPath: "{{.Vault.AdminCredentialsPath}}"
  # Database credentials (username/password)
  dbCredentialsPath: "{{.Vault.DbCredentialsPath}}"

# Service configuration
service:
  type: {{.Service.Type}}
  port: {{.Service.Port}}

# CA Server configuration
ca:
  # Common name for the CA
  commonName: "{{.CA.CommonName}}"
  # Organizational unit
  organizationalUnit: "{{.CA.OrganizationalUnit}}"

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
