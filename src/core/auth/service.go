package authentication

import (
	"context"
	"fmt"
	"log"

	"github.com/MetaDandy/go-fiber-skeleton/constant"
	"github.com/MetaDandy/go-fiber-skeleton/src/enum"
	"github.com/MetaDandy/go-fiber-skeleton/src/model"
	"github.com/MetaDandy/go-fiber-skeleton/src/service/mail"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Service interface {
	UserAuthProviders(email string) ([]string, error)
	SignUpPassword(input SignUpPassword) error
	SendTestEmail(email, name string) error
}

type uRepo interface {
	FindByEmail(email string) (model.User, error)
	ExistsByEmail(email string) error
}

type service struct {
	repo        Repo
	uRepo       uRepo
	mailService mail.EmailService
}

func NewService(repo Repo, uRepo uRepo, mailService mail.EmailService) Service {
	return &service{
		repo:        repo,
		uRepo:       uRepo,
		mailService: mailService,
	}
}

func (s *service) UserAuthProviders(email string) ([]string, error) {
	user, err := s.uRepo.FindByEmail(email)
	if err != nil {
		return []string{}, err
	}

	providers := s.repo.UserAuthProviders(user.ID.String())
	if user.Password != "" {
		providers = append(providers, "password")
	}

	return providers, nil
}

func (s *service) SignUpPassword(input SignUpPassword) error {
	if err := s.uRepo.ExistsByEmail(input.Email); err == nil {
		return fmt.Errorf("%s already exist", input.Email)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	u := model.User{
		ID:            uuid.New(),
		Email:         input.Email,
		Password:      string(hash),
		EmailVerified: false,
		RoleID:        constant.GenericID,
	}

	al := model.AuthLog{
		ID:        uuid.New(),
		Event:     enum.SignUpSuccess,
		UserID:    u.ID,
		Ip:        input.Ip,
		UserAgent: input.UserAgent,
	}

	if err := s.repo.Create(u, al, nil); err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	// Generar token de verificación de email
	token, err := mail.GenerateVerificationToken()
	if err != nil {
		log.Printf("failed to generate verification token: %v", err)
		// No retornar error, el signup ya fue exitoso
		// El usuario puede solicitar resend después
		return nil
	}

	// Enviar email de verificación (agnóstico - funciona con Mailpit o Resend)
	ctx := context.Background()
	if err := s.mailService.SendVerificationEmail(ctx, u.Email, u.Name, token); err != nil {
		log.Printf("failed to send verification email to %s: %v", u.Email, err)
		// No retornar error, el signup ya fue exitoso
		// El usuario puede solicitar resend después
		return nil
	}

	return nil
}

// SendTestEmail envía un email de prueba (para testing durante desarrollo)
func (s *service) SendTestEmail(email, name string) error {
	if email == "" {
		return fmt.Errorf("email is required")
	}

	if name == "" {
		name = "User"
	}

	ctx := context.Background()

	// Enviar email de bienvenida como prueba
	if err := s.mailService.SendWelcome(ctx, email, name); err != nil {
		return fmt.Errorf("failed to send test email: %w", err)
	}

	log.Printf("Test email sent to %s", email)
	return nil
}
