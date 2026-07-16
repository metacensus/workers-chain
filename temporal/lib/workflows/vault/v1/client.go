package vault

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/metacensus/infra/temporal/lib/workers/config"
)

// Sentinel errors for Vault API responses
var (
	ErrNotFound            = errors.New("vault: not found")
	ErrUnauthorized        = errors.New("vault: unauthorized")
	ErrForbidden           = errors.New("vault: forbidden")
	ErrInvalidResponseData = errors.New("vault: invalid response data")
)

// VaultClient defines generic path-based operations for Vault
type VaultClient interface {
	// Low-level generic operations (path-based)
	Write(ctx context.Context, path string, data map[string]any) error
	Read(ctx context.Context, path string) (map[string]any, error)
	Delete(ctx context.Context, path string) error

	// Low-level secret operations (path-based)
	FetchSecret(ctx context.Context, path string) (map[string]interface{}, error)

	// High-level secret operations (ref-based, for SecretClient interface)
	FetchSecretByRef(ctx context.Context, ref map[string]interface{}) (map[string]interface{}, error)
	StoreSecretByRef(ctx context.Context, ref map[string]interface{}, secret map[string]interface{}) error
	UpdateSecretByRef(ctx context.Context, ref map[string]interface{}, partialSecret map[string]interface{}) error
}

// vaultClient implements VaultClient
type vaultClient struct {
	config     *config.VaultConfig
	httpClient *http.Client
	token      string // Actual auth token (from config or k8s auth)
}

// NewVaultClient creates a new Vault client with the given configuration
// It authenticates using Kubernetes service account auth method.
func NewVaultClient(cfg *config.VaultConfig) (VaultClient, error) {
	client := &vaultClient{
		config: cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	// Authenticate using Kubernetes service account
	token, err := client.authenticateKubernetes(context.Background())
	if err != nil {
		return nil, fmt.Errorf("kubernetes auth failed: %w", err)
	}
	client.token = token

	return client, nil
}

// Write performs a PUT request to the given path with data
func (vc *vaultClient) Write(ctx context.Context, path string, data map[string]any) error {
	url := vc.buildURL(path)

	bodyBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPut, url, strings.NewReader(string(bodyBytes)))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	vc.setHeaders(httpReq)
	resp, err := vc.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if err := vc.checkError(resp); err != nil {
		return err
	}

	return nil
}

// Read performs a GET request to the given path and returns the response data
func (vc *vaultClient) Read(ctx context.Context, path string) (map[string]any, error) {
	url := vc.buildURL(path)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	vc.setHeaders(httpReq)
	resp, err := vc.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if err := vc.checkError(resp); err != nil {
		return nil, err
	}

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return result, nil
}

// Delete performs a DELETE request to the given path
func (vc *vaultClient) Delete(ctx context.Context, path string) error {
	url := vc.buildURL(path)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	vc.setHeaders(httpReq)
	resp, err := vc.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if err := vc.checkError(resp); err != nil {
		return err
	}

	return nil
}

// FetchSecret fetches a secret from KV v2 secrets engine and unwraps the response
// path should be the logical path without the mount prefix (e.g., "fabric/shard/org1/credentials")
// Returns the unwrapped secret data as map[string]interface{}
func (vc *vaultClient) FetchSecret(ctx context.Context, path string) (map[string]interface{}, error) {
	// Vault KV v2 path structure: <mount>/data/<path>
	// We assume "secret" mount by default
	kvv2Path := fmt.Sprintf("secret/data/%s", path)

	// Read from Vault using the low-level Read method
	result, err := vc.Read(ctx, kvv2Path)
	if err != nil {
		return nil, err
	}

	// Parse Vault KV v2 response structure: {data: {data: {...}, metadata: {version: N}}}
	dataSection, ok := result["data"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%w: missing 'data' field in response", ErrInvalidResponseData)
	}

	secretData, ok := dataSection["data"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%w: missing 'data.data' field in response", ErrInvalidResponseData)
	}

	return secretData, nil
}

// buildURL constructs the full Vault API URL from a path
func (vc *vaultClient) buildURL(path string) string {
	baseURL := strings.TrimSuffix(vc.config.Address, "/")
	return fmt.Sprintf("%s/v1/%s", baseURL, path)
}

// authenticateKubernetes performs Kubernetes auth method login
func (vc *vaultClient) authenticateKubernetes(ctx context.Context) (string, error) {
	// Read service account JWT token
	jwtPath := "/var/run/secrets/kubernetes.io/serviceaccount/token"
	jwtBytes, err := os.ReadFile(jwtPath)
	if err != nil {
		return "", fmt.Errorf("read service account token: %w", err)
	}
	jwt := string(jwtBytes)

	// Prepare login request
	loginData := map[string]any{
		"role": vc.config.Role,
		"jwt":  jwt,
	}

	bodyBytes, err := json.Marshal(loginData)
	if err != nil {
		return "", fmt.Errorf("marshal login request: %w", err)
	}

	// Send login request
	url := vc.buildURL("auth/kubernetes/login")
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(string(bodyBytes)))
	if err != nil {
		return "", fmt.Errorf("create login request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := vc.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("send login request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("login failed (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse response to get client token
	var loginResp struct {
		Auth struct {
			ClientToken string `json:"client_token"`
		} `json:"auth"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		return "", fmt.Errorf("decode login response: %w", err)
	}

	if loginResp.Auth.ClientToken == "" {
		return "", errors.New("login response missing client_token")
	}

	return loginResp.Auth.ClientToken, nil
}

// setHeaders sets authentication and content-type headers on the request
func (vc *vaultClient) setHeaders(req *http.Request) {
	req.Header.Set("X-Vault-Token", vc.token)
	req.Header.Set("Content-Type", "application/json")
}

// checkError checks the HTTP response status and returns appropriate errors
func (vc *vaultClient) checkError(resp *http.Response) error {
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNoContent {
		return nil
	}

	// Read error body for Vault error message
	body, _ := io.ReadAll(resp.Body)
	errMsg := string(body)

	switch resp.StatusCode {
	case http.StatusNotFound:
		return fmt.Errorf("vault: %w (status: %d)", ErrNotFound, resp.StatusCode)
	case http.StatusForbidden:
		return fmt.Errorf("vault: %w (status: %d)", ErrForbidden, resp.StatusCode)
	case http.StatusUnauthorized:
		return fmt.Errorf("vault: %w (status: %d)", ErrUnauthorized, resp.StatusCode)
	default:
		return fmt.Errorf("vault: request failed with status %d: %s", resp.StatusCode, errMsg)
	}
}

// =============================================================================
// High-level secret operations (for SecretClient compatibility)
// =============================================================================

const vaultFieldData = "data"

// extractVaultPath extracts the "path" field from a ref map
func extractVaultPath(ref map[string]interface{}) (string, error) {
	if ref == nil {
		return "", fmt.Errorf("ref is nil")
	}
	pathValue, ok := ref["path"]
	if !ok {
		return "", fmt.Errorf("missing 'path' field in ref")
	}
	path, ok := pathValue.(string)
	if !ok || path == "" {
		return "", fmt.Errorf("'path' field is invalid or empty")
	}
	return path, nil
}

func buildKVPath(path string) string {
	return fmt.Sprintf("secret/data/%s", path)
}

// FetchSecretByRef fetches a secret using a ref map (implements SecretClient interface)
func (vc *vaultClient) FetchSecretByRef(ctx context.Context, ref map[string]interface{}) (map[string]interface{}, error) {
	path, err := extractVaultPath(ref)
	if err != nil {
		return nil, err
	}
	return vc.FetchSecret(ctx, path)
}

// StoreSecretByRef stores a secret using a ref map (implements SecretClient interface)
func (vc *vaultClient) StoreSecretByRef(ctx context.Context, ref map[string]interface{}, secret map[string]interface{}) error {
	path, err := extractVaultPath(ref)
	if err != nil {
		return err
	}

	body := map[string]any{
		vaultFieldData: secret,
	}
	return vc.Write(ctx, buildKVPath(path), body)
}

// UpdateSecretByRef merges new fields with existing secret data (implements SecretClient interface)
func (vc *vaultClient) UpdateSecretByRef(ctx context.Context, ref map[string]interface{}, partialSecret map[string]interface{}) error {
	path, err := extractVaultPath(ref)
	if err != nil {
		return err
	}

	// Fetch existing secret (if it exists)
	existingSecret, err := vc.FetchSecret(ctx, path)
	if err != nil {
		// If secret doesn't exist, create it with the partial data
		existingSecret = make(map[string]interface{})
	}

	// Merge: new fields override existing ones
	for key, value := range partialSecret {
		existingSecret[key] = value
	}

	// Write back the merged secret
	body := map[string]any{
		vaultFieldData: existingSecret,
	}
	return vc.Write(ctx, buildKVPath(path), body)
}
