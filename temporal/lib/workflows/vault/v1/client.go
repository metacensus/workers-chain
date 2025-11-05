package vault

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
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
	Token   string // Authentication token
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
}

// NewVaultClient creates a new Vault client with the given configuration
func NewVaultClient(config VaultConfig) VaultClient {
	return &vaultClient{
		config: config,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
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

// setHeaders sets authentication and content-type headers on the request
func (vc *vaultClient) setHeaders(req *http.Request) {
	req.Header.Set("X-Vault-Token", vc.config.Token)
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
