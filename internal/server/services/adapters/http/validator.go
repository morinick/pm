package http

import (
	"errors"
	"fmt"

	vldtr "github.com/go-playground/validator/v10"
)

type validator struct {
	v *vldtr.Validate
}

func newValidator(v *vldtr.Validate) *validator {
	return &validator{v: v}
}

func (v *validator) ValidateServiceNames(names ...string) error {
	var errs []error

	validatingStruct := struct {
		Name string `validate:"required,excludesall=~!@#$%^&*?<>"`
	}{}

	for _, name := range names {
		validatingStruct.Name = name
		if err := v.v.Struct(validatingStruct); err != nil {
			errs = append(errs, fmt.Errorf("%s is invalid", name))
		}
	}

	return errors.Join(errs...)
}
