// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package queryengine

import (
	"slices"
	"testing"
)

// runEventsQuery runs a request on every test database, with the session a
// number of hours east of UTC.
func runEventsQuery(t *testing.T, hoursEastOfUTC int, request Request) []databaseResult {
	t.Helper()

	request.Now = tuesdayAfternoon

	return runOnEachDatabaseInTimeZone(t, eventsModels(), request, hoursEastOfUTC)
}

// Tests open the database with its session in UTC unless they check another
// time zone, so they do not depend on the time zone of the server they run on.

func countsByStep(step DateStep, dateRange DateRange) Request {
	return Request{
		Measures:   []string{"events.count"},
		DateFields: []DateField{{Field: "events.happened_at", Step: step, Range: dateRange}},
		SortBy:     []SortField{{Field: "events.happened_at"}},
		RowLimit:   100,
	}
}

func TestRunQueryCountsByDay(t *testing.T) {
	request := countsByStep(DateStepDay, DateRange{Start: "2025-03-01", End: "2025-03-31"})

	results := runEventsQuery(t, 0, request)

	wantRows := []map[string]any{
		{"events.happened_at": "2025-03-02T00:00:00Z", "events.count": 1.0},
		{"events.happened_at": "2025-03-03T00:00:00Z", "events.count": 2.0},
		{"events.happened_at": "2025-03-31T00:00:00Z", "events.count": 1.0},
	}

	checkRows(t, results, wantRows)

	wantRanges := []ResolvedDateRange{
		{Field: "events.happened_at", Start: "2025-03-01T00:00:00Z", End: "2025-03-31T23:59:59.999Z"},
	}

	for _, databaseResult := range results {
		gotRanges := databaseResult.result.DateRanges

		if !slices.Equal(gotRanges, wantRanges) {
			t.Errorf("%s: want date ranges %v, got %v", databaseResult.database.name, wantRanges, gotRanges)
		}
	}
}

func TestRunQueryWeeksStartOnMonday(t *testing.T) {
	request := countsByStep(DateStepWeek, DateRange{})

	results := runEventsQuery(t, 0, request)

	// Sunday 2 March belongs to the week of Monday 24 February; Monday
	// 3 March starts a new week. Rows with no date form their own group.
	wantRows := []map[string]any{
		{"events.happened_at": nil, "events.count": 1.0},
		{"events.happened_at": "2025-02-24T00:00:00Z", "events.count": 1.0},
		{"events.happened_at": "2025-03-03T00:00:00Z", "events.count": 2.0},
		{"events.happened_at": "2025-03-31T00:00:00Z", "events.count": 2.0},
	}

	checkSortedRows(t, results, "events.happened_at", wantRows)
}

func TestRunQueryMonthsQuartersAndYears(t *testing.T) {
	cases := []struct {
		step     DateStep
		wantRows []map[string]any
	}{
		{
			DateStepMonth,
			[]map[string]any{
				{"events.happened_at": "2025-03-01T00:00:00Z", "events.count": 4.0},
				{"events.happened_at": "2025-04-01T00:00:00Z", "events.count": 1.0},
			},
		},
		{
			DateStepQuarter,
			[]map[string]any{
				{"events.happened_at": "2025-01-01T00:00:00Z", "events.count": 4.0},
				{"events.happened_at": "2025-04-01T00:00:00Z", "events.count": 1.0},
			},
		},
		{
			DateStepYear,
			[]map[string]any{
				{"events.happened_at": "2025-01-01T00:00:00Z", "events.count": 5.0},
			},
		},
	}

	for _, testCase := range cases {
		t.Run(string(testCase.step), func(t *testing.T) {
			request := countsByStep(testCase.step, DateRange{Start: "2025-01-01", End: "2025-12-31"})

			results := runEventsQuery(t, 0, request)

			checkRows(t, results, testCase.wantRows)
		})
	}
}

// With the database session at UTC+3, a stored 01:30 on 3 March is 22:30 on
// 2 March in UTC, so all of the first three events fall on 2 March, and a
// stored midnight on 1 April is 21:00 on 31 March.
func TestRunQueryReadsDatesInUTCWhateverTheDatabaseTimeZone(t *testing.T) {
	request := countsByStep(DateStepDay, DateRange{Start: "2025-03-01", End: "2025-03-31"})

	results := runEventsQuery(t, 3, request)

	wantRows := []map[string]any{
		{"events.happened_at": "2025-03-02T00:00:00Z", "events.count": 3.0},
		{"events.happened_at": "2025-03-31T00:00:00Z", "events.count": 2.0},
	}

	checkRows(t, results, wantRows)
}

func TestRunQueryRangeIsInUTCWhateverTheDatabaseTimeZone(t *testing.T) {
	request := Request{
		Measures:   []string{"events.count"},
		DateFields: []DateField{{Field: "events.happened_at", Range: DateRange{Start: "2025-03-03", End: "2025-03-03"}}},
		RowLimit:   100,
	}

	results := runEventsQuery(t, 3, request)

	wantRows := []map[string]any{{"events.count": 0.0}}
	checkRows(t, results, wantRows)
}

func TestRunQueryDateFieldWithoutStep(t *testing.T) {
	request := Request{
		Dimensions: []string{"events.happened_at"},
		SortBy:     []SortField{{Field: "events.id"}},
		RowLimit:   100,
		RowsOnly:   true,
	}

	results := runEventsQuery(t, 0, request)

	wantRows := []map[string]any{
		{"events.happened_at": "2025-03-02T23:59:59Z"},
		{"events.happened_at": "2025-03-03T00:00:00Z"},
		{"events.happened_at": "2025-03-03T01:30:00Z"},
		{"events.happened_at": "2025-03-31T12:00:00Z"},
		{"events.happened_at": "2025-04-01T00:00:00Z"},
		{"events.happened_at": nil},
	}

	checkRows(t, results, wantRows)
}

func TestRunQueryDateFilters(t *testing.T) {
	cases := []struct {
		name      string
		filter    Filter
		wantCount float64
	}{
		{"before 3 March", Filter{Operator: FilterBeforeDate, Values: []string{"2025-03-03"}}, 1},
		{"after 31 March", Filter{Operator: FilterAfterDate, Values: []string{"2025-03-31"}}, 1},
		{"from 3 to 31 March", Filter{Operator: FilterInDateRange, Values: []string{"2025-03-03", "2025-03-31"}}, 3},
		{"outside 3 to 31 March", Filter{Operator: FilterNotInDateRange, Values: []string{"2025-03-03", "2025-03-31"}}, 2},
		{"between two times", Filter{Operator: FilterInDateRange, Values: []string{"2025-03-03T00:00:00", "2025-03-03T01:00:00"}}, 1},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			filter := testCase.filter
			filter.Field = "events.happened_at"

			request := Request{
				Measures: []string{"events.count"},
				Filters:  []Filter{filter},
				RowLimit: 100,
			}

			results := runEventsQuery(t, 0, request)

			wantRows := []map[string]any{{"events.count": testCase.wantCount}}
			checkRows(t, results, wantRows)
		})
	}
}

// eventsWithShiftedTime gives events a time field moved from happened_at by
// the database's own interval SQL, such as "200 years" back.
func eventsWithShiftedTime(database testDatabase, mysqlShift string, postgresShift string) Models {
	shiftedSQL := "{DATASET}.happened_at " + mysqlShift
	if database.databaseType == DatabasePostgres {
		shiftedSQL = "{DATASET}.happened_at " + postgresShift
	}

	dataset := eventsDataset()
	dataset.Dimensions = append(dataset.Dimensions, Dimension{Name: "shifted", Type: FieldTypeTime, SQL: shiftedSQL})

	return Models{Datasets: []Dataset{dataset}}
}

// A time half a millisecond before midnight still belongs to its day: the
// range ends before the next day starts, not at 23:59:59.999.
func TestRunQueryWholeDayIncludesItsLastMicroseconds(t *testing.T) {
	eachTestDatabase(t, func(t *testing.T, database testDatabase) {
		models := eventsWithShiftedTime(database, "+ INTERVAL 999500 MICROSECOND", "+ interval '999500 microseconds'")

		request := Request{
			Measures:   []string{"events.count"},
			DateFields: []DateField{{Field: "events.shifted", Range: DateRange{Start: "2025-03-02", End: "2025-03-02"}}},
			RowLimit:   10,
		}

		result := runOn(t, database, database.connection, models, request)

		wantRows := []map[string]any{{"events.count": 1.0}}
		checkRowsOf(t, databaseResult{database: database, result: result}, wantRows)
	})
}

// Weeks and quarters before 1900 start on the right day on every database.
func TestRunQueryWeeksAndQuartersBefore1900(t *testing.T) {
	eachTestDatabase(t, func(t *testing.T, database testDatabase) {
		models := eventsWithShiftedTime(database, "- INTERVAL 200 YEAR", "- interval '200 years'")

		weeks := Request{
			Measures:   []string{"events.count"},
			DateFields: []DateField{{Field: "events.shifted", Step: DateStepWeek}},
			SortBy:     []SortField{{Field: "events.shifted"}},
			RowLimit:   10,
		}

		result := runOn(t, database, database.connection, models, weeks)

		wantWeeks := []map[string]any{
			{"events.shifted": nil, "events.count": 1.0},
			{"events.shifted": "1825-02-28T00:00:00Z", "events.count": 3.0},
			{"events.shifted": "1825-03-28T00:00:00Z", "events.count": 2.0},
		}
		checkRowsOf(t, databaseResult{database: database, result: result}, wantWeeks)

		quarters := weeks
		quarters.DateFields = []DateField{{Field: "events.shifted", Step: DateStepQuarter}}

		result = runOn(t, database, database.connection, models, quarters)

		wantQuarters := []map[string]any{
			{"events.shifted": nil, "events.count": 1.0},
			{"events.shifted": "1825-01-01T00:00:00Z", "events.count": 4.0},
			{"events.shifted": "1825-04-01T00:00:00Z", "events.count": 1.0},
		}
		checkRowsOf(t, databaseResult{database: database, result: result}, wantQuarters)
	})
}
