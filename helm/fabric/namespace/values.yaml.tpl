# Fabric namespace configuration
namespace:
  name: {{.NamespaceName}}

# Common labels to apply to all resources
commonLabels:
  app.kubernetes.io/managed-by: {{.CommonLabelsAppManagedBy}}
  app.kubernetes.io/part-of: {{.CommonLabelsAppPartOf}}

