package rbac

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/kytnacode/inventure/api"
	"github.com/kytnacode/inventure/internal/web"
	"github.com/kytnacode/inventure/logging"
)

// roleGetter read roles from a database or another store.
//
// MUST be safe for concurrent use.
type roleGetter interface {
	// GetRoles return roles with the matching IDs, should drop non-existing IDs, doesn't require
	// roles slice to match ids order.
	GetRoles(ctx context.Context, ids ...uuid.UUID) ([]Role, error)
}

// AuthData contains necessary authentication data to proceed with authorization.
type AuthData struct {
	// RoleIDs are the ids of the roles the user belongs to.
	RoleIDs []uuid.UUID

	// Resource is the resource the action is performed on.
	Resource Resource
}

// Middleware handles RBAC middleware.
//
// Safe for concurrent use.
type Middleware struct {
	eng     *Engine
	repo    roleGetter
	authExt api.Extractor[*AuthData]
}

// NewMiddleware creates a new [Middleware]. authExt extracts [AuthData] from request's context, if
// returned error is not nil, then authExt MUST return a not nil, fully-populated [AuthData]
// object.
func NewMiddleware(eng *Engine, repo roleGetter, authExt api.Extractor[*AuthData]) *Middleware {
	return &Middleware{
		eng:     eng,
		repo:    repo,
		authExt: authExt,
	}
}

// RequirePerms extracts user's roles and request's resource from request's context and check
// if the user has the given permissions on that resource.
func (m *Middleware) RequirePerms(perms ...Perm) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r = m.withLogger(r, "auth/rbac.Middleware.RequireScopes")

			logger := logging.FromCtx(r.Context())

			data, err := m.authExt(r.Context())
			if err != nil {
				logger.Error(
					"could not extract authentication data from context, missing authentication middleware?",
					logging.Error(err),
				)

				w.WriteHeader(http.StatusInternalServerError)

				err = api.WriteError(w, "internal server error", nil, nil)
				if err != nil {
					logger.Error("could not send error response", logging.Error(err))
				}

				return
			}

			roles, err := m.repo.GetRoles(r.Context(), data.RoleIDs...)
			if err != nil {
				logger.Error("could not get roles from database", logging.Error(err))

				w.WriteHeader(http.StatusInternalServerError)

				err = api.WriteError(w, "internal server error", nil, nil)
				if err != nil {
					logger.Error("could not write error response", logging.Error(err))
				}

				return
			}

			err = m.eng.Authorize(r.Context(), roles, data.Resource, perms...)
			if accessErr, ok := errors.AsType[*AccessError](err); ok {
				logger.Warn(
					"forbidden access attempt",
					logging.Error(accessErr),
					slog.Any("missing_perms", accessErr.MissingPermissions),
				)

				w.WriteHeader(http.StatusForbidden)

				err = api.WriteError(w, "forbidden", web.CodeMissingPerms, nil)
				if err != nil {
					logger.Error("could not write error response", logging.Error(err))
				}

				return
			}

			if err != nil {
				logger.Error("could not authorize request", logging.Error(err))

				w.WriteHeader(http.StatusInternalServerError)

				err = api.WriteError(w, "internal server error", nil, nil)
				if err != nil {
					logger.Error("could not write error response", logging.Error(err))
				}

				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// withLogger set middleware attribute in request's logger.
func (m *Middleware) withLogger(r *http.Request, name string) *http.Request {
	logger := logging.FromCtx(r.Context())

	logger = logger.With(logging.Middleware(name))

	ctx := logging.WithLogger(r.Context(), logger)

	return r.WithContext(ctx)
}
