// © 2019-present nextmv.io inc

// Package errors contains errors contains information about errors returned by nextmv functions.
package errors

import (
	"fmt"
)

// Error is the base interface for all errors returned by nextroute functions.
type Error interface {
	error
}

// InputDataError is returned when there is an input data error.
type InputDataError struct {
	error error
}

// Error returns the error message.
func (e InputDataError) Error() string {
	return e.error.Error()
}

// NewInputDataError creates a new InputDataError.
func NewInputDataError(err error) Error {
	return InputDataError{
		error: fmt.Errorf("input data error: %w", err),
	}
}

// ModelCustomizationError is returned when there is an error in the custom model.
type ModelCustomizationError struct {
	error error
}

// Error returns the error message.
func (e ModelCustomizationError) Error() string {
	return e.error.Error()
}

// NewModelCustomizationError creates a new ModelCustomizationError.
func NewModelCustomizationError(err error) Error {
	return ModelCustomizationError{
		error: fmt.Errorf("input data error: %w", err),
	}
}

// ArgumentMismatchError is returned when an argument contains incorrect data.
type ArgumentMismatchError struct {
	error error
}

// Error returns the error message.
func (e ArgumentMismatchError) Error() string {
	return e.error.Error()
}

// NewArgumentMismatchError creates a new ArgumentMismatchError.
func NewArgumentMismatchError(err error) Error {
	return ArgumentMismatchError{
		error: fmt.Errorf("input data error: %w", err),
	}
}
