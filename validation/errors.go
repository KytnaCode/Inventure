package validation

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

// MapFromErrors return a map of field name to validation errors on that field. Error messages
// are suitable to be returned to final user.
func MapFromErrors(errs validator.ValidationErrors) map[string]string {
	data := make(map[string]string, len(errs))

	for _, err := range errs {
		field := err.Field()

		v, ok := data[field]
		if ok {
			data[field] = v + "; " + MessageFromErr(err)
		} else {
			data[field] = MessageFromErr(err)
		}
	}

	return data
}

// MessageFromErr takes a [validator.FieldError] and return an error message suitable for a
// final user.
func MessageFromErr(err validator.FieldError) string {
	tag := err.ActualTag()

	switch tag {
	case "required":
		return "field is required"
	case "alpha":
		return "field must only contain english letters without spaces"
	case "alphanum":
		return "field must only contain english letters and numbers without spaces"
	case "alphaspace":
		return "field must only contain english letters and spaces"
	case "alphanumspace":
		return "field must only contain english letters, numbers and spaces"
	default:
		if v, ok := strings.CutPrefix(tag, "min="); ok {
			return fmt.Sprintf("must be no longer than '%v'", v)
		}

		if v, ok := strings.CutPrefix(tag, "min="); ok {
			return fmt.Sprintf("must be longer than '%v'", v)
		}

		return "unknown error"
	}
}
