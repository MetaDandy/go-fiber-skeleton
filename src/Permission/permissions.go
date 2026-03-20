package permission

import "strings"

type Permission struct {
	ID          string
	Name        string
	Description string
}

type ModuleActions map[string]Permission

type PermissionCatalogue map[string]ModuleActions

// BuildPermissionID crea IDs estandarizados module.action
func BuildPermissionID(module, action string) string {
	return strings.ToLower(module) + "." + strings.ToLower(action)
}

// BuildPermission crea un permiso normalizado para una acción.
func BuildPermission(module, action string) Permission {
	id := BuildPermissionID(module, action)
	return Permission{
		ID:          id,
		Name:        id,
		Description: "Permite " + action + " en " + module,
	}
}

var Modules = map[string][]string{
	"user":      {"create", "read", "list", "update", "delete"},
	"role":      {"create", "read", "list", "update", "delete"},
	"permission": {"read", "list"},
	"session":   {"list", "revoke", "revoke-all"},
	"auth-log":  {"read", "list"},
}

var Catalogue = BuildCatalogue(Modules)

// BuildCatalogue genera una lista de permisos por módulo/acción.
func BuildCatalogue(modules map[string][]string) PermissionCatalogue {
	catalogue := PermissionCatalogue{}
	for module, actions := range modules {
		actionMap := ModuleActions{}
		for _, a := range actions {
			p := BuildPermission(module, a)
			actionMap[a] = p
		}
		catalogue[module] = actionMap
	}
	return catalogue
}

// Flatten devuelve todos los permisos como lista plana.
func (c PermissionCatalogue) Flatten() []Permission {
	var perms []Permission
	for _, actionMap := range c {
		for _, p := range actionMap {
			perms = append(perms, p)
		}
	}
	return perms
}

// Get devuelve el permiso id del módulo/acción.
func (c PermissionCatalogue) Get(module, action string) (Permission, bool) {
	actionMap, ok := c[module]
	if !ok {
		return Permission{}, false
	}
	p, ok := actionMap[action]
	return p, ok
}

// MustGet devuelve permiso o panic (para inicialización segura).
func (c PermissionCatalogue) MustGet(module, action string) Permission {
	p, ok := c.Get(module, action)
	if !ok {
		panic("permission not found: " + module + "." + action)
	}
	return p
}
