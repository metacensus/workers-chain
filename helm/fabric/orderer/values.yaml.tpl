# Hyperledger Fabric Orderer Node Configuration
# This chart deploys a SINGLE orderer instance

# Organization context
organization:
  id: "{{.Organization.ID}}"
  mspId: "{{.Organization.MSPId}}"

# Orderer component and instance number
# Used to construct unique identity: orderer-{number}
# Examples: orderer-0, orderer-1, orderer-2
component: "{{.Component}}"
number: {{.Number}}

# Orderer container image
image:
  repository: {{.Image.Repository}}
  tag: "{{.Image.Tag}}"
  pullPolicy: {{.Image.PullPolicy}}
  pullSecrets: []

# Orderer service configuration
service:
  type: {{.Service.Type}}
  port: {{.Service.Port}}

# Orderer resource requests and limits
resources:
  requests:
    cpu: {{.Resources.Requests.CPU}}
    memory: {{.Resources.Requests.Memory}}
  limits:
    cpu: {{.Resources.Limits.CPU}}
    memory: {{.Resources.Limits.Memory}}

# Orderer-specific configuration
# MSP ID for the orderer (typically different from organization)
mspId: "{{.OrdererMSPID}}"
loggingSpec: "{{.LoggingSpec}}"
tls:
  enabled: {{.TLS.Enabled}}
  clientAuthRequired: {{.TLS.ClientAuthRequired}}
readinessProbe:
  httpGet:
    path: {{.ReadinessProbe.HTTPGet.Path}}
    port: {{.ReadinessProbe.HTTPGet.Port}}
  initialDelaySeconds: {{.ReadinessProbe.InitialDelaySeconds}}
  periodSeconds: {{.ReadinessProbe.PeriodSeconds}}

# MSP (Membership Service Provider) configuration
msp:
  basePath: {{.MSP.BasePath}}
  paths:
    signcerts: {{.MSP.Paths.Signcerts}}
    keystore: {{.MSP.Paths.Keystore}}
    cacerts: {{.MSP.Paths.Cacerts}}
  
  # CSI (Vault) volume mounting for certificates
  csi:
    enabled: {{.MSP.CSI.Enabled}}
    secretProviderClassName: "{{.MSP.CSI.SecretProviderClassName}}"

# Namespace configuration
namespace:
  create: {{.Namespace.Create}}
  name: "{{.Namespace.Name}}"

# Common labels applied to all resources
commonLabels:
  managedBy: {{.CommonLabels.ManagedBy}}
  partOf: {{.CommonLabels.PartOf}}

