package sqltypes

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"strings"
)

var (
	// ErrInvalidList is returned if the database returns an invalid list representation.
	ErrInvalidList = errors.New("sql value is an invalid list value")

	// ErrInvalidElementType is returned if list element type could not be converted to a SQL value.
	ErrInvalidElementType = errors.New("invalid element type")
)

const listSep = ","

// List encodes a list of strings as a string of comma separated values.
type List []string

// Scan implements [sql.Scanner].
func (t *List) Scan(src any) error {
	var source string

	switch v := src.(type) {
	case string:
		source = v
	case []byte:
		source = string(v)
	default:
		return ErrInvalidList
	}

	*t = strings.Split(source, listSep)

	return nil
}

// Value implements [driver.Valuer].
func (t List) Value() (driver.Value, error) {
	result, err := driver.String.ConvertValue(strings.Join(t, listSep))
	if err != nil {
		return nil, fmt.Errorf("invalid list value: %w", err)
	}

	return result, nil
}
