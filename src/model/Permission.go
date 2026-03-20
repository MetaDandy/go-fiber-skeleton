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
	RolePermissions          []RoleEffectivePermission `gorm:"foreignKey:PermissionID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	RoleEffectivePermissions []RoleEffectivePermission `gorm:"foreignKey:PermissionID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (Permission) TableName() string {
	return "Permissions"
}
