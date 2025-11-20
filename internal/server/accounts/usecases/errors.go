package usecases

import (
	"errors"
	"fmt"
)

type accountsError struct {
	Code      int
	Component string
	Msg       string
	Err       error
}

func (ce *accountsError) Error() string {
	return fmt.Sprintf("%s: %s", ce.Component, ce.Msg)
}

func (ce *accountsError) Unwrap() error {
	return ce.Err
}

func (ce *accountsError) Is(target error) bool {
	return ce.Error() == target.Error()
}

func newClientError(msg string) error {
	return &accountsError{Code: 400, Component: "ClientError", Msg: msg, Err: nil}
}

func newInternalError(component, msg string, err error) error {
	return &accountsError{Code: 500, Component: component, Msg: msg, Err: err}
}

func parseAccountsError(err error) (int, string, error) {
	var ce *accountsError
	if errors.As(err, &ce) {
		return ce.Code, ce.Error(), ce.Err
	}
	return 0, "", nil
}
