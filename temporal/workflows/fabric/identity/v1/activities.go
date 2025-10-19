package identity

import (
	"context"
	"fmt"

	pb "github.com/metacensus/infra/temporal/gen/workflows/fabric/identity/v1"
	"go.temporal.io/sdk/activity"
)

// Activities implements the CertificateActivities interface from generated code
type Activities struct {
	client FabricCAClient
}

// NewActivities creates a new Activities instance
func NewActivities(client FabricCAClient) *Activities {
	return &Activities{
		client: client,
	}
}

// RegisterIdentity registers a new identity with Fabric CA
func (a *Activities) RegisterIdentity(ctx context.Context, input *pb.RegisterIdentityInput) (*pb.RegisterIdentityOutput, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Registering identity",
		"name", input.GetName(),
		"type", input.GetType(),
		"affiliation", input.GetAffiliation())

	result, err := a.client.Register(ctx, input)
	if err != nil {
		logger.Error("Failed to register identity", "error", err)
		return nil, fmt.Errorf("fabric CA client error: %w", err)
	}

	logger.Info("Identity registered successfully",
		"name", result.GetName(),
		"type", result.GetType(),
		"secret_generated", result.GetSecret() != "")

	return result, nil
}

// EnrollIdentity enrolls an identity and retrieves certificates
func (a *Activities) EnrollIdentity(ctx context.Context, input *pb.EnrollIdentityInput) (*pb.EnrollIdentityOutput, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Enrolling identity",
		"name", input.GetName(),
		"profile", input.GetProfile())

	result, err := a.client.Enroll(ctx, input)
	if err != nil {
		logger.Error("Failed to enroll identity", "error", err)
		return nil, fmt.Errorf("fabric CA client error: %w", err)
	}

	logger.Info("Identity enrolled successfully",
		"name", input.GetName(),
		"cert_present", len(result.GetCertificate()) > 0,
		"key_present", len(result.GetPrivateKey()) > 0)

	return result, nil
}

// AssembleMsp assembles an MSP directory structure from enrollment results
func (a *Activities) AssembleMsp(ctx context.Context, input *pb.AssembleMspInput) (*pb.AssembleMspOutput, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Assembling MSP",
		"signing_cert_present", len(input.GetSigningEnrollment().GetCertificate()) > 0,
		"tls_cert_present", len(input.GetTlsEnrollment().GetCertificate()) > 0)

	// This is a pure data transformation activity (no external calls needed)
	result := &pb.AssembleMspOutput{
		SigningCert:  input.GetSigningEnrollment().GetCertificate(),
		SigningKey:   input.GetSigningEnrollment().GetPrivateKey(),
		TlsCert:      input.GetTlsEnrollment().GetCertificate(),
		TlsKey:       input.GetTlsEnrollment().GetPrivateKey(),
		CaCerts:      input.GetCaCerts(),
		NodeOuConfig: input.GetNodeOuConfig(),
	}

	logger.Info("MSP assembled successfully")
	return result, nil
}

// GenerateSecureSecret generates a cryptographically secure random secret
func (a *Activities) GenerateSecureSecret(ctx context.Context, input *pb.GenerateSecureSecretInput) (*pb.GenerateSecureSecretOutput, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Generating secure secret", "length", input.GetLength())

	// Use client method for random secret generation
	// (The client can implement this using crypto/rand or delegate to an external service)
	secret, err := a.client.GenerateSecret(input.GetLength())
	if err != nil {
		logger.Error("Failed to generate secure secret", "error", err)
		return nil, fmt.Errorf("failed to generate secret: %w", err)
	}

	logger.Info("Secure secret generated successfully")
	return &pb.GenerateSecureSecretOutput{Secret: secret}, nil
}

// CreateAdminIdentity creates a new admin identity using the default admin credentials
// This is a multi-step process:
// 1. Enroll as default admin to get an identity for registration
// 2. Register the new admin using the default admin identity
// 3. Enroll the new admin to get certificates
func (a *Activities) CreateAdminIdentity(ctx context.Context, input *pb.CreateAdminIdentityInput) (*pb.CreateAdminIdentityOutput, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Creating admin identity",
		"admin_id", input.GetAdminId(),
		"default_admin_id", input.GetDefaultAdminId())

	// Step 1: Enroll as default admin
	// Note: The Fabric CA client internally maintains the enrolled identity for subsequent operations
	_, err := a.client.Enroll(ctx, &pb.EnrollIdentityInput{
		Name:   input.GetDefaultAdminId(),
		Secret: input.GetDefaultSecret(),
	})
	if err != nil {
		logger.Error("Failed to enroll as default admin", "error", err)
		return nil, fmt.Errorf("failed to enroll as default admin: %w", err)
	}

	// Step 2: Register new admin identity
	// Build attributes for admin role
	attributes := input.GetAttributes()
	// Add admin-specific attributes if not present
	attributes = append(attributes,
		&pb.Attribute{Name: "hf.Registrar.Roles", Value: "admin", Ecert: true},
		&pb.Attribute{Name: "hf.Registrar.Attributes", Value: "*", Ecert: true},
	)

	registration, err := a.client.Register(ctx, &pb.RegisterIdentityInput{
		Name:           input.GetAdminId(),
		Type:           "client", // Admin is type "client" in Fabric CA
		Secret:         input.GetAdminSecret(),
		Affiliation:    "", // Root affiliation
		Attributes:     attributes,
		MaxEnrollments: -1, // Unlimited enrollments for admin
	})
	if err != nil {
		logger.Error("Failed to register admin identity", "error", err)
		return nil, fmt.Errorf("failed to register admin: %w", err)
	}

	// Step 3: Enroll the new admin to get certificates
	enrollment, err := a.client.Enroll(ctx, &pb.EnrollIdentityInput{
		Name:   registration.GetName(),
		Secret: registration.GetSecret(),
	})
	if err != nil {
		logger.Error("Failed to enroll admin identity", "error", err)
		return nil, fmt.Errorf("failed to enroll admin: %w", err)
	}

	// Build AdminIdentity result
	adminIdentity := &pb.AdminIdentity{
		Id:          input.GetAdminId(),
		Secret:      registration.GetSecret(),
		Certificate: enrollment.GetCertificate(),
		PrivateKey:  enrollment.GetPrivateKey(),
		CaCert:      enrollment.GetCaCert(),
		Attributes:  attributes,
	}

	logger.Info("Admin identity created successfully", "admin_id", adminIdentity.GetId())
	return &pb.CreateAdminIdentityOutput{
		AdminIdentity: adminIdentity,
	}, nil
}

// RevokeDefaultAdmin revokes the default admin identity
func (a *Activities) RevokeDefaultAdmin(ctx context.Context, input *pb.RevokeDefaultAdminInput) (*pb.RevokeDefaultAdminOutput, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Revoking default admin", "admin_id", input.GetDefaultAdminId())

	err := a.client.Revoke(ctx, input.GetDefaultAdminId())
	if err != nil {
		logger.Error("Failed to revoke default admin", "error", err)
		return nil, fmt.Errorf("fabric CA client error: %w", err)
	}

	logger.Info("Default admin revoked successfully")
	return &pb.RevokeDefaultAdminOutput{Success: true, Message: "Default admin revoked"}, nil
}

// CreateAffiliations creates multiple affiliations in Fabric CA
func (a *Activities) CreateAffiliations(ctx context.Context, input *pb.CreateAffiliationsInput) (*pb.CreateAffiliationsOutput, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Creating affiliations",
		"count", len(input.GetAffiliations()),
		"admin_id", input.GetAdminIdentity().GetId())

	for _, name := range input.GetAffiliations() {
		err := a.client.CreateAffiliation(ctx, name)
		if err != nil {
			logger.Error("Failed to create affiliation", "name", name, "error", err)
			return nil, fmt.Errorf("fabric CA client error for %q: %w", name, err)
		}
		logger.Info("Affiliation created", "name", name)
	}

	logger.Info("All affiliations created successfully")
	return &pb.CreateAffiliationsOutput{}, nil
}
