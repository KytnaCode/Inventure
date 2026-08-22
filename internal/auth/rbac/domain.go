package rbac

// Role is a role for RBAC.
type Role struct {
	// ID is role's unique ID.
	ID string

	// Name is role's unique name.
	Name string

	// Accesses are a list of permission grants over a resource.
	Accesses []Access

	// On is the resource the role belongs to, role permission are not applicable to resources
	// that is not the role's resource or a child of it.
	On Resource
}

// Access is a permission grant for a role on a given resource.
type Access struct {
	// ID is access' unique ID.
	ID string

	// Role is the role which the permission is granted on.
	Role *Role

	// Perms are the permissions granted.
	Perms []Perm

	// On are the resource on which the access is granted.
	On Resource
}

// Resource is a resource that can contain roles and be subject of permission grants.
type Resource struct {
	// Parent is the resource parent, nil for a root resource.
	Parent *Resource

	// Typ is the resource type.
	Typ string

	// ID is the resource unique ID.
	ID string
}

// Perm is a permission that can be granted.
type Perm string
