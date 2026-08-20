package rbac

import (
	"context"
	"net/http"

	"github.com/kytnacode/inventure/pkg/api"
	"github.com/kytnacode/inventure/pkg/logging"
)

// Engine extracts necessary authentication data from request and authorize actions for a given
// role, resource, and permissions.
type Engine interface {
	RoleFrom(r *http.Request) (*Role, error)
	ResourceIDFrom(r *http.Request) (string, error)
	Authorize(ctx context.Context, role *Role, res Resource, perms ...Perm) (bool, error)
}

// Middleware implements RBAC middleware.
type Middleware struct {
	eng Engine
}

// NewMiddleware creates a new [Middleware].
func NewMiddleware(eng Engine) *Middleware {
	return &Middleware{
		eng: eng,
	}
}

// RequirePerms checks if user has required permissions to perform an action on a resource.
func (m *Middleware) RequirePerms(
	resourceTyp string,
	perms ...Perm,
) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r = withLogger(r, "auth/rbac.Middleware")

			logger := logging.FromCtx(r.Context())

			role, err := m.eng.RoleFrom(r)
			if err != nil {
				logger.Error("could not get role from request", logging.Error(err))

				w.WriteHeader(http.StatusUnauthorized)

				err = api.WriteError(w, "unauthorized", nil, nil)
				if err != nil {
					logger.Error("could not write error response", logging.Error(err))
				}

				return
			}

			resID, err := m.eng.ResourceIDFrom(r)
			if err != nil {
				logger.Error("could not get resource ID", logging.Error(err))

				w.WriteHeader(http.StatusInternalServerError)

				err = api.WriteError(w, "internal server error", nil, nil)
				if err != nil {
					logger.Error("could not write error response", logging.Error(err))
				}

				return
			}

			res := NewResource(resourceTyp, resID)

			ok, err := m.eng.Authorize(r.Context(), role, res, perms...)
			if err != nil {
				logger.Error("could not authorize request", logging.Error(err))

				w.WriteHeader(http.StatusInternalServerError)

				err = api.WriteError(w, "internal server error", nil, nil)
				if err != nil {
					logger.Error("could not write error response", logging.Error(err))
				}

				return
			}

			if !ok {
				logger.Warn("forbidden access")

				w.WriteHeader(http.StatusForbidden)

				err = api.WriteError(w, "forbidden", nil, nil)
				if err != nil {
					logger.Error("could not write error response", logging.Error(err))
				}

				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func withLogger(r *http.Request, middleware string) *http.Request {
	logger := logging.FromCtx(r.Context())

	logger = logger.With(logging.Middleware(middleware))

	ctx := logging.WithLogger(r.Context(), logger)

	return r.WithContext(ctx)
}
