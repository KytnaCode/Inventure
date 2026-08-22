package validation

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

var nameReg = regexp.MustCompile(
	"^[\\p{L}_=\\-\\.,'@]" + // First character must be a letter or in `_=-.,'@`.
		"[\\p{L}\\p{N}\\p{S} ]*$") // Remaining characters can be any letter, numbers, or symbols.

// ValidateResourceName validates if a string is a valid resource name.
func ValidateResourceName(fl validator.FieldLevel) bool {
	return nameReg.Match([]byte(fl.Field().String()))
}
