package http

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/LambdatestIncPrivate/exemplar/internal/middleware"
	"github.com/LambdatestIncPrivate/exemplar/pkg/errs"
	"github.com/LambdatestIncPrivate/exemplar/pkg/lumber"
	"github.com/jmoiron/sqlx"
	"github.com/lestrrat-go/backoff"

	"github.com/gin-contrib/cors"

	"github.com/LambdatestIncPrivate/exemplar/config"
	"github.com/LambdatestIncPrivate/exemplar/internal/ws"
	globalConfig "github.com/LambdatestIncPrivate/exemplar/pkg/global"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	_ "github.com/swaggo/gin-swagger/example/basic/docs"
)

// Setup initializes all crons on service startup

// @title Exemplar API
// @version 1.0
// @description Exemplar is a model microservice with best practices
// @host localhost:9876
// @BasePath /v1.0
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
func StartAPIServer(config *config.Config, ctx context.Context, wg *sync.WaitGroup, db *sqlx.DB, logger lumber.Logger) error {
	defer wg.Done()

	// set gin to release mode
	gin.SetMode(gin.ReleaseMode)

	//Initialize logger for packages

	middleware.RegisterLogger(logger)

	logger.Infof("Setting up http handler")
	router := gin.Default()

	// Recovery middleware recovers from any panics and writes a 500 if there was one.
	router.Use(gin.Recovery())

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AddAllowHeaders("authorization", "cache-control", "pragma")
	router.Use(cors.New(corsConfig))

	errChan := make(chan error)

	globalRoute := router.Group("/")
	v1Route := router.Group("/v1.0")

	// registerGlobalRoutes(globalRoute, logger)
	RegisterGlobalRoutes(globalRoute, db, logger)
	RegisterV1Routes(v1Route, db, logger)
	ws.RegisterRoutes(globalRoute, logger)

	// attach swagger endpoint
	url := ginSwagger.URL("http://localhost:9876/swagger/doc.json")
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, url))

	// HTTP server instance
	srv := &http.Server{
		Addr:    ":" + config.Port,
		Handler: router,
	}

	// channel to signal server process exit
	done := make(chan struct{})
	go func() {
		logger.Infof("Starting server on port %s", config.Port)
		// service connections
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Errorf("listen: %#v", err)
			errChan <- err
		}
	}()

	select {
	case <-ctx.Done():
		logger.Infof("Caller has requested graceful shutdown. shutting down the server")
		if err := srv.Shutdown(ctx); err != nil {
			logger.Errorf("Server Shutdown:", "error", err)
		}
		return nil
	case err := <-errChan:
		return err
	case <-done:
		return nil
	}

}

func Setup(config *config.Config, ctx context.Context, wg *sync.WaitGroup, db *sqlx.DB, logger lumber.Logger) error {
	logger.Debugf("Starting API server")

	var policy = backoff.NewExponential(
		backoff.WithInterval(250*time.Millisecond),                         // base interval
		backoff.WithJitterFactor(0.05),                                     // 5% jitter
		backoff.WithMaxRetries(globalConfig.MAX_API_SERVER_START_ATTEMPTS), // If not specified, default number of retries is 10
	)

	b, cancel := policy.Start(context.Background())
	defer cancel()

	for backoff.Continue(b) {
		// channel to mark completion
		done := make(chan struct{})

		// check for context
		select {
		case <-ctx.Done():
			logger.Debugf("Context cancelled. Stopping sink proxy")
			select {
			case <-done:
			case <-time.After(500 * time.Microsecond):
			}
			return nil
		default:
			err := StartAPIServer(config, ctx, wg, db, logger)
			if err != nil {
				logger.Errorf("API server start error: %s", err)
				logger.Warnf("Restarting api server")
			} else {
				return nil
			}

		}
	}

	// in the case of retry attempt exceed, return with errror
	return errs.ERR_INF_API_MAX_ATTEMPT
}
