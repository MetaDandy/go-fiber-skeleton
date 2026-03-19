package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey;"`
	Name            string
	Email           string `gorm:"uniqueIndex"`
	EmailVerified   bool
	EmailVerifiedAt time.Time
	Phone           string
	Password        string
	Picture         string

	RoleID uuid.UUID `gorm:"type:uuid;"`
	Role   Role      `gorm:"foreignKey:RoleID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`

	AuthProviders           []AuthProvider           `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	AuthLogs                []AuthLog                `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	UserPermissions         []UserPermission         `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	Sessions                []Session                `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	EmailVerificationTokens []EmailVerificationToken `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	PasswordResetTokens     []PasswordResetToken     `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (User) TableName() string {
	return "users"
}
