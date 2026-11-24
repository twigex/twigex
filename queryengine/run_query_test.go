// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package queryengine

import (
	"context"
	"database/sql"
	"reflect"
	"testing"
)

func runSalesQuery(t *testing.T, request Request) []databaseResult {
	t.Helper()

	return runOnEachDatabase(t, salesModels(), request)
}

// checkRows checks the rows on every database.
func checkRows(t *testing.T, results []databaseResult, wantRows []map[string]any) {
	t.Helper()

	for _, databaseResult := range results {
		checkRowsOf(t, databaseResult, wantRows)
	}
}

// checkRowsOn checks the rows on one kind of database, for the few results
// in which databases really differ, such as how many decimals a division has.
func checkRowsOn(t *testing.T, results []databaseResult, databaseType DatabaseType, wantRows []map[string]any) {
	t.Helper()

	for _, databaseResult := range results {
		if databaseResult.database.databaseType == databaseType {
			checkRowsOf(t, databaseResult, wantRows)
		}
	}
}

func checkRowsOf(t *testing.T, databaseResult databaseResult, wantRows []map[string]any) {
	t.Helper()

	gotRows := databaseResult.result.Rows

	if !reflect.DeepEqual(gotRows, wantRows) {
		t.Errorf("%s: want rows\n%v\ngot\n%v", databaseResult.database.name, wantRows, gotRows)
	}
}

func TestRunQueryMeasuresByCountry(t *testing.T) {
	request := Request{
		Dimensions: []string{"sales.country"},
		Measures:   []string{"sales.count", "sales.revenue", "sales.average_sale", "sales.margin"},
		SortBy:     []SortField{{Field: "sales.country"}},
		RowLimit:   100,
	}

	results := runSalesQuery(t, request)

	wantRows := []map[string]any{
		{"sales.country": nil, "sales.count": 1.0, "sales.revenue": 10.0, "sales.average_sale": 10.0, "sales.margin": 50.0},
		{"sales.country": "Estonia", "sales.count": 1.0, "sales.revenue": 300.0, "sales.average_sale": 300.0, "sales.margin": 66.7},
		{"sales.country": "Latvia", "sales.count": 2.0, "sales.revenue": 300.5, "sales.average_sale": 150.25, "sales.margin": 30.1},
		{"sales.country": "Poland", "sales.count": 1.0, "sales.revenue": 50.25, "sales.average_sale": 50.25, "sales.margin": 60.2},
	}

	checkRows(t, results, wantRows)
}

// Empty values count as the smallest on every database: last when sorting
// largest first, as they are first in TestRunQueryMeasuresByCountry.
func TestRunQuerySortsEmptyLastWhenLargestFirst(t *testing.T) {
	request := Request{
		Dimensions: []string{"sales.country"},
		Measures:   []string{"sales.count"},
		SortBy:     []SortField{{Field: "sales.country", Descending: true}},
		RowLimit:   100,
	}

	results := runSalesQuery(t, request)

	wantRows := []map[string]any{
		{"sales.country": "Poland", "sales.count": 1.0},
		{"sales.country": "Latvia", "sales.count": 2.0},
		{"sales.country": "Estonia", "sales.count": 1.0},
		{"sales.country": nil, "sales.count": 1.0},
	}

	checkRows(t, results, wantRows)
}

func TestRunQuerySortsByMeasureLargestFirst(t *testing.T) {
	request := Request{
		Dimensions: []string{"sales.country"},
		Measures:   []string{"sales.revenue"},
		SortBy:     []SortField{{Field: "sales.revenue", Descending: true}},
		RowLimit:   100,
	}

	results := runSalesQuery(t, request)

	wantRows := []map[string]any{
		{"sales.country": "Latvia", "sales.revenue": 300.5},
		{"sales.country": "Estonia", "sales.revenue": 300.0},
		{"sales.country": "Poland", "sales.revenue": 50.25},
		{"sales.country": nil, "sales.revenue": 10.0},
	}

	checkRows(t, results, wantRows)
}

func TestRunQueryValueTypes(t *testing.T) {
	request := Request{
		Dimensions: []string{"sales.big_code", "sales.price", "sales.quantity", "sales.is_paid", "sales.rating"},
		SortBy:     []SortField{{Field: "sales.id"}},
		RowLimit:   100,
		RowsOnly:   true,
	}

	results := runSalesQuery(t, request)

	wantRows := []map[string]any{
		{"sales.big_code": "9007199254740993", "sales.price": "100.50", "sales.quantity": 2.0, "sales.is_paid": true, "sales.rating": 4.5},
		{"sales.big_code": "2", "sales.price": "200.00", "sales.quantity": 1.0, "sales.is_paid": false, "sales.rating": 3.0},
		{"sales.big_code": "3", "sales.price": "50.25", "sales.quantity": 3.0, "sales.is_paid": true, "sales.rating": nil},
		{"sales.big_code": "4", "sales.price": "10.00", "sales.quantity": 1.0, "sales.is_paid": nil, "sales.rating": 2.0},
		{"sales.big_code": "5", "sales.price": "300.00", "sales.quantity": 5.0, "sales.is_paid": true, "sales.rating": 5.0},
	}

	checkRows(t, results, wantRows)

	wantKeptAsText := map[string]bool{
		"sales.big_code": true,
		"sales.price":    true,
		"sales.quantity": false,
		"sales.is_paid":  false,
		"sales.rating":   false,
	}

	for _, databaseResult := range results {
		for _, column := range databaseResult.result.Columns {
			if column.KeptAsText != wantKeptAsText[column.Key] {
				t.Errorf("%s, %s: want KeptAsText %v, got %v", databaseResult.database.name, column.Key, wantKeptAsText[column.Key], column.KeptAsText)
			}
		}
	}
}

func TestRunQueryContainsMatchesPercentAndUnderscoreAsThemselves(t *testing.T) {
	cases := []struct {
		searchFor string
		wantNote  string
	}{
		{"50%", "50% off"},
		{"_", "half_price"},
		{"FREE", "free"},
	}

	for _, testCase := range cases {
		t.Run(testCase.searchFor, func(t *testing.T) {
			request := Request{
				Dimensions: []string{"sales.note"},
				Filters: []Filter{
					{Field: "sales.note", Operator: FilterContains, Values: []string{testCase.searchFor}},
				},
				RowLimit: 100,
				RowsOnly: true,
			}

			results := runSalesQuery(t, request)

			wantRows := []map[string]any{{"sales.note": testCase.wantNote}}
			checkRows(t, results, wantRows)
		})
	}
}

func TestRunQueryNotEqualsIncludesEmptyValues(t *testing.T) {
	request := Request{
		Measures: []string{"sales.count"},
		Filters: []Filter{
			{Field: "sales.country", Operator: FilterNotEquals, Values: []string{"Latvia"}},
		},
		RowLimit: 100,
	}

	results := runSalesQuery(t, request)

	wantRows := []map[string]any{{"sales.count": 3.0}}
	checkRows(t, results, wantRows)
}

// Requests filter on a yes/no value as results give it: true or false.
func TestRunQueryFiltersYesNoFieldsByTrueAndFalse(t *testing.T) {
	cases := []struct {
		values    []string
		wantCount float64
	}{
		{[]string{"true"}, 3.0},
		{[]string{"false"}, 1.0},
		{[]string{"1"}, 3.0},
		{[]string{"true", "false"}, 4.0},
	}

	for _, testCase := range cases {
		request := Request{
			Measures: []string{"sales.count"},
			Filters:  []Filter{{Field: "sales.is_paid", Operator: FilterEquals, Values: testCase.values}},
			RowLimit: 100,
		}

		results := runSalesQuery(t, request)

		checkRows(t, results, []map[string]any{{"sales.count": testCase.wantCount}})
	}
}

func TestRunQueryFilterOnMeasure(t *testing.T) {
	request := Request{
		Dimensions: []string{"sales.country"},
		Measures:   []string{"sales.count"},
		Filters: []Filter{
			{Field: "sales.count", Operator: FilterGreaterThan, Values: []string{"1"}},
		},
		RowLimit: 100,
	}

	results := runSalesQuery(t, request)

	wantRows := []map[string]any{{"sales.country": "Latvia", "sales.count": 2.0}}
	checkRows(t, results, wantRows)
}

// An OR in a field's own SQL stays inside its filter: only the Latvian sale
// in a shop matches, not the Estonian one too.
func TestRunQueryKeepsWrittenSQLInsideItsFilter(t *testing.T) {
	dataset := salesDataset()
	dataset.Dimensions = append(dataset.Dimensions, Dimension{
		Name: "shop_or_estonia",
		Type: FieldTypeYesNo,
		SQL:  "{DATASET}.channel = 'Shop' OR {DATASET}.country = 'Estonia'",
	})
	models := Models{Datasets: []Dataset{dataset}}

	request := Request{
		Measures: []string{"sales.count"},
		Filters: []Filter{
			{Field: "sales.country", Operator: FilterEquals, Values: []string{"Latvia"}},
			{Field: "sales.shop_or_estonia", Operator: FilterEquals, Values: []string{"true"}},
		},
		RowLimit: 10,
	}

	results := runOnEachDatabase(t, models, request)

	checkRows(t, results, []map[string]any{{"sales.count": 1.0}})
}

// A time worked out by hand, such as a time plus an hour, is turned into UTC
// as a whole. On Postgres, a conversion on only its last part fails.
func TestRunQueryTimeWrittenAsSum(t *testing.T) {
	eachTestDatabase(t, func(t *testing.T, database testDatabase) {
		laterSQL := "{DATASET}.happened_at + INTERVAL 1 HOUR"
		if database.databaseType == DatabasePostgres {
			laterSQL = "{DATASET}.happened_at + interval '1 hour'"
		}

		dataset := eventsDataset()
		dataset.Dimensions = append(dataset.Dimensions, Dimension{Name: "hour_later", Type: FieldTypeTime, SQL: laterSQL})
		models := Models{Datasets: []Dataset{dataset}}

		request := Request{
			Measures:   []string{"events.count"},
			DateFields: []DateField{{Field: "events.hour_later", Step: DateStepDay, Range: DateRange{Start: "2025-03-03", End: "2025-03-03"}}},
			SortBy:     []SortField{{Field: "events.hour_later"}},
			RowLimit:   10,
		}

		result := runOn(t, database, database.connection, models, request)

		wantRows := []map[string]any{{"events.hour_later": "2025-03-03T00:00:00Z", "events.count": 3.0}}
		checkRowsOf(t, databaseResult{database: database, result: result}, wantRows)
	})
}

// Values with no JSON number or date to be sent as are sent as empty, rather
// than failing the whole result.
func TestConvertValueSendsNumbersAndDatesWithoutAValueAsEmpty(t *testing.T) {
	number := Column{Key: "rating", Type: FieldTypeNumber}
	moment := Column{Key: "happened_at", Type: FieldTypeTime}

	cases := []struct {
		column Column
		text   string
	}{
		{number, "NaN"},
		{number, "Infinity"},
		{number, "-Infinity"},
		{moment, "infinity"},
		{moment, "-infinity"},
		{Column{Key: "exact", Type: FieldTypeNumber, KeptAsText: true}, "NaN"},
		{Column{Key: "exact", Type: FieldTypeNumber, KeptAsText: true}, "-Infinity"},
	}

	for _, testCase := range cases {
		value, err := convertValue(testCase.column, sql.NullString{String: testCase.text, Valid: true})
		if err != nil || value != nil {
			t.Errorf("%s %q: want empty, got %v, %v", testCase.column.Key, testCase.text, value, err)
		}
	}
}

// Postgres has NaN, Infinity and infinite times; MySQL and MariaDB do not.
func TestRunQueryPostgresNaNAndInfinityAreEmpty(t *testing.T) {
	eachTestDatabase(t, func(t *testing.T, database testDatabase) {
		if database.databaseType != DatabasePostgres {
			t.Skip("only Postgres has NaN and infinity")
		}

		dataset := salesDataset()
		dataset.Dimensions = append(dataset.Dimensions,
			Dimension{Name: "not_a_number", Type: FieldTypeNumber, SQL: "CASE WHEN {DATASET}.id = 1 THEN 'NaN'::float8 ELSE 'Infinity'::float8 END"},
			Dimension{Name: "endless", Type: FieldTypeTime, SQL: "'infinity'::timestamptz"},
			Dimension{Name: "exact_not_a_number", Type: FieldTypeNumber, SQL: "'NaN'::numeric"},
		)
		dataset.Measures = append(dataset.Measures, Measure{Name: "largest_rating", Type: MeasureMaximum, SQL: "CASE WHEN {DATASET}.id = 1 THEN 'NaN'::float8 ELSE {DATASET}.rating END"})
		models := Models{Datasets: []Dataset{dataset}}

		request := Request{
			Dimensions: []string{"sales.not_a_number", "sales.endless", "sales.exact_not_a_number"},
			Measures:   []string{"sales.largest_rating"},
			SortBy:     []SortField{{Field: "sales.not_a_number"}},
			RowLimit:   10,
		}

		result := runOn(t, database, database.connection, models, request)

		wantRows := []map[string]any{
			{"sales.not_a_number": nil, "sales.endless": nil, "sales.exact_not_a_number": nil, "sales.largest_rating": 5.0},
			{"sales.not_a_number": nil, "sales.endless": nil, "sales.exact_not_a_number": nil, "sales.largest_rating": nil},
		}
		checkRowsOf(t, databaseResult{database: database, result: result}, wantRows)
	})
}

// Numbers and times can be searched as text on every database: Postgres has
// no ILIKE for them until they are made text.
func TestRunQuerySearchesNumbersAndTimesAsText(t *testing.T) {
	request := Request{
		Measures: []string{"sales.count"},
		Filters: []Filter{
			{Field: "sales.price", Operator: FilterContains, Values: []string{"0.5"}},
		},
		RowLimit: 10,
	}

	checkRows(t, runSalesQuery(t, request), []map[string]any{{"sales.count": 1.0}})

	timeRequest := Request{
		Measures: []string{"events.count"},
		Filters: []Filter{
			{Field: "events.happened_at", Operator: FilterStartsWith, Values: []string{"2025-03-03"}},
		},
		RowLimit: 10,
	}

	results := runOnEachDatabase(t, eventsModels(), timeRequest)

	checkRows(t, results, []map[string]any{{"events.count": 2.0}})
}

func TestRunQueryWithNoMatchingRows(t *testing.T) {
	request := Request{
		Dimensions: []string{"sales.country"},
		Filters: []Filter{
			{Field: "sales.country", Operator: FilterEquals, Values: []string{"Nowhere"}},
		},
		RowLimit: 100,
	}

	results := runSalesQuery(t, request)

	for _, databaseResult := range results {
		rows := databaseResult.result.Rows

		if rows == nil || len(rows) != 0 {
			t.Errorf("%s: want an empty list of rows, got %v", databaseResult.database.name, rows)
		}
	}
}

func TestRunQueryStopsWhenCancelled(t *testing.T) {
	eachTestDatabase(t, func(t *testing.T, database testDatabase) {
		models := salesModels()
		models.DatabaseTypes = map[string]DatabaseType{"connection-1": database.databaseType}

		request := Request{Measures: []string{"sales.count"}, RowLimit: 10}

		query, err := BuildQuery(models, request)
		if err != nil {
			t.Fatalf("BuildQuery: %v", err)
		}

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err = RunQuery(ctx, database.connection, query)
		if err == nil {
			t.Error("want an error from a cancelled query, got none")
		}
	})
}

func TestIsExactAsNumber(t *testing.T) {
	cases := []struct {
		text    string
		isExact bool
	}{
		{"151.50", true},
		{"0.1", true},
		{"-1.500000", true},
		{"9007199254740992", true},
		{"9007199254740993", false},
		{"9223372036854775807", false},
		// Exact as a float64, but written back as -9223372036854776000.
		{"-9223372036854775808", false},
		{"1000000000000000000", true},
		{"12345678901234.123456", false},
		{"1.2345678901234566e17", true},
		{"not a number", true},
	}

	for _, testCase := range cases {
		if got := isExactAsNumber(testCase.text); got != testCase.isExact {
			t.Errorf("isExactAsNumber(%q) = %v, want %v", testCase.text, got, testCase.isExact)
		}
	}
}

// A sum or largest value too long for a float64 comes back as the database's
// text, with every digit; a count and an average stay numbers. The sales table
// has 5 rows: 5 × 9007199254740993 = 45035996273704965.
func TestRunQueryKeepsTooLongMeasuresAsText(t *testing.T) {
	models := salesModels()
	models.Datasets[0].Measures = append(models.Datasets[0].Measures,
		Measure{Name: "largest_code", Type: MeasureMaximum, SQL: "9223372036854775807"},
		Measure{Name: "code_total", Type: MeasureSum, SQL: "9007199254740993"},
		Measure{Name: "average_code", Type: MeasureAverage, SQL: "9007199254740993"},
	)

	request := Request{
		Measures: []string{"sales.count", "sales.largest_code", "sales.code_total", "sales.average_code"},
		RowLimit: 100,
	}

	results := runOnEachDatabase(t, models, request)

	wantRows := []map[string]any{
		{
			"sales.count":        5.0,
			"sales.largest_code": "9223372036854775807",
			"sales.code_total":   "45035996273704965",
			"sales.average_code": 9007199254740992.0,
		},
	}

	checkRows(t, results, wantRows)

	wantKeptAsText := map[string]bool{
		"sales.count":        false,
		"sales.largest_code": true,
		"sales.code_total":   true,
		"sales.average_code": false,
	}

	for _, databaseResult := range results {
		for _, column := range databaseResult.result.Columns {
			if column.KeptAsText != wantKeptAsText[column.Key] {
				t.Errorf("%s, %s: want KeptAsText %v, got %v", databaseResult.database.name, column.Key, wantKeptAsText[column.Key], column.KeptAsText)
			}
		}
	}
}
