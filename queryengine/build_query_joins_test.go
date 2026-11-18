// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package queryengine

import (
	"errors"
	"slices"
	"strings"
	"testing"
)

func buildShopQuery(t *testing.T, request Request) BuiltQuery {
	t.Helper()

	query, err := BuildQuery(shopModels(), request)
	if err != nil {
		t.Fatalf("BuildQuery: %v", err)
	}

	return query
}

func TestBuildQueryJoinsAnotherDataset(t *testing.T) {
	request := Request{
		Dimensions: []string{"shops.format"},
		Measures:   []string{"shop_sales.revenue"},
		RowLimit:   100,
	}

	query := buildShopQuery(t, request)

	wantSQL := "SELECT `shops`.`format` AS `shops__format`, SUM(`shop_sales`.`amount`) AS `shop_sales__revenue`\n" +
		"FROM `store_sales` AS `shop_sales`\n" +
		"LEFT JOIN `stores` AS `shops` ON `shop_sales`.`store_id` = `shops`.`id`\n" +
		"GROUP BY 1\n" +
		"LIMIT 100"

	checkSQL(t, query, wantSQL, nil)
}

func TestBuildQueryJoinsThroughDatasetNotAskedFor(t *testing.T) {
	request := Request{
		Dimensions: []string{"shop_regions.name"},
		Measures:   []string{"shop_sales.revenue"},
		RowLimit:   100,
	}

	query := buildShopQuery(t, request)

	wantSQL := "SELECT `shop_regions`.`name` AS `shop_regions__name`, SUM(`shop_sales`.`amount`) AS `shop_sales__revenue`\n" +
		"FROM `store_sales` AS `shop_sales`\n" +
		"LEFT JOIN `stores` AS `shops` ON `shop_sales`.`store_id` = `shops`.`id`\n" +
		"LEFT JOIN `regions` AS `shop_regions` ON `shops`.`region_id` = `shop_regions`.`id`\n" +
		"GROUP BY 1\n" +
		"LIMIT 100"

	checkSQL(t, query, wantSQL, nil)
}

func TestBuildQueryJoinWithWrittenCondition(t *testing.T) {
	request := Request{
		Dimensions: []string{"sale_channels.label"},
		Measures:   []string{"shop_sales.count"},
		RowLimit:   100,
	}

	query := buildShopQuery(t, request)

	wantSQL := "SELECT `sale_channels`.`label` AS `sale_channels__label`, COUNT(`shop_sales`.`id`) AS `shop_sales__count`\n" +
		"FROM `store_sales` AS `shop_sales`\n" +
		"LEFT JOIN `channels` AS `sale_channels` ON `shop_sales`.channel = `sale_channels`.code\n" +
		"GROUP BY 1\n" +
		"LIMIT 100"

	checkSQL(t, query, wantSQL, nil)
}

// shops and shop_regions join each other, so a query using both could start
// from either; findJoinPlan's rule decides.
func TestBuildQueryChoosesWhereTheJoinsStart(t *testing.T) {
	cases := []struct {
		name      string
		request   Request
		wantStart string
	}{
		{
			"dimension before measure",
			Request{Dimensions: []string{"shop_regions.name"}, Measures: []string{"shops.count"}},
			"`regions` AS `shop_regions`",
		},
		{
			"first dimension",
			Request{Dimensions: []string{"shops.format", "shop_regions.name"}, Measures: []string{"shops.count"}},
			"`stores` AS `shops`",
		},
		{
			"sort field before date field",
			Request{RowsOnly: true, SortBy: []SortField{{Field: "shop_regions.name"}}, DateFields: []DateField{{Field: "shops.name"}}},
			"`regions` AS `shop_regions`",
		},
		{
			"date field before filter",
			Request{Measures: []string{"shops.count"}, DateFields: []DateField{{Field: "shops.name"}}, Filters: []Filter{{Field: "shop_regions.name", Operator: FilterIsSet}}},
			"`stores` AS `shops`",
		},
		{
			"filter before measure",
			Request{Measures: []string{"shops.count"}, Filters: []Filter{{Field: "shop_regions.name", Operator: FilterIsSet}}},
			"`regions` AS `shop_regions`",
		},
		{
			"first filter",
			Request{Measures: []string{"shops.count"}, Filters: []Filter{{Field: "shops.format", Operator: FilterIsSet}, {Field: "shop_regions.name", Operator: FilterIsSet}}},
			"`stores` AS `shops`",
		},
		{
			"filter on a field before filter on a measure",
			Request{Measures: []string{"shops.count"}, Filters: []Filter{
				{Field: "shops.count", Operator: FilterGreaterThan, Values: []string{"0"}},
				{Field: "shop_regions.name", Operator: FilterIsSet},
			}},
			"`regions` AS `shop_regions`",
		},
		{
			"only one dataset reaches the other",
			Request{Dimensions: []string{"shops.format"}, Measures: []string{"shop_sales.revenue"}},
			"`store_sales` AS `shop_sales`",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			request := testCase.request
			request.RowLimit = 10

			// The shop models have no date field. Finding the start does not
			// look at field types, so a text field stands in for one.
			if len(request.DateFields) > 0 {
				start := findStartForTest(t, request)
				if start != testCase.wantStart {
					t.Errorf("want start %s, got %s", testCase.wantStart, start)
				}

				return
			}

			query := buildShopQuery(t, request)

			wantFrom := "FROM " + testCase.wantStart + "\n"
			if !strings.Contains(query.SQL, wantFrom) {
				t.Errorf("want %q in\n%s", wantFrom, query.SQL)
			}
		})
	}
}

func findStartForTest(t *testing.T, request Request) string {
	t.Helper()

	builder := newQueryBuilder(shopModels(), request)

	var keys []string
	keys = append(keys, request.Dimensions...)
	keys = append(keys, request.Measures...)

	for _, dateField := range request.DateFields {
		keys = append(keys, dateField.Field)
	}

	for _, key := range keys {
		if _, err := builder.readFieldOnce(key); err != nil {
			t.Fatalf("read %s: %v", key, err)
		}
	}

	if err := builder.readSortAndFilterFields(); err != nil {
		t.Fatalf("readSortAndFilterFields: %v", err)
	}

	plan, err := builder.planJoins(nil)
	if err != nil {
		t.Fatalf("planJoins: %v", err)
	}

	return "`" + plan.start.Table + "` AS `" + plan.start.Name + "`"
}

func TestBuildQueryCountsRepeatedRowsOnce(t *testing.T) {
	cases := []struct {
		name       string
		request    Request
		wantSelect string
	}{
		{
			"many to one repeats the dataset it points to",
			Request{Dimensions: []string{"shops.format"}, Measures: []string{"shop_regions.count"}},
			"COUNT(DISTINCT `shop_regions`.`id`) AS `shop_regions__count`",
		},
		{
			"one to many further away repeats the start",
			Request{Dimensions: []string{"shop_staff.name"}, Measures: []string{"shop_sales.count"}},
			"COUNT(DISTINCT `shop_sales`.`id`) AS `shop_sales__count`",
		},
		{
			"counting distinct values is not changed",
			Request{Dimensions: []string{"shops.format"}, Measures: []string{"shop_regions.names"}},
			"COUNT(DISTINCT `shop_regions`.`name`) AS `shop_regions__names`",
		},
		{
			"one to many from the start does not repeat the other side",
			Request{Dimensions: []string{"shop_regions.name"}, Measures: []string{"shops.count"}},
			"COUNT(`shops`.`id`) AS `shops__count`",
		},
		{
			"in a view",
			Request{Dimensions: []string{"shop_sales_view.channel"}, Measures: []string{"shop_sales_view.shops_count"}},
			"COUNT(DISTINCT `shops`.`id`) AS `shop_sales_view__shops_count`",
		},
		{
			"one to one repeats nothing",
			Request{Dimensions: []string{"shops.format"}, Measures: []string{"shop_managers.count"}},
			"COUNT(`shop_managers`.`id`) AS `shop_managers__count`",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			request := testCase.request
			request.RowLimit = 10

			query := buildShopQuery(t, request)

			if !strings.Contains(query.SQL, testCase.wantSelect) {
				t.Errorf("want %q in\n%s", testCase.wantSelect, query.SQL)
			}
		})
	}
}

func TestBuildQueryViewFollowsItsPaths(t *testing.T) {
	request := Request{
		Dimensions: []string{"shop_sales_view.shop_regions_name"},
		Measures:   []string{"shop_sales_view.revenue", "shop_sales_view.count"},
		RowLimit:   100,
	}

	query := buildShopQuery(t, request)

	wantSQL := "SELECT `shop_regions`.`name` AS `shop_sales_view__shop_regions_name`, " +
		"SUM(`shop_sales`.`amount`) AS `shop_sales_view__revenue`, " +
		"COUNT(`shop_sales`.`id`) AS `shop_sales_view__count`\n" +
		"FROM `store_sales` AS `shop_sales`\n" +
		"LEFT JOIN `stores` AS `shops` ON `shop_sales`.`store_id` = `shops`.`id`\n" +
		"LEFT JOIN `regions` AS `shop_regions` ON `shops`.`region_id` = `shop_regions`.`id`\n" +
		"GROUP BY 1\n" +
		"LIMIT 100"

	checkSQL(t, query, wantSQL, nil)

	wantColumns := []Column{
		{Key: "shop_sales_view.shop_regions_name", Title: "Shop Regions Name", Kind: FieldKindDimension, Type: FieldTypeText},
		{Key: "shop_sales_view.revenue", Title: "Revenue", Kind: FieldKindMeasure, Type: FieldTypeNumber},
		{Key: "shop_sales_view.count", Title: "Count", Kind: FieldKindMeasure, Type: FieldTypeNumber},
	}

	if !slices.Equal(query.Columns, wantColumns) {
		t.Errorf("want columns %v, got %v", wantColumns, query.Columns)
	}
}

// A view field is reached along its path from the view's first dataset, even
// when nothing else from that dataset is asked for.
func TestBuildQueryViewJoinsWholePath(t *testing.T) {
	request := Request{Dimensions: []string{"shop_sales_view.shops_format"}, RowLimit: 100}

	query := buildShopQuery(t, request)

	wantSQL := "SELECT `shops`.`format` AS `shop_sales_view__shops_format`\n" +
		"FROM `store_sales` AS `shop_sales`\n" +
		"LEFT JOIN `stores` AS `shops` ON `shop_sales`.`store_id` = `shops`.`id`\n" +
		"GROUP BY 1\n" +
		"LIMIT 100"

	checkSQL(t, query, wantSQL, nil)
}

func TestBuildQueryViewJoinsOnlyWhatIsUsed(t *testing.T) {
	request := Request{
		Dimensions: []string{"shops_view.format"},
		Measures:   []string{"shops_view.count"},
		RowLimit:   100,
	}

	query := buildShopQuery(t, request)

	wantSQL := "SELECT `shops`.`format` AS `shops_view__format`, COUNT(`shops`.`id`) AS `shops_view__count`\n" +
		"FROM `stores` AS `shops`\n" +
		"GROUP BY 1\n" +
		"LIMIT 100"

	checkSQL(t, query, wantSQL, nil)
}

func TestBuildQueryJoinErrors(t *testing.T) {
	models := shopModels()

	otherConnection := peopleDataset("other_connection")
	otherConnection.ConnectionID = "connection-2"

	brokenOnTheWay := peopleDataset("broken_on_the_way")
	brokenOnTheWay.Joins = []Join{
		{OtherDataset: "shops", Relationship: ManyToOne, Column: "store_id", OtherColumn: "id"},
		{OtherDataset: "missing", Relationship: ManyToOne, Column: "x", OtherColumn: "y"},
	}

	startOfBroken := peopleDataset("start_of_broken")
	startOfBroken.Joins = []Join{
		{OtherDataset: "broken_on_the_way", Relationship: ManyToOne, Column: "id", OtherColumn: "id"},
	}

	badColumns := peopleDataset("bad_columns")
	badColumns.Joins = []Join{
		{OtherDataset: "shops", Relationship: ManyToOne, Column: "store_id; DROP", OtherColumn: "id"},
		{OtherDataset: "shop_regions", Relationship: ManyToOne, CustomCondition: "{DATASET}.region = {shop_regions.id}"},
	}

	loops := peopleDataset("loops")
	loops.Joins = []Join{{OtherDataset: "shops", Relationship: ManyToOne, Column: "store_id", OtherColumn: "id"}}
	loops.Measures = []Measure{
		{Name: "first", Type: MeasureFormula, SQL: "{second} + 1"},
		{Name: "second", Type: MeasureFormula, SQL: "{first} * 2"},
		{Name: "columns_and_other_dataset", Type: MeasureFormula, SQL: "SUM({DATASET}.id) / {shops.total_floor_area}"},
		{Name: "unknown_reference", Type: MeasureFormula, SQL: "{nothing}"},
	}

	keyless := Dataset{
		Name:         "keyless",
		Table:        "regions",
		ConnectionID: "connection-1",
		Dimensions:   []Dimension{{Name: "name", Type: FieldTypeText, SQL: "name"}},
		Measures: []Measure{
			{Name: "count", Type: MeasureCount},
			{Name: "total", Type: MeasureSum, SQL: "id"},
		},
	}

	shops := models.Datasets[1]
	shops.Joins = append(shops.Joins,
		Join{OtherDataset: "other_connection", Relationship: OneToOne, Column: "id", OtherColumn: "store_id"},
		Join{OtherDataset: "keyless", Relationship: ManyToOne, Column: "region_id", OtherColumn: "id"},
	)
	models.Datasets[1] = shops

	models.Datasets = append(models.Datasets, otherConnection, brokenOnTheWay, startOfBroken, badColumns, loops, keyless)

	startsApart := View{Name: "starts_apart", Parts: []ViewPart{
		{Path: []string{"shops"}, AddPrefix: true},
		{Path: []string{"shop_regions"}, AddPrefix: true},
	}}
	models.Views = append(models.Views, startsApart)

	cases := []struct {
		name      string
		request   Request
		wantCode  ErrorCode
		wantField string
	}{
		{
			"no path between datasets",
			Request{Dimensions: []string{"shop_managers.name", "sale_channels.label"}},
			ErrorNoJoinPath, "shop_managers, sale_channels",
		},
		{
			"different connections",
			Request{Dimensions: []string{"shops.name", "other_connection.name"}},
			ErrorDifferentConnections, "other_connection",
		},
		{
			"dataset with errors on the way",
			Request{Dimensions: []string{"start_of_broken.name", "shops.name"}},
			ErrorDatasetHasErrors, "broken_on_the_way",
		},
		{
			"join columns that are not plain names",
			Request{Dimensions: []string{"bad_columns.name", "shops.name"}},
			ErrorBadJoin, "bad_columns → shops",
		},
		{
			"written condition naming a field",
			Request{Dimensions: []string{"bad_columns.name", "shop_regions.name"}},
			ErrorFieldReferenceInSQL, "bad_columns",
		},
		{
			"view with a dataset",
			Request{Dimensions: []string{"shops_view.name", "shops.name"}},
			ErrorViewMixedWithOthers, "shops",
		},
		{
			"view whose parts start from different datasets",
			Request{Dimensions: []string{"starts_apart.shops_format"}, Measures: []string{"starts_apart.shop_regions_count"}},
			ErrorViewPathsStartApart, "starts_apart",
		},
		{
			"dataset with a view",
			Request{Dimensions: []string{"shops.name"}, Filters: []Filter{{Field: "shops_view.name", Operator: FilterIsSet}}},
			ErrorViewMixedWithOthers, "shops_view",
		},
		{
			"two views",
			Request{Dimensions: []string{"shops_view.name", "shop_sales_view.channel"}},
			ErrorViewMixedWithOthers, "shop_sales_view",
		},
		{
			"unknown view field",
			Request{Dimensions: []string{"shops_view.floor_area"}},
			ErrorUnknownField, "shops_view.floor_area",
		},
		{
			"formula referring to itself through another",
			Request{Measures: []string{"loops.first"}},
			ErrorFormulaLoop, "loops.first",
		},
		{
			"formula using another dataset and its own columns",
			Request{Dimensions: []string{"shops.format"}, Measures: []string{"loops.columns_and_other_dataset"}},
			ErrorFormulaMixesDatasets, "loops.columns_and_other_dataset",
		},
		{
			"formula referring to an unknown measure",
			Request{Measures: []string{"loops.unknown_reference"}},
			ErrorUnknownField, "loops.unknown_reference: nothing",
		},
		{
			"repeated rows of a dataset without a primary key",
			Request{Dimensions: []string{"shops.format"}, Measures: []string{"keyless.count"}},
			ErrorNeedsPrimaryKey, "keyless",
		},
		{
			"sum over repeated rows of a dataset without a primary key",
			Request{Dimensions: []string{"shops.format"}, Measures: []string{"keyless.total"}},
			ErrorNeedsPrimaryKey, "keyless",
		},
		{
			"sorting a combined query by a field not shown",
			Request{Dimensions: []string{"shops.format"}, Measures: []string{"shops.count", "shop_managers.count"}, SortBy: []SortField{{Field: "shops.name"}}},
			ErrorSortFieldNotShown, "shops.name",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			request := testCase.request
			request.RowLimit = 10

			_, err := BuildQuery(models, request)

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
