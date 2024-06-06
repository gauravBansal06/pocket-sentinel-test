package services

import (
	"byod/common"
	"byod/storage"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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

// getSessionID extracts the session ID from the URL path using a regular expression.
func getSessionID(path string) string {
	if matches := regexSessionID.FindStringSubmatch(path); len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// getOrCreateProxy retrieves an existing reverse proxy for the target URL or creates a new one.
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
		if strings.Contains(resp.Header.Get("Content-Type"), "application/json") {
			originalBody, err := io.ReadAll(resp.Body)
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

			resp.Body = io.NopCloser(bytes.NewReader(originalBody))
		}
		return nil
	}

	ReverseProxyMap.Store(targetURL, proxy)
	return proxy
}

// SessionHandler handles incoming session requests, either creating a new session or managing existing ones.
func SessionHandler(res http.ResponseWriter, req *http.Request) {
	var testInfo common.TestInfo
	if err := json.NewDecoder(req.Body).Decode(&testInfo); err != nil {
		http.Error(res, "Invalid request body", http.StatusBadRequest)
		return
	}

	sessionID := getSessionID(req.URL.Path)
	if sessionID == "" && req.URL.Path == "/wd/hub/session" && req.Method == "POST" {
		handleNewSession(res, req, testInfo)
	} else if proxy, ok := ReverseProxyMap.Load(sessionID); ok {
		if req.Method == "DELETE" && strings.HasPrefix(req.URL.Path, "/wd/hub/session") {
			handleSessionDeletion(res, req, proxy.(*httputil.ReverseProxy), testInfo.UDID)
		} else {
			proxy.(*httputil.ReverseProxy).ServeHTTP(res, req)
		}
	} else {
		http.Error(res, `{"status": "invalid session"}`, http.StatusBadRequest)
	}
}

// handleNewSession processes the creation of a new Appium session.
func handleNewSession(res http.ResponseWriter, req *http.Request, testInfo common.TestInfo) {
	go launchApp(testInfo.OS, testInfo.UDID, testInfo.AppPackage)
	os.Create(fmt.Sprintf("%s/%s.json", common.AppDirs.TestInfo, testInfo.TestID))

	port := startAppium(testInfo.UDID, testInfo.TestID)
	targetURL := "http://localhost:" + port
	proxy := getOrCreateProxy(targetURL)

	if testInfo.TestType == "manual" {
		req.Body, req.ContentLength = getSessionPayload(testInfo)
	}
	proxy.ServeHTTP(res, req)
}

// handleSessionDeletion handles the deletion of an Appium session.
func handleSessionDeletion(res http.ResponseWriter, req *http.Request, proxy *httputil.ReverseProxy, udid string) {
	req.Body = nil
	req.ContentLength = 0
	proxy.ServeHTTP(res, req)
	go stopAppium(udid)
}

// startAppium starts the Appium server for the given UDID and test ID.
func startAppium(udid, testId string) string {
	var port string
	appiumLogs := fmt.Sprintf("%s/%s.log", common.AppDirs.AppiumLogs, testId)
	os.Remove(appiumLogs)
	storage.Store.Get("Appium_Port_"+udid, &port)
	cmd, err := common.ExecuteAsync(fmt.Sprintf("appium --base-path /wd/hub -p %s --log %s", port, appiumLogs))
	if err != nil {
		log.Printf("Failed to execute command: %s\n", err)
		return ""
	}
	AppiumServers.Store(udid, cmd)
	time.Sleep(5 * time.Second)
	return port
}

// stopAppium stops the Appium server for the given UDID.
func stopAppium(udid string) {
	if cmd, _ := AppiumServers.Load(udid); cmd != nil {
		cmd.(*exec.Cmd).Process.Kill()
		AppiumServers.Delete(udid)
	}
	var port string
	storage.Store.Get("Appium_Port_"+udid, &port)
	if port != "" {
		common.KillProcessOnPort(port)
	}
}

// getSessionPayload generates the payload for starting a new Appium session.
func getSessionPayload(testInfo common.TestInfo) (io.ReadCloser, int64) {
	automationName := "UiAutomator2"
	if testInfo.OS == "ios" {
		automationName = "XCUITest"
	}
	payload := common.WebDriver{
		Capabilities: common.Capabilities{
			AlwaysMatch: common.AlwaysMatch{
				AppiumUDID:                    testInfo.UDID,
				PlatformName:                  testInfo.OS,
				AppiumAutomationName:          automationName,
				AppiumNoReset:                 true,
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
	readCloser := io.NopCloser(reader)
	return readCloser, int64(len(jsonData))
}
