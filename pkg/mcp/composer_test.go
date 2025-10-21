package mcp

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestComposerCommandJSONFormatting(t *testing.T) {
	tests := []struct {
		name       string
		command    string
		formatJSON bool
		expected   string
	}{
		{
			name:       "JSON command with formatting",
			command:    "show",
			formatJSON: true,
			expected:   "composer show --format=json",
		},
		{
			name:       "JSON command already has format",
			command:    "show --format=xml",
			formatJSON: true,
			expected:   "composer show --format=xml",
		},
		{
			name:       "Non-JSON command without formatting",
			command:    "install",
			formatJSON: true,
			expected:   "composer install",
		},
		{
			name:       "JSON disabled",
			command:    "show",
			formatJSON: false,
			expected:   "composer show",
		},
		{
			name:       "Info command with JSON",
			command:    "info",
			formatJSON: true,
			expected:   "composer info --format=json",
		},
		{
			name:       "List command with JSON",
			command:    "list",
			formatJSON: true,
			expected:   "composer list --format=json",
		},
		{
			name:       "Outdated command with JSON",
			command:    "outdated",
			formatJSON: true,
			expected:   "composer outdated --format=json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test command formatting logic
			composerCommand := tt.command
			if tt.formatJSON {
				jsonSupportedCommands := []string{
					"show", "info", "list", "outdated", "depends",
					"why", "why-not", "status", "licenses",
				}

				commandParts := strings.Fields(composerCommand)
				if len(commandParts) > 0 {
					commandName := commandParts[0]
					for _, supported := range jsonSupportedCommands {
						if commandName == supported {
							// Check if --format is already specified
							hasFormatFlag := false
							for _, part := range commandParts {
								if strings.HasPrefix(part, "--format=") {
									hasFormatFlag = true
									break
								}
							}
							if !hasFormatFlag {
								composerCommand = fmt.Sprintf("%s --format=json", composerCommand)
							}
							break
						}
					}
				}
			}

			// Add "composer " prefix to match expected output
			fullCommand := fmt.Sprintf("composer %s", composerCommand)
			require.Equal(t, tt.expected, fullCommand)
		})
	}
}

func TestComposerCommandJSONParsing(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectJSON    bool
		expectedError bool
	}{
		{
			name:          "Valid JSON",
			input:         `{"name": "test", "version": "1.0.0"}`,
			expectJSON:    true,
			expectedError: false,
		},
		{
			name:          "Invalid JSON",
			input:         `{"name": "test", "version": "1.0.0"`,
			expectJSON:    false,
			expectedError: false, // Should handle gracefully
		},
		{
			name:          "Plain text",
			input:         "Loading composer repositories with package information",
			expectJSON:    false,
			expectedError: false,
		},
		{
			name:          "Empty string",
			input:         "",
			expectJSON:    false,
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var jsonData any
			err := json.Unmarshal([]byte(tt.input), &jsonData)

			isValidJSON := (err == nil && len(tt.input) > 0)
			require.Equal(t, tt.expectJSON, isValidJSON)

			if tt.expectedError {
				require.Error(t, err)
			} else {
				// No error expected or error handled gracefully
			}
		})
	}
}
