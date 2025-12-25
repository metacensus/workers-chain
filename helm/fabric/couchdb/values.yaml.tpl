# CouchDB State Database Configuration
# This chart deploys a SINGLE CouchDB instance for state database storage
# Typically paired with a peer via pod co-location in the peer chart
# Can also be deployed standalone if needed

# Organization context
organization:
  id: "{{.Organization.ID}}"

# CouchDB component and instance number
# Used to construct unique identity: couchdb-{number}
# Examples: couchdb-0, couchdb-1
component: "{{.Component}}"
number: {{.Number}}

# CouchDB container image
image:
  repository: {{.Image.Repository}}
  tag: "{{.Image.Tag}}"
  pullPolicy: {{.Image.PullPolicy}}
  pullSecrets: []

# CouchDB service configuration
service:
  type: {{.Service.Type}}
  port: {{.Service.Port}}

# CouchDB credentials (for initial setup)
# NOTE: Credentials MUST be generated and stored in Vault
# These are fetched from Vault at runtime, do NOT set here
credentials:
  username: "{{.Credentials.Username}}"
  password: "{{.Credentials.Password}}"

# CouchDB resource requests and limits
resources:
  requests:
    cpu: {{.Resources.Requests.CPU}}
    memory: {{.Resources.Requests.Memory}}
  limits:
    cpu: {{.Resources.Limits.CPU}}
    memory: {{.Resources.Limits.Memory}}

# CouchDB health check configuration
readinessProbe:
  httpGet:
    path: {{.ReadinessProbe.HTTPGet.Path}}
    port: {{.ReadinessProbe.HTTPGet.Port}}
  initialDelaySeconds: {{.ReadinessProbe.InitialDelaySeconds}}
  periodSeconds: {{.ReadinessProbe.PeriodSeconds}}

# Namespace configuration
namespace:
  create: {{.Namespace.Create}}
  name: "{{.Namespace.Name}}"

# Common labels applied to all resources
commonLabels:
  managedBy: {{.CommonLabels.ManagedBy}}
  partOf: {{.CommonLabels.PartOf}}

