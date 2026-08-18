package routes

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
	"github.com/kytnacode/inventure/internal/auth/csrf"
	"github.com/kytnacode/inventure/internal/auth/session"
	userrepository "github.com/kytnacode/inventure/internal/user/repository"
	"github.com/kytnacode/inventure/internal/web"
	"github.com/kytnacode/inventure/pkg/api"
	"github.com/kytnacode/inventure/pkg/logging"
	"github.com/kytnacode/inventure/pkg/passhash"
)

// Config is the configuration for authentication routes.
type Config struct {
	UserRepo              *userrepository.Repository
	LoginAttemptLimit     int
	LoginAttempTimeWindow time.Duration
	SessionManager        *scs.SessionManager
	RequestLimit          int
	TimeWindow            time.Duration
	Validator             *validator.Validate
	RedirectURL           string
}

// Routes handle password based authentication routes.
type Routes struct {
	conf         *Config
	loginLimiter *httprate.RateLimiter
}

// New creates a new [Routes].
func New(conf *Config) *Routes {
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

	r = withLogger(r, "auth/routes/Routes.SignUp")

	data := new(SignUpData)

	ok := decodeData(w, r, data)
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

	model := &userrepository.User{
		Name:         data.Name,
		Email:        data.Email,
		PasswordHash: &hash,
	}

	ok = api.ValidateModel(r.Context(), w, ro.conf.Validator, model)
	if !ok {
		return
	}

	id, ok := signUpUser(r.Context(), w, ro.conf.UserRepo, model, hash)
	if !ok {
		return
	}

	ok = destroyExistingSession(r.Context(), w, ro.conf.SessionManager)
	if !ok {
		return
	}

	ro.conf.SessionManager.Put(r.Context(), session.KeySessionData, &session.Session{
		ID: id,
	})
}

// SignIn is the handler for password-based sign in.
func (ro *Routes) SignIn(w http.ResponseWriter, r *http.Request) {
	api.AcceptJSON(w)

	r = withLogger(r, "auth/routes/Routes.SignIn")

	data := new(SignInData)

	ok := decodeData(w, r, data)
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

	id := signInUser(r.Context(), w, ro.conf.UserRepo, &userrepository.SignInUserData{
		Email:         data.Email,
		ClearPassword: data.Password,
	})
	if id == "" {
		return
	}

	ok = destroyExistingSession(r.Context(), w, ro.conf.SessionManager)
	if !ok {
		return
	}

	ro.conf.SessionManager.Put(r.Context(), session.KeySessionData, &session.Session{
		ID: id,
	})
}

// SetupRouter set ups authentication router. If `RequestLimit` or `TimeWindow` are set to zero
// value in the given [Config], then, no rate limit will be applied.
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

func withLogger(r *http.Request, handler string) *http.Request {
	logger := logging.FromCtx(r.Context())

	logger = logger.With(logging.Handler(handler))

	ctx := logging.WithLogger(r.Context(), logger)

	return r.WithContext(ctx)
}

func decodeData(w http.ResponseWriter, r *http.Request, data any) (ok bool) {
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

func signInUser(
	ctx context.Context,
	w http.ResponseWriter,
	repo *userrepository.Repository,
	data *userrepository.SignInUserData,
) (id string) {
	logger := logging.FromCtx(ctx)

	id, err := repo.SignInUser(ctx, data)
	if err != nil {
		if errors.Is(err, userrepository.ErrUserNotFound) {
			logger.Warn("user not found", logging.Error(err))

			w.WriteHeader(http.StatusNotFound)

			err = api.WriteError(w, "user not found", web.CodeUserNotFound, nil)
			if err != nil {
				logger.Error("could not write fail response", logging.Error(err))
			}

			return ""
		}

		if errors.Is(err, userrepository.ErrNotPasswordAuth) {
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

			return ""
		}

		if errors.Is(err, userrepository.ErrWrongCredentials) {
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

			return ""
		}

		logger.Error("unknown server error on user sign in", logging.Error(err))

		w.WriteHeader(http.StatusInternalServerError)

		err = api.WriteError(w, "unknown server error", nil, nil)
		if err != nil {
			logger.Error("could not send error response", logging.Error(err))
		}

		return ""
	}

	return id
}

func destroyExistingSession(
	ctx context.Context,
	w http.ResponseWriter,
	m *scs.SessionManager,
) (ok bool) {
	logger := logging.FromCtx(ctx)

	err := m.Destroy(ctx)
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
