package secrets

import "fmt"

// Credentials represents username/password
type Credentials struct {
	Username string
	Password string
}

// KeyPair represents a public/private key pair (PEM-encoded)
type KeyPair struct {
	PrivateKey []byte
	PublicKey  []byte
}

// Certificate represents a certificate with its chain (PEM-encoded)
// Per cert-design.md, all certificate paths store both:
//   - certificate: the leaf certificate only
//   - chain: the full chain (leaf + issuer chain, or just leaf for root CAs)
type Certificate struct {
	Certificate []byte // Leaf certificate only
	Chain       []byte // Full chain (leaf + issuer chain to root)
}

// IdentityCredentials represents a complete identity (keypair + certificate)
// This is used for storing admin credentials where both key and cert are needed together
type IdentityCredentials struct {
	PrivateKey  []byte
	PublicKey   []byte
	Certificate []byte
	CAChain     []byte // Optional CA certificate chain
}

// Secret is generic secret data (untyped key-value pairs)
type Secret map[string]any

// FromCredentials creates a Secret from Credentials
func FromCredentials(creds *Credentials) Secret {
	return Secret{
		"username": creds.Username,
		"password": creds.Password,
	}
}

// FromKeyPair creates a Secret from KeyPair
func FromKeyPair(kp *KeyPair) Secret {
	return Secret{
		"private_key": string(kp.PrivateKey),
		"public_key":  string(kp.PublicKey),
	}
}

// FromCertificate creates a Secret from Certificate
// Stores both certificate (leaf) and chain (full chain including leaf)
func FromCertificate(cert *Certificate) Secret {
	return Secret{
		"certificate": string(cert.Certificate),
		"chain":       string(cert.Chain),
	}
}

// FromIdentityCredentials creates a Secret from IdentityCredentials
func FromIdentityCredentials(creds *IdentityCredentials) Secret {
	secret := Secret{
		"private_key": string(creds.PrivateKey),
		"public_key":  string(creds.PublicKey),
		"certificate": string(creds.Certificate),
	}
	if len(creds.CAChain) > 0 {
		secret["ca_chain"] = string(creds.CAChain)
	}
	return secret
}

// Credentials extracts username/password credentials from the secret
func (s Secret) Credentials() (*Credentials, error) {
	username, ok := s["username"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid username in secret")
	}
	password, ok := s["password"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid password in secret")
	}
	return &Credentials{Username: username, Password: password}, nil
}

// KeyPair extracts a keypair from the secret
func (s Secret) KeyPair() (*KeyPair, error) {
	privateKey, ok := s["private_key"].(string)
	if !ok {
		// Try "key" as alternative
		privateKey, ok = s["key"].(string)
		if !ok {
			return nil, fmt.Errorf("missing or invalid private_key in secret")
		}
	}
	publicKey, ok := s["public_key"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid public_key in secret")
	}
	return &KeyPair{
		PrivateKey: []byte(privateKey),
		PublicKey:  []byte(publicKey),
	}, nil
}

// Certificate extracts a certificate with chain from the secret
func (s Secret) Certificate() (*Certificate, error) {
	certificate, ok := s["certificate"].(string)
	if !ok {
		// Try "cert" as alternative
		certificate, ok = s["cert"].(string)
		if !ok {
			return nil, fmt.Errorf("missing or invalid certificate in secret")
		}
	}
	// Chain is optional for backwards compatibility
	chain := certificate // Default to certificate if no chain
	if c, ok := s["chain"].(string); ok && c != "" {
		chain = c
	}
	return &Certificate{
		Certificate: []byte(certificate),
		Chain:       []byte(chain),
	}, nil
}

// IdentityCredentials extracts identity credentials (keypair + certificate) from the secret
func (s Secret) IdentityCredentials() (*IdentityCredentials, error) {
	privateKey, ok := s["private_key"].(string)
	if !ok {
		// Try "key" as alternative
		privateKey, ok = s["key"].(string)
		if !ok {
			return nil, fmt.Errorf("missing or invalid private_key in secret")
		}
	}

	publicKey, ok := s["public_key"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid public_key in secret")
	}

	certificate, ok := s["certificate"].(string)
	if !ok {
		// Try "cert" as alternative
		certificate, ok = s["cert"].(string)
		if !ok {
			return nil, fmt.Errorf("missing or invalid certificate in secret")
		}
	}

	// CA chain is optional
	caChain := ""
	if val, ok := s["ca_chain"].(string); ok {
		caChain = val
	}

	return &IdentityCredentials{
		PrivateKey:  []byte(privateKey),
		PublicKey:   []byte(publicKey),
		Certificate: []byte(certificate),
		CAChain:     []byte(caChain),
	}, nil
}
