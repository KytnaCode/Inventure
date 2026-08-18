package rbac

import (
	"context"
	"database/sql/driver"
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// ErrInvalidScopeString is returned if scope field is malformed in the database.
var ErrInvalidScopeString = errors.New("invalid scope string format")

// ScopeString is a list of scopes. Implements [sql.Scanner] and [sql/driver.Valuer] to marshal
// and unmarshal to and from a string of comma separated elements.
type ScopeString []string

// Scan implements [sql.Scanner].
func (s *ScopeString) Scan(src any) error {
	var source string

	// Some driver returns value as a byte slice and other as a string, try both.
	switch v := src.(type) {
	case string:
		source = v
	case []byte:
		source = string(v)
	default:
		return fmt.Errorf("could not unmarshal scope string: %w", ErrInvalidScopeString)
	}

	*s = ScopeString(strings.Split(source, ","))

	return nil
}

// Value implements [sql/driver.Valuer].
func (s *ScopeString) Value() (driver.Value, error) {
	if s == nil {
		return "", nil
	}

	return strings.Join(*s, ","), nil
}

// CreateRoleData is the data necessary to create a new role.
type CreateRoleData struct {
	// Name is the role display name, must not contain commas.
	Name string `validate:"required,min=3,max=96,alphanum"`

	// Scopes is the list of scopes or permissions the role grants.
	Scopes []string
}

// Model is a role database model.
type Model struct {
	gorm.Model

	// ID is the role primary key.
	ID datatypes.UUID `gorm:"primaryKey"`

	// Name is role's display name, must not contain commas.
	Name string

	// Scopes is the list of scopes the role grants.
	Scopes ScopeString
}

// TableName implements [gorm/schema.Tabler].
func (m *Model) TableName() string {
	return "roles"
}

// Repository handles persistence logic for roles.
type Repository struct {
	table gorm.Interface[Model]
	v     *validator.Validate
}

// NewRepository creates a new [Repository].
func NewRepository(table gorm.Interface[Model], v *validator.Validate) *Repository {
	return &Repository{
		table: table,
		v:     v,
	}
}

// CreateRole creates a new role.
func (r *Repository) CreateRole(ctx context.Context, data *CreateRoleData) (id string, err error) {
	if err := r.v.Struct(data); err != nil {
		return id, fmt.Errorf("role validation error: %w", err)
	}

	scopes := data.Scopes
	if scopes == nil {
		scopes = []string{}
	}

	m := Model{
		ID:     datatypes.NewUUIDv4(),
		Name:   data.Name,
		Scopes: scopes,
	}

	err = r.table.Create(ctx, &m)
	if err != nil {
		return "", fmt.Errorf("could not create role: %w", err)
	}

	return m.ID.String(), nil
}
