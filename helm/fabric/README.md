# Hyperledger Fabric Helm Chart

This Helm chart deploys a Hyperledger Fabric network for a single organization on Kubernetes.

## Features

- **Configurable component counts**: Deploy any number of CAs, Peers, and Orderers
- **Automatic CouchDB pairing**: Each peer gets a dedicated CouchDB instance
- **Standard certificate locations**: Follows Hyperledger Fabric MSP conventions
- **TLS enabled**: Support for mutual TLS between components
- **Resource management**: Configurable CPU and memory limits
- **Production-ready**: Health checks, proper labeling, and service discovery

## Prerequisites

- Kubernetes 1.19+
- Helm 3.0+
- Certificate secrets created before deployment (see below)

## Installation

### 1. Create Certificate Secrets

Before installing the chart, create Kubernetes secrets containing the MSP material for each component. Each secret must contain:

- `signcerts/cert.pem` - The node's certificate
- `keystore/key.pem` - The node's private key
- `cacerts/ca.pem` - The CA's certificate

Example:
```bash
kubectl create secret generic fabric-peer-0-certs \
  --from-file=signcerts/cert.pem=./crypto/peer0/signcerts/cert.pem \
  --from-file=keystore/key.pem=./crypto/peer0/keystore/key.pem \
  --from-file=cacerts/ca.pem=./crypto/peer0/cacerts/ca.pem \
  -n fabric-org1
```

### 2. Install the Chart

```bash
# Basic installation with defaults
helm install my-fabric ./helm/fabric -n fabric-org1 --create-namespace

# With custom values
helm install my-fabric ./helm/fabric \
  -n fabric-org1 \
  --create-namespace \
  -f custom-values.yaml
```

### 3. Verify Installation

```bash
# Check pod status
kubectl get pods -n fabric-org1

# View services
kubectl get services -n fabric-org1

# Check deployment notes
helm get notes my-fabric -n fabric-org1
```

## Configuration

The following table lists the key configurable parameters:

| Parameter | Description | Default |
|-----------|-------------|---------|
| `organization.id` | Organization identifier | `org1` |
| `organization.mspId` | MSP ID for the organization | `Org1MSP` |
| `ca.count` | Number of Certificate Authorities | `1` |
| `ca.image.repository` | CA image repository | `hyperledger/fabric-ca` |
| `ca.image.tag` | CA image tag | `1.5.5` |
| `ca.certSecrets` | Array of secret names for CA certificates | `["fabric-ca-0-certs"]` |
| `peer.count` | Number of peers | `2` |
| `peer.image.repository` | Peer image repository | `hyperledger/fabric-peer` |
| `peer.image.tag` | Peer image tag | `2.5.0` |
| `peer.certSecrets` | Array of secret names for peer certificates | `["fabric-peer-0-certs", "fabric-peer-1-certs"]` |
| `peer.tls.enabled` | Enable TLS for peers | `true` |
| `couchdb.image.repository` | CouchDB image repository | `couchdb` |
| `couchdb.image.tag` | CouchDB image tag | `3.3.2` |
| `couchdb.user` | CouchDB username | `admin` |
| `couchdb.password` | CouchDB password | `adminpw` |
| `orderer.count` | Number of orderers | `3` |
| `orderer.image.repository` | Orderer image repository | `hyperledger/fabric-orderer` |
| `orderer.image.tag` | Orderer image tag | `2.5.0` |
| `orderer.mspId` | MSP ID for orderers | `OrdererMSP` |
| `orderer.certSecrets` | Array of secret names for orderer certificates | `["fabric-orderer-0-certs", ...]` |

### Example: Custom Organization

```yaml
# custom-values.yaml
organization:
  id: "metacensus"
  mspId: "MetaCensusMSP"

peer:
  count: 4
  certSecrets:
    - "metacensus-peer-0-certs"
    - "metacensus-peer-1-certs"
    - "metacensus-peer-2-certs"
    - "metacensus-peer-3-certs"
  resources:
    requests:
      cpu: 500m
      memory: 1Gi
    limits:
      cpu: 2000m
      memory: 4Gi

orderer:
  count: 5
  mspId: "MetaCensusOrdererMSP"
  certSecrets:
    - "metacensus-orderer-0-certs"
    - "metacensus-orderer-1-certs"
    - "metacensus-orderer-2-certs"
    - "metacensus-orderer-3-certs"
    - "metacensus-orderer-4-certs"
```

Install with:
```bash
helm install metacensus-fabric ./helm/fabric \
  -n fabric-metacensus \
  --create-namespace \
  -f custom-values.yaml
```

## Architecture

### Component Naming

Components are named using predictable patterns:

- **CAs**: `ca{index}-{org-id}` (e.g., `ca0-org1`)
- **Peers**: `peer{index}-{org-id}` (e.g., `peer0-org1`, `peer1-org1`)
- **CouchDB**: `couchdb-peer{index}-{org-id}` (e.g., `couchdb-peer0-org1`)
- **Orderers**: `orderer{index}-{org-id}` (e.g., `orderer0-org1`)

### CouchDB Pairing

Each peer is automatically paired with a dedicated CouchDB instance:
- `peer0-org1` → `couchdb-peer0-org1:5984`
- `peer1-org1` → `couchdb-peer1-org1:5984`

### MSP Structure

All components expect certificates at standard Fabric paths:

```
/etc/hyperledger/fabric/msp/
├── signcerts/
│   └── cert.pem       # Node certificate
├── keystore/
│   └── key.pem        # Private key
└── cacerts/
    └── ca.pem         # CA certificate
```

## Upgrading

To upgrade an existing release:

```bash
helm upgrade my-fabric ./helm/fabric \
  -n fabric-org1 \
  -f custom-values.yaml
```

### Adding More Peers

1. Update your values file:
```yaml
peer:
  count: 4  # Increased from 2
  certSecrets:
    - "fabric-peer-0-certs"
    - "fabric-peer-1-certs"
    - "fabric-peer-2-certs"
    - "fabric-peer-3-certs"
```

2. Create secrets for new peers:
```bash
kubectl create secret generic fabric-peer-2-certs --from-file=... -n fabric-org1
kubectl create secret generic fabric-peer-3-certs --from-file=... -n fabric-org1
```

3. Upgrade the release:
```bash
helm upgrade my-fabric ./helm/fabric -n fabric-org1 -f values.yaml
```

## Uninstalling

```bash
helm uninstall my-fabric -n fabric-org1
```

**Note**: This does not delete certificate secrets. Delete them manually if needed:
```bash
kubectl delete secrets -n fabric-org1 -l app.kubernetes.io/instance=my-fabric
```

## Integration with Temporal Workflows

This chart is designed to be deployed via Temporal workflows using the Helm Go SDK:

```go
import "helm.sh/helm/v3/pkg/action"

func DeployFabricOrg(ctx context.Context, orgID string, peerCount int) error {
    client := action.NewUpgrade(actionConfig)
    client.Namespace = fmt.Sprintf("fabric-%s", orgID)
    client.CreateNamespace = true
    client.Install = true  // Install if not exists
    client.Wait = true
    
    chart, _ := loader.Load("./helm/fabric")
    values := map[string]interface{}{
        "organization": map[string]interface{}{
            "id": orgID,
            "mspId": fmt.Sprintf("%sMSP", orgID),
        },
        "peer": map[string]interface{}{
            "count": peerCount,
        },
    }
    
    _, err := client.Run(fmt.Sprintf("fabric-%s", orgID), chart, values)
    return err  // Idempotent - safe to retry
}
```

## Troubleshooting

### Pods stuck in Pending

Check if certificate secrets exist:
```bash
kubectl get secrets -n fabric-org1
```

### Pods crash with permission errors

Verify secret contents have correct structure:
```bash
kubectl get secret fabric-peer-0-certs -n fabric-org1 -o jsonpath='{.data}' | jq 'keys'
# Should show: ["cacerts/ca.pem", "keystore/key.pem", "signcerts/cert.pem"]
```

### Peer can't connect to CouchDB

Check CouchDB pod logs:
```bash
kubectl logs -n fabric-org1 -l app=fabric-couchdb
```

Verify CouchDB service exists:
```bash
kubectl get svc -n fabric-org1 -l app=fabric-couchdb
```

## License

Apache License 2.0

