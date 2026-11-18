// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package queryengine

import (
	"slices"
	"testing"
)

func TestBuildQuerySumOverRepeatedRowsUsesKeys(t *testing.T) {
	request := Request{
		Dimensions: []string{"shop_sales.channel"},
		Measures:   []string{"shops.total_floor_area"},
		RowLimit:   10,
	}

	query := buildShopQuery(t, request)

	wantSQL := "WITH `part_1` AS (\n" +
		"SELECT `distinct_rows`.`shop_sales__channel` AS `shop_sales__channel`, SUM(`shops`.`floor_area`) AS `shops__total_floor_area`\n" +
		"FROM (SELECT DISTINCT `shop_sales`.`channel` AS `shop_sales__channel`, `shops`.`id` AS `distinct_rows__primary_key`\n" +
		"FROM `store_sales` AS `shop_sales`\n" +
		"LEFT JOIN `stores` AS `shops` ON `shop_sales`.`store_id` = `shops`.`id`) AS `distinct_rows`\n" +
		"LEFT JOIN `stores` AS `shops` ON `distinct_rows`.`distinct_rows__primary_key` = `shops`.`id`\n" +
		"GROUP BY 1\n" +
		")\n" +
		"SELECT `part_1`.`shop_sales__channel` AS `shop_sales__channel`, `part_1`.`shops__total_floor_area` AS `shops__total_floor_area`\n" +
		"FROM `part_1`\n" +
		"LIMIT 10"

	checkSQL(t, query, wantSQL, nil)
}

// Each part joins only what it needs: the stores part does not join sales.
func TestBuildQueryFormulaUsingAnotherDatasetCombinesParts(t *testing.T) {
	request := Request{
		Dimensions: []string{"shops.format"},
		Measures:   []string{"shop_sales.revenue_per_floor_area"},
		Filters:    []Filter{{Field: "shops.name", Operator: FilterNotEquals, Values: []string{"Liepaja"}}},
		RowLimit:   10,
	}

	query := buildShopQuery(t, request)

	wantSQL := "WITH `part_1` AS (\n" +
		"SELECT `shops`.`format` AS `shops__format`, SUM(`shop_sales`.`amount`) AS `shop_sales__revenue`\n" +
		"FROM `store_sales` AS `shop_sales`\n" +
		"LEFT JOIN `stores` AS `shops` ON `shop_sales`.`store_id` = `shops`.`id`\n" +
		"WHERE (`shops`.`name` <> ? OR `shops`.`name` IS NULL)\n" +
		"GROUP BY 1\n" +
		"),\n" +
		"`part_2` AS (\n" +
		"SELECT `shops`.`format` AS `shops__format`, SUM(`shops`.`floor_area`) AS `shops__total_floor_area`\n" +
		"FROM `stores` AS `shops`\n" +
		"WHERE (`shops`.`name` <> ? OR `shops`.`name` IS NULL)\n" +
		"GROUP BY 1\n" +
		")\n" +
		"SELECT `all_groups`.`shops__format` AS `shops__format`, " +
		"(`part_1`.`shop_sales__revenue` / NULLIF(`part_2`.`shops__total_floor_area`, 0)) AS `shop_sales__revenue_per_floor_area`\n" +
		"FROM (SELECT `shops__format` FROM `part_1` UNION SELECT `shops__format` FROM `part_2`) AS `all_groups`\n" +
		"LEFT JOIN `part_1` ON `all_groups`.`shops__format` <=> `part_1`.`shops__format`\n" +
		"LEFT JOIN `part_2` ON `all_groups`.`shops__format` <=> `part_2`.`shops__format`\n" +
		"LIMIT 10"

	checkSQL(t, query, wantSQL, []any{"Liepaja", "Liepaja"})
}

func TestBuildQueryFiltersOnCombinedMeasuresComeLast(t *testing.T) {
	request := Request{
		Measures: []string{"shops.count", "shop_regions.count"},
		Filters: []Filter{
			{Field: "shops.count", Operator: FilterGreaterThan, Values: []string{"1"}},
			{Field: "shops.name", Operator: FilterEquals, Values: []string{"Liepaja"}},
		},
		RowLimit: 10,
	}

	query := buildShopQuery(t, request)

	wantValues := []any{"Liepaja", "Liepaja", "1"}
	if !slices.Equal(query.Values, wantValues) {
		t.Errorf("want values %v, got %v", wantValues, query.Values)
	}
}

// A formula referring to another formula keeps it in brackets.
func TestBuildQueryFormulaUsingFormula(t *testing.T) {
	models := shopModels()

	sales := models.Datasets[0]
	sales.Measures = append(sales.Measures,
		Measure{Name: "difference", Type: MeasureFormula, SQL: "{revenue} - {items}"},
		Measure{Name: "double_difference", Type: MeasureFormula, SQL: "{difference} * 2"},
	)
	models.Datasets[0] = sales

	request := Request{Measures: []string{"shop_sales.double_difference"}, RowLimit: 10}

	query, err := BuildQuery(models, request)
	if err != nil {
		t.Fatalf("BuildQuery: %v", err)
	}

	wantSQL := "SELECT ((SUM(`shop_sales`.`amount`) - SUM(`shop_sales`.`quantity`)) * 2) AS `shop_sales__double_difference`\n" +
		"FROM `store_sales` AS `shop_sales`\n" +
		"LIMIT 10"

	checkSQL(t, query, wantSQL, nil)
}
