package web

import "github.com/kytnacode/inventure/pkg/api"

type Code = api.Code

var (
	CodeUserNotFound     = api.NewCode(1)
	CodeNoPasswordAuth   = api.NewCode(2)
	CodeWrongCredentials = api.NewCode(3)

	// CodeMissingCSRFTokenSession is returned if session doesn't contain a valid CSRF token.
	CodeMissingCSRFTokenSession = api.NewCode(4)

	// CodeWrongCSRFToken is returned if session's token mismatch with request header's.
	CodeWrongCSRFToken = api.NewCode(5)
)
