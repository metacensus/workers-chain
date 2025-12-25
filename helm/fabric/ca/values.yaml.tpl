# Hyperledger Fabric Certificate Authority Configuration
# This chart deploys a SINGLE Fabric CA instance

# Organization context
organization:
  id: "{{.Organization.ID}}"
  mspId: "{{.Organization.MSPId}}"

# CA component and instance number
# Used to construct unique identity: ca-{number}
# Examples: ca-0, ca-1
component: "{{.Component}}"
number: {{.Number}}

# CA container image
image:
  repository: {{.Image.Repository}}
  tag: "{{.Image.Tag}}"
  pullPolicy: {{.Image.PullPolicy}}
  pullSecrets: []

# CA service configuration
service:
  type: {{.Service.Type}}
  port: {{.Service.Port}}

# CA resource requests and limits
resources:
  requests:
    cpu: {{.Resources.Requests.CPU}}
    memory: {{.Resources.Requests.Memory}}
  limits:
    cpu: {{.Resources.Limits.CPU}}
    memory: {{.Resources.Limits.Memory}}

# CA-specific configuration
loggingSpec: "{{.LoggingSpec}}"
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

