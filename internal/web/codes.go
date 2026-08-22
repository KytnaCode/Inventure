package web

import "github.com/kytnacode/inventure/api"

// Code is an API's error code.
type Code = api.Code

// Error codes.
var (
	// CodeUserNotFound is returned when trying to login as a user who could be found.
	CodeUserNotFound = api.NewCode(1)

	// CodeNoPasswordAuth is returned when trying to login using a password for a user without
	// password-based authentication enabled.
	CodeNoPasswordAuth = api.NewCode(2)

	// CodeWrongCredentials is returned when trying to login as a user but password or email are
	// wrong.
	CodeWrongCredentials = api.NewCode(3)

	// CodeMissingCSRFTokenSession is returned if session doesn't contain a valid CSRF token.
	CodeMissingCSRFTokenSession = api.NewCode(4)

	// CodeWrongCSRFToken is returned if session's token mismatch with request header's.
	CodeWrongCSRFToken = api.NewCode(5)
)
