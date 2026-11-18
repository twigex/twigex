// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package queryengine

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// postgresDatabase writes SQL for PostgreSQL. As in MySQL, a stored time
// without a time zone is read in the session's time zone, which Connector
// sets to UTC on every connection, and times are compared and cut in UTC.
type postgresDatabase struct{}

func (postgresDatabase) quoteName(name string) string {
	escaped := strings.ReplaceAll(name, `"`, `""`)

	return `"` + escaped + `"`
}

// replaceValuePlaceholders numbers the values in the order they appear:
// $1, $2, …
func (postgresDatabase) replaceValuePlaceholders(sql string) string {
	pieces := strings.Split(sql, valuePlaceholder)

	var numbered strings.Builder
	numbered.WriteString(pieces[0])

	for position, piece := range pieces[1:] {
		numbered.WriteString("$" + strconv.Itoa(position+1))
		numbered.WriteString(piece)
	}

	return numbered.String()
}

func (postgresDatabase) textMatch(expression string, kind textMatchKind, isNegated bool) string {
	matchWord := "ILIKE"
	if isNegated {
		matchWord = "NOT ILIKE"
	}

	pattern := valuePlaceholder
	if kind.allowsTextBefore() {
		pattern = "'%' || " + pattern
	}

	if kind.allowsTextAfter() {
		pattern = pattern + " || '%'"
	}

	// As text, so that numbers and times can be searched as they are on
	// MySQL, which turns them into text itself.
	return "CAST(" + expression + " AS TEXT) " + matchWord + " (" + pattern + ") ESCAPE '!'"
}

func (postgresDatabase) sameOrBothEmpty(left string, right string) string {
	return left + " IS NOT DISTINCT FROM " + right
}

func (postgresDatabase) timeInUTC(expression string) string {
	return "(" + expression + "::timestamptz AT TIME ZONE 'UTC')"
}

// startOfStep uses date_trunc. Its weeks start on Monday.
func (postgresDatabase) startOfStep(utcExpression string, step DateStep) string {
	return "date_trunc('" + string(step) + "', " + utcExpression + ")"
}

func (postgresDatabase) datePlaceholder() string {
	return valuePlaceholder + "::timestamptz"
}

// dateForSQL writes a moment in UTC with the Z that says so:
// 2025-03-01T00:00:00.000Z.
func (postgresDatabase) dateForSQL(moment time.Time) string {
	return moment.UTC().Format("2006-01-02T15:04:05.000Z")
}

func (postgresDatabase) rowLimit(rowCount int) string {
	return fmt.Sprintf("LIMIT %d", rowCount)
}
