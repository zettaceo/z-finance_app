package usecase

import (
	"errors"
	"fmt"
)

type ExternalDependencyError struct {
	Provider  string
	Operation string
	Err       error
}

func (e *ExternalDependencyError) Error() string {
	if e == nil {
		return "external dependency error"
	}
	if e.Operation != "" {
		return fmt.Sprintf("external dependency failure (%s:%s): %v", e.Provider, e.Operation, e.Err)
	}
	return fmt.Sprintf("external dependency failure (%s): %v", e.Provider, e.Err)
}

func (e *ExternalDependencyError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func NewExternalDependencyError(provider, operation string, err error) error {
	if err == nil {
		return nil
	}
	return &ExternalDependencyError{Provider: provider, Operation: operation, Err: err}
}

func AsExternalDependencyError(err error) (*ExternalDependencyError, bool) {
	if err == nil {
		return nil, false
	}
	var externalErr *ExternalDependencyError
	if errors.As(err, &externalErr) {
		return externalErr, true
	}
	return nil, false
}
