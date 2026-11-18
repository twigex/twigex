// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package queryengine

import (
	"strings"
	"unicode"
)

// longestName is the longest name Postgres keeps whole, in bytes. MySQL
// keeps 64 for table names and 256 for column names.
const longestName = 63

// nameAsKept is a name as the database keeps it: Postgres cuts longer names
// off, so two names alike in their first 63 bytes are the same name there.
func nameAsKept(name string) string {
	if len(name) <= longestName {
		return name
	}

	return name[:longestName]
}

// isReservedName says whether a dataset is named like a table the engine
// makes inside its own queries (combined_query.go).
func isReservedName(name string) bool {
	if name == allGroupsName || name == distinctRowsName {
		return true
	}

	number, isPart := strings.CutPrefix(name, "part_")
	if !isPart || number == "" {
		return false
	}

	for _, character := range number {
		if character < '0' || character > '9' {
			return false
		}
	}

	return true
}

// isPlainName reports whether text is a bare column or table name, as opposed
// to an SQL expression.
func isPlainName(text string) bool {
	if text == "" {
		return false
	}

	for position, character := range text {
		isLetter := unicode.IsLetter(character) || character == '_'
		isDigit := unicode.IsDigit(character)

		if position == 0 && isDigit {
			return false
		}

		if !isLetter && !isDigit {
			return false
		}
	}

	return true
}

// quotedTableName quotes a table name, which may include its database or
// schema: "sales.orders" becomes `sales`.`orders` in MySQL and "sales"."orders"
// in Postgres.
func quotedTableName(syntax sqlSyntax, table string) (string, error) {
	parts := strings.Split(table, ".")

	quotedParts := make([]string, 0, len(parts))
	for _, part := range parts {
		if !isPlainName(part) {
			return "", newRequestError(ErrorInvalidTableName, table)
		}

		quotedParts = append(quotedParts, syntax.quoteName(part))
	}

	return strings.Join(quotedParts, "."), nil
}
