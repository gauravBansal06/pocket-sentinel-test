package services

import (
	"byod/common"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

// RequestInfo represents the JSON structure for incoming requests.
type RequestInfo struct {
	OS      string `json:"os"`
	UDID    string `json:"udid"`
	AppPath string `json:"appPath"`
	Package string `json:"package"`
	Action  string `json:"action"`
}

// AppInfo represents the structure for a single application.
type AppInfo struct {
	Name    string `json:"name"`
	Package string `json:"package"`
	Version string `json:"version"`
}

// AppResponse represents the JSON structure for outgoing responses.
type AppResponse struct {
	Status string    `json:"status"`
	Apps   []AppInfo `json:"apps"`
}

// ApplicationHandler handles different application actions such as install, uninstall, etc.
func ApplicationHandler(w http.ResponseWriter, r *http.Request) {
	var requestInfo RequestInfo
	if err := json.NewDecoder(r.Body).Decode(&requestInfo); err != nil {
		http.Error(w, `{"status":"failed"}`, http.StatusBadRequest)
		return
	}

	log.Println("action", requestInfo.Action, "os", requestInfo.OS, "udid", requestInfo.UDID, "appPath", requestInfo.AppPath, "package", requestInfo.Package)
	var response AppResponse
	switch requestInfo.Action {
	case "install":
		response.Status = executeAppAction(installApp(requestInfo.OS, requestInfo.UDID, requestInfo.AppPath))
	case "uninstall":
		response.Status = executeAppAction(uninstallApp(requestInfo.OS, requestInfo.UDID, requestInfo.Package))
	case "launch":
		response.Status = executeAppAction(launchApp(requestInfo.OS, requestInfo.UDID, requestInfo.Package))
	case "kill":
		response.Status = executeAppAction(killApp(requestInfo.OS, requestInfo.UDID, requestInfo.Package))
	case "apps":
		response.Apps = ListApps(requestInfo.OS, requestInfo.UDID)
		response.Status = "success"
	default:
		response.Status = "invalid action"
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// executeAppAction processes the error from app management actions and returns the appropriate status.
func executeAppAction(err error) string {
	if err != nil {
		log.Printf("ApplicationHandler: Error occurred: %v\n", err)
		return "failed"
	}
	return "success"
}

// installApp installs an app on a device identified by OS and UDID.
func installApp(os, udid, appPath string) error {
	filePath, err := common.DownloadAppIfRequired(appPath)
	if err != nil {
		return err
	}
	var command string
	if os == "android" {
		command = fmt.Sprintf("%s -s %s install -t %s", common.Adb, udid, filePath)
	} else {
		command = fmt.Sprintf("%s install --path=%s --udid %s", common.GoIOS, filePath, udid)
	}
	_, err = common.Execute(command)
	return err
}

// uninstallApp uninstalls an app from a device.
func uninstallApp(os, udid, bundle string) error {
	var command string
	if os == "android" {
		command = fmt.Sprintf("%s -s %s uninstall %s", common.Adb, udid, bundle)
	} else {
		command = fmt.Sprintf("%s uninstall %s --udid %s", common.GoIOS, bundle, udid)
	}
	_, err := common.Execute(command)
	return err
}

// launchApp launches an app on a device.
func launchApp(os, udid, bundle string) error {
	var command string
	if os == "android" {
		command = fmt.Sprintf("%s -s %s shell monkey -p %s -c android.intent.category.LAUNCHER 1", common.Adb, udid, bundle)
	} else {
		command = fmt.Sprintf("%s launch %s --udid %s", common.GoIOS, bundle, udid)
	}
	_, err := common.Execute(command)
	return err
}

// killApp force-stops an app on a device.
func killApp(os, udid, bundle string) error {
	var command string
	if os == "android" {
		command = fmt.Sprintf("%s -s %s shell am force-stop %s", common.Adb, udid, bundle)
	} else {
		command = fmt.Sprintf("%s kill %s --udid %s", common.GoIOS, bundle, udid)
	}
	_, err := common.Execute(command)
	return err
}

// ListApps lists all installed apps on a device.
func ListApps(os, udid string) []AppInfo {
	var appList []AppInfo
	if os == "android" {
		command := fmt.Sprintf("%s -s %s shell 'pm list packages -3 | cut -d ':' -f2 | while read line; do version=`dumpsys package $line | grep versionName | cut -d '=' -f2`; echo \"$line $version\"; done'", common.Adb, udid)
		output, err := common.Execute(command)
		if err == nil {
			apps := strings.Split(output, "\n")
			for _, app := range apps {
				appInfo := strings.Split(app, " ")
				appList = append(appList, AppInfo{
					Package: appInfo[0],
					Version: appInfo[1],
				})
			}
			return appList
		}
		log.Println("error while getting app list", err)
	} else {
		command := fmt.Sprintf("%s apps --list --udid %s", common.GoIOS, udid)
		output, err := common.Execute(command)
		if err == nil {
			apps := strings.Split(output, "\n")
			for _, app := range apps {
				appInfo := strings.Split(app, " ")
				appList = append(appList, AppInfo{
					Name:    appInfo[1],
					Package: appInfo[0],
					Version: appInfo[2],
				})
			}
			return appList
		}
	}
	return appList
}

// parseAppList parses the command line output into a slice of AppInfo.
func parseAppList(output string) []AppInfo {
	var apps []AppInfo
	for _, line := range strings.Split(output, "\n") {
		parts := strings.Split(line, " ")
		if len(parts) >= 3 {
			apps = append(apps, AppInfo{Name: parts[1], Package: parts[0], Version: parts[2]})
		}
	}
	return apps
}
