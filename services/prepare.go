package services

import (
	"byod/common"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func Initialize() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Failed to get home directory: %s", err)
	}
	baseDir := filepath.Join(homeDir, ".lambdatest")

	common.AppDirs = common.AppDirectories{
		WorkingDir:   baseDir,
		Assets:       filepath.Join(baseDir, "assets"),
		TestInfo:     filepath.Join(baseDir, "tests"),
		Videos:       filepath.Join(baseDir, "videos"),
		CommandLogs:  filepath.Join(baseDir, "commandlogs"),
		AppiumLogs:   filepath.Join(baseDir, "appiumlogs"),
		BinaryLogs:   filepath.Join(baseDir, "binarylogs"),
		Screenshots:  filepath.Join(baseDir, "screenshots"),
		Applications: filepath.Join(baseDir, "applications"),
		DiskImages:   filepath.Join(baseDir, "assets", "diskimages"),
		AppiumDir:    filepath.Join(baseDir, "assets", "appium"),
	}

	createDirectories()

	prepare()
	setEnvironmentVariables()
	setToolPaths()
}

func createDirectories() {
	for _, dir := range []string{
		common.AppDirs.WorkingDir,
		common.AppDirs.Assets,
		common.AppDirs.TestInfo,
		common.AppDirs.Videos,
		common.AppDirs.CommandLogs,
		common.AppDirs.AppiumLogs,
		common.AppDirs.BinaryLogs,
		common.AppDirs.Screenshots,
		common.AppDirs.Applications,
		common.AppDirs.DiskImages,
	} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Fatalf("Failed to create directory '%s': %s", dir, err)
		}
	}
}

func setEnvironmentVariables() {
	os.Setenv("ANDROID_HOME", common.AppDirs.Assets)
}

func setToolPaths() {
	common.Adb = fmt.Sprintf("%s/adb", common.AppDirs.Assets)
	common.GoIOS = fmt.Sprintf("%s/go-ios", common.AppDirs.Assets)
	common.Appium = fmt.Sprintf("%s/node_modules/.bin/appium", common.AppDirs.AppiumDir)
}

func prepare() {
	items := []struct {
		name       string
		compressed bool
	}{
		{"adb", false},
		{"go-ios", false},
		{"DYLIBS", true},
		{"optool", false},
		{"WebDriverAgentRunner-Runner.app", true},
		{"node", false},
		{"scrcpy-server.jar", false},
	}

	for _, item := range items {
		ensureFileExists(item.name, item.compressed)
	}

	appiumSetup()
}

func ensureFileExists(item string, isCompressed bool) {
	target := filepath.Join(common.AppDirs.Assets, item)
	_, err := os.Stat(target)
	if err == nil {
		return
	}
	ext := ""
	if isCompressed {
		ext = ".zip"
	}
	source := fmt.Sprintf("%s/%s%s", common.SanitisatioEndpoint, item, ext)
	common.Download(source, target+ext)
	if isCompressed {
		common.Unzip(target+ext, common.AppDirs.Assets)
	}
}

func appiumSetup() {
	appiumDir := common.AppDirs.AppiumDir

	// Commands to install Appium and its drivers
	commands := []string{
		fmt.Sprintf("npm install --prefix %s appium", common.AppDirs.AppiumDir),
		fmt.Sprintf("%s/node_modules/.bin/appium driver install xcuitest", appiumDir),
		fmt.Sprintf("%s/node_modules/.bin/appium driver install uiautomator2", appiumDir),
	}

	// Execute each command sequentially
	for _, cmd := range commands {
		log.Println(cmd)
		if _, err := common.Execute(cmd); err != nil {
			log.Printf("Failed to execute command '%s': %v", cmd, err)
		}
	}

}
