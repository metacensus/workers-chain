package vault

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// VaultConfig represents Vault configuration
type VaultConfig struct {
	Address    string `json:"address"`
	Token      string `json:"token"`
	MountPath  string `json:"mount_path"`
	EngineType string `json:"engine_type"` // kv or transit
}

// VaultClient defines the interface for Vault operations
type VaultClient interface {
	StoreSecret(ctx context.Context, input StoreSecretInput) (*StoreSecretOutput, error)
	FetchSecret(ctx context.Context, input FetchSecretInput) (*FetchSecretOutput, error)
}

// vaultClient implements VaultClient
type vaultClient struct {
	config VaultConfig
	client *http.Client
}

// NewVaultClient creates a new Vault client
func NewVaultClient(config VaultConfig) VaultClient {
	return &vaultClient{
		config: config,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// StoreSecret stores a secret in Vault
func (vc *vaultClient) StoreSecret(ctx context.Context, input StoreSecretInput) (*StoreSecretOutput, error) {
	// Prepare the data to store in Vault
	secretData := make(map[string]interface{})
	for key, value := range input.Data {
		secretData[key] = base64.StdEncoding.EncodeToString(value)
	}
	secretData["version"] = input.Version
	secretData["labels"] = input.Labels
	secretData["created_at"] = time.Now().UTC().Format(time.RFC3339)

	jsonData, err := json.Marshal(secretData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal secret data: %w", err)
	}

	url := fmt.Sprintf("%s/v1/%s/data/%s", vc.config.Address, vc.config.MountPath, input.Name)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create Vault request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Vault-Token", vc.config.Token)

	resp, err := vc.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make Vault request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read Vault response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return &StoreSecretOutput{
			Name:    input.Name,
			Success: false,
			Message: fmt.Sprintf("Vault request failed with status %d: %s", resp.StatusCode, string(body)),
		}, fmt.Errorf("Vault request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return &StoreSecretOutput{
		Name:      input.Name,
		Namespace: "vault",
		Version:   input.Version,
		Success:   true,
		Message:   "Secret stored successfully",
	}, nil
}

// FetchSecret fetches a secret from Vault
func (vc *vaultClient) FetchSecret(ctx context.Context, input FetchSecretInput) (*FetchSecretOutput, error) {
	url := fmt.Sprintf("%s/v1/%s/data/%s", vc.config.Address, vc.config.MountPath, input.Name)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Vault request: %w", err)
	}
	req.Header.Set("X-Vault-Token", vc.config.Token)

	resp, err := vc.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make Vault request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read Vault response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return &FetchSecretOutput{
			Name:    input.Name,
			Found:   false,
			Message: fmt.Sprintf("Vault request failed with status %d: %s", resp.StatusCode, string(body)),
		}, fmt.Errorf("Vault request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Simulate decoding the response (in real code, parse the actual Vault response structure)
	var respData struct {
		Data struct {
			Data    map[string]string `json:"data"`
			Labels  map[string]string `json:"labels"`
			Version int64             `json:"version"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &respData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Vault response: %w", err)
	}

	decodedData := make(map[string][]byte)
	for k, v := range respData.Data.Data {
		decoded, _ := base64.StdEncoding.DecodeString(v)
		decodedData[k] = decoded
	}

	return &FetchSecretOutput{
		Name:    input.Name,
		Data:    decodedData,
		Labels:  respData.Data.Labels,
		Version: respData.Data.Version,
		Found:   true,
		Message: "Secret fetched successfully",
	}, nil
}

// GetVaultConfig returns Vault configuration from environment variables
func GetVaultConfig() VaultConfig {
	address := os.Getenv("VAULT_ADDR")
	if address == "" {
		address = "http://vault.secrets.svc.cluster.local:8200"
	}

	token := os.Getenv("VAULT_TOKEN")
	if token == "" {
		token = "root" // Default dev token
	}

	mountPath := os.Getenv("VAULT_MOUNT_PATH")
	if mountPath == "" {
		mountPath = "secret"
	}

	engineType := os.Getenv("VAULT_ENGINE_TYPE")
	if engineType == "" {
		engineType = "kv"
	}

	return VaultConfig{
		Address:    address,
		Token:      token,
		MountPath:  mountPath,
		EngineType: engineType,
	}
}
