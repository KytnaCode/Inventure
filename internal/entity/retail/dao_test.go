package retail_test

import (
	"errors"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/kytnacode/inventure/internal/entity/item"
	"github.com/kytnacode/inventure/internal/entity/retail"
	"github.com/kytnacode/inventure/validation"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newDAO(t *testing.T) (*retail.DAO, *gorm.DB) {
	db, err := gorm.Open(sqlite.Open(":memory:"))
	if err != nil {
		t.Fatalf("could not open test database instance: %v", err)
	}

	err = db.AutoMigrate(&retail.Model{}, &retail.PlaceModel{}, &item.Model{})
	if err != nil {
		t.Fatalf("could not run test database migrations: %v", err)
	}

	return retail.NewDAO(validation.New()), db
}

func TestDAO_CreateRetailShouldRequireRetailName(t *testing.T) {
	t.Parallel()

	dao, db := newDAO(t)

	data := retail.Data{
		// Name:     "real retail",
		TenantID: uuid.New(),
	}

	id, err := dao.CreateRetail(db, &data)
	if id.String() == "" {
		t.Errorf("expected an empty id: got '%v'", id.String())
	}

	if err == nil {
		t.Fatalf("expected a non-nil error: %v", err)
	}

	if _, ok := errors.AsType[validator.ValidationErrors](err); !ok {
		t.Errorf("expected error to be validation errors: got '%v'", err)
	}
}

func TestDAO_CreateRetailShouldCreateRetail(t *testing.T) {
	t.Parallel()

	dao, db := newDAO(t)

	data := retail.Data{
		Name:     "real retail",
		TenantID: uuid.New(),
	}

	id, err := dao.CreateRetail(db, &data)
	if err != nil {
		t.Fatalf("expected a nil error: %v", err)
	}

	var got retail.Model

	err = db.WithContext(t.Context()).Where("id = ?", id).Take(&got).Error
	if err != nil {
		t.Fatalf("could not get inserted retail: %v", err)
	}

	if data.Name != got.Name {
		t.Errorf("expected name to be '%v': got '%v'", data.Name, got.Name)
	}

	if data.TenantID != got.TenantID {
		t.Errorf("expected tenant id to be '%v': got '%v'", data.TenantID, got.TenantID)
	}
}

func TestDAO_CreatePlaceShouldCreatePlace(t *testing.T) {
	t.Parallel()

	dao, db := newDAO(t)

	retailData := retail.Data{
		Name: "abc",
	}

	retailID, err := dao.CreateRetail(db, &retailData)
	if err != nil {
		t.Fatalf("could not create place parent retail: %v", err)
	}

	data := retail.PlaceData{
		Name: "place 1",
		Items: []item.Data{
			{
				Name:  "hello",
				Desc:  "world",
				Stock: 2,
			},
		},
		Children: []retail.PlaceData{
			{
				Name: "place 2",
				Items: []item.Data{
					{
						Name:  "item 2",
						Stock: 10,
					},
				},
			},
			{
				Name: "Place 3",
			},
		},
	}

	err = dao.CreatePlace(db, retailID, &data, "/")
	if err != nil {
		t.Fatalf("expected a nil error: %v", err)
	}

	var got []retail.PlaceModel

	err = db.WithContext(t.Context()).
		Where("retail_id = ?", retailID).
		Where("path LIKE '/%'").
		Preload("Items").
		Find(&got).Error
	if err != nil {
		t.Fatalf("could not get inserted places: %v", err)
	}

	domain := retail.PlaceFromModel(got)

	comparePlace(t, &data, domain)
}
