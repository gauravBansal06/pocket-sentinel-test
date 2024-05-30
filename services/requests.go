package services

import (
	"byod/common"
	"bytes"
	"io/ioutil"
	"net/http"
	"time"
)

func MakePostRequest(endpoint string, payload []byte) (string, []byte, error) {
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(payload))
	if err != nil {
		return "400 BAD REQUEST", nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic "+common.SyncToken)
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "400 BAD REQUEST", nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return resp.Status, nil, err
	}
	return resp.Status, body, err
}
