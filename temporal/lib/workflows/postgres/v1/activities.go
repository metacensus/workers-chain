package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	pb "github.com/metacensus/infra/temporal/gen/workflows/postgres/v1"
	secrets "github.com/metacensus/infra/temporal/lib/workflows/secrets/v1"
	"go.temporal.io/sdk/activity"
)

// Sentinel errors for postgres activities
var (
	ErrCheckDatabaseExistence = errors.New("check database existence")
	ErrCreateDatabase         = errors.New("create database")
	ErrCheckUserExistence     = errors.New("check user existence")
	ErrCreateUser             = errors.New("create user")
	ErrGrantPrivileges        = errors.New("grant privileges")
	ErrFetchCredentials       = errors.New("fetch credentials")
)

// ClientProvider is an interface for getting postgres clients
type ClientProvider interface {
	GetClient(ctx context.Context) (PostgresClient, error)
}

// Activities implements the PostgresActivities interface
type Activities struct {
	clientProvider ClientProvider
	secretProvider secrets.SecretProvider
}

// NewActivities creates a new Activities instance
func NewActivities(clientProvider ClientProvider, secretProvider secrets.SecretProvider) *Activities {
	return &Activities{
		clientProvider: clientProvider,
		secretProvider: secretProvider,
	}
}

// CreateDatabase creates a new PostgreSQL database
func (a *Activities) CreateDatabase(ctx context.Context, req *pb.CreateDatabaseInput) (*pb.CreateDatabaseOutput, error) {
	logger := activity.GetLogger(ctx)

	if req.DatabaseName == "" {
		return nil, errors.New("database name cannot be empty")
	}

	client, err := a.clientProvider.GetClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get postgres client: %w", err)
	}

	exists, err := a.databaseExists(ctx, client, req.DatabaseName)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCheckDatabaseExistence, err)
	}

	if exists {
		logger.Info("Database already exists", "database_name", req.DatabaseName)
		return &pb.CreateDatabaseOutput{}, nil
	}

	escapedName := escapeIdentifier(req.DatabaseName)
	createSQL := fmt.Sprintf("CREATE DATABASE %s", escapedName)

	if err := client.Exec(ctx, createSQL); err != nil {
		if strings.Contains(err.Error(), "already exists") {
			logger.Info("Database already exists (race condition)", "database_name", req.DatabaseName)
			return &pb.CreateDatabaseOutput{}, nil
		}
		return nil, fmt.Errorf("%w: %w", ErrCreateDatabase, err)
	}

	return &pb.CreateDatabaseOutput{}, nil
}

// CreateUser creates a new PostgreSQL user with credentials fetched from SecretProvider
func (a *Activities) CreateUser(ctx context.Context, req *pb.CreateUserInput) (*pb.CreateUserOutput, error) {
	logger := activity.GetLogger(ctx)

	if req.CredentialsRef == nil {
		return nil, errors.New("credentials_ref is required")
	}

	// Fetch credentials from secret provider
	secret, err := a.secretProvider.FetchSecret(ctx, req.CredentialsRef)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFetchCredentials, err)
	}

	creds, err := secret.Credentials()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFetchCredentials, err)
	}

	client, err := a.clientProvider.GetClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get postgres client: %w", err)
	}

	exists, err := a.userExists(ctx, client, creds.Username)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCheckUserExistence, err)
	}

	if exists {
		logger.Info("User already exists", "username", creds.Username)
		return &pb.CreateUserOutput{}, nil
	}

	escapedUsername := escapeIdentifier(creds.Username)
	escapedPassword := escapeStringLiteral(creds.Password)
	createSQL := fmt.Sprintf("CREATE ROLE %s WITH LOGIN PASSWORD %s", escapedUsername, escapedPassword)

	if err := client.Exec(ctx, createSQL); err != nil {
		if strings.Contains(err.Error(), "already exists") {
			logger.Info("User already exists (race condition)", "username", creds.Username)
			return &pb.CreateUserOutput{}, nil
		}
		return nil, fmt.Errorf("%w: %w", ErrCreateUser, err)
	}

	return &pb.CreateUserOutput{}, nil
}

// GrantPrivileges grants privileges to a user on a database
func (a *Activities) GrantPrivileges(ctx context.Context, req *pb.GrantPrivilegesInput) (*pb.GrantPrivilegesOutput, error) {
	if req.Username == "" {
		return nil, errors.New("username cannot be empty")
	}
	if req.DatabaseName == "" {
		return nil, errors.New("database name cannot be empty")
	}

	client, err := a.clientProvider.GetClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get postgres client: %w", err)
	}

	escapedUsername := escapeIdentifier(req.Username)
	escapedDatabaseName := escapeIdentifier(req.DatabaseName)

	alterOwnerSQL := fmt.Sprintf("ALTER DATABASE %s OWNER TO %s", escapedDatabaseName, escapedUsername)
	if err := client.Exec(ctx, alterOwnerSQL); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrGrantPrivileges, err)
	}

	return &pb.GrantPrivilegesOutput{}, nil
}

func (a *Activities) databaseExists(ctx context.Context, client PostgresClient, dbName string) (bool, error) {
	query := "SELECT 1 FROM pg_database WHERE datname = $1"
	rows, err := client.Query(ctx, query, dbName)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	return rows.Next(), nil
}

func (a *Activities) userExists(ctx context.Context, client PostgresClient, username string) (bool, error) {
	query := "SELECT 1 FROM pg_roles WHERE rolname = $1"
	rows, err := client.Query(ctx, query, username)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	return rows.Next(), nil
}

func escapeIdentifier(identifier string) string {
	escaped := strings.ReplaceAll(identifier, `"`, `""`)
	return fmt.Sprintf(`"%s"`, escaped)
}

func escapeStringLiteral(s string) string {
	escaped := strings.ReplaceAll(s, `'`, `''`)
	return fmt.Sprintf(`'%s'`, escaped)
}

// ConvertPrivilegesToSQL converts Privilege enums to PostgreSQL SQL privilege strings
func ConvertPrivilegesToSQL(privileges []pb.Privilege) string {
	if len(privileges) == 0 {
		return "ALL"
	}

	var sqlPrivileges []string
	for _, priv := range privileges {
		switch priv {
		case pb.Privilege_PRIVILEGE_UNSPECIFIED:
			continue
		case pb.Privilege_PRIVILEGE_ALL:
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

	if len(sqlPrivileges) == 0 {
		return "ALL"
	}

	return strings.Join(sqlPrivileges, ", ")
}
