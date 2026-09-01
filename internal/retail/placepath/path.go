// Package placepath implement utilities for working with place paths.
package placepath

import (
	"path"
	"strings"
)

// Separator is place path separator.
const Separator = '/'

// TrimLeftPath remove the left most place in a path.
//
//	TrimLeftPath("/id1/id2/id3") // '/id2/id3'
//	TrimLeftPath(TrimLeftPath("/id1/id2/id3")) // '/id3'
//	TrimLeftPath("/id1") // '/'
//	TrimLeftPath("/") // '/'
func TrimLeftPath(s string) string {
	if len(s) == 0 {
		return s
	}

	if s[0] == Separator {
		s = s[1:]
	}

	i := strings.IndexRune(s, Separator)
	if i == -1 {
		return "/"
	}

	return s[i:]
}

// CutPrefix cuts a given path prefix and returns the new path.
//
// If len(s) == 0 then an empty string is returned, if prefix is not present s is returned as is.
// If prefix matches all s then '/' will be returned, not an empty string, resulting string will
// start with '/' and not end with '/' (unless is the root path '/').
//
//	CutPrefix("/prefix/id1", "/prefix") // '/id1'
//	CutPrefix("/id1", "/") // '/id1'
func CutPrefix(s, prefix string) string {
	if len(s) == 0 {
		return s
	}

	// If not start by '/', prepend it.
	if s[0] != Separator {
		s = "/" + s
	}

	// If end by '/' remove the leading slash.
	if s[len(s)-1] == Separator {
		s = s[:len(s)-1]
	}

	// If prefix is empty return `s` as is.
	if len(prefix) == 0 {
		return s
	}

	// If prefix doesn't start by separator prepend it, this is necessary for call to
	// strings.CutPrefix work.
	if prefix[0] != Separator {
		prefix = string(Separator) + prefix
	}

	// If prefix end by slash remove it, this is necessary to ensure result string start with a
	// slash.
	if prefix[len(prefix)-1] == Separator {
		prefix = prefix[:len(prefix)-1]
	}

	s, _ = strings.CutPrefix(s, prefix)

	// If result will be empty return the root path.
	if s == "" {
		return string(Separator)
	}

	return s
}

// Components return components of a path, if the path is root ('/') then an empty slice will be
// returned.
func Components(pathStr string) []string {
	if len(pathStr) == 0 || pathStr == string(Separator) {
		return []string{}
	}

	// If start with slash remove it to avoid an empty component at the start.
	if pathStr[0] == Separator {
		pathStr = pathStr[1:]
	}

	// Same as above but for the end.
	if pathStr[len(pathStr)-1] == Separator {
		pathStr = pathStr[:len(pathStr)-1]
	}

	return strings.Split(pathStr, string(Separator))
}

// Join joins place's paths.
func Join(elems ...string) string {
	return path.Join(elems...)
}
