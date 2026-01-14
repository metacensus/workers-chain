# Service account configuration
namespace: {{.Namespace}}

serviceAccountName: {{.ServiceAccountName}}

roleName: {{.RoleName}}

commonLabels:
  managedBy: {{.CommonLabels.ManagedBy}}
  partOf: {{.CommonLabels.PartOf}}

