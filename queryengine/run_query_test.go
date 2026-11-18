// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package queryengine

import (
	"context"
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

// checkSortedRows checks rows sorted by sortKey, smallest first. wantRows are
// in MySQL's order, where empty values come first. Postgres puts them last;
// the engine leaves this to the database.
func checkSortedRows(t *testing.T, results []databaseResult, sortKey string, wantRows []map[string]any) {
	t.Helper()

	var emptyLast []map[string]any
	var emptyRows []map[string]any

	for _, row := range wantRows {
		if row[sortKey] == nil {
			emptyRows = append(emptyRows, row)
			continue
		}

		emptyLast = append(emptyLast, row)
	}

	emptyLast = append(emptyLast, emptyRows...)

	checkRowsOn(t, results, DatabaseMySQL, wantRows)
	checkRowsOn(t, results, DatabasePostgres, emptyLast)
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

	checkSortedRows(t, results, "sales.country", wantRows)
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

	wantSentAsText := map[string]bool{
		"sales.big_code": true,
		"sales.price":    true,
		"sales.quantity": false,
		"sales.is_paid":  false,
		"sales.rating":   false,
	}

	for _, databaseResult := range results {
		for _, column := range databaseResult.result.Columns {
			if column.SentAsText != wantSentAsText[column.Key] {
				t.Errorf("%s, %s: want SentAsText %v, got %v", databaseResult.database.name, column.Key, wantSentAsText[column.Key], column.SentAsText)
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

// Charts filter on a yes/no value as results give it: true or false.
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
