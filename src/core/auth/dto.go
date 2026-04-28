package authentication

import (
	"github.com/MetaDandy/go-fiber-skeleton/api_error"
)

type SignUpPassword struct {
	Email          string `json:"email" validate:"required,email"`
	Password       string `json:"password" validate:"required,min=8"`
	RepeatPassword string `json:"repeat_password" validate:"required,min=8"`
	Ip             string `json:"ip" validate:"required,ip"`
	UserAgent      string `json:"user_agent" validate:"required"`
}

func (s SignUpPassword) Validate() error {
	if len(s.Password) < 8 {
		return api_error.BadRequest("password must be at least 8 characters")
	}

	if s.Password != s.RepeatPassword {
		return api_error.BadRequest("password and repeat password do not match")
	}

	return nil
}

type ForgotPassword struct {
	Email string `json:"email" validate:"required,email"`
	Ip    string `json:"ip" validate:"required,ip"`
}

func (f ForgotPassword) Validate() error {
	if f.Email == "" {
		return api_error.BadRequest("email is required")
	}
	if f.Ip == "" {
		return api_error.BadRequest("ip is required")
	}
	return nil
}

type ResetPassword struct {
	Token          string `json:"token" validate:"required"`
	NewPassword    string `json:"new_password" validate:"required,min=8"`
	RepeatPassword string `json:"repeat_password" validate:"required,min=8"`
	Ip             string `json:"ip" validate:"required,ip"`
	UserAgent      string `json:"user_agent" validate:"required"`
}

func (r ResetPassword) Validate() error {
	if r.Token == "" {
		return api_error.BadRequest("token is required")
	}
	if len(r.NewPassword) < 8 {
		return api_error.BadRequest("new password must be at least 8 characters")
	}
	if r.NewPassword != r.RepeatPassword {
		return api_error.BadRequest("passwords do not match")
	}
	return nil
}

type ChangePassword struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=8"`
	RepeatPassword  string `json:"repeat_password" validate:"required,min=8"`
}

func (c ChangePassword) Validate() error {
	if c.CurrentPassword == "" {
		return api_error.BadRequest("current password is required")
	}
	if len(c.NewPassword) < 8 {
		return api_error.BadRequest("new password must be at least 8 characters")
	}
	if c.NewPassword != c.RepeatPassword {
		return api_error.BadRequest("passwords do not match")
	}
	if c.CurrentPassword == c.NewPassword {
		return api_error.BadRequest("new password must be different from current password")
	}
	return nil
}

type LoginPassword struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
	Ip       string `json:"ip"`
	UserAgent string `json:"user_agent"`
}

func (l LoginPassword) Validate() error {
	if l.Email == "" {
		return api_error.BadRequest("email is required")
	}
	if l.Password == "" {
		return api_error.BadRequest("password is required")
	}
	return nil
}

// OAuth DTOs

type OAuthUserInfo struct {
	ID    string // provider_user_id
	Email string
	Name  string
	Image string
}

type OAuthCallback struct {
	Code  string `json:"code" validate:"required"`
	State string `json:"state" validate:"required"`
	Ip    string
	UserAgent string
}

// OAuthCallbackInternal es el DTO interno después de procesar el callback
type OAuthCallbackInternal struct {
	Provider  string
	State     string // State original para validación y consumo atómico
	UserInfo  OAuthUserInfo
	Ip        string
	UserAgent string
}
