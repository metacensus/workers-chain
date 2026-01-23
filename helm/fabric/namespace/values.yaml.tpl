# Fabric namespace configuration
name: {{.Name}}

# isOperator indicates whether this is the operator namespace
# When true, network policies are created to allow cross-namespace
# access from fabric org namespaces to TLS CA services
isOperator: {{.IsOperator}}

commonLabels:
  managedBy: {{.CommonLabels.ManagedBy}}
  partOf: {{.CommonLabels.PartOf}}

