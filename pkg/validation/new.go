// Package validation contains validation helpers.
package validation

import "github.com/go-playground/validator/v10"

// New creates a new [validator.Validate].
func New() *validator.Validate {
	return validator.New(validator.WithRequiredStructEnabled())
}
