package mcp

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

// DatabaseType represents the supported database types
type DatabaseType string

const (
	DatabaseTypeMySQL    DatabaseType = "mysql"
	DatabaseTypeMariaDB  DatabaseType = "mariadb"
	DatabaseTypePostgres DatabaseType = "postgres"
)

// DatabaseConfig contains database-specific command configurations
type DatabaseConfig struct {
	Client        string
	CommandFlag   string
	ListTables    string
	DescribeTable func(table string) string
	ListDatabases string
}

// GetDatabaseType determines the database type from project description
func GetDatabaseType(desc map[string]any) DatabaseType {
	if dbType, ok := desc["database_type"].(string); ok {
		switch strings.ToLower(dbType) {
		case "postgres", "postgresql":
			return DatabaseTypePostgres
		case "mariadb":
			return DatabaseTypeMariaDB
		case "mysql":
			return DatabaseTypeMySQL
		}
	}

	// Check dbinfo as fallback
	if dbInfo, ok := desc["dbinfo"].(map[string]any); ok {
		if dbType, ok := dbInfo["database_type"].(string); ok {
			switch strings.ToLower(dbType) {
			case "postgres", "postgresql":
				return DatabaseTypePostgres
			case "mariadb":
				return DatabaseTypeMariaDB
			case "mysql":
				return DatabaseTypeMySQL
			}
		}
	}

	return DatabaseTypeMySQL // Default fallback
}

// GetDatabaseConfig returns database-specific configuration
func GetDatabaseConfig(dbType DatabaseType) DatabaseConfig {
	switch dbType {
	case DatabaseTypePostgres:
		return DatabaseConfig{
			Client:      "psql",
			CommandFlag: "-c",
			ListTables:  "\\dt",
			DescribeTable: func(table string) string {
				return fmt.Sprintf("\\d %s", table)
			},
			ListDatabases: "\\l",
		}
	case DatabaseTypeMySQL, DatabaseTypeMariaDB:
		return DatabaseConfig{
			Client:      "mysql",
			CommandFlag: "-e",
			ListTables:  "SHOW TABLES;",
			DescribeTable: func(table string) string {
				return fmt.Sprintf("DESCRIBE %s;", table)
			},
			ListDatabases: "SHOW DATABASES;",
		}
	default:
		// Default to MySQL configuration
		return DatabaseConfig{
			Client:      "mysql",
			CommandFlag: "-e",
			ListTables:  "SHOW TABLES;",
			DescribeTable: func(table string) string {
				return fmt.Sprintf("DESCRIBE %s;", table)
			},
			ListDatabases: "SHOW DATABASES;",
		}
	}
}

// IsReadOnlyQuery checks if a SQL query is read-only
func IsReadOnlyQuery(query string) bool {
	normalizedQuery := normalizeQuery(query)

	if normalizedQuery == "" {
		return true
	}

	// Check for multiple statements (SQL injection attempts)
	statements := strings.Split(normalizedQuery, ";")
	var meaningfulStatements []string
	for _, stmt := range statements {
		trimmed := strings.TrimSpace(stmt)
		if trimmed != "" {
			meaningfulStatements = append(meaningfulStatements, trimmed)
		}
	}

	if len(meaningfulStatements) > 1 {
		return false // Multiple statements detected - likely SQL injection
	}

	// First check for dangerous keywords that should make it non-read-only
	dangerousPatterns := []*regexp.Regexp{
		regexp.MustCompile(`\bINSERT\b`),
		regexp.MustCompile(`\bUPDATE\b`),
		regexp.MustCompile(`\bDELETE\b`),
		regexp.MustCompile(`\bCREATE\b`),
		regexp.MustCompile(`\bDROP\b`),
		regexp.MustCompile(`\bALTER\b`),
		regexp.MustCompile(`\bTRUNCATE\b`),
		regexp.MustCompile(`\bREPLACE\b`),
		regexp.MustCompile(`\bMERGE\b`),
		regexp.MustCompile(`\bGRANT\b`),
		regexp.MustCompile(`\bREVOKE\b`),
		regexp.MustCompile(`\bSET\b`),
		regexp.MustCompile(`\bRESET\b`),
		regexp.MustCompile(`\bCALL\b`),
		regexp.MustCompile(`\bEXECUTE\b`),
		regexp.MustCompile(`\bEXEC\b`),
		regexp.MustCompile(`\bCOPY\b`),
		regexp.MustCompile(`\bVACUUM\b`),
		regexp.MustCompile(`\bANALYZE\b`),
		regexp.MustCompile(`\bCLUSTER\b`),
		regexp.MustCompile(`\bREINDEX\b`),
		regexp.MustCompile(`\bLOAD\b`),
		regexp.MustCompile(`\bIMPORT\b`),
		regexp.MustCompile(`\bFLUSH\b`),
		regexp.MustCompile(`\bOPTIMIZE\b`),
		regexp.MustCompile(`\bREPAIR\b`),
		regexp.MustCompile(`\bCHECKSUM\b`),
		regexp.MustCompile(`\bBEGIN\b`),
		regexp.MustCompile(`\bSTART\b`),
		regexp.MustCompile(`\bCOMMIT\b`),
		regexp.MustCompile(`\bROLLBACK\b`),
		regexp.MustCompile(`\bSAVEPOINT\b`),
		regexp.MustCompile(`\bRENAME\b`),
		regexp.MustCompile(`\bCOMMENT\b`),
		regexp.MustCompile(`\bHANDLER\b`),
		regexp.MustCompile(`\bLOCK\b`),
		regexp.MustCompile(`\bUNLOCK\b`),
	}

	for _, pattern := range dangerousPatterns {
		if pattern.MatchString(normalizedQuery) {
			return false
		}
	}

	// Safe read-only patterns
	safeReadOnlyPatterns := []*regexp.Regexp{
		regexp.MustCompile(`^SELECT\b`),
		regexp.MustCompile(`^SHOW\s+(TABLES|DATABASES|SCHEMAS|COLUMNS|INDEX|INDEXES|INDICES|STATUS|VARIABLES|PROCESSLIST|ENGINES|CHARSET|COLLATION|CREATE\s+TABLE|CREATE\s+DATABASE|CREATE\s+VIEW|TABLE\s+STATUS|FULL\s+TABLES|GRANTS|PRIVILEGES)\b`),
		regexp.MustCompile(`^DESCRIBE\b`),
		regexp.MustCompile(`^DESC\b`),
		regexp.MustCompile(`^EXPLAIN\b`),
		regexp.MustCompile(`^EXPLAIN\s+(ANALYZE|VERBOSE|FORMAT)\b`),
		regexp.MustCompile(`^\\DT$`),
		regexp.MustCompile(`^\\D\s+\w+$`),
		regexp.MustCompile(`^\\L$`),
		regexp.MustCompile(`^\\DN$`),
		regexp.MustCompile(`^\\DF$`),
		regexp.MustCompile(`^\\DV$`),
		regexp.MustCompile(`^\\DI$`),
		regexp.MustCompile(`^\\DU$`),
		regexp.MustCompile(`^\\DP$`),
		regexp.MustCompile(`^\\Z$`),
		regexp.MustCompile(`^WITH\b.*\bSELECT\b`),
		regexp.MustCompile(`^SELECT\s+.*\s+FROM\s+(INFORMATION_SCHEMA|PERFORMANCE_SCHEMA|mysql\.)\w+`),
		regexp.MustCompile(`^SELECT\s+.*\s+FROM\s+(pg_catalog|information_schema)\.\w+`),
		regexp.MustCompile(`^SELECT\s+.*\s+FROM\s+pg_\w+`),
	}

	for _, pattern := range safeReadOnlyPatterns {
		if pattern.MatchString(normalizedQuery) {
			return true
		}
	}

	return false
}

// ValidateQuerySecurity validates a SQL query against security rules
func ValidateQuerySecurity(query string, allowWriteOperations bool) (allowed bool, reason string) {
	normalizedQuery := strings.TrimSpace(strings.ToUpper(query))

	// Catastrophic operation patterns that should always be blocked
	catastrophicPatterns := []*regexp.Regexp{
		regexp.MustCompile(`DROP\s+(DATABASE|SCHEMA|TABLESPACE)\b`),
		regexp.MustCompile(`SHUTDOWN\b`),
		regexp.MustCompile(`KILL\b`),
		regexp.MustCompile(`LOAD_FILE\b`),
		regexp.MustCompile(`INTO\s+(OUTFILE|DUMPFILE)\b`),
		regexp.MustCompile(`\/\*!.*?\*\/`),
		regexp.MustCompile(`\\!\s*`),
		regexp.MustCompile(`COPY\s+.*FROM\s+PROGRAM\b`),
		regexp.MustCompile(`SELECT\s+.*INTO\s+(OUTFILE|DUMPFILE)\b`),
		regexp.MustCompile(`LOAD\s+DATA\s+LOCAL\s+INFILE\b`),
		regexp.MustCompile(`GRANT\s+(ALL\s+PRIVILEGES|CREATE|DROP|ALTER|DELETE|INSERT|UPDATE|SELECT|SUPER|RELOAD|LOCK\s+TABLES|REPLICATION|BINLOG|PROCESS|FILE|REFERENCES|INDEX|CREATE\s+USER|SHUTDOWN|CREATE\s+TEMPORARY\s+TABLES|EXECUTE|REPLICATION\s+SLAVE|REPLICATION\s+CLIENT|CREATE\s+VIEW|SHOW\s+VIEW|CREATE\s+ROUTINE|ALTER\s+ROUTINE|EVENT|TRIGGER)`),
		regexp.MustCompile(`CREATE\s+USER\b`),
		regexp.MustCompile(`SET\s+(GLOBAL|SESSION|@@)`),
		regexp.MustCompile(`^\s*UNION\s+SELECT\b`),
		regexp.MustCompile(`SELECT\s+@@DATADIR`),
		regexp.MustCompile(`SELECT\s+@@BASEDIR`),
		regexp.MustCompile(`SELECT\s+@@TMPDIR`),
		regexp.MustCompile(`SELECT\s+@@SECURE_FILE_PRIV`),
		regexp.MustCompile(`SELECT\s+@@PLUGIN_DIR`),
		regexp.MustCompile(`SHOW\s+GRANTS\b`),
	}

	for _, pattern := range catastrophicPatterns {
		if pattern.MatchString(normalizedQuery) {
			return false, "This operation is permanently blocked as it could be catastrophic to the system or expose sensitive data."
		}
	}

	if !allowWriteOperations && !IsReadOnlyQuery(query) {
		return false, "Query not in whitelist of safe read-only operations. Only SELECT, SHOW, DESCRIBE, EXPLAIN, and database introspection commands are allowed. Use allow_write_operations=true to enable write operations."
	}

	return true, ""
}

// NormalizeQuery normalizes a SQL query for analysis
func normalizeQuery(query string) string {
	// Remove comments
	query = regexp.MustCompile(`\/\*[\s\S]*?\*\/`).ReplaceAllString(query, " ")
	query = regexp.MustCompile(`--.*$`).ReplaceAllString(query, " ")

	// Normalize whitespace
	var normalized strings.Builder
	for i, char := range query {
		if unicode.IsSpace(char) {
			if i == 0 || !unicode.IsSpace(rune(query[i-1])) {
				normalized.WriteRune(' ')
			}
		} else {
			normalized.WriteRune(char)
		}
	}

	return strings.TrimSpace(strings.ToUpper(normalized.String()))
}

// BuildDatabaseCommand builds a DDEV exec command for database operations
func BuildDatabaseCommand(dbType DatabaseType, database string, query string) string {
	config := GetDatabaseConfig(dbType)

	// Escape quotes in query
	escapedQuery := strings.ReplaceAll(query, `"`, `\"`)

	var command strings.Builder
	command.WriteString(fmt.Sprintf("exec %s", config.Client))

	if database != "" {
		if dbType == DatabaseTypePostgres {
			command.WriteString(fmt.Sprintf(" -d %s", database))
		} else {
			command.WriteString(fmt.Sprintf(" -D %s", database))
		}
	}

	command.WriteString(fmt.Sprintf(" %s \"%s\"", config.CommandFlag, escapedQuery))

	return command.String()
}
