package services

import (
	"byod/common"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
)

// ValidationResponse defines the structure for API responses of validation requests.
type ValidationResponse struct {
	Status    string `json:"status"`
	ElementId string `json:"elementId"`
	Data      string `json:"data"`
}

// ValidationInfo defines the expected structure for validation requests.
type ValidationInfo struct {
	OS                   string `json:"os"`
	UDID                 string `json:"udid"`
	Package              string `json:"package"`
	Action               string `json:"action"`
	XPath                string `json:"xpath"`
	Value                string `json:"value"`
	Keys                 string `json:"keys"`
	Context              string `json:"context"`
	SessionId            string `json:"sessionId"`
	ElementId            string `json:"elementId"`
	WdaPort              string `json:"wdaPort"`
	CaseSensitiveLocator bool   `json:"caseSensitiveLocator"`
	InstrumentedFallback bool   `json:"instrumentedFallback"`
}

// ValidationHandler processes device interaction validation requests.
func ValidationHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Read the entire request body
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error reading request body: %v", err), http.StatusInternalServerError)
		return
	}

	// Parse the JSON request body into ValidationInfo struct
	var validationInfo ValidationInfo
	if err := json.Unmarshal(bodyBytes, &validationInfo); err != nil {
		http.Error(w, fmt.Sprintf("Error decoding request body: %v", err), http.StatusBadRequest)
		return
	}

	// Set up proxy to forward the request to the appropriate device IP and port
	deviceIP, port := getDeviceNetworkConfig(validationInfo.UDID, validationInfo.OS, validationInfo.Package)
	targetURL, err := url.Parse(fmt.Sprintf("http://%s:%s", deviceIP, port))
	if err != nil {
		http.Error(w, fmt.Sprintf("Error parsing target URL: %v", err), http.StatusInternalServerError)
		return
	}

	// Configure the reverse proxy
	proxy := configureProxy(targetURL, validationInfo.Action)
	proxy.ServeHTTP(w, r)
}

// getDeviceNetworkConfig retrieves the IP and port for the given device, package, and OS.
func getDeviceNetworkConfig(udid, os, pkg string) (string, string) {
	deviceIp, _ := common.FindDeviceIP(udid, os) // In production, handle errors appropriately.
	port, _ := common.GetForegroundApp(udid, pkg, os)
	return deviceIp, port
}

// configureProxy sets up a reverse proxy with specified target URL and action.
func configureProxy(targetURL *url.URL, action string) *httputil.ReverseProxy {
	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	proxy.Director = func(req *http.Request) {
		req.URL.Scheme = targetURL.Scheme
		req.URL.Host = targetURL.Host
		req.Host = targetURL.Host
		req.URL.Path = action
		req.RequestURI = "" // Clear the RequestURI to prevent errors on client redirects
	}
	proxy.ModifyResponse = func(resp *http.Response) error {
		resp.Header.Set("Connection", "close")
		return nil
	}
	return proxy
}
