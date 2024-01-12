package user_test

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"runtime/debug"
	"testing"
	"time"

	"github.com/aleury/service/business/core/user"
	"github.com/aleury/service/business/data/dbtest"
	"github.com/aleury/service/business/data/order"
	"github.com/aleury/service/foundation/docker"
	"github.com/google/go-cmp/cmp"
)

var c *docker.Container

func TestMain(m *testing.M) {
	var err error
	c, err = dbtest.StartDB()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer dbtest.StopDB(c)

	m.Run()
}

func Test_User(t *testing.T) {
	t.Run("crud", crud)
	t.Run("paging", paging)
}

// =============================================================================

func crud(t *testing.T) {
	seed := func(ctx context.Context, usrCore *user.Core) ([]user.User, error) {
		// Users were already created by seed.sql.
		// TODO: Create test helpers to generate users.
		// See product_test.go for an example in the Ardan Labs repo.
		usrs, err := usrCore.Query(ctx, user.QueryFilter{}, order.By{Field: user.OrderByName, Direction: order.ASC}, 1, 1)
		if err != nil {
			return nil, fmt.Errorf("seeding users: %w", err)
		}
		return usrs, nil
	}

	// --------------------------------------------------------------------------

	test := dbtest.NewTest(t, c)
	defer func() {
		if r := recover(); r != nil {
			t.Log(r)
			t.Error(string(debug.Stack()))
		}
		test.Teardown()
	}()

	api := test.CoreAPIs

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	t.Log("Go seeding...")

	users, err := seed(ctx, api.User)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// -------------------------------------------------------------------------

	saved, err := api.User.QueryByID(ctx, users[0].ID)
	if err != nil {
		t.Fatalf("Should be able to retrieve user by ID: %s.", err)
	}

	if users[0].DateCreated.UnixMilli() != saved.DateCreated.UnixMilli() {
		t.Logf("got:  %v", saved.DateCreated)
		t.Logf("want: %v", users[0].DateCreated)
		t.Logf("diff: %v", saved.DateCreated.Sub(users[0].DateCreated))
		t.Error("Should get back the same date created")
	}

	if users[0].DateUpdated.UnixMilli() != saved.DateUpdated.UnixMilli() {
		t.Logf("got:  %v", saved.DateUpdated)
		t.Logf("want: %v", users[0].DateUpdated)
		t.Logf("diff: %v", saved.DateUpdated.Sub(users[0].DateUpdated))
		t.Error("Should get back the same date updated")
	}

	users[0].DateCreated = time.Time{}
	users[0].DateUpdated = time.Time{}
	saved.DateCreated = time.Time{}
	saved.DateUpdated = time.Time{}

	if diff := cmp.Diff(users[0], saved); diff != "" {
		t.Fatalf("Should get back the same user. diff:\n%s", diff)
	}

	// -------------------------------------------------------------------------

	email, err := mail.ParseAddress("jacob@ardanlabs.com")
	if err != nil {
		t.Fatalf("Should be able to parse email address: %s.", err)
	}

	upd := user.UpdateUser{
		Name:       dbtest.StringPointer("Jacob Walker"),
		Email:      email,
		Department: dbtest.StringPointer("development"),
	}
	if _, err := api.User.Update(ctx, users[0], upd); err != nil {
		t.Fatalf("Should be able to update user: %s.", err)
	}

	saved, err = api.User.QueryByEmail(ctx, *upd.Email)
	if err != nil {
		t.Fatalf("Should be able to retrieve user by email: %s.", err)
	}

	diff := users[0].DateUpdated.Sub(saved.DateUpdated)
	if diff > 0 {
		t.Errorf("Should have a larger DateUpdated: updated=%v original=%v, diff=%v", saved.DateUpdated, users[0].DateUpdated, diff)
	}

	if saved.Name != *upd.Name {
		t.Logf("got:  %v", saved.Name)
		t.Logf("want: %v", *upd.Name)
		t.Error("Should be able to see updates to Name")
	}

	if saved.Email != *upd.Email {
		t.Logf("got:  %v", saved.Email)
		t.Logf("want: %v", *upd.Email)
		t.Error("Should be able to see updates to Email")
	}

	if saved.Department != *upd.Department {
		t.Logf("got:  %v", saved.Department)
		t.Logf("want: %v", *upd.Department)
		t.Error("Should be able to see updates to Department")
	}

	if err := api.User.Delete(ctx, saved); err != nil {
		t.Fatalf("Should be able to delete user: %s.", err)
	}

	_, err = api.User.QueryByID(ctx, saved.ID)
	if !errors.Is(err, user.ErrNotFound) {
		t.Fatalf("Should NOT be able to retrieve user: %s.", err)
	}
}

func paging(t *testing.T) {
	test := dbtest.NewTest(t, c)
	defer func() {
		if r := recover(); r != nil {
			t.Log(r)
			t.Error(string(debug.Stack()))
		}
		test.Teardown()
	}()

	api := test.CoreAPIs

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// -------------------------------------------------------------------------

	name := "User Gopher"
	users1, err := api.User.Query(ctx, user.QueryFilter{Name: &name}, user.DefaultOrderBy, 1, 1)
	if err != nil {
		t.Fatalf("Should be able to retrieve user %q: %s.", name, err)
	}

	n, err := api.User.Count(ctx, user.QueryFilter{Name: &name})
	if err != nil {
		t.Fatalf("Should be able to retrieve user count %q: %s.", name, err)
	}

	if len(users1) != n && users1[0].Name == name {
		t.Errorf("Should have a single user for %q", name)
	}

	name = "Admin Gopher"
	users2, err := api.User.Query(ctx, user.QueryFilter{Name: &name}, user.DefaultOrderBy, 1, 1)
	if err != nil {
		t.Fatalf("Should be able to retrieve user %q: %s.", name, err)
	}

	n, err = api.User.Count(ctx, user.QueryFilter{Name: &name})
	if err != nil {
		t.Fatalf("Should be able to retrieve user count %q: %s.", name, err)
	}

	if len(users2) != n && users2[0].Name == name {
		t.Errorf("Should have a single user for %q", name)
	}

	users3, err := api.User.Query(ctx, user.QueryFilter{}, user.DefaultOrderBy, 1, 2)
	if err != nil {
		t.Fatalf("Should be able to retrieve 2 users for page 1: %s.", err)
	}

	n, err = api.User.Count(ctx, user.QueryFilter{})
	if err != nil {
		t.Fatalf("Should be able to retrieve user count: %s.", err)
	}

	if len(users3) != n {
		t.Logf("got:  %v", len(users3))
		t.Logf("want: %v", n)
		t.Error("Should have 2 users for page 1")
	}

	if users3[0].ID == users3[1].ID {
		t.Logf("User1: %v", users3[0].ID)
		t.Logf("User2: %v", users3[1].ID)
		t.Error("Should have different users")
	}
}
