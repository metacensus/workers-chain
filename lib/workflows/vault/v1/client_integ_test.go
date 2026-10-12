//go:build integration
// +build integration

package vault

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/metacensus/workers-chain/lib/workers/config"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// TestVaultClient_GenericOperations tests the generic Write/Read/Delete operations
// Run with: go test -tags=integration ./workflows/vault/v1
func TestVaultClient_GenericOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	// Start Vault in dev mode
	req := testcontainers.ContainerRequest{
		Image:        "hashicorp/vault:1.15",
		ExposedPorts: []string{"8200/tcp"},
		Env: map[string]string{
			"VAULT_DEV_ROOT_TOKEN_ID":  "root",
			"VAULT_DEV_LISTEN_ADDRESS": "0.0.0.0:8200",
		},
		Cmd: []string{"server", "-dev", "-dev-listen-address=0.0.0.0:8200"},
		WaitingFor: wait.ForHTTP("/v1/sys/health").
			WithPort("8200/tcp").
			WithStartupTimeout(30 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("failed to start vault container: %v", err)
	}
	defer func() {
		if err := container.Terminate(ctx); err != nil {
			t.Logf("failed to terminate container: %v", err)
		}
	}()

	// Get connection details
	host, err := container.Host(ctx)
	if err != nil {
		t.Fatalf("failed to get container host: %v", err)
	}
	port, err := container.MappedPort(ctx, "8200")
	if err != nil {
		t.Fatalf("failed to get container port: %v", err)
	}

	vaultAddr := fmt.Sprintf("http://%s:%s", host, port.Port())

	// Create generic client using token (test helper for dev mode vault)
	client := newVaultClientWithToken(vaultAddr, "root")

	t.Run("write and read secret data", func(t *testing.T) {
		// Write secret data using generic Write
		path := "secret/data/test-namespace/my-secret"
		data := map[string]any{
			"data": map[string]string{
				"username": "admin",
				"password": "secret123",
			},
		}

		err := client.Write(ctx, path, data)
		if err != nil {
			t.Fatalf("failed to write secret: %v", err)
		}

		// Read it back using generic Read
		result, err := client.Read(ctx, path)
		if err != nil {
			t.Fatalf("failed to read secret: %v", err)
		}

		// Verify response structure
		dataSection, ok := result["data"].(map[string]any)
		if !ok {
			t.Fatal("result missing 'data' section")
		}

		secretData, ok := dataSection["data"].(map[string]any)
		if !ok {
			t.Fatal("data section missing 'data' subsection")
		}

		if username, ok := secretData["username"].(string); !ok || username != "admin" {
			t.Errorf("username mismatch: got %q", username)
		}
	})

	t.Run("delete secret", func(t *testing.T) {
		// First write a secret
		path := "secret/data/test-namespace/to-delete"
		data := map[string]any{
			"data": map[string]string{
				"key": "value",
			},
		}
		err := client.Write(ctx, path, data)
		if err != nil {
			t.Fatalf("failed to write secret: %v", err)
		}

		// Delete it
		err = client.Delete(ctx, path)
		if err != nil {
			t.Fatalf("failed to delete secret: %v", err)
		}

		// Verify it's gone
		_, err = client.Read(ctx, path)
		if err == nil {
			t.Fatal("expected error reading deleted secret, got nil")
		}
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got: %v", err)
		}
	})

	t.Run("read nonexistent secret returns not found", func(t *testing.T) {
		path := "secret/data/test-namespace/does-not-exist"
		_, err := client.Read(ctx, path)
		if err == nil {
			t.Fatal("expected error for nonexistent secret, got nil")
		}
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got: %v", err)
		}
	})

	t.Run("write and read policies", func(t *testing.T) {
		// Write a policy
		policyPath := "sys/policies/acl/test-policy"
		policyData := map[string]any{
			"policy": `path "secret/data/test/*" { capabilities = ["read"] }`,
		}

		err := client.Write(ctx, policyPath, policyData)
		if err != nil {
			t.Fatalf("failed to write policy: %v", err)
		}

		// Read it back
		result, err := client.Read(ctx, policyPath)
		if err != nil {
			t.Fatalf("failed to read policy: %v", err)
		}

		// Vault wraps policy response in "data" field
		data, ok := result["data"].(map[string]any)
		if !ok {
			t.Fatalf("expected 'data' field in policy response, got: %v", result)
		}

		if policy, ok := data["policy"].(string); !ok || policy == "" {
			t.Error("policy not found in response data")
		}

		// Delete the policy
		err = client.Delete(ctx, policyPath)
		if err != nil {
			t.Fatalf("failed to delete policy: %v", err)
		}
	})

	t.Run("write and read auth role", func(t *testing.T) {
		// Enable Kubernetes auth method first (required in dev mode)
		enablePath := "sys/auth/kubernetes"
		enableData := map[string]any{
			"type": "kubernetes",
		}
		err := client.Write(ctx, enablePath, enableData)
		if err != nil {
			t.Fatalf("failed to enable kubernetes auth: %v", err)
		}

		// Write a Kubernetes auth role
		rolePath := "auth/kubernetes/role/test-role"
		roleData := map[string]any{
			"bound_service_account_names":      []string{"test-sa"},
			"bound_service_account_namespaces": []string{"default"},
			"policies":                         []string{"test-policy"},
		}

		err = client.Write(ctx, rolePath, roleData)
		if err != nil {
			t.Fatalf("failed to write auth role: %v", err)
		}

		// Read it back
		result, err := client.Read(ctx, rolePath)
		if err != nil {
			t.Fatalf("failed to read auth role: %v", err)
		}

		// Vault wraps auth role response in "data" field
		data, ok := result["data"].(map[string]any)
		if !ok {
			t.Fatalf("expected 'data' field in auth role response, got: %v", result)
		}

		// Verify role data
		if names, ok := data["bound_service_account_names"].([]any); !ok || len(names) == 0 {
			t.Error("bound_service_account_names not found or empty")
		}

		// Delete the role
		err = client.Delete(ctx, rolePath)
		if err != nil {
			t.Fatalf("failed to delete auth role: %v", err)
		}

		// Disable the auth method (cleanup)
		err = client.Delete(ctx, enablePath)
		if err != nil {
			t.Logf("failed to disable kubernetes auth (cleanup): %v", err)
		}
	})
}

// TestVaultClient_ErrorHandling tests error cases with real Vault
func TestVaultClient_ErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	// Start Vault
	req := testcontainers.ContainerRequest{
		Image:        "hashicorp/vault:1.15",
		ExposedPorts: []string{"8200/tcp"},
		Env: map[string]string{
			"VAULT_DEV_ROOT_TOKEN_ID": "root",
		},
		Cmd: []string{"server", "-dev", "-dev-listen-address=0.0.0.0:8200"},
		WaitingFor: wait.ForHTTP("/v1/sys/health").
			WithPort("8200/tcp").
			WithStartupTimeout(30 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("failed to start vault container: %v", err)
	}
	defer container.Terminate(ctx)

	host, _ := container.Host(ctx)
	port, _ := container.MappedPort(ctx, "8200")
	vaultAddr := fmt.Sprintf("http://%s:%s", host, port.Port())

	t.Run("unauthorized token returns error", func(t *testing.T) {
		client := newVaultClientWithToken(vaultAddr, "invalid-token")

		path := "secret/data/test/secret"
		data := map[string]any{
			"data": map[string]string{"key": "value"},
		}

		err := client.Write(ctx, path, data)
		if err == nil {
			t.Fatal("expected error with invalid token, got nil")
		}
		// Vault returns 403 Forbidden for invalid tokens, not 401 Unauthorized
		if !errors.Is(err, ErrForbidden) {
			t.Errorf("expected ErrForbidden, got: %v", err)
		}
	})

	t.Run("read nonexistent path returns not found", func(t *testing.T) {
		client := newVaultClientWithToken(vaultAddr, "root")

		_, err := client.Read(ctx, "secret/data/nonexistent/secret")
		if err == nil {
			t.Fatal("expected error for nonexistent secret, got nil")
		}
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got: %v", err)
		}
	})
}

// newVaultClientWithToken creates a Vault client with a pre-configured token.
// This is for testing only - production code should use NewVaultClient with k8s auth.
func newVaultClientWithToken(address, token string) VaultClient {
	return &vaultClient{
		config: &config.VaultConfig{
			Address: address,
		},
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		token: token,
	}
}
