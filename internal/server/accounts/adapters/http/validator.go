package http

import (
	"fmt"

	vldtr "github.com/go-playground/validator/v10"
)

type validator struct {
	v *vldtr.Validate
}

func newValidator(v *vldtr.Validate) *validator {
	return &validator{v: v}
}

func (v *validator) ValidateName(name string) error {
	validatingStruct := struct {
		Name string `validate:"required,excludesall=~!@#$%^&*?<>"`
	}{Name: name}

	if err := v.v.Struct(validatingStruct); err != nil {
		return fmt.Errorf("invalid name")
	}

	return nil
}

func (v *validator) ValidateAccount(name string, login string, password string) error {
	validatingStruct := struct {
		Name     string `validate:"required,min=3,excludesall=~!@#$%^&*?<>"`
		Login    string `validate:"required,min=1"`
		Password string `validate:"required,min=1"`
	}{
		Name:     name,
		Login:    login,
		Password: password,
	}

	if err := v.v.Struct(validatingStruct); err != nil {
		return fmt.Errorf("invalid account parameters")
	}

	return nil
}
