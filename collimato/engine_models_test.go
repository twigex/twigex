// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package collimato

import (
	"slices"
	"testing"

	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/queryengine"
)

const savedOrders = `{
	"name": "orders",
	"sql_table": "orders",
	"connection_id": "connection-1",
	"dimensions": [
		{"name": "id", "sql": "id", "type": "number", "primary_key": true},
		{"name": "code", "sql": "code", "type": "string", "primary_key": true, "public": true},
		{"name": "ordered_at", "sql": "ordered_at", "type": "time", "title": "Order date"},
		{"name": "is_paid", "sql": "is_paid", "type": "boolean"}
	],
	"measures": [
		{"name": "count", "type": "count", "title": "Orders"},
		{"name": "profit", "type": "sum", "sql": "amount - cost"},
		{"name": "customers", "type": "count_distinct_approx", "sql": "customer_id"},
		{"name": "margin", "type": "number", "sql": "SUM({DATASET}.amount)", "custom": true, "format": "percent"}
	],
	"joins": [
		{"name": "stores", "relationship": "many_to_one", "this_column": "store_id", "other_column": "id"},
		{"name": "regions", "relationship": "many_to_one", "sql": "{DATASET}.region = {regions}.code"}
	]
}`

const savedSalesView = `{
	"name": "sales",
	"parts": [
		{"join_path": "orders", "includes": ["count", "ordered_at"]},
		{"join_path": "orders.stores", "includes": ["*"], "prefix": true}
	]
}`

func savedFile(name string, fileType string, builderModel string) model.CollimatoWorkspaceFile {
	file := model.CollimatoWorkspaceFile{
		Name:     name,
		FileType: fileType,
	}

	if builderModel != "" {
		file.BuilderModel = &builderModel
	}

	return file
}

func TestEngineModelsConvertsDataset(t *testing.T) {
	files := []model.CollimatoWorkspaceFile{
		savedFile("orders.yml", model.CollimatoFileTypeDataset, savedOrders),
	}

	models, unsupported := EngineModels(files)

	if len(unsupported) > 0 {
		t.Fatalf("want no unsupported files, got %v", unsupported)
	}

	if len(models.Datasets) != 1 {
		t.Fatalf("want one dataset, got %d", len(models.Datasets))
	}

	orders := models.Datasets[0]

	if orders.Name != "orders" || orders.Table != "orders" || orders.ConnectionID != "connection-1" {
		t.Errorf("wrong dataset details: %+v", orders)
	}

	wantDimensions := []queryengine.Dimension{
		{Name: "id", Type: queryengine.FieldTypeNumber, SQL: "id", IsPrimaryKey: true, IsHidden: true},
		{Name: "code", Type: queryengine.FieldTypeText, SQL: "code", IsPrimaryKey: true},
		{Name: "ordered_at", Title: "Order date", Type: queryengine.FieldTypeTime, SQL: "ordered_at"},
		{Name: "is_paid", Type: queryengine.FieldTypeYesNo, SQL: "is_paid"},
	}

	if !slices.Equal(orders.Dimensions, wantDimensions) {
		t.Errorf("want dimensions\n%+v\ngot\n%+v", wantDimensions, orders.Dimensions)
	}

	wantMeasures := []queryengine.Measure{
		{Name: "count", Title: "Orders", Type: queryengine.MeasureCount},
		{Name: "profit", Type: queryengine.MeasureSum, SQL: "amount - cost"},
		{Name: "customers", Type: queryengine.MeasureCountDistinct, SQL: "customer_id"},
		{Name: "margin", Type: queryengine.MeasureFormula, SQL: "SUM({DATASET}.amount)", Format: "percent"},
	}

	if !slices.Equal(orders.Measures, wantMeasures) {
		t.Errorf("want measures\n%+v\ngot\n%+v", wantMeasures, orders.Measures)
	}

	wantJoins := []queryengine.Join{
		{OtherDataset: "stores", Relationship: queryengine.ManyToOne, Column: "store_id", OtherColumn: "id"},
		{OtherDataset: "regions", Relationship: queryengine.ManyToOne, CustomCondition: "{DATASET}.region = {regions}.code"},
	}

	if !slices.Equal(orders.Joins, wantJoins) {
		t.Errorf("want joins\n%+v\ngot\n%+v", wantJoins, orders.Joins)
	}
}

func TestEngineModelsConvertsView(t *testing.T) {
	files := []model.CollimatoWorkspaceFile{
		savedFile("sales.yml", model.CollimatoFileTypeView, savedSalesView),
	}

	models, unsupported := EngineModels(files)

	if len(unsupported) > 0 {
		t.Fatalf("want no unsupported files, got %v", unsupported)
	}

	if len(models.Views) != 1 {
		t.Fatalf("want one view, got %d", len(models.Views))
	}

	sales := models.Views[0]

	if sales.Name != "sales" || len(sales.Parts) != 2 {
		t.Fatalf("wrong view: %+v", sales)
	}

	firstPart := sales.Parts[0]
	if !slices.Equal(firstPart.Path, []string{"orders"}) {
		t.Errorf("want path [orders], got %v", firstPart.Path)
	}

	if !slices.Equal(firstPart.ChosenFields, []string{"count", "ordered_at"}) {
		t.Errorf("want chosen fields [count ordered_at], got %v", firstPart.ChosenFields)
	}

	if firstPart.AddPrefix {
		t.Error("first part adds a prefix")
	}

	secondPart := sales.Parts[1]
	if !slices.Equal(secondPart.Path, []string{"orders", "stores"}) {
		t.Errorf("want path [orders stores], got %v", secondPart.Path)
	}

	if len(secondPart.ChosenFields) != 0 {
		t.Errorf("all fields should mean no chosen fields, got %v", secondPart.ChosenFields)
	}

	if !secondPart.AddPrefix {
		t.Error("second part does not add a prefix")
	}
}

func TestEngineModelsReportsHandWrittenFile(t *testing.T) {
	files := []model.CollimatoWorkspaceFile{
		savedFile("legacy.js", model.CollimatoFileTypeDataset, ""),
		savedFile("legacy_view.yml", model.CollimatoFileTypeView, ""),
	}

	models, unsupported := EngineModels(files)

	want := []UnsupportedFile{
		{Name: "legacy.js", Reason: UnsupportedWrittenByHand},
		{Name: "legacy_view.yml", Reason: UnsupportedWrittenByHand},
	}

	if !slices.Equal(unsupported, want) {
		t.Errorf("want unsupported %v, got %v", want, unsupported)
	}

	if len(models.Datasets) > 0 || len(models.Views) > 0 {
		t.Errorf("hand-written files became models: %+v", models)
	}
}

func TestEngineModelsReportsFilesItCannotRead(t *testing.T) {
	cases := []struct {
		name         string
		fileType     string
		builderModel string
	}{
		{"broken JSON", model.CollimatoFileTypeDataset, `{"name": `},
		{"broken view JSON", model.CollimatoFileTypeView, `[`},
		{"unknown dimension type", model.CollimatoFileTypeDataset, `{"name": "a", "dimensions": [{"name": "b", "type": "geo"}]}`},
		{"unknown measure type", model.CollimatoFileTypeDataset, `{"name": "a", "measures": [{"name": "b", "type": "median"}]}`},
		{"unknown relationship", model.CollimatoFileTypeDataset, `{"name": "a", "joins": [{"name": "b", "relationship": "many_to_many"}]}`},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			files := []model.CollimatoWorkspaceFile{
				savedFile("broken.yml", testCase.fileType, testCase.builderModel),
			}

			_, unsupported := EngineModels(files)

			if len(unsupported) != 1 {
				t.Fatalf("want one unsupported file, got %v", unsupported)
			}

			if unsupported[0].Reason != UnsupportedCannotRead {
				t.Errorf("want reason %q, got %q", UnsupportedCannotRead, unsupported[0].Reason)
			}

			if unsupported[0].Detail == "" {
				t.Error("no detail says what is wrong")
			}
		})
	}
}

func TestEngineModelsSkipsFilesOfOtherTypes(t *testing.T) {
	files := []model.CollimatoWorkspaceFile{
		savedFile("settings", "settings", ""),
	}

	models, unsupported := EngineModels(files)

	if len(unsupported) > 0 {
		t.Errorf("a file of another type is reported as unsupported: %v", unsupported)
	}

	if len(models.Datasets) > 0 || len(models.Views) > 0 {
		t.Errorf("a file of another type became a model: %+v", models)
	}
}
