package rbac

// Role contains a set of permissions granted to a user.
type Role struct {
	// ID is the role's unique ID.
	ID string

	// Role display name.
	Name string

	// Scopes are the scopes or permissions granted to the user who the role is granted.
	Scopes []Scope

	// Resource is the resource the role belongs to, role permission for resources in a larger scope
	// will be ignored.
	Resource Resource
}
