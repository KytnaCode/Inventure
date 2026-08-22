// Package validation contains validation helpers.
package validation

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

// New creates a new [validator.Validate].
func New() *validator.Validate {
	v := validator.New(validator.WithRequiredStructEnabled())

	err := v.RegisterValidation("resourcename", ValidateResourceName)
	if err != nil {
		panic(fmt.Sprintf("unexpected error: '%v'", err))
	}

	return v
}
