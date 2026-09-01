package retail_test

import (
	"cmp"
	"errors"
	"maps"
	"slices"
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

func comparePlace(t *testing.T, data *retail.PlaceData, domain *retail.Place) {
	if data.Name != domain.Name {
		t.Errorf("expected name to be '%v': got '%v'", data.Name, domain.Name)
	}

	if len(data.Items) != len(domain.Items) {
		t.Fatalf("different items length: expeceted '%v' got '%v'", len(data.Items), len(domain.Items))
	}

	itemDataSlice := slices.SortedFunc(
		slices.Values(data.Items),
		func(a, b item.Data) int {
			return cmp.Compare(a.Name, b.Name)
		},
	)

	itemModelSlice := slices.SortedFunc(
		slices.Values(domain.Items),
		func(a, b item.Item) int {
			return cmp.Compare(a.Name, b.Name)
		},
	)

	for i, itemData := range itemDataSlice {
		itemModel := itemModelSlice[i]

		if itemData.Name != itemModel.Name {
			t.Errorf("expected item name to be '%v': got '%v'", itemData.Name, itemModel.Name)
		}

		if itemData.Desc != itemModel.Desc {
			t.Errorf("expected item desc to be '%v': got '%v'", itemData.Desc, itemModel.Desc)
		}

		if itemData.Stock != itemModel.Stock {
			t.Errorf("expected item stock to be '%v': got '%v'", itemData.Stock, itemModel.Stock)
		}

		if !maps.Equal(itemData.Attrs, itemModel.Attrs) {
			t.Errorf("expected item attrs to be '%v': got '%v'", itemData.Attrs, itemModel.Attrs)
		}
	}

	if len(data.Children) != len(domain.Children) {
		t.Fatalf("expected '%v' children: got '%v'", len(data.Children), len(domain.Children))
	}

	dataChildren := slices.SortedFunc(
		slices.Values(data.Children),
		func(a, b retail.PlaceData) int {
			return cmp.Compare(a.Name, b.Name)
		},
	)

	modelChildren := slices.SortedFunc(
		slices.Values(domain.Children),
		func(a, b retail.Place) int {
			return cmp.Compare(a.Name, b.Name)
		},
	)

	for i, child := range dataChildren {
		modelChild := modelChildren[i]

		comparePlace(t, &child, &modelChild)
	}
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

	got, err := dao.GetRetailPlaces(db, retailID)
	if err != nil {
		t.Fatalf("could not get places: %v", err)
	}

	domain := retail.PlaceFromModel(got)

	comparePlace(t, &data, domain)
}
