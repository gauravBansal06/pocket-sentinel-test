package services

import (
	"byod/common"
	"byod/instrument"
	"encoding/json"
	"fmt"
	"net/http"
)

type RequestInfo struct {
	OS      string
	UDID    string
	AppPath string
	Package string
	Action  string
}

type AppInfo struct {
	Package string
	Version string
	Size    string
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
		err = installApp(requestInfo)
	} else if requestInfo.Action == "uninstall" {
		err = uninstallApp(requestInfo)
	} else if requestInfo.Action == "launch" {
		err = launchApp(requestInfo)
	} else if requestInfo.Action == "kill" {
		err = killApp(requestInfo)
	} else if requestInfo.Action == "apps" {
		response.Apps = ListApps(requestInfo)
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

func installApp(requestInfo RequestInfo) error {
	var command string
	filaPath, err := common.DownloadAppIfRequired(requestInfo.AppPath)
	if err == nil {
		requestInfo.AppPath = filaPath
		if requestInfo.OS == "android" {
			command = fmt.Sprintf("%s -s %s install -t %s", common.Adb, requestInfo.UDID, requestInfo.AppPath)
		} else {
			instrument.ResigniOSApp(requestInfo.AppPath)
			command = fmt.Sprintf("%s install --path=%s --udid %s", common.GoIOS, requestInfo.AppPath, requestInfo.UDID)
		}
		_, err = common.Execute(command)
	}
	return err
}

func uninstallApp(requestInfo RequestInfo) error {
	var command string
	if requestInfo.OS == "android" {
		command = fmt.Sprintf("%s -s %s uninstall %s", common.Adb, requestInfo.UDID, requestInfo.Package)
	} else {
		command = fmt.Sprintf("%s uninstall %s --udid %s", common.GoIOS, requestInfo.Package, requestInfo.UDID)
	}
	_, err := common.Execute(command)
	return err
}

func launchApp(requestInfo RequestInfo) error {
	var command string
	if requestInfo.OS == "android" {
		command = fmt.Sprintf("%s -s %s shell monkey -p %s -c android.intent.category.LAUNCHER 1", common.Adb, requestInfo.UDID, requestInfo.Package)
	} else {
		command = fmt.Sprintf("%s launch %s --udid %s", common.GoIOS, requestInfo.Package, requestInfo.UDID)
	}
	_, err := common.Execute(command)
	return err
}

func killApp(requestInfo RequestInfo) error {
	var command string
	if requestInfo.OS == "android" {
		command = fmt.Sprintf("%s -s %s, shell am force-stop %s", common.Adb, requestInfo.UDID, requestInfo.Package)
	} else {
		command = fmt.Sprintf("%s kill %s --udid %s", common.GoIOS, requestInfo.Package, requestInfo.UDID)
	}
	_, err := common.Execute(command)
	return err
}

func ListApps(requestInfo RequestInfo) []AppInfo {
	var appList []AppInfo
	return appList
}
