package authentication

import (
	"context"
	"errors"
	"testing"

	"github.com/MetaDandy/go-fiber-skeleton/constant"
	"github.com/MetaDandy/go-fiber-skeleton/src/model"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"golang.org/x/crypto/bcrypt"
)

// testUserRepo implements passwordURepo for testing
type testUserRepo struct {
	db *gorm.DB
}

func NewTestUserRepo(db *gorm.DB) *testUserRepo {
	return &testUserRepo{db: db}
}

func (r *testUserRepo) FindByEmail(email string) (model.User, error) {
	var user model.User
	err := r.db.Where("email = ?", email).First(&user).Error
	return user, err
}

func (r *testUserRepo) ExistsByEmail(email string) error {
	var count int64
	err := r.db.Model(&model.User{}).Where("email = ?", email).Count(&count).Error
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New("record found")
	}
	return gorm.ErrRecordNotFound
}

func (r *testUserRepo) FindByID(id string) (model.User, error) {
	var user model.User
	err := r.db.Where("id = ?", id).First(&user).Error
	return user, err
}

func (r *testUserRepo) UpdatePassword(userID string, passwordHash string) error {
	return r.db.Model(&model.User{}).Where("id = ?", userID).Update("password", passwordHash).Error
}

// mockEmailService is a test double for mail.EmailService
type mockEmailService struct {
	sentEmails []sentEmail
}

type sentEmail struct {
	to   string
	name string
}

func (m *mockEmailService) SendVerificationEmail(ctx context.Context, to, name, token string) error {
	m.sentEmails = append(m.sentEmails, sentEmail{to: to, name: name})
	return nil
}

func (m *mockEmailService) SendPasswordReset(ctx context.Context, to, name, resetLink string) error {
	m.sentEmails = append(m.sentEmails, sentEmail{to: to, name: name})
	return nil
}

func (m *mockEmailService) SendWelcome(ctx context.Context, to, name string) error {
	m.sentEmails = append(m.sentEmails, sentEmail{to: to, name: name})
	return nil
}

// TestSignUpPassword_CreatesUser_WithHashedPassword tests the happy path for password signup.
// It verifies:
// 1. User is created successfully
// 2. Password is hashed (not stored as plaintext)
// 3. Email verification token is generated
// 4. Verification email is sent
func TestSignUpPassword_CreatesUser_WithHashedPassword(t *testing.T) {
	db := setupTestContainer(t)

	// Create the default role that the signup will reference
	defaultRole := &model.Role{
		ID:   constant.GenericID,
		Name: "Generic Role",
	}
	err := db.Create(defaultRole).Error
	assert.NoError(t, err, "should create default role")

	// Create real auth repo and test user repo
	authRepo := NewRepo(db)
	uRepo := NewTestUserRepo(db)
	mockMailSvc := &mockEmailService{}
	appURL := "http://localhost:3000"

	// Create password service
	passwordSvc := NewPasswordService(authRepo, uRepo, mockMailSvc, appURL)

	// Input for signup
	input := SignUpPassword{
		Email:          "test@example.com",
		Password:       "SecurePass123!",
		RepeatPassword: "SecurePass123!",
		Ip:             "127.0.0.1",
		UserAgent:      "test-agent",
	}

	// Execute signup
	signUpErr := passwordSvc.SignUpPassword(input)

	// Verify no error - the service returns *api_error.Error
	if signUpErr != nil {
		t.Fatalf("SignUpPassword should not return an error, got: %v", signUpErr)
	}

	// Verify user was created in database
	var createdUser model.User
	result := db.Where("email = ?", input.Email).First(&createdUser)
	assert.NoError(t, result.Error, "User should be created in database")
	assert.Equal(t, input.Email, createdUser.Email, "Email should match")
	assert.False(t, createdUser.EmailVerified, "Email should not be verified initially")

	// Verify password is hashed (not plaintext)
	assert.NotNil(t, createdUser.Password, "Password should be stored")
	assert.NotEqual(t, input.Password, *createdUser.Password, "Password should be hashed, not plaintext")

	// Verify password hash is valid bcrypt
	compareErr := bcrypt.CompareHashAndPassword([]byte(*createdUser.Password), []byte(input.Password))
	assert.NoError(t, compareErr, "Password hash should be valid bcrypt")

	// Verify email verification token was created
	var tokenCount int64
	db.Model(&model.EmailVerificationToken{}).Where("user_id = ?", createdUser.ID).Count(&tokenCount)
	assert.Equal(t, int64(1), tokenCount, "One verification token should be created")

	// Verify verification email was sent
	assert.Len(t, mockMailSvc.sentEmails, 1, "One email should be sent")
	assert.Equal(t, input.Email, mockMailSvc.sentEmails[0].to, "Verification email should be sent to correct address")
}