package role

import (
	"github.com/MetaDandy/go-fiber-skeleton/api_error"
	"github.com/MetaDandy/go-fiber-skeleton/helper"
	"github.com/MetaDandy/go-fiber-skeleton/src/model"
	"github.com/MetaDandy/go-fiber-skeleton/src/response"
	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

type Service interface {
	Create(input Create) error
	FindByID(id string) (*response.Role, error)
	FindAll(opts *helper.FindAllOptions) (*response.Paginated[response.Role], error)
	UpdateHeader(id string, input UpdateHeader) error
	UpdateDetails(id string, input UpdateDetails) error
}

type PermissionChecker interface {
	AllExists(ids []string) error
}

type service struct {
	repo              Repo
	permissionChecker PermissionChecker
}

func NewService(repo Repo, permissionChecker PermissionChecker) Service {
	return &service{
		repo:              repo,
		permissionChecker: permissionChecker,
	}
}

func (s *service) Create(input Create) error {
	if len(input.Permissions) == 0 {
		return api_error.BadRequest("Role must have at least one direct permission")
	}

	if err := s.permissionChecker.AllExists(input.Permissions); err != nil {
		return err
	}

	roleID := uuid.New()

	var parentID *uuid.UUID
	var parentEffectivePermissions []model.RoleEffectivePermission

	if input.RoleID != nil && *input.RoleID != "" {
		parsedRoleID, err := uuid.Parse(*input.RoleID)
		if err != nil {
			return err
		}

		parentRole, err := s.repo.FindByID(parsedRoleID.String())
		if err != nil {
			return err
		}

		parentID = &parentRole.ID
		parentEffectivePermissions = parentRole.Role_effective_permissions

		parentEffectiveMap := make(map[string]struct{}, len(parentEffectivePermissions))
		for _, rep := range parentEffectivePermissions {
			parentEffectiveMap[rep.PermissionID] = struct{}{}
		}

		// Nueva regla:
		// si al menos un permiso ya existe en el árbol del padre, error.
		for _, permissionID := range input.Permissions {
			if _, exists := parentEffectiveMap[permissionID]; exists {
				return api_error.BadRequest(
					"Permission '" + permissionID + "' is already inherited from parent or ancestor role",
				)
			}
		}
	}

	role := model.Role{
		ID:          roleID,
		Name:        input.Name,
		Description: input.Description,
		RoleID:      parentID,
	}

	rolePermissions := make([]model.RolePermission, 0, len(input.Permissions))
	for _, permissionID := range input.Permissions {
		rolePermissions = append(rolePermissions, model.RolePermission{
			ID:           uuid.New(),
			RoleID:       roleID,
			PermissionID: permissionID,
		})
	}

	roleEffectivePermissions := make([]model.RoleEffectivePermission, 0, len(parentEffectivePermissions)+len(input.Permissions))

	// heredados del árbol del padre
	for _, inherited := range parentEffectivePermissions {
		roleEffectivePermissions = append(roleEffectivePermissions, model.RoleEffectivePermission{
			ID:           uuid.New(),
			RoleID:       roleID,
			SourceRoleID: inherited.SourceRoleID,
			PermissionID: inherited.PermissionID,
		})
	}

	// propios del rol nuevo
	for _, permissionID := range input.Permissions {
		roleEffectivePermissions = append(roleEffectivePermissions, model.RoleEffectivePermission{
			ID:           uuid.New(),
			RoleID:       roleID,
			SourceRoleID: roleID,
			PermissionID: permissionID,
		})
	}

	return s.repo.Create(role, rolePermissions, roleEffectivePermissions)
}

func (s *service) FindByID(id string) (*response.Role, error) {
	role, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	dto := response.RoleToDto(&role)
	return &dto, nil
}

func (s *service) FindAll(opts *helper.FindAllOptions) (*response.Paginated[response.Role], error) {
	finded, total, err := s.repo.FindAll(opts)
	if err != nil {
		return nil, err
	}

	dtos := response.RoleToListDto(finded)
	pages := uint((total + int64(opts.Limit) - 1) / int64(opts.Limit))

	return &response.Paginated[response.Role]{
		Data:   dtos,
		Total:  total,
		Limit:  opts.Limit,
		Offset: opts.Offset,
		Pages:  pages,
	}, nil
}

func hasAtLeastOneElement(add, remove []string) bool {
	return len(add) > 0 || len(remove) > 0
}

func removeDuplicatesBetweenArrays(add, remove []string) ([]string, []string) {
	removeSet := make(map[string]struct{}, len(remove))
	for _, id := range remove {
		removeSet[id] = struct{}{}
	}

	filteredAdd := make([]string, 0, len(add))
	common := make(map[string]struct{})

	for _, id := range add {
		if _, exists := removeSet[id]; exists {
			common[id] = struct{}{}
			continue
		}
		filteredAdd = append(filteredAdd, id)
	}

	filteredRemove := make([]string, 0, len(remove))
	for _, id := range remove {
		if _, exists := common[id]; exists {
			continue
		}
		filteredRemove = append(filteredRemove, id)
	}

	return filteredAdd, filteredRemove
}

func (s *service) UpdateHeader(id string, input UpdateHeader) error {
	tx := s.repo.BeginTx()
	if tx.Error != nil {
		return tx.Error
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	role, err := s.repo.FindByIDTx(tx, id)
	if err != nil {
		tx.Rollback()
		return err
	}

	opt := copier.Option{
		IgnoreEmpty: true,
		DeepCopy:    true,
	}

	if err := copier.CopyWithOption(&role, &input, opt); err != nil {
		tx.Rollback()
		return err
	}

	if input.RoleID != nil {
		if *input.RoleID != "" {
			parsedRoleID, err := uuid.Parse(*input.RoleID)
			if err != nil {
				tx.Rollback()
				return err
			}
			role.RoleID = &parsedRoleID
		} else {
			role.RoleID = nil
		}
	}

	if err := tx.Save(&role).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := s.normalizeDirectPermissionsAgainstParentTx(tx, role.ID); err != nil {
		tx.Rollback()
		return err
	}

	if err := s.rebuildRoleTreeTx(tx, role.ID); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (s *service) UpdateDetails(id string, input UpdateDetails) error {
	add := input.Add
	remove := input.Remove

	if len(add) == 0 && len(remove) == 0 {
		return api_error.BadRequest("At least one of add or remove must contain one element")
	}

	if len(add) > 0 {
		if err := s.permissionChecker.AllExists(add); err != nil {
			return err
		}
	}

	tx := s.repo.BeginTx()
	if tx.Error != nil {
		return tx.Error
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	role, err := s.repo.FindByIDTx(tx, id)
	if err != nil {
		tx.Rollback()
		return err
	}

	add, remove = removeDuplicatesBetweenArrays(add, remove)

	// mapa de permisos directos actuales
	currentDirect := make(map[string]struct{}, len(role.Role_permissions))
	for _, rp := range role.Role_permissions {
		currentDirect[rp.PermissionID] = struct{}{}
	}

	// filtrar adds que ya son directos del rol
	filteredAdd := make([]string, 0, len(add))
	for _, permissionID := range add {
		if _, exists := currentDirect[permissionID]; exists {
			continue
		}
		filteredAdd = append(filteredAdd, permissionID)
	}
	add = filteredAdd

	// Validar adds contra ancestros heredados
	if role.RoleID != nil && len(add) > 0 {
		parentRole, err := s.repo.FindByIDTx(tx, role.RoleID.String())
		if err != nil {
			tx.Rollback()
			return err
		}

		parentEffectiveMap := make(map[string]struct{}, len(parentRole.Role_effective_permissions))
		for _, rep := range parentRole.Role_effective_permissions {
			parentEffectiveMap[rep.PermissionID] = struct{}{}
		}

		for _, permissionID := range add {
			if _, exists := parentEffectiveMap[permissionID]; exists {
				tx.Rollback()
				return api_error.BadRequest(
					"Permission '" + permissionID + "' is already inherited from parent or ancestor role",
				)
			}
		}
	}

	// Validar removes: solo directos
	if len(remove) > 0 {
		for _, permissionID := range remove {
			if _, exists := currentDirect[permissionID]; !exists {
				tx.Rollback()
				return api_error.BadRequest(
					"Permission '" + permissionID + "' is not directly assigned to the role",
				)
			}
		}
	}

	// Regla: el rol actual debe conservar al menos un permiso directo propio
	remainingDirectCount, err := s.repo.CountDirectPermissionsNotInSetTx(tx, role.ID.String(), remove)
	if err != nil {
		tx.Rollback()
		return err
	}

	finalDirectCount := remainingDirectCount + int64(len(add))
	if finalDirectCount == 0 {
		tx.Rollback()
		return api_error.BadRequest("Role must keep at least one direct permission of its own")
	}

	// Validar add con descendientes si strictMode = false
	if !input.StrictMode {
		for _, permissionID := range add {
			descendants, err := s.repo.DescendantsWithDirectPermissionTx(tx, id, permissionID)
			if err != nil {
				tx.Rollback()
				return err
			}
			if len(descendants) > 0 {
				tx.Rollback()
				return api_error.BadRequest(
					"Permission '" + permissionID + "' is directly assigned in descendant role: " + descendants[0].Name,
				)
			}
		}
	}

	// Actualizar directos del rol actual
	rolePermissionsToAdd := make([]model.RolePermission, 0, len(add))
	for _, permissionID := range add {
		rolePermissionsToAdd = append(rolePermissionsToAdd, model.RolePermission{
			ID:           uuid.New(),
			RoleID:       role.ID,
			PermissionID: permissionID,
		})
	}

	if err := s.repo.UpdateRolePermissionsTx(tx, id, rolePermissionsToAdd, remove); err != nil {
		tx.Rollback()
		return err
	}

	// Propagación incremental add
	for _, permissionID := range add {
		if err := s.propagateAddTx(tx, role.ID, permissionID, input.StrictMode); err != nil {
			tx.Rollback()
			return err
		}
	}

	// Propagación incremental remove
	for _, permissionID := range remove {
		if err := s.propagateRemoveTx(tx, role.ID, permissionID); err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

func (s *service) propagateAddTx(tx *gorm.DB, roleID uuid.UUID, permissionID string, strictMode bool) error {
	// 1. el rol actual recibe effective propio
	exists, err := s.repo.HasEffectivePermissionTx(tx, roleID.String(), permissionID)
	if err != nil {
		return err
	}
	if !exists {
		if err := s.repo.UpsertEffectivePermissionTx(tx, model.RoleEffectivePermission{
			ID:           uuid.New(),
			RoleID:       roleID,
			SourceRoleID: roleID,
			PermissionID: permissionID,
		}); err != nil {
			return err
		}
	}

	// 2. recorrer descendientes
	descendants, err := s.repo.FindDescendantsOrderedTx(tx, roleID.String())
	if err != nil {
		return err
	}

	for _, d := range descendants {
		if strictMode {
			hasDirect, err := s.repo.HasDirectPermissionTx(tx, d.ID.String(), permissionID)
			if err != nil {
				return err
			}

			if hasDirect {
				remainingDirectCount, err := s.repo.CountDirectPermissionsNotInSetTx(
					tx,
					d.ID.String(),
					[]string{permissionID},
				)
				if err != nil {
					return err
				}

				if remainingDirectCount == 0 {
					return api_error.BadRequest(
						"Descendant role '" + d.Name + "' must keep at least one direct permission of its own",
					)
				}

				if err := s.repo.DeleteDirectPermissionTx(tx, d.ID.String(), permissionID); err != nil {
					return err
				}

				// borrar solo el effective propio del descendiente
				if err := s.repo.DeleteEffectivePermissionBySourceTx(
					tx,
					d.ID.String(),
					permissionID,
					d.ID.String(),
				); err != nil {
					return err
				}
			}
		}

		hasEffective, err := s.repo.HasEffectivePermissionTx(tx, d.ID.String(), permissionID)
		if err != nil {
			return err
		}
		if !hasEffective {
			if err := s.repo.UpsertEffectivePermissionTx(tx, model.RoleEffectivePermission{
				ID:           uuid.New(),
				RoleID:       d.ID,
				SourceRoleID: roleID,
				PermissionID: permissionID,
			}); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *service) propagateRemoveTx(tx *gorm.DB, roleID uuid.UUID, permissionID string) error {
	// borrar effective del rol actual originado por él mismo
	if err := s.repo.DeleteEffectivePermissionBySourceTx(tx, roleID.String(), permissionID, roleID.String()); err != nil {
		return err
	}

	descendants, err := s.repo.FindDescendantsOrderedTx(tx, roleID.String())
	if err != nil {
		return err
	}

	for _, d := range descendants {
		if err := s.repo.DeleteEffectivePermissionBySourceTx(tx, d.ID.String(), permissionID, roleID.String()); err != nil {
			return err
		}
	}

	return nil
}

func (s *service) rebuildSingleRoleEffectivePermissionsTx(tx *gorm.DB, role model.Role) error {
	// borrar todos los effective actuales del rol
	if err := tx.
		Where("role_id = ?", role.ID).
		Delete(&model.RoleEffectivePermission{}).Error; err != nil {
		return err
	}

	seen := make(map[string]struct{})

	// 1. heredar del nuevo padre
	if role.RoleID != nil {
		parentRole, err := s.repo.FindByIDTx(tx, role.RoleID.String())
		if err != nil {
			return err
		}

		for _, inherited := range parentRole.Role_effective_permissions {
			if _, exists := seen[inherited.PermissionID]; exists {
				continue
			}

			if err := s.repo.UpsertEffectivePermissionTx(tx, model.RoleEffectivePermission{
				ID:           uuid.New(),
				RoleID:       role.ID,
				SourceRoleID: inherited.SourceRoleID,
				PermissionID: inherited.PermissionID,
			}); err != nil {
				return err
			}

			seen[inherited.PermissionID] = struct{}{}
		}
	}

	// 2. reinsertar directos propios del rol
	for _, rp := range role.Role_permissions {
		if _, exists := seen[rp.PermissionID]; exists {
			continue
		}

		if err := s.repo.UpsertEffectivePermissionTx(tx, model.RoleEffectivePermission{
			ID:           uuid.New(),
			RoleID:       role.ID,
			SourceRoleID: role.ID,
			PermissionID: rp.PermissionID,
		}); err != nil {
			return err
		}

		seen[rp.PermissionID] = struct{}{}
	}

	return nil
}
func (s *service) rebuildRoleTreeTx(tx *gorm.DB, roleID uuid.UUID) error {
	role, err := s.repo.FindByIDTx(tx, roleID.String())
	if err != nil {
		return err
	}

	if err := s.rebuildSingleRoleEffectivePermissionsTx(tx, role); err != nil {
		return err
	}

	children, err := s.repo.FindChildrenTx(tx, roleID.String())
	if err != nil {
		return err
	}

	for _, child := range children {
		if err := s.rebuildRoleTreeTx(tx, child.ID); err != nil {
			return err
		}
	}

	return nil
}

func (s *service) normalizeDirectPermissionsAgainstParentTx(tx *gorm.DB, roleID uuid.UUID) error {
	role, err := s.repo.FindByIDTx(tx, roleID.String())
	if err != nil {
		return err
	}

	if role.RoleID == nil {
		return nil
	}

	parentRole, err := s.repo.FindByIDTx(tx, role.RoleID.String())
	if err != nil {
		return err
	}

	// permisos que el nuevo árbol ya le aporta
	parentEffectiveMap := make(map[string]string, len(parentRole.Role_effective_permissions))
	duplicatedPermissionIDs := make([]string, 0)

	for _, ep := range parentRole.Role_effective_permissions {
		parentEffectiveMap[ep.PermissionID] = ep.SourceRoleID.String()
	}

	// detectar cuáles directos del hijo chocan con el nuevo árbol
	for _, rp := range role.Role_permissions {
		if _, exists := parentEffectiveMap[rp.PermissionID]; exists {
			duplicatedPermissionIDs = append(duplicatedPermissionIDs, rp.PermissionID)
		}
	}

	// si no hay conflictos, no hay nada que hacer
	if len(duplicatedPermissionIDs) == 0 {
		return nil
	}

	// regla nueva: el rol debe conservar al menos un permiso directo propio
	remainingDirectCount, err := s.repo.CountDirectPermissionsNotInSetTx(tx, role.ID.String(), duplicatedPermissionIDs)
	if err != nil {
		return err
	}

	if remainingDirectCount == 0 {
		return api_error.BadRequest("Role must keep at least one direct permission of its own")
	}

	// para cada permiso que ahora será heredado:
	// 1. borrar de rolepermissions
	// 2. actualizar source_role_id en effective en vez de delete+insert
	for _, permissionID := range duplicatedPermissionIDs {
		if err := s.repo.DeleteDirectPermissionTx(tx, role.ID.String(), permissionID); err != nil {
			return err
		}

		newSourceRoleID := parentEffectiveMap[permissionID]

		if err := s.repo.UpdateEffectivePermissionSourceTx(
			tx,
			role.ID.String(),
			permissionID,
			role.ID.String(),
			newSourceRoleID,
		); err != nil {
			return err
		}
	}

	return nil
}
