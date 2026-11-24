// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package queryengine

import (
	"errors"
	"slices"
	"strings"
	"testing"
)

func buildSalesQuery(t *testing.T, request Request) BuiltQuery {
	t.Helper()

	query, err := BuildQuery(salesModels(), request)
	if err != nil {
		t.Fatalf("BuildQuery: %v", err)
	}

	return query
}

func checkSQL(t *testing.T, query BuiltQuery, wantSQL string, wantValues []any) {
	t.Helper()

	if query.SQL != wantSQL {
		t.Errorf("want SQL\n%s\ngot\n%s", wantSQL, query.SQL)
	}

	if !slices.Equal(query.Values, wantValues) {
		t.Errorf("want values %v, got %v", wantValues, query.Values)
	}
}

func TestBuildQueryMeasuresByField(t *testing.T) {
	request := Request{
		Dimensions: []string{"sales.country"},
		Measures:   []string{"sales.count", "sales.revenue"},
		RowLimit:   100,
	}

	query := buildSalesQuery(t, request)

	wantSQL := "SELECT `sales`.`country` AS `sales__country`, COUNT(`sales`.`id`) AS `sales__count`, SUM(`sales`.`amount`) AS `sales__revenue`\n" +
		"FROM `sales` AS `sales`\n" +
		"GROUP BY 1\n" +
		"LIMIT 100"

	checkSQL(t, query, wantSQL, nil)

	if query.ConnectionID != "connection-1" {
		t.Errorf("want connection-1, got %q", query.ConnectionID)
	}

	wantColumns := []Column{
		{Key: "sales.country", Title: "Country", Kind: FieldKindDimension, Type: FieldTypeText},
		{Key: "sales.count", Title: "Sales", Kind: FieldKindMeasure, Type: FieldTypeNumber, MeasureType: MeasureCount},
		{Key: "sales.revenue", Title: "Revenue", Kind: FieldKindMeasure, Type: FieldTypeNumber, MeasureType: MeasureSum},
	}

	if !slices.Equal(query.Columns, wantColumns) {
		t.Errorf("want columns %v, got %v", wantColumns, query.Columns)
	}
}

func TestBuildQueryEveryMeasureType(t *testing.T) {
	cases := []struct {
		measure    string
		wantSelect string
	}{
		{"sales.count", "COUNT(`sales`.`id`)"},
		{"sales.countries", "COUNT(DISTINCT `sales`.`country`)"},
		{"sales.revenue", "SUM(`sales`.`amount`)"},
		{"sales.average_sale", "AVG(`sales`.`amount`)"},
		{"sales.smallest_sale", "MIN(`sales`.`amount`)"},
		{"sales.largest_sale", "MAX(`sales`.`amount`)"},
		{"sales.profit", "SUM((amount - cost))"},
		{"sales.margin", "(ROUND(100 * SUM(`sales`.amount - `sales`.cost) / NULLIF(SUM(`sales`.amount), 0), 1))"},
	}

	for _, testCase := range cases {
		t.Run(testCase.measure, func(t *testing.T) {
			request := Request{Measures: []string{testCase.measure}, RowLimit: 10}

			query := buildSalesQuery(t, request)

			alias := mysqlDatabase{}.quoteName(requestedField{key: testCase.measure}.alias())
			wantSQL := "SELECT " + testCase.wantSelect + " AS " + alias + "\n" +
				"FROM `sales` AS `sales`\n" +
				"LIMIT 10"

			checkSQL(t, query, wantSQL, nil)
		})
	}
}

func TestBuildQueryCountsEveryRowWithoutPrimaryKey(t *testing.T) {
	dataset := salesDataset()
	dataset.Dimensions = dataset.Dimensions[1:]

	models := Models{Datasets: []Dataset{dataset}, DatabaseTypes: testDatabaseTypes()}
	request := Request{Measures: []string{"sales.count"}, RowLimit: 10}

	query, err := BuildQuery(models, request)
	if err != nil {
		t.Fatalf("BuildQuery: %v", err)
	}

	wantSQL := "SELECT COUNT(*) AS `sales__count`\n" +
		"FROM `sales` AS `sales`\n" +
		"LIMIT 10"

	checkSQL(t, query, wantSQL, nil)
}

func TestBuildQueryFilterTypes(t *testing.T) {
	cases := []struct {
		name          string
		filter        Filter
		wantCondition string
		wantValues    []any
	}{
		{
			"equals one value",
			Filter{Field: "sales.country", Operator: FilterEquals, Values: []string{"Latvia"}},
			"`sales`.`country` = ?",
			[]any{"Latvia"},
		},
		{
			"equals several values",
			Filter{Field: "sales.country", Operator: FilterEquals, Values: []string{"Latvia", "Poland"}},
			"`sales`.`country` IN (?, ?)",
			[]any{"Latvia", "Poland"},
		},
		{
			"not equals one value",
			Filter{Field: "sales.country", Operator: FilterNotEquals, Values: []string{"Latvia"}},
			"(`sales`.`country` <> ? OR `sales`.`country` IS NULL)",
			[]any{"Latvia"},
		},
		{
			"not equals several values",
			Filter{Field: "sales.country", Operator: FilterNotEquals, Values: []string{"Latvia", "Poland"}},
			"(`sales`.`country` NOT IN (?, ?) OR `sales`.`country` IS NULL)",
			[]any{"Latvia", "Poland"},
		},
		{
			"contains",
			Filter{Field: "sales.note", Operator: FilterContains, Values: []string{"50%", "a_b"}},
			"(LOWER(`sales`.`note`) LIKE CONCAT('%', LOWER(?), '%') ESCAPE '!' OR LOWER(`sales`.`note`) LIKE CONCAT('%', LOWER(?), '%') ESCAPE '!')",
			[]any{"50!%", "a!_b"},
		},
		{
			"does not contain",
			Filter{Field: "sales.note", Operator: FilterNotContains, Values: []string{"x", "y!"}},
			"((LOWER(`sales`.`note`) NOT LIKE CONCAT('%', LOWER(?), '%') ESCAPE '!' AND LOWER(`sales`.`note`) NOT LIKE CONCAT('%', LOWER(?), '%') ESCAPE '!') OR `sales`.`note` IS NULL)",
			[]any{"x", "y!!"},
		},
		{
			"starts with",
			Filter{Field: "sales.note", Operator: FilterStartsWith, Values: []string{"fr"}},
			"(LOWER(`sales`.`note`) LIKE CONCAT('', LOWER(?), '%') ESCAPE '!')",
			[]any{"fr"},
		},
		{
			"ends with",
			Filter{Field: "sales.note", Operator: FilterEndsWith, Values: []string{"ee"}},
			"(LOWER(`sales`.`note`) LIKE CONCAT('%', LOWER(?), '') ESCAPE '!')",
			[]any{"ee"},
		},
		{
			"greater than uses only the first value",
			Filter{Field: "sales.quantity", Operator: FilterGreaterThan, Values: []string{"2", "9"}},
			"`sales`.`quantity` > ?",
			[]any{"2"},
		},
		{
			"greater or equal",
			Filter{Field: "sales.quantity", Operator: FilterGreaterOrEqual, Values: []string{"2"}},
			"`sales`.`quantity` >= ?",
			[]any{"2"},
		},
		{
			"less than",
			Filter{Field: "sales.quantity", Operator: FilterLessThan, Values: []string{"2"}},
			"`sales`.`quantity` < ?",
			[]any{"2"},
		},
		{
			"less or equal",
			Filter{Field: "sales.quantity", Operator: FilterLessOrEqual, Values: []string{"2"}},
			"`sales`.`quantity` <= ?",
			[]any{"2"},
		},
		{
			"is set",
			Filter{Field: "sales.country", Operator: FilterIsSet},
			"`sales`.`country` IS NOT NULL",
			nil,
		},
		{
			"is not set",
			Filter{Field: "sales.country", Operator: FilterIsNotSet},
			"`sales`.`country` IS NULL",
			nil,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			request := Request{
				Measures: []string{"sales.count"},
				Filters:  []Filter{testCase.filter},
				RowLimit: 10,
			}

			query := buildSalesQuery(t, request)

			wantSQL := "SELECT COUNT(`sales`.`id`) AS `sales__count`\n" +
				"FROM `sales` AS `sales`\n" +
				"WHERE " + testCase.wantCondition + "\n" +
				"LIMIT 10"

			checkSQL(t, query, wantSQL, testCase.wantValues)
		})
	}
}

func TestBuildQueryFiltersOnMeasuresGoAfterAddingUp(t *testing.T) {
	request := Request{
		Dimensions: []string{"sales.country"},
		Measures:   []string{"sales.count"},
		Filters: []Filter{
			{Field: "sales.revenue", Operator: FilterGreaterThan, Values: []string{"100"}},
			{Field: "sales.channel", Operator: FilterEquals, Values: []string{"Shop"}},
		},
		RowLimit: 10,
	}

	query := buildSalesQuery(t, request)

	wantSQL := "SELECT `sales`.`country` AS `sales__country`, COUNT(`sales`.`id`) AS `sales__count`\n" +
		"FROM `sales` AS `sales`\n" +
		"WHERE `sales`.`channel` = ?\n" +
		"GROUP BY 1\n" +
		"HAVING SUM(`sales`.`amount`) > ?\n" +
		"LIMIT 10"

	checkSQL(t, query, wantSQL, []any{"Shop", "100"})
}

func TestBuildQueryFilterGroups(t *testing.T) {
	request := Request{
		Measures: []string{"sales.count"},
		Filters: []Filter{
			{
				AllOf: []Filter{
					{Field: "sales.channel", Operator: FilterEquals, Values: []string{"Shop"}},
					{
						AnyOf: []Filter{
							{Field: "sales.country", Operator: FilterIsNotSet},
							{Field: "sales.quantity", Operator: FilterGreaterThan, Values: []string{"2"}},
						},
					},
				},
			},
		},
		RowLimit: 10,
	}

	query := buildSalesQuery(t, request)

	wantSQL := "SELECT COUNT(`sales`.`id`) AS `sales__count`\n" +
		"FROM `sales` AS `sales`\n" +
		"WHERE (`sales`.`channel` = ? AND (`sales`.`country` IS NULL OR `sales`.`quantity` > ?))\n" +
		"LIMIT 10"

	checkSQL(t, query, wantSQL, []any{"Shop", "2"})
}

func TestBuildQuerySorting(t *testing.T) {
	request := Request{
		Dimensions: []string{"sales.country"},
		Measures:   []string{"sales.count"},
		SortBy: []SortField{
			{Field: "sales.count", Descending: true},
			{Field: "sales.revenue"},
		},
		RowLimit: 10,
	}

	query := buildSalesQuery(t, request)

	wantSQL := "SELECT `sales`.`country` AS `sales__country`, COUNT(`sales`.`id`) AS `sales__count`\n" +
		"FROM `sales` AS `sales`\n" +
		"GROUP BY 1\n" +
		"ORDER BY `sales__count` DESC, SUM(`sales`.`amount`) ASC\n" +
		"LIMIT 10"

	checkSQL(t, query, wantSQL, nil)
}

func TestBuildQueryRowsOnly(t *testing.T) {
	request := Request{
		Dimensions: []string{"sales.country", "sales.channel"},
		Filters: []Filter{
			{Field: "sales.channel", Operator: FilterGreaterThan, Values: []string{"M"}},
		},
		SortBy:   []SortField{{Field: "sales.channel"}},
		RowLimit: 10,
		RowsOnly: true,
	}

	query := buildSalesQuery(t, request)

	wantSQL := "SELECT `sales`.`country` AS `sales__country`, `sales`.`channel` AS `sales__channel`\n" +
		"FROM `sales` AS `sales`\n" +
		"WHERE `sales`.`channel` > ?\n" +
		"ORDER BY `sales__channel` ASC\n" +
		"LIMIT 10"

	checkSQL(t, query, wantSQL, []any{"M"})
}

func TestBuildQueryTableWithDatabaseName(t *testing.T) {
	dataset := salesDataset()
	dataset.Table = "shop.sales"

	models := Models{Datasets: []Dataset{dataset}, DatabaseTypes: testDatabaseTypes()}
	request := Request{Measures: []string{"sales.count"}, RowLimit: 10}

	query, err := BuildQuery(models, request)
	if err != nil {
		t.Fatalf("BuildQuery: %v", err)
	}

	wantSQL := "SELECT COUNT(`sales`.`id`) AS `sales__count`\n" +
		"FROM `shop`.`sales` AS `sales`\n" +
		"LIMIT 10"

	checkSQL(t, query, wantSQL, nil)
}

func TestBuildQueryErrors(t *testing.T) {
	brokenDataset := salesDataset()
	brokenDataset.Name = "broken"
	brokenDataset.Joins = []Join{{OtherDataset: "missing", Relationship: ManyToOne}}

	badTableDataset := salesDataset()
	badTableDataset.Name = "bad_table"
	badTableDataset.Table = "sales; DROP TABLE sales"

	timeDataset := salesDataset()
	timeDataset.Name = "dated"
	timeDataset.Dimensions = append(timeDataset.Dimensions, Dimension{Name: "sold_at", Type: FieldTypeTime, SQL: "sold_at"})
	timeDataset.Measures = append(timeDataset.Measures, Measure{Name: "total_without_sql", Type: MeasureSum})
	timeDataset.Measures = append(timeDataset.Measures, Measure{Name: "median_amount", Type: "median", SQL: "amount"})

	view := View{Name: "sales_view", Parts: []ViewPart{{Path: []string{"sales"}}}}

	models := Models{
		Datasets:      []Dataset{salesDataset(), brokenDataset, badTableDataset, timeDataset},
		Views:         []View{view},
		DatabaseTypes: testDatabaseTypes(),
	}

	countOnly := []string{"sales.count"}

	cases := []struct {
		name      string
		request   Request
		wantCode  ErrorCode
		wantField string
	}{
		{"no fields", Request{RowLimit: 10}, ErrorNoFields, ""},
		{"no row limit", Request{Measures: countOnly}, ErrorMissingRowLimit, ""},
		{"measures in rows only", Request{Measures: countOnly, RowLimit: 10, RowsOnly: true}, ErrorMeasuresInRowsOnly, "sales.count"},
		{"unknown dataset", Request{Measures: []string{"nothing.count"}, RowLimit: 10}, ErrorUnknownField, "nothing.count"},
		{"unknown field", Request{Measures: []string{"sales.nothing"}, RowLimit: 10}, ErrorUnknownField, "sales.nothing"},
		{"key without dot", Request{Measures: []string{"count"}, RowLimit: 10}, ErrorUnknownField, "count"},
		{"measure asked as dimension", Request{Dimensions: countOnly, RowLimit: 10}, ErrorWrongFieldKind, "sales.count"},
		{
			"wrong kind before an unknown sort field",
			Request{Dimensions: countOnly, SortBy: []SortField{{Field: "sales.nothing"}}, RowLimit: 10},
			ErrorWrongFieldKind, "sales.count",
		},
		{
			"field shown twice before a filter on a view",
			Request{
				Dimensions: []string{"sales.country", "sales.country"},
				Filters:    []Filter{{Field: "sales_view.country", Operator: FilterIsSet}},
				RowLimit:   10,
			},
			ErrorFieldShownTwice, "sales.country",
		},
		{"dimension asked as measure", Request{Measures: []string{"sales.country"}, RowLimit: 10}, ErrorWrongFieldKind, "sales.country"},
		{"dataset with errors", Request{Measures: []string{"broken.count"}, RowLimit: 10}, ErrorDatasetHasErrors, "broken"},
		{"unsafe table name", Request{Measures: []string{"bad_table.count"}, RowLimit: 10}, ErrorInvalidTableName, "sales; DROP TABLE sales"},
		{"measure without SQL", Request{Measures: []string{"dated.total_without_sql"}, RowLimit: 10}, ErrorFieldHasNoSQL, "dated.total_without_sql"},
		{"unknown measure type", Request{Measures: []string{"dated.median_amount"}, RowLimit: 10}, ErrorUnknownMeasureType, "dated.median_amount"},
		{"two datasets without a join", Request{Measures: []string{"sales.count", "dated.count"}, RowLimit: 10}, ErrorNoJoinPath, "sales, dated"},
		{
			"unknown filter type",
			Request{Measures: countOnly, RowLimit: 10, Filters: []Filter{{Field: "sales.country", Operator: "between"}}},
			ErrorUnknownFilterType, "sales.country",
		},
		{
			"missing filter value",
			Request{Measures: countOnly, RowLimit: 10, Filters: []Filter{{Field: "sales.country", Operator: FilterEquals}}},
			ErrorMissingFilterValue, "sales.country",
		},
		{
			"empty filter",
			Request{Measures: countOnly, RowLimit: 10, Filters: []Filter{{}}},
			ErrorBadFilterGroup, "",
		},
		{
			"group with both all and any",
			Request{Measures: countOnly, RowLimit: 10, Filters: []Filter{{
				AllOf: []Filter{{Field: "sales.country", Operator: FilterIsSet}},
				AnyOf: []Filter{{Field: "sales.country", Operator: FilterIsSet}},
			}}},
			ErrorBadFilterGroup, "",
		},
		{
			"group mixing fields and measures",
			Request{Measures: countOnly, RowLimit: 10, Filters: []Filter{{
				AnyOf: []Filter{
					{Field: "sales.country", Operator: FilterIsSet},
					{Field: "sales.revenue", Operator: FilterGreaterThan, Values: []string{"1"}},
				},
			}}},
			ErrorMixedFilterGroup, "sales.revenue",
		},
		{
			"filter on measure in rows only",
			Request{Dimensions: []string{"sales.country"}, RowLimit: 10, RowsOnly: true, Filters: []Filter{
				{Field: "sales.revenue", Operator: FilterGreaterThan, Values: []string{"1"}},
			}},
			ErrorMeasuresInRowsOnly, "sales.revenue",
		},
		{
			"sort by measure in rows only",
			Request{Dimensions: []string{"sales.country"}, RowLimit: 10, RowsOnly: true, SortBy: []SortField{{Field: "sales.revenue"}}},
			ErrorMeasuresInRowsOnly, "sales.revenue",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := BuildQuery(models, testCase.request)

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

// SQL written by hand is put in brackets, so that its OR cannot escape the
// filter's AND and match rows the other filter left out.
func TestBuildQueryPutsWrittenSQLInBrackets(t *testing.T) {
	dataset := salesDataset()
	dataset.Dimensions = append(dataset.Dimensions, Dimension{
		Name: "shop_or_estonia",
		Type: FieldTypeYesNo,
		SQL:  "{DATASET}.channel = 'Shop' OR {DATASET}.country = 'Estonia'",
	})
	models := Models{Datasets: []Dataset{dataset}, DatabaseTypes: testDatabaseTypes()}

	request := Request{
		Measures: []string{"sales.count"},
		Filters: []Filter{
			{Field: "sales.country", Operator: FilterEquals, Values: []string{"Latvia"}},
			{Field: "sales.shop_or_estonia", Operator: FilterEquals, Values: []string{"true"}},
		},
		RowLimit: 10,
	}

	query, err := BuildQuery(models, request)
	if err != nil {
		t.Fatalf("BuildQuery: %v", err)
	}

	wantSQL := "SELECT COUNT(`sales`.`id`) AS `sales__count`\n" +
		"FROM `sales` AS `sales`\n" +
		"WHERE `sales`.`country` = ? AND (`sales`.channel = 'Shop' OR `sales`.country = 'Estonia') = ?\n" +
		"LIMIT 10"

	checkSQL(t, query, wantSQL, []any{"Latvia", "1"})
}

// Grouped rows have no single value of a field they are not grouped by, so
// sorting by one is refused. Rows that are not grouped can be sorted by any
// field, and grouped rows by any measure.
func TestBuildQuerySortsGroupedRowsOnlyByWhatTheyShow(t *testing.T) {
	grouped := Request{
		Dimensions: []string{"sales.country"},
		Measures:   []string{"sales.count"},
		SortBy:     []SortField{{Field: "sales.channel"}},
		RowLimit:   10,
	}

	_, err := BuildQuery(salesModels(), grouped)

	checkRequestError(t, err, ErrorSortFieldNotShown)

	rowsOnly := Request{
		Dimensions: []string{"sales.country"},
		SortBy:     []SortField{{Field: "sales.channel"}},
		RowsOnly:   true,
		RowLimit:   10,
	}

	if _, err := BuildQuery(salesModels(), rowsOnly); err != nil {
		t.Errorf("rows that are not grouped: %v", err)
	}

	byMeasure := Request{
		Dimensions: []string{"sales.country"},
		SortBy:     []SortField{{Field: "sales.revenue", Descending: true}},
		RowLimit:   10,
	}

	if _, err := BuildQuery(salesModels(), byMeasure); err != nil {
		t.Errorf("grouped rows by a measure not shown: %v", err)
	}
}

func checkRequestError(t *testing.T, err error, wantCode ErrorCode) {
	t.Helper()

	var requestError *RequestError
	if !errors.As(err, &requestError) || requestError.Code != wantCode {
		t.Errorf("want %s, got %v", wantCode, err)
	}
}

// a.b__c and a__b.c would both be the column a__b__c.
func TestBuildQueryRefusesColumnNamesThatClash(t *testing.T) {
	first := Dataset{
		Name: "a", Table: "a", ConnectionID: "connection-1",
		Dimensions: []Dimension{{Name: "id", Type: FieldTypeNumber, SQL: "id", IsPrimaryKey: true}, {Name: "b__c", Type: FieldTypeText, SQL: "c"}},
		Joins:      []Join{{OtherDataset: "a__b", Relationship: ManyToOne, Column: "b_id", OtherColumn: "id"}},
	}
	second := Dataset{
		Name: "a__b", Table: "b", ConnectionID: "connection-1",
		Dimensions: []Dimension{{Name: "id", Type: FieldTypeNumber, SQL: "id", IsPrimaryKey: true}, {Name: "c", Type: FieldTypeText, SQL: "c"}},
	}
	models := Models{Datasets: []Dataset{first, second}, DatabaseTypes: testDatabaseTypes()}

	request := Request{Dimensions: []string{"a.b__c", "a__b.c"}, RowLimit: 10}

	_, err := BuildQuery(models, request)

	checkRequestError(t, err, ErrorColumnNamesClash)
}

// Postgres keeps the first 63 bytes of a name, so two long names alike that
// far are one name there.
func TestBuildQueryRefusesLongColumnNamesAlikeIn63Bytes(t *testing.T) {
	sameStart := strings.Repeat("x", 60)

	dataset := salesDataset()
	dataset.Dimensions = append(dataset.Dimensions,
		Dimension{Name: sameStart + "_one", Type: FieldTypeText, SQL: "country"},
		Dimension{Name: sameStart + "_two", Type: FieldTypeText, SQL: "channel"},
	)
	models := Models{Datasets: []Dataset{dataset}, DatabaseTypes: testDatabaseTypes()}

	request := Request{Dimensions: []string{"sales." + sameStart + "_one", "sales." + sameStart + "_two"}, RowLimit: 10}

	_, err := BuildQuery(models, request)

	checkRequestError(t, err, ErrorColumnNamesClash)
}
