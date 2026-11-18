// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package queryengine

import (
	"errors"
	"slices"
	"testing"
)

const happenedAtInUTC = "CONVERT_TZ(`events`.`happened_at`, @@session.time_zone, '+00:00')"

const dateValue = "TIMESTAMP(CONVERT_TZ(?, '+00:00', @@session.time_zone))"

func buildEventsQuery(t *testing.T, request Request) BuiltQuery {
	t.Helper()

	request.Now = tuesdayAfternoon

	query, err := BuildQuery(eventsModels(), request)
	if err != nil {
		t.Fatalf("BuildQuery: %v", err)
	}

	return query
}

func TestBuildQueryDateFieldByDayInRange(t *testing.T) {
	request := Request{
		Measures: []string{"events.count"},
		DateFields: []DateField{
			{Field: "events.happened_at", Step: DateStepDay, Range: DateRange{Start: "2025-03-01", End: "2025-03-31"}},
		},
		RowLimit: 10,
	}

	query := buildEventsQuery(t, request)

	wantSQL := "SELECT CAST(DATE_FORMAT(" + happenedAtInUTC + ", '%Y-%m-%dT00:00:00.000') AS DATETIME) AS `events__happened_at`, COUNT(`events`.`id`) AS `events__count`\n" +
		"FROM `events` AS `events`\n" +
		"WHERE (`events`.`happened_at` >= " + dateValue + " AND `events`.`happened_at` < " + dateValue + ")\n" +
		"GROUP BY 1\n" +
		"LIMIT 10"

	checkSQL(t, query, wantSQL, []any{"2025-03-01T00:00:00.000", "2025-04-01T00:00:00.000"})

	wantColumns := []Column{
		{Key: "events.happened_at", Title: "Happened", Kind: FieldKindDimension, Type: FieldTypeTime, DateStep: DateStepDay},
		{Key: "events.count", Title: "Events", Kind: FieldKindMeasure, Type: FieldTypeNumber},
	}

	if !slices.Equal(query.Columns, wantColumns) {
		t.Errorf("want columns %v, got %v", wantColumns, query.Columns)
	}

	wantRanges := []ResolvedDateRange{
		{Field: "events.happened_at", Start: "2025-03-01T00:00:00Z", End: "2025-03-31T23:59:59.999Z"},
	}

	if !slices.Equal(query.DateRanges, wantRanges) {
		t.Errorf("want date ranges %v, got %v", wantRanges, query.DateRanges)
	}
}

func TestBuildQueryEveryDateStep(t *testing.T) {
	cases := []struct {
		step       DateStep
		wantSelect string
	}{
		{DateStepSecond, "CAST(DATE_FORMAT(" + happenedAtInUTC + ", '%Y-%m-%dT%H:%i:%S.000') AS DATETIME)"},
		{DateStepMinute, "CAST(DATE_FORMAT(" + happenedAtInUTC + ", '%Y-%m-%dT%H:%i:00.000') AS DATETIME)"},
		{DateStepHour, "CAST(DATE_FORMAT(" + happenedAtInUTC + ", '%Y-%m-%dT%H:00:00.000') AS DATETIME)"},
		{DateStepDay, "CAST(DATE_FORMAT(" + happenedAtInUTC + ", '%Y-%m-%dT00:00:00.000') AS DATETIME)"},
		{DateStepWeek, "CAST(DATE_FORMAT(DATE_ADD('1000-01-06', INTERVAL TIMESTAMPDIFF(WEEK, '1000-01-06', " + happenedAtInUTC + ") WEEK), '%Y-%m-%dT00:00:00.000') AS DATETIME)"},
		{DateStepMonth, "CAST(DATE_FORMAT(" + happenedAtInUTC + ", '%Y-%m-01T00:00:00.000') AS DATETIME)"},
		{DateStepQuarter, "CAST(DATE_ADD('1000-01-01', INTERVAL TIMESTAMPDIFF(QUARTER, '1000-01-01', " + happenedAtInUTC + ") QUARTER) AS DATETIME)"},
		{DateStepYear, "CAST(DATE_FORMAT(" + happenedAtInUTC + ", '%Y-01-01T00:00:00.000') AS DATETIME)"},
	}

	for _, testCase := range cases {
		t.Run(string(testCase.step), func(t *testing.T) {
			request := Request{
				DateFields: []DateField{{Field: "events.happened_at", Step: testCase.step}},
				RowLimit:   10,
			}

			query := buildEventsQuery(t, request)

			wantSQL := "SELECT " + testCase.wantSelect + " AS `events__happened_at`\n" +
				"FROM `events` AS `events`\n" +
				"GROUP BY 1\n" +
				"LIMIT 10"

			checkSQL(t, query, wantSQL, nil)
		})
	}
}

func TestBuildQueryDateRangeWithoutStepIsNotShown(t *testing.T) {
	request := Request{
		Measures:   []string{"events.count"},
		DateFields: []DateField{{Field: "events.happened_at", Range: DateRange{Named: "this month"}}},
		RowLimit:   10,
	}

	query := buildEventsQuery(t, request)

	wantSQL := "SELECT COUNT(`events`.`id`) AS `events__count`\n" +
		"FROM `events` AS `events`\n" +
		"WHERE (`events`.`happened_at` >= " + dateValue + " AND `events`.`happened_at` < " + dateValue + ")\n" +
		"LIMIT 10"

	checkSQL(t, query, wantSQL, []any{"2026-11-01T00:00:00.000", "2026-12-01T00:00:00.000"})

	wantRanges := []ResolvedDateRange{
		{Field: "events.happened_at", Start: "2026-11-01T00:00:00Z", End: "2026-11-30T23:59:59.999Z"},
	}

	if !slices.Equal(query.DateRanges, wantRanges) {
		t.Errorf("want date ranges %v, got %v", wantRanges, query.DateRanges)
	}
}

func TestBuildQueryDateFieldWithoutStepIsShownInUTC(t *testing.T) {
	request := Request{
		Dimensions: []string{"events.happened_at"},
		RowLimit:   10,
		RowsOnly:   true,
	}

	query := buildEventsQuery(t, request)

	wantSQL := "SELECT " + happenedAtInUTC + " AS `events__happened_at`\n" +
		"FROM `events` AS `events`\n" +
		"LIMIT 10"

	checkSQL(t, query, wantSQL, nil)
}

func TestBuildQueryDateFilters(t *testing.T) {
	cases := []struct {
		name          string
		filter        Filter
		wantCondition string
		wantValues    []any
	}{
		{
			"in date range",
			Filter{Field: "events.happened_at", Operator: FilterInDateRange, Values: []string{"2025-03-01", "2025-03-31"}},
			"(`events`.`happened_at` >= " + dateValue + " AND `events`.`happened_at` < " + dateValue + ")",
			[]any{"2025-03-01T00:00:00.000", "2025-04-01T00:00:00.000"},
		},
		{
			"not in date range",
			Filter{Field: "events.happened_at", Operator: FilterNotInDateRange, Values: []string{"2025-03-01", "2025-03-31"}},
			"(`events`.`happened_at` < " + dateValue + " OR `events`.`happened_at` >= " + dateValue + ")",
			[]any{"2025-03-01T00:00:00.000", "2025-04-01T00:00:00.000"},
		},
		{
			"before date means before the day starts",
			Filter{Field: "events.happened_at", Operator: FilterBeforeDate, Values: []string{"2025-03-01"}},
			"`events`.`happened_at` < " + dateValue,
			[]any{"2025-03-01T00:00:00.000"},
		},
		{
			"after date means after the day ends",
			Filter{Field: "events.happened_at", Operator: FilterAfterDate, Values: []string{"2025-03-01"}},
			"`events`.`happened_at` >= " + dateValue,
			[]any{"2025-03-02T00:00:00.000"},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			request := Request{
				Measures: []string{"events.count"},
				Filters:  []Filter{testCase.filter},
				RowLimit: 10,
			}

			query := buildEventsQuery(t, request)

			wantSQL := "SELECT COUNT(`events`.`id`) AS `events__count`\n" +
				"FROM `events` AS `events`\n" +
				"WHERE " + testCase.wantCondition + "\n" +
				"LIMIT 10"

			checkSQL(t, query, wantSQL, testCase.wantValues)
		})
	}
}

func TestBuildQueryValuesFollowTheirPlaceInTheSQL(t *testing.T) {
	request := Request{
		Measures: []string{"events.count"},
		DateFields: []DateField{
			{Field: "events.happened_at", Step: DateStepMonth, Range: DateRange{Start: "2025-01-01", End: "2025-12-31"}},
		},
		Filters: []Filter{
			{Field: "events.count", Operator: FilterGreaterThan, Values: []string{"1"}},
			{Field: "events.happened_at", Operator: FilterAfterDate, Values: []string{"2025-02-01"}},
		},
		SortBy:   []SortField{{Field: "events.happened_at", Descending: true}},
		RowLimit: 10,
	}

	query := buildEventsQuery(t, request)

	wantSQL := "SELECT CAST(DATE_FORMAT(" + happenedAtInUTC + ", '%Y-%m-01T00:00:00.000') AS DATETIME) AS `events__happened_at`, COUNT(`events`.`id`) AS `events__count`\n" +
		"FROM `events` AS `events`\n" +
		"WHERE `events`.`happened_at` >= " + dateValue + " AND (`events`.`happened_at` >= " + dateValue + " AND `events`.`happened_at` < " + dateValue + ")\n" +
		"GROUP BY 1\n" +
		"HAVING COUNT(`events`.`id`) > ?\n" +
		"ORDER BY `events__happened_at` DESC\n" +
		"LIMIT 10"

	wantValues := []any{"2025-02-02T00:00:00.000", "2025-01-01T00:00:00.000", "2026-01-01T00:00:00.000", "1"}

	checkSQL(t, query, wantSQL, wantValues)
}

func TestBuildQueryDateErrors(t *testing.T) {
	cases := []struct {
		name      string
		request   Request
		wantCode  ErrorCode
		wantField string
	}{
		{
			"step on a field that is not a date",
			Request{DateFields: []DateField{{Field: "sales.country", Step: DateStepDay}}, RowLimit: 10},
			ErrorNotADateField, "sales.country",
		},
		{
			"measure used as a date field",
			Request{DateFields: []DateField{{Field: "events.count", Step: DateStepDay}}, RowLimit: 10},
			ErrorWrongFieldKind, "events.count",
		},
		{
			"unknown step",
			Request{DateFields: []DateField{{Field: "events.happened_at", Step: "fortnight"}}, RowLimit: 10},
			ErrorUnknownDateStep, "events.happened_at",
		},
		{
			"unknown named range",
			Request{Measures: []string{"events.count"}, DateFields: []DateField{{Field: "events.happened_at", Range: DateRange{Named: "someday"}}}, RowLimit: 10},
			ErrorUnknownDateRange, "events.happened_at",
		},
		{
			"same field shown twice",
			Request{
				Dimensions: []string{"events.happened_at"},
				DateFields: []DateField{{Field: "events.happened_at", Step: DateStepDay}},
				RowLimit:   10,
			},
			ErrorFieldShownTwice, "events.happened_at",
		},
		{
			"date filter with a bad date",
			Request{Measures: []string{"events.count"}, Filters: []Filter{{Field: "events.happened_at", Operator: FilterBeforeDate, Values: []string{"later"}}}, RowLimit: 10},
			ErrorBadDate, "events.happened_at",
		},
		{
			"date range filter with one date",
			Request{Measures: []string{"events.count"}, Filters: []Filter{{Field: "events.happened_at", Operator: FilterInDateRange, Values: []string{"2025-03-01"}}}, RowLimit: 10},
			ErrorMissingFilterValue, "events.happened_at",
		},
		{
			"range only, nothing shown",
			Request{DateFields: []DateField{{Field: "events.happened_at", Range: DateRange{Named: "today"}}}, RowLimit: 10},
			ErrorNoFields, "",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.request.Now = tuesdayAfternoon

			_, err := BuildQuery(eventsModels(), testCase.request)

			var requestError *RequestError
			if !errors.As(err, &requestError) {
				t.Fatalf("want a RequestError, got %v", err)
			}

			if requestError.Code != testCase.wantCode {
				t.Errorf("want code %q, got %q", testCase.wantCode, requestError.Code)
			}

			if requestError.Field != testCase.wantField {
				t.Errorf("want field %q, got %q", testCase.wantField, requestError.Field)
			}
		})
	}
}
