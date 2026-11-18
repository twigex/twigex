// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package collimato

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/twigex/twigex/queryengine"
)

func answerModels() queryengine.Models {
	orders := queryengine.Dataset{
		Name:         "orders",
		Table:        "orders",
		ConnectionID: "connection-1",
		Dimensions: []queryengine.Dimension{
			{Name: "id", Type: queryengine.FieldTypeNumber, SQL: "id", IsPrimaryKey: true},
			{Name: "ordered_at", Title: "Order date", Type: queryengine.FieldTypeTime, SQL: "ordered_at"},
			{Name: "country", Type: queryengine.FieldTypeText, SQL: "country"},
		},
		Measures: []queryengine.Measure{
			{Name: "count", Title: "Orders", Type: queryengine.MeasureCount},
			{Name: "revenue", Type: queryengine.MeasureSum, SQL: "amount"},
		},
		Joins: []queryengine.Join{
			{OtherDataset: "stores", Relationship: queryengine.ManyToOne, Column: "store_id", OtherColumn: "id"},
		},
	}

	stores := queryengine.Dataset{
		Name:         "stores",
		Table:        "stores",
		ConnectionID: "connection-1",
		Dimensions: []queryengine.Dimension{
			{Name: "is_open", Type: queryengine.FieldTypeYesNo, SQL: "is_open"},
		},
	}

	visits := queryengine.Dataset{
		Name:         "website_visits",
		Table:        "website_visits",
		ConnectionID: "connection-1",
		Measures: []queryengine.Measure{
			{Name: "visits", Type: queryengine.MeasureAverage, SQL: "visits"},
		},
	}

	broken := queryengine.Dataset{
		Name:         "returns",
		Table:        "returns",
		ConnectionID: "connection-1",
		Joins: []queryengine.Join{
			{OtherDataset: "missing", Relationship: queryengine.ManyToOne, Column: "a", OtherColumn: "b"},
		},
	}

	sales := queryengine.View{
		Name:  "sales",
		Parts: []queryengine.ViewPart{{Path: []string{"orders"}, ChosenFields: []string{"country", "count"}}},
	}

	return queryengine.Models{
		Datasets: []queryengine.Dataset{orders, stores, visits, broken},
		Views:    []queryengine.View{sales},
	}
}

func findDatasetFields(t *testing.T, fieldList FieldList, name string) DatasetFields {
	t.Helper()

	for _, dataset := range fieldList.Datasets {
		if dataset.Name == name {
			return dataset
		}
	}

	t.Fatalf("%s is not in the field list", name)

	return DatasetFields{}
}

func TestFieldListForDataset(t *testing.T) {
	models := answerModels()
	fieldList := FieldListFor(models, queryengine.ListFields(models))

	orders := findDatasetFields(t, fieldList, "orders")

	if orders.Title != "Orders" || orders.IsView {
		t.Errorf("orders: %+v", orders)
	}

	want := []Field{
		{Key: "orders.id", Title: "ID", FullTitle: "Orders ID", Kind: "dimension", Type: "number", IsPrimaryKey: true},
		{Key: "orders.ordered_at", Title: "Order date", FullTitle: "Orders Order date", Kind: "dimension", Type: "time"},
		{Key: "orders.country", Title: "Country", FullTitle: "Orders Country", Kind: "dimension", Type: "string"},
		{Key: "orders.count", Title: "Orders", FullTitle: "Orders Orders", Kind: "measure", Type: "number", MeasureType: "count"},
		{Key: "orders.revenue", Title: "Revenue", FullTitle: "Orders Revenue", Kind: "measure", Type: "number", MeasureType: "sum"},
	}

	if !reflect.DeepEqual(orders.Fields, want) {
		t.Errorf("want fields\n%+v\ngot\n%+v", want, orders.Fields)
	}

	stores := findDatasetFields(t, fieldList, "stores")
	if stores.Fields[0].Type != "boolean" {
		t.Errorf("want a yes/no field as the builder's boolean, got %q", stores.Fields[0].Type)
	}

	sales := findDatasetFields(t, fieldList, "sales")
	if !sales.IsView || sales.Fields[1].MeasureType != "count" {
		t.Errorf("sales: %+v", sales)
	}
}

func TestFieldListGroupsJoinedDatasets(t *testing.T) {
	models := answerModels()
	fieldList := FieldListFor(models, queryengine.ListFields(models))

	orders := findDatasetFields(t, fieldList, "orders")
	stores := findDatasetFields(t, fieldList, "stores")
	visits := findDatasetFields(t, fieldList, "website_visits")
	sales := findDatasetFields(t, fieldList, "sales")

	if orders.JoinGroup == 0 || orders.JoinGroup != stores.JoinGroup {
		t.Errorf("orders and stores are joined but in groups %d and %d", orders.JoinGroup, stores.JoinGroup)
	}

	if visits.JoinGroup != 0 || sales.JoinGroup != 0 {
		t.Errorf("want no group without joins, got %d and %d", visits.JoinGroup, sales.JoinGroup)
	}
}

func TestFieldListLeavesOutModelsWithErrors(t *testing.T) {
	models := answerModels()
	engineFieldList := queryengine.ListFields(models)
	fieldList := FieldListFor(models, engineFieldList)

	for _, dataset := range fieldList.Datasets {
		if dataset.Name == "returns" {
			t.Errorf("a dataset with errors is listed: %+v", dataset)
		}
	}

	unsupported := []UnsupportedFile{
		{Name: "legacy.js", Reason: UnsupportedWrittenByHand},
		{Name: "odd", Reason: UnsupportedCannotRead, Detail: `measure "b" has unknown type "median"`},
	}

	modelErrors := ModelErrors(unsupported, engineFieldList)

	want := []ModelError{
		{Name: "legacy.js", Reason: UnsupportedWrittenByHand},
		{Name: "odd", Reason: UnsupportedCannotRead, Detail: `measure "b" has unknown type "median"`},
		{Name: "returns", Reason: UnsupportedHasErrors, Detail: "unknown_join_target: missing"},
	}

	if !reflect.DeepEqual(modelErrors, want) {
		t.Errorf("want model errors\n%+v\ngot\n%+v", want, modelErrors)
	}
}

func TestModelsThatMayComeBack(t *testing.T) {
	writtenByHand := FieldList{ModelErrors: []ModelError{{Name: "legacy.js", Reason: UnsupportedWrittenByHand}}}
	if writtenByHand.HasModelsThatMayComeBack() {
		t.Error("a file written by hand stays unsupported, so it cannot come back")
	}

	withErrors := FieldList{ModelErrors: []ModelError{{Name: "returns", Reason: UnsupportedHasErrors}}}
	if !withErrors.HasModelsThatMayComeBack() {
		t.Error("a model with errors comes back once fixed")
	}
}

func TestAnswerFor(t *testing.T) {
	result := queryengine.Result{
		Columns: []queryengine.Column{
			{Key: "orders.ordered_at", Title: "Order date", Kind: queryengine.FieldKindDimension, Type: queryengine.FieldTypeTime, DateStep: queryengine.DateStepMonth},
			{Key: "stores.is_open", Title: "Is Open", Kind: queryengine.FieldKindDimension, Type: queryengine.FieldTypeYesNo},
			{Key: "orders.revenue", Title: "Revenue", Kind: queryengine.FieldKindMeasure, Type: queryengine.FieldTypeNumber},
		},
		Rows: []map[string]any{
			{"orders.ordered_at": "2025-11-01T00:00:00Z", "stores.is_open": true, "orders.revenue": 43412.27},
		},
		DateRanges: []queryengine.ResolvedDateRange{
			{Field: "orders.ordered_at", Start: "2025-01-01T00:00:00Z", End: "2025-12-31T23:59:59.999Z"},
		},
	}

	answer := AnswerFor(result, answerModels())

	encoded, err := json.Marshal(answer)
	if err != nil {
		t.Fatal(err)
	}

	want := `{"columns":[` +
		`{"key":"orders.ordered_at","title":"Order date","fullTitle":"Orders Order date","kind":"dimension","type":"time","dateStep":"month"},` +
		`{"key":"stores.is_open","title":"Is Open","fullTitle":"Stores Is Open","kind":"dimension","type":"boolean"},` +
		`{"key":"orders.revenue","title":"Revenue","fullTitle":"Orders Revenue","kind":"measure","type":"number"}],` +
		`"rows":[{"orders.ordered_at":"2025-11-01T00:00:00Z","orders.revenue":43412.27,"stores.is_open":true}],` +
		`"dateRanges":[{"field":"orders.ordered_at","start":"2025-01-01T00:00:00Z","end":"2025-12-31T23:59:59.999Z"}]}`

	if string(encoded) != want {
		t.Errorf("want\n%s\ngot\n%s", want, encoded)
	}
}

func TestAnswerForNoRowsHasEmptyLists(t *testing.T) {
	answer := AnswerFor(queryengine.Result{Rows: []map[string]any{}}, answerModels())

	encoded, err := json.Marshal(answer)
	if err != nil {
		t.Fatal(err)
	}

	want := `{"columns":[],"rows":[],"dateRanges":[]}`
	if string(encoded) != want {
		t.Errorf("want %s, got %s", want, encoded)
	}
}
