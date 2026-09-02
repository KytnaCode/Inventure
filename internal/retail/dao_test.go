package retail_test

import (
	"cmp"
	"slices"
	"testing"

	"github.com/google/uuid"
	"github.com/kytnacode/inventure/internal/retail"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newDAO(t *testing.T) (*retail.DAO, *gorm.DB) {
	db, err := gorm.Open(sqlite.Open(":memory:"))
	if err != nil {
		t.Fatalf("could not open test database instance: %v", err)
	}

	err = db.AutoMigrate(&retail.Model{}, &retail.PlaceModel{}, &retail.ItemModel{})
	if err != nil {
		t.Fatalf("could not run test database migrations: %v", err)
	}

	return retail.NewDAO(), db
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

func TestDAO_CreateRetailListShouldCreateRetails(t *testing.T) {
	t.Parallel()

	dao, db := newDAO(t)

	data := []retail.Data{
		{
			Name:     "real retail",
			TenantID: uuid.New(),
		},
		{
			Name:     "other real retail",
			TenantID: uuid.New(),
		},
	}

	ids, err := dao.CreateRetailsList(db, data)
	if err != nil {
		t.Fatalf("expected a nil error: %v", err)
	}

	var gotRetails []retail.Model

	err = db.WithContext(t.Context()).Where("id IN ?", ids).Find(&gotRetails).Error
	if err != nil {
		t.Fatalf("could not get inserted retail: %v", err)
	}

	slices.SortFunc(gotRetails, func(a, b retail.Model) int {
		return cmp.Compare(a.Name, b.Name)
	})

	slices.SortFunc(data, func(a, b retail.Data) int {
		return cmp.Compare(a.Name, b.Name)
	})

	if len(gotRetails) != len(data) {
		t.Fatalf("expected '%v' retails: got '%v'", len(data), len(gotRetails))
	}

	for i, got := range gotRetails {
		expected := data[i]

		if expected.Name != got.Name {
			t.Errorf("expected name to be '%v': got '%v'", expected.Name, got.Name)
		}

		if expected.TenantID != got.TenantID {
			t.Errorf("expected tenant id to be '%v': got '%v'", expected.TenantID, got.TenantID)
		}
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
		Items: []retail.ItemData{
			{
				Name:  "hello",
				Desc:  "world",
				Stock: 2,
			},
		},
		Children: []retail.PlaceData{
			{
				Name: "place 2",
				Items: []retail.ItemData{
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

	got, err := dao.GetRetailPlaces(db, retailID)
	if err != nil {
		t.Fatalf("could not get places: %v", err)
	}

	domain := retail.PlaceFromModel(got)

	comparePlace(t, &data, domain)
}
