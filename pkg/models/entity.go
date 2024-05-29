package models

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// BaseEntity defines entities that can be used with Common repository
type BaseEntity interface {
	GetID() string
	GetTableName() string
}

// CommonModel defines common attributes to shared by all entities
type CommonModel struct {
	ID      string    `json:"id"`
	Created time.Time `json:"created_time" time_format:"2006-01-02 15:04:05"`
	Updated time.Time `json:"updated_time" time_format:"2006-01-02 15:04:05"`
}

// GetID returns string representation of ID
func (m *CommonModel) GetID() string {
	if m == nil {
		return ""
	}
	return m.ID
}

// Validate checks if ID is a valid UUID or not
func (m *CommonModel) Validate() error {

	// return error if id is not a valid uuid
	if _, err := uuid.Parse(m.ID); err != nil {
		return errors.New("Invalid ID. ID must a valid UUID")
	}

	return nil
}
