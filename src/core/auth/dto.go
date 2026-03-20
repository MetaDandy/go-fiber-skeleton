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
