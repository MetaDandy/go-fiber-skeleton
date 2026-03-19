package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RolePermission struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey;"`

	RoleID uuid.UUID `gorm:"type:uuid;"`
	Role   Role      `gorm:"foreignKey:RoleID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`

	PermissionID uuid.UUID  `gorm:"type:uuid;"`
	Permission   Permission `gorm:"foreignKey:PermissionID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (RolePermission) TableName() string {
	return "rolepermissions"
}
