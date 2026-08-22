package user

// User is the domain model for a user.
type User struct {
	// Name is user display name.
	Name string `validate:"required,min=3,max=80,resourcename"`
}
