package http

import (
	"context"
	"net/http"

	"github.com/LambdatestIncPrivate/exemplar/internal/middleware"
	"github.com/LambdatestIncPrivate/exemplar/pkg/endpoint"
	"github.com/LambdatestIncPrivate/exemplar/pkg/lumber"
	"github.com/LambdatestIncPrivate/exemplar/pkg/service"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

var logger lumber.Logger

// RegisterRoutes registers routes and logger in the package
func RegisterV1Routes(router *gin.RouterGroup, db *sqlx.DB, logger lumber.Logger) {
	svcLogger := logger.WithFields(lumber.Fields{"service": "model-service"})
	var (
		service   = service.NewHostService(svcLogger, db)
		endpoints = endpoint.NewHostEndpoints(service, logger)
	)
	router.POST("/hello", middleware.Authentication(), versionedHello)

	// host API
	router.GET("/host/:id", NewHTTPHandler(endpoints.GetHost, DecodeHostGetRequestfunc, EncodeHostGetRequestfunc, logger))
	router.GET("/host/:id/default-encoder-example", NewHTTPHandler(endpoints.GetHost, DecodeHostGetRequestfunc, EncodeJSONResponse, logger))
	router.Static("/docs", "pkg/swagger-ui")
	logger.Infof("v1.0 routes injected")
}

func versionedHello(c *gin.Context) {
	c.Data(http.StatusOK, gin.MIMEJSON, []byte("{\"status\": \"VALID\"}"))
}

// @Summary Get host by id
// @Description API to fetch host information using ID
// @id get-host-by-id
// @Produce  json
// @Param id path string true "UUID for host object	"
// @Success 200 {object} models.Host
// @Failure 500 {object} errs.Err
// @Router /v1.0/host/{id} [get]
func DecodeHostGetRequestfunc(c context.Context, g *gin.Context, logger lumber.Logger) (request interface{}, err error) {
	id := g.Param("id")
	return service.GetByIDRequest{ID: id}, nil
}

func EncodeHostGetRequestfunc(c context.Context, g *gin.Context, response interface{}, logger lumber.Logger) (err error) {
	getByIDResponse, ok := response.(service.GetByIDResponse)
	if !ok {
		logger.Errorf("Unable to cast response to getById struct: %s", err)
		return err
	}
	g.JSON(200, getByIDResponse.Host)
	return nil
}
