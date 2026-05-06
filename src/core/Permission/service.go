package permission

import (
	"github.com/MetaDandy/go-fiber-skeleton/api_error"
	"github.com/MetaDandy/go-fiber-skeleton/helper"
	"github.com/MetaDandy/go-fiber-skeleton/src/response"
)

type Service interface {
	FindByID(id string) (*response.Permission, *api_error.Error)
	FindAll(opts *helper.FindAllOptions) (*response.Paginated[response.Permission], *api_error.Error)
	AllExists(ids []string) *api_error.Error
}

type service struct {
	repo Repo
}

func NewService(repo Repo) Service {
	return &service{repo: repo}
}

func (s *service) FindByID(id string) (*response.Permission, *api_error.Error) {
	permission, err := s.repo.FindByID(id)
	if err != nil {
		return nil, api_error.NotFound("Permission not found").WithErr(err)
	}
	dto := response.PermissionToDto(&permission)

	return &dto, nil
}

func (s *service) FindAll(opts *helper.FindAllOptions) (*response.Paginated[response.Permission], *api_error.Error) {
	finded, total, err := s.repo.FindAll(opts)
	if err != nil {
		return nil, api_error.InternalServerError("Could not retrieve permissions").WithErr(err)
	}
	dtos := response.PermissionToListDto(finded)
	pages := uint((total + int64(opts.Limit) - 1) / int64(opts.Limit))

	return &response.Paginated[response.Permission]{
		Data:   dtos,
		Total:  total,
		Limit:  opts.Limit,
		Offset: opts.Offset,
		Pages:  pages,
	}, nil
}

func (s *service) AllExists(ids []string) *api_error.Error {
	if err := s.repo.AllExists(ids); err != nil {
		return api_error.InternalServerError("Could not verify permissions").WithErr(err)
	}
	return nil
}
