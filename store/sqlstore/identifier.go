// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package sqlstore

import (
	"fmt"
	"regexp"

	"github.com/twigex/twigex/model"
)

// Every table and column name the server makes fits this, the generated
// t<id> and c<id>, the workspace prefix and the built-in columns alike, and
// so do the names fields were given before they were generated.
var identifierPattern = regexp.MustCompile(`^[A-Za-z0-9_]{1,64}$`)

// quoteIdent returns name ready to put in SQL as a table or column, or an
// error for a name that could be anything else. It never repairs a name: one
// that does not fit was not made by the server.
func quoteIdent(name string) (string, error) {
	if !identifierPattern.MatchString(name) {
		return "", fmt.Errorf("not a table or column name: %q", name)
	}

	return "`" + name + "`", nil
}

func checkSourceTables(sources []model.TaskSource) error {
	for _, s := range sources {
		if err := checkIdents(s.TableName); err != nil {
			return err
		}
	}

	return nil
}

// checkIdents refuses the lot unless every name would quote, for code that
// wraps the names in backticks itself.
func checkIdents(names ...string) error {
	for _, name := range names {
		if _, err := quoteIdent(name); err != nil {
			return err
		}
	}

	return nil
}
