package services

import (
	"byod/common"
	"bytes"
	"io"
	"net/http"
	"time"
)

// MakePostRequest sends a POST request to the specified endpoint with the given payload.
// It returns the HTTP status code as a string, the response body as a byte slice, and any error encountered.
func MakePostRequest(endpoint string, payload []byte) (string, []byte, error) {
	client := &http.Client{Timeout: 5 * time.Second}

	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return http.StatusText(http.StatusBadRequest), nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic "+common.SyncToken)

	resp, err := client.Do(req)
	if err != nil {
		// Return a general "Service Unavailable" as status since the request didn't go through
		return http.StatusText(http.StatusServiceUnavailable), nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		// Return the actual status from the response, even if reading the body failed
		return resp.Status, nil, err
	}

	return resp.Status, body, nil
}
