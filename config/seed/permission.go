package seed

import (
	"log"

	"github.com/MetaDandy/go-fiber-skeleton/src/enum"
	"github.com/MetaDandy/go-fiber-skeleton/src/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var permissionModules = []model.Permission{
	// Existing (5)
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
	{
		ID:          enum.RoleCreate.String(),
		Name:        "Permiso para crear un rol",
		Description: "Permiso para crear un rol específico",
	},
	{
		ID:          enum.UserPermissionCreate.String(),
		Name:        "Permiso para asignar permiso directo a usuario",
		Description: "Permiso para crear user_permission",
	},
	{
		ID:          enum.UserPermissionRead.String(),
		Name:        "Permiso para leer permisos de usuario",
		Description: "Permiso para leer user_permission",
	},
	// New (14)
	{
		ID:          enum.PermissionCreate.String(),
		Name:        "Permiso para crear un permiso",
		Description: "Permiso para crear un recurso específico",
	},
	{
		ID:          enum.UserCreate.String(),
		Name:        "Permiso para crear un usuario",
		Description: "Permiso para crear un usuario",
	},
	{
		ID:          enum.UserUpdate.String(),
		Name:        "Permiso para actualizar un usuario",
		Description: "Permiso para actualizar un usuario",
	},
	{
		ID:          enum.UserDelete.String(),
		Name:        "Permiso para eliminar un usuario",
		Description: "Permiso para eliminar un usuario",
	},
	{
		ID:          enum.UserList.String(),
		Name:        "Permiso para listar usuarios",
		Description: "Permiso para listar usuarios",
	},
	{
		ID:          enum.RoleUpdate.String(),
		Name:        "Permiso para actualizar un rol",
		Description: "Permiso para actualizar un rol",
	},
	{
		ID:          enum.RoleList.String(),
		Name:        "Permiso para listar roles",
		Description: "Permiso para listar roles",
	},
	{
		ID:          enum.SessionList.String(),
		Name:        "Permiso para listar sesiones",
		Description: "Permiso para listar sesiones de usuario",
	},
	{
		ID:          enum.SessionRevoke.String(),
		Name:        "Permiso para revocar una sesión",
		Description: "Permiso para revocar una sesión específica",
	},
	{
		ID:          enum.SessionRevokeAll.String(),
		Name:        "Permiso para revocar todas las sesiones",
		Description: "Permiso para revocar todas las sesiones de un usuario",
	},
	{
		ID:          enum.AuthLogRead.String(),
		Name:        "Permiso para leer un log de autenticación",
		Description: "Permiso para leer un log de autenticación",
	},
	{
		ID:          enum.AuthLogList.String(),
		Name:        "Permiso para listar logs de autenticación",
		Description: "Permiso para listar logs de autenticación",
	},
	{
		ID:          enum.UserPermissionDelete.String(),
		Name:        "Permiso para eliminar un permiso de usuario",
		Description: "Permiso para eliminar user_permission",
	},
	{
		ID:          enum.UserPermissionList.String(),
		Name:        "Permiso para listar permisos de usuario",
		Description: "Permiso para listar user_permissions",
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
