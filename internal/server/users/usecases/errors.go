package usecases

import (
	"errors"
	"fmt"
)

type userError struct {
	Code      int
	Component string
	Msg       string
	Err       error
}

func (ue *userError) Error() string {
	return fmt.Sprintf("%s: %s", ue.Component, ue.Msg)
}

func (ue *userError) Unwrap() error {
	return ue.Err
}

func (ue *userError) Is(targer error) bool {
	return ue.Error() == targer.Error()
}

func newClientError(msg string) error {
	return &userError{Code: 400, Component: "ClientError", Msg: msg, Err: nil}
}

func newInternalError(component string, msg string, err error) error {
	return &userError{Code: 500, Component: component, Msg: msg, Err: err}
}

func parseError(err error) (int, string, error) {
	var ue *userError
	if errors.As(err, &ue) {
		return ue.Code, ue.Error(), ue.Err
	}
	return 0, "", nil
}
