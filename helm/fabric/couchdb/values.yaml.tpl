# CouchDB State Database Configuration
# This chart deploys a SINGLE CouchDB instance for state database storage
# Typically paired with a peer via pod co-location in the peer chart
# Can also be deployed standalone if needed

# Organization context
organization:
  id: "{{.OrgID}}"

# CouchDB component and instance number
# Used to construct unique identity: couchdb-{number}
# Examples: couchdb-0, couchdb-1
component: "couchdb"
number: {{.Number}}

# CouchDB container image
image:
  repository: {{.ImageRepository}}
  tag: "{{.ImageTag}}"
  pullPolicy: {{.ImagePullPolicy}}
  pullSecrets: {{.ImagePullSecretsYAML}}

# CouchDB service configuration
service:
  type: {{.ServiceType}}
  port: {{.ServicePort}}

# CouchDB credentials (for initial setup)
# NOTE: Credentials MUST be generated and stored in Vault
# These are fetched from Vault at runtime, do NOT set here
credentials:
  username: "{{.Username}}"
  password: "{{.Password}}"

# CouchDB resource requests and limits
resources:
  requests:
    cpu: {{.ResourcesRequestsCPU}}
    memory: {{.ResourcesRequestsMemory}}
  limits:
    cpu: {{.ResourcesLimitsCPU}}
    memory: {{.ResourcesLimitsMemory}}

# CouchDB health check configuration
readinessProbe:
  httpGet:
    path: {{.ReadinessProbeHTTPGetPath}}
    port: {{.ReadinessProbeHTTPGetPort}}
  initialDelaySeconds: {{.ReadinessProbeInitialDelaySeconds}}
  periodSeconds: {{.ReadinessProbePeriodSeconds}}

# Namespace configuration
namespace:
  create: {{.NamespaceCreate}}
  name: "{{.NamespaceName}}"

# Common labels applied to all resources
commonLabels:
  app.kubernetes.io/managed-by: {{.CommonLabelsAppManagedBy}}
  app.kubernetes.io/part-of: {{.CommonLabelsAppPartOf}}

