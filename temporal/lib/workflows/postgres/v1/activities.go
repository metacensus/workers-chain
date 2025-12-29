package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	pb "github.com/metacensus/infra/temporal/gen/workflows/postgres/v1"
	"go.temporal.io/sdk/activity"
)

// Activities implements the PostgresActivities interface from generated code
type Activities struct {
	client PostgresClient
}

// NewActivities creates a new Activities instance
func NewActivities(client PostgresClient) *Activities {
	return &Activities{
		client: client,
	}
}

// CreateDatabase creates a new PostgreSQL database
// Idempotent: returns success if database already exists
func (a *Activities) CreateDatabase(ctx context.Context, req *pb.CreateDatabaseInput) (*pb.CreateDatabaseOutput, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Creating database",
		"database_name", req.DatabaseName)

	// Validate required fields
	if req.DatabaseName == "" {
		return nil, errors.New("database name cannot be empty")
	}

	// Check if database already exists
	exists, err := a.databaseExists(ctx, a.client, req.DatabaseName)
	if err != nil {
		logger.Error("Failed to check if database exists", "error", err)
		return nil, fmt.Errorf("check database existence: %w", err)
	}

	if exists {
		logger.Info("Database already exists", "database_name", req.DatabaseName)
		return &pb.CreateDatabaseOutput{}, nil
	}

	// Create the database
	// PostgreSQL doesn't support IF NOT EXISTS for CREATE DATABASE, so we check first
	// We need to escape the database name to prevent SQL injection
	escapedName := escapeIdentifier(req.DatabaseName)
	createSQL := fmt.Sprintf("CREATE DATABASE %s", escapedName)

	if err := a.client.Exec(ctx, createSQL); err != nil {
		// Check if error is because database already exists (race condition)
		if strings.Contains(err.Error(), "already exists") {
			logger.Info("Database already exists (race condition)", "database_name", req.DatabaseName)
			return &pb.CreateDatabaseOutput{}, nil
		}
		logger.Error("Failed to create database", "error", err)
		return nil, fmt.Errorf("create database: %w", err)
	}

	logger.Info("Database created successfully", "database_name", req.DatabaseName)
	return &pb.CreateDatabaseOutput{}, nil
}

// CreateUser creates a new PostgreSQL user
// Idempotent: returns success if user already exists
func (a *Activities) CreateUser(ctx context.Context, req *pb.CreateUserInput) (*pb.CreateUserOutput, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Creating user",
		"username", req.Username)

	// Validate required fields
	if req.Username == "" {
		return nil, errors.New("username cannot be empty")
	}
	if req.Password == "" {
		return nil, errors.New("password cannot be empty")
	}

	// Check if user already exists
	exists, err := a.userExists(ctx, a.client, req.Username)
	if err != nil {
		logger.Error("Failed to check if user exists", "error", err)
		return nil, fmt.Errorf("check user existence: %w", err)
	}

	if exists {
		logger.Info("User already exists", "username", req.Username)
		return &pb.CreateUserOutput{}, nil
	}

	// Create the user (PostgreSQL uses CREATE ROLE)
	// We need to escape identifiers and quote the password
	escapedUsername := escapeIdentifier(req.Username)
	escapedPassword := escapeStringLiteral(req.Password)
	createSQL := fmt.Sprintf("CREATE ROLE %s WITH LOGIN PASSWORD %s", escapedUsername, escapedPassword)

	if err := a.client.Exec(ctx, createSQL); err != nil {
		// Check if error is because user already exists (race condition)
		if strings.Contains(err.Error(), "already exists") {
			logger.Info("User already exists (race condition)", "username", req.Username)
			return &pb.CreateUserOutput{}, nil
		}
		logger.Error("Failed to create user", "error", err)
		return nil, fmt.Errorf("create user: %w", err)
	}

	logger.Info("User created successfully", "username", req.Username)
	return &pb.CreateUserOutput{}, nil
}

// GrantPrivileges grants privileges to a user on a database
// Idempotent: can be called multiple times safely
func (a *Activities) GrantPrivileges(ctx context.Context, req *pb.GrantPrivilegesInput) (*pb.GrantPrivilegesOutput, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Granting privileges",
		"username", req.Username,
		"database_name", req.DatabaseName,
		"privileges", req.Privileges)

	// Validate required fields
	if req.Username == "" {
		return nil, errors.New("username cannot be empty")
	}
	if req.DatabaseName == "" {
		return nil, errors.New("database name cannot be empty")
	}

	// Default to ALL if privileges not specified
	privileges := req.Privileges
	if privileges == "" {
		privileges = "ALL"
	}

	// Grant privileges on the database
	// We need to escape identifiers
	escapedUsername := escapeIdentifier(req.Username)
	escapedDatabaseName := escapeIdentifier(req.DatabaseName)
	grantSQL := fmt.Sprintf("GRANT %s ON DATABASE %s TO %s", privileges, escapedDatabaseName, escapedUsername)

	if err := a.client.Exec(ctx, grantSQL); err != nil {
		logger.Error("Failed to grant privileges", "error", err)
		return nil, fmt.Errorf("grant privileges: %w", err)
	}

	logger.Info("Privileges granted successfully",
		"username", req.Username,
		"database_name", req.DatabaseName,
		"privileges", privileges)

	return &pb.GrantPrivilegesOutput{}, nil
}

// databaseExists checks if a database exists
func (a *Activities) databaseExists(ctx context.Context, client PostgresClient, dbName string) (bool, error) {
	// Query pg_database to check if database exists
	query := "SELECT 1 FROM pg_database WHERE datname = $1"
	rows, err := client.Query(ctx, query, dbName)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	return rows.Next(), nil
}

// userExists checks if a user (role) exists
func (a *Activities) userExists(ctx context.Context, client PostgresClient, username string) (bool, error) {
	// Query pg_roles to check if role exists
	query := "SELECT 1 FROM pg_roles WHERE rolname = $1"
	rows, err := client.Query(ctx, query, username)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	return rows.Next(), nil
}

// escapeIdentifier escapes a PostgreSQL identifier to prevent SQL injection
// PostgreSQL identifiers are case-insensitive unless quoted, and should be quoted if they contain special characters
func escapeIdentifier(identifier string) string {
	// For simplicity, we'll quote all identifiers
	// Replace double quotes with double double quotes (PostgreSQL escaping)
	escaped := strings.ReplaceAll(identifier, `"`, `""`)
	return fmt.Sprintf(`"%s"`, escaped)
}

// escapeStringLiteral escapes a PostgreSQL string literal to prevent SQL injection
func escapeStringLiteral(s string) string {
	// Replace single quotes with two single quotes (PostgreSQL escaping)
	escaped := strings.ReplaceAll(s, `'`, `''`)
	return fmt.Sprintf(`'%s'`, escaped)
}
