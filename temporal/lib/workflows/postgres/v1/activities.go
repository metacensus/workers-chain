package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	pb "github.com/metacensus/infra/temporal/gen/workflows/postgres/v1"
	"go.temporal.io/sdk/activity"
)

// Sentinel errors for postgres activities
var (
	ErrCheckDatabaseExistence = errors.New("check database existence")
	ErrCreateDatabase         = errors.New("create database")
	ErrCheckUserExistence     = errors.New("check user existence")
	ErrCreateUser             = errors.New("create user")
	ErrGrantPrivileges        = errors.New("grant privileges")
)

// ClientProvider is an interface for getting postgres clients
// This allows for easier testing by injecting mock providers
type ClientProvider interface {
	GetClient(ctx context.Context) (PostgresClient, error)
}

// Activities implements the PostgresActivities interface from generated code
type Activities struct {
	clientProvider ClientProvider
}

// NewActivities creates a new Activities instance
func NewActivities(clientProvider ClientProvider) *Activities {
	return &Activities{
		clientProvider: clientProvider,
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

	// Get postgres client (fetches credentials from Vault on first call)
	client, err := a.clientProvider.GetClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get postgres client: %w", err)
	}

	// Check if database already exists
	exists, err := a.databaseExists(ctx, client, req.DatabaseName)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCheckDatabaseExistence, err)
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

	if err := client.Exec(ctx, createSQL); err != nil {
		// Check if error is because database already exists (race condition)
		if strings.Contains(err.Error(), "already exists") {
			logger.Info("Database already exists (race condition)", "database_name", req.DatabaseName)
			return &pb.CreateDatabaseOutput{}, nil
		}
		return nil, fmt.Errorf("%w: %w", ErrCreateDatabase, err)
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

	// Get postgres client (fetches credentials from Vault on first call)
	client, err := a.clientProvider.GetClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get postgres client: %w", err)
	}

	// Check if user already exists
	exists, err := a.userExists(ctx, client, req.Username)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCheckUserExistence, err)
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

	if err := client.Exec(ctx, createSQL); err != nil {
		// Check if error is because user already exists (race condition)
		if strings.Contains(err.Error(), "already exists") {
			logger.Info("User already exists (race condition)", "username", req.Username)
			return &pb.CreateUserOutput{}, nil
		}
		return nil, fmt.Errorf("%w: %w", ErrCreateUser, err)
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

	// Get postgres client (fetches credentials from Vault on first call)
	client, err := a.clientProvider.GetClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get postgres client: %w", err)
	}

	// Convert enum privileges to PostgreSQL privilege strings
	privilegesStr := ConvertPrivilegesToSQL(req.Privileges)

	// Grant privileges on the database
	// We need to escape identifiers
	escapedUsername := escapeIdentifier(req.Username)
	escapedDatabaseName := escapeIdentifier(req.DatabaseName)
	grantSQL := fmt.Sprintf("GRANT %s ON DATABASE %s TO %s", privilegesStr, escapedDatabaseName, escapedUsername)

	if err := client.Exec(ctx, grantSQL); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrGrantPrivileges, err)
	}

	logger.Info("Privileges granted successfully",
		"username", req.Username,
		"database_name", req.DatabaseName,
		"privileges", privilegesStr)

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

// ConvertPrivilegesToSQL converts Privilege enums to PostgreSQL SQL privilege strings
func ConvertPrivilegesToSQL(privileges []pb.Privilege) string {
	// Default to ALL if no privileges specified or only UNSPECIFIED
	if len(privileges) == 0 {
		return "ALL"
	}

	// Filter out UNSPECIFIED and convert to SQL strings
	var sqlPrivileges []string
	for _, priv := range privileges {
		switch priv {
		case pb.Privilege_PRIVILEGE_UNSPECIFIED:
			// Skip unspecified
			continue
		case pb.Privilege_PRIVILEGE_ALL:
			// ALL is a special case - if present, just return ALL
			return "ALL"
		case pb.Privilege_PRIVILEGE_SELECT:
			sqlPrivileges = append(sqlPrivileges, "SELECT")
		case pb.Privilege_PRIVILEGE_INSERT:
			sqlPrivileges = append(sqlPrivileges, "INSERT")
		case pb.Privilege_PRIVILEGE_UPDATE:
			sqlPrivileges = append(sqlPrivileges, "UPDATE")
		case pb.Privilege_PRIVILEGE_DELETE:
			sqlPrivileges = append(sqlPrivileges, "DELETE")
		case pb.Privilege_PRIVILEGE_TRUNCATE:
			sqlPrivileges = append(sqlPrivileges, "TRUNCATE")
		case pb.Privilege_PRIVILEGE_REFERENCES:
			sqlPrivileges = append(sqlPrivileges, "REFERENCES")
		case pb.Privilege_PRIVILEGE_TRIGGER:
			sqlPrivileges = append(sqlPrivileges, "TRIGGER")
		case pb.Privilege_PRIVILEGE_CREATE:
			sqlPrivileges = append(sqlPrivileges, "CREATE")
		case pb.Privilege_PRIVILEGE_CONNECT:
			sqlPrivileges = append(sqlPrivileges, "CONNECT")
		case pb.Privilege_PRIVILEGE_TEMPORARY:
			sqlPrivileges = append(sqlPrivileges, "TEMPORARY")
		case pb.Privilege_PRIVILEGE_EXECUTE:
			sqlPrivileges = append(sqlPrivileges, "EXECUTE")
		case pb.Privilege_PRIVILEGE_USAGE:
			sqlPrivileges = append(sqlPrivileges, "USAGE")
		}
	}

	// If all were UNSPECIFIED, default to ALL
	if len(sqlPrivileges) == 0 {
		return "ALL"
	}

	// Join multiple privileges with commas
	return strings.Join(sqlPrivileges, ", ")
}
