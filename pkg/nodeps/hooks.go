package nodeps

import "fmt"

// YAMLTask defines tasks like Exec to be run in hooks
type YAMLTask map[string]interface{}

// ValidateHooks checks that the given hooks map uses only valid hook names and task types.
// The source parameter is used in error messages to identify where the invalid config came from.
func ValidateHooks(hooks map[string][]YAMLTask, source string) error {
	for hookName, tasks := range hooks {
		if !contains(ValidHookNames, hookName) {
			return fmt.Errorf("invalid hook %s defined in %s", hookName, source)
		}
		for _, task := range tasks {
			var match bool
			for _, validTaskName := range ValidTaskNames {
				if _, ok := task[validTaskName]; ok {
					match = true
				}
			}
			if !match {
				return fmt.Errorf("invalid task '%v' defined for hook %s in %s", task, hookName, source)
			}
		}
	}
	return nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// ValidHookNames is the list of valid hook names for DDEV config
var ValidHookNames = []string{
	"pre-start",
	"post-start",
	"pre-import-db",
	"post-import-db",
	"pre-import-files",
	"post-import-files",
	"pre-composer",
	"post-composer",
	"pre-stop",
	"post-stop",
	"pre-config",
	"post-config",
	"pre-describe",
	"post-describe",
	"pre-exec",
	"post-exec",
	"pre-pause",
	"post-pause",
	"pre-pull",
	"post-pull",
	"pre-push",
	"post-push",
	"pre-share",
	"post-share",
	"pre-snapshot",
	"post-snapshot",
	"pre-restore-snapshot",
	"post-restore-snapshot",
}

// ValidTaskNames is the list of valid task types for DDEV hooks
var ValidTaskNames = []string{
	"exec",
	"exec-host",
	"composer",
}
