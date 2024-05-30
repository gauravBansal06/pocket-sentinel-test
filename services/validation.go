package services

import (
	"byod/common"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
)

type ValidationResponse struct {
	Status    string `json:"status"`
	ElementId string `json:"elementId"`
	Data      string `json:"data"`
}

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

func ValidationHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var response ValidationResponse
	response.Status = "success"

	var validationInfo ValidationInfo
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error reading request body: %v", err), http.StatusInternalServerError)
		return
	}
	if err := json.Unmarshal(bodyBytes, &validationInfo); err != nil {
		http.Error(w, fmt.Sprintf("Error decoding request body: %v", err), http.StatusBadRequest)
		return
	}
	r.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	r.ContentLength = int64(len(bodyBytes))
	r.Header.Set("Content-Length", strconv.Itoa(len(bodyBytes)))

	udid := validationInfo.UDID
	deviceIp, _ := common.FindDeviceIP(udid, validationInfo.OS)
	port, _ := common.GetForegroundApp(udid, validationInfo.Package, validationInfo.OS)
	targetURL, _ := url.Parse(fmt.Sprintf("http://%s:%s", deviceIp, port))

	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	proxy.Director = func(req *http.Request) {
		req.URL.Scheme = targetURL.Scheme
		req.URL.Host = targetURL.Host
		req.Host = targetURL.Host
		req.Header = r.Clone(r.Context()).Header
		req.URL.Path = validationInfo.Action
		req.RequestURI = validationInfo.Action
	}
	proxy.ModifyResponse = func(resp *http.Response) error {
		resp.Header.Set("Connection", "close")
		return nil
	}
	proxy.ServeHTTP(w, r)
}
