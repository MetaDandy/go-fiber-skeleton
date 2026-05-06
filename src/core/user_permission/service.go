package user_permission

import (
	"github.com/MetaDandy/go-fiber-skeleton/api_error"
	"github.com/MetaDandy/go-fiber-skeleton/src/model"
	"github.com/google/uuid"
)

type PermissionChecker interface {
	AllExists(ids []string) *api_error.Error
}

type Service interface {
	// UpdateDetails updates a user's permissions by adding and/or removing permission IDs.
	//
	// Flow:
	//   1. Validates that at least one of Add or Remove contains elements
	//   2. Parses and validates the userID as a valid UUID
	//   3. Verifies that all permission IDs in Add exist (via PermissionChecker)
	//   4. Removes duplicate IDs that appear in both Add and Remove arrays
	//   5. Begins a database transaction
	//   6. Calls UpdatePermissionsTx to persist changes
	//   7. Commits the transaction
	//
	// Possible errors:
	//   - BadRequest (400): Invalid user ID, or both Add and Remove are empty
	//   - InternalServerError (500): Database error, permission check failure, or commit failure
	UpdateDetails(userID string, input UpdateDetails) *api_error.Error
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

func (s *service) UpdateDetails(userID string, input UpdateDetails) *api_error.Error {
	if len(input.Add) == 0 && len(input.Remove) == 0 {
		return api_error.BadRequest("At least one of add or remove must contain one element")
	}

	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		return api_error.BadRequest("Invalid user ID")
	}

	if len(input.Add) > 0 {
		if err := s.permissionChecker.AllExists(input.Add); err != nil {
			return api_error.InternalServerError("Internal error").WithErr(err)
		}
	}

	add, remove := removeDuplicatesBetweenArrays(input.Add, input.Remove)

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
		return api_error.InternalServerError("Internal error").WithErr(err)
	}

	if err := tx.Commit().Error; err != nil {
		return api_error.InternalServerError("Failed to commit").WithErr(err)
	}
	return nil
}
