package crypto

import (
	"errors"
	"fmt"
	"time"

	cryptopb "github.com/metacensus/infra/temporal/gen/workflows/crypto/v1"
	vaultpb "github.com/metacensus/infra/temporal/gen/workflows/vault/v1"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// Sentinel errors for crypto workflows
var (
	ErrGenerateAndStorePassword = errors.New("generate and store password workflow failed")
	ErrGenerateAndStoreKeyPair  = errors.New("generate and store key pair workflow failed")
)

// Workflows implements the CryptoWorkflows interface from generated code
type Workflows struct{}

// NewWorkflows creates a new Workflows instance
func NewWorkflows() *Workflows {
	return &Workflows{}
}

// GenerateAndStorePassword implements the GenerateAndStorePassword workflow
func (w *Workflows) GenerateAndStorePassword(ctx workflow.Context, input *cryptopb.GenerateAndStorePasswordWorkflowInput) (cryptopb.GenerateAndStorePasswordWorkflow, error) {
	return &generateAndStorePasswordWorkflow{input: input}, nil
}

// GenerateAndStoreKeyPair implements the GenerateAndStoreKeyPair workflow
func (w *Workflows) GenerateAndStoreKeyPair(ctx workflow.Context, input *cryptopb.GenerateAndStoreKeyPairWorkflowInput) (cryptopb.GenerateAndStoreKeyPairWorkflow, error) {
	return &generateAndStoreKeyPairWorkflow{input: input}, nil
}

// generateAndStorePasswordWorkflow implements cryptopb.GenerateAndStorePasswordWorkflow
type generateAndStorePasswordWorkflow struct {
	input *cryptopb.GenerateAndStorePasswordWorkflowInput
}

// Execute orchestrates password generation and storage in Vault
func (w *generateAndStorePasswordWorkflow) Execute(ctx workflow.Context) (*cryptopb.GenerateAndStorePasswordOutput, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("GenerateAndStorePassword workflow started", "vault_path", w.input.Req.GetVaultPath())

	// Set up activity options
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    30 * time.Second,
			MaximumAttempts:    3,
		},
	}
	activityCtx := workflow.WithActivityOptions(ctx, ao)

	// Step 1: Generate password
	logger.Info("Generating password")
	var passwordOutput *cryptopb.GeneratePasswordOutput
	err := workflow.ExecuteActivity(activityCtx, cryptopb.GeneratePasswordActivityName,
		&cryptopb.GeneratePasswordInput{}).Get(ctx, &passwordOutput)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to generate password: %w", ErrGenerateAndStorePassword, err)
	}

	// Step 2: Store password in Vault with key name "store-password"
	logger.Info("Storing password in Vault", "vault_path", w.input.Req.GetVaultPath())
	var storeOutput *vaultpb.StoreSecretOutput
	err = workflow.ExecuteActivity(activityCtx, vaultpb.StoreSecretActivityName,
		&vaultpb.StoreSecretInput{
			Path: w.input.Req.GetVaultPath(),
			Data: map[string][]byte{
				"store-password": []byte(passwordOutput.GetPassword()),
			},
		}).Get(ctx, &storeOutput)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to store password: %w", ErrGenerateAndStorePassword, err)
	}

	logger.Info("GenerateAndStorePassword workflow completed successfully",
		"vault_path", storeOutput.SecretRef.GetPath(),
		"version", storeOutput.SecretRef.GetVersion())

	return &cryptopb.GenerateAndStorePasswordOutput{
		SecretRef: storeOutput.GetSecretRef(),
	}, nil
}

// generateAndStoreKeyPairWorkflow implements cryptopb.GenerateAndStoreKeyPairWorkflow
type generateAndStoreKeyPairWorkflow struct {
	input *cryptopb.GenerateAndStoreKeyPairWorkflowInput
}

// Execute orchestrates key pair generation and storage in Vault
func (w *generateAndStoreKeyPairWorkflow) Execute(ctx workflow.Context) (*cryptopb.GenerateAndStoreKeyPairOutput, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("GenerateAndStoreKeyPair workflow started", "vault_path", w.input.Req.GetVaultPath())

	// Set up activity options
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    30 * time.Second,
			MaximumAttempts:    3,
		},
	}
	activityCtx := workflow.WithActivityOptions(ctx, ao)

	// Step 1: Generate key pair
	logger.Info("Generating ECDSA key pair")
	var keyPairOutput *cryptopb.GenerateKeyPairOutput
	err := workflow.ExecuteActivity(activityCtx, cryptopb.GenerateKeyPairActivityName,
		&cryptopb.GenerateKeyPairInput{}).Get(ctx, &keyPairOutput)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to generate key pair: %w", ErrGenerateAndStoreKeyPair, err)
	}

	// Step 2: Store both keys in Vault with key names "private" and "public"
	logger.Info("Storing key pair in Vault", "vault_path", w.input.Req.GetVaultPath())
	var storeOutput *vaultpb.StoreSecretOutput
	err = workflow.ExecuteActivity(activityCtx, vaultpb.StoreSecretActivityName,
		&vaultpb.StoreSecretInput{
			Path: w.input.Req.GetVaultPath(),
			Data: map[string][]byte{
				"private": keyPairOutput.GetPrivateKey(),
				"public":  keyPairOutput.GetPublicKey(),
			},
		}).Get(ctx, &storeOutput)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to store key pair: %w", ErrGenerateAndStoreKeyPair, err)
	}

	logger.Info("GenerateAndStoreKeyPair workflow completed successfully",
		"vault_path", storeOutput.SecretRef.GetPath(),
		"version", storeOutput.SecretRef.GetVersion())

	return &cryptopb.GenerateAndStoreKeyPairOutput{
		SecretRef: storeOutput.GetSecretRef(),
	}, nil
}

// TestGenerateAndStorePasswordWorkflow is a test helper function
func TestGenerateAndStorePasswordWorkflow(ctx workflow.Context, input *cryptopb.GenerateAndStorePasswordWorkflowInput) (*cryptopb.GenerateAndStorePasswordOutput, error) {
	w := &generateAndStorePasswordWorkflow{input: input}
	return w.Execute(ctx)
}

// TestGenerateAndStoreKeyPairWorkflow is a test helper function
func TestGenerateAndStoreKeyPairWorkflow(ctx workflow.Context, input *cryptopb.GenerateAndStoreKeyPairWorkflowInput) (*cryptopb.GenerateAndStoreKeyPairOutput, error) {
	w := &generateAndStoreKeyPairWorkflow{input: input}
	return w.Execute(ctx)
}

