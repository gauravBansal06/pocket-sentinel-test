package services

import (
	"byod/common"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// AuthenticatedUsers stores a map of authenticated users to avoid reauthentication overhead.
var AuthenticatedUsers sync.Map
var AuthenticatedJwtUsers sync.Map

// credentials stores the username and password extracted from the token.
type credentials struct {
	username string
	password string
}

// IsValidUser checks if the user associated with the given token is valid.
func IsValidUser(authToken string) (common.UserDetails, bool) {
	tokenSlice := strings.Split(authToken, " ")
	if len(tokenSlice) <= 1 {
		log.Println("invalid token")
		return common.UserDetails{}, false
	}

	var userInfo common.UserDetails
	var err error

	scheme := strings.ToLower(tokenSlice[0])
	if scheme == "basic" {
		log.Println("Proceeding with basic token auth")
		userInfo, err = BasicAuthentication(tokenSlice[1])
	} else if scheme == "bearer" {
		log.Println("Proceeding with JWT auth")
		userInfo, err = JWTAuthentication(authToken)
	} else {
		log.Println("got unsupported auth scheme: ", scheme)
		return common.UserDetails{}, false
	}

	if err != nil {
		log.Printf("%v auth error: %v\n", scheme, err)
		return common.UserDetails{}, false
	}
	return userInfo, true
}

// auth using basic token
func BasicAuthentication(token string) (common.UserDetails, error) {
	// Parse the token to retrieve credentials
	bytes, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return common.UserDetails{}, fmt.Errorf("invalid token: %v", err)
	}
	parts := strings.Split(string(bytes), ":")
	if len(parts) != 2 {
		return common.UserDetails{}, fmt.Errorf("token does not contain expected format")
	}
	creds := &credentials{username: parts[0], password: parts[1]}

	// Check if the user is already authenticated by looking up the credentials in the map
	if userDetails, ok := AuthenticatedUsers.Load(creds); ok {
		userInfo, _ := userDetails.(common.UserDetails)
		return userInfo, nil
	}

	// Authenticate the user with the credentials extracted from the token
	userInfo, err := AuthenticateUser(creds.username, creds.password)
	if err != nil {
		return common.UserDetails{}, fmt.Errorf("authentication failed for user: %s", creds.username)
	}

	// Check if the authenticated user is part of the same organization as the current user
	if userInfo.Organization.OrgID == common.UserInfo.Organization.OrgID {
		AuthenticatedUsers.Store(creds, userInfo)
		return userInfo, nil
	}

	return common.UserDetails{}, fmt.Errorf("authentication failed due to different organization for user: %+v", userInfo)
}

// AuthenticateUser sends a POST request to authenticate a user with a given username and password.
func AuthenticateUser(username, password string) (common.UserDetails, error) {
	payload := fmt.Sprintf(`{"username": "%s", "token": "%s"}`, username, password)
	body := []byte(payload)

	// Make a POST request to the authentication endpoint with the user credentials
	_, resp, err := MakePostRequest(common.BasicAuthenticateEndpoint, body)
	if err != nil {
		return common.UserDetails{}, err
	}

	var userInfo common.UserDetails
	// Parse the JSON response into the userInfo struct
	if err := json.Unmarshal(resp, &userInfo); err != nil {
		return common.UserDetails{}, err
	}
	userInfo.OrgID = userInfo.Organization.OrgID
	return userInfo, nil
}

// auth using jwt token
func JWTAuthentication(bearerToken string) (common.UserDetails, error) {
	token := strings.Split(bearerToken, " ")[1]

	// Check if the user is already authenticated by looking up the token in the map
	if userDetails, ok := AuthenticatedJwtUsers.Load(token); ok {
		if !IsJWTExpired(token) {
			userInfo, _ := userDetails.(common.UserDetails)
			return userInfo, nil
		} else {
			AuthenticatedJwtUsers.Delete(token)
		}
	}

	// Authenticate the user from LUMS using bearer token
	userInfo, err := BearerAuthenticateUser(bearerToken)
	if err != nil {
		return common.UserDetails{}, fmt.Errorf("authentication failed for user: %v", token)
	}

	// Check if the authenticated user is part of the same organization as the current user
	if userInfo.Organization.OrgID == common.UserInfo.Organization.OrgID {
		AuthenticatedJwtUsers.Store(token, userInfo)
		return userInfo, nil
	}

	return common.UserDetails{}, fmt.Errorf("authentication failed due to different organization for user: %+v", userInfo)
}

// Authentication using jwt bearer token
func BearerAuthenticateUser(token string) (common.UserDetails, error) {
	// Make a POST request to the authentication endpoint with the user credentials
	_, resp, err := MakeGetRequest(common.BearerAuthenticateEndpoint, token, nil)
	if err != nil {
		return common.UserDetails{}, err
	}

	var userDetails common.BearerAuthUserDetails
	// Parse the JSON response into the userInfo struct
	if err := json.Unmarshal(resp, &userDetails); err != nil {
		return common.UserDetails{}, err
	}

	var userInfo = common.UserDetails{
		UserID:   userDetails.UserID,
		Name:     userDetails.Name,
		Email:    userDetails.Username,
		Username: userDetails.Username,
		Status:   userDetails.Status,
		Role:     userDetails.Role,
		OrgID:    userDetails.OrgID,
		ApiToken: userDetails.Token,
		Organization: common.Organization{
			OrgID: userDetails.OrgID,
		},
	}
	return userInfo, nil
}

// check jwt token is not expired
func IsJWTExpired(token string) bool {
	// Parse the JWT token without validating the signature as signature already verified from LUMS
	jwtToken, _, err := new(jwt.Parser).ParseUnverified(token, jwt.MapClaims{})
	if err != nil {
		log.Println("Failed to parse token: ", err)
		return true
	}

	// Get the claims from the token
	if claims, ok := jwtToken.Claims.(jwt.MapClaims); ok {
		// Check the "exp" claim
		if exp, ok := claims["exp"].(float64); ok {
			expirationTime := time.Unix(int64(exp), 0)
			if time.Now().After(expirationTime) {
				return true
			} else {
				return false
			}
		} else {
			return false
		}
	} else {
		return true //mark expired in case no claims found
	}
}

// a watcher to reset AuthenticatedJwtUsers
func ResetAuthenticatedJwtUsersCron(stopChan chan struct{}) {
	log.Println("starting ResetAuthenticatedJwtUsersCron.....")
	for {
		select {
		case <-stopChan:
			log.Println("received termination signal: stopping ResetAuthenticatedJwtUsersCron")
			return
		default:
			AuthenticatedJwtUsers.Range(func(key, value interface{}) bool {
				AuthenticatedJwtUsers.Delete(key)
				return true
			})
			time.Sleep(30 * time.Minute)
		}
	}
}
