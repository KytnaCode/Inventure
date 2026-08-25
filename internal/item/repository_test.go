package item_test

import (
	"errors"
	"maps"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/kytnacode/inventure/internal/item"
	"github.com/kytnacode/inventure/validation"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newRepo(t *testing.T) (*item.Repository, *gorm.DB) {
	db, err := gorm.Open(sqlite.Open(":memory:"))
	if err != nil {
		t.Fatalf("could not open sqlite instance: %v", err)
	}

	err = db.AutoMigrate(item.Model{})
	if err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	v := validation.New()

	repo := item.NewRepository(db, v)

	return repo, db
}

func TestRepository_CreateItemShouldRequireName(t *testing.T) {
	t.Parallel()

	repo, _ := newRepo(t)

	id, err := repo.CreateItem(t.Context(), &item.Data{
		// Name: ,
		Desc:    "description",
		Stock:   10,
		Attrs:   map[string]any{},
		PlaceID: uuid.New(),
	})
	if id != "" {
		t.Errorf("expected an empty id: got '%v'", id)
	}

	if err == nil {
		t.Fatal("expected a non-nil error")
	}

	if _, ok := errors.AsType[validator.ValidationErrors](err); !ok {
		t.Fatalf("expected error to be validation errors: got '%v'", err)
	}
}

func TestRepository_CreateItemShouldStoreItem(t *testing.T) {
	t.Parallel()

	repo, db := newRepo(t)

	data := &item.Data{
		Name:  "real name",
		Desc:  "description",
		Stock: 10,
		Attrs: map[string]any{
			"hello": "world",
		},
		PlaceID: uuid.New(),
	}

	id, err := repo.CreateItem(t.Context(), data)
	if err != nil {
		t.Fatalf("expected a nil error: %v", err)
	}

	var item item.Model

	err = db.Where("id = ?", id).
		Take(&item).Error
	if err != nil {
		t.Fatalf("could not get new item: %v", err)
	}

	if item.Name != data.Name {
		t.Errorf("expected name to be '%v': got '%v'", data.Name, item.Name)
	}

	if item.Desc != data.Desc {
		t.Errorf("expected desc to be '%v': got '%v'", data.Desc, item.Desc)
	}

	if item.Stock != data.Stock {
		t.Errorf("expected stock to be '%v': got '%v'", data.Stock, item.Stock)
	}

	if !maps.Equal(item.Attrs, data.Attrs) {
		t.Errorf("expected attrs to be '%v': got '%v'", data.Attrs, item.Attrs)
	}
}
