package seed

import (
	"log"

	"github.com/MetaDandy/go-fiber-skeleton/src/enum"
	"github.com/MetaDandy/go-fiber-skeleton/src/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var permissionModules = []model.Permission{
	{
		ID:          enum.PermissionRead.String(),
		Name:        "Permiso para leer un permiso",
		Description: "Permiso para leer un recurso específico",
	},
	{
		ID:          enum.PermissionList.String(),
		Name:        "Permiso para listar permisos",
		Description: "Permiso para listar recursos específicos",
	},
}

func verifyAndCreatePermissions(db *gorm.DB, permissions []model.Permission) ([]model.Permission, error) {
	if len(permissions) == 0 {
		return []model.Permission{}, nil
	}

	result := db.
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			DoNothing: true,
		}).
		CreateInBatches(permissions, 50)

	if result.Error != nil {
		log.Printf("error creando permisos: %v", result.Error)
		return nil, result.Error
	}

	log.Printf("%d permisos nuevos fueron insertados", result.RowsAffected)
	log.Printf("%d permisos ya existían", len(permissions)-int(result.RowsAffected))

	return permissions, nil
}

func SeedPermissions(db *gorm.DB) {
	createdPermissions, err := verifyAndCreatePermissions(db, permissionModules)
	if err != nil {
		log.Printf("error en el proceso de creación de permisos: %v", err)
		return
	}

	if len(createdPermissions) > 0 {
		log.Printf("proceso de seed ejecutado correctamente")
	} else {
		log.Println("no había permisos para procesar")
	}
}
