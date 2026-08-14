package web

import "github.com/kytnacode/inventure/pkg/api"

type Code = api.Code

var (
	CodeUserNotFound     = api.NewCode(1)
	CodeNoPasswordAuth   = api.NewCode(2)
	CodeWrongCredentials = api.NewCode(3)
)
