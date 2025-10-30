package http

import (
	"fmt"
	"passman/cmd/internal/creds"

	vldtr "github.com/go-playground/validator/v10"
)

type validator struct {
	v *vldtr.Validate
}

func newValidator(v *vldtr.Validate) *validator {
	return &validator{v: v}
}

func (v *validator) ValidateServiceName(name string) error {
	validatingStruct := struct {
		Name string `validate:"required,excludesall=~!@#$%^&*?<>"`
	}{Name: name}

	if err := v.v.Struct(validatingStruct); err != nil {
		return fmt.Errorf("invalid service name")
	}

	return nil
}

func (v *validator) ValidateCreds(serviceCreds creds.Service) error {
	validatingStruct := struct {
		Name     string `validate:"required,min=3,excludesall=~!@#$%^&*?<>"`
		Login    string `validate:"required,min=1"`
		Password string `validate:"required,min=1"`
	}{
		Name:     serviceCreds.Name,
		Login:    serviceCreds.Login,
		Password: serviceCreds.Password,
	}

	if err := v.v.Struct(validatingStruct); err != nil {
		return fmt.Errorf("invalid creds")
	}

	return nil
}
