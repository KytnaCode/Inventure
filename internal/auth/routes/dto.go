package routes

// SignUpData is the necessary data for a new user to sign up with password-based authentication.
// For validation rules see [user/repository.User].
type SignUpData struct {
	// Name is user's display name.
	Name string `json:"name" validate:"required"`

	// Email is user's email.
	Email string `json:"email" validate:"required,email"`

	// Password is user's password in clear text.
	Password string `json:"password" validate:"required"`
}

// SignInData is the necessary data for a user to sign in with password-based authentication.
type SignInData struct {
	// Email is user's email.
	Email string `json:"email" validate:"required,email"`

	// Password is user's password in clear text.
	Password string `json:"password" validate:"required"`
}
