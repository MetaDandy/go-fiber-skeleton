package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RoleEffectivePermission struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey;"`

	RoleID uuid.UUID `gorm:"type:uuid;uniqueIndex:uq_role_eff_perm;not null"`
	Role   Role      `gorm:"foreignKey:RoleID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	SourceRoleID uuid.UUID `gorm:"type:uuid;index;not null"`
	SourceRole   Role      `gorm:"foreignKey:SourceRoleID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	PermissionID string     `gorm:"type:varchar(255);uniqueIndex:uq_role_eff_perm;not null"`
	Permission   Permission `gorm:"foreignKey:PermissionID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (RoleEffectivePermission) TableName() string {
	return "roleeffectivepermissions"
}
