# Fabric Certificate Matrix

## TLS Infrastructure

| Service/Entity | Purpose | Issued By | Notes |
|---|---|---|---|
| TLS Root CA | CA Signing | Self-signed | Offline after bootstrap |
| TLS Root CA | Server TLS | Self-signed | For CA's HTTPS endpoint |
| TLS Int CA | CA Signing | TLS Root CA | Signs all TLS certs |
| TLS Int CA | Server TLS | TLS Root CA | For CA's HTTPS endpoint |

## TLS Certificate Authority Admins

| Service/Entity | Purpose | Issued By | Notes |
|---|---|---|---|
| TLS Root CA Admin | Identity | TLS Root CA | TLS Root; admin privs |
| TLS Root CA Admin | Client TLS | TLS Root CA | TLS Root; admin privs |
| TLS Int CA Admin | Identity | TLS Int CA | Admin privs on TLS Int CA (must be in CA's database) |
| TLS Int CA Admin | Client TLS | TLS Int CA | mTLS to TLS Int CA |

## Organization Infrastructure

| Service/Entity | Purpose | Issued By | Notes |
|---|---|---|---|
| Org Root CA | CA Signing | Self-signed | Offline after bootstrap |
| Org Root CA | Server TLS | TLS Int CA | For CA's HTTPS endpoint |
| Org Int CA | CA Signing | Org Root CA | Signs all identity certs |
| Org Int CA | Server TLS | TLS Int CA | For CA's HTTPS endpoint |

## Organization Certificate Authority Admins

| Service/Entity | Purpose | Issued By | Notes |
|---|---|---|---|
| Org Root CA Admin | Identity | Org Root CA | Admin privs on Org Root CA (must be in CA's database) |
| Org Root CA Admin | Client TLS | TLS Int CA | mTLS to Org Root CA |
| Org Int CA Admin | Identity | Org Int CA | Admin privs on Org Int CA (must be in CA's database) |
| Org Int CA Admin | Client TLS | TLS Int CA | mTLS to Org Int CA |

## Organization Entities

| Service/Entity | Purpose | Issued By | Notes |
|---|---|---|---|
| Orderer _ | Identity | Org Int CA | MSP signing |
| Orderer _ | Server/Client TLS | TLS Int CA | mTLS with peers and orderers |
| Peer _ | Identity | Org Int CA | Endorsement signing |
| Peer _ | Server/Client TLS | TLS Int CA | mTLS with orderers and clients |
| Org Admin | Identity | Org Int CA | Channel/chaincode ops |
| Org Admin | Client TLS | TLS Int CA | mTLS to peers and orderers |
| Client App | Identity | Org Int CA | Transaction signing |
| Client App | Client TLS | TLS Int CA | mTLS to peers and orderers |

## Other Identities

| Service/Entity | Purpose | Issued By | Notes |
|---|---|---|---|
| Chaincode | Server TLS | TLS Int CA | Peer-to-chaincode mTLS |
