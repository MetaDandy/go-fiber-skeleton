package model

import (
	"time"

	"github.com/google/uuid"
)

type OAuthState struct {
	ID        uuid.UUID  `gorm:"primaryKey" json:"id"`
	State     string     `gorm:"uniqueIndex;not null" json:"state"`
	Provider  string     `gorm:"not null" json:"provider"`
	CreatedAt time.Time  `json:"created_at"`
	ExpiresAt time.Time  `gorm:"not null" json:"expires_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

func (OAuthState) TableName() string {
	return "oauth_states"
}
