package auth

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

var (
	// ErrInvalidScopeString is returned if scope field is malformed in the database.
	ErrInvalidScopeString = errors.New("invalid scope string format")

	// ErrRoleNotFound is returned when a given record could not be found.
	ErrRoleNotFound = errors.New("role not found")
)

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
func (s ScopeString) Value() (driver.Value, error) {
	return strings.Join(s, ","), nil
}

// CreateRoleData is the data necessary to create a new role.
type CreateRoleData struct {
	// Name is the role display name, must not contain commas.
	Name string `validate:"required,min=3,max=96,alphanum"`

	// Scopes is the list of scopes or permissions the role grants.
	Scopes []string

	// Resource is the resource the new role will belong to.
	Resource Resource
}

// RoleModel is a role database model.
type RoleModel struct {
	gorm.Model

	// ID is the role primary key.
	ID datatypes.UUID `gorm:"primaryKey"`

	// Name is role's display name, must not contain commas.
	Name string

	// Scopes is the list of scopes the role grants.
	Scopes ScopeString

	// ResourceType is the type of the resource the role belongs to.
	ResourceType string

	// ResourceID is the ID of the resource the role belongs to.
	ResourceID datatypes.UUID
}

// TableName implements [gorm/schema.Tabler].
func (m *RoleModel) TableName() string {
	return "roles"
}

// ToDomain converts a [RoleModel] into a [Role].
func (m *RoleModel) ToDomain() *Role {
	scopes := make([]Scope, 0, len(m.Scopes))

	for _, v := range m.Scopes {
		scopes = append(scopes, Scope(v))
	}

	return &Role{
		ID:       m.ID.String(),
		Name:     m.Name,
		Scopes:   scopes,
		Resource: NewResource(m.ResourceType, m.ResourceID.String()),
	}
}

// Repository handles persistence logic for roles.
type Repository struct {
	table gorm.Interface[RoleModel]
	v     *validator.Validate
}

// NewRepository creates a new [Repository].
func NewRepository(table gorm.Interface[RoleModel], v *validator.Validate) *Repository {
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

	resID, err := toUUID(data.Resource.ID)
	if err != nil {
		return "", fmt.Errorf("expected resource ID to be a valid UUID: %w", err)
	}

	m := RoleModel{
		ID:           datatypes.NewUUIDv4(),
		Name:         data.Name,
		Scopes:       scopes,
		ResourceType: data.Resource.Typ,
		ResourceID:   resID,
	}

	err = r.table.Create(ctx, &m)
	if err != nil {
		return "", fmt.Errorf("could not create role: %w", err)
	}

	return m.ID.String(), nil
}

// GetRoleByID returns a role with the given ID, if not found returns [ErrRoleNotFound].
func (r *Repository) GetRoleByID(ctx context.Context, id string) (*Role, error) {
	m, err := r.table.Where("id = ?", id).Take(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRoleNotFound
		}

		return nil, fmt.Errorf("could not get role by id: %w", err)
	}

	role := m.ToDomain()

	return role, nil
}
