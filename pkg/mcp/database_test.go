package mcp

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetDatabaseType(t *testing.T) {
	tests := []struct {
		name     string
		desc     map[string]any
		expected DatabaseType
	}{
		{
			name: "PostgreSQL database_type",
			desc: map[string]any{
				"database_type": "postgres",
			},
			expected: DatabaseTypePostgres,
		},
		{
			name: "MySQL database_type",
			desc: map[string]any{
				"database_type": "mysql",
			},
			expected: DatabaseTypeMySQL,
		},
		{
			name: "MariaDB database_type",
			desc: map[string]any{
				"database_type": "mariadb",
			},
			expected: DatabaseTypeMariaDB,
		},
		{
			name: "PostgreSQL in dbinfo",
			desc: map[string]any{
				"dbinfo": map[string]any{
					"database_type": "postgres",
				},
			},
			expected: DatabaseTypePostgres,
		},
		{
			name:     "No database info defaults to MySQL",
			desc:     map[string]any{},
			expected: DatabaseTypeMySQL,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetDatabaseType(tt.desc)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestGetDatabaseConfig(t *testing.T) {
	tests := []struct {
		name     string
		dbType   DatabaseType
		expected DatabaseConfig
	}{
		{
			name:   "PostgreSQL config",
			dbType: DatabaseTypePostgres,
			expected: DatabaseConfig{
				Client:        "psql",
				CommandFlag:   "-c",
				ListTables:    "\\dt",
				ListDatabases: "\\l",
			},
		},
		{
			name:   "MySQL config",
			dbType: DatabaseTypeMySQL,
			expected: DatabaseConfig{
				Client:        "mysql",
				CommandFlag:   "-e",
				ListTables:    "SHOW TABLES;",
				ListDatabases: "SHOW DATABASES;",
			},
		},
		{
			name:   "MariaDB config",
			dbType: DatabaseTypeMariaDB,
			expected: DatabaseConfig{
				Client:        "mysql",
				CommandFlag:   "-e",
				ListTables:    "SHOW TABLES;",
				ListDatabases: "SHOW DATABASES;",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetDatabaseConfig(tt.dbType)
			require.Equal(t, tt.expected.Client, result.Client)
			require.Equal(t, tt.expected.CommandFlag, result.CommandFlag)
			require.Equal(t, tt.expected.ListTables, result.ListTables)
			require.Equal(t, tt.expected.ListDatabases, result.ListDatabases)
		})
	}
}

func TestIsReadOnlyQuery(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected bool
	}{
		{
			name:     "Simple SELECT",
			query:    "SELECT * FROM users",
			expected: true,
		},
		{
			name:     "SHOW TABLES",
			query:    "SHOW TABLES",
			expected: true,
		},
		{
			name:     "DESCRIBE table",
			query:    "DESCRIBE users",
			expected: true,
		},
		{
			name:     "EXPLAIN query",
			query:    "EXPLAIN SELECT * FROM users",
			expected: true,
		},
		{
			name:     "PostgreSQL \\dt",
			query:    "\\dt",
			expected: true,
		},
		{
			name:     "INSERT statement",
			query:    "INSERT INTO users (name) VALUES ('test')",
			expected: false,
		},
		{
			name:     "UPDATE statement",
			query:    "UPDATE users SET name = 'test'",
			expected: false,
		},
		{
			name:     "DELETE statement",
			query:    "DELETE FROM users WHERE id = 1",
			expected: false,
		},
		{
			name:     "Multiple statements (SQL injection)",
			query:    "SELECT * FROM users; DROP TABLE users;",
			expected: false,
		},
		{
			name:     "Empty query",
			query:    "",
			expected: true,
		},
		{
			name:     "Comments only",
			query:    "/* This is a comment */ -- Another comment",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsReadOnlyQuery(tt.query)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateQuerySecurity(t *testing.T) {
	tests := []struct {
		name            string
		query           string
		allowWriteOps   bool
		expectedAllowed bool
		expectedReason  string
	}{
		{
			name:            "Safe SELECT query",
			query:           "SELECT * FROM users",
			allowWriteOps:   false,
			expectedAllowed: true,
		},
		{
			name:            "INSERT with write allowed",
			query:           "INSERT INTO users (name) VALUES ('test')",
			allowWriteOps:   true,
			expectedAllowed: true,
		},
		{
			name:            "INSERT without write allowed",
			query:           "INSERT INTO users (name) VALUES ('test')",
			allowWriteOps:   false,
			expectedAllowed: false,
			expectedReason:  "Query not in whitelist of safe read-only operations",
		},
		{
			name:            "DROP DATABASE always blocked",
			query:           "DROP DATABASE test",
			allowWriteOps:   true,
			expectedAllowed: false,
			expectedReason:  "This operation is permanently blocked",
		},
		{
			name:            "SHUTDOWN always blocked",
			query:           "SHUTDOWN",
			allowWriteOps:   true,
			expectedAllowed: false,
			expectedReason:  "This operation is permanently blocked",
		},
		{
			name:            "LOAD_FILE blocked",
			query:           "SELECT LOAD_FILE('/etc/passwd')",
			allowWriteOps:   true,
			expectedAllowed: false,
			expectedReason:  "This operation is permanently blocked",
		},
		{
			name:            "CREATE USER blocked",
			query:           "CREATE USER 'test'@'localhost'",
			allowWriteOps:   true,
			expectedAllowed: false,
			expectedReason:  "This operation is permanently blocked",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allowed, reason := ValidateQuerySecurity(tt.query, tt.allowWriteOps)
			require.Equal(t, tt.expectedAllowed, allowed)
			if tt.expectedReason != "" {
				require.Contains(t, reason, tt.expectedReason)
			}
		})
	}
}

func TestBuildDatabaseCommand(t *testing.T) {
	tests := []struct {
		name     string
		dbType   DatabaseType
		database string
		query    string
		expected string
	}{
		{
			name:     "MySQL without database",
			dbType:   DatabaseTypeMySQL,
			database: "",
			query:    "SELECT * FROM users",
			expected: "exec mysql -e \"SELECT * FROM users\"",
		},
		{
			name:     "MySQL with database",
			dbType:   DatabaseTypeMySQL,
			database: "testdb",
			query:    "SELECT * FROM users",
			expected: "exec mysql -D testdb -e \"SELECT * FROM users\"",
		},
		{
			name:     "PostgreSQL without database",
			dbType:   DatabaseTypePostgres,
			database: "",
			query:    "SELECT * FROM users",
			expected: "exec psql -c \"SELECT * FROM users\"",
		},
		{
			name:     "PostgreSQL with database",
			dbType:   DatabaseTypePostgres,
			database: "testdb",
			query:    "SELECT * FROM users",
			expected: "exec psql -d testdb -c \"SELECT * FROM users\"",
		},
		{
			name:     "Query with quotes",
			dbType:   DatabaseTypeMySQL,
			database: "testdb",
			query:    `SELECT * FROM users WHERE name = "test"`,
			expected: `exec mysql -D testdb -e "SELECT * FROM users WHERE name = \"test\""`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildDatabaseCommand(tt.dbType, tt.database, tt.query)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestHandleDatabaseQuery_SecurityValidation(t *testing.T) {
	// Test security validation directly without project dependencies
	t.Run("Security validation only", func(t *testing.T) {
		allowed, reason := ValidateQuerySecurity("SELECT * FROM users", false)
		require.True(t, allowed)
		require.Empty(t, reason)

		allowed, reason = ValidateQuerySecurity("DROP DATABASE test", true)
		require.False(t, allowed)
		require.Contains(t, reason, "permanently blocked")

		allowed, reason = ValidateQuerySecurity("INSERT INTO users VALUES ('test')", false)
		require.False(t, allowed)
		require.Contains(t, reason, "read-only operations")

		allowed, reason = ValidateQuerySecurity("INSERT INTO users VALUES ('test')", true)
		require.True(t, allowed)
		require.Empty(t, reason)
	})
}

// MockSecurityManager for testing
type MockSecurityManager struct {
	permissions map[string]bool
}

func (m *MockSecurityManager) CheckPermission(toolName string, args map[string]any) error {
	if !m.permissions[toolName] {
		return fmt.Errorf("permission denied for tool %s", toolName)
	}
	return nil
}

func (m *MockSecurityManager) RequiresApproval(toolName string, args map[string]any) bool {
	return false
}

func (m *MockSecurityManager) RequestApproval(toolName string, args map[string]any, description string) error {
	return nil
}

func (m *MockSecurityManager) LogOperation(toolName string, args map[string]any, result any, err error) {
	// Mock implementation - do nothing
}
