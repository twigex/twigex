// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package queryengine

import (
	"errors"
	"fmt"
	"testing"

	"github.com/go-sql-driver/mysql"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestQueryReachedTimeLimit(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"MySQL", &mysql.MySQLError{Number: 3024}, true},
		{"MariaDB", &mysql.MySQLError{Number: 1969}, true},
		{"Postgres", &pgconn.PgError{Code: "57014", Message: "canceling statement due to statement timeout"}, true},
		{"wrapped", fmt.Errorf("run: %w", &mysql.MySQLError{Number: 3024}), true},
		{"Postgres in another language", &pgconn.PgError{Code: "57014", Message: "Abbruch der Anfrage wegen Zeitüberschreitung"}, true},
		{"another MySQL error", &mysql.MySQLError{Number: 1054}, false},
		{"another error", errors.New("broken"), false},
		{"no error", nil, false},
	}

	for _, testCase := range cases {
		got := QueryReachedTimeLimit(testCase.err)
		if got != testCase.want {
			t.Errorf("%s: want %v, got %v", testCase.name, testCase.want, got)
		}
	}
}
