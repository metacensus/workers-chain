package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// Sentinel errors for PostgreSQL operations
var (
	ErrDatabaseExists = errors.New("postgres: database already exists")
	ErrUserExists     = errors.New("postgres: user already exists")
	ErrNotFound       = errors.New("postgres: not found")
)

// PostgresConfig represents PostgreSQL connection configuration
type PostgresConfig struct {
	ConnectionString string // PostgreSQL connection string (e.g., "postgres://user:password@host:port/database?sslmode=disable")
}

// PostgresClient defines the interface for PostgreSQL operations
// This interface uses type aliases to pull pgx interfaces into the package for mocking
type PostgresClient interface {
	// Exec executes a SQL statement that doesn't return rows
	Exec(ctx context.Context, sql string, arguments ...interface{}) error

	// Query executes a SQL query that returns rows
	Query(ctx context.Context, sql string, arguments ...interface{}) (pgx.Rows, error)

	// QueryRow executes a SQL query that returns a single row
	QueryRow(ctx context.Context, sql string, arguments ...interface{}) pgx.Row

	// Begin starts a transaction
	Begin(ctx context.Context) (pgx.Tx, error)

	// Close closes the connection
	Close(ctx context.Context) error
}

// Type alias for pgx.Conn to make it mockable
type Conn = pgx.Conn

// Type alias for pgx.Rows to make it mockable
type Rows = pgx.Rows

// Type alias for pgx.Row to make it mockable
type Row = pgx.Row

// Type alias for pgx.Tx to make it mockable
type Tx = pgx.Tx

// postgresClient implements PostgresClient using pgx
type postgresClient struct {
	conn *pgx.Conn
}

// NewPostgresClient creates a new PostgreSQL client with the given connection string
func NewPostgresClient(config PostgresConfig) (PostgresClient, error) {
	if config.ConnectionString == "" {
		return nil, errors.New("connection string cannot be empty")
	}

	conn, err := pgx.Connect(context.Background(), config.ConnectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	return &postgresClient{
		conn: conn,
	}, nil
}

// Exec executes a SQL statement that doesn't return rows
func (c *postgresClient) Exec(ctx context.Context, sql string, arguments ...interface{}) error {
	_, err := c.conn.Exec(ctx, sql, arguments...)
	if err != nil {
		return fmt.Errorf("postgres exec error: %w", err)
	}
	return nil
}

// Query executes a SQL query that returns rows
func (c *postgresClient) Query(ctx context.Context, sql string, arguments ...interface{}) (pgx.Rows, error) {
	rows, err := c.conn.Query(ctx, sql, arguments...)
	if err != nil {
		return nil, fmt.Errorf("postgres query error: %w", err)
	}
	return rows, nil
}

// QueryRow executes a SQL query that returns a single row
func (c *postgresClient) QueryRow(ctx context.Context, sql string, arguments ...interface{}) pgx.Row {
	return c.conn.QueryRow(ctx, sql, arguments...)
}

// Begin starts a transaction
func (c *postgresClient) Begin(ctx context.Context) (pgx.Tx, error) {
	tx, err := c.conn.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("postgres begin transaction error: %w", err)
	}
	return tx, nil
}

// Close closes the connection
func (c *postgresClient) Close(ctx context.Context) error {
	return c.conn.Close(ctx)
}
