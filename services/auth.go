package services

import (
	"byod/common"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
)

var AuthenticatedUsers sync.Map

func isValidUser(token string) bool {
	token = strings.Replace(token, "Basic ", "", -1)
	bytes, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return false
	}
	token = string(bytes)
	array := strings.Split(token, ":")

	if _, ok := AuthenticatedUsers.Load(token); ok {
		return true
	}

	userInfo, err := AuthenticateUser(array[0], array[1])
	if err != nil {
		fmt.Println("Authentication failed for", token)
		return false
	}

	if userInfo.Organization.OrgID == common.UserInfo.Organization.OrgID {
		AuthenticatedUsers.Store(token, userInfo)
		return true
	} else {
		fmt.Println("Authentication failed due to different organization", token)
		return false
	}
}

func AuthenticateUser(user, key string) (common.UserDetails, error) {
	var userInfo common.UserDetails
	payload := fmt.Sprintf(`{"username": "%s", "token": "%s"}`, user, key)
	body := []byte(payload)
	_, resp, err := MakePostRequest(common.AuthenticateEndpoint, body)
	if err == nil && resp != nil {
		err = json.Unmarshal(resp, &userInfo)
	}
	return userInfo, err
}
