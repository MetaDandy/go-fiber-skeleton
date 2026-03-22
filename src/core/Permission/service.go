package permission

import (
	"github.com/MetaDandy/go-fiber-skeleton/helper"
	"github.com/MetaDandy/go-fiber-skeleton/src/response"
)

type Service interface {
	FindByID(id string) (*response.Permission, error)
	FindAll(opts *helper.FindAllOptions) (*response.Paginated[response.Permission], error)
}

type service struct {
	repo Repo
}

func NewService(repo Repo) Service {
	return &service{repo: repo}
}

func (s *service) FindByID(id string) (*response.Permission, error) {
	permission, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	dto := response.PermissionToDto(&permission)

	return &dto, nil
}

func (s *service) FindAll(opts *helper.FindAllOptions) (*response.Paginated[response.Permission], error) {
	finded, total, err := s.repo.FindAll(opts)
	if err != nil {
		return nil, err
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
