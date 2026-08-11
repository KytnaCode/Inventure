package session

import "encoding/gob"

const KeySessionData = "session-data"

type Session struct {
	ID string
}

func init() {
	gob.Register(&Session{})
}
