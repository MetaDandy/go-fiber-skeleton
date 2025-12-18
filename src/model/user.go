package model

import "github.com/google/uuid"

type User struct {
	ID    uuid.UUID `gorm:"type:uuid;primaryKey;"`
	Name  string
	Email string `gorm:"uniqueIndex"`
}

func (User) TableName() string {
	return "users"
}
