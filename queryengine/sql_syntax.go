// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package queryengine

import (
	"strings"
	"time"
)

// sqlSyntax is how one kind of database writes the few things that differ
// between them. Everything else in a query is written the same for every
// database, including GROUP BY with column positions (GROUP BY 1, 2).
type sqlSyntax interface {
	// quoteName puts a table, column or other name in the database's quotes.
	quoteName(name string) string

	// replaceValuePlaceholders turns every valuePlaceholder in finished SQL into the
	// database's own mark for a value.
	replaceValuePlaceholders(sql string) string

	// textMatch compares a field with one searched text, ignoring upper and
	// lower case. The kind says where in the value the text must be.
	textMatch(expression string, kind textMatchKind, isNegated bool) string

	// sameOrBothEmpty is true when two values are equal or both empty.
	sameOrBothEmpty(left string, right string) string

	// timeInUTC gives a stored time in UTC.
	timeInUTC(expression string) string

	// startOfStep cuts a time, already in UTC, down to the start of its
	// second, minute, hour, day, week, month, quarter or year.
	startOfStep(utcExpression string, step DateStep) string

	// datePlaceholder is where a date value goes in a comparison with a stored
	// time, and dateForSQL writes the date the way that place expects it.
	datePlaceholder() string
	dateForSQL(moment time.Time) string

	// rowLimit is the line that ends a query, keeping only its first rows.
	rowLimit(rowCount int) string

	// sortEmptyAsSmallest sorts by an expression in a direction, ASC or DESC,
	// with empty values counted as the smallest: first going up, last going
	// down. Databases differ here, so without it the same request would get
	// different answers on different databases.
	sortEmptyAsSmallest(expression string, direction string) string
}

// valuePlaceholder stands for a value while SQL is being written. BuildQuery
// turns each one into the database's own mark at the end (replaceValuePlaceholders). It
// cannot appear in SQL written by hand, so a ? in such SQL, which Postgres uses
// as an operator, is never taken for a value.
const valuePlaceholder = "\x00value\x00"

func valuePlaceholders(count int) string {
	marks := make([]string, 0, count)

	for range count {
		marks = append(marks, valuePlaceholder)
	}

	return strings.Join(marks, ", ")
}

// textMatchKind says where in a value a searched text must be.
type textMatchKind string

const (
	matchAnywhere textMatchKind = "anywhere"
	matchAtStart  textMatchKind = "start"
	matchAtEnd    textMatchKind = "end"
)

// allowsTextBefore says whether other text may come before the searched text.
func (kind textMatchKind) allowsTextBefore() bool {
	return kind != matchAtStart
}

// allowsTextAfter says whether other text may come after the searched text.
func (kind textMatchKind) allowsTextAfter() bool {
	return kind != matchAtEnd
}
