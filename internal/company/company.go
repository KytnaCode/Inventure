package company

// Company is a company that can control multiple namespaces.
type Company struct {
	Name string `validate:"required,min=3,max=80,alphanumspace"`
}
