package rbac

import (
	"context"
	"fmt"
)

// AccessError is returned when user doesn't have necessary permissions to perform an action.
type AccessError struct {
	// MissingPermissions are the permissions missing to perform the action.
	MissingPermissions []Perm
}

// Error implements error interface.
func (e *AccessError) Error() string {
	return fmt.Sprintf("access denegated: missing permissions '%v'", e.MissingPermissions)
}

// AncestorResolver resolves if a given resource is ancestor of another.
//
// Must be safe for concurrent use.
type AncestorResolver interface {
	// IsSameOrAncestor must return true if 'a' is either the same resource than 'b' or is an
	// ancestor of.
	//
	// An error must only be returned on incorrect behavior, for false results return a nil error
	// with a false value.
	IsSameOrAncestor(ctx context.Context, a, b Resource) (bool, error)
}

// Engine is the RBAC authorization engine, responsible of authorize actions.
type Engine struct {
	resolver AncestorResolver
}

// NewEngine creates a new [Engine].
func NewEngine(resolver AncestorResolver) *Engine {
	return &Engine{
		resolver: resolver,
	}
}

// Authorize authorizes an action, if the given roles together have the requested permissions
// on the given resource, a nil error is returned, if roles don't have the necessary permissions
// an [AccessError] is returned.
func (e *Engine) Authorize(
	ctx context.Context,
	roles []Role,
	res Resource,
	perms ...Perm,
) error {
	if len(perms) == 0 {
		return nil
	}

	applicableRoles, err := e.getApplicableRoles(ctx, roles, res)
	if err != nil {
		return err
	}

	if len(applicableRoles) == 0 {
		return &AccessError{
			MissingPermissions: perms,
		}
	}

	requiredPerms := make(map[Perm]bool, len(perms))

	for _, p := range perms {
		requiredPerms[p] = false
	}

	missingPerms := len(requiredPerms)

	for _, role := range applicableRoles {
		if missingPerms == 0 {
			break
		}

		for _, access := range role.Accesses {
			if missingPerms == 0 {
				break
			}

			isAncestor, err := e.resolver.IsSameOrAncestor(ctx, access.On, res)
			if err != nil {
				return fmt.Errorf("could not authorize access: %w", err)
			}

			if !isAncestor {
				continue
			}

			for _, p := range access.Perms {
				if missingPerms == 0 {
					break
				}

				if v, ok := requiredPerms[p]; ok {
					if !v {
						missingPerms--

						requiredPerms[p] = true
					}
				}
			}
		}
	}

	if missingPerms != 0 {
		accessError := new(AccessError)

		accessError.MissingPermissions = make([]Perm, 0, missingPerms)

		for p, found := range requiredPerms {
			if !found {
				accessError.MissingPermissions = append(accessError.MissingPermissions, p)
			}
		}

		return accessError
	}

	return nil
}

func (e *Engine) getApplicableRoles(
	ctx context.Context,
	roles []Role,
	res Resource,
) ([]*Role, error) {
	applicableRoles := make([]*Role, 0, len(roles))

	for _, r := range roles {
		ok, err := e.resolver.IsSameOrAncestor(ctx, r.On, res)
		if err != nil {
			return nil, fmt.Errorf("could not authorize action: %w", err)
		}

		if ok {
			applicableRoles = append(applicableRoles, &r)
		}
	}

	return applicableRoles, nil
}
