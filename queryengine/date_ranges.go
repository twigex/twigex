// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package queryengine

import (
	"strings"
	"time"
)

// ResolvedDateRange is the actual start and end of a date field's range, so a
// chart asking for "this month" knows which days that was.
type ResolvedDateRange struct {
	Field string
	Start string
	End   string
}

type dateRangeMoments struct {
	start time.Time
	end   time.Time
}

const lastMillisecondOfDay = 24*time.Hour - time.Millisecond

// Dates without a time zone are read as UTC.
var dateLayouts = []string{
	time.RFC3339Nano,
	"2006-01-02T15:04:05.999999999",
	"2006-01-02 15:04:05.999999999",
}

const dayLayout = "2006-01-02"

// resolveDateRange works out the start and end of a range. hasRange is false
// for an empty range, which keeps every row.
func resolveDateRange(field string, dateRange DateRange, now time.Time) (resolved dateRangeMoments, hasRange bool, err error) {
	if dateRange.Named != "" {
		resolved, err := namedDateRange(field, dateRange.Named, now)
		return resolved, true, err
	}

	if dateRange.Start == "" && dateRange.End == "" {
		return dateRangeMoments{}, false, nil
	}

	start, err := rangeStart(field, dateRange.Start)
	if err != nil {
		return dateRangeMoments{}, false, err
	}

	end, err := rangeEnd(field, dateRange.End)
	if err != nil {
		return dateRangeMoments{}, false, err
	}

	return dateRangeMoments{start: start, end: end}, true, nil
}

// namedDateRange works out ranges such as "this month" from today's date in
// UTC. Weeks start on Monday.
func namedDateRange(field string, name string, now time.Time) (dateRangeMoments, error) {
	nowInUTC := now.UTC()
	today := time.Date(nowInUTC.Year(), nowInUTC.Month(), nowInUTC.Day(), 0, 0, 0, 0, time.UTC)

	thisMonday := today.AddDate(0, 0, -daysSinceMonday(today))

	firstOfMonth := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, time.UTC)
	firstOfQuarter := time.Date(today.Year(), firstMonthOfQuarter(today.Month()), 1, 0, 0, 0, 0, time.UTC)

	firstOfYear := time.Date(today.Year(), time.January, 1, 0, 0, 0, 0, time.UTC)

	switch strings.ToLower(strings.TrimSpace(name)) {
	case "today":
		return wholeDays(today, today), nil
	case "yesterday":
		yesterday := today.AddDate(0, 0, -1)
		return wholeDays(yesterday, yesterday), nil
	case "this week":
		return wholeDays(thisMonday, thisMonday.AddDate(0, 0, 6)), nil
	case "last week":
		return wholeDays(thisMonday.AddDate(0, 0, -7), thisMonday.AddDate(0, 0, -1)), nil
	case "this month":
		return wholeDays(firstOfMonth, firstOfMonth.AddDate(0, 1, -1)), nil
	case "last month":
		return wholeDays(firstOfMonth.AddDate(0, -1, 0), firstOfMonth.AddDate(0, 0, -1)), nil
	case "this quarter":
		return wholeDays(firstOfQuarter, firstOfQuarter.AddDate(0, 3, -1)), nil
	case "last quarter":
		return wholeDays(firstOfQuarter.AddDate(0, -3, 0), firstOfQuarter.AddDate(0, 0, -1)), nil
	case "this year":
		return wholeDays(firstOfYear, firstOfYear.AddDate(1, 0, -1)), nil
	case "last year":
		return wholeDays(firstOfYear.AddDate(-1, 0, 0), firstOfYear.AddDate(0, 0, -1)), nil
	case "last 7 days":
		return wholeDays(today.AddDate(0, 0, -7), today.AddDate(0, 0, -1)), nil
	case "last 30 days":
		return wholeDays(today.AddDate(0, 0, -30), today.AddDate(0, 0, -1)), nil
	}

	return dateRangeMoments{}, newRequestError(ErrorUnknownDateRange, field)
}

func wholeDays(firstDay time.Time, lastDay time.Time) dateRangeMoments {
	return dateRangeMoments{
		start: firstDay,
		end:   lastDay.Add(lastMillisecondOfDay),
	}
}

// rangeStart reads the start of a range: a date alone means the start of
// that day.
func rangeStart(field string, text string) (time.Time, error) {
	moment, _, err := parseDate(field, text)

	return moment, err
}

// rangeEnd reads the end of a range: a date alone means the end of that day,
// so the whole day is included.
func rangeEnd(field string, text string) (time.Time, error) {
	moment, hasTime, err := parseDate(field, text)
	if err != nil {
		return time.Time{}, err
	}

	if !hasTime {
		moment = moment.Add(lastMillisecondOfDay)
	}

	return moment, nil
}

func parseDate(field string, text string) (moment time.Time, hasTime bool, err error) {
	trimmed := strings.TrimSpace(text)

	day, err := time.Parse(dayLayout, trimmed)
	if err == nil {
		return day, false, nil
	}

	for _, layout := range dateLayouts {
		moment, err := time.Parse(layout, trimmed)
		if err == nil {
			return moment.UTC(), true, nil
		}
	}

	return time.Time{}, false, newRequestError(ErrorBadDate, field)
}

// dateForResult writes a moment the way results send it, with the time zone
// written out: 2025-03-01T00:00:00Z.
func dateForResult(moment time.Time) string {
	return moment.UTC().Format(time.RFC3339Nano)
}

// daysSinceMonday is 0 on a Monday and 6 on a Sunday. Go counts weekdays from
// Sunday, as 0.
func daysSinceMonday(day time.Time) int {
	daysSinceSunday := int(day.Weekday())

	return (daysSinceSunday + 6) % 7
}

// firstMonthOfQuarter gives January, April, July or October.
func firstMonthOfQuarter(month time.Month) time.Month {
	quarterPosition := (int(month) - 1) / 3

	return time.Month(quarterPosition*3 + 1)
}
