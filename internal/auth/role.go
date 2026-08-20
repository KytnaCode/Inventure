package auth

import "github.com/kytnacode/inventure/internal/auth/rbac"

// Role contains a set of permissions granted to a user.
type Role struct {
	// ID is the role's unique ID.
	ID string

	// Role display name.
	Name string

	// Allow are the permissions granted to the user who the role is given.
	Allow []Scope

	// Forbid are the permissions forbidden for the user who the role is given.
	Forbid []Scope

	// Resource is the resource the role belongs to, role permission for resources in a larger scope
	// will be ignored.
	Resource rbac.Resource
}
