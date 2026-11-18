// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package queryengine

import (
	"slices"
	"testing"
)

func listView(t *testing.T, view View, datasets ...Dataset) SourceFields {
	t.Helper()

	models := Models{
		Datasets: datasets,
		Views:    []View{view},
	}

	list := ListFields(models)

	return findDataset(t, list, view.Name)
}

func TestListFieldsForView(t *testing.T) {
	view := View{
		Name: "sales",
		Parts: []ViewPart{
			{Path: []string{"orders"}, ChosenFields: []string{"profit", "count", "country"}},
			{Path: []string{"orders", "stores"}, AddPrefix: true},
		},
	}

	sales := listView(t, view, ordersDataset(), storesDataset())

	want := []Field{
		{Key: "sales.country", Title: "Country", Kind: FieldKindDimension, Type: FieldTypeText},
		{Key: "sales.stores_id", Title: "Stores ID", Kind: FieldKindDimension, Type: FieldTypeNumber, IsPrimaryKey: true},
		{Key: "sales.stores_name", Title: "Store", Kind: FieldKindDimension, Type: FieldTypeText},
		{Key: "sales.stores_is_open", Title: "Stores Is Open", Kind: FieldKindDimension, Type: FieldTypeYesNo},
		{Key: "sales.profit", Title: "Profit", Kind: FieldKindMeasure, Type: FieldTypeNumber, MeasureType: MeasureSum},
		{Key: "sales.count", Title: "Orders", Kind: FieldKindMeasure, Type: FieldTypeNumber, MeasureType: MeasureCount},
		{Key: "sales.stores_count", Title: "Stores", Kind: FieldKindMeasure, Type: FieldTypeNumber, MeasureType: MeasureCount},
	}

	if !slices.Equal(sales.Fields, want) {
		t.Errorf("want fields\n%v\ngot\n%v", want, sales.Fields)
	}

	if !sales.IsView {
		t.Error("a view is not marked as a view")
	}

	if sales.Title != "Sales" {
		t.Errorf("want title %q, got %q", "Sales", sales.Title)
	}
}

func TestListFieldsForViewPathNotFollowingJoins(t *testing.T) {
	view := View{
		Name:  "sales",
		Parts: []ViewPart{{Path: []string{"stores", "orders"}}},
	}

	sales := listView(t, view, ordersDataset(), storesDataset())

	want := []ModelError{{Code: ErrorBrokenViewPath, Source: "sales", Field: "orders"}}
	if !slices.Equal(sales.Errors, want) {
		t.Errorf("want errors %v, got %v", want, sales.Errors)
	}

	if len(sales.Fields) > 0 {
		t.Errorf("a view with errors lists fields: %v", sales.Fields)
	}
}

func TestListFieldsForViewPathStartingAtUnknownDataset(t *testing.T) {
	view := View{
		Name:  "sales",
		Parts: []ViewPart{{Path: []string{"missing"}}},
	}

	sales := listView(t, view, storesDataset())

	want := []ModelError{{Code: ErrorBrokenViewPath, Source: "sales", Field: "missing"}}
	if !slices.Equal(sales.Errors, want) {
		t.Errorf("want errors %v, got %v", want, sales.Errors)
	}
}

func TestListFieldsForViewWithEmptyPath(t *testing.T) {
	view := View{
		Name:  "sales",
		Parts: []ViewPart{{}},
	}

	sales := listView(t, view, storesDataset())

	want := []ModelError{{Code: ErrorBrokenViewPath, Source: "sales"}}
	if !slices.Equal(sales.Errors, want) {
		t.Errorf("want errors %v, got %v", want, sales.Errors)
	}
}

func TestListFieldsForViewWithUnknownChosenField(t *testing.T) {
	view := View{
		Name:  "sales",
		Parts: []ViewPart{{Path: []string{"stores"}, ChosenFields: []string{"name", "missing"}}},
	}

	sales := listView(t, view, storesDataset())

	want := []ModelError{{Code: ErrorUnknownChosenField, Source: "sales", Field: "missing"}}
	if !slices.Equal(sales.Errors, want) {
		t.Errorf("want errors %v, got %v", want, sales.Errors)
	}
}

func TestListFieldsForViewUsingBrokenDataset(t *testing.T) {
	brokenStores := storesDataset()
	brokenStores.Joins = []Join{{OtherDataset: "missing", Relationship: ManyToOne, Column: "a", OtherColumn: "b"}}

	view := View{
		Name:  "sales",
		Parts: []ViewPart{{Path: []string{"orders", "stores"}}},
	}

	sales := listView(t, view, ordersDataset(), brokenStores)

	want := []ModelError{{Code: ErrorUsesBrokenDataset, Source: "sales", Field: "stores"}}
	if !slices.Equal(sales.Errors, want) {
		t.Errorf("want errors %v, got %v", want, sales.Errors)
	}
}

func TestListFieldsForViewWithSameFieldNameTwice(t *testing.T) {
	view := View{
		Name: "sales",
		Parts: []ViewPart{
			{Path: []string{"orders"}, ChosenFields: []string{"count"}},
			{Path: []string{"orders", "stores"}, ChosenFields: []string{"count"}},
		},
	}

	sales := listView(t, view, ordersDataset(), storesDataset())

	want := []ModelError{{Code: ErrorDuplicateFieldInView, Source: "sales", Field: "count"}}
	if !slices.Equal(sales.Errors, want) {
		t.Errorf("want errors %v, got %v", want, sales.Errors)
	}
}

// regionsTwiceDatasets are sales that reach regions through their store and
// through their customer.
func regionsTwiceDatasets() []Dataset {
	regions := Dataset{
		Name: "regions", Table: "regions", ConnectionID: "connection-1",
		Dimensions: []Dimension{{Name: "id", Type: FieldTypeNumber, SQL: "id", IsPrimaryKey: true}, {Name: "name", Type: FieldTypeText, SQL: "name"}},
	}

	stores := storesDataset()
	stores.Joins = []Join{{OtherDataset: "regions", Relationship: ManyToOne, Column: "region_id", OtherColumn: "id"}}

	customers := Dataset{
		Name: "customers", Table: "customers", ConnectionID: "connection-1",
		Dimensions: []Dimension{{Name: "id", Type: FieldTypeNumber, SQL: "id", IsPrimaryKey: true}},
		Joins:      []Join{{OtherDataset: "regions", Relationship: ManyToOne, Column: "region_id", OtherColumn: "id"}},
	}

	orders := ordersDataset()
	orders.Joins = append(orders.Joins, Join{OtherDataset: "customers", Relationship: ManyToOne, Column: "customer_id", OtherColumn: "id"})

	return []Dataset{orders, stores, customers, regions}
}

func TestListFieldsForViewReachingADatasetTwice(t *testing.T) {
	view := View{
		Name: "sales",
		Parts: []ViewPart{
			{Path: []string{"orders"}},
			{Path: []string{"orders", "stores", "regions"}, AddPrefix: true, ChosenFields: []string{"name"}},
			{Path: []string{"orders", "customers", "regions"}, AddPrefix: true, ChosenFields: []string{"name"}},
		},
	}

	sales := listView(t, view, regionsTwiceDatasets()...)

	want := []ModelError{{Code: ErrorViewReachesDatasetTwice, Source: "sales", Field: "regions"}}
	if !slices.Equal(sales.Errors, want) {
		t.Errorf("want errors %v, got %v", want, sales.Errors)
	}

	if len(sales.Fields) > 0 {
		t.Errorf("a view with errors lists fields: %v", sales.Fields)
	}

	models := Models{Datasets: regionsTwiceDatasets(), Views: []View{view}, DatabaseTypes: testDatabaseTypes()}
	request := Request{Dimensions: []string{"sales.regions_name"}, RowLimit: 10}

	_, err := BuildQuery(models, request)

	checkRequestError(t, err, ErrorDatasetHasErrors)
}

// Two paths through the same datasets are one way to reach them.
func TestListFieldsForViewReachingADatasetTheSameWayTwice(t *testing.T) {
	view := View{
		Name: "sales",
		Parts: []ViewPart{
			{Path: []string{"orders", "stores"}, ChosenFields: []string{"name"}},
			{Path: []string{"orders", "stores", "regions"}, AddPrefix: true, ChosenFields: []string{"name"}},
		},
	}

	sales := listView(t, view, regionsTwiceDatasets()...)

	if len(sales.Errors) > 0 {
		t.Errorf("want no errors, got %v", sales.Errors)
	}
}
