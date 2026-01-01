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
	ErrGenerateAndStorePassword    = errors.New("generate and store password workflow failed")
	ErrGenerateAndStoreKeyPair     = errors.New("generate and store key pair workflow failed")
	ErrCreateSelfSignedCertificate = errors.New("create self-signed certificate workflow failed")
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

// CreateSelfSignedCertificate implements the CreateSelfSignedCertificate workflow
func (w *Workflows) CreateSelfSignedCertificate(ctx workflow.Context, input *cryptopb.CreateSelfSignedCertificateWorkflowInput) (cryptopb.CreateSelfSignedCertificateWorkflow, error) {
	return &createSelfSignedCertificateWorkflow{input: input}, nil
}

// generateAndStorePasswordWorkflow implements cryptopb.GenerateAndStorePasswordWorkflow
type generateAndStorePasswordWorkflow struct {
	input *cryptopb.GenerateAndStorePasswordWorkflowInput
}

// Execute orchestrates password generation and storage in Vault along with username
func (w *generateAndStorePasswordWorkflow) Execute(ctx workflow.Context) (*cryptopb.GenerateAndStorePasswordOutput, error) {
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
	var passwordOutput *cryptopb.GeneratePasswordOutput
	err := workflow.ExecuteActivity(activityCtx, cryptopb.GeneratePasswordActivityName,
		&cryptopb.GeneratePasswordInput{}).Get(ctx, &passwordOutput)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to generate password: %w", ErrGenerateAndStorePassword, err)
	}

	// Step 2: Store username and password in Vault with key names "username" and "password"
	var storeOutput *vaultpb.StoreSecretOutput
	err = workflow.ExecuteActivity(activityCtx, vaultpb.StoreSecretActivityName,
		&vaultpb.StoreSecretInput{
			Path: w.input.Req.GetVaultPath(),
			Data: map[string][]byte{
				"username": []byte(w.input.Req.GetUsername()),
				"password": []byte(passwordOutput.GetPassword()),
			},
		}).Get(ctx, &storeOutput)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to store username and password: %w", ErrGenerateAndStorePassword, err)
	}

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
	var keyPairOutput *cryptopb.GenerateKeyPairOutput
	err := workflow.ExecuteActivity(activityCtx, cryptopb.GenerateKeyPairActivityName,
		&cryptopb.GenerateKeyPairInput{}).Get(ctx, &keyPairOutput)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to generate key pair: %w", ErrGenerateAndStoreKeyPair, err)
	}

	// Step 2: Store both keys in Vault with key names "private" and "public"
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

// createSelfSignedCertificateWorkflow implements cryptopb.CreateSelfSignedCertificateWorkflow
type createSelfSignedCertificateWorkflow struct {
	input *cryptopb.CreateSelfSignedCertificateWorkflowInput
}

// Execute orchestrates self-signed certificate generation and storage in Vault
func (w *createSelfSignedCertificateWorkflow) Execute(ctx workflow.Context) (*cryptopb.CreateSelfSignedCertificateOutput, error) {
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

	// Step 1: Fetch private key from Vault
	var fetchOutput *vaultpb.FetchSecretOutput
	err := workflow.ExecuteActivity(activityCtx, vaultpb.FetchSecretActivityName,
		&vaultpb.FetchSecretInput{
			Path: w.input.Req.GetVaultKeyPath(),
		}).Get(ctx, &fetchOutput)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to fetch private key: %w", ErrCreateSelfSignedCertificate, err)
	}

	// Extract private key from the fetched secret data
	// Keys are stored with key name "private" by GenerateAndStoreKeyPair
	privateKeyBytes, ok := fetchOutput.GetData()["private"]
	if !ok {
		return nil, fmt.Errorf("%w: private key not found in vault secret", ErrCreateSelfSignedCertificate)
	}

	// Step 2: Generate self-signed certificate
	var certOutput *cryptopb.GenerateCertificateOutput
	err = workflow.ExecuteActivity(activityCtx, cryptopb.GenerateCertificateActivityName,
		&cryptopb.GenerateCertificateInput{
			PrivateKey:         privateKeyBytes,
			CommonName:         w.input.Req.GetCommonName(),
			Organization:       w.input.Req.GetOrganization(),
			OrganizationalUnit: w.input.Req.GetOrganizationalUnit(),
			ValidityDuration:   w.input.Req.GetValidityDuration(),
			MaxPathLength:      w.input.Req.GetMaxPathLength(),
		}).Get(ctx, &certOutput)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to generate certificate: %w", ErrCreateSelfSignedCertificate, err)
	}

	// Step 3: Store certificate in Vault
	var storeOutput *vaultpb.StoreSecretOutput
	err = workflow.ExecuteActivity(activityCtx, vaultpb.StoreSecretActivityName,
		&vaultpb.StoreSecretInput{
			Path: w.input.Req.GetVaultCertificatePath(),
			Data: map[string][]byte{
				"certificate": certOutput.GetCertificate(),
			},
		}).Get(ctx, &storeOutput)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to store certificate: %w", ErrCreateSelfSignedCertificate, err)
	}

	return &cryptopb.CreateSelfSignedCertificateOutput{
		SecretRef: storeOutput.GetSecretRef(),
	}, nil
}

// TestCreateSelfSignedCertificateWorkflow is a test helper function
func TestCreateSelfSignedCertificateWorkflow(ctx workflow.Context, input *cryptopb.CreateSelfSignedCertificateWorkflowInput) (*cryptopb.CreateSelfSignedCertificateOutput, error) {
	w := &createSelfSignedCertificateWorkflow{input: input}
	return w.Execute(ctx)
}
