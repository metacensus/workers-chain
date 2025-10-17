package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	pb "github.com/metacensus/infra/temporal/gen/workflows/vault/v1"
)

// VaultConfig represents Vault configuration
type VaultConfig struct {
	Address    string
	Token      string
	MountPath  string
	EngineType string // kv or transit
}

// VaultClient defines the interface for Vault operations
type VaultClient interface {
	StoreSecret(ctx context.Context, req *pb.Secret) (*pb.SecretRef, error)
	FetchSecret(ctx context.Context, req *pb.SecretRef) (*pb.Secret, error)
}

// vaultClient implements VaultClient
type vaultClient struct {
	config     VaultConfig
	httpClient *http.Client
}

// NewVaultClient creates a new Vault client with the given configuration
// This is the constructor that fx will use to provide VaultClient
func NewVaultClient(config VaultConfig) VaultClient {
	return &vaultClient{
		config: config,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// StoreSecret stores a secret in Vault
func (vc *vaultClient) StoreSecret(ctx context.Context, req *pb.Secret) (*pb.SecretRef, error) {
	path := fmt.Sprintf("%s/data/%s/%s", vc.config.MountPath, req.GetNamespace(), req.GetName())
	url := fmt.Sprintf("%s/v1/%s", strings.TrimSuffix(vc.config.Address, "/"), path)

	// Convert byte data to strings for Vault KV v2 storage
	dataStr := make(map[string]string)
	for k, v := range req.GetData() {
		dataStr[k] = string(v)
	}

	// Prepare request body with Vault KV v2 format
	body := map[string]any{
		"data": dataStr,
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(bodyBytes)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Vault-Token", vc.config.Token)

	resp, err := vc.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return nil, fmt.Errorf("vault returned status %d", resp.StatusCode)
	}

	// Parse response to get version
	var vaultResp struct {
		Data struct {
			Version int64 `json:"version"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&vaultResp); err != nil {
		// If we can't decode, still consider it success but with unknown version
		return &pb.SecretRef{
			Name:      req.GetName(),
			Namespace: req.GetNamespace(),
			Version:   req.GetVersion(),
		}, nil
	}

	return &pb.SecretRef{
		Name:      req.GetName(),
		Namespace: req.GetNamespace(),
		Version:   vaultResp.Data.Version,
	}, nil
}

// FetchSecret retrieves a secret from Vault
func (vc *vaultClient) FetchSecret(ctx context.Context, req *pb.SecretRef) (*pb.Secret, error) {
	path := fmt.Sprintf("%s/data/%s/%s", vc.config.MountPath, req.GetNamespace(), req.GetName())
	url := fmt.Sprintf("%s/v1/%s", strings.TrimSuffix(vc.config.Address, "/"), path)

	// Add version parameter if specified
	if req.GetVersion() > 0 {
		url = fmt.Sprintf("%s?version=%d", url, req.GetVersion())
	}

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("X-Vault-Token", vc.config.Token)

	resp, err := vc.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("secret not found: %s/%s", req.GetNamespace(), req.GetName())
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("vault returned status %d", resp.StatusCode)
	}

	// Parse Vault KV v2 response
	var vaultResp struct {
		Data struct {
			Data     map[string]string `json:"data"`
			Metadata struct {
				Version int64 `json:"version"`
			} `json:"metadata"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&vaultResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Convert strings to bytes
	data := make(map[string][]byte)
	for k, v := range vaultResp.Data.Data {
		data[k] = []byte(v)
	}

	return &pb.Secret{
		Name:      req.GetName(),
		Namespace: req.GetNamespace(),
		Data:      data,
		Version:   vaultResp.Data.Metadata.Version,
	}, nil
}
