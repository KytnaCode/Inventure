package user

// User is the domain model for a user.
type User struct {
	// TODO: Allow Unicode alphabetic characters.
	// Name is user display name.
	Name string `validate:"required,min=3,max=80,alphanumspace"`
}
