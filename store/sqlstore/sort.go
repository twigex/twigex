// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package sqlstore

import (
	"strings"

	"github.com/twigex/twigex/model"
)

// orderBy only ever writes column names from the columns map, so a sort key
// sent by a client can choose between them but never reach the query itself.
func orderBy(sort model.Sort, columns map[string][]string, fallback, tieBreaker string) string {
	exprs, ok := columns[sort.Key]
	if !ok {
		exprs = columns[fallback]
	}

	dir := " ASC"
	if sort.Desc {
		dir = " DESC"
	}

	parts := make([]string, 0, len(exprs)+1)
	for _, expr := range exprs {
		parts = append(parts, expr+dir)
	}

	parts = append(parts, tieBreaker+dir)

	return " ORDER BY " + strings.Join(parts, ", ")
}
