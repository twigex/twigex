// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package queryengine

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"testing"
)

// The grid runs every combination of a few request options on every test
// database, to find bugs that come from two options together. Each one must
// run or be refused for the expected reason, give the same rows on every
// database in the order asked for, and, for one dataset, the answer worked
// out in Go (grid_answers_test.go).

// gridShape is a kind of query: its models and the fields a request uses.
type gridShape struct {
	name   string
	models func(databaseType DatabaseType) Models

	dimensions []string
	measures   []string
	timeField  string

	// notShown is a dimension the request does not show, to sort by.
	notShown string

	filters []gridFilter

	// answer works out the expected rows in Go, or is nil when only the
	// databases are compared with each other.
	answer func(combination gridCombination) []map[string]any
}

type gridFilter struct {
	name    string
	filters []Filter

	// onMeasure is a filter on a measure, which rows only cannot have.
	onMeasure bool

	// keeps says whether the filter keeps a sales row, for the answer
	// worked out in Go.
	keeps func(row salesGridRow) bool
}

type gridSort struct {
	name string

	// field picks the field to sort by from the shape, or is nil for none.
	field      func(shape gridShape) string
	descending bool
	onMeasure  bool
	notShown   bool
}

type gridCombination struct {
	shape    gridShape
	rowsOnly bool
	step     DateStep
	filter   gridFilter
	sort     gridSort
}

var gridSteps = []DateStep{"", DateStepDay, DateStepWeek, DateStepQuarter}

var gridSorts = []gridSort{
	{name: "no sort"},
	{name: "by dimension", field: func(shape gridShape) string { return shape.dimensions[0] }},
	{name: "by measure, largest first", field: func(shape gridShape) string { return shape.measures[0] }, descending: true, onMeasure: true},
	{name: "by field not shown", field: func(shape gridShape) string { return shape.notShown }, notShown: true},
}

func (combination gridCombination) name() string {
	rows := "grouped"
	if combination.rowsOnly {
		rows = "rows only"
	}

	step := "no step"
	if combination.step != "" {
		step = string(combination.step)
	}

	return strings.Join([]string{combination.shape.name, rows, step, combination.filter.name, combination.sort.name}, " / ")
}

func (combination gridCombination) request() Request {
	shape := combination.shape

	request := Request{
		Dimensions: shape.dimensions,
		RowsOnly:   combination.rowsOnly,
		RowLimit:   1000,
	}

	if !combination.rowsOnly {
		request.Measures = shape.measures
	}

	if combination.step != "" {
		request.DateFields = []DateField{{Field: shape.timeField, Step: combination.step}}
	}

	request.Filters = combination.filter.filters

	if combination.sort.field != nil {
		sortField := SortField{Field: combination.sort.field(shape), Descending: combination.sort.descending}
		request.SortBy = []SortField{sortField}
	}

	return request
}

// expectedRefusal is the reason the engine must refuse the combination, or
// empty when it must run.
func (combination gridCombination) expectedRefusal() ErrorCode {
	if combination.rowsOnly && (combination.filter.onMeasure || combination.sort.onMeasure) {
		return ErrorMeasuresInRowsOnly
	}

	if !combination.rowsOnly && combination.sort.notShown {
		return ErrorSortFieldNotShown
	}

	return ""
}

func gridCombinations(shape gridShape) []gridCombination {
	var combinations []gridCombination

	for _, rowsOnly := range []bool{false, true} {
		for _, step := range gridSteps {
			// Rows only are not grouped, so they have no date step.
			if rowsOnly && step != "" {
				continue
			}

			for _, filter := range shape.filters {
				for _, sort := range gridSorts {
					combination := gridCombination{shape: shape, rowsOnly: rowsOnly, step: step, filter: filter, sort: sort}
					combinations = append(combinations, combination)
				}
			}
		}
	}

	return combinations
}

// gridOutcome is what one database gave for a combination.
type gridOutcome struct {
	database testDatabase
	refusal  ErrorCode
	rows     []map[string]any
}

func TestGridOfRequests(t *testing.T) {
	if len(testDatabases) == 0 {
		t.Skip("neither TEST_MYSQL_DSN nor TEST_POSTGRES_DSN is set; skipping tests that run queries")
	}

	shapes := []gridShape{salesGridShape(), joinGridShape(), combinedGridShape(), viewGridShape(), eventsGridShape()}

	for _, shape := range shapes {
		for _, combination := range gridCombinations(shape) {
			t.Run(combination.name(), func(t *testing.T) {
				checkGridCombination(t, combination)
			})
		}
	}
}

func checkGridCombination(t *testing.T, combination gridCombination) {
	var outcomes []gridOutcome

	for _, database := range testDatabases {
		outcome, err := runGridCombination(database, combination)
		if err != nil {
			t.Fatalf("%s: %v", database.name, err)
		}

		outcomes = append(outcomes, outcome)
	}

	wantRefusal := combination.expectedRefusal()

	for _, outcome := range outcomes {
		if outcome.refusal != wantRefusal {
			t.Fatalf("%s: want refusal %q, got %q", outcome.database.name, wantRefusal, outcome.refusal)
		}
	}

	if wantRefusal != "" {
		return
	}

	for _, outcome := range outcomes {
		checkGridOrder(t, outcome, combination)
	}

	firstRows := inAnyOrder(outcomes[0].rows)

	for _, outcome := range outcomes[1:] {
		rows := inAnyOrder(outcome.rows)
		if !reflect.DeepEqual(rows, firstRows) {
			t.Errorf("%s and %s differ:\n%v\n%v", outcomes[0].database.name, outcome.database.name, firstRows, rows)
		}
	}

	if combination.shape.answer == nil {
		return
	}

	wantRows := inAnyOrder(combination.shape.answer(combination))

	for _, outcome := range outcomes {
		rows := inAnyOrder(outcome.rows)
		if !reflect.DeepEqual(rows, wantRows) {
			t.Errorf("%s: want rows\n%v\ngot\n%v", outcome.database.name, wantRows, rows)
		}
	}
}

// runGridCombination builds and runs a combination. A refusal is an outcome;
// any other failure, such as the database refusing the SQL, is an error.
func runGridCombination(database testDatabase, combination gridCombination) (gridOutcome, error) {
	models := combination.shape.models(database.databaseType)
	models.DatabaseTypes = map[string]DatabaseType{"connection-1": database.databaseType}

	outcome := gridOutcome{database: database}

	query, err := BuildQuery(models, combination.request())

	var requestError *RequestError
	if errors.As(err, &requestError) {
		outcome.refusal = requestError.Code
		return outcome, nil
	}

	if err != nil {
		return gridOutcome{}, fmt.Errorf("BuildQuery: %w", err)
	}

	result, err := RunQuery(context.Background(), database.connection, query)
	if err != nil {
		return gridOutcome{}, fmt.Errorf("RunQuery: %w\n%s", err, query.SQL)
	}

	outcome.rows = result.Rows

	return outcome, nil
}

// inAnyOrder sorts rows by how they print, so that two lists of the same rows
// in another order compare equal.
func inAnyOrder(rows []map[string]any) []map[string]any {
	sorted := make([]map[string]any, 0, len(rows))
	sorted = append(sorted, rows...)

	slices.SortFunc(sorted, func(first, second map[string]any) int {
		return strings.Compare(fmt.Sprint(first), fmt.Sprint(second))
	})

	return sorted
}

// checkGridOrder checks the rows come in the order asked for, empty values
// counting as the smallest.
func checkGridOrder(t *testing.T, outcome gridOutcome, combination gridCombination) {
	t.Helper()

	if combination.sort.field == nil {
		return
	}

	key := combination.sort.field(combination.shape)

	for position := 1; position < len(outcome.rows); position++ {
		previous := outcome.rows[position-1][key]
		current := outcome.rows[position][key]

		order := compareGridValues(previous, current)
		if combination.sort.descending {
			order = -order
		}

		if order > 0 {
			t.Errorf("%s: rows %d and %d are out of order by %s: %v, %v", outcome.database.name, position, position+1, key, previous, current)
		}
	}
}

func compareGridValues(first any, second any) int {
	if first == nil || second == nil {
		switch {
		case first == nil && second == nil:
			return 0
		case first == nil:
			return -1
		default:
			return 1
		}
	}

	firstNumber, isNumber := first.(float64)
	if isNumber {
		secondNumber, _ := second.(float64)

		switch {
		case firstNumber < secondNumber:
			return -1
		case firstNumber > secondNumber:
			return 1
		}

		return 0
	}

	return strings.Compare(strings.ToLower(fmt.Sprint(first)), strings.ToLower(fmt.Sprint(second)))
}

// gridTimeSQL is a time worked out by hand: a day 30 days apart for each id,
// written in each database's own way.
func gridTimeSQL(databaseType DatabaseType) string {
	if databaseType == DatabasePostgres {
		return "TIMESTAMP '2025-01-27' + ({DATASET}.id * 30) * INTERVAL '1 day'"
	}

	return "TIMESTAMP('2025-01-27') + INTERVAL ({DATASET}.id * 30) DAY"
}

// shopSalesForGrid adds fields written by hand to the shop sales.
func shopSalesForGrid(databaseType DatabaseType) Dataset {
	dataset := shopSalesDataset()

	dataset.Dimensions = append(dataset.Dimensions,
		Dimension{Name: "size", Type: FieldTypeText, SQL: "CASE WHEN {DATASET}.amount >= 50 THEN 'big' ELSE 'small' END"},
		Dimension{Name: "web_or_large", Type: FieldTypeYesNo, SQL: "{DATASET}.channel = 'web' OR {DATASET}.amount > 100"},
		Dimension{Name: "doubled", Type: FieldTypeNumber, SQL: "{DATASET}.quantity * 2 + 1"},
		Dimension{Name: "sold_at", Type: FieldTypeTime, SQL: gridTimeSQL(databaseType)},
	)

	return dataset
}

func shopModelsForGrid(databaseType DatabaseType) Models {
	models := shopModels()

	for position, dataset := range models.Datasets {
		if dataset.Name == "shop_sales" {
			models.Datasets[position] = shopSalesForGrid(databaseType)
		}
	}

	view := View{
		Name: "grid_view",
		Parts: []ViewPart{
			{Path: []string{"shop_sales"}},
			{Path: []string{"shop_sales", "shops"}, ChosenFields: []string{"name", "format", "total_floor_area"}, AddPrefix: true},
			{Path: []string{"shop_sales", "shops", "shop_regions"}, ChosenFields: []string{"name"}, AddPrefix: true},
		},
	}
	models.Views = append(models.Views, view)

	return models
}

// shopGridFilters are the filters on the shop models. salesName is where the
// shop sales' fields are asked for: the dataset itself or the view.
func shopGridFilters(salesName string, shopNameField string, measure string) []gridFilter {
	return []gridFilter{
		{name: "no filter"},
		{name: "text equals", filters: []Filter{{Field: shopNameField, Operator: FilterEquals, Values: []string{"Riga Centre"}}}},
		{name: "text not equals", filters: []Filter{{Field: shopNameField, Operator: FilterNotEquals, Values: []string{"Riga Centre"}}}},
		{name: "text contains", filters: []Filter{{Field: shopNameField, Operator: FilterContains, Values: []string{"RIGA"}}}},
		{name: "written text equals", filters: []Filter{{Field: salesName + ".size", Operator: FilterEquals, Values: []string{"big"}}}},
		{name: "written number greater", filters: []Filter{{Field: salesName + ".doubled", Operator: FilterGreaterThan, Values: []string{"4"}}}},
		{name: "number contains", filters: []Filter{{Field: salesName + ".amount", Operator: FilterContains, Values: []string{"0.0"}}}},
		{name: "written yes/no with OR", filters: []Filter{{Field: salesName + ".web_or_large", Operator: FilterEquals, Values: []string{"true"}}}},
		{name: "date range", filters: []Filter{{Field: salesName + ".sold_at", Operator: FilterInDateRange, Values: []string{"2025-05-01", "2025-07-26"}}}},
		{name: "after date", filters: []Filter{{Field: salesName + ".sold_at", Operator: FilterAfterDate, Values: []string{"2025-07-26"}}}},
		{name: "not set", filters: []Filter{{Field: shopNameField, Operator: FilterIsNotSet}}},
		{name: "measure greater", filters: []Filter{{Field: measure, Operator: FilterGreaterThan, Values: []string{"1"}}}, onMeasure: true},
		{name: "two filters", filters: []Filter{
			{Field: shopNameField, Operator: FilterContains, Values: []string{"riga"}},
			{Field: salesName + ".web_or_large", Operator: FilterEquals, Values: []string{"true"}},
		}},
	}
}

func joinGridShape() gridShape {
	return gridShape{
		name:       "joins",
		models:     shopModelsForGrid,
		dimensions: []string{"shops.format", "shop_sales.size"},
		measures:   []string{"shop_sales.count", "shop_sales.revenue"},
		timeField:  "shop_sales.sold_at",
		notShown:   "shop_sales.channel",
		filters:    shopGridFilters("shop_sales", "shops.name", "shop_sales.count"),
	}
}

// combinedGridShape has measures of two datasets, which a combined query
// works out apart.
func combinedGridShape() gridShape {
	return gridShape{
		name:       "combined",
		models:     shopModelsForGrid,
		dimensions: []string{"shops.format"},
		measures:   []string{"shop_sales.revenue", "shops.total_floor_area"},
		timeField:  "shop_sales.sold_at",
		notShown:   "shop_sales.channel",
		filters:    shopGridFilters("shop_sales", "shops.name", "shop_sales.revenue"),
	}
}

func viewGridShape() gridShape {
	return gridShape{
		name:       "view",
		models:     shopModelsForGrid,
		dimensions: []string{"grid_view.shops_format", "grid_view.size"},
		measures:   []string{"grid_view.count", "grid_view.revenue"},
		timeField:  "grid_view.sold_at",
		notShown:   "grid_view.channel",
		filters:    shopGridFilters("grid_view", "grid_view.shops_name", "grid_view.count"),
	}
}

func eventsGridShape() gridShape {
	models := func(databaseType DatabaseType) Models {
		dataset := eventsDataset()
		dataset.Dimensions = append(dataset.Dimensions,
			Dimension{Name: "half", Type: FieldTypeText, SQL: "CASE WHEN {DATASET}.id <= 3 THEN 'first' ELSE 'second' END"},
			Dimension{Name: "late_or_first", Type: FieldTypeYesNo, SQL: "{DATASET}.id > 4 OR {DATASET}.id = 1"},
		)

		return Models{Datasets: []Dataset{dataset}}
	}

	filters := []gridFilter{
		{name: "no filter"},
		{name: "written text equals", filters: []Filter{{Field: "events.half", Operator: FilterEquals, Values: []string{"first"}}}},
		{name: "number greater", filters: []Filter{{Field: "events.id", Operator: FilterGreaterThan, Values: []string{"2"}}}},
		{name: "number contains", filters: []Filter{{Field: "events.id", Operator: FilterContains, Values: []string{"1"}}}},
		{name: "time contains", filters: []Filter{{Field: "events.happened_at", Operator: FilterStartsWith, Values: []string{"2025-03-03"}}}},
		{name: "written yes/no with OR", filters: []Filter{{Field: "events.late_or_first", Operator: FilterEquals, Values: []string{"true"}}}},
		{name: "date range", filters: []Filter{{Field: "events.happened_at", Operator: FilterInDateRange, Values: []string{"2025-03-02", "2025-03-31"}}}},
		{name: "after date", filters: []Filter{{Field: "events.happened_at", Operator: FilterAfterDate, Values: []string{"2025-03-03"}}}},
		{name: "not set", filters: []Filter{{Field: "events.happened_at", Operator: FilterIsNotSet}}},
		{name: "measure greater", filters: []Filter{{Field: "events.count", Operator: FilterGreaterThan, Values: []string{"1"}}}, onMeasure: true},
		{name: "two filters", filters: []Filter{
			{Field: "events.half", Operator: FilterEquals, Values: []string{"second"}},
			{Field: "events.late_or_first", Operator: FilterEquals, Values: []string{"true"}},
		}},
	}

	return gridShape{
		name:       "plain time",
		models:     models,
		dimensions: []string{"events.half"},
		measures:   []string{"events.count"},
		timeField:  "events.happened_at",
		notShown:   "events.id",
		filters:    filters,
	}
}
