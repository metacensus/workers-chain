package postgres

import (
	"strings"
	"testing"
)

func TestPostgresConfig_ConnectionString(t *testing.T) {
	tests := []struct {
		name     string
		config   PostgresConfig
		expected string
	}{
		{
			name: "builds connection string with all fields",
			config: PostgresConfig{
				Host:     "localhost",
				Port:     5432,
				User:     "admin",
				Password: "secret",
				Database: "testdb",
			},
			expected: "postgres://admin:secret@localhost:5432/testdb?sslmode=disable",
		},
		{
			name: "handles different port",
			config: PostgresConfig{
				Host:     "db.example.com",
				Port:     5433,
				User:     "myuser",
				Password: "mypass",
				Database: "mydb",
			},
			expected: "postgres://myuser:mypass@db.example.com:5433/mydb?sslmode=disable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.ConnectionString()
			if result != tt.expected {
				t.Errorf("ConnectionString() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestNewPostgresClient_Validation(t *testing.T) {
	tests := []struct {
		name        string
		config      PostgresConfig
		expectErr   bool
		errContains string
	}{
		{
			name: "error - empty host",
			config: PostgresConfig{
				Host:     "",
				Port:     5432,
				User:     "admin",
				Password: "secret",
				Database: "testdb",
			},
			expectErr:   true,
			errContains: "host cannot be empty",
		},
		{
			name: "error - zero port",
			config: PostgresConfig{
				Host:     "localhost",
				Port:     0,
				User:     "admin",
				Password: "secret",
				Database: "testdb",
			},
			expectErr:   true,
			errContains: "port cannot be zero",
		},
		{
			name: "error - empty user",
			config: PostgresConfig{
				Host:     "localhost",
				Port:     5432,
				User:     "",
				Password: "secret",
				Database: "testdb",
			},
			expectErr:   true,
			errContains: "user cannot be empty",
		},
		{
			name: "error - empty database",
			config: PostgresConfig{
				Host:     "localhost",
				Port:     5432,
				User:     "admin",
				Password: "secret",
				Database: "",
			},
			expectErr:   true,
			errContains: "database cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewPostgresClient(tt.config)

			if tt.expectErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("error %q does not contain %q", err.Error(), tt.errContains)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}
		})
	}
}
