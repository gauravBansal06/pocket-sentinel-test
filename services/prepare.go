package services

import (
	"byod/common"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func Initialize() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Failed to get home directory: %s", err)
	}

	common.AppDirs = common.AppDirectories{
		WorkingDir:   filepath.Join(homeDir, ".lambdatest"),
		Assets:       filepath.Join(homeDir, ".lambdatest", "assets"),
		TestInfo:     filepath.Join(homeDir, ".lambdatest", "tests"),
		Videos:       filepath.Join(homeDir, ".lambdatest", "videos"),
		CommandLogs:  filepath.Join(homeDir, ".lambdatest", "commandlogs"),
		AppiumLogs:   filepath.Join(homeDir, ".lambdatest", "appiumlogs"),
		BinaryLogs:   filepath.Join(homeDir, ".lambdatest", "binarylogs"),
		Screenshots:  filepath.Join(homeDir, ".lambdatest", "screenshots"),
		Applications: filepath.Join(homeDir, ".lambdatest", "applications"),
		DiskImages:   filepath.Join(homeDir, ".lambdatest", "assets", "diskimages"),
	}

	dirs := []string{
		common.AppDirs.WorkingDir,
		common.AppDirs.Assets,
		common.AppDirs.TestInfo,
		common.AppDirs.Videos,
		common.AppDirs.BinaryLogs,
		common.AppDirs.AppiumLogs,
		common.AppDirs.CommandLogs,
		common.AppDirs.Screenshots,
		common.AppDirs.Applications,
		common.AppDirs.DiskImages,
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Fatalf("Failed to create directory '%s': %s", dir, err)
		}
	}

	sanitize()

	common.Adb = fmt.Sprintf("%s/adb", common.AppDirs.Assets)
	common.GoIOS = fmt.Sprintf("%s/go-ios", common.AppDirs.Assets)
}

func sanitize() {
	target := fmt.Sprintf("%s/%s", common.AppDirs.Assets, "adb")
	_, err := os.Stat(target)
	if err != nil {
		source := fmt.Sprintf("%s/%s", common.SanitisatioEndpoint, "adb")
		common.Download(source, target)
	}
	target = fmt.Sprintf("%s/%s", common.AppDirs.Assets, "go-ios")
	_, err = os.Stat(target)
	if err != nil {
		source := fmt.Sprintf("%s/%s", common.SanitisatioEndpoint, "go-ios")
		common.Download(source, target)
	}
	target = fmt.Sprintf("%s/%s", common.AppDirs.Assets, "DYLIBS")
	_, err = os.Stat(target)
	if err != nil {
		source := fmt.Sprintf("%s/%s", common.SanitisatioEndpoint, "DYLIBS.zip")
		target = fmt.Sprintf("%s.zip", target)
		common.Download(source, target)
		common.Unzip(target, common.AppDirs.Assets)
	}
	target = fmt.Sprintf("%s/%s", common.AppDirs.Assets, "optool")
	_, err = os.Stat(target)
	if err != nil {
		source := fmt.Sprintf("%s/%s", common.SanitisatioEndpoint, "optool")
		common.Download(source, target)
	}
	target = fmt.Sprintf("%s/%s", common.AppDirs.Assets, "WebDriverAgentRunner-Runner.app")
	_, err = os.Stat(target)
	if err != nil {
		source := fmt.Sprintf("%s/%s", common.SanitisatioEndpoint, "WebDriverAgentRunner-Runner.zip")
		target = strings.Replace(target, ".app", ".zip", -1)
		common.Download(source, target)
		common.Unzip(target, common.AppDirs.Assets)
	}
	target = fmt.Sprintf("%s/%s", common.AppDirs.Assets, "android-sdk.zip")
	_, err = os.Stat(target)
	if err != nil {
		source := fmt.Sprintf("%s/%s", common.SanitisatioEndpoint, "android-sdk.zip")
		common.Download(source, target)
	}
}
