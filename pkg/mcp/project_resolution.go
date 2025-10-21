package mcp

import (
	"fmt"

	"github.com/ddev/ddev/pkg/ddevapp"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ResolveProjectName resolves the project name based on DefaultProject settings
// Returns the resolved project name or an error if access is denied
func ResolveProjectName(requestedProject string, settings ServerSettings) (string, error) {
	// If no default project is set, return the requested project (multi-project mode)
	if settings.DefaultProject == "" {
		if requestedProject == "" {
			return "", fmt.Errorf("no project specified and no default project configured")
		}
		return requestedProject, nil
	}

	// If default project is set, enforce single-project mode
	if requestedProject == "" {
		return settings.DefaultProject, nil
	}

	if requestedProject != settings.DefaultProject {
		return "", fmt.Errorf("access denied: project '%s' is not the configured single project '%s'", requestedProject, settings.DefaultProject)
	}

	return settings.DefaultProject, nil
}

// GetPinnedProjectDescription returns a description of the current pinning mode
func GetPinnedProjectDescription(settings ServerSettings) string {
	if settings.DefaultProject == "" {
		return "🔄 Multi-Project Mode: Can access any DDEV project"
	}
	return fmt.Sprintf("📌 Single-Project Mode: Pinned to '%s' (access denied to other projects)", settings.DefaultProject)
}

// ValidateProjectAccess checks if a tool call should be allowed based on project settings
func ValidateProjectAccess(requestedProject string, settings ServerSettings, toolName string) error {
	resolvedProject, err := ResolveProjectName(requestedProject, settings)
	if err != nil {
		return err
	}

	// For security-sensitive operations, add extra validation in single-project mode
	if settings.DefaultProject != "" && isDestructiveOperation(toolName) {
		return fmt.Errorf("access denied: destructive operation '%s' is not allowed in single-project mode", toolName)
	}

	return nil
}

// isDestructiveOperation returns true if the tool is potentially destructive
func isDestructiveOperation(toolName string) bool {
	destructiveTools := []string{
		"ddev_exec_command",
		"ddev_update_config",
		"ddev_composer_command",
	}

	for _, tool := range destructiveTools {
		if toolName == tool {
			return true
		}
	}
	return false
}
