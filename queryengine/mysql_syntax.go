// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package queryengine

import (
	"fmt"
	"strings"
	"time"
)

// mysqlDatabase writes SQL for MySQL and MariaDB. Stored times are read in the
// session's time zone and turned into UTC, and dates in ranges are given in
// UTC and turned into the session's time zone before comparing. Connector
// sets every session to UTC, so stored times are taken as UTC.
type mysqlDatabase struct{}

func (mysqlDatabase) quoteName(name string) string {
	escaped := strings.ReplaceAll(name, "`", "``")

	return "`" + escaped + "`"
}

func (mysqlDatabase) replaceValuePlaceholders(sql string) string {
	return strings.ReplaceAll(sql, valuePlaceholder, "?")
}

func (mysqlDatabase) textMatch(expression string, kind textMatchKind, isNegated bool) string {
	matchWord := "LIKE"
	if isNegated {
		matchWord = "NOT LIKE"
	}

	before := "''"
	if kind.allowsTextBefore() {
		before = "'%'"
	}

	after := "''"
	if kind.allowsTextAfter() {
		after = "'%'"
	}

	pattern := "CONCAT(" + before + ", LOWER(" + valuePlaceholder + "), " + after + ")"

	return "LOWER(" + expression + ") " + matchWord + " " + pattern + " ESCAPE '!'"
}

func (mysqlDatabase) sameOrBothEmpty(left string, right string) string {
	return left + " <=> " + right
}

func (mysqlDatabase) timeInUTC(expression string) string {
	return "CONVERT_TZ(" + expression + ", @@session.time_zone, '+00:00')"
}

// startOfStep works out the start of each step. Weeks start on Monday. Weeks
// are counted from 1000-01-06, a Monday, and quarters from 1000-01-01, the
// earliest year MySQL's dates hold: TIMESTAMPDIFF rounds toward zero, so a
// date before the day counted from would land in the next week or quarter.
func (mysqlDatabase) startOfStep(utcExpression string, step DateStep) string {
	switch step {
	case DateStepSecond:
		return mysqlFormattedAsTime(utcExpression, "%Y-%m-%dT%H:%i:%S.000")
	case DateStepMinute:
		return mysqlFormattedAsTime(utcExpression, "%Y-%m-%dT%H:%i:00.000")
	case DateStepHour:
		return mysqlFormattedAsTime(utcExpression, "%Y-%m-%dT%H:00:00.000")
	case DateStepDay:
		return mysqlFormattedAsTime(utcExpression, "%Y-%m-%dT00:00:00.000")
	case DateStepWeek:
		mondayOfWeek := "DATE_ADD('1000-01-06', INTERVAL TIMESTAMPDIFF(WEEK, '1000-01-06', " + utcExpression + ") WEEK)"
		return mysqlFormattedAsTime(mondayOfWeek, "%Y-%m-%dT00:00:00.000")
	case DateStepMonth:
		return mysqlFormattedAsTime(utcExpression, "%Y-%m-01T00:00:00.000")
	case DateStepQuarter:
		firstDayOfQuarter := "DATE_ADD('1000-01-01', INTERVAL TIMESTAMPDIFF(QUARTER, '1000-01-01', " + utcExpression + ") QUARTER)"
		return "CAST(" + firstDayOfQuarter + " AS DATETIME)"
	case DateStepYear:
		return mysqlFormattedAsTime(utcExpression, "%Y-01-01T00:00:00.000")
	}

	return ""
}

func mysqlFormattedAsTime(expression string, format string) string {
	return "CAST(DATE_FORMAT(" + expression + ", '" + format + "') AS DATETIME)"
}

func (mysqlDatabase) datePlaceholder() string {
	return "TIMESTAMP(CONVERT_TZ(" + valuePlaceholder + ", '+00:00', @@session.time_zone))"
}

// dateForSQL writes a moment in UTC, without a time zone, because CONVERT_TZ
// does not read one: 2025-03-01T00:00:00.000.
func (mysqlDatabase) dateForSQL(moment time.Time) string {
	return moment.UTC().Format("2006-01-02T15:04:05.000")
}

func (mysqlDatabase) rowLimit(rowCount int) string {
	return fmt.Sprintf("LIMIT %d", rowCount)
}

// sortEmptyAsSmallest adds nothing: MySQL already counts empty values as the
// smallest, and has no NULLS FIRST or NULLS LAST.
func (mysqlDatabase) sortEmptyAsSmallest(expression string, direction string) string {
	return expression + " " + direction
}
