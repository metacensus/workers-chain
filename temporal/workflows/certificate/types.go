package certificate

import (
	"time"

	"github.com/cloudflare/cfssl/csr"
	"github.com/hyperledger/fabric-ca/api"
)

// IdentityParams represents parameters for identity registration
type IdentityParams struct {
	Type        string            `json:"type"`        // peer|orderer|client|admin
	Name        string            `json:"name"`        // Identity name
	CAEndpoint  string            `json:"ca_endpoint"` // Fabric CA server endpoint
	Attributes  map[string]string `json:"attributes"`  // Identity attributes
	Affiliation string            `json:"affiliation"` // Organization hierarchy
	Secret      string            `json:"secret"`      // Registration secret
}

// Registration represents a registered identity
type Registration struct {
	ID          string            `json:"id"`
	Type        string            `json:"type"`
	Attributes  map[string]string `json:"attributes"`
	Affiliation string            `json:"affiliation"`
	CAEndpoint  string            `json:"ca_endpoint"`
	Secret      string            `json:"secret"` // Registration secret returned from CA
}

// EnrollmentParams represents parameters for certificate enrollment
type EnrollmentParams struct {
	Registration Registration `json:"registration"`
	CSRConfig    CSRConfig    `json:"csr_config"`
	Profile      string       `json:"profile"` // tls|sign
	CAEndpoint   string       `json:"ca_endpoint"`
}

// CSRConfig represents certificate signing request configuration
type CSRConfig struct {
	Hosts        []string `json:"hosts"`
	CommonName   string   `json:"common_name"`
	Organization string   `json:"organization"`
	Country      string   `json:"country"`
	Province     string   `json:"province"`
	Locality     string   `json:"locality"`
}

// Enrollment represents an enrolled certificate
type Enrollment struct {
	Certificate []byte `json:"certificate"`
	PrivateKey  []byte `json:"private_key"`
	CACert      []byte `json:"ca_cert"`
	Profile     string `json:"profile"`
}

// MSPParams represents parameters for MSP assembly
type MSPParams struct {
	SigningEnrollment Enrollment   `json:"signing_enrollment"`
	TLSEnrollment     *Enrollment  `json:"tls_enrollment,omitempty"`
	CACerts           [][]byte     `json:"ca_certs"`
	NodeOUConfig      NodeOUConfig `json:"node_ou_config"`
}

// NodeOUConfig represents Node OU configuration
type NodeOUConfig struct {
	Enable  bool              `json:"enable"`
	Client  map[string]string `json:"client"`
	Peer    map[string]string `json:"peer"`
	Admin   map[string]string `json:"admin"`
	Orderer map[string]string `json:"orderer"`
}

// MSPBundle represents a complete MSP structure
type MSPBundle struct {
	SigningCert  []byte       `json:"signing_cert"`
	SigningKey   []byte       `json:"signing_key"`
	TLSCert      []byte       `json:"tls_cert,omitempty"`
	TLSKey       []byte       `json:"tls_key,omitempty"`
	CACerts      [][]byte     `json:"ca_certs"`
	NodeOUConfig NodeOUConfig `json:"node_ou_config"`
}

// SecretParams represents parameters for secret storage
type SecretParams struct {
	Name    string            `json:"name"`
	Data    map[string][]byte `json:"data"`
	Labels  map[string]string `json:"labels"`
	Version int64             `json:"version"`
}

// SecretRef represents a reference to a stored secret
type SecretRef struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Version   int64  `json:"version"`
}

// DeploymentParams represents parameters for deployment updates
type DeploymentParams struct {
	ResourceType   string            `json:"resource_type"`
	SecretRef      SecretRef         `json:"secret_ref"`
	UpdateStrategy string            `json:"update_strategy"`
	HealthCheck    HealthCheckConfig `json:"health_check"`
}

// HealthCheckConfig represents health check configuration
type HealthCheckConfig struct {
	Endpoint string        `json:"endpoint"`
	Timeout  time.Duration `json:"timeout"`
	Interval time.Duration `json:"interval"`
}

// PeerConfig represents peer-specific configuration
type PeerConfig struct {
	PeerID       string       `json:"peer_id"`
	Organization string       `json:"organization"`
	EnrollmentCA string       `json:"enrollment_ca"`
	TLSCA        string       `json:"tls_ca"`
	OrgNodeOUs   NodeOUConfig `json:"org_node_ous"`
	Namespace    string       `json:"namespace"`
}

// OrdererConfig represents orderer-specific configuration
type OrdererConfig struct {
	OrdererID    string       `json:"orderer_id"`
	Organization string       `json:"organization"`
	EnrollmentCA string       `json:"enrollment_ca"`
	TLSCA        string       `json:"tls_ca"`
	OrgNodeOUs   NodeOUConfig `json:"org_node_ous"`
	Namespace    string       `json:"namespace"`
}

// ClientConfig represents client-specific configuration
type ClientConfig struct {
	ClientID         string            `json:"client_id"`
	CustomAttributes map[string]string `json:"custom_attributes"`
	Namespace        string            `json:"namespace"`
}

// RotationParams represents parameters for certificate rotation
type RotationParams struct {
	IdentityName string `json:"identity_name"`
	SecretName   string `json:"secret_name"`
	CAEndpoint   string `json:"ca_endpoint"`
	ResourceType string `json:"resource_type"`
	Version      int64  `json:"version"`
}

// RotationStrategy represents a rotation strategy
type RotationStrategy struct {
	Type              string            `json:"type"`
	Identity          string            `json:"identity"`
	SecretName        string            `json:"secret_name"`
	CAEndpoint        string            `json:"ca_endpoint"`
	ResourceType      string            `json:"resource_type"`
	Version           int64             `json:"version"`
	HealthCheckConfig HealthCheckConfig `json:"health_check_config"`
	GracePeriod       time.Duration     `json:"grace_period"`
}

// BatchParams represents parameters for batch operations
type BatchParams struct {
	CertificateType string `json:"certificate_type"`
	Namespace       string `json:"namespace"`
	Concurrency     int    `json:"concurrency"`
}

// FabricCAConfig represents Fabric CA client configuration
type FabricCAConfig struct {
	URL      string    `json:"url"`
	CAName   string    `json:"ca_name"`
	Username string    `json:"username"`
	Password string    `json:"password"`
	TLS      TLSConfig `json:"tls"`
	CSP      CSPConfig `json:"csp"`
}

// TLSConfig represents TLS configuration
type TLSConfig struct {
	Enabled bool `json:"enabled"`
}

// CSPConfig represents crypto service provider configuration
type CSPConfig struct {
	Provider string `json:"provider"`
}

// VaultConfig represents Vault configuration
type VaultConfig struct {
	Address    string `json:"address"`
	Token      string `json:"token"`
	MountPath  string `json:"mount_path"`
	EngineType string `json:"engine_type"` // kv or transit
}

// K8sSecretConfig represents Kubernetes secret configuration
type K8sSecretConfig struct {
	Namespace string            `json:"namespace"`
	Labels    map[string]string `json:"labels"`
}

// CertificateMetadata represents metadata for a certificate
type CertificateMetadata struct {
	ID           string    `json:"id"`
	Type         string    `json:"type"`
	Profile      string    `json:"profile"`
	NotBefore    time.Time `json:"not_before"`
	NotAfter     time.Time `json:"not_after"`
	SerialNumber string    `json:"serial_number"`
	Issuer       string    `json:"issuer"`
	Subject      string    `json:"subject"`
	Version      int64     `json:"version"`
	CreatedAt    time.Time `json:"created_at"`
}

// ToSecretData converts enrollment to secret data format
func (e *Enrollment) ToSecretData() map[string][]byte {
	return map[string][]byte{
		"cert.pem": e.Certificate,
		"key.pem":  e.PrivateKey,
		"ca.pem":   e.CACert,
	}
}

// ToTLSSecretData converts enrollment to TLS secret data format
func (e *Enrollment) ToTLSSecretData() map[string][]byte {
	return map[string][]byte{
		"tls.crt": e.Certificate,
		"tls.key": e.PrivateKey,
		"ca.crt":  e.CACert,
	}
}

// ToCSRInfo converts CSRConfig to Fabric CA CSR info
func (c *CSRConfig) ToCSRInfo() *api.CSRInfo {
	return &api.CSRInfo{
		CN:    c.CommonName,
		Hosts: c.Hosts,
		Names: []csr.Name{
			{
				C:  c.Country,
				ST: c.Province,
				L:  c.Locality,
				O:  c.Organization,
			},
		},
	}
}

// AdminIdentity represents an admin identity for Fabric CA
type AdminIdentity struct {
	ID          string            `json:"id"`
	Secret      string            `json:"secret"`
	Certificate []byte            `json:"certificate"`
	PrivateKey  []byte            `json:"private_key"`
	CACert      []byte            `json:"ca_cert"`
	Attributes  map[string]string `json:"attributes"`
	CAEndpoint  string            `json:"ca_endpoint"`
}

// BootstrapAdminParams contains parameters for bootstrap admin workflow
type BootstrapAdminParams struct {
	CAEndpoint     string            `json:"ca_endpoint"`
	AdminID        string            `json:"admin_id"`
	AdminSecret    string            `json:"admin_secret"`
	DefaultAdminID string            `json:"default_admin_id"`
	DefaultSecret  string            `json:"default_secret"`
	Attributes     map[string]string `json:"attributes"`
	VaultPath      string            `json:"vault_path"`
}

// BootstrapAdminResult contains the result of bootstrap admin workflow
type BootstrapAdminResult struct {
	AdminIdentity *AdminIdentity `json:"admin_identity"`
	Success       bool           `json:"success"`
	Message       string         `json:"message"`
}

// BootstrapConfig contains configuration for bootstrap workflows
type BootstrapConfig struct {
	CAEndpoint string       `json:"ca_endpoint"`
	OrgNodeOUs NodeOUConfig `json:"org_node_ous"`
	Namespace  string       `json:"namespace"`
}
