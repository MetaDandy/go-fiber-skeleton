package user

import (
	"github.com/MetaDandy/go-fiber-skeleton/helper"
	"github.com/MetaDandy/go-fiber-skeleton/src/model"
	"github.com/MetaDandy/go-fiber-skeleton/src/response"
	"github.com/google/uuid"
	"github.com/jinzhu/copier"
)

type Service interface {
	Create(input Create) (*response.User, error)
	FindByID(id string) (*response.User, error)
	FindAll(opts *helper.FindAllOptions) (*response.Paginated[response.User], error)
	Update(id string, input Update) (*response.User, error)
	Delete(id string) error
}

type service struct {
	repo Repo
}

func NewService(repo Repo) Service {
	return &service{repo: repo}
}

func (s *service) Create(input Create) (*response.User, error) {
	user := model.User{}
	copier.Copy(user, input)
	user.ID = uuid.New()

	if err := s.repo.Create(user); err != nil {
		return nil, err
	}

	dto := response.UserToDto(&user)
	return &dto, nil
}

func (s *service) FindByID(id string) (*response.User, error) {
	user, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	dto := response.UserToDto(&user)

	return &dto, nil
}

func (s *service) FindAll(opts *helper.FindAllOptions) (*response.Paginated[response.User], error) {
	finded, total, err := s.repo.FindAll(opts)
	if err != nil {
		return nil, err
	}
	dtos := response.UserToListDto(finded)
	pages := uint((total + int64(opts.Limit) - 1) / int64(opts.Limit))

	return &response.Paginated[response.User]{
		Data:   dtos,
		Total:  total,
		Limit:  opts.Limit,
		Offset: opts.Offset,
		Pages:  pages,
	}, nil
}

func (s *service) Update(id string, input Update) (*response.User, error) {
	user, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	opt := copier.Option{
		IgnoreEmpty: true,
		DeepCopy:    true,
	}

	if err := copier.CopyWithOption(user, &input, opt); err != nil {
		return nil, err
	}

	if err := s.repo.Update(user); err != nil {
		return nil, err
	}

	dto := response.UserToDto(&user)
	return &dto, nil
}

func (s *service) Delete(id string) error {
	return s.repo.Delete(id)
}
