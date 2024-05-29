package endpoint

import (
	"context"

	"github.com/LambdatestIncPrivate/exemplar/pkg/lumber"
	"github.com/LambdatestIncPrivate/exemplar/pkg/service"
	"github.com/go-kit/kit/endpoint"
	kep "github.com/go-kit/kit/endpoint"
)

type HealthEndpoints struct {
	GetHealth kep.Endpoint
}

func NewHealthEndpoints(svc service.HealthService, logger lumber.Logger) HealthEndpoints {
	return HealthEndpoints{
		GetHealth: makeGetHealthEndpoint(svc, logger),
	}
}

func makeGetHealthEndpoint(s service.HealthService, logger lumber.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		err := s.Health(ctx)
		if err != nil {
			logger.Errorf("Service is not healthy: %s", err)
			return nil, err
		}
		logger.Infof("Service is healthy")
		return nil, nil
	}
}
