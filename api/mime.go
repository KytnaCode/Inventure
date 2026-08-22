package api

import "net/http"

// Header names.
const (
	HeaderContentType = "Content-Type"
	HeaderAccept      = "Accept"
)

// MimeTypeJSON is JSON's mime type.
const MimeTypeJSON = "application/json"

// AcceptJSON sets [HeaderAccept] header to [MimeTypeJSON].
func AcceptJSON(w http.ResponseWriter) {
	w.Header().Set(HeaderAccept, MimeTypeJSON)
}

// ContentJSON sets [HeaderContentType] to [MimeTypeJSON].
func ContentJSON(w http.ResponseWriter) {
	w.Header().Set(HeaderContentType, MimeTypeJSON)
}
