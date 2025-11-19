package http

import (
	"fmt"
	"regexp"

	vldtr "github.com/go-playground/validator/v10"
)

type validator struct {
	v *vldtr.Validate
}

func newValidator(v *vldtr.Validate) *validator {
	usernameRegexp := regexp.MustCompile("^[a-zA-Z][a-zA-Z0-9]{3,14}$")

	_ = v.RegisterValidation("username", func(fl vldtr.FieldLevel) bool {
		return usernameRegexp.MatchString(fl.Field().String())
	})
	_ = v.RegisterValidation("password", func(fl vldtr.FieldLevel) bool {
		return len(fl.Field().String()) > 7
	})

	return &validator{v: v}
}

func (v *validator) ValidateUserCreds(username, password string) error {
	errInvalidParams := fmt.Errorf("invalid username or password")

	if err := v.v.Var(username, "username"); err != nil {
		return errInvalidParams
	}
	if err := v.v.Var(password, "password"); err != nil {
		return errInvalidParams
	}

	return nil
}

func (v *validator) ValidatePasswords(passwords ...string) error {
	for _, pass := range passwords {
		if err := v.v.Var(pass, "password"); err != nil {
			return fmt.Errorf("invalid old or new password")
		}
	}
	return nil
}
