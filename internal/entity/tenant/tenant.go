package tenant

import (
	"os/user"

	"github.com/google/uuid"
	"github.com/kytnacode/inventure/internal/entity/retail"
)

// Tenant represents a company that can manage multiple retails.
type Tenant struct {
	// ID is tenant's unique ID.
	ID uuid.UUID

	// Name is tenant's display name.
	Name string

	// Users are tenant-scoped users for this tenant.
	Users []user.User

	// Retails are retails owned by this tenant.
	Retails []retail.Retail
}
