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
	Create(input Create) *api_error.Error
	FindByID(id string) (*response.Role, *api_error.Error)
	FindAll(opts *helper.FindAllOptions) (*response.Paginated[response.Role], *api_error.Error)
	UpdateHeader(id string, input UpdateHeader) *api_error.Error
	UpdateDetails(id string, input UpdateDetails) *api_error.Error
}

type PermissionChecker interface {
	AllExists(ids []string) *api_error.Error
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

func (s *service) Create(input Create) *api_error.Error {
	if len(input.Permissions) == 0 {
		return api_error.BadRequest("Role must have at least one direct permission")
	}

	if err := s.permissionChecker.AllExists(input.Permissions); err != nil {
		return api_error.InternalServerError("Internal error").WithErr(err)
	}

	roleID := uuid.New()

	var parentID *uuid.UUID
	var parentEffectivePermissions []model.RoleEffectivePermission

	if input.RoleID != nil && *input.RoleID != "" {
		parsedRoleID, err := uuid.Parse(*input.RoleID)
		if err != nil {
			return api_error.InternalServerError("Internal error").WithErr(err)
		}

		parentRole, err := s.repo.FindByID(parsedRoleID)
		if err != nil {
			return api_error.InternalServerError("Internal error").WithErr(err)
		}

		parentID = &parentRole.ID
		parentEffectivePermissions = parentRole.Role_effective_permissions

		parentEffectiveMap := make(map[string]struct{}, len(parentEffectivePermissions))
		for _, rep := range parentEffectivePermissions {
			parentEffectiveMap[rep.PermissionID] = struct{}{}
		}

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

	for _, inherited := range parentEffectivePermissions {
		roleEffectivePermissions = append(roleEffectivePermissions, model.RoleEffectivePermission{
			ID:           uuid.New(),
			RoleID:       roleID,
			SourceRoleID: inherited.SourceRoleID,
			PermissionID: inherited.PermissionID,
		})
	}

	for _, permissionID := range input.Permissions {
		roleEffectivePermissions = append(roleEffectivePermissions, model.RoleEffectivePermission{
			ID:           uuid.New(),
			RoleID:       roleID,
			SourceRoleID: roleID,
			PermissionID: permissionID,
		})
	}

	if err := s.repo.Create(role, rolePermissions, roleEffectivePermissions); err != nil {
		return api_error.InternalServerError("Repository error").WithErr(err)
	}
	return nil
}

func (s *service) FindByID(id string) (*response.Role, *api_error.Error) {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return nil, api_error.BadRequest("Invalid role ID: " + id)
	}

	role, err := s.repo.FindByID(parsedID)
	if err != nil {
		return nil, api_error.InternalServerError("Database error").WithErr(err)
	}

	dto := response.RoleToDto(&role)
	return &dto, nil
}

func (s *service) FindAll(opts *helper.FindAllOptions) (*response.Paginated[response.Role], *api_error.Error) {
	finded, total, err := s.repo.FindAll(opts)
	if err != nil {
		return nil, api_error.InternalServerError("Database error").WithErr(err)
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

func removeDuplicatesBetweenStringArrays(add, remove []string) ([]string, []string) {
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

func (s *service) UpdateHeader(id string, input UpdateHeader) *api_error.Error {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return api_error.BadRequest("Invalid role ID: " + id)
	}

	tx := s.repo.BeginTx()
	if tx.Error != nil {
		return api_error.InternalServerError("Database error").WithErr(tx.Error)
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	role, err := s.repo.FindByIDTx(tx, parsedID)
	if err != nil {
		tx.Rollback()
		return api_error.InternalServerError("Internal error").WithErr(err)
	}

	opt := copier.Option{
		IgnoreEmpty: true,
		DeepCopy:    true,
	}

	if err := copier.CopyWithOption(&role, &input, opt); err != nil {
		tx.Rollback()
		return api_error.InternalServerError("Internal error").WithErr(err)
	}

	if input.RoleID != nil {
		if *input.RoleID != "" {
			parsedRoleID, err := uuid.Parse(*input.RoleID)
			if err != nil {
				tx.Rollback()
				return api_error.InternalServerError("Internal error").WithErr(err)
			}
			role.RoleID = &parsedRoleID
		} else {
			role.RoleID = nil
		}
	}

	if err := tx.Save(&role).Error; err != nil {
		tx.Rollback()
		return api_error.InternalServerError("Internal error").WithErr(err)
	}

	if err := s.normalizeDirectPermissionsAgainstParentTx(tx, role.ID, input.StrictMode); err != nil {
		tx.Rollback()
		return api_error.InternalServerError("Internal error").WithErr(err)
	}

	if err := s.rebuildRoleTreeTx(tx, role.ID); err != nil {
		tx.Rollback()
		return api_error.InternalServerError("Internal error").WithErr(err)
	}

	return api_error.InternalServerError("Failed to commit").WithErr(tx.Commit().Error)
}

func (s *service) UpdateDetails(id string, input UpdateDetails) *api_error.Error {
	add := input.Add
	remove := input.Remove

	if len(add) == 0 && len(remove) == 0 {
		return api_error.BadRequest("At least one of add or remove must contain one element")
	}

	if len(add) > 0 {
		if err := s.permissionChecker.AllExists(add); err != nil {
			return api_error.InternalServerError("Internal error").WithErr(err)
		}
	}

	parsedID, err := uuid.Parse(id)
	if err != nil {
		return api_error.BadRequest("Invalid role ID: " + id)
	}

	tx := s.repo.BeginTx()
	if tx.Error != nil {
		return api_error.InternalServerError("Database error").WithErr(tx.Error)
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	role, err := s.repo.FindByIDTx(tx, parsedID)
	if err != nil {
		tx.Rollback()
		return api_error.InternalServerError("Internal error").WithErr(err)
	}

	// Remove duplicates between add and remove
	add, remove = removeDuplicatesBetweenStringArrays(add, remove)

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
		parentRole, err := s.repo.FindByIDTx(tx, *role.RoleID)
		if err != nil {
			tx.Rollback()
			return api_error.InternalServerError("Internal error").WithErr(err)
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
	remainingDirectCount, err := s.repo.CountDirectPermissionsNotInSetTx(tx, role.ID, remove)
	if err != nil {
		tx.Rollback()
		return api_error.InternalServerError("Internal error").WithErr(err)
	}

	finalDirectCount := remainingDirectCount + int64(len(add))
	if finalDirectCount == 0 {
		tx.Rollback()
		return api_error.BadRequest("Role must keep at least one direct permission of its own")
	}

	// Validar add con descendientes si strictMode = false
	if !input.StrictMode {
		for _, permissionID := range add {
			descendants, err := s.repo.DescendantsWithDirectPermissionTx(tx, parsedID, permissionID)
			if err != nil {
				tx.Rollback()
				return api_error.InternalServerError("Internal error").WithErr(err)
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

	if err := s.repo.UpdateRolePermissionsTx(tx, parsedID, rolePermissionsToAdd, remove); err != nil {
		tx.Rollback()
		return api_error.InternalServerError("Internal error").WithErr(err)
	}

	// Propagación incremental add
	for _, permissionID := range add {
		if err := s.propagateAddTx(tx, role.ID, permissionID, input.StrictMode); err != nil {
			tx.Rollback()
			return api_error.InternalServerError("Internal error").WithErr(err)
		}
	}

	// Propagación incremental remove
	for _, permissionID := range remove {
		if err := s.propagateRemoveTx(tx, role.ID, permissionID); err != nil {
			tx.Rollback()
			return api_error.InternalServerError("Internal error").WithErr(err)
		}
	}

	return api_error.InternalServerError("Failed to commit").WithErr(tx.Commit().Error)
}

func (s *service) propagateAddTx(tx *gorm.DB, roleID uuid.UUID, permissionID string, strictMode bool) *api_error.Error {
	descendants, err := s.repo.FindDescendantsOrderedTx(tx, roleID)
	if err != nil {
		return api_error.InternalServerError("Internal error").WithErr(err)
	}

	bulkUpserts := make([]model.RoleEffectivePermission, 0, len(descendants)+1)
	bulkUpserts = append(bulkUpserts, model.RoleEffectivePermission{
		ID:           uuid.New(),
		RoleID:       roleID,
		SourceRoleID: roleID,
		PermissionID: permissionID,
	})

	for _, d := range descendants {
		bulkUpserts = append(bulkUpserts, model.RoleEffectivePermission{
			ID:           uuid.New(),
			RoleID:       d.ID,
			SourceRoleID: roleID,
			PermissionID: permissionID,
		})
	}

	if strictMode && len(descendants) > 0 {
		descendantIDs := make([]uuid.UUID, 0, len(descendants))
		idToName := make(map[uuid.UUID]string)
		for _, d := range descendants {
			descendantIDs = append(descendantIDs, d.ID)
			idToName[d.ID] = d.Name
		}

		rolesWithDirect, err := s.repo.GetRolesWithDirectPermissionTx(tx, descendantIDs, permissionID)
		if err != nil {
			return api_error.InternalServerError("Internal error").WithErr(err)
		}

		if len(rolesWithDirect) > 0 {
			counts, err := s.repo.GetDirectPermissionsCountsTx(tx, rolesWithDirect)
			if err != nil {
				return api_error.InternalServerError("Internal error").WithErr(err)
			}

			for _, id := range rolesWithDirect {
				if counts[id] <= 1 {
					return api_error.BadRequest(
						"Descendant role '" + idToName[id] + "' must keep at least one direct permission of its own",
					)
				}
			}

			if err := s.repo.DeleteDirectPermissionsBatchTx(tx, rolesWithDirect, permissionID); err != nil {
				return api_error.InternalServerError("Internal error").WithErr(err)
			}
			if err := s.repo.DeleteOwnEffectivePermissionsTx(tx, rolesWithDirect, permissionID); err != nil {
				return api_error.InternalServerError("Internal error").WithErr(err)
			}
		}
	}

	if err := s.repo.UpsertEffectivePermissionsBatchTx(tx, bulkUpserts); err != nil {
		return api_error.InternalServerError("Repository error").WithErr(err)
	}
	return nil
}

func (s *service) propagateRemoveTx(tx *gorm.DB, roleID uuid.UUID, permissionID string) *api_error.Error {
	descendants, err := s.repo.FindDescendantsOrderedTx(tx, roleID)
	if err != nil {
		return api_error.InternalServerError("Internal error").WithErr(err)
	}

	roleIDs := make([]uuid.UUID, 0, len(descendants)+1)
	roleIDs = append(roleIDs, roleID)

	for _, d := range descendants {
		roleIDs = append(roleIDs, d.ID)
	}

	if err := s.repo.DeleteEffectivePermissionsBySourceAndRolesTx(tx, roleIDs, permissionID, roleID); err != nil {
		return api_error.InternalServerError("Repository error").WithErr(err)
	}
	return nil
}

func (s *service) rebuildSingleRoleEffectivePermissionsTx(tx *gorm.DB, role model.Role) *api_error.Error {
	if err := tx.
		Where("role_id = ?", role.ID).
		Delete(&model.RoleEffectivePermission{}).Error; err != nil {
		return api_error.InternalServerError("Internal error").WithErr(err)
	}

	seen := make(map[string]struct{})

	if role.RoleID != nil {
		parentRole, err := s.repo.FindByIDTx(tx, *role.RoleID)
		if err != nil {
			return api_error.InternalServerError("Internal error").WithErr(err)
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
				return api_error.InternalServerError("Internal error").WithErr(err)
			}

			seen[inherited.PermissionID] = struct{}{}
		}
	}

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
			return api_error.InternalServerError("Internal error").WithErr(err)
		}

		seen[rp.PermissionID] = struct{}{}
	}

	return nil
}

func (s *service) rebuildRoleTreeTx(tx *gorm.DB, roleID uuid.UUID) *api_error.Error {
	role, err := s.repo.FindByIDTx(tx, roleID)
	if err != nil {
		return api_error.InternalServerError("Internal error").WithErr(err)
	}

	if err := s.rebuildSingleRoleEffectivePermissionsTx(tx, role); err != nil {
		return api_error.InternalServerError("Internal error").WithErr(err)
	}

	children, err := s.repo.FindDescendantsOrderedTx(tx, roleID)
	if err != nil {
		return api_error.InternalServerError("Internal error").WithErr(err)
	}

	for _, child := range children {
		freshChild, err := s.repo.FindByIDTx(tx, child.ID)
		if err != nil {
			return api_error.InternalServerError("Internal error").WithErr(err)
		}
		if err := s.rebuildSingleRoleEffectivePermissionsTx(tx, freshChild); err != nil {
			return api_error.InternalServerError("Internal error").WithErr(err)
		}
	}

	return nil
}

func (s *service) normalizeDirectPermissionsAgainstParentTx(tx *gorm.DB, roleID uuid.UUID, strictMode bool) *api_error.Error {
	role, err := s.repo.FindByIDTx(tx, roleID)
	if err != nil {
		return api_error.InternalServerError("Internal error").WithErr(err)
	}

	if role.RoleID == nil {
		return nil
	}

	parentRole, err := s.repo.FindByIDTx(tx, *role.RoleID)
	if err != nil {
		return api_error.InternalServerError("Internal error").WithErr(err)
	}

	parentEffectiveMap := make(map[string]string, len(parentRole.Role_effective_permissions))
	duplicatedPermissionIDs := make([]string, 0)

	for _, ep := range parentRole.Role_effective_permissions {
		parentEffectiveMap[ep.PermissionID] = ep.SourceRoleID.String()
	}

	for _, rp := range role.Role_permissions {
		if _, exists := parentEffectiveMap[rp.PermissionID]; exists {
			duplicatedPermissionIDs = append(duplicatedPermissionIDs, rp.PermissionID)
		}
	}

	if len(duplicatedPermissionIDs) == 0 {
		return nil
	}

	if !strictMode {
		return api_error.BadRequest("The new parent role already provides one or more identical permissions currently assigned directly to this role. Please remove these overlapped permissions manually or send strict_mode=true to automatically override them.")
	}

	remainingDirectCount, err := s.repo.CountDirectPermissionsNotInSetTx(tx, role.ID, duplicatedPermissionIDs)
	if err != nil {
		return api_error.InternalServerError("Internal error").WithErr(err)
	}

	if remainingDirectCount == 0 {
		return api_error.BadRequest("Role would be left with zero direct permissions due to overlap with the new parent role.")
	}

	for _, permissionID := range duplicatedPermissionIDs {
		if err := s.repo.DeleteDirectPermissionTx(tx, role.ID, permissionID); err != nil {
			return api_error.InternalServerError("Internal error").WithErr(err)
		}

		newSourceRoleID := parentEffectiveMap[permissionID]

		if err := s.repo.UpdateEffectivePermissionSourceTx(
			tx,
			role.ID,
			permissionID,
			role.ID,
			uuid.MustParse(newSourceRoleID),
		); err != nil {
			return api_error.InternalServerError("Internal error").WithErr(err)
		}
	}

	return nil
}
