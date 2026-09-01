package user_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/kytnacode/inventure/internal/user"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newDAO(t *testing.T) (*user.DAO, *gorm.DB) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"))
	if err != nil {
		t.Fatalf("could not open test database: %v", err)
	}

	err = db.AutoMigrate(user.Model{})
	if err != nil {
		t.Fatalf("could not run test database migrations: %v", err)
	}

	return user.NewDAO(), db
}

func TestDAO_ExistsShouldReturnFalseIfNotExists(t *testing.T) {
	t.Parallel()

	dao, db := newDAO(t)

	found, err := dao.Exists(db.WithContext(t.Context()), uuid.New())
	if err != nil {
		t.Fatalf("expected a nil error: %v", err)
	}

	if found {
		t.Fatalf("expected found to be false: got '%v'", found)
	}
}

func TestDAO_ExistsShouldReturnTrueIfExists(t *testing.T) {
	t.Parallel()

	dao, db := newDAO(t)

	userID := uuid.New()

	err := db.WithContext(t.Context()).Create(&user.Model{
		ID:           userID,
		Name:         "username",
		Email:        "real@email.com",
		PasswordHash: nil,
	}).Error
	if err != nil {
		t.Fatalf("expected a nil error: %v", err)
	}

	found, err := dao.Exists(db.WithContext(t.Context()), userID)
	if err != nil {
		t.Fatalf("expected a nil error: %v", err)
	}

	if !found {
		t.Fatalf("expected found to be true: got '%v'", found)
	}
}
