// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package queryengine

import "testing"

// Every expected value here is worked out by hand from the rows in
// test_database_test.go. Store floor areas: Riga Centre 500 (Large, North),
// Riga Airport 120 (Small, North), Daugavpils 400 (Large, South), Liepaja 150
// (Small, no region), Ventspils 100 (Small, a region that does not exist).

// Each store is added once per channel it sold in. Adding up the joined rows
// would give 1520 for "shop", because Riga Centre has two shop sales.
//
//	shop: Riga Centre, Riga Airport, Daugavpils (sale 10's store does not exist) = 1020
//	web:  Riga Centre, Riga Airport, Liepaja (sale 9 has no store)              = 770
//	post: Daugavpils                                                             = 400
func TestRunQuerySumOverRepeatedRows(t *testing.T) {
	request := Request{
		Dimensions: []string{"shop_sales.channel"},
		Measures:   []string{"shops.total_floor_area", "shops.largest_floor_area", "shops.average_floor_area", "shops.count"},
		SortBy:     []SortField{{Field: "shop_sales.channel"}},
		RowLimit:   100,
	}

	results := runShopQuery(t, request)

	wantRows := []map[string]any{
		{"shop_sales.channel": "post", "shops.total_floor_area": 400.0, "shops.largest_floor_area": 400.0, "shops.average_floor_area": 400.0, "shops.count": 1.0},
		{"shop_sales.channel": "shop", "shops.total_floor_area": 1020.0, "shops.largest_floor_area": 500.0, "shops.average_floor_area": 340.0, "shops.count": 3.0},
		{"shop_sales.channel": "web", "shops.total_floor_area": 770.0, "shops.largest_floor_area": 500.0, "shops.average_floor_area": 256.6667, "shops.count": 3.0},
	}

	checkRowsOn(t, results, DatabaseMySQL, wantRows)

	// Postgres gives an average many more decimals than MySQL's four.
	wantRows[2]["shops.average_floor_area"] = 256.6666666666667
	checkRowsOn(t, results, DatabasePostgres, wantRows)
}

// A formula over repeated rows is worked out over each store once too:
// "shop" is 1020 / 3 stores.
func TestRunQueryFormulaOverRepeatedRows(t *testing.T) {
	request := Request{
		Dimensions: []string{"shop_sales.channel"},
		Measures:   []string{"shops.area_per_store", "shops.area_written_out"},
		SortBy:     []SortField{{Field: "shop_sales.channel"}},
		RowLimit:   100,
	}

	results := runShopQuery(t, request)

	wantRows := []map[string]any{
		{"shop_sales.channel": "post", "shops.area_per_store": 400.0, "shops.area_written_out": 400.0},
		{"shop_sales.channel": "shop", "shops.area_per_store": 340.0, "shops.area_written_out": 1020.0},
		{"shop_sales.channel": "web", "shops.area_per_store": 256.6667, "shops.area_written_out": 770.0},
	}

	checkRowsOn(t, results, DatabaseMySQL, wantRows)

	// In Postgres a whole number divided by a whole number is a whole number:
	// 770 / 3 is 256. The formula is the model's own SQL, so the database
	// decides.
	wantRows[2]["shops.area_per_store"] = 256.0
	checkRowsOn(t, results, DatabasePostgres, wantRows)
}

// Revenue comes from sales, floor area from stores. No store has an empty
// format, so the empty format, from sales without a store, has no floor area:
// it is empty, not 0.
//
//	empty: revenue 15 + 20 = 35
//	Large: revenue 175.5 + 360 = 535.5, floor area 500 + 400 = 900
//	Small: revenue 50 + 70 = 120,       floor area 120 + 150 + 100 = 370
func TestRunQueryMeasuresFromTwoDatasets(t *testing.T) {
	request := Request{
		Dimensions: []string{"shops.format"},
		Measures:   []string{"shop_sales.revenue", "shops.total_floor_area"},
		SortBy:     []SortField{{Field: "shops.format"}},
		RowLimit:   100,
	}

	results := runShopQuery(t, request)

	wantRows := []map[string]any{
		{"shops.format": nil, "shop_sales.revenue": 35.0, "shops.total_floor_area": nil},
		{"shops.format": "Large", "shop_sales.revenue": 535.5, "shops.total_floor_area": 900.0},
		{"shops.format": "Small", "shop_sales.revenue": 120.0, "shops.total_floor_area": 370.0},
	}

	checkSortedRows(t, results, "shops.format", wantRows)
}

// Large 535.5 / 900 = 0.595, Small 120 / 370 = 0.324324, empty 35 / nothing.
func TestRunQueryFormulaUsingAnotherDataset(t *testing.T) {
	request := Request{
		Dimensions: []string{"shops.format"},
		Measures:   []string{"shop_sales.revenue_per_floor_area"},
		SortBy:     []SortField{{Field: "shops.format"}},
		RowLimit:   100,
	}

	results := runShopQuery(t, request)

	wantRows := []map[string]any{
		{"shops.format": nil, "shop_sales.revenue_per_floor_area": nil},
		{"shops.format": "Large", "shop_sales.revenue_per_floor_area": 0.595},
		{"shops.format": "Small", "shop_sales.revenue_per_floor_area": 0.324324},
	}

	checkRowsOn(t, results, DatabaseMySQL, wantRows)

	// Postgres gives a division more decimals than MySQL's six, and puts the
	// empty format last.
	postgresRows := []map[string]any{
		{"shops.format": "Large", "shop_sales.revenue_per_floor_area": 0.595},
		{"shops.format": "Small", "shop_sales.revenue_per_floor_area": 0.32432432432432434},
		{"shops.format": nil, "shop_sales.revenue_per_floor_area": nil},
	}

	checkRowsOn(t, results, DatabasePostgres, postgresRows)
}

// The formula belongs to sales but uses only the stores' floor area, whose rows
// repeat per channel: twice the floor area in TestRunQuerySumOverRepeatedRows.
func TestRunQueryFormulaUsingOnlyAnotherDataset(t *testing.T) {
	request := Request{
		Dimensions: []string{"shop_sales.channel"},
		Measures:   []string{"shop_sales.floor_area_doubled"},
		SortBy:     []SortField{{Field: "shop_sales.channel"}},
		RowLimit:   100,
	}

	results := runShopQuery(t, request)

	wantRows := []map[string]any{
		{"shop_sales.channel": "post", "shop_sales.floor_area_doubled": 800.0},
		{"shop_sales.channel": "shop", "shop_sales.floor_area_doubled": 2040.0},
		{"shop_sales.channel": "web", "shop_sales.floor_area_doubled": 1540.0},
	}

	checkRows(t, results, wantRows)
}

// Empty 35 / 3 items, Large 535.5 / 9, Small 120 / 6.
func TestRunQueryFormulaUsingOwnMeasures(t *testing.T) {
	request := Request{
		Dimensions: []string{"shops.format"},
		Measures:   []string{"shop_sales.revenue_per_item"},
		SortBy:     []SortField{{Field: "shops.format"}},
		RowLimit:   100,
	}

	results := runShopQuery(t, request)

	wantRows := []map[string]any{
		{"shops.format": nil, "shop_sales.revenue_per_item": 11.666667},
		{"shops.format": "Large", "shop_sales.revenue_per_item": 59.5},
		{"shops.format": "Small", "shop_sales.revenue_per_item": 20.0},
	}

	checkRowsOn(t, results, DatabaseMySQL, wantRows)

	// Postgres gives a division more decimals, and puts the empty format last.
	postgresRows := []map[string]any{
		{"shops.format": "Large", "shop_sales.revenue_per_item": 59.5},
		{"shops.format": "Small", "shop_sales.revenue_per_item": 20.0},
		{"shops.format": nil, "shop_sales.revenue_per_item": 11.666666666666666},
	}

	checkRowsOn(t, results, DatabasePostgres, postgresRows)
}

// Only stores in a region that exists count: Riga Centre and Daugavpils
// (Large, both with a manager) and Riga Airport (Small, with a manager). The
// filter on managers and the sort apply to the measures put together.
func TestRunQueryTwoDatasetsWithFiltersSortAndLimit(t *testing.T) {
	request := Request{
		Dimensions: []string{"shops.format"},
		Measures:   []string{"shops.count", "shop_managers.count"},
		Filters: []Filter{
			{Field: "shop_regions.name", Operator: FilterIsSet},
			{Field: "shop_managers.count", Operator: FilterGreaterThan, Values: []string{"0"}},
		},
		SortBy:   []SortField{{Field: "shops.count", Descending: true}},
		RowLimit: 1,
	}

	results := runShopQuery(t, request)

	wantRows := []map[string]any{
		{"shops.format": "Large", "shops.count": 2.0, "shop_managers.count": 2.0},
	}

	checkRows(t, results, wantRows)
}

// Regions per format: Large has North and South, Small only North (Liepaja
// has no region, Ventspils's does not exist). Only Large has more than one,
// and the regions' count is filtered on without being shown.
func TestRunQueryFilterOnAnotherDatasetsMeasure(t *testing.T) {
	request := Request{
		Dimensions: []string{"shops.format"},
		Measures:   []string{"shops.count"},
		Filters:    []Filter{{Field: "shop_regions.count", Operator: FilterGreaterThan, Values: []string{"1"}}},
		RowLimit:   100,
	}

	results := runShopQuery(t, request)

	wantRows := []map[string]any{
		{"shops.format": "Large", "shops.count": 2.0},
	}

	checkRows(t, results, wantRows)
}

func TestRunQueryTwoDatasetsWithNothingShown(t *testing.T) {
	request := Request{
		Measures: []string{"shops.count", "shop_regions.count"},
		RowLimit: 100,
	}

	results := runShopQuery(t, request)

	wantRows := []map[string]any{
		{"shops.count": 5.0, "shop_regions.count": 3.0},
	}

	checkRows(t, results, wantRows)
}

// Revenue by channel: shop 100 + 25.5 + 10 + 300 + 20 = 455.5, web 50 + 40 +
// 70 + 15 = 175, post 60. Floor area as in TestRunQuerySumOverRepeatedRows.
func TestRunQueryViewWithMeasuresFromTwoDatasets(t *testing.T) {
	request := Request{
		Dimensions: []string{"shop_sales_view.channel"},
		Measures:   []string{"shop_sales_view.revenue", "shop_sales_view.shops_total_floor_area"},
		SortBy:     []SortField{{Field: "shop_sales_view.channel"}},
		RowLimit:   100,
	}

	results := runShopQuery(t, request)

	wantRows := []map[string]any{
		{"shop_sales_view.channel": "post", "shop_sales_view.revenue": 60.0, "shop_sales_view.shops_total_floor_area": 400.0},
		{"shop_sales_view.channel": "shop", "shop_sales_view.revenue": 455.5, "shop_sales_view.shops_total_floor_area": 1020.0},
		{"shop_sales_view.channel": "web", "shop_sales_view.revenue": 175.0, "shop_sales_view.shops_total_floor_area": 770.0},
	}

	checkRows(t, results, wantRows)
}
