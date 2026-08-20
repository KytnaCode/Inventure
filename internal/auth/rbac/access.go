package rbac

// Access represents a permission grant on a resource.
type Access struct {
	Perms []Perm
	On    Resource
}
