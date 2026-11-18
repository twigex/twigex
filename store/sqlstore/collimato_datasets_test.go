// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package sqlstore

import (
	"database/sql"
	"encoding/json"
	"reflect"
	"testing"
)

type savedFileRow struct {
	name         string
	fileType     string
	content      string
	builderModel sql.NullString
}

func seedSavedFile(t *testing.T, id, name, fileType, content string, builderModel any) {
	t.Helper()

	mustExec(t,
		`INSERT INTO collimato_workspace_files (id, workspace_id, name, file_type, content, encrypted, created_at, updated_at, builder_model)
		 VALUES (?, 'w1', ?, ?, ?, 0, 1, 1, ?)`,
		id, name, fileType, content, builderModel)
}

func savedFileRows(t *testing.T) map[string]savedFileRow {
	t.Helper()

	rows, err := testDB.Query(`SELECT id, name, file_type, content, builder_model FROM collimato_workspace_files`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	saved := map[string]savedFileRow{}

	for rows.Next() {
		var id string
		var row savedFileRow

		if err := rows.Scan(&id, &row.name, &row.fileType, &row.content, &row.builderModel); err != nil {
			t.Fatal(err)
		}

		saved[id] = row
	}

	return saved
}

func decodedModel(t *testing.T, row savedFileRow) map[string]any {
	t.Helper()

	var decoded map[string]any
	if err := json.Unmarshal([]byte(row.builderModel.String), &decoded); err != nil {
		t.Fatalf("builder model is not JSON: %v (%s)", err, row.builderModel.String)
	}

	return decoded
}

func TestRenameCubesToDatasets(t *testing.T) {
	db := requireDB(t)
	cleanTables(t, "collimato_workspace_files", "graph_datasets")
	repo := &collimatoRepository{Db: db}

	seedSavedFile(t, "orders", "orders.yml", "cube", "cubes: [...]",
		`{"name":"orders","measures":[{"name":"margin","type":"number","sql":"SUM({CUBE}.amount) / {revenue}"}],`+
			`"joins":[{"name":"stores","relationship":"many_to_one","sql":"{CUBE}.store_id = {stores}.id"}]}`)
	seedSavedFile(t, "stores-old", "Stores", "cube", "cube(`stores`, {})", nil)
	seedSavedFile(t, "stores", "stores.yml", "cube", "cubes: [...]", `{"name":"stores"}`)
	seedSavedFile(t, "sales", "sales.yml", "view", "views: [...]",
		`{"name":"sales","cubes":[{"join_path":"orders.stores","includes":["*"],"prefix":true}]}`)
	seedSavedFile(t, "legacy", "legacy.js", "cube", "cube(`legacy`, {})", nil)
	seedSavedFile(t, "env", ".env", "env", "CUBEJS_DB_TYPE=mysql", nil)
	seedSavedFile(t, "cubejs", "cube.js", "cubejs", "module.exports = {}", nil)
	seedSavedFile(t, "broken", "broken.yml", "cube", "x", `not json`)

	mustExec(t, `INSERT INTO graph_datasets (id, workspace_id, name, kind, created_at) VALUES ('g1', 'w1', 'orders', 'cube', 1), ('g2', 'w1', 'sales', 'view', 1)`)

	keptOldName, err := repo.RenameCubesToDatasets()
	if err != nil {
		t.Fatalf("RenameCubesToDatasets: %v", err)
	}

	if !reflect.DeepEqual(keptOldName, []string{"stores.yml"}) {
		t.Errorf("want stores.yml to keep its name, got %v", keptOldName)
	}

	saved := savedFileRows(t)

	wantNames := map[string][3]string{
		"orders":     {"orders", "dataset", ""},
		"stores-old": {"Stores", "dataset", "cube(`stores`, {})"},
		"stores":     {"stores.yml", "dataset", "cubes: [...]"},
		"sales":      {"sales", "view", ""},
		"legacy":     {"legacy.js", "dataset", "cube(`legacy`, {})"},
		"broken":     {"broken", "dataset", ""},
	}

	if len(saved) != len(wantNames) {
		t.Errorf("want the Cube.js settings files deleted, got %v", saved)
	}

	for id, want := range wantNames {
		row := saved[id]
		got := [3]string{row.name, row.fileType, row.content}

		if got != want {
			t.Errorf("%s: want name, type and content %q, got %q", id, want, got)
		}
	}

	orders := decodedModel(t, saved["orders"])
	measure := orders["measures"].([]any)[0].(map[string]any)
	join := orders["joins"].([]any)[0].(map[string]any)

	if measure["sql"] != "SUM({DATASET}.amount) / {revenue}" || join["sql"] != "{DATASET}.store_id = {stores}.id" {
		t.Errorf("want {DATASET} in written SQL, got %v and %v", measure["sql"], join["sql"])
	}

	sales := decodedModel(t, saved["sales"])
	if sales["cubes"] != nil || len(sales["parts"].([]any)) != 1 {
		t.Errorf("want the view's parts under parts, got %v", sales)
	}

	if saved["broken"].builderModel.String != "not json" {
		t.Errorf("want a model that is not JSON left as it was, got %q", saved["broken"].builderModel.String)
	}

	var kinds []string
	rows, err := testDB.Query(`SELECT kind FROM graph_datasets ORDER BY id`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var kind string
		if err := rows.Scan(&kind); err != nil {
			t.Fatal(err)
		}

		kinds = append(kinds, kind)
	}

	if !reflect.DeepEqual(kinds, []string{"dataset", "view"}) {
		t.Errorf("want graph kinds dataset and view, got %v", kinds)
	}

	if _, err := repo.RenameCubesToDatasets(); err != nil {
		t.Fatalf("running again: %v", err)
	}

	if again := savedFileRows(t); !reflect.DeepEqual(again, saved) {
		t.Errorf("running again changed the files:\nbefore %v\nafter  %v", saved, again)
	}
}

// A failure part way leaves every file as it was: the file type is changed
// first, and the knowledge graph table, missing here, comes after it.
func TestRenameCubesToDatasetsChangesNothingWhenItFails(t *testing.T) {
	db := requireDB(t)
	cleanTables(t, "collimato_workspace_files", "graph_datasets")
	repo := &collimatoRepository{Db: db}

	seedSavedFile(t, "orders", "orders.yml", "cube", "cubes: [...]", `{"name":"orders"}`)
	seedSavedFile(t, "env", ".env", "env", "CUBEJS_DB_TYPE=mysql", nil)

	before := savedFileRows(t)

	mustExec(t, "RENAME TABLE graph_datasets TO graph_datasets_away")
	t.Cleanup(func() { _, _ = testDB.Exec("RENAME TABLE graph_datasets_away TO graph_datasets") })

	if _, err := repo.RenameCubesToDatasets(); err == nil {
		t.Fatal("want an error without the knowledge graph table")
	}

	if after := savedFileRows(t); !reflect.DeepEqual(after, before) {
		t.Errorf("a failed run changed the files:\nbefore %v\nafter  %v", before, after)
	}
}
