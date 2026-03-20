package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuthProvider struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey;"`
	Provider       string
	ProviderUserID string

	UserID uuid.UUID `gorm:"type:uuid;"`
	User   User      `gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (AuthProvider) TableName() string {
	return "authproviders"
}
