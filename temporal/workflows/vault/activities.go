package vault

import (
	"context"

	"go.temporal.io/sdk/activity"
)

// StoreSecretInput is the input for storing a secret in Vault
type StoreSecretInput struct {
	Name    string            `json:"name"`
	Data    map[string][]byte `json:"data"`
	Labels  map[string]string `json:"labels,omitempty"`
	Version int64             `json:"version,omitempty"`
}

// StoreSecretOutput is the output for storing a secret in Vault
type StoreSecretOutput struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Version   int64  `json:"version"`
	Success   bool   `json:"success"`
	Message   string `json:"message,omitempty"`
}

// FetchSecretInput is the input for fetching a secret from Vault
type FetchSecretInput struct {
	Name string `json:"name"`
}

// FetchSecretOutput is the output for fetching a secret from Vault
type FetchSecretOutput struct {
	Name    string            `json:"name"`
	Data    map[string][]byte `json:"data"`
	Labels  map[string]string `json:"labels,omitempty"`
	Version int64             `json:"version,omitempty"`
	Found   bool              `json:"found"`
	Message string            `json:"message,omitempty"`
}

// Activities contains all vault-related activities
type Activities struct {
	client VaultClient
}

// NewActivities creates a new Activities instance
func NewActivities(client VaultClient) *Activities {
	return &Activities{
		client: client,
	}
}

// StoreSecretActivity stores a secret in Vault
func (a *Activities) StoreSecretActivity(ctx context.Context, input StoreSecretInput) (*StoreSecretOutput, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Storing secret in Vault", "name", input.Name)

	output, err := a.client.StoreSecret(ctx, input)
	if err != nil {
		logger.Error("Failed to store secret in Vault", "error", err)
		return &StoreSecretOutput{
			Name:    input.Name,
			Success: false,
			Message: err.Error(),
		}, err
	}

	logger.Info("Secret stored successfully in Vault", "name", input.Name)
	return output, nil
}

// FetchSecretActivity fetches a secret from Vault
func (a *Activities) FetchSecretActivity(ctx context.Context, input FetchSecretInput) (*FetchSecretOutput, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Fetching secret from Vault", "name", input.Name)

	output, err := a.client.FetchSecret(ctx, input)
	if err != nil {
		logger.Error("Failed to fetch secret from Vault", "error", err)
		return &FetchSecretOutput{
			Name:    input.Name,
			Found:   false,
			Message: err.Error(),
		}, err
	}

	logger.Info("Secret fetched successfully from Vault", "name", input.Name)
	return output, nil
}
