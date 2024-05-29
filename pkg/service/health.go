package service

import (
	"context"

	"github.com/LambdatestIncPrivate/exemplar/pkg/errs"
	"github.com/LambdatestIncPrivate/exemplar/pkg/lumber"
)

type healthService struct {
	logger lumber.Logger
}

// NewHealthService generates new health service
func NewHealthService(logger lumber.Logger) HealthService {
	return &healthService{logger: logger}
}

func (svc *healthService) Health(ctx context.Context) error {
	return errs.ERR_DUMMY
}
