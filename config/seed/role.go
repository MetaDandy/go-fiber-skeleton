package seed

import (
	"log"

	"github.com/MetaDandy/go-fiber-skeleton/constant"
	"github.com/MetaDandy/go-fiber-skeleton/src/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var roles = []model.Role{
	{
		ID:          constant.GenericID,
		Name:        "Generic User",
		Description: "Rol genérico para usuarios registrados",
	},
}

func verifyAndCreateRoles(db *gorm.DB, rolesToCreate []model.Role) ([]model.Role, error) {
	if len(rolesToCreate) == 0 {
		return []model.Role{}, nil
	}

	result := db.
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			DoNothing: true,
		}).
		CreateInBatches(rolesToCreate, 50)

	if result.Error != nil {
		log.Printf("error creando roles: %v", result.Error)
		return nil, result.Error
	}

	log.Printf("%d roles nuevos fueron insertados", result.RowsAffected)
	log.Printf("%d roles ya existían", len(rolesToCreate)-int(result.RowsAffected))

	return rolesToCreate, nil
}

func SeedRoles(db *gorm.DB) {
	createdRoles, err := verifyAndCreateRoles(db, roles)
	if err != nil {
		log.Printf("error en el proceso de creación de roles: %v", err)
		return
	}

	if len(createdRoles) > 0 {
		log.Printf("proceso de seed de roles ejecutado correctamente")
	} else {
		log.Println("no había roles para procesar")
	}
}
