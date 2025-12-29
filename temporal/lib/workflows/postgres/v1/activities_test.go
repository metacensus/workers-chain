package postgres_test

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	pb "github.com/metacensus/infra/temporal/gen/workflows/postgres/v1"
	postgresv1 "github.com/metacensus/infra/temporal/lib/workflows/postgres/v1"
	postgresmocks "github.com/metacensus/infra/temporal/mocks/workflows/postgres/v1"
	"github.com/stretchr/testify/mock"
	"go.temporal.io/sdk/testsuite"
)

var (
	// Test error sentinels for client errors
	errClientConnectionRefused = errors.New("connection refused")
	errClientQueryFailed       = errors.New("query failed")
)

// mockRows is a simple mock implementation of pgx.Rows for testing
type mockRows struct {
	hasNext bool
}

func (m *mockRows) Next() bool {
	if m.hasNext {
		m.hasNext = false
		return true
	}
	return false
}

func (m *mockRows) Close() {
	// no-op
}

func (m *mockRows) CommandTag() pgconn.CommandTag {
	return pgconn.CommandTag{}
}

func (m *mockRows) Err() error {
	return nil
}

func (m *mockRows) FieldDescriptions() []pgconn.FieldDescription {
	return nil
}

func (m *mockRows) Scan(dest ...interface{}) error {
	return nil
}

func (m *mockRows) Values() ([]interface{}, error) {
	return nil, nil
}

func (m *mockRows) RawValues() [][]byte {
	return nil
}

func (m *mockRows) Conn() *pgx.Conn {
	return nil
}

func TestCreateDatabaseActivity(t *testing.T) {
	tests := []struct {
		name      string
		input     *pb.CreateDatabaseInput
		setupMock func(*postgresmocks.MockPostgresClient)
		expectErr bool
	}{
		{
			name: "success - creates new database",
			input: &pb.CreateDatabaseInput{
				DatabaseName: "testdb",
			},
			setupMock: func(m *postgresmocks.MockPostgresClient) {
				// Mock database existence check (returns false - doesn't exist)
				rows := &mockRows{hasNext: false}
				m.EXPECT().
					Query(mock.Anything, "SELECT 1 FROM pg_database WHERE datname = $1", "testdb").
					Return(rows, nil)

				// Mock database creation
				m.EXPECT().
					Exec(mock.Anything, `CREATE DATABASE "testdb"`).
					Return(nil)
			},
		},
		{
			name: "success - database already exists",
			input: &pb.CreateDatabaseInput{
				DatabaseName: "existingdb",
			},
			setupMock: func(m *postgresmocks.MockPostgresClient) {
				// Mock database existence check (returns true - exists)
				rows := &mockRows{hasNext: true}
				m.EXPECT().
					Query(mock.Anything, "SELECT 1 FROM pg_database WHERE datname = $1", "existingdb").
					Return(rows, nil)
			},
		},
		{
			name: "error - missing database name",
			input: &pb.CreateDatabaseInput{
				DatabaseName: "",
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestActivityEnvironment()

			mockClient := postgresmocks.NewMockPostgresClient(t)
			if tt.setupMock != nil {
				tt.setupMock(mockClient)
			}

			activities := postgresv1.NewActivities(mockClient)
			env.RegisterActivity(activities.CreateDatabase)

			// Execute activity
			val, err := env.ExecuteActivity(activities.CreateDatabase, tt.input)

			if tt.expectErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error from ExecuteActivity: %v", err)
				}
				var result *pb.CreateDatabaseOutput
				if err := val.Get(&result); err != nil {
					t.Fatalf("unexpected error from Get: %v", err)
				}
				if result == nil {
					t.Fatal("result is nil")
				}
			}
		})
	}
}

func TestCreateUserActivity(t *testing.T) {
	tests := []struct {
		name      string
		input     *pb.CreateUserInput
		setupMock func(*postgresmocks.MockPostgresClient)
		expectErr bool
	}{
		{
			name: "success - creates new user",
			input: &pb.CreateUserInput{
				Username: "testuser",
				Password: "testpass",
			},
			setupMock: func(m *postgresmocks.MockPostgresClient) {
				// Mock user existence check (returns false - doesn't exist)
				rows := &mockRows{hasNext: false}
				m.EXPECT().
					Query(mock.Anything, "SELECT 1 FROM pg_roles WHERE rolname = $1", "testuser").
					Return(rows, nil)

				// Mock user creation
				m.EXPECT().
					Exec(mock.Anything, mock.MatchedBy(func(sql string) bool {
						return sql == `CREATE ROLE "testuser" WITH LOGIN PASSWORD 'testpass'`
					})).
					Return(nil)
			},
		},
		{
			name: "success - user already exists",
			input: &pb.CreateUserInput{
				Username: "existinguser",
				Password: "testpass",
			},
			setupMock: func(m *postgresmocks.MockPostgresClient) {
				// Mock user existence check (returns true - exists)
				rows := &mockRows{hasNext: true}
				m.EXPECT().
					Query(mock.Anything, "SELECT 1 FROM pg_roles WHERE rolname = $1", "existinguser").
					Return(rows, nil)
			},
		},
		{
			name: "error - missing username",
			input: &pb.CreateUserInput{
				Username: "",
				Password: "testpass",
			},
			expectErr: true,
		},
		{
			name: "error - missing password",
			input: &pb.CreateUserInput{
				Username: "testuser",
				Password: "",
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestActivityEnvironment()

			mockClient := postgresmocks.NewMockPostgresClient(t)
			if tt.setupMock != nil {
				tt.setupMock(mockClient)
			}

			activities := postgresv1.NewActivities(mockClient)
			env.RegisterActivity(activities.CreateUser)

			// Execute activity
			val, err := env.ExecuteActivity(activities.CreateUser, tt.input)

			if tt.expectErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error from ExecuteActivity: %v", err)
				}
				var result *pb.CreateUserOutput
				if err := val.Get(&result); err != nil {
					t.Fatalf("unexpected error from Get: %v", err)
				}
				if result == nil {
					t.Fatal("result is nil")
				}
			}
		})
	}
}

func TestGrantPrivilegesActivity(t *testing.T) {
	tests := []struct {
		name      string
		input     *pb.GrantPrivilegesInput
		setupMock func(*postgresmocks.MockPostgresClient)
		expectErr bool
	}{
		{
			name: "success - grants ALL privileges",
			input: &pb.GrantPrivilegesInput{
				Username:     "testuser",
				DatabaseName: "testdb",
				Privileges:   "ALL",
			},
			setupMock: func(m *postgresmocks.MockPostgresClient) {
				m.EXPECT().
					Exec(mock.Anything, `GRANT ALL ON DATABASE "testdb" TO "testuser"`).
					Return(nil)
			},
		},
		{
			name: "success - grants specific privileges",
			input: &pb.GrantPrivilegesInput{
				Username:     "testuser",
				DatabaseName: "testdb",
				Privileges:   "SELECT,INSERT,UPDATE",
			},
			setupMock: func(m *postgresmocks.MockPostgresClient) {
				m.EXPECT().
					Exec(mock.Anything, `GRANT SELECT,INSERT,UPDATE ON DATABASE "testdb" TO "testuser"`).
					Return(nil)
			},
		},
		{
			name: "success - defaults to ALL when privileges empty",
			input: &pb.GrantPrivilegesInput{
				Username:     "testuser",
				DatabaseName: "testdb",
				Privileges:   "",
			},
			setupMock: func(m *postgresmocks.MockPostgresClient) {
				m.EXPECT().
					Exec(mock.Anything, `GRANT ALL ON DATABASE "testdb" TO "testuser"`).
					Return(nil)
			},
		},
		{
			name: "error - missing username",
			input: &pb.GrantPrivilegesInput{
				Username:     "",
				DatabaseName: "testdb",
				Privileges:   "ALL",
			},
			expectErr: true,
		},
		{
			name: "error - missing database name",
			input: &pb.GrantPrivilegesInput{
				Username:     "testuser",
				DatabaseName: "",
				Privileges:   "ALL",
			},
			expectErr: true,
		},
		{
			name: "error - client exec failed",
			input: &pb.GrantPrivilegesInput{
				Username:     "testuser",
				DatabaseName: "testdb",
				Privileges:   "ALL",
			},
			setupMock: func(m *postgresmocks.MockPostgresClient) {
				m.EXPECT().
					Exec(mock.Anything, mock.Anything).
					Return(errClientQueryFailed)
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestActivityEnvironment()

			mockClient := postgresmocks.NewMockPostgresClient(t)
			if tt.setupMock != nil {
				tt.setupMock(mockClient)
			}

			activities := postgresv1.NewActivities(mockClient)
			env.RegisterActivity(activities.GrantPrivileges)

			// Execute activity
			val, err := env.ExecuteActivity(activities.GrantPrivileges, tt.input)

			if tt.expectErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error from ExecuteActivity: %v", err)
				}
				var result *pb.GrantPrivilegesOutput
				if err := val.Get(&result); err != nil {
					t.Fatalf("unexpected error from Get: %v", err)
				}
				if result == nil {
					t.Fatal("result is nil")
				}
			}
		})
	}
}
