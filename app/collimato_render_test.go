// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"strings"
	"testing"

	"github.com/twigex/twigex/model"
	yaml "go.yaml.in/yaml/v3"
)

func TestRenderCubeYAML(t *testing.T) {
	m := model.Dataset{
		Name:     "files",
		SqlTable: "files",
		Dimensions: []model.CubeDimension{
			{Name: "id", SQL: "id", Type: "string", PrimaryKey: true},
			{
				Name:   "file_category",
				Type:   "string",
				Title:  "File Category",
				Custom: true,
				SQL:    "CASE\n  WHEN type LIKE 'image/%' THEN 'Image'\n  ELSE 'Other'\nEND",
			},
		},
		Measures: []model.CubeMeasure{
			{Name: "count", Type: "count"},
			{
				Name:   "total_size_mb",
				Type:   "sum",
				SQL:    "ROUND(size/1048576.0, 2)",
				Title:  "Total Size (MB)",
				Format: "number",
			},
		},
	}

	got, err := renderCubeYAML(m, &model.Connection{Default: true, Database: "cloud"})
	if err != nil {
		t.Fatalf("render: %v", err)
	}

	// Output must be valid YAML and round-trip the structure.
	var parsed yamlCubeFile
	if err := yaml.Unmarshal([]byte(got), &parsed); err != nil {
		t.Fatalf("output is not valid YAML: %v\n%s", err, got)
	}
	if len(parsed.Cubes) != 1 {
		t.Fatalf("expected 1 cube, got %d", len(parsed.Cubes))
	}
	c := parsed.Cubes[0]
	if c.Name != "files" || c.SqlTable != "files" || c.DataSource != "default" {
		t.Errorf("header wrong: name=%q sql_table=%q data_source=%q", c.Name, c.SqlTable, c.DataSource)
	}

	var count, sum yamlMeasure
	for _, ms := range c.Measures {
		switch ms.Name {
		case "count":
			count = ms
		case "total_size_mb":
			sum = ms
		}
	}
	if count.Type != "count" || count.SQL != "" {
		t.Errorf("count measure wrong: %+v", count)
	}
	if sum.Type != "sum" || !strings.Contains(sum.SQL, "ROUND") {
		t.Errorf("sum measure wrong: %+v", sum)
	}

	// Multi-line CASE round-trips intact.
	var caseDim yamlDimension
	for _, d := range c.Dimensions {
		if d.Name == "file_category" {
			caseDim = d
		}
	}
	if !strings.Contains(caseDim.SQL, "CASE") || !strings.Contains(caseDim.SQL, "ELSE 'Other'") {
		t.Errorf("CASE dimension not preserved: %q", caseDim.SQL)
	}
}

func TestRenderCubeYAMLJoins(t *testing.T) {
	m := model.Dataset{
		Name:       "files",
		SqlTable:   "files",
		Dimensions: []model.CubeDimension{{Name: "id", SQL: "id", Type: "string", PrimaryKey: true}},
		Measures:   []model.CubeMeasure{{Name: "count", Type: "count"}},
		Joins: []model.CubeJoin{
			{Name: "users", Relationship: "many_to_one", ThisColumn: "owner", OtherColumn: "id"},
			{Name: "channels", Relationship: "many_to_one", SQL: "{CUBE}.channel_id = {channels}.id"},
		},
	}

	got, err := renderCubeYAML(m, &model.Connection{Default: true})
	if err != nil {
		t.Fatalf("render: %v", err)
	}

	var parsed yamlCubeFile
	if err := yaml.Unmarshal([]byte(got), &parsed); err != nil {
		t.Fatalf("output is not valid YAML: %v\n%s", err, got)
	}
	c := parsed.Cubes[0]
	if len(c.Joins) != 2 {
		t.Fatalf("expected 2 joins, got %d\n%s", len(c.Joins), got)
	}
	if c.Joins[0].SQL != "{CUBE}.owner = {users}.id" {
		t.Errorf("structured join sql wrong: %q", c.Joins[0].SQL)
	}
	if c.Joins[1].SQL != "{CUBE}.channel_id = {channels}.id" {
		t.Errorf("raw join sql override wrong: %q", c.Joins[1].SQL)
	}
}

func TestRenderViewYAML(t *testing.T) {
	m := model.View{
		Name: "files_overview",
		Cubes: []model.ViewCubeRef{
			{JoinPath: "files", Includes: []string{"file_category", "count"}},
			{JoinPath: "files.channels", Includes: []string{"name"}, Prefix: true},
			{JoinPath: "files.users", Includes: []string{"*"}},
		},
	}

	got, err := renderViewYAML(m)
	if err != nil {
		t.Fatalf("render: %v", err)
	}

	var parsed yamlViewFile
	if err := yaml.Unmarshal([]byte(got), &parsed); err != nil {
		t.Fatalf("output is not valid YAML: %v\n%s", err, got)
	}
	if len(parsed.Views) != 1 {
		t.Fatalf("expected 1 view, got %d", len(parsed.Views))
	}
	v := parsed.Views[0]
	if v.Name != "files_overview" || len(v.Cubes) != 3 {
		t.Fatalf("view header wrong: name=%q cubes=%d", v.Name, len(v.Cubes))
	}
	if v.Cubes[1].JoinPath != "files.channels" || v.Cubes[1].Prefix != true {
		t.Errorf("joined cube wrong: %+v", v.Cubes[1])
	}
	// A single "*" include emits the splat scalar, not a list.
	if s, ok := v.Cubes[2].Includes.(string); !ok || s != "*" {
		t.Errorf("splat include should be scalar \"*\", got %#v", v.Cubes[2].Includes)
	}
	// A member list stays a list.
	if _, ok := v.Cubes[0].Includes.([]interface{}); !ok {
		t.Errorf("member list should stay a list, got %#v", v.Cubes[0].Includes)
	}
}

func TestCubeDataSourceNormalization(t *testing.T) {
	if got := cubeDataSource(&model.Connection{Default: true, Database: "Anything"}); got != "default" {
		t.Errorf("default conn should be 'default', got %q", got)
	}
	if got := cubeDataSource(&model.Connection{Default: false, Database: "My DB"}); got != "mydb" {
		t.Errorf("non-default normalize failed, got %q", got)
	}
}

func TestRenderCubeYAMLMeasureTypeNormalisation(t *testing.T) {
	m := model.Dataset{
		Name:     "orders",
		SqlTable: "orders",
		Dimensions: []model.CubeDimension{
			{Name: "id", SQL: "id", Type: "string", PrimaryKey: true},
		},
		Measures: []model.CubeMeasure{
			{Name: "count", Type: "count"},
			{Name: "buyers", Type: "countDistinct", SQL: "user_id"},
			{Name: "sellers", Type: "count_distinct", SQL: "seller_id"},
		},
	}

	out, err := renderCubeYAML(m, &model.Connection{Type: "mysql", Database: "shop"})
	if err != nil {
		t.Fatalf("renderCubeYAML: %v", err)
	}

	if strings.Contains(out, "countDistinct") {
		t.Errorf("YAML kept the JavaScript spelling:\n%s", out)
	}
	if n := strings.Count(out, "type: count_distinct"); n != 2 {
		t.Errorf("want 2 count_distinct measures, got %d:\n%s", n, out)
	}
}

func TestDatasetFromTable(t *testing.T) {
	conn := &model.Connection{ID: "conn1", Database: "Analytics"}
	columns := []model.ColumnDetail{
		{Name: "id", Type: "int"},
		{Name: "title", Type: "varchar(255)"},
		{Name: "created_at", Type: "datetime"},
		{Name: "active", Type: "tinyint(1)"},
		{Name: "rank", Type: "tinyint(10)"},
		{Name: "payload", Type: "json"},
		{Name: "order-id", Type: "int"},
	}

	dataset, skipped := datasetFromTable(conn, "orders", columns)

	if dataset.Name != "orders" || dataset.SqlTable != "orders" {
		t.Errorf("name/table: got %q/%q", dataset.Name, dataset.SqlTable)
	}
	if dataset.ConnectionID != "conn1" {
		t.Errorf("connection: got %q, want conn1", dataset.ConnectionID)
	}

	want := []model.CubeDimension{
		{Name: "id", SQL: "id", Type: "number"},
		{Name: "title", SQL: "title", Type: "string"},
		{Name: "created_at", SQL: "created_at", Type: "time"},
		{Name: "active", SQL: "active", Type: "boolean"},
		{Name: "rank", SQL: "rank", Type: "number"},
	}
	if len(dataset.Dimensions) != len(want) {
		t.Fatalf("dimensions: got %d, want %d: %+v", len(dataset.Dimensions), len(want), dataset.Dimensions)
	}
	for i, d := range want {
		if dataset.Dimensions[i] != d {
			t.Errorf("dimension %d: got %+v, want %+v", i, dataset.Dimensions[i], d)
		}
	}

	if len(skipped) != 2 || skipped[0] != "payload" || skipped[1] != "order-id" {
		t.Errorf("skipped: got %v, want [payload order-id]", skipped)
	}

	if len(dataset.Measures) != 1 || dataset.Measures[0].Type != "count" {
		t.Errorf("measures: got %+v", dataset.Measures)
	}
}

func TestDatasetFromTableSkipsEveryColumn(t *testing.T) {
	conn := &model.Connection{ID: "conn1", Database: "shop"}

	dataset, skipped := datasetFromTable(conn, "blobs", []model.ColumnDetail{
		{Name: "data", Type: "longblob"},
	})

	if len(dataset.Dimensions) != 0 {
		t.Errorf("want no dimensions, got %+v", dataset.Dimensions)
	}
	if len(skipped) != 1 {
		t.Errorf("want 1 skipped, got %v", skipped)
	}
}
