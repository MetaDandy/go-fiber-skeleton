package model

import (
	"time"

	"github.com/MetaDandy/go-fiber-skeleton/src/enum"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuthLog struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;"`
	Event     enum.Event
	UserAgent string
	Ip        string

	UserID uuid.UUID `gorm:"type:uuid;"`
	User   User      `gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (AuthLog) TableName() string {
	return "AuthLogs"
}
