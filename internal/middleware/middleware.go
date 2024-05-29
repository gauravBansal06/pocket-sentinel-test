package middleware

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/LambdatestIncPrivate/exemplar/pkg/global"
	"github.com/LambdatestIncPrivate/exemplar/pkg/lumber"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

var logger lumber.Logger

//RegisterLogger Initialize logger for the package
func RegisterLogger(jack lumber.Logger) {
	logger = jack
}

// Authentication middleware
func Authentication() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenSlice := strings.Split(c.Request.Header.Get("Authorization"), " ")

		if len(tokenSlice) <= 1 {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "Unauthorized"})
			return
		}
		if scheme := strings.ToLower(tokenSlice[0]); scheme != "basic" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Only Support http basic authentication"})
			return
		}
		decoded, err := base64.StdEncoding.DecodeString(tokenSlice[1])
		if err != nil {
			logger.Errorf("Base64 Decode error:", "error", err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Unable to authenticate user, please try again"})
			return
		}
		details := strings.Split(string(decoded), ":")

		if len(details) != 2 {
			logger.Errorf("bad auth credentials")
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Unable to authenticate user, please try again"})
			return
		}

		c.Set("stamped", "yes")
		logger.Debugf("request stamping done")
		c.Next()
	}

}

//JWTDecode authenticate user using jwt token
func JWTDecode() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenSlice := strings.Split(c.Request.Header.Get("Authorization"), " ")

		if len(tokenSlice) <= 1 {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "Authorization token missing. Unable to get tunnel details"})
			return
		}
		if scheme := strings.ToLower(tokenSlice[0]); scheme != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Only Support bearer token"})
			return
		}

		parsedToken, _ := jwt.Parse(tokenSlice[1], func(token *jwt.Token) (interface{}, error) {
			return []byte(""), nil
		})

		claims := parsedToken.Claims.(jwt.MapClaims)
		data, err := json.Marshal(claims)
		if err != nil {
			logger.Errorf("Failed to marshall jwt claim", "error", err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Invalid token"})
			return
		}

		logger.Debugf("JWT claimed data %s", data)

		c.Set("stamped", "yes")
		c.Next()

	}
}

//WsValidator middleware to validate all request headers before initiating a websocket handshake
func WsValidator() gin.HandlerFunc {
	return func(c *gin.Context) {

		if strings.ToLower(c.Request.Header.Get("Connection")) != "upgrade" {
			c.Data(http.StatusBadRequest, gin.MIMEPlain, []byte(global.BadHandshake+"'upgrade' token not found in 'Connection' header"))
			c.Abort()
			return
		}

		if strings.ToLower(c.Request.Header.Get("Upgrade")) != "websocket" {
			c.Data(http.StatusBadRequest, gin.MIMEPlain, []byte(global.BadHandshake+"'websocket' token not found in 'Upgrade' header"))
			c.Abort()
			return
		}
		if c.Request.Method != "GET" {
			c.Data(http.StatusMethodNotAllowed, gin.MIMEPlain, []byte(global.BadHandshake+"request method is not GET"))
			c.Abort()
			return
		}
		if c.Request.Header.Get("Sec-Websocket-Version") != "13" {
			c.Data(http.StatusMethodNotAllowed, gin.MIMEPlain, []byte("websocket: unsupported version: 13 not found in 'Sec-Websocket-Version' header"))
			c.Abort()
			return

		}
		if c.Request.Header.Get("Sec-Websocket-Key") == "" {
			c.Data(http.StatusBadRequest, gin.MIMEPlain, []byte("websocket: not a websocket handshake: 'Sec-WebSocket-Key' header is missing or blank"))
			c.Abort()
			return
		}
		c.Next()
	}
}
