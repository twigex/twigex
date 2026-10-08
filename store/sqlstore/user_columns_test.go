// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package sqlstore

// A transposition in the column list puts one field's value in another's
// column rather than failing, and a round trip cannot see it: the insert and
// the scan share the list, so they agree with each other while disagreeing
// with the table. Both directions are checked against names written out here.
// Runs against real MySQL (see testhelper_test.go).

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/twigex/twigex/model"
)

func newUserForColumns() model.NewUser {
	return model.NewUser{
		Email:        "ada@example.com",
		Role:         model.SystemUserRoleId,
		Password:     "hashed-password",
		AuthService:  "ldap",
		Username:     "ada",
		Name:         "Ada",
		LastName:     "Abele",
		StorageLimit: 21474836480,
		ClockDisplay: "24h",
		AuthData:     "auth-data-value",
	}
}

func TestInsertUserWritesEachValueToItsOwnColumn(t *testing.T) {
	db := requireDB(t)
	cleanTables(t, "preferences", "users")

	repo := userRepository{Db: db}
	want := newUserForColumns()

	inserted, err := repo.Create(want)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	var got struct {
		email, role, password, authService string
		username, name, lastname           string
		authData                           string
		storageLimit                       int64
		mfaActive                          bool
		createdAt, deactivatedAt           int64
	}
	err = db.QueryRow(`SELECT email, role, password, auth_service, username,
	                          name, lastname, storage_limit, auth_data, mfa_active,
	                          created_at, deactivated_at
	                   FROM users WHERE id = ?`, inserted.ID).
		Scan(&got.email, &got.role, &got.password, &got.authService, &got.username,
			&got.name, &got.lastname, &got.storageLimit, &got.authData, &got.mfaActive,
			&got.createdAt, &got.deactivatedAt)
	if err != nil {
		t.Fatalf("read back: %v", err)
	}

	for _, c := range []struct{ column, got, want string }{
		{"email", got.email, want.Email},
		{"role", got.role, want.Role},
		{"password", got.password, want.Password},
		{"auth_service", got.authService, want.AuthService},
		{"username", got.username, want.Username},
		{"name", got.name, want.Name},
		{"lastname", got.lastname, want.LastName},
		{"auth_data", got.authData, want.AuthData},
	} {
		if c.got != c.want {
			t.Errorf("column %s holds %q, want %q", c.column, c.got, c.want)
		}
	}

	if got.storageLimit != want.StorageLimit {
		t.Errorf("column storage_limit holds %d, want %d", got.storageLimit, want.StorageLimit)
	}
	if got.mfaActive {
		t.Error("mfa_active is true, want false")
	}
	if got.createdAt == 0 {
		t.Error("created_at is zero")
	}
	if got.deactivatedAt != 0 {
		t.Errorf("deactivated_at is %d, want 0", got.deactivatedAt)
	}
}

// seedDistinctUser writes a row naming its columns, so the values are placed
// independently of the constant the scans use.
func seedDistinctUser(t *testing.T, db *sql.DB) string {
	t.Helper()

	id := model.NewID()
	now := time.Now().Unix()
	_, err := db.Exec(`INSERT INTO users
	    (id, email, role, password, auth_service, username, name, lastname,
	     storage_limit, timezone, mfa_active, mfa_secret,
	     created_at, updated_at, deactivated_at, auth_data)
	    VALUES (?, 'grace@example.com', 'system_user', 'pw-value', 'ldap',
	            'grace', 'Grace', 'Hopper', 53687091200, 'null', 0, 'mfa-value',
	            ?, ?, 0, 'authdata-value')`, id, now, now)
	if err != nil {
		t.Fatalf("seedDistinctUser: %v", err)
	}

	return id
}

func assertScannedIntoTheRightFields(t *testing.T, label string, u *model.User) {
	t.Helper()

	for _, c := range []struct{ field, got, want string }{
		{"Email", u.Email, "grace@example.com"},
		{"Role", u.Role, "system_user"},
		{"Password", u.Password, "pw-value"},
		{"AuthService", u.AuthService, "ldap"},
		{"Username", u.Username, "grace"},
		{"Name", u.Name, "Grace"},
		{"LastName", u.LastName, "Hopper"},
		{"MfaSecret", u.MfaSecret, "mfa-value"},
		{"AuthData", u.AuthData, "authdata-value"},
		{"Photo", u.Photo, "Photo-test"},
	} {
		if c.got != c.want {
			t.Errorf("%s: %s = %q, want %q", label, c.field, c.got, c.want)
		}
	}

	if u.StorageLimit != 53687091200 {
		t.Errorf("%s: StorageLimit = %d, want 53687091200", label, u.StorageLimit)
	}
}

func TestUserScansReadEachColumnIntoItsOwnField(t *testing.T) {
	db := requireDB(t)
	cleanTables(t, "preferences", "users")

	repo := userRepository{Db: db}
	id := seedDistinctUser(t, db)

	if _, err := db.Exec(`INSERT INTO user_photos (user_id, photo_id) VALUES (?, ?)`, id, "Photo-test"); err != nil {
		t.Fatalf("insert photo record: %v", err)
	}

	photo, err := NewUserPhotoRepository(db).Get(context.Background(), id)
	if err != nil || photo == nil || photo.PhotoID != "Photo-test" || photo.StorageID != "" {
		t.Fatalf("legacy photo record = %#v, error = %v", photo, err)
	}

	byID, err := repo.Get(id)
	if err != nil || byID == nil {
		t.Fatalf("Get: %v", err)
	}
	assertScannedIntoTheRightFields(t, "Get", byID)

	byEmail, err := repo.GetByEmail("grace@example.com")
	if err != nil || byEmail == nil {
		t.Fatalf("GetByEmail: %v", err)
	}
	assertScannedIntoTheRightFields(t, "GetByEmail", byEmail)

	byUsername, err := repo.GetByUsername("grace")
	if err != nil || byUsername == nil {
		t.Fatalf("GetByUsername: %v", err)
	}
	assertScannedIntoTheRightFields(t, "GetByUsername", byUsername)

	all, err := repo.GetAll()
	if err != nil {
		t.Fatalf("GetAll: %v", err)
	}
	if len(all) != 1 {
		t.Fatalf("GetAll returned %d users, want 1", len(all))
	}
	assertScannedIntoTheRightFields(t, "GetAll", &all[0])

	byIDs, err := repo.GetByIDs([]string{id})
	if err != nil {
		t.Fatalf("GetByIDs: %v", err)
	}
	if len(byIDs) != 1 {
		t.Fatalf("GetByIDs returned %d users, want 1", len(byIDs))
	}
	assertScannedIntoTheRightFields(t, "GetByIDs", &byIDs[0])
}
