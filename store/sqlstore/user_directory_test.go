// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package sqlstore

import (
	"context"
	"database/sql"
	"testing"
)

func seedAuthUser(t *testing.T, db *sql.DB, username, email, authService, authData string, deactivatedAt int64) string {
	t.Helper()
	return insertUser(t, db, userRow{
		name:          "First",
		lastname:      "Last",
		email:         email,
		username:      username,
		authService:   authService,
		authData:      authData,
		deactivatedAt: deactivatedAt,
	})
}

// The sync needs deactivated accounts too, otherwise a returning user could never
// be reactivated.
func TestFindUsersByAuthServiceIncludesDeactivated(t *testing.T) {
	db := requireDB(t)
	cleanTables(t, "users")

	repo := userRepository{Db: db}

	active := seedAuthUser(t, db, "jane", "jane@example.com", "ldap", "uuid-a", 0)
	gone := seedAuthUser(t, db, "bob", "bob@example.com", "ldap", "uuid-b", 1700000000)
	seedAuthUser(t, db, "admin", "admin@example.com", "", "", 0)
	seedAuthUser(t, db, "carol", "carol@example.com", "google-workspace", "sub-1", 0)

	users, err := repo.GetByAuthService(context.Background(), "ldap")
	if err != nil {
		t.Fatalf("GetByAuthService: %v", err)
	}
	if len(users) != 2 {
		t.Fatalf("got %d users, want the 2 ldap ones", len(users))
	}

	found := map[string]int64{}
	for _, u := range users {
		found[u.ID] = u.DeactivatedAt
		if u.AuthService != "ldap" {
			t.Errorf("%s has auth_service %q", u.Username, u.AuthService)
		}
	}
	if found[active] != 0 {
		t.Error("the active user should come back active")
	}
	if found[gone] == 0 {
		t.Error("the deactivated user should keep its timestamp")
	}
}

func TestFindUsersByAuthServiceCarriesTheFieldsTheDiffNeeds(t *testing.T) {
	db := requireDB(t)
	cleanTables(t, "users")

	repo := userRepository{Db: db}
	seedAuthUser(t, db, "jane", "jane@example.com", "ldap", "uuid-a", 0)

	users, err := repo.GetByAuthService(context.Background(), "ldap")
	if err != nil {
		t.Fatalf("GetByAuthService: %v", err)
	}

	u := users[0]
	if u.AuthData != "uuid-a" {
		t.Errorf("auth_data = %q, the diff matches on it", u.AuthData)
	}
	if u.Username != "jane" || u.Email != "jane@example.com" {
		t.Errorf("got %q / %q", u.Username, u.Email)
	}
}

func TestUpdateUserProfileWritesEveryField(t *testing.T) {
	db := requireDB(t)
	cleanTables(t, "users")

	repo := userRepository{Db: db}
	id := seedAuthUser(t, db, "jane", "jane@example.com", "ldap", "uuid-a", 0)

	if _, err := repo.UpdateProfile(id, "jsmith@example.com", "Jane", "Smith"); err != nil {
		t.Fatalf("UpdateProfile: %v", err)
	}

	u, err := repo.Get(id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if u.Email != "jsmith@example.com" || u.Name != "Jane" || u.LastName != "Smith" {
		t.Errorf("got %q %q %q", u.Email, u.Name, u.LastName)
	}
	if u.Username != "jane" {
		t.Errorf("username is %q, want it left to UpdateUsername", u.Username)
	}
}

// The rename is its own statement, so a username the unique index refuses does
// not hold back the fields that can be applied. These columns share no
// invariant, and an unresolvable name collision would otherwise leave a stale
// email behind on every run.
func TestProfileFieldsApplyDespiteUsernameCollision(t *testing.T) {
	db := requireDB(t)
	cleanTables(t, "users")

	repo := userRepository{Db: db}
	jane := seedAuthUser(t, db, "jane", "jane@example.com", "ldap", "uuid-a", 0)
	seedAuthUser(t, db, "taken", "taken@example.com", "ldap", "uuid-b", 0)

	if _, err := repo.UpdateUsername(jane, "taken"); err == nil {
		t.Fatal("renaming onto an existing username must fail")
	}

	if _, err := repo.UpdateProfile(jane, "new@example.com", "New", "Name"); err != nil {
		t.Fatalf("UpdateProfile: %v", err)
	}

	u, err := repo.Get(jane)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if u.Email != "new@example.com" || u.Name != "New" || u.LastName != "Name" {
		t.Errorf("got %q %q %q, want the directory's values applied", u.Email, u.Name, u.LastName)
	}
	if u.Username != "jane" {
		t.Errorf("username is %q, want the refused rename to have left it alone", u.Username)
	}
}
