//go:build integration
// +build integration

package postgres

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// TestPostgresClient_Operations tests real PostgreSQL operations using testcontainers.
// Run with: go test -tags=integration ./lib/workflows/postgres/v1
func TestPostgresClient_Operations(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx := context.Background()

	// Start PostgreSQL container
	req := testcontainers.ContainerRequest{
		Image:        "postgres:latest",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "testuser",
			"POSTGRES_PASSWORD": "testpass",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections").
			WithOccurrence(2). // PostgreSQL logs this twice during startup
			WithStartupTimeout(60 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
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
	port, err := container.MappedPort(ctx, "5432")
	if err != nil {
		t.Fatalf("failed to get container port: %v", err)
	}

	config := PostgresConfig{
		Host:     host,
		Port:     port.Int(),
		User:     "testuser",
		Password: "testpass",
		Database: "testdb",
	}

	t.Run("success - creates client with valid config", func(t *testing.T) {
		client, err := NewPostgresClient(config)
		if err != nil {
			t.Fatalf("failed to create client: %v", err)
		}
		defer client.Close(ctx)
	})

	t.Run("success - exec creates table", func(t *testing.T) {
		client, err := NewPostgresClient(config)
		if err != nil {
			t.Fatalf("failed to create client: %v", err)
		}
		defer client.Close(ctx)

		err = client.Exec(ctx, `CREATE TABLE IF NOT EXISTS test_exec (id SERIAL PRIMARY KEY, name TEXT)`)
		if err != nil {
			t.Fatalf("failed to create table: %v", err)
		}
	})

	t.Run("success - exec inserts data", func(t *testing.T) {
		client, err := NewPostgresClient(config)
		if err != nil {
			t.Fatalf("failed to create client: %v", err)
		}
		defer client.Close(ctx)

		// Ensure table exists
		err = client.Exec(ctx, `CREATE TABLE IF NOT EXISTS test_insert (id SERIAL PRIMARY KEY, value TEXT)`)
		if err != nil {
			t.Fatalf("failed to create table: %v", err)
		}

		err = client.Exec(ctx, `INSERT INTO test_insert (value) VALUES ($1)`, "test-value")
		if err != nil {
			t.Fatalf("failed to insert data: %v", err)
		}
	})

	t.Run("success - query returns rows", func(t *testing.T) {
		client, err := NewPostgresClient(config)
		if err != nil {
			t.Fatalf("failed to create client: %v", err)
		}
		defer client.Close(ctx)

		// Setup test data
		err = client.Exec(ctx, `CREATE TABLE IF NOT EXISTS test_query (id SERIAL PRIMARY KEY, name TEXT)`)
		if err != nil {
			t.Fatalf("failed to create table: %v", err)
		}
		err = client.Exec(ctx, `INSERT INTO test_query (name) VALUES ('row1'), ('row2')`)
		if err != nil {
			t.Fatalf("failed to insert test data: %v", err)
		}

		rows, err := client.Query(ctx, `SELECT name FROM test_query ORDER BY name`)
		if err != nil {
			t.Fatalf("failed to query: %v", err)
		}
		defer rows.Close()

		var names []string
		for rows.Next() {
			var name string
			if err := rows.Scan(&name); err != nil {
				t.Fatalf("failed to scan row: %v", err)
			}
			names = append(names, name)
		}

		if len(names) < 2 {
			t.Errorf("expected at least 2 rows, got %d", len(names))
		}
	})

	t.Run("success - query row returns single row", func(t *testing.T) {
		client, err := NewPostgresClient(config)
		if err != nil {
			t.Fatalf("failed to create client: %v", err)
		}
		defer client.Close(ctx)

		// Setup test data
		err = client.Exec(ctx, `CREATE TABLE IF NOT EXISTS test_queryrow (id SERIAL PRIMARY KEY, value INT)`)
		if err != nil {
			t.Fatalf("failed to create table: %v", err)
		}
		err = client.Exec(ctx, `INSERT INTO test_queryrow (value) VALUES (42) ON CONFLICT DO NOTHING`)
		if err != nil {
			t.Fatalf("failed to insert test data: %v", err)
		}

		row := client.QueryRow(ctx, `SELECT value FROM test_queryrow LIMIT 1`)
		var value int
		if err := row.Scan(&value); err != nil {
			t.Fatalf("failed to scan row: %v", err)
		}

		if value != 42 {
			t.Errorf("expected value 42, got %d", value)
		}
	})

	t.Run("success - begin starts transaction", func(t *testing.T) {
		client, err := NewPostgresClient(config)
		if err != nil {
			t.Fatalf("failed to create client: %v", err)
		}
		defer client.Close(ctx)

		tx, err := client.Begin(ctx)
		if err != nil {
			t.Fatalf("failed to begin transaction: %v", err)
		}

		// Rollback to clean up
		if err := tx.Rollback(ctx); err != nil {
			t.Fatalf("failed to rollback: %v", err)
		}
	})

	t.Run("success - transaction commit persists changes", func(t *testing.T) {
		client, err := NewPostgresClient(config)
		if err != nil {
			t.Fatalf("failed to create client: %v", err)
		}
		defer client.Close(ctx)

		// Setup table
		err = client.Exec(ctx, `CREATE TABLE IF NOT EXISTS test_tx_commit (id SERIAL PRIMARY KEY, data TEXT)`)
		if err != nil {
			t.Fatalf("failed to create table: %v", err)
		}

		// Start transaction and insert
		tx, err := client.Begin(ctx)
		if err != nil {
			t.Fatalf("failed to begin transaction: %v", err)
		}

		_, err = tx.Exec(ctx, `INSERT INTO test_tx_commit (data) VALUES ('committed')`)
		if err != nil {
			t.Fatalf("failed to insert in transaction: %v", err)
		}

		if err := tx.Commit(ctx); err != nil {
			t.Fatalf("failed to commit: %v", err)
		}

		// Verify data persisted
		row := client.QueryRow(ctx, `SELECT data FROM test_tx_commit WHERE data = 'committed'`)
		var data string
		if err := row.Scan(&data); err != nil {
			t.Fatalf("committed data not found: %v", err)
		}
	})

	t.Run("success - transaction rollback discards changes", func(t *testing.T) {
		client, err := NewPostgresClient(config)
		if err != nil {
			t.Fatalf("failed to create client: %v", err)
		}
		defer client.Close(ctx)

		// Setup table with unique constraint for testing
		tableName := fmt.Sprintf("test_tx_rollback_%d", time.Now().UnixNano())
		err = client.Exec(ctx, fmt.Sprintf(`CREATE TABLE %s (id SERIAL PRIMARY KEY, data TEXT UNIQUE)`, tableName))
		if err != nil {
			t.Fatalf("failed to create table: %v", err)
		}

		// Start transaction and insert
		tx, err := client.Begin(ctx)
		if err != nil {
			t.Fatalf("failed to begin transaction: %v", err)
		}

		_, err = tx.Exec(ctx, fmt.Sprintf(`INSERT INTO %s (data) VALUES ('to-be-rolled-back')`, tableName))
		if err != nil {
			t.Fatalf("failed to insert in transaction: %v", err)
		}

		if err := tx.Rollback(ctx); err != nil {
			t.Fatalf("failed to rollback: %v", err)
		}

		// Verify data was not persisted
		rows, err := client.Query(ctx, fmt.Sprintf(`SELECT data FROM %s WHERE data = 'to-be-rolled-back'`, tableName))
		if err != nil {
			t.Fatalf("failed to query: %v", err)
		}
		defer rows.Close()

		if rows.Next() {
			t.Error("rolled back data should not exist")
		}
	})

	t.Run("error - exec with invalid SQL", func(t *testing.T) {
		client, err := NewPostgresClient(config)
		if err != nil {
			t.Fatalf("failed to create client: %v", err)
		}
		defer client.Close(ctx)

		err = client.Exec(ctx, `INVALID SQL STATEMENT`)
		if err == nil {
			t.Fatal("expected error for invalid SQL, got nil")
		}
	})

	t.Run("error - query nonexistent table", func(t *testing.T) {
		client, err := NewPostgresClient(config)
		if err != nil {
			t.Fatalf("failed to create client: %v", err)
		}
		defer client.Close(ctx)

		_, err = client.Query(ctx, `SELECT * FROM nonexistent_table_xyz`)
		if err == nil {
			t.Fatal("expected error for nonexistent table, got nil")
		}
	})
}

// TestPostgresClient_ConnectionError tests connection failure scenarios
func TestPostgresClient_ConnectionError(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	t.Run("error - connection refused", func(t *testing.T) {
		config := PostgresConfig{
			Host:     "localhost",
			Port:     59999, // Port unlikely to have PostgreSQL
			User:     "testuser",
			Password: "testpass",
			Database: "testdb",
		}

		_, err := NewPostgresClient(config)
		if err == nil {
			t.Fatal("expected connection error, got nil")
		}
	})
}
