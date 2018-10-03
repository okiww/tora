package model

import (
	"time"

	"github.com/jinzhu/gorm"
	"github.com/satori/go.uuid"
)

// BaseModel base model definition for common entity's field
type BaseModel struct {
	ID        uuid.UUID  `gorm:"type:char(36); primary_key"`
	CreatedAt time.Time  `json:"-"`
	UpdatedAt time.Time  `json:"-"`
	DeletedAt *time.Time `sql:"index" json:"-"`
}

// BeforeCreate gorm callback
func (m *BaseModel) BeforeCreate(scope *gorm.Scope) error {
	if m.ID == uuid.Nil {
		newID := uuid.NewV4()
		return scope.SetColumn("ID", newID)
	}

	return nil
}
