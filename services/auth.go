package services

import (
	"byod/common"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
)

// AuthenticatedUsers stores a map of authenticated users to avoid reauthentication overhead.
var AuthenticatedUsers sync.Map

// credentials stores the username and password extracted from the token.
type credentials struct {
	username string
	password string
}

// IsValidUser checks if the user associated with the given token is valid.
func IsValidUser(token string) bool {
	// Parse the token to retrieve credentials
	creds, err := parseToken(token)
	if err != nil {
		log.Printf("Invalid token format: %s\n", err)
		return false
	}

	// Check if the user is already authenticated by looking up the credentials in the map
	if _, ok := AuthenticatedUsers.Load(creds); ok {
		return true
	}

	// Authenticate the user with the credentials extracted from the token
	userInfo, err := AuthenticateUser(creds.username, creds.password)
	if err != nil {
		log.Printf("Authentication failed for user: %s\n", creds.username)
		return false
	}

	// Check if the authenticated user is part of the same organization as the current user
	if userInfo.Organization.OrgID == common.UserInfo.Organization.OrgID {
		AuthenticatedUsers.Store(creds, userInfo)
		return true
	}

	log.Printf("Authentication failed due to different organization for user: %s\n", creds.username)
	return false
}

// parseToken decodes a Base64 encoded token to extract credentials.
func parseToken(token string) (*credentials, error) {
	parts := strings.Split(token, " ")
	if len(parts) != 2 {
		return nil, fmt.Errorf("token is not valid base64")
	}
	bytes, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, err
	}

	parts = strings.Split(string(bytes), ":")
	if len(parts) != 2 {
		return nil, fmt.Errorf("token does not contain expected format")
	}

	return &credentials{username: parts[0], password: parts[1]}, nil
}

// AuthenticateUser sends a POST request to authenticate a user with a given username and password.
func AuthenticateUser(username, password string) (common.UserDetails, error) {
	payload := fmt.Sprintf(`{"username": "%s", "token": "%s"}`, username, password)
	body := []byte(payload)

	// Make a POST request to the authentication endpoint with the user credentials
	_, resp, err := MakePostRequest(common.AuthenticateEndpoint, body)
	if err != nil {
		return common.UserDetails{}, err
	}

	var userInfo common.UserDetails
	// Parse the JSON response into the userInfo struct
	if err := json.Unmarshal(resp, &userInfo); err != nil {
		return common.UserDetails{}, err
	}
	return userInfo, nil
}
