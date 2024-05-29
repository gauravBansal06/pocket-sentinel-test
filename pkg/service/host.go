package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/LambdatestIncPrivate/exemplar/pkg/lumber"
	"github.com/LambdatestIncPrivate/exemplar/pkg/models"
	"github.com/LambdatestIncPrivate/exemplar/pkg/repositories"
	"github.com/LambdatestIncPrivate/exemplar/pkg/utils"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type hostService struct {
	db     *sqlx.DB
	logger lumber.Logger
	repo   repositories.HostRepository
}

func NewHostService(logger lumber.Logger, db *sqlx.DB) HostService {
	repo := repositories.NewMySQLHostRepository(db, logger)
	return &hostService{db: db, logger: logger, repo: repo}
}

func (svc *hostService) GetByID(ctx context.Context, id string) (*models.Host, error) {
	svc.logger.Debugf("Fetching host details for id %s", id)
	var h models.Host
	// fetch host using id
	err := svc.repo.GetByID(context.Background(), id, &h)
	if err != nil {
		svc.logger.Errorf("Unable to fetch host: %s", err)
		return nil, errors.New(fmt.Sprintf("Host fetch error %s", err))
	}

	return &h, nil

}

func (svc *hostService) Create(ctx context.Context, host *models.Host) (string, string, error) {
	svc.logger.Debugf("Creating new host id: %s", "xxx-yyy-zzz")

	uuid, err := uuid.New().MarshalText()
	if err != nil {
		svc.logger.Errorf("Unable to generate new uuid: %s", err)
		return "", "", err
	}
	// set uuid
	host.ID = string(uuid)
	host.Hash = utils.RandStringBytesMaskImprSrcSB(15)
	host.Created = time.Now()
	host.Updated = time.Now()

	err = host.Validate()
	if err != nil {
		svc.logger.Debugf("Model validation error %s", err)
	}

	id, err := svc.repo.Create(context.Background(), host)
	if err != nil {
		return "", "", fmt.Errorf("Unable to create new host record using payload %s ", err)
	}
	return id, host.Hash, nil
}
