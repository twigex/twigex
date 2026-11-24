// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package queryengine

import (
	"slices"
	"strings"
	"testing"
)

func ordersDataset() Dataset {
	return Dataset{
		Name:         "orders",
		Table:        "orders",
		ConnectionID: "connection-1",
		Dimensions: []Dimension{
			{Name: "id", Type: FieldTypeNumber, SQL: "id", IsPrimaryKey: true, IsHidden: true},
			{Name: "ordered_at", Title: "Order date", Type: FieldTypeTime, SQL: "ordered_at"},
			{Name: "country", Title: "Country", Type: FieldTypeText, SQL: "country"},
		},
		Measures: []Measure{
			{Name: "count", Title: "Orders", Type: MeasureCount},
			{Name: "profit", Type: MeasureSum, SQL: "amount - cost"},
		},
		Joins: []Join{
			{OtherDataset: "stores", Relationship: ManyToOne, Column: "store_id", OtherColumn: "id"},
		},
	}
}

func storesDataset() Dataset {
	return Dataset{
		Name:         "stores",
		Table:        "stores",
		ConnectionID: "connection-1",
		Dimensions: []Dimension{
			{Name: "id", Type: FieldTypeNumber, SQL: "id", IsPrimaryKey: true, IsHidden: true},
			{Name: "name", Title: "Store", Type: FieldTypeText, SQL: "name"},
			{Name: "is_open", Type: FieldTypeYesNo, SQL: "is_open"},
		},
		Measures: []Measure{
			{Name: "count", Title: "Stores", Type: MeasureCount},
		},
	}
}

func findDataset(t *testing.T, list FieldList, name string) SourceFields {
	t.Helper()

	for _, dataset := range list.Sources {
		if dataset.Name == name {
			return dataset
		}
	}

	t.Fatalf("dataset %q is not in the list", name)

	return SourceFields{}
}

func listedFieldKeys(fields []Field) []string {
	keys := make([]string, 0, len(fields))

	for _, field := range fields {
		keys = append(keys, field.Key)
	}

	return keys
}

func errorCodes(modelErrors []ModelError) []ErrorCode {
	codes := make([]ErrorCode, 0, len(modelErrors))

	for _, modelError := range modelErrors {
		codes = append(codes, modelError.Code)
	}

	return codes
}

func TestListFieldsForDataset(t *testing.T) {
	models := Models{Datasets: []Dataset{ordersDataset(), storesDataset()}}

	list := ListFields(models)
	orders := findDataset(t, list, "orders")

	want := []Field{
		{Key: "orders.ordered_at", Title: "Order date", Kind: FieldKindDimension, Type: FieldTypeTime},
		{Key: "orders.country", Title: "Country", Kind: FieldKindDimension, Type: FieldTypeText},
		{Key: "orders.count", Title: "Orders", Kind: FieldKindMeasure, Type: FieldTypeNumber, MeasureType: MeasureCount},
		{Key: "orders.profit", Title: "Profit", Kind: FieldKindMeasure, Type: FieldTypeNumber, MeasureType: MeasureSum},
	}

	if !slices.Equal(orders.Fields, want) {
		t.Errorf("want fields %v, got %v", want, orders.Fields)
	}

	if orders.Title != "Orders" {
		t.Errorf("want title %q, got %q", "Orders", orders.Title)
	}

	if orders.IsView {
		t.Error("a dataset is listed as a view")
	}

	if len(orders.Errors) > 0 {
		t.Errorf("want no errors, got %v", orders.Errors)
	}
}

func TestListFieldsLeavesOutHiddenFields(t *testing.T) {
	models := Models{Datasets: []Dataset{storesDataset()}}

	list := ListFields(models)
	stores := findDataset(t, list, "stores")
	keys := listedFieldKeys(stores.Fields)

	if slices.Contains(keys, "stores.id") {
		t.Errorf("hidden primary key is listed: %v", keys)
	}
}

func TestListFieldsShowsPrimaryKeyThatIsNotHidden(t *testing.T) {
	stores := storesDataset()
	stores.Dimensions[0].IsHidden = false

	models := Models{Datasets: []Dataset{stores}}

	list := ListFields(models)
	listed := findDataset(t, list, "stores")
	keys := listedFieldKeys(listed.Fields)

	if !slices.Contains(keys, "stores.id") {
		t.Errorf("shown primary key is missing: %v", keys)
	}

	if !listed.Fields[0].IsPrimaryKey {
		t.Errorf("the primary key is not marked as one: %+v", listed.Fields[0])
	}
}

func TestListFieldsMakesMissingTitlesFromNames(t *testing.T) {
	models := Models{Datasets: []Dataset{storesDataset()}}

	list := ListFields(models)
	stores := findDataset(t, list, "stores")

	isOpen := stores.Fields[1]
	if isOpen.Title != "Is Open" {
		t.Errorf("want title %q, got %q", "Is Open", isOpen.Title)
	}
}

func TestListFieldsSortsDatasetsByNameThenViews(t *testing.T) {
	view := View{
		Name:  "a_view",
		Parts: []ViewPart{{Path: []string{"stores"}}},
	}

	models := Models{
		Datasets: []Dataset{storesDataset(), ordersDataset()},
		Views:    []View{view},
	}

	list := ListFields(models)

	var names []string
	for _, dataset := range list.Sources {
		names = append(names, dataset.Name)
	}

	want := []string{"orders", "stores", "a_view"}
	if !slices.Equal(names, want) {
		t.Errorf("want order %v, got %v", want, names)
	}
}

func TestListFieldsUnknownJoinTargetMarksOnlyThatDataset(t *testing.T) {
	models := Models{Datasets: []Dataset{ordersDataset()}}

	list := ListFields(models)
	orders := findDataset(t, list, "orders")

	want := []ModelError{{Code: ErrorUnknownJoinTarget, Source: "orders", Field: "stores"}}
	if !slices.Equal(orders.Errors, want) {
		t.Errorf("want errors %v, got %v", want, orders.Errors)
	}

	if len(orders.Fields) > 0 {
		t.Errorf("a dataset with errors lists fields: %v", orders.Fields)
	}
}

func TestListFieldsOtherDatasetsWorkWhenOneIsBroken(t *testing.T) {
	brokenOrders := ordersDataset()
	brokenOrders.Joins[0].OtherDataset = "missing"

	models := Models{Datasets: []Dataset{brokenOrders, storesDataset()}}

	list := ListFields(models)
	stores := findDataset(t, list, "stores")

	if len(stores.Errors) > 0 {
		t.Errorf("want no errors, got %v", stores.Errors)
	}

	if len(stores.Fields) == 0 {
		t.Error("stores lists no fields")
	}
}

func TestListFieldsDuplicateDatasetName(t *testing.T) {
	models := Models{Datasets: []Dataset{storesDataset(), storesDataset()}}

	list := ListFields(models)

	for _, dataset := range list.Sources {
		codes := errorCodes(dataset.Errors)

		if !slices.Contains(codes, ErrorDuplicateName) {
			t.Errorf("want a duplicate name error, got %v", dataset.Errors)
		}
	}
}

func TestListFieldsViewWithSameNameAsDataset(t *testing.T) {
	view := View{
		Name:  "stores",
		Parts: []ViewPart{{Path: []string{"stores"}}},
	}

	models := Models{
		Datasets: []Dataset{storesDataset()},
		Views:    []View{view},
	}

	list := ListFields(models)

	for _, dataset := range list.Sources {
		codes := errorCodes(dataset.Errors)

		if !slices.Contains(codes, ErrorDuplicateName) {
			t.Errorf("want a duplicate name error on %q, got %v", dataset.Name, dataset.Errors)
		}
	}
}

func TestListFieldsDuplicateFieldName(t *testing.T) {
	stores := storesDataset()
	duplicate := Measure{Name: "name", Type: MeasureCount}
	stores.Measures = append(stores.Measures, duplicate)

	models := Models{Datasets: []Dataset{stores}}

	list := ListFields(models)
	listed := findDataset(t, list, "stores")

	want := []ModelError{{Code: ErrorDuplicateName, Source: "stores", Field: "name"}}
	if !slices.Equal(listed.Errors, want) {
		t.Errorf("want errors %v, got %v", want, listed.Errors)
	}
}

func TestListFieldsKeepsEachSameNamedDatasetsOwnProblems(t *testing.T) {
	plain := Dataset{Name: "orders", Table: "orders", Dimensions: []Dimension{{Name: "x", Type: FieldTypeText, SQL: "x"}}}

	withDuplicateField := plain
	withDuplicateField.Dimensions = []Dimension{
		{Name: "x", Type: FieldTypeText, SQL: "x"},
		{Name: "x", Type: FieldTypeText, SQL: "y"},
	}

	list := ListFields(Models{Datasets: []Dataset{withDuplicateField, plain}})

	if len(list.Sources) != 2 {
		t.Fatalf("want 2 sources, got %d", len(list.Sources))
	}

	problemCounts := []int{len(list.Sources[0].Errors), len(list.Sources[1].Errors)}
	if problemCounts[0]+problemCounts[1] != 3 {
		t.Errorf("want the duplicate name on both and the duplicate field on one, got %v and %v",
			list.Sources[0].Errors, list.Sources[1].Errors)
	}
}

// Datasets named like the tables the engine makes inside its own queries, or
// too long for Postgres to keep whole, cannot be used.
func TestListFieldsForDatasetsWithReservedOrTooLongNames(t *testing.T) {
	cases := []struct {
		name     string
		wantCode ErrorCode
	}{
		{"all_groups", ErrorReservedName},
		{"distinct_rows", ErrorReservedName},
		{"part_1", ErrorReservedName},
		{"part_12", ErrorReservedName},
		{strings.Repeat("x", 64), ErrorNameTooLong},
	}

	for _, testCase := range cases {
		dataset := storesDataset()
		dataset.Name = testCase.name

		list := ListFields(Models{Datasets: []Dataset{dataset}})
		got := findDataset(t, list, testCase.name)

		want := []ModelError{{Code: testCase.wantCode, Source: testCase.name}}
		if !slices.Equal(got.Errors, want) {
			t.Errorf("%s: want errors %v, got %v", testCase.name, want, got.Errors)
		}
	}

	for _, allowedName := range []string{"part", "part_one", "parts_1", strings.Repeat("x", 63)} {
		dataset := storesDataset()
		dataset.Name = allowedName

		list := ListFields(Models{Datasets: []Dataset{dataset}})
		got := findDataset(t, list, allowedName)

		if len(got.Errors) > 0 {
			t.Errorf("%s: want no errors, got %v", allowedName, got.Errors)
		}
	}
}
