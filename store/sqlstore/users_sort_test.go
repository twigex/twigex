// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package sqlstore

import (
	"context"
	"testing"

	"github.com/twigex/twigex/model"
)

func TestUsersPagedFollowTheRequestedSort(t *testing.T) {
	db := requireDB(t)
	cleanTables(t, "users")
	repo := &userRepository{Db: db}
	ctx := context.Background()

	seedNamedUser(t, db, "Anna", "Ozols", "c@example.com", "anna", 0)
	seedNamedUser(t, db, "Maris", "Berzins", "a@example.com", "maris", 0)
	seedNamedUser(t, db, "Liga", "Kalnina", "b@example.com", "liga", 1)

	emails := func(sort model.Sort) []string {
		t.Helper()

		users, err := repo.GetAllPaged(ctx, "", sort, 10, 0, true)
		if err != nil {
			t.Fatalf("GetAllPaged(%+v): %v", sort, err)
		}

		out := make([]string, len(users))
		for i, u := range users {
			out[i] = u.Email
		}

		return out
	}

	cases := []struct {
		sort model.Sort
		want []string
	}{
		{model.Sort{}, []string{"c@example.com", "b@example.com", "a@example.com"}},
		{model.Sort{Key: "email", Desc: true}, []string{"c@example.com", "b@example.com", "a@example.com"}},
		{model.Sort{Key: "email"}, []string{"a@example.com", "b@example.com", "c@example.com"}},
		{model.Sort{Key: "status", Desc: true}, []string{"b@example.com"}},
	}

	for _, c := range cases {
		got := emails(c.sort)

		for i, want := range c.want {
			if i >= len(got) || got[i] != want {
				t.Errorf("sort %+v: got %v, want it to start with %v", c.sort, got, c.want)
				break
			}
		}
	}
}
