package seed

import (
	//"log"

	"gorm.io/gorm"
)

func Seeder(db *gorm.DB) {

	SeedRoles(db)
	SeedPermissions(db)

	/*if err := Users(db); err != nil {
		log.Fatalf("Error al seedear usuarios: %v", err)
	}*/

}
