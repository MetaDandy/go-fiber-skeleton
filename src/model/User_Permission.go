package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserPermission struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey;"`

	PermissionID uuid.UUID  `gorm:"type:uuid;"`
	Permission   Permission `gorm:"foreignKey:PermissionID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`

	UserID uuid.UUID `gorm:"type:uuid;"`
	User   User      `gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (UserPermission) TableName() string {
	return "UserPermissions"
}
