package cron

import (
	"context"
	"sync"

	"github.com/LambdatestIncPrivate/exemplar/config"
	"github.com/LambdatestIncPrivate/exemplar/pkg/lumber"
	"github.com/robfig/cron/v3"
)

var logger lumber.Logger

// Setup initializes all crons on service startup
func Setup(config *config.Config, ctx context.Context, wg *sync.WaitGroup, initializedLogger lumber.Logger) {
	defer wg.Done()

	logger = initializedLogger
	logger.Infof("Setting up crons")
	c := cron.New()
	c.AddFunc("* * * * *", func() { logger.Infof("Cron alert ...") })
	c.Start()

	select {
	case <-ctx.Done():
		c.Stop()
		logger.Infof("Caller has requested graceful shutdown. Returning.....")
		return
	}

}
