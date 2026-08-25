package retail_test

import (
	"cmp"
	"maps"
	"slices"
	"testing"

	"github.com/kytnacode/inventure/internal/item"
	"github.com/kytnacode/inventure/internal/retail"
	"github.com/kytnacode/inventure/internal/user"
	"github.com/kytnacode/inventure/validation"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newRepo(t *testing.T) (*retail.Repository, *gorm.DB) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"))
	if err != nil {
		t.Fatalf("could not open test database instance: %v", err)
	}

	err = db.AutoMigrate(retail.Model{}, retail.PlaceModel{}, item.Model{}, user.Model{})
	if err != nil {
		t.Fatalf("could not run test database migrations: %v", err)
	}

	v := validation.New()

	repo := retail.NewRepository(db, v)

	return repo, db
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

func TestRepository_CreateRetailShouldInsertRetailData(t *testing.T) {
	t.Parallel()

	repo, db := newRepo(t)

	data := &retail.Data{
		Name: "read retail name",
	}

	id, err := repo.CreateRetail(t.Context(), data, &retail.PlaceData{})
	if err != nil {
		t.Fatalf("expected a nil error: %v", err)
	}

	var m retail.Model

	err = db.WithContext(t.Context()).Where("id = ?", id).Take(&m).Error
	if err != nil {
		t.Fatalf("could not get inserted retail: %v", err)
	}

	if m.Name != data.Name {
		t.Fatalf("expected retail name to be '%v': got '%v'", data.Name, m.Name)
	}
}

func TestRepository_CreateRetailShouldRequireName(t *testing.T) {
	t.Parallel()

	repo, _ := newRepo(t)

	data := &retail.Data{}

	id, err := repo.CreateRetail(t.Context(), data, &retail.PlaceData{})
	if id != "" {
		t.Errorf("expected an empty ID: got '%v'", id)
	}

	if err == nil {
		t.Fatal("expected a non-nil error")
	}
}

func TestRepository_CreateRetailShouldCreatePlacesAndItems(t *testing.T) {
	t.Parallel()

	repo, db := newRepo(t)

	data := &retail.Data{
		Name: "real retail name",
	}

	storage := &retail.PlaceData{
		Name: "place 1",
		Items: []item.Data{
			{
				Name:  "real item 1",
				Desc:  "desc",
				Stock: 10,
			},
			{
				Name:  "real item 2",
				Stock: 2,
			},
		},
		Children: []retail.PlaceData{
			{
				Name: "place 2",
				Items: []item.Data{
					{
						Name: "real item 3",
					},
				},
			},
			{
				Name: "place 3",
				Children: []retail.PlaceData{
					{
						Name: "place 4",
						Items: []item.Data{
							{
								Name:  "real item 4",
								Stock: 57,
							},
						},
					},
				},
			},
		},
	}

	id, err := repo.CreateRetail(t.Context(), data, storage)
	if err != nil {
		t.Fatalf("could not create retail: %v", err)
	}

	var m []retail.PlaceModel

	err = db.WithContext(t.Context()).
		Where("retail_id = ?", id).
		Where("path LIKE '/%'").
		Preload("Items").
		Find(&m).Error
	if err != nil {
		t.Fatalf("could not get inserted retail: %v", err)
	}

	p := retail.PlaceFromModel(m)

	comparePlace(t, storage, p)
}
