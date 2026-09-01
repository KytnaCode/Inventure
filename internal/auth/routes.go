package auth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httprate"
	"github.com/go-playground/validator/v10"
	"github.com/kytnacode/inventure/api"
	"github.com/kytnacode/inventure/api/csrf"
	"github.com/kytnacode/inventure/internal/user"
	"github.com/kytnacode/inventure/internal/web"
	"github.com/kytnacode/inventure/logging"
	"github.com/kytnacode/inventure/passhash"
)

// SignUpDto is the necessary data for a new user to sign up with password-based authentication.
// For validation rules see [user/repository.User].
type SignUpDto struct {
	// Name is user's display name.
	Name string `json:"name" validate:"required"`

	// Email is user's email.
	Email string `json:"email" validate:"required,email"`

	// Password is user's password in clear text.
	Password string `json:"password" validate:"required"`
}

// SignInDto is the necessary data for a user to sign in with password-based authentication.
type SignInDto struct {
	// Email is user's email.
	Email string `json:"email" validate:"required,email"`

	// Password is user's password in clear text.
	Password string `json:"password" validate:"required"`
}

// RoutesConfig is the configuration for authentication routes.
type RoutesConfig struct {
	UserRepo              *user.Repository
	LoginAttemptLimit     int
	LoginAttempTimeWindow time.Duration
	SessionManager        *scs.SessionManager
	RequestLimit          int
	TimeWindow            time.Duration
	Validator             *validator.Validate
}

// Routes handle password based authentication routes.
type Routes struct {
	conf         *RoutesConfig
	loginLimiter *httprate.RateLimiter
}

// NewRoutes creates a new [Routes].
func NewRoutes(conf *RoutesConfig) *Routes {
	limiter := new(httprate.RateLimiter)

	if conf.LoginAttemptLimit != 0 && conf.LoginAttempTimeWindow != 0 {
		limiter = httprate.NewRateLimiter(
			conf.LoginAttemptLimit,
			conf.LoginAttempTimeWindow,
		)
	}

	return &Routes{
		conf:         conf,
		loginLimiter: limiter,
	}
}

// SignUp is the API handler for sign up a new user with password based authentication.
func (ro *Routes) SignUp(w http.ResponseWriter, r *http.Request) {
	api.AcceptJSON(w)

	r = ro.withLogger(r, "auth.Routes.SignUp")

	data := new(SignUpDto)

	ok := ro.decodeData(w, r, data)
	if !ok {
		return
	}

	ok = api.ValidateModel(r.Context(), w, ro.conf.Validator, data)
	if !ok {
		return
	}

	limiter := ro.loginLimiter

	if limiter != nil && limiter.RespondOnLimit(w, r, data.Email) {
		return
	}

	hash := passhash.Hash([]byte(data.Password))

	model := &user.Model{
		Name:         data.Name,
		Email:        data.Email,
		PasswordHash: &hash,
	}

	ok = api.ValidateModel(r.Context(), w, ro.conf.Validator, model)
	if !ok {
		return
	}

	userData, ok := ro.signUpUser(r.Context(), w, model, hash)
	if !ok {
		return
	}

	ok = ro.destroyExistingSession(r.Context(), w)
	if !ok {
		return
	}

	ro.conf.SessionManager.Put(r.Context(), KeySessionData, &Session{
		ID:      userData.ID,
		RoleIDs: userData.RoleIDs,
	})
}

// SignIn is the handler for password-based sign in.
func (ro *Routes) SignIn(w http.ResponseWriter, r *http.Request) {
	api.AcceptJSON(w)

	r = ro.withLogger(r, "auth.Routes.SignIn")

	data := new(SignInDto)

	ok := ro.decodeData(w, r, data)
	if !ok {
		return
	}

	ok = api.ValidateModel(r.Context(), w, ro.conf.Validator, data)
	if !ok {
		return
	}

	limiter := ro.loginLimiter

	if limiter != nil && limiter.RespondOnLimit(w, r, data.Email) {
		return
	}

	userData := ro.signInUser(r.Context(), w, &user.SignInData{
		Email:         data.Email,
		ClearPassword: data.Password,
	})
	if userData == nil {
		return
	}

	ok = ro.destroyExistingSession(r.Context(), w)
	if !ok {
		return
	}

	ro.conf.SessionManager.Put(r.Context(), KeySessionData, &Session{
		ID:      userData.ID,
		RoleIDs: userData.RoleIDs,
	})
}

// SetupRouter set ups authentication router. If `RequestLimit` or `TimeWindow` are set to zero
// value in the given [RoutesConfig], then, no rate limit will be applied.
func (ro *Routes) SetupRouter() http.Handler {
	r := chi.NewRouter()

	r.Use(ro.conf.SessionManager.LoadAndSave)
	r.Use(csrf.RequireCSRF(ro.conf.SessionManager))

	if ro.conf.RequestLimit != 0 && ro.conf.TimeWindow != 0 {
		r.Use(httprate.LimitBy(ro.conf.RequestLimit, ro.conf.TimeWindow, api.KeyIP))
	}

	r.Post("/signin", ro.SignIn)
	r.Post("/signup", ro.SignUp)

	return r
}

func (ro *Routes) withLogger(r *http.Request, handler string) *http.Request {
	logger := logging.FromCtx(r.Context())

	logger = logger.With(logging.Handler(handler))

	ctx := logging.WithLogger(r.Context(), logger)

	return r.WithContext(ctx)
}

func (ro *Routes) decodeData(w http.ResponseWriter, r *http.Request, data any) (ok bool) {
	logger := logging.FromCtx(r.Context())

	if err := json.NewDecoder(r.Body).Decode(data); err != nil {
		logger.Warn("could not decode request body", logging.Error(err))

		w.WriteHeader(http.StatusBadRequest)

		err = api.WriteFail(w, map[string]any{
			"body": "invalid json or server error",
		})
		if err != nil {
			logger.Error("could write warn response", logging.Error(err))
		}

		return false
	}

	return true
}

func (ro *Routes) signUpUser(
	ctx context.Context,
	w http.ResponseWriter,
	model *user.Model,
	passHash string,
) (userData *user.Claims, ok bool) {
	logger := logging.FromCtx(ctx)

	userData, err := ro.conf.UserRepo.SignUp(ctx, &user.SignUpData{
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

		return nil, false
	}

	return userData, true
}

func (ro *Routes) signInUser(
	ctx context.Context,
	w http.ResponseWriter,
	data *user.SignInData,
) (userData *user.Claims) {
	logger := logging.FromCtx(ctx)

	userData, err := ro.conf.UserRepo.SignIn(ctx, data)
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			logger.Warn("user not found", logging.Error(err))

			w.WriteHeader(http.StatusNotFound)

			err = api.WriteError(w, "user not found", web.CodeUserNotFound, nil)
			if err != nil {
				logger.Error("could not write fail response", logging.Error(err))
			}

			return nil
		}

		if errors.Is(err, user.ErrNotPasswordAuth) {
			logger.Warn("user doesn't support password based authentication", logging.Error(err))

			w.WriteHeader(http.StatusBadRequest)

			err = api.WriteError(
				w,
				"user doesn't support password authentication",
				web.CodeNoPasswordAuth,
				nil,
			)
			if err != nil {
				logger.Error("could not send error response", logging.Error(err))
			}

			return nil
		}

		if errors.Is(err, user.ErrWrongCredentials) {
			logger.Warn("user attempt to sign in with wrong credentials", logging.Error(err))

			w.WriteHeader(http.StatusUnauthorized)

			err = api.WriteError(
				w,
				"wrong email or password",
				web.CodeWrongCredentials,
				nil,
			)
			if err != nil {
				logger.Error("could not send error response", logging.Error(err))
			}

			return nil
		}

		logger.Error("unknown server error on user sign in", logging.Error(err))

		w.WriteHeader(http.StatusInternalServerError)

		err = api.WriteError(w, "unknown server error", nil, nil)
		if err != nil {
			logger.Error("could not send error response", logging.Error(err))
		}

		return nil
	}

	return userData
}

func (ro *Routes) destroyExistingSession(
	ctx context.Context,
	w http.ResponseWriter,
) (ok bool) {
	logger := logging.FromCtx(ctx)

	err := ro.conf.SessionManager.Destroy(ctx)
	if err != nil {
		logger.Error("could not destroy session", logging.Error(err))

		err = api.WriteError(w, "internal server error", nil, nil)
		if err != nil {
			logger.Error("could not write error response", logging.Error(err))
		}

		return false
	}

	return true
}
