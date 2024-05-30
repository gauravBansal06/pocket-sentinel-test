package services

import (
	"byod/common"
	"byod/storage"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"
)

var (
	AppiumServers   sync.Map
	ReverseProxyMap sync.Map
	regexSessionID  = regexp.MustCompile(`^/wd/hub/session(?:/([^/]+))?$`)
)

func getSessionID(path string) string {
	if matches := regexSessionID.FindStringSubmatch(path); len(matches) > 1 {
		return matches[1]
	}
	return ""
}

func getOrCreateProxy(targetURL string) *httputil.ReverseProxy {
	if proxy, found := ReverseProxyMap.Load(targetURL); found {
		return proxy.(*httputil.ReverseProxy)
	}
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		log.Printf("Error parsing server URL '%s': %v", targetURL, err)
		return nil
	}
	proxy := httputil.NewSingleHostReverseProxy(parsedURL)
	proxy.Director = func(req *http.Request) {
		req.URL.Scheme = parsedURL.Scheme
		req.URL.Host = parsedURL.Host
		req.Host = parsedURL.Host
		req.Header.Set("X-Forwarded-Host", req.Header.Get("Host"))
	}

	proxy.ModifyResponse = func(resp *http.Response) error {
		if resp.Header.Get("Content-Type") == "application/json" {
			originalBody, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return err
			}
			defer resp.Body.Close()
			var testInfo common.TestInfo
			if err := json.Unmarshal(originalBody, &testInfo); err != nil {
				return err
			}
			if testInfo.Value.SessionID != "" {
				ReverseProxyMap.Store(testInfo.Value.SessionID, proxy)
			}
		}
		return nil
	}
	return proxy
}

func SessionHandler(res http.ResponseWriter, req *http.Request) {
	var testInfo common.TestInfo
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&testInfo)
	if err != nil {
		return
	}

	sessionID := getSessionID(req.URL.Path)
	if sessionID == "" && req.URL.Path == "/wd/hub/session" && req.Method == "POST" {
		go launchApp(testInfo.OS, testInfo.UDID, testInfo.AppPackage)
		os.Create(fmt.Sprintf("%s/%s.json", common.AppDirs.TestInfo, testInfo.TestID))
		port := startAppium(testInfo.UDID, testInfo.TestID)
		targetURL := "http://localhost:" + port
		proxy := getOrCreateProxy(targetURL)
		if testInfo.TestType == "manual" {
			req.Body, req.ContentLength = getSessionPayload(testInfo)
		}
		proxy.ServeHTTP(res, req)
	} else if proxy, ok := ReverseProxyMap.Load(sessionID); ok {
		if strings.HasPrefix(req.URL.Path, "/wd/hub/session") && req.Method == "DELETE" {
			req.Body = nil
			req.ContentLength = 0
			proxy.(*httputil.ReverseProxy).ServeHTTP(res, req)
			go stopAppium(testInfo.UDID)
		} else {
			proxy.(*httputil.ReverseProxy).ServeHTTP(res, req)
		}
	} else {
		res.Write([]byte(`{}"status": "invalid session"}`))
	}
}

func startAppium(udid, testId string) string {
	var port string
	appiumLogs := fmt.Sprintf("%s/%s.log", common.AppDirs.AppiumLogs, testId)
	os.Remove(appiumLogs)
	storage.Store.Get("Appium_Port_"+udid, &port)
	cmd, err := common.ExecuteAsync(fmt.Sprintf("appium --base-path /wd/hub -p %s --log %s", port, appiumLogs))
	if err != nil {
		fmt.Printf("Failed to execute command: %s\n", err)
		return ""
	}
	AppiumServers.Store(udid, cmd)
	time.Sleep(5 * time.Second)
	return port
}

func stopAppium(udid string) {
	cmd, _ := AppiumServers.Load(udid)
	if cmd != nil {
		cmd.(*exec.Cmd).Process.Kill()
		AppiumServers.Delete(udid)
	}
	var port string
	storage.Store.Get("Appium_Port_"+udid, &port)
	if port != "" {
		common.KillProcessOnPort(port)
	}
}

func getSessionPayload(testInfo common.TestInfo) (io.ReadCloser, int64) {
	automationName := "UiAutomator2"
	if testInfo.OS == "iOS" {
		automationName = "XCUITest"
	}
	payload := common.WebDriver{
		Capabilities: common.Capabilities{
			AlwaysMatch: common.AlwaysMatch{
				AppiumUDID:                    testInfo.UDID,
				PlatformName:                  testInfo.OS,
				AppiumAutomationName:          automationName,
				AppiumNoReset:                 true,
				AppiumAppPackage:              testInfo.AppPackage,
				AppiumEnsureWebviewsHavePages: true,
				AppiumNativeWebScreenshot:     true,
				AppiumNewCommandTimeout:       7200,
				AppiumConnectHardwareKeyboard: true,
			},
			FirstMatch: []struct{}{},
		},
		DesiredCapabilities: common.DesiredCapabilities{
			AppiumUDID:                    testInfo.UDID,
			PlatformName:                  testInfo.OS,
			AutomationName:                automationName,
			AppiumNoReset:                 true,
			AppiumAppPackage:              testInfo.AppPackage,
			AppiumAppActivity:             testInfo.AppActivity,
			AppiumEnsureWebviewsHavePages: true,
			AppiumNativeWebScreenshot:     true,
			AppiumNewCommandTimeout:       7200,
			AppiumConnectHardwareKeyboard: true,
		},
	}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, 0
	}
	reader := bytes.NewReader(jsonData)
	readCloser := ioutil.NopCloser(reader)
	return readCloser, int64(len(jsonData))
}
