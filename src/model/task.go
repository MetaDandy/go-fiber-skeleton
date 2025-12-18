package model

import (
	"github.com/MetaDandy/go-fiber-skeleton/src/enum"
	"github.com/google/uuid"
)

type Task struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;"`
	Title       string
	Description string
	Status      enum.StatusEnum

	UserID uuid.UUID `gorm:"type:uuid;"`
	User   User      `gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}

func (Task) TableName() string {
	return "task"
}
