// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/store"
)

// Answers every lookup the way the store reports a missing row: nil, no error.
type missingRowUserStore struct {
	store.UserStore
}

func (missingRowUserStore) Get(string) (*model.User, error) { return nil, nil }

type sessionAuthStore struct {
	store.AuthStore
}

func (sessionAuthStore) GetUserID(*http.Request) (*string, error) {
	id := "deleted-user"
	return &id, nil
}

// A caller that checks only the error dereferences nil and takes the request
// down with it, reachable whenever a row is deleted while an id to it is still
// in flight: a live session whose user was removed, or an id held by a client.
func TestMissingRowIsNotFoundNotAPanic(t *testing.T) {
	a := &App{Store: store.Store{User: missingRowUserStore{}, Auth: sessionAuthStore{}}}

	t.Run("GetCurrentUser reports not found", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("panicked on a user that no longer exists: %v", r)
			}
		}()

		user, appErr := a.GetCurrentUser(httptest.NewRequest(http.MethodGet, "/", nil))
		if user != nil {
			t.Errorf("returned a user for an id with no row: %+v", user)
		}
		if appErr == nil {
			t.Fatal("returned no error for a user that does not exist")
		}
		if appErr.Status != http.StatusNotFound {
			t.Errorf("status is %d, want %d", appErr.Status, http.StatusNotFound)
		}
	})
}
