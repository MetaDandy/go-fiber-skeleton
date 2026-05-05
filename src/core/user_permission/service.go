package user_permission

import (
	"github.com/MetaDandy/go-fiber-skeleton/api_error"
	"github.com/MetaDandy/go-fiber-skeleton/src/model"
	"github.com/google/uuid"
)

type PermissionChecker interface {
	AllExists(ids []string) error
}

type Service interface {
	UpdateDetails(userID string, input UpdateDetails) error
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

func (s *service) UpdateDetails(userID string, input UpdateDetails) error {
	if len(input.Add) == 0 && len(input.Remove) == 0 {
		return api_error.BadRequest("At least one of add or remove must contain one element")
	}

	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		return api_error.BadRequest("Invalid user ID")
	}

	if len(input.Add) > 0 {
		if err := s.permissionChecker.AllExists(input.Add); err != nil {
			return err
		}
	}

	add, remove := removeDuplicatesBetweenArrays(input.Add, input.Remove)

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

	userPermissionsToAdd := make([]model.UserPermission, 0, len(add))
	for _, permissionID := range add {
		userPermissionsToAdd = append(userPermissionsToAdd, model.UserPermission{
			ID:           uuid.New(),
			UserID:       parsedUserID,
			PermissionID: permissionID,
		})
	}

	if err := s.repo.UpdatePermissionsTx(tx, parsedUserID, userPermissionsToAdd, remove); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}
