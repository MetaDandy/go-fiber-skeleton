package role

import (
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
	Update(id string, input Update) error
	Delete(id string) error
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

func (s *service) Update(id string, input Update) error {
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

	if err := s.repo.Update(role); err != nil {
		return err
	}

	return nil
}

func (s *service) Delete(id string) error {
	return s.repo.Delete(id)
}
