package routes

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/alexedwards/scs/v2"
	"github.com/go-playground/validator/v10"
	"github.com/kytnacode/inventure/internal/auth/session"
	userrepository "github.com/kytnacode/inventure/internal/user/repository"
	"github.com/kytnacode/inventure/pkg/api"
	"github.com/kytnacode/inventure/pkg/logging"
	"github.com/kytnacode/inventure/pkg/passhash"
	"github.com/kytnacode/inventure/pkg/validation"
)

// Routes handle password based authentication routes.
type Routes struct {
	userRepo       *userrepository.Repository
	v              *validator.Validate
	sessionManager *scs.SessionManager
	redirectURL    string
}

// New creates a new [Routes].
func New(
	userRepo *userrepository.Repository,
	sessionManager *scs.SessionManager,
	v *validator.Validate,
	redirectURL string,
) *Routes {
	return &Routes{
		userRepo:       userRepo,
		sessionManager: sessionManager,
		v:              v,
		redirectURL:    redirectURL,
	}
}

// SignUp is the API handler for sign up a new user with password based authentication.
func (ro *Routes) SignUp(w http.ResponseWriter, r *http.Request) {
	api.AcceptJSON(w)

	r = withLogger(r, "auth/routes/Routes.SignUp")

	data, ok := decodeSignUpData(w, r)
	if !ok {
		return
	}

	hash := passhash.Hash([]byte(data.Password))

	model := &userrepository.User{
		Name:         data.Name,
		Email:        data.Email,
		PasswordHash: &hash,
	}

	ok = validateModel(w, r, ro.v, model)
	if !ok {
		return
	}

	id, ok := signUpUser(r.Context(), w, ro.userRepo, model, hash)
	if !ok {
		return
	}

	ro.sessionManager.Put(r.Context(), session.KeySessionData, &session.Session{
		ID: id,
	})

	http.Redirect(w, r, ro.redirectURL, http.StatusTemporaryRedirect)
}

func withLogger(r *http.Request, handler string) *http.Request {
	logger := logging.FromCtx(r.Context())

	logger = logger.With(logging.Handler(handler))

	ctx := logging.WithLogger(r.Context(), logger)

	return r.WithContext(ctx)
}

func decodeSignUpData(w http.ResponseWriter, r *http.Request) (data *SignUpData, ok bool) {
	logger := logging.FromCtx(r.Context())

	data = new(SignUpData)

	if err := json.NewDecoder(r.Body).Decode(data); err != nil {
		logger.Warn("could not decode sign up data", logging.Error(err))

		w.WriteHeader(http.StatusBadRequest)

		err = api.WriteFail(w, map[string]any{
			"body": "invalid json or server error",
		})
		if err != nil {
			logger.Error("could write warn response", logging.Error(err))
		}

		return nil, false
	}

	return data, true
}

func validateModel(
	w http.ResponseWriter,
	r *http.Request,
	v *validator.Validate,
	m *userrepository.User,
) (ok bool) {
	logger := logging.FromCtx(r.Context())

	err := v.Struct(m)
	if err != nil {
		logger.Warn("invalid user model", logging.Error(err))

		if validationErrors, ok := errors.AsType[validator.ValidationErrors](err); ok {
			w.WriteHeader(http.StatusBadRequest)

			err = api.WriteFail(w, validation.MapFromErrors(validationErrors))
			if err != nil {
				logger.Error("could not write fail response", logging.Error(err))
			}

			return false
		}

		w.WriteHeader(http.StatusInternalServerError)

		err = api.WriteError(w, "internal server error", nil, nil)
		if err != nil {
			logger.Error("could not write error response", logging.Error(err))
		}

		return false
	}

	return true
}

func signUpUser(
	ctx context.Context,
	w http.ResponseWriter,
	userRepo *userrepository.Repository,
	model *userrepository.User,
	passHash string,
) (id string, ok bool) {
	logger := logging.FromCtx(ctx)

	id, err := userRepo.SignUpUser(ctx, &userrepository.SignUpUserData{
		Email:        model.Email,
		Name:         model.Name,
		PasswordHash: passHash,
	})
	if err != nil {
		logger.Error("could not sign up user", logging.Error(err))

		w.WriteHeader(http.StatusInternalServerError)

		err = api.WriteError(w, "internal server error", nil, nil)
		if err != nil {
			logger.Error("could not write error response", logging.Error(err))
		}

		return "", false
	}

	return id, true
}
