package role

import (
	"github.com/MetaDandy/go-fiber-skeleton/api_error"
	"github.com/MetaDandy/go-fiber-skeleton/helper"
	"github.com/MetaDandy/go-fiber-skeleton/src/model"
	"github.com/MetaDandy/go-fiber-skeleton/src/response"
	"github.com/google/uuid"
	"github.com/jinzhu/copier"
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
	if err := s.permissionChecker.AllExists(input.Permissions); err != nil {
		return err
	}

	roleID := uuid.New()

	var parentID *uuid.UUID
	if input.RoleID != nil && *input.RoleID != "" {
		parsedRoleID, err := uuid.Parse(*input.RoleID)
		if err != nil {
			return err
		}
		parentID = &parsedRoleID
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

	return s.repo.Create(role, rolePermissions)
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
	role, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}

	opt := copier.Option{
		IgnoreEmpty: true,
		DeepCopy:    true,
	}

	if err := copier.CopyWithOption(&role, &input, opt); err != nil {
		return err
	}

	if input.RoleID != nil {
		if *input.RoleID != "" {
			parsedRoleID, err := uuid.Parse(*input.RoleID)
			if err != nil {
				return err
			}
			role.RoleID = &parsedRoleID
		} else {
			role.RoleID = nil
		}
	}

	return s.repo.UpdateHeader(role)
}

func (s *service) UpdateDetails(id string, input UpdateDetails) error {
	add := input.Add
	remove := input.Remove

	if !hasAtLeastOneElement(add, remove) {
		return api_error.BadRequest("At least one of add or remove must contain one element")
	}

	add, remove = removeDuplicatesBetweenArrays(add, remove)

	if !hasAtLeastOneElement(add, remove) {
		return api_error.BadRequest("At least one of add or remove must contain one element")
	}

	role, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}

	currentPermissions := make(map[string]model.RolePermission, len(role.Role_permissions))
	for _, rp := range role.Role_permissions {
		currentPermissions[rp.PermissionID] = rp
	}

	filteredAdd := make([]string, 0, len(add))
	for _, permissionID := range add {
		if _, exists := currentPermissions[permissionID]; exists {
			continue
		}
		filteredAdd = append(filteredAdd, permissionID)
	}

	filteredRemove := make([]string, 0, len(remove))
	for _, permissionID := range remove {
		if _, exists := currentPermissions[permissionID]; !exists {
			continue
		}
		filteredRemove = append(filteredRemove, permissionID)
	}

	if !hasAtLeastOneElement(filteredAdd, filteredRemove) {
		return api_error.BadRequest("No valid permissions to add or remove")
	}

	if err := s.permissionChecker.AllExists(filteredAdd); err != nil {
		return err
	}

	rolePermissionsToAdd := make([]model.RolePermission, 0, len(filteredAdd))
	for _, permissionID := range filteredAdd {
		rolePermissionsToAdd = append(rolePermissionsToAdd, model.RolePermission{
			ID:           uuid.New(),
			RoleID:       role.ID,
			PermissionID: permissionID,
		})
	}

	return s.repo.UpdateDetails(role.ID.String(), rolePermissionsToAdd, filteredRemove)
}
