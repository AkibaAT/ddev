package ddevapp

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/ddev/ddev/pkg/archive"
	"github.com/ddev/ddev/pkg/fileutil"
	"github.com/ddev/ddev/pkg/nodeps"
	"github.com/ddev/ddev/pkg/util"
	copy2 "github.com/otiai10/copy"
)

// isCraftCmsApp returns true if the app is of type craftcms
func isCraftCmsApp(app *DdevApp) bool {
	return fileutil.FileExists(filepath.Join(app.AppRoot, app.ComposerRoot, "craft"))
}

// craftCmsImportFilesAction defines the workflow for importing project files.
func craftCmsImportFilesAction(app *DdevApp, uploadDir, importPath, extPath string) error {
	destPath := app.calculateHostUploadDirFullPath(uploadDir)

	// parent of destination dir should exist
	if !fileutil.FileExists(filepath.Dir(destPath)) {
		return fmt.Errorf("unable to import to %s: parent directory does not exist", destPath)
	}

	// parent of destination dir should be writable.
	if err := util.Chmod(filepath.Dir(destPath), 0755); err != nil {
		return err
	}

	// If the destination path exists, purge it as was warned
	if fileutil.FileExists(destPath) {
		if err := fileutil.PurgeDirectory(destPath); err != nil {
			return fmt.Errorf("failed to cleanup %s before import: %v", destPath, err)
		}
	}

	if isTar(importPath) {
		if err := archive.Untar(importPath, destPath, extPath); err != nil {
			return fmt.Errorf("failed to extract provided archive: %v", err)
		}

		return nil
	}

	if isZip(importPath) {
		if err := archive.Unzip(importPath, destPath, extPath); err != nil {
			return fmt.Errorf("failed to extract provided archive: %v", err)
		}

		return nil
	}

	if err := copy2.Copy(importPath, destPath); err != nil {
		return err
	}

	return nil
}

// Set up the .ddev/.env.web file for ddev
func updateCraftCMSDotEnv(app *DdevApp) (string, error) {
	// If settings management is disabled, do nothing
	if app.DisableSettingsManagement {
		return "", nil
	}

	// Check version is v4 or higher or warn user about app type mismatch.
	if !isCraftCms4orHigher(app) {
		util.Warning("It looks like the installed Craft CMS is lower than version 4 where it's recommended to use project type `php` or disable settings management with `ddev config --disable-settings-management`")
		if !util.Confirm("Would you like to stop here, not do the automatic configuration and change project type?") {
			return "", nil
		}
	}

	envFilePath := app.SiteSettingsPath

	port := "3306"
	driver := "mysql"
	if app.Database.Type == nodeps.Postgres {
		driver = "pgsql"
		port = "5432"
	}

	newCraftEnvMap := map[string]string{
		"CRAFT_DB_DRIVER":       driver,
		"CRAFT_DB_SERVER":       "db",
		"CRAFT_DB_PORT":         port,
		"CRAFT_DB_DATABASE":     "db",
		"CRAFT_DB_USER":         "db",
		"CRAFT_DB_PASSWORD":     "db",
		"CRAFT_WEB_ROOT":        app.GetAbsDocroot(true),
		"MAILPIT_SMTP_HOSTNAME": "127.0.0.1",
		"MAILPIT_SMTP_PORT":     "1025",
		"PRIMARY_SITE_URL":      app.GetPrimaryURL(),
	}

	// Read existing env file
	_, existingEnvText, err := ReadProjectEnvFile(envFilePath)
	// If envFilePath doesn't exist, that's not really an error, continue
	if err != nil && !os.IsNotExist(err) {
		return "", err
	}

	err = WriteProjectEnvFile(envFilePath, newCraftEnvMap, existingEnvText)
	if err != nil {
		return "", err
	}

	return "", nil
}

func craftCmsConfigOverrideAction(app *DdevApp) error {
	app.Database = DatabaseDesc{nodeps.MySQL, nodeps.MySQL80}
	return nil
}

// isCraftCms4orHigher returns true if the Craft CMS version is 4 or higher. The
// proper detection will fail if the vendor folder location is changed in the
// composer.json.
// The detection is based on a change starting with 4.0.0-RC1 where deprecated
// constants were removed in src/Craft.php see
// https://github.com/craftcms/cms/commit/1660ff90a3a69cec425271d47ade66523a4bd44e#diff-21e22a30e7c48265a4dcedc1b1c8b9372eca5d3fdeff6d72c7d9c6b671365c56
func isCraftCms4orHigher(app *DdevApp) bool {
	craftFilePath := filepath.Join(app.GetComposerRoot(false, false), "vendor", "craftcms", "cms", "src", "Craft.php")
	if !fileutil.FileExists(craftFilePath) {
		// Sources are not installed, assuming v4 or higher.
		return true
	}

	craftFileContent, err := fileutil.ReadFileIntoString(craftFilePath)
	if err != nil {
		util.Warning("unable to read file `%s` in project `%s`: %v", craftFilePath, app.Name, err)

		return true
	}

	return !regexp.MustCompile(`const\s+Personal\s*=\s*0`).MatchString(craftFileContent)
}

func setCraftCMSDotFileLocation(app *DdevApp) {
	app.SiteSettingsPath = app.GetConfigPath(".env.web")
}
