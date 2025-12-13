package user

import (
	"github.com/MetaDandy/go-fiber-skeleton/helper"
	"github.com/MetaDandy/go-fiber-skeleton/src/model"
	"github.com/MetaDandy/go-fiber-skeleton/src/response"
	"github.com/google/uuid"
	"github.com/jinzhu/copier"
)

type UserService interface {
	Create(input CreateUserDto) (*response.UserResponse, error)
	FindByID(id string) (*response.UserResponse, error)
	FindAll(opts *helper.FindAllOptions) (*helper.PaginatedResponse[response.UserResponse], error)
	Update(id string, input UpdateUserDto) (*response.UserResponse, error)
	Delete(id string) error
}

type Service struct {
	repo UserRepo
}

func NewService(repo UserRepo) UserService {
	return &Service{repo: repo}
}

func (s *Service) Create(input CreateUserDto) (*response.UserResponse, error) {
	user := model.User{}
	copier.Copy(user, input)
	user.ID = uuid.New()

	if err := s.repo.Create(user); err != nil {
		return nil, err
	}

	dto := response.UserToDto(&user)
	return &dto, nil
}

func (s *Service) FindByID(id string) (*response.UserResponse, error) {
	user, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	dto := response.UserToDto(&user)

	return &dto, nil
}

func (s *Service) FindAll(opts *helper.FindAllOptions) (*helper.PaginatedResponse[response.UserResponse], error) {
	finded, total, err := s.repo.FindAll(opts)
	if err != nil {
		return nil, err
	}
	dtos := response.UserToListDto(finded)
	pages := uint((total + int64(opts.Limit) - 1) / int64(opts.Limit))

	return &helper.PaginatedResponse[response.UserResponse]{
		Data:   dtos,
		Total:  total,
		Limit:  opts.Limit,
		Offset: opts.Offset,
		Pages:  pages,
	}, nil
}

func (s *Service) Update(id string, input UpdateUserDto) (*response.UserResponse, error) {
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

func (s *Service) Delete(id string) error {
	return s.repo.Delete(id)
}
