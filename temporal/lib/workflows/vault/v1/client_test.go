//go:build integration
// +build integration

package vault

import (
	"context"
	"fmt"
	"testing"
	"time"

	pb "github.com/metacensus/infra/temporal/gen/workflows/vault/v1"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// TestVaultClient_RealVault tests against an actual Vault instance in Docker
// Run with: go test -tags=integration ./workflows/vault/v1
func TestVaultClient_RealVault(t *testing.T) {
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

	// Create client
	client := NewVaultClient(VaultConfig{
		Address:    vaultAddr,
		Token:      "root",
		MountPath:  "secret",
		EngineType: "kv",
	})

	t.Run("store and fetch secret", func(t *testing.T) {
		// Store a secret
		secret := &pb.Secret{
			Name:      "test-secret",
			Namespace: "test-namespace",
			Data: map[string][]byte{
				"username": []byte("admin"),
				"password": []byte("secret123"),
			},
		}

		ref, err := client.StoreSecret(ctx, secret)
		if err != nil {
			t.Fatalf("failed to store secret: %v", err)
		}

		if ref.Name != secret.Name {
			t.Errorf("name mismatch: got %q, want %q", ref.Name, secret.Name)
		}
		if ref.Namespace != secret.Namespace {
			t.Errorf("namespace mismatch: got %q, want %q", ref.Namespace, secret.Namespace)
		}
		if ref.Version != 1 {
			t.Errorf("expected version 1, got %d", ref.Version)
		}

		// Fetch it back
		fetchRef := &pb.SecretRef{
			Name:      secret.Name,
			Namespace: secret.Namespace,
		}

		fetched, err := client.FetchSecret(ctx, fetchRef)
		if err != nil {
			t.Fatalf("failed to fetch secret: %v", err)
		}

		if fetched.Name != secret.Name {
			t.Errorf("fetched name mismatch: got %q, want %q", fetched.Name, secret.Name)
		}
		if string(fetched.Data["username"]) != string(secret.Data["username"]) {
			t.Errorf("username mismatch: got %q, want %q", fetched.Data["username"], secret.Data["username"])
		}
		if string(fetched.Data["password"]) != string(secret.Data["password"]) {
			t.Errorf("password mismatch: got %q, want %q", fetched.Data["password"], secret.Data["password"])
		}
	})

	t.Run("secret not found", func(t *testing.T) {
		ref := &pb.SecretRef{
			Name:      "nonexistent",
			Namespace: "test",
		}

		_, err := client.FetchSecret(ctx, ref)
		if err == nil {
			t.Fatal("expected error for nonexistent secret, got nil")
		}
	})

	t.Run("multiple versions", func(t *testing.T) {
		// Store version 1
		secret1 := &pb.Secret{
			Name:      "versioned-secret",
			Namespace: "test",
			Data: map[string][]byte{
				"value": []byte("version1"),
			},
		}
		ref1, err := client.StoreSecret(ctx, secret1)
		if err != nil {
			t.Fatalf("failed to store version 1: %v", err)
		}
		if ref1.Version != 1 {
			t.Errorf("expected version 1, got %d", ref1.Version)
		}

		// Store version 2 (update)
		secret2 := &pb.Secret{
			Name:      "versioned-secret",
			Namespace: "test",
			Data: map[string][]byte{
				"value": []byte("version2"),
			},
		}
		ref2, err := client.StoreSecret(ctx, secret2)
		if err != nil {
			t.Fatalf("failed to store version 2: %v", err)
		}
		if ref2.Version != 2 {
			t.Errorf("expected version 2, got %d", ref2.Version)
		}

		// Fetch version 1 specifically
		fetchRef1 := &pb.SecretRef{
			Name:      "versioned-secret",
			Namespace: "test",
			Version:   1,
		}
		fetched1, err := client.FetchSecret(ctx, fetchRef1)
		if err != nil {
			t.Fatalf("failed to fetch version 1: %v", err)
		}
		if string(fetched1.Data["value"]) != "version1" {
			t.Errorf("version 1 value: got %q, want %q", fetched1.Data["value"], "version1")
		}

		// Fetch version 2 (latest)
		fetchRef2 := &pb.SecretRef{
			Name:      "versioned-secret",
			Namespace: "test",
			Version:   0, // Latest
		}
		fetched2, err := client.FetchSecret(ctx, fetchRef2)
		if err != nil {
			t.Fatalf("failed to fetch latest: %v", err)
		}
		if string(fetched2.Data["value"]) != "version2" {
			t.Errorf("latest value: got %q, want %q", fetched2.Data["value"], "version2")
		}
		if fetched2.Version != 2 {
			t.Errorf("latest version: got %d, want 2", fetched2.Version)
		}
	})

	t.Run("multiple secrets in different namespaces", func(t *testing.T) {
		secrets := []*pb.Secret{
			{
				Name:      "secret1",
				Namespace: "prod",
				Data:      map[string][]byte{"key": []byte("prod-value")},
			},
			{
				Name:      "secret1",
				Namespace: "staging",
				Data:      map[string][]byte{"key": []byte("staging-value")},
			},
		}

		// Store both
		for _, s := range secrets {
			if _, err := client.StoreSecret(ctx, s); err != nil {
				t.Fatalf("failed to store %s/%s: %v", s.Namespace, s.Name, err)
			}
		}

		// Fetch and verify they're independent
		for _, s := range secrets {
			ref := &pb.SecretRef{
				Name:      s.Name,
				Namespace: s.Namespace,
			}
			fetched, err := client.FetchSecret(ctx, ref)
			if err != nil {
				t.Fatalf("failed to fetch %s/%s: %v", s.Namespace, s.Name, err)
			}
			if string(fetched.Data["key"]) != string(s.Data["key"]) {
				t.Errorf("%s/%s: got %q, want %q", s.Namespace, s.Name, fetched.Data["key"], s.Data["key"])
			}
		}
	})
}

// TestVaultClient_RealVault_ErrorHandling tests error cases with real Vault
func TestVaultClient_RealVault_ErrorHandling(t *testing.T) {
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

	t.Run("invalid token", func(t *testing.T) {
		client := NewVaultClient(VaultConfig{
			Address:    vaultAddr,
			Token:      "invalid-token",
			MountPath:  "secret",
			EngineType: "kv",
		})

		secret := &pb.Secret{
			Name:      "test",
			Namespace: "test",
			Data:      map[string][]byte{"key": []byte("value")},
		}

		_, err := client.StoreSecret(ctx, secret)
		if err == nil {
			t.Fatal("expected error with invalid token, got nil")
		}
	})

	t.Run("fetch nonexistent secret", func(t *testing.T) {
		client := NewVaultClient(VaultConfig{
			Address:    vaultAddr,
			Token:      "root",
			MountPath:  "secret",
			EngineType: "kv",
		})

		ref := &pb.SecretRef{
			Name:      "does-not-exist",
			Namespace: "test",
		}

		_, err := client.FetchSecret(ctx, ref)
		if err == nil {
			t.Fatal("expected error for nonexistent secret, got nil")
		}
	})
}
