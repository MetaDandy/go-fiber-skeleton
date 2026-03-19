package authentication

import "fmt"

type SignUpPassword struct {
	Email          string `json:"email" validate:"required,email"`
	Password       string `json:"password" validate:"required,min=8"`
	RepeatPassword string `json:"repeat_password" validate:"required,min=8"`
	Ip             string `json:"ip" validate:"required,ip"`
	UserAgent      string `json:"user_agent" validate:"required"`
}

func (s SignUpPassword) Validate() error {
	if s.Password != s.RepeatPassword {
		return fmt.Errorf("password and repeat password do not match")
	}

	return nil
}
