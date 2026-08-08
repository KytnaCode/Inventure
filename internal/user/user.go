package user

type User struct {
	// TODO: Allow Unicode alphabetic characters.
	// Name is user display name.
	Name string `validate:"required,min=3,max=80,alphanumspace"`
}
