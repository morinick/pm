package http

import (
	"errors"
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
	var errs []error

	if err := v.v.Var(username, "username"); err != nil {
		errs = append(errs, fmt.Errorf("invalid username"))
	}
	if err := v.v.Var(password, "password"); err != nil {
		errs = append(errs, fmt.Errorf("invalid password"))
	}

	return errors.Join(errs...)
}
