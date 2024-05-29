package repositories

import (
	"context"
	"fmt"

	"github.com/LambdatestIncPrivate/exemplar/pkg/lumber"
	"github.com/LambdatestIncPrivate/exemplar/pkg/models"
	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/mysql"
	"github.com/jmoiron/sqlx"
)

// CommonRepository is a standard repo that can be embedded in all repsitories
type CommonRepository interface {
	GetByID(context.Context, string, models.BaseEntity) error
	Create(context.Context, models.BaseEntity) (string, error)
}

type MySQLCommonRepository struct {
	db     *sqlx.DB
	logger lumber.Logger
}

func NewMySQLCommonRepository(db *sqlx.DB, logger lumber.Logger) *MySQLCommonRepository {
	return &MySQLCommonRepository{db: db, logger: logger}
}

//GetByID fetches entity record from db using id
func (repo *MySQLCommonRepository) GetByID(ctx context.Context, id string, model models.BaseEntity) error {
	query := fmt.Sprintf("SELECT * FROM %s WHERE id=?", model.GetTableName())
	err := repo.db.Get(model, query, id)
	if err != nil {
		return err
	}

	return nil
}

//Create creates host entry and saves in DB
func (repo *MySQLCommonRepository) Create(ctx context.Context, model models.BaseEntity) (string, error) {

	dialect := goqu.Dialect("mysql")

	// create insert statement
	ds := dialect.Insert(model.GetTableName()).Rows(
		model,
	)
	insertSQL, _, _ := ds.ToSQL()
	_, err := repo.db.Exec(insertSQL)
	if err != nil {
		return "", err
	}

	return model.GetID(), nil
}
