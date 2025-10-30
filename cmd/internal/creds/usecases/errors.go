package usecases

import (
	"errors"
	"fmt"
)

type credsError struct {
	Code      int
	Component string
	Msg       string
	Err       error
}

func (ce *credsError) Error() string {
	return fmt.Sprintf("%s: %s", ce.Component, ce.Msg)
}

func (ce *credsError) Unwrap() error {
	return ce.Err
}

func (ce *credsError) Is(target error) bool {
	return ce.Error() == target.Error()
}

func newClientError(msg string) error {
	return &credsError{Code: 400, Component: "ClientError", Msg: msg, Err: nil}
}

func newInternalError(component, msg string, err error) error {
	return &credsError{Code: 500, Component: component, Msg: msg, Err: err}
}

func parseCredsError(err error) (int, string, error) {
	var ce *credsError
	if errors.As(err, &ce) {
		return ce.Code, ce.Error(), ce.Err
	}
	return 0, "", nil
}
