package repositories

import (
	"github.com/LambdatestIncPrivate/exemplar/pkg/lumber"
	_ "github.com/doug-martin/goqu/v9/dialect/mysql"
	"github.com/jmoiron/sqlx"
)

// HostRepository as of now doesn't support custom methods and promotes methods from common repo
type HostRepository interface {
	CommonRepository
}

type MySQLHostRepository struct {
	*MySQLCommonRepository
	db     *sqlx.DB
	logger lumber.Logger
}

func NewMySQLHostRepository(db *sqlx.DB, logger lumber.Logger) *MySQLHostRepository {
	return &MySQLHostRepository{db: db, logger: logger, MySQLCommonRepository: NewMySQLCommonRepository(db, logger)}
}
