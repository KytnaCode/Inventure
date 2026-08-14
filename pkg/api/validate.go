package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/kytnacode/inventure/pkg/logging"
	"github.com/kytnacode/inventure/pkg/validation"
)

// ValidateModel validates a model using the given validator instance, if model is valid
// returns true, if is invalid returns false and sends an error response. If a validation
// error occurs, a fail response indicating failed rules is written, if an unknown error
// occurs an internal server error is returned.
func ValidateModel[T any](
	ctx context.Context,
	w http.ResponseWriter,
	v *validator.Validate,
	m T,
) (ok bool) {
	logger := logging.FromCtx(ctx)

	err := v.Struct(m)
	if err != nil {
		logger.Warn("invalid model or dto", logging.Error(err))

		if validationErrors, ok := errors.AsType[validator.ValidationErrors](err); ok {
			w.WriteHeader(http.StatusBadRequest)

			err = WriteFail(w, validation.MapFromErrors(validationErrors))
			if err != nil {
				logger.Error("could not write fail response", logging.Error(err))
			}

			return false
		}

		w.WriteHeader(http.StatusInternalServerError)

		err = WriteError(w, "internal server error", nil, nil)
		if err != nil {
			logger.Error("could not write error response", logging.Error(err))
		}

		return false
	}

	return true
}
