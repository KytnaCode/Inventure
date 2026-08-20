package rbac

type Role struct {
	Allow    []Perm
	Forbid   []Perm
	Resource Resource
}
