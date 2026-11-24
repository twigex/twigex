// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package queryengine

import (
	"errors"
	"testing"
	"time"
)

// A Tuesday.
var tuesdayAfternoon = time.Date(2026, time.November, 17, 15, 0, 0, 0, time.UTC)

func checkRange(t *testing.T, got dateRangeMoments, wantStart string, wantEnd string) {
	t.Helper()

	gotStart := mysqlDatabase{}.dateForSQL(got.start)
	gotEnd := mysqlDatabase{}.dateForSQL(got.end)

	if gotStart != wantStart || gotEnd != wantEnd {
		t.Errorf("want %s to %s, got %s to %s", wantStart, wantEnd, gotStart, gotEnd)
	}
}

func TestNamedDateRanges(t *testing.T) {
	cases := []struct {
		name      string
		wantStart string
		wantEnd   string
	}{
		{"today", "2026-11-17T00:00:00.000", "2026-11-17T23:59:59.999"},
		{"yesterday", "2026-11-16T00:00:00.000", "2026-11-16T23:59:59.999"},
		{"this week", "2026-11-16T00:00:00.000", "2026-11-22T23:59:59.999"},
		{"last week", "2026-11-09T00:00:00.000", "2026-11-15T23:59:59.999"},
		{"this month", "2026-11-01T00:00:00.000", "2026-11-30T23:59:59.999"},
		{"last month", "2026-10-01T00:00:00.000", "2026-10-31T23:59:59.999"},
		{"this quarter", "2026-10-01T00:00:00.000", "2026-12-31T23:59:59.999"},
		{"last quarter", "2026-07-01T00:00:00.000", "2026-09-30T23:59:59.999"},
		{"this year", "2026-01-01T00:00:00.000", "2026-12-31T23:59:59.999"},
		{"last year", "2025-01-01T00:00:00.000", "2025-12-31T23:59:59.999"},
		{"last 7 days", "2026-11-10T00:00:00.000", "2026-11-16T23:59:59.999"},
		{"last 30 days", "2026-10-18T00:00:00.000", "2026-11-16T23:59:59.999"},
		{"This Month", "2026-11-01T00:00:00.000", "2026-11-30T23:59:59.999"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			dateRange := DateRange{Named: testCase.name}

			got, hasRange, err := resolveDateRange("sales.sold_at", dateRange, tuesdayAfternoon)
			if err != nil {
				t.Fatalf("resolveDateRange: %v", err)
			}

			if !hasRange {
				t.Fatal("a named range was read as no range")
			}

			checkRange(t, got, testCase.wantStart, testCase.wantEnd)
		})
	}
}

func TestNamedDateRangesUseTodayInUTC(t *testing.T) {
	riga := time.FixedZone("Riga", 3*60*60)

	// 01:00 on Wednesday in Riga is still Tuesday in UTC.
	earlyWednesdayInRiga := time.Date(2026, time.November, 18, 1, 0, 0, 0, riga)

	got, _, err := resolveDateRange("sales.sold_at", DateRange{Named: "today"}, earlyWednesdayInRiga)
	if err != nil {
		t.Fatalf("resolveDateRange: %v", err)
	}

	checkRange(t, got, "2026-11-17T00:00:00.000", "2026-11-17T23:59:59.999")
}

func TestThisWeekOnSundayAndMonday(t *testing.T) {
	sunday := time.Date(2026, time.November, 22, 12, 0, 0, 0, time.UTC)
	monday := time.Date(2026, time.November, 23, 12, 0, 0, 0, time.UTC)

	sundayWeek, _, err := resolveDateRange("sales.sold_at", DateRange{Named: "this week"}, sunday)
	if err != nil {
		t.Fatalf("resolveDateRange: %v", err)
	}

	checkRange(t, sundayWeek, "2026-11-16T00:00:00.000", "2026-11-22T23:59:59.999")

	mondayWeek, _, err := resolveDateRange("sales.sold_at", DateRange{Named: "this week"}, monday)
	if err != nil {
		t.Fatalf("resolveDateRange: %v", err)
	}

	checkRange(t, mondayWeek, "2026-11-23T00:00:00.000", "2026-11-29T23:59:59.999")
}

func TestCustomDateRanges(t *testing.T) {
	cases := []struct {
		name      string
		dateRange DateRange
		wantStart string
		wantEnd   string
	}{
		{
			"dates alone cover whole days",
			DateRange{Start: "2025-03-01", End: "2025-03-31"},
			"2025-03-01T00:00:00.000", "2025-03-31T23:59:59.999",
		},
		{
			"times are kept",
			DateRange{Start: "2025-03-01T10:30:00", End: "2025-03-02T08:00:00"},
			"2025-03-01T10:30:00.000", "2025-03-02T08:00:00.000",
		},
		{
			"milliseconds and a time zone",
			DateRange{Start: "2025-03-01T10:30:00.250Z", End: "2025-03-02"},
			"2025-03-01T10:30:00.250", "2025-03-02T23:59:59.999",
		},
		{
			"another time zone becomes UTC",
			DateRange{Start: "2025-03-01T10:30:00+03:00", End: "2025-03-02T00:00:00.000"},
			"2025-03-01T07:30:00.000", "2025-03-02T00:00:00.000",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			got, hasRange, err := resolveDateRange("sales.sold_at", testCase.dateRange, tuesdayAfternoon)
			if err != nil {
				t.Fatalf("resolveDateRange: %v", err)
			}

			if !hasRange {
				t.Fatal("a custom range was read as no range")
			}

			checkRange(t, got, testCase.wantStart, testCase.wantEnd)
		})
	}
}

func TestEmptyDateRangeKeepsEveryRow(t *testing.T) {
	_, hasRange, err := resolveDateRange("sales.sold_at", DateRange{}, tuesdayAfternoon)
	if err != nil {
		t.Fatalf("resolveDateRange: %v", err)
	}

	if hasRange {
		t.Error("an empty range was read as a range")
	}
}

func TestDateRangeErrors(t *testing.T) {
	cases := []struct {
		name      string
		dateRange DateRange
		wantCode  ErrorCode
	}{
		{"unknown name", DateRange{Named: "next century"}, ErrorUnknownDateRange},
		{"start that is not a date", DateRange{Start: "soon", End: "2025-03-01"}, ErrorBadDate},
		{"missing end", DateRange{Start: "2025-03-01"}, ErrorBadDate},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			_, _, err := resolveDateRange("sales.sold_at", testCase.dateRange, tuesdayAfternoon)

			var requestError *RequestError
			if !errors.As(err, &requestError) {
				t.Fatalf("want a RequestError, got %v", err)
			}

			if requestError.Code != testCase.wantCode {
				t.Errorf("want code %q, got %q", testCase.wantCode, requestError.Code)
			}

			if requestError.Field != "sales.sold_at" {
				t.Errorf("want field sales.sold_at, got %q", requestError.Field)
			}
		})
	}
}
