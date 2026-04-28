package model

import (
	"time"

	"gorm.io/gorm"
)

type Permission struct {
	ID          string `gorm:"primaryKey;type:varchar(255)"`
	Description string
	Name        string

	UserPermissions          []UserPermission          `gorm:"foreignKey:PermissionID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	RolePermissions          []RolePermission          `gorm:"foreignKey:PermissionID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	RoleEffectivePermissions []RoleEffectivePermission `gorm:"foreignKey:PermissionID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (Permission) TableName() string {
	return "permissions"
}
