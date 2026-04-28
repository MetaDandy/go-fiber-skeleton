package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Session struct {
	ID               uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Provider         string
	RefreshTokenHash string
	ExpiresAt        time.Time `gorm:"type:timestamptz"`
	Ip               string
	UserAgent        string
	RevokedAt        *time.Time `gorm:"type:timestamptz"`

	UserID uuid.UUID `gorm:"type:uuid"`
	User   User      `gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`

	CreatedAt time.Time      `gorm:"type:timestamptz;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time      `gorm:"type:timestamptz;default:CURRENT_TIMESTAMP"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (Session) TableName() string {
	return "sessions"
}
