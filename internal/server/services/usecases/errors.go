package usecases

import (
	"errors"
	"fmt"
)

type serviceError struct {
	Code      int
	Component string
	Msg       string
	Err       error
}

func (ce *serviceError) Error() string {
	return fmt.Sprintf("%s: %s", ce.Component, ce.Msg)
}

func (ce *serviceError) Unwrap() error {
	return ce.Err
}

func (ce *serviceError) Is(target error) bool {
	return ce.Error() == target.Error()
}

func newClientError(msg string) error {
	return &serviceError{Code: 400, Component: "ClientError", Msg: msg, Err: nil}
}

func newInternalError(component, msg string, err error) error {
	return &serviceError{Code: 500, Component: component, Msg: msg, Err: err}
}

func parseServiceError(err error) (int, string, error) {
	var ce *serviceError
	if errors.As(err, &ce) {
		return ce.Code, ce.Error(), ce.Err
	}
	return 0, "", nil
}
