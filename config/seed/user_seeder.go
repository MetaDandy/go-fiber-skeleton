package seed

import (
	"github.com/MetaDandy/go-fiber-skeleton/src/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func SeedUsers(db *gorm.DB) error {
	users := []model.User{
		{
			ID:    uuid.New(),
			Name:  "Juan García",
			Email: "juan.garcia@example.com",
		},
		{
			ID:    uuid.New(),
			Name:  "María López",
			Email: "maria.lopez@example.com",
		},
		{
			ID:    uuid.New(),
			Name:  "Carlos Rodríguez",
			Email: "carlos.rodriguez@example.com",
		},
		{
			ID:    uuid.New(),
			Name:  "Ana Martínez",
			Email: "ana.martinez@example.com",
		},
		{
			ID:    uuid.New(),
			Name:  "Pedro Sánchez",
			Email: "pedro.sanchez@example.com",
		},
	}

	for _, user := range users {
		// Verificar si el usuario ya existe por email
		var existingUser model.User
		if err := db.Where("email = ?", user.Email).First(&existingUser).Error; err == nil {
			// Usuario ya existe, saltar
			continue
		}

		// Insertar solo si no existe
		if err := db.Create(&user).Error; err != nil {
			return err
		}
	}

	return nil
}
