// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package sqlstore

import (
	"testing"

	"github.com/twigex/twigex/model"
)

func TestOrderBy(t *testing.T) {
	columns := map[string][]string{
		"name":  {"name", "lastname"},
		"email": {"email"},
	}

	cases := []struct {
		sort model.Sort
		want string
	}{
		{model.Sort{Key: "email"}, " ORDER BY email ASC, id ASC"},
		{model.Sort{Key: "name", Desc: true}, " ORDER BY name DESC, lastname DESC, id DESC"},
		{model.Sort{}, " ORDER BY name ASC, lastname ASC, id ASC"},
		{model.Sort{Key: "email; DROP TABLE users"}, " ORDER BY name ASC, lastname ASC, id ASC"},
	}

	for _, c := range cases {
		if got := orderBy(c.sort, columns, "name", "id"); got != c.want {
			t.Errorf("orderBy(%+v) = %q, want %q", c.sort, got, c.want)
		}
	}
}
