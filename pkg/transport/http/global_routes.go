package http

import (
	"net/http"

	"github.com/LambdatestIncPrivate/exemplar/pkg/endpoint"
	"github.com/LambdatestIncPrivate/exemplar/pkg/lumber"
	"github.com/LambdatestIncPrivate/exemplar/pkg/service"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

// RegisterGlobalRoutes registers global routes and logger in the package
func RegisterGlobalRoutes(router *gin.RouterGroup, db *sqlx.DB, logger lumber.Logger) {
	var (
		svcLogger = logger.WithFields(lumber.Fields{"service": "health-service"})
		service   = service.NewHealthService(svcLogger)
		endpoints = endpoint.NewHealthEndpoints(service, svcLogger)
	)

	router.GET("/health", health)
	router.GET("/status", status)
	router.GET("/custom-health", NewHTTPHandler(endpoints.GetHealth, NopRequestDecoder, EncodeJSONResponse, logger))
	logger.Infof("global routes injected")
}

// health API
// @Summary Health API
// @Description checks if service is healthy
// @Tags health

// @Success 200 {object} string
// @Router /health [get] {}
func health(c *gin.Context) {
	c.Data(http.StatusOK, gin.MIMEPlain, []byte("OK"))
}

// status api
// @Summary Status API
// @Description Sample JSON api
// @Tags sample

// @Success 200 {object} string
// @Router /status [get]
func status(c *gin.Context) {
	c.Data(http.StatusOK, gin.MIMEJSON, []byte("{\"status\": \"OK\"}"))
}
