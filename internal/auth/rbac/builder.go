package rbac

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ErrMissingRoleData is returned by [Builder] when required role data is not set with the given
// setter method.
var ErrMissingRoleData = errors.New("missing role data")

type roleCreator interface {
	// CreateRole creates a role with the given data and accesses and return its IDs.
	CreateRole(ctx context.Context, data *RoleData, accesses ...AccessData) (id uuid.UUID, err error)
}

// Builder expose a simple and elegant API to create roles. MUST not be reused. Not safe for
// concurrent use.
//
//	// Usage:
//	id, err := RoleBuilder().WithRepo(repo).
//	  Name("role-name").BelongsTo(ownerResource).
//	  On(resource1).Allow("item-add", "item-del"). // grants permissions on resource1.
//	  Remove(resource1). // remove granted permissions on resource1.
//	  On(resource2).Allow("item-read"). // grants permissions on resource2.
//	  // Belove is same as: On(resource3).Allow("item-del", "item-del").
//	  On(resource3).Allow("item-add").Allow("item-del"). // Multiple allow calls are merged
//	  Build(ctx)
type Builder struct {
	// db is used if [Builder.WithDB] is called, used for low-level operations.
	db *gorm.DB

	// dao is used if [Builder.WithDB] is called, used for low-level operations.
	dao *DAO

	// repo is used if [Builder.WithRepo] is called, used for high-level operations to create
	// the role on [Builder.Build] call.
	repo roleCreator

	// roleOn is the final role resource, using a pointer to check nil resource easier.
	roleOn *Resource

	// role data is the final role data used to create the role on [Builder.Build] call, role.On
	// must be set to `roleOn` before creation.
	role RoleData

	// accesses is a map from resource string representation `[typ: uuid]` to an access pointer,
	// is used to avoid duplicating entries on call [Builder.On] for the same resource multiple
	// times, also allows removing accesses with [Builder.Remove].
	accesses map[string]*AccessData

	// currAccess is the access being edited by calls to [Builder.Allow], [Builder.On] changes
	// current access.
	currAccess *AccessData

	// temporaryPermissions contains permissions added with [Builder.Allow], to avoid computing
	// set union on multiple calls to [Builder.Allow] for the same resource, perms are temporary
	// stored here, on resource change with [Builder.On] or in call to [Builder.Build] set union
	// is computed and permissions are added.
	temporaryPermissions []Perm
}

// RoleBuilder creates a new [Builder].
func RoleBuilder() *Builder {
	return &Builder{
		dao:                  NewDAO(),
		accesses:             make(map[string]*AccessData, 8),
		temporaryPermissions: make([]Perm, 0, 10),
	}
}

// WithRepo uses a high-level repository for operations.
func (b *Builder) WithRepo(repo roleCreator) *Builder {
	b.repo = repo
	b.db = nil

	return b
}

// WithDB uses a database for persistence, used ofr low-level operations.
func (b *Builder) WithDB(db *gorm.DB) *Builder {
	b.repo = nil
	b.db = db

	return b
}

// On is used before [Builder.Allow] to specify on which resource allow permissions
//
//	builder.On(resource).Allow("perm1", "perm2")
func (b *Builder) On(res Resource) *Builder {
	b.appendStalePerms()

	key := res.String()

	v, ok := b.accesses[key]
	if !ok {
		v = &AccessData{
			On: res,
		}

		b.accesses[key] = v
	}

	b.currAccess = v

	return b
}

// Name set role's name, must be called at least once.
func (b *Builder) Name(name string) *Builder {
	b.role.Name = name

	return b
}

// BelongsTo set role's owner, must be called at least once.
func (b *Builder) BelongsTo(res Resource) *Builder {
	on := res

	b.roleOn = &on

	return b
}

// Allow add permissions to role, must be preceded by a call to [Builder.On]. Can be called
// multiple time to specify permissions to different resources, if called two or more times
// on the same resource permissions will be joined.
//
//	builder.
//	  // `item-add` and `item-del` will be grant on `resource1`.
//	  On(resource1).Allow("item-add", "item-del").
//	  // `item-add`, `item-del` and `item-read` will be granted on `resource2`.
//	  On(resource2).Allow("item-add").Allow("item-del", "item-read")
func (b *Builder) Allow(perms ...Perm) *Builder {
	if b.currAccess == nil {
		return b
	}

	b.temporaryPermissions = append(b.temporaryPermissions, perms...)

	return b
}

// Remove removes granted permissions for a given resource.
//
//	builder.
//	  On(resource1).Allow("item-read").
//	  Remove(resource1) // No permissions will be granted on `resource1`.
func (b *Builder) Remove(res Resource) *Builder {
	if b.currAccess.On.Equal(res) {
		b.currAccess = nil
	}

	delete(b.accesses, res.String())

	return b
}

// Build inserts builded role into the database and returns its ID. If [Builder.Name] or
// [Builder.BelongsTo] are not called, an [ErrMissingRoleData] will be returned.
func (b *Builder) Build(ctx context.Context) (uuid.UUID, error) {
	b.appendStalePerms()

	if b.role.Name == "" {
		return uuid.UUID{}, fmt.Errorf(
			"missing role name, use Builder.Name() to set one: %w",
			ErrMissingRoleData,
		)
	}

	if b.roleOn == nil {
		return uuid.UUID{}, fmt.Errorf(
			"missing role resource, use Builder.BelongsTo() to set one: %w",
			ErrMissingRoleData,
		)
	}

	b.role.On = *b.roleOn

	accesses := make([]AccessData, 0, len(b.accesses))

	for _, v := range b.accesses {
		if v == nil {
			continue
		}

		accesses = append(accesses, *v)
	}

	var (
		id  uuid.UUID
		err error
	)

	switch {
	case b.repo != nil:
		id, err = b.repo.CreateRole(ctx, &b.role, accesses...)
	case b.db != nil && b.dao != nil:
		id, err = b.dao.CreateRole(b.db, &b.role, accesses...)
	default:
		return uuid.UUID{}, fmt.Errorf(
			"missing persitence backend, use Builder.WithRepo() or Builder.WithDB() to set one: %w",
			ErrMissingRoleData,
		)
	}

	if err != nil {
		return uuid.UUID{}, fmt.Errorf("could not create role: %w", err)
	}

	return id, nil
}

// appendStalePerms checks if there are permissions stored in `temporaryPermissions` and if
// there are, appends them to current resource and then resets `temporaryPermissions` slice.
func (b *Builder) appendStalePerms() {
	if b.currAccess != nil && len(b.temporaryPermissions) != 0 {
		b.currAccess.Perms = union(b.currAccess.Perms, b.temporaryPermissions)
	}

	b.temporaryPermissions = b.temporaryPermissions[:0]
}
