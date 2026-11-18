// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package queryengine

import "testing"

func runShopQuery(t *testing.T, request Request) []databaseResult {
	t.Helper()

	return runOnEachDatabase(t, shopModels(), request)
}

// Sales 9 has no store and sale 10 a store that does not exist. Both are kept,
// with no format, because the join keeps rows that match nothing.
func TestRunQueryJoinKeepsRowsWithoutMatch(t *testing.T) {
	request := Request{
		Dimensions: []string{"shops.format"},
		Measures:   []string{"shop_sales.revenue"},
		SortBy:     []SortField{{Field: "shops.format"}},
		RowLimit:   100,
	}

	results := runShopQuery(t, request)

	wantRows := []map[string]any{
		{"shops.format": nil, "shop_sales.revenue": 35.0},
		{"shops.format": "Large", "shop_sales.revenue": 535.5},
		{"shops.format": "Small", "shop_sales.revenue": 120.0},
	}

	checkSortedRows(t, results, "shops.format", wantRows)
}

func TestRunQueryJoinsThroughDatasetNotAskedFor(t *testing.T) {
	request := Request{
		Dimensions: []string{"shop_regions.name"},
		Measures:   []string{"shop_sales.revenue"},
		SortBy:     []SortField{{Field: "shop_regions.name"}},
		RowLimit:   100,
	}

	results := runShopQuery(t, request)

	wantRows := []map[string]any{
		{"shop_regions.name": nil, "shop_sales.revenue": 105.0},
		{"shop_regions.name": "North", "shop_sales.revenue": 225.5},
		{"shop_regions.name": "South", "shop_sales.revenue": 360.0},
	}

	checkSortedRows(t, results, "shop_regions.name", wantRows)
}

func TestRunQueryJoinWithWrittenCondition(t *testing.T) {
	request := Request{
		Dimensions: []string{"sale_channels.label"},
		Measures:   []string{"shop_sales.count"},
		SortBy:     []SortField{{Field: "sale_channels.label"}},
		RowLimit:   100,
	}

	results := runShopQuery(t, request)

	wantRows := []map[string]any{
		{"sale_channels.label": nil, "shop_sales.count": 1.0},
		{"sale_channels.label": "In store", "shop_sales.count": 5.0},
		{"sale_channels.label": "Website", "shop_sales.count": 4.0},
	}

	checkSortedRows(t, results, "sale_channels.label", wantRows)
}

// Starting from regions keeps East, which has no stores.
func TestRunQueryStartsFromDimensionDataset(t *testing.T) {
	request := Request{
		Dimensions: []string{"shop_regions.name"},
		Measures:   []string{"shops.count"},
		SortBy:     []SortField{{Field: "shop_regions.name"}},
		RowLimit:   100,
	}

	results := runShopQuery(t, request)

	wantRows := []map[string]any{
		{"shop_regions.name": "East", "shops.count": 0.0},
		{"shop_regions.name": "North", "shops.count": 2.0},
		{"shop_regions.name": "South", "shops.count": 1.0},
	}

	checkRows(t, results, wantRows)
}

// North has two stores, so without counting each region once it would be
// counted twice among the large and small stores.
func TestRunQueryCountsRepeatedRowsOnce(t *testing.T) {
	request := Request{
		Dimensions: []string{"shops.format"},
		Measures:   []string{"shop_regions.count"},
		SortBy:     []SortField{{Field: "shops.format"}},
		RowLimit:   100,
	}

	results := runShopQuery(t, request)

	wantRows := []map[string]any{
		{"shops.format": "Large", "shop_regions.count": 2.0},
		{"shops.format": "Small", "shop_regions.count": 1.0},
	}

	checkRows(t, results, wantRows)
}

func TestRunQueryView(t *testing.T) {
	request := Request{
		Dimensions: []string{"shop_sales_view.shop_regions_name"},
		Measures:   []string{"shop_sales_view.revenue", "shop_sales_view.count"},
		SortBy:     []SortField{{Field: "shop_sales_view.shop_regions_name"}},
		RowLimit:   100,
	}

	results := runShopQuery(t, request)

	wantRows := []map[string]any{
		{"shop_sales_view.shop_regions_name": nil, "shop_sales_view.revenue": 105.0, "shop_sales_view.count": 3.0},
		{"shop_sales_view.shop_regions_name": "North", "shop_sales_view.revenue": 225.5, "shop_sales_view.count": 5.0},
		{"shop_sales_view.shop_regions_name": "South", "shop_sales_view.revenue": 360.0, "shop_sales_view.count": 2.0},
	}

	checkSortedRows(t, results, "shop_sales_view.shop_regions_name", wantRows)
}

func TestRunQueryViewRowsOnly(t *testing.T) {
	request := Request{
		Dimensions: []string{"shops_view.name", "shops_view.shop_managers_name"},
		SortBy:     []SortField{{Field: "shops_view.name"}},
		RowsOnly:   true,
		RowLimit:   100,
	}

	results := runShopQuery(t, request)

	wantRows := []map[string]any{
		{"shops_view.name": "Daugavpils", "shops_view.shop_managers_name": "Liga"},
		{"shops_view.name": "Liepaja", "shops_view.shop_managers_name": nil},
		{"shops_view.name": "Riga Airport", "shops_view.shop_managers_name": "Janis"},
		{"shops_view.name": "Riga Centre", "shops_view.shop_managers_name": "Anna"},
		{"shops_view.name": "Ventspils", "shops_view.shop_managers_name": nil},
	}

	checkRows(t, results, wantRows)
}
