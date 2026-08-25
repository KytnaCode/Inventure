package retail

import (
	"github.com/google/uuid"
	"github.com/kytnacode/inventure/internal/item"
	"github.com/kytnacode/inventure/internal/user"
)

// Model is the database representation of a [Retail].
type Model struct {
	// ID is the retail's unique ID.
	ID uuid.UUID

	// Name is the retail's display name.
	Name string

	// Users are the users the retail has.
	Users []user.Model `gorm:"polymorphic:Resource;"`

	// Storage is the root place where items are stored.
	Storage PlaceModel `gorm:"polymorphic:Parent;"`
}

// PlaceModel is the database representation of a [Place].
type PlaceModel struct {
	// ParentType is the type of the parent resource.
	ParentType string

	// ParentID is the ID of the parent resource.
	ParentID uuid.UUID

	// Name is place's name, must be unique between siblings.
	Name string

	// Children are the children places.
	Children []PlaceModel `gorm:"polymorphic:Parent;"`

	// Items are the items that reside directly on this place.
	Items []item.Model
}
