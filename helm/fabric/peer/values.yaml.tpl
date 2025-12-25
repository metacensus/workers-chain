# Hyperledger Fabric Peer Node Configuration
# This chart deploys a SINGLE peer instance with attached CouchDB state database

# Organization context
organization:
  id: "{{.Organization.ID}}"
  mspId: "{{.Organization.MSPId}}"

# Peer component and instance number
# Used to construct unique identity: peer-{number}
# Examples: peer-0, peer-1, peer-2
component: "{{.Component}}"
number: {{.Number}}

# Peer container image
image:
  repository: {{.Image.Repository}}
  tag: "{{.Image.Tag}}"
  pullPolicy: {{.Image.PullPolicy}}
  pullSecrets: []

# Peer service configuration
service:
  type: {{.Service.Type}}
  port: {{.Service.Port}}

# Peer resource requests and limits
resources:
  requests:
    cpu: {{.Resources.Requests.CPU}}
    memory: {{.Resources.Requests.Memory}}
  limits:
    cpu: {{.Resources.Limits.CPU}}
    memory: {{.Resources.Limits.Memory}}

# Peer-specific configuration
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

# CouchDB state database (paired with this peer)
couchdb:
  enabled: {{.CouchDB.Enabled}}
  image:
    repository: {{.CouchDB.Image.Repository}}
    tag: "{{.CouchDB.Image.Tag}}"
  service:
    type: {{.CouchDB.Service.Type}}
    port: {{.CouchDB.Service.Port}}
  # NOTE: Credentials MUST be generated and stored in Vault
  # These are fetched from Vault at runtime, do NOT set here
  credentials:
    username: "{{.CouchDB.Credentials.Username}}"
    password: "{{.CouchDB.Credentials.Password}}"
  resources:
    requests:
      cpu: {{.CouchDB.Resources.Requests.CPU}}
      memory: {{.CouchDB.Resources.Requests.Memory}}
    limits:
      cpu: {{.CouchDB.Resources.Limits.CPU}}
      memory: {{.CouchDB.Resources.Limits.Memory}}
  readinessProbe:
    httpGet:
      path: {{.CouchDB.ReadinessProbe.HTTPGet.Path}}
      port: {{.CouchDB.ReadinessProbe.HTTPGet.Port}}
    initialDelaySeconds: {{.CouchDB.ReadinessProbe.InitialDelaySeconds}}
    periodSeconds: {{.CouchDB.ReadinessProbe.PeriodSeconds}}

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

