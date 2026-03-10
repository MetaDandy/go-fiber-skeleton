package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID    uuid.UUID `gorm:"type:uuid;primaryKey;"`
	Name  string
	Email string `gorm:"uniqueIndex"`

	Tasks []Task `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (User) TableName() string {
	return "users"
}
