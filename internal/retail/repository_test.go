package retail_test

import (
	"cmp"
	"maps"
	"slices"
	"testing"

	"github.com/google/uuid"
	"github.com/kytnacode/inventure/internal/retail"
	"github.com/kytnacode/inventure/internal/user"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newRepo(t *testing.T) (*retail.Repository, *gorm.DB) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"))
	if err != nil {
		t.Fatalf("could not open test database: %v", err)
	}

	err = db.AutoMigrate(
		retail.TenantModel{},
		user.Model{},
		retail.Model{},
		retail.PlaceModel{},
		retail.ItemModel{},
	)
	if err != nil {
		t.Fatalf("could not run migrations on test database: %v", err)
	}

	repo := retail.NewRepository(db)

	return repo, db
}

func compareUsers(t *testing.T, got, existingUsers []user.Model) {
	t.Helper()

	if len(got) != len(existingUsers) {
		t.Errorf(
			"expected tenant's users list to have %v users: got %v users",
			len(existingUsers),
			len(got),
		)

		return
	}

	sortByName := func(a, b user.Model) int {
		return cmp.Compare(a.Name, b.Name)
	}

	slices.SortFunc(got, sortByName)
	slices.SortFunc(existingUsers, sortByName)

	for i, gotUser := range got {
		expected := existingUsers[i] //nolint:gosec // slices have same length.

		if gotUser.Name != expected.Name {
			t.Errorf("expected user name to be '%v': got '%v'", expected.Name, gotUser.Name)
		}

		if gotUser.Email != expected.Email {
			t.Errorf("expected user email to be '%v': got '%v'", expected.Email, gotUser.Email)
		}

		if gotUser.PasswordHash != expected.PasswordHash {
			t.Errorf(
				"expected user password hash to be '%v': got '%v'",
				expected.PasswordHash,
				gotUser.PasswordHash,
			)
		}
	}
}

func compareRetails(t *testing.T, got []retail.Model, expected []retail.Data) {
	t.Helper()

	if len(got) != len(expected) {
		t.Errorf(
			"expected tenant to have %v retails: got %v",
			len(expected),
			len(got),
		)
	}

	slices.SortFunc(got, func(a, b retail.Model) int {
		return cmp.Compare(a.Name, b.Name)
	})

	slices.SortFunc(expected, func(a, b retail.Data) int {
		return cmp.Compare(a.Name, b.Name)
	})

	for i, gotRetail := range got {
		expected := expected[i] //nolint:gosec // slices have the same length.

		if gotRetail.Name != expected.Name {
			t.Errorf("expected user name to be '%v': got '%v'", expected.Name, gotRetail.Name)
		}
	}
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
		func(a, b retail.ItemData) int {
			return cmp.Compare(a.Name, b.Name)
		},
	)

	itemModelSlice := slices.SortedFunc(
		slices.Values(domain.Items),
		func(a, b retail.Item) int {
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

func TestRepository_CreateFullTenantShouldCreateAnEmptyTenant(t *testing.T) {
	t.Parallel()

	repo, db := newRepo(t)

	data := retail.TenantData{
		Name: "tenant name",
	}

	id, err := repo.CreateFullTenant(t.Context(), &data, nil, nil)
	if err != nil {
		t.Fatalf("expected a nil error: %v", err)
	}

	var got retail.TenantModel

	err = db.WithContext(t.Context()).
		Where("id = ?", id).
		Preload("Retails").
		Preload("Users").
		Take(&got).Error
	if err != nil {
		t.Fatalf("could not get inserted tenant: %v", err)
	}

	if got.Name != data.Name {
		t.Errorf("expected tenant's name to be '%v': got '%v'", data.Name, got.Name)
	}

	compareRetails(t, got.Retails, []retail.Data{})

	compareUsers(t, got.Users, []user.Model{})
}

func TestRepository_CreateFullTenantShouldCreateAnTenantWithExistingUsers(t *testing.T) {
	t.Parallel()

	repo, db := newRepo(t)

	data := &retail.TenantData{
		Name: "tenant name",
	}

	user1 := user.Model{
		ID:           uuid.New(),
		Name:         "Luz Noceda",
		Email:        "luz.noceda@gmail.com",
		PasswordHash: nil,
	}

	user2 := user.Model{
		ID:           uuid.New(),
		Name:         "Amity Blight",
		Email:        "amity.blight@penstagram.com",
		PasswordHash: nil,
	}

	existingUsers := []user.Model{user1, user2}

	err := db.Create(existingUsers).Error
	if err != nil {
		t.Fatalf("could not create test users: %v", err)
	}

	userIDs := []uuid.UUID{
		user1.ID,
		user2.ID,
	}

	id, err := repo.CreateFullTenant(t.Context(), data, nil, userIDs)
	if err != nil {
		t.Fatalf("expected a nil error: %v", err)
	}

	var got retail.TenantModel

	err = db.WithContext(t.Context()).
		Where("id = ?", id).
		Preload("Retails").
		Preload("Users").
		Take(&got).Error
	if err != nil {
		t.Fatalf("could not get inserted tenant: %v", err)
	}

	if got.Name != data.Name {
		t.Errorf("expected tenant's name to be '%v': got '%v'", data.Name, got.Name)
	}

	compareRetails(t, got.Retails, []retail.Data{})

	compareUsers(t, got.Users, existingUsers)
}

func TestRepository_CreateFullTenantShouldCreateAFullFeaturedTenant(t *testing.T) {
	t.Parallel()

	repo, db := newRepo(t)

	data := &retail.TenantData{
		Name: "tenant name",
	}

	user1 := user.Model{
		ID:           uuid.New(),
		Name:         "Luz Noceda",
		Email:        "luz.noceda@gmail.com",
		PasswordHash: nil,
	}

	user2 := user.Model{
		ID:           uuid.New(),
		Name:         "Amity Blight",
		Email:        "amity.blight@penstagram.com",
		PasswordHash: nil,
	}

	existingUsers := []user.Model{user1, user2}

	retails := []retail.Data{
		{
			Name: "retail 1",
			Storage: &retail.PlaceData{
				Name: "Storage",
				Items: []retail.ItemData{
					{
						Name: "item 1",
					},
					{
						Name: "item 2",
					},
				},
			},
		},
		{
			Name: "retail 2",
			Storage: &retail.PlaceData{
				Name: "Storage",
				Items: []retail.ItemData{
					{
						Name: "item 3",
					},
				},
				Children: []retail.PlaceData{
					{
						Name: "place 1",
						Items: []retail.ItemData{
							{
								Name: "item 4",
							},
						},
					},
					{
						Name: "place 2",
					},
				},
			},
		},
	}

	err := db.Create(existingUsers).Error
	if err != nil {
		t.Fatalf("could not create test users: %v", err)
	}

	userIDs := []uuid.UUID{
		user1.ID,
		user2.ID,
	}

	id, err := repo.CreateFullTenant(t.Context(), data, retails, userIDs)
	if err != nil {
		t.Fatalf("expected a nil error: %v", err)
	}

	var got retail.TenantModel

	err = db.WithContext(t.Context()).
		Where("id = ?", id).
		Preload("Retails").
		Preload("Users").
		Take(&got).Error
	if err != nil {
		t.Fatalf("could not get inserted tenant: %v", err)
	}

	if got.Name != data.Name {
		t.Errorf("expected tenant's name to be '%v': got '%v'", data.Name, got.Name)
	}

	compareRetails(t, got.Retails, retails)

	compareUsers(t, got.Users, existingUsers)
}

func TestRepository_CreateFullTenantShouldCreateATenantWithRetails(t *testing.T) {
	t.Parallel()

	repo, db := newRepo(t)

	data := &retail.TenantData{
		Name: "tenant name",
	}

	retails := []retail.Data{
		{
			Name: "retail 1",
			Storage: &retail.PlaceData{
				Name: "Storage",
				Items: []retail.ItemData{
					{
						Name: "item 1",
					},
					{
						Name: "item 2",
					},
				},
			},
		},
		{
			Name: "retail 2",
			Storage: &retail.PlaceData{
				Name: "Storage",
				Items: []retail.ItemData{
					{
						Name: "item 3",
					},
				},
				Children: []retail.PlaceData{
					{
						Name: "place 1",
						Items: []retail.ItemData{
							{
								Name: "item 4",
							},
						},
					},
					{
						Name: "place 2",
					},
				},
			},
		},
	}

	id, err := repo.CreateFullTenant(t.Context(), data, retails, nil)
	if err != nil {
		t.Fatalf("expected a nil error: %v", err)
	}

	var got retail.TenantModel

	err = db.WithContext(t.Context()).
		Where("id = ?", id).
		Preload("Retails").
		Preload("Users").
		Take(&got).Error
	if err != nil {
		t.Fatalf("could not get inserted tenant: %v", err)
	}

	if got.Name != data.Name {
		t.Errorf("expected tenant's name to be '%v': got '%v'", data.Name, got.Name)
	}

	slices.SortFunc(retails, func(a, b retail.Data) int {
		return cmp.Compare(a.Name, b.Name)
	})

	slices.SortFunc(got.Retails, func(a, b retail.Model) int {
		return cmp.Compare(a.Name, b.Name)
	})

	for i, r := range got.Retails {
		domain, err := repo.GetRetailStorage(db, r.ID)
		if err != nil {
			t.Fatalf("could not get retail places: %v", err)
		}

		expected := retails[i]

		comparePlace(t, expected.Storage, domain)
	}

	compareRetails(t, got.Retails, retails)

	compareUsers(t, got.Users, []user.Model{})
}
