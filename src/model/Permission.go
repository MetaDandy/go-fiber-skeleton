package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Permission struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;"`
	Description string
	Name        string

	UserPermissions          []UserPermission          `gorm:"foreignKey:PermissionID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	RolePermissions          []RoleEffectivePermission `gorm:"foreignKey:PermissionID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	RoleEffectivePermissions []RoleEffectivePermission `gorm:"foreignKey:PermissionID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (Permission) TableName() string {
	return "permissions"
}
