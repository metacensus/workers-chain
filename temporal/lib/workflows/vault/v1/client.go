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
)

// Sentinel errors for Vault API responses
var (
	ErrNotFound     = errors.New("vault: not found")
	ErrUnauthorized = errors.New("vault: unauthorized")
	ErrForbidden    = errors.New("vault: forbidden")
)

// VaultConfig represents Vault connection configuration
type VaultConfig struct {
	Address string // Vault server address (e.g., "https://vault.example.com")
	Role    string // Kubernetes auth role name
}

// VaultClient defines generic path-based operations for Vault
type VaultClient interface {
	// Write performs a PUT request to the given path with data
	Write(ctx context.Context, path string, data map[string]any) error

	// Read performs a GET request to the given path and returns the response data
	Read(ctx context.Context, path string) (map[string]any, error)

	// Delete performs a DELETE request to the given path
	Delete(ctx context.Context, path string) error
}

// vaultClient implements VaultClient
type vaultClient struct {
	config     VaultConfig
	httpClient *http.Client
	token      string // Actual auth token (from config or k8s auth)
}

// NewVaultClient creates a new Vault client with the given configuration
// It authenticates using Kubernetes service account auth method.
func NewVaultClient(config VaultConfig) (VaultClient, error) {
	client := &vaultClient{
		config: config,
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
