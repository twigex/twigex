// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"net/http"
	"testing"

	"github.com/twigex/twigex/config"
	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/store"
)

type fakeUsernameRoleStore struct {
	store.RoleStore
}

func (fakeUsernameRoleStore) GetByNames([]string) ([]model.Role, error) {
	roles := []model.Role{{
		Name: model.SystemUserRoleId,
		Permissions: []string{
			model.AdminPermissions.PermissionCreateUsers.Id,
			model.AdminPermissions.PermissionEditUsers.Id,
		},
	}}
	return roles, nil
}

type fakeUsernameUserStore struct {
	store.UserStore
	existing      *model.User
	takenUsername *model.User
	takenEmail    *model.User
	inserted      bool
	edited        bool
}

func (f *fakeUsernameUserStore) Get(string) (*model.User, error) {
	return f.existing, nil
}

func (f *fakeUsernameUserStore) GetByUsername(username string) (*model.User, error) {
	if f.takenUsername != nil && f.takenUsername.Username == username {
		return f.takenUsername, nil
	}
	return nil, nil
}

func (f *fakeUsernameUserStore) GetByEmail(email string) (*model.User, error) {
	if f.takenEmail != nil && f.takenEmail.Email == email {
		return f.takenEmail, nil
	}
	return nil, nil
}

func (f *fakeUsernameUserStore) Create(model.NewUser) (*model.User, error) {
	f.inserted = true
	return &model.User{}, nil
}

func (f *fakeUsernameUserStore) Update(model.UserPatch) error {
	f.edited = true
	return nil
}

func usernameApp() (*App, *fakeUsernameUserStore) {
	u := &fakeUsernameUserStore{existing: &model.User{ID: "u1", Username: "jane"}}
	return &App{
		Store: store.Store{User: u, Roles: fakeUsernameRoleStore{}},
		ConfigStore: config.ConfigStore{
			Config: &model.ServerConfig{
				SqlSettings: model.SqlSettings{QueryTimeout: model.NewInt(25)},
			},
		},
	}, u
}

func newUserNamed(name string) model.NewUser {
	return model.NewUser{
		Username: name,
		Email:    "someone@example.com",
		Name:     "Some",
		LastName: "One",
		Role:     model.SystemUserRoleId,
		Password: "Sufficiently1Long!",
	}
}

func patchNamed(name string) model.UserPatch {
	return model.UserPatch{
		ID:       "u1",
		Username: name,
		Email:    "someone@example.com",
		Name:     "Some",
		LastName: "One",
		Role:     model.SystemUserRoleId,
	}
}

func TestCreateUserRejectsReservedUsername(t *testing.T) {
	for _, name := range []string{"all", "here", "unknown-user", "ALL"} {
		a, s := usernameApp()

		_, appErr := a.CreateUser(model.User{}, newUserNamed(name))
		if appErr == nil {
			t.Fatalf("%q: expected an error", name)
		}
		if appErr.Message != "user.username_reserved" {
			t.Errorf("%q: got %q, want user.username_reserved", name, appErr.Message)
		}
		if appErr.Status != http.StatusBadRequest {
			t.Errorf("%q: got status %d, want 400", name, appErr.Status)
		}
		if s.inserted {
			t.Errorf("%q: reached the store", name)
		}
	}
}

func TestEditUserRejectsReservedUsername(t *testing.T) {
	for _, name := range []string{"all", "here", "unknown-user", "Here"} {
		a, s := usernameApp()

		_, appErr := a.EditUser(model.User{}, patchNamed(name))
		if appErr == nil {
			t.Fatalf("%q: expected an error", name)
		}
		if appErr.Message != "user.username_reserved" {
			t.Errorf("%q: got %q, want user.username_reserved", name, appErr.Message)
		}
		if s.edited {
			t.Errorf("%q: reached the store", name)
		}
	}
}

func TestEditUserAcceptsAnOrdinaryUsername(t *testing.T) {
	a, s := usernameApp()

	if _, appErr := a.EditUser(model.User{}, patchNamed("allan")); appErr != nil {
		t.Fatalf("unexpected error: %v", appErr.Message)
	}
	if !s.edited {
		t.Error("the edit never reached the store")
	}
}

func TestEditUserReportsATakenUsername(t *testing.T) {
	a, s := usernameApp()
	s.takenUsername = &model.User{ID: "u2", Username: "taken"}

	_, appErr := a.EditUser(model.User{}, patchNamed("taken"))
	if appErr == nil {
		t.Fatal("expected an error")
	}
	if appErr.Message != "user.username_taken" || appErr.Status != http.StatusConflict {
		t.Errorf("got %q/%d, want user.username_taken/409", appErr.Message, appErr.Status)
	}
	if s.edited {
		t.Error("reached the store")
	}
}

func TestEditUserReportsATakenEmail(t *testing.T) {
	a, s := usernameApp()
	s.takenEmail = &model.User{ID: "u2", Email: "someone@example.com"}

	_, appErr := a.EditUser(model.User{}, patchNamed("fresh"))
	if appErr == nil {
		t.Fatal("expected an error")
	}
	if appErr.Message != "user.email_taken" || appErr.Status != http.StatusConflict {
		t.Errorf("got %q/%d, want user.email_taken/409", appErr.Message, appErr.Status)
	}
	if s.edited {
		t.Error("reached the store")
	}
}

func TestEditUserAllowsAUserToKeepTheirOwnIdentity(t *testing.T) {
	a, s := usernameApp()
	s.takenUsername = &model.User{ID: "u1", Username: "jane"}
	s.takenEmail = &model.User{ID: "u1", Email: "someone@example.com"}

	if _, appErr := a.EditUser(model.User{}, patchNamed("jane")); appErr != nil {
		t.Fatalf("unexpected error: %v", appErr.Message)
	}
	if !s.edited {
		t.Error("the edit never reached the store")
	}
}

func TestCreateUserReportsATakenUsername(t *testing.T) {
	a, s := usernameApp()
	s.takenUsername = &model.User{ID: "u2", Username: "taken"}

	_, appErr := a.CreateUser(model.User{}, newUserNamed("taken"))
	if appErr == nil {
		t.Fatal("expected an error")
	}
	if appErr.Message != "user.username_taken" || appErr.Status != http.StatusConflict {
		t.Errorf("got %q/%d, want user.username_taken/409", appErr.Message, appErr.Status)
	}
	if s.inserted {
		t.Error("reached the store")
	}
}
