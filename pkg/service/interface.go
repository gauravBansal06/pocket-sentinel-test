package service

import (
	"context"

	"github.com/LambdatestIncPrivate/exemplar/pkg/models"
)

// request and response params
type GetByIDRequest struct {
	ID string
}

type GetByIDResponse struct {
	Host *models.Host
}

type CreateRequest struct {
	Host *models.Host
}

type CreateResponse struct {
	UUID string
	Hash string
}

// HostService
type HostService interface {
	Create(ctx context.Context, host *models.Host) (string, string, error)
	GetByID(ctx context.Context, id string) (*models.Host, error)
}

// HealthService
type HealthService interface {
	Health(ctx context.Context) error
}
