package identity

import "fmt"

// VaultPaths provides fabric-specific path construction for Vault secrets
type VaultPaths struct {
	organization string
}

// NewVaultPaths creates a new VaultPaths helper for an organization
func NewVaultPaths(organization string) *VaultPaths {
	return &VaultPaths{
		organization: organization,
	}
}

// PeerMSP returns the path for storing peer MSP certificates
func (vp *VaultPaths) PeerMSP(peerID string) string {
	return fmt.Sprintf("fabric/%s/msp/peer-%s", vp.organization, peerID)
}

// OrdererMSP returns the path for storing orderer MSP certificates
func (vp *VaultPaths) OrdererMSP(ordererID string) string {
	return fmt.Sprintf("fabric/%s/msp/orderer-%s", vp.organization, ordererID)
}

// ClientMSP returns the path for storing client MSP certificates
func (vp *VaultPaths) ClientMSP(clientID string) string {
	return fmt.Sprintf("fabric/%s/msp/client-%s", vp.organization, clientID)
}

// ClientTLSMSP returns the path for storing client TLS certificates
func (vp *VaultPaths) ClientTLSMSP(clientID string) string {
	return fmt.Sprintf("fabric/%s/msp/client-%s-tls", vp.organization, clientID)
}

// AdminMSP returns the path for storing admin identity
func (vp *VaultPaths) AdminMSP() string {
	return fmt.Sprintf("fabric/%s/msp/admin", vp.organization)
}

// CACerts returns the path for storing CA certificates
func (vp *VaultPaths) CACerts() string {
	return fmt.Sprintf("fabric/%s/msp/ca-certs", vp.organization)
}

// Config returns the path for storing configuration (NodeOU, etc)
func (vp *VaultPaths) Config() string {
	return fmt.Sprintf("fabric/%s/msp/config", vp.organization)
}
