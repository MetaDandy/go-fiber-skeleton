package task

import (
	"fmt"

	"github.com/MetaDandy/go-fiber-skeleton/helper"
	"github.com/MetaDandy/go-fiber-skeleton/src/enum"
	"github.com/MetaDandy/go-fiber-skeleton/src/model"
	"github.com/MetaDandy/go-fiber-skeleton/src/response"
	"github.com/google/uuid"
	"github.com/jinzhu/copier"
)

type Service interface {
	Create(input Create) (*response.Task, error)
	FindByID(id string) (*response.Task, error)
	FindAll(opts *helper.FindAllOptions) (*response.Paginated[response.Task], error)
	Update(id string, input Update) (*response.Task, error)
	Delete(id string) error
}

type userRepo interface {
	Exists(id string) error
}

type service struct {
	repo  Repo
	uRepo userRepo
}

func NewService(repo Repo, uRepo userRepo) Service {
	return &service{repo: repo, uRepo: uRepo}
}

func (s *service) Create(input Create) (*response.Task, error) {
	status, err := enum.ParseStatus(input.Status)
	if err != nil {
		return nil, err
	}

	err = s.uRepo.Exists(input.UserID)
	if err != nil {
		return nil, fmt.Errorf("user with id: %v not exist %v", input.UserID, err)
	}

	task := model.Task{
		ID:          uuid.New(),
		Title:       input.Title,
		Description: input.Description,
		Status:      status,
		UserID:      uuid.MustParse(input.UserID),
	}

	if err := s.repo.Create(task); err != nil {
		return nil, err
	}

	dto := response.TaskToDto(&task)
	return &dto, nil
}

func (s *service) FindByID(id string) (*response.Task, error) {
	task, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	dto := response.TaskToDto(&task)

	return &dto, nil
}

func (s *service) FindAll(opts *helper.FindAllOptions) (*response.Paginated[response.Task], error) {
	finded, total, err := s.repo.FindAll(opts)
	if err != nil {
		return nil, err
	}
	dtos := response.TaskToListDto(finded)
	pages := uint((total + int64(opts.Limit) - 1) / int64(opts.Limit))

	return &response.Paginated[response.Task]{
		Data:   dtos,
		Total:  total,
		Limit:  opts.Limit,
		Offset: opts.Offset,
		Pages:  pages,
	}, nil
}

func (s *service) Update(id string, input Update) (*response.Task, error) {
	task, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	opt := copier.Option{
		IgnoreEmpty: true,
		DeepCopy:    true,
	}

	if err := copier.CopyWithOption(&task, &input, opt); err != nil {
		return nil, err
	}

	if err := s.repo.Update(task); err != nil {
		return nil, err
	}

	dto := response.TaskToDto(&task)
	return &dto, nil
}

func (s *service) Delete(id string) error {
	return s.repo.Delete(id)
}
