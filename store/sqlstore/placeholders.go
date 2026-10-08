// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package sqlstore

import "strings"

// Three gives "?,?,?".
func sqlPlaceholders(n int) string {
	if n <= 0 {
		return ""
	}

	return strings.Repeat("?,", n-1) + "?"
}

// Three rows of two columns give "(?,?),(?,?),(?,?)".
func rowPlaceholders(rows, columns int) string {
	if rows <= 0 || columns <= 0 {
		return ""
	}

	row := "(" + strings.Repeat("?,", columns-1) + "?)"

	return strings.TrimSuffix(strings.Repeat(row+",", rows), ",")
}
