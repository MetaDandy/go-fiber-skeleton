package config

import (
	"log"

	"github.com/MetaDandy/go-fiber-skeleton/src/model"
	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) {
	err := db.AutoMigrate(
		&model.User{},
		&model.Task{},
	)

	if err != nil {
		log.Fatal("Failed to migrate database", err)
	}
}
