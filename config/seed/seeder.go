package seed

import (
	"log"

	"gorm.io/gorm"
)

func Seeder(db *gorm.DB) {
	if err := SeedUsers(db); err != nil {
		log.Fatalf("Error al seedear usuarios: %v", err)
	}
}
