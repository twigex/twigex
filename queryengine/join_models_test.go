// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package queryengine

// The shop models join the tables made in test_database_test.go:
//
//	shop_sales ──many to one──▶ shops ──many to one──▶ shop_regions
//	     │                       │ ├──one to one──▶ shop_managers
//	     │                       │ └──one to many─▶ shop_staff
//	     └──many to one (written condition)──▶ sale_channels
//	shop_regions ──one to many──▶ shops
//
// shops and shop_regions join each other, so either can be the start.

func shopSalesDataset() Dataset {
	return Dataset{
		Name:         "shop_sales",
		Table:        "store_sales",
		ConnectionID: "connection-1",
		Dimensions: []Dimension{
			{Name: "id", Type: FieldTypeNumber, SQL: "id", IsPrimaryKey: true, IsHidden: true},
			{Name: "channel", Type: FieldTypeText, SQL: "channel"},
			{Name: "amount", Type: FieldTypeNumber, SQL: "amount"},
		},
		Measures: []Measure{
			{Name: "count", Type: MeasureCount},
			{Name: "revenue", Type: MeasureSum, SQL: "amount"},
			{Name: "average_sale", Type: MeasureAverage, SQL: "amount"},
			{Name: "largest_sale", Type: MeasureMaximum, SQL: "amount"},
			{Name: "items", Type: MeasureSum, SQL: "quantity"},
			{Name: "stores_sold_in", Type: MeasureCountDistinct, SQL: "store_id"},
			{Name: "revenue_per_item", Type: MeasureFormula, SQL: "{revenue} / NULLIF({items}, 0)"},
			{Name: "revenue_per_floor_area", Type: MeasureFormula, SQL: "{revenue} / NULLIF({shops.total_floor_area}, 0)"},
			{Name: "floor_area_doubled", Type: MeasureFormula, SQL: "{shops.total_floor_area} * 2"},
		},
		Joins: []Join{
			{OtherDataset: "shops", Relationship: ManyToOne, Column: "store_id", OtherColumn: "id"},
			{OtherDataset: "sale_channels", Relationship: ManyToOne, CustomCondition: "{DATASET}.channel = {sale_channels}.code"},
		},
	}
}

func shopsDataset() Dataset {
	return Dataset{
		Name:         "shops",
		Table:        "stores",
		ConnectionID: "connection-1",
		Dimensions: []Dimension{
			{Name: "id", Type: FieldTypeNumber, SQL: "id", IsPrimaryKey: true, IsHidden: true},
			{Name: "name", Type: FieldTypeText, SQL: "name"},
			{Name: "format", Type: FieldTypeText, SQL: "format"},
		},
		Measures: []Measure{
			{Name: "count", Type: MeasureCount},
			{Name: "total_floor_area", Type: MeasureSum, SQL: "floor_area"},
			{Name: "largest_floor_area", Type: MeasureMaximum, SQL: "floor_area"},
			{Name: "average_floor_area", Type: MeasureAverage, SQL: "floor_area"},
			{Name: "area_per_store", Type: MeasureFormula, SQL: "{total_floor_area} / NULLIF({count}, 0)"},
			{Name: "area_written_out", Type: MeasureFormula, SQL: "SUM({DATASET}.floor_area)"},
		},
		Joins: []Join{
			{OtherDataset: "shop_regions", Relationship: ManyToOne, Column: "region_id", OtherColumn: "id"},
			{OtherDataset: "shop_managers", Relationship: OneToOne, Column: "id", OtherColumn: "store_id"},
			{OtherDataset: "shop_staff", Relationship: OneToMany, Column: "id", OtherColumn: "store_id"},
		},
	}
}

func shopRegionsDataset() Dataset {
	return Dataset{
		Name:         "shop_regions",
		Table:        "regions",
		ConnectionID: "connection-1",
		Dimensions: []Dimension{
			{Name: "id", Type: FieldTypeNumber, SQL: "id", IsPrimaryKey: true, IsHidden: true},
			{Name: "name", Type: FieldTypeText, SQL: "name"},
		},
		Measures: []Measure{
			{Name: "count", Type: MeasureCount},
			{Name: "names", Type: MeasureCountDistinct, SQL: "name"},
		},
		Joins: []Join{
			{OtherDataset: "shops", Relationship: OneToMany, Column: "id", OtherColumn: "region_id"},
		},
	}
}

func peopleDataset(name string) Dataset {
	return Dataset{
		Name:         name,
		Table:        "managers",
		ConnectionID: "connection-1",
		Dimensions: []Dimension{
			{Name: "id", Type: FieldTypeNumber, SQL: "id", IsPrimaryKey: true, IsHidden: true},
			{Name: "name", Type: FieldTypeText, SQL: "name"},
		},
		Measures: []Measure{
			{Name: "count", Type: MeasureCount},
		},
	}
}

func saleChannelsDataset() Dataset {
	return Dataset{
		Name:         "sale_channels",
		Table:        "channels",
		ConnectionID: "connection-1",
		Dimensions: []Dimension{
			{Name: "code", Type: FieldTypeText, SQL: "code", IsPrimaryKey: true},
			{Name: "label", Type: FieldTypeText, SQL: "label"},
		},
		Measures: []Measure{
			{Name: "count", Type: MeasureCount},
		},
	}
}

func shopSalesView() View {
	return View{
		Name: "shop_sales_view",
		Parts: []ViewPart{
			{Path: []string{"shop_sales"}, ChosenFields: []string{"revenue", "count", "channel"}},
			{Path: []string{"shop_sales", "shops"}, AddPrefix: true},
			{Path: []string{"shop_sales", "shops", "shop_regions"}, ChosenFields: []string{"name"}, AddPrefix: true},
		},
	}
}

func shopsView() View {
	return View{
		Name: "shops_view",
		Parts: []ViewPart{
			{Path: []string{"shops"}, ChosenFields: []string{"name", "format", "count", "total_floor_area"}},
			{Path: []string{"shops", "shop_managers"}, ChosenFields: []string{"name"}, AddPrefix: true},
		},
	}
}

func shopModels() Models {
	return Models{
		Datasets: []Dataset{
			shopSalesDataset(),
			shopsDataset(),
			shopRegionsDataset(),
			peopleDataset("shop_managers"),
			peopleDataset("shop_staff"),
			saleChannelsDataset(),
		},
		Views:         []View{shopSalesView(), shopsView()},
		DatabaseTypes: testDatabaseTypes(),
	}
}
