package services

import (
	"byod/common"
	"byod/instrument"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type RequestInfo struct {
	OS      string `json:"os"`
	UDID    string `json:"udid"`
	AppPath string `json:"appPath"`
	Package string `json:"package"`
	Action  string `json:"action"`
}

type AppInfo struct {
	Name    string `json:"name"`
	Package string `json:"package"`
	Version string `json:"version"`
}

type AppResponse struct {
	Status string    `json:"status"`
	Apps   []AppInfo `json:"apps"`
}

func ApplicationHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	var requestInfo RequestInfo
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&requestInfo)
	if err != nil {
		w.Write([]byte(`{ "status": "failed"}`))
	}

	var response AppResponse
	if requestInfo.Action == "install" {
		err = installApp(requestInfo.OS, requestInfo.UDID, requestInfo.AppPath)
	} else if requestInfo.Action == "uninstall" {
		err = uninstallApp(requestInfo.OS, requestInfo.UDID, requestInfo.Package)
	} else if requestInfo.Action == "launch" {
		err = launchApp(requestInfo.OS, requestInfo.UDID, requestInfo.Package)
	} else if requestInfo.Action == "kill" {
		err = killApp(requestInfo.OS, requestInfo.UDID, requestInfo.Package)
	} else if requestInfo.Action == "apps" {
		response.Apps = ListApps(requestInfo.OS, requestInfo.UDID)
	}

	if err != nil {
		fmt.Printf("ApplicationHandler: Error in %s, error : %s\n", requestInfo.Action, err.Error())
		response.Status = "failed"
	} else {
		response.Status = "success"
	}
	encoder := json.NewEncoder(w)
	if err := encoder.Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func installApp(os, udid, appPath string) error {
	var command string
	filePath, err := common.DownloadAppIfRequired(appPath)
	if err == nil {
		appPath = filePath
		if os == "android" {
			command = fmt.Sprintf("%s -s %s install -t %s", common.Adb, udid, appPath)
		} else {
			instrument.ResigniOSApp(appPath)
			command = fmt.Sprintf("%s install --path=%s --udid %s", common.GoIOS, appPath, udid)
		}
		_, err = common.Execute(command)
	}
	return err
}

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

func killApp(os, udid, bundle string) error {
	var command string
	if os == "android" {
		command = fmt.Sprintf("%s -s %s, shell am force-stop %s", common.Adb, udid, bundle)
	} else {
		command = fmt.Sprintf("%s kill %s --udid %s", common.GoIOS, bundle, udid)
	}
	_, err := common.Execute(command)
	return err
}

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
		fmt.Println("error while getting app list", err)
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
		fmt.Println("error while getting app list", err)
	}
	return appList
}
