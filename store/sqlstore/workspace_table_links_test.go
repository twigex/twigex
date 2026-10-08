// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package sqlstore

import (
	"context"
	"strings"
	"testing"

	"github.com/twigex/twigex/model"
)

func seedLinkWorkspace(t *testing.T) *workspaceRepository {
	t.Helper()
	requireDB(t)
	cleanTables(t, "workspaces", "workspace_folders", "workspace_tables", "workspace_fields", "workspace_relationships")

	mustExec(t, `INSERT INTO workspaces (id, title, description, start_date, end_date, pre_fix, created_at, updated_at, deleted_at, created_by)
		VALUES ('w1', 'w1', '', 0, 0, ?, 0, 0, 0, 'u1')`, "lk"+strings.ToLower(model.NewID()[:6])+"_")
	mustExec(t, `INSERT INTO workspace_folders (id, workspace_id, parent_folder_id, name, description, created_at, updated_at, deleted_at)
		VALUES ('folder', 'w1', '', 'folder', '', 0, 0, 0)`)
	return &workspaceRepository{Db: testDB}
}

func physicalName() string {
	return "t" + strings.ReplaceAll(model.NewID(), "-", "")
}

func createLinkTestTable(t *testing.T, w *workspaceRepository, name, folderID string) string {
	t.Helper()
	id := model.NewID()
	statuses := []model.KanbanStatusOption{{ID: model.NewID(), Name: "Completed", Color: "#16a34a", StatusType: "Done"}}
	if _, err := w.CreateTable(id, "w1", name, name, false, "", "u1", false, false, "", folderID, "[]", statuses, model.NewID(), physicalName(), model.NewID()); err != nil {
		t.Fatalf("creating table %s: %v", name, err)
	}

	return id
}

func createLinkTestField(t *testing.T, w *workspaceRepository, tableID, name, fieldType, linkedTableID string, bothWays bool) (string, string) {
	t.Helper()
	sqlType := "TEXT"
	if fieldType == "number" {
		sqlType = "NUMERIC"
	}
	p := model.CreateFieldParams{
		WorkspaceID:                 "w1",
		TableID:                     tableID,
		FieldName:                   name,
		FieldType:                   fieldType,
		SQLFieldType:                sqlType,
		LinkedTableID:               linkedTableID,
		UserID:                      "u1",
		SelectedType:                fieldType,
		LinkBothDirections:          bothWays,
		FieldNameInSecondTable:      name + "_back",
		FieldNameDisplay:            name,
		MainFieldID:                 model.NewID(),
		SecondFieldID:               model.NewID(),
		LinkTableID:                 model.NewID(),
		SecondLinkID:                model.NewID(),
		LinkTablePhysicalName:       physicalName(),
		SecondLinkTablePhysicalName: physicalName(),
		SingleSelectTableID:         model.NewID(),
		SingleSelectPhysicalName:    physicalName(),
		LinkViewID:                  model.NewID(),
		SecondLinkViewID:            model.NewID(),
	}
	if _, err := w.CreateTableField(p); err != nil {
		t.Fatalf("creating field %s: %v", name, err)
	}

	if !bothWays {
		return p.MainFieldID, ""
	}
	return p.MainFieldID, p.SecondFieldID
}

func fieldDeleted(t *testing.T, id string) bool {
	t.Helper()
	var at int64
	if err := testDB.QueryRow("SELECT deleted_at FROM workspace_fields WHERE id = ?", id).Scan(&at); err != nil {
		t.Fatalf("reading field %s: %v", id, err)
	}

	return at != 0
}

func linkedFieldsOf(result []map[string]interface{}) map[string]string {
	tables := map[string]string{}
	for _, group := range result {
		for _, id := range group["field_ids"].([]string) {
			tables[id] = group["table_id"].(string)
		}
	}
	return tables
}

func TestGeneratedNamesTakeAnyNameAsText(t *testing.T) {
	w := seedLinkWorkspace(t)
	label := "Client `Projects`'; DROP TABLE workspace_tables; -- Pārdošana €"

	tableID := model.NewID()
	statuses := []model.KanbanStatusOption{{ID: model.NewID(), Name: "Completed", Color: "#16a34a", StatusType: "Done"}}
	if _, err := w.CreateTable(tableID, "w1", label, "t"+strings.ReplaceAll(tableID, "-", ""), false, "", "u1", false, false, "", "",
		"[]", statuses, model.NewID(), physicalName(), model.NewID()); err != nil {
		t.Fatal(err)
	}
	other := createLinkTestTable(t, w, "other", "")

	fieldID, secondID := model.NewID(), model.NewID()
	p := model.CreateFieldParams{
		WorkspaceID:                   "w1",
		TableID:                       tableID,
		FieldName:                     "c" + strings.ReplaceAll(fieldID, "-", ""),
		FieldType:                     "link",
		SQLFieldType:                  "TEXT",
		LinkedTableID:                 other,
		UserID:                        "u1",
		SelectedType:                  "link",
		LinkBothDirections:            true,
		FieldNameInSecondTable:        "c" + strings.ReplaceAll(secondID, "-", ""),
		FieldNameDisplay:              label,
		FieldNameInSecondTableDisplay: label,
		MainFieldID:                   fieldID,
		SecondFieldID:                 secondID,
		LinkTableID:                   model.NewID(),
		SecondLinkID:                  model.NewID(),
		LinkTablePhysicalName:         physicalName(),
		SecondLinkTablePhysicalName:   physicalName(),
		SingleSelectTableID:           model.NewID(),
		SingleSelectPhysicalName:      physicalName(),
		LinkViewID:                    model.NewID(),
		SecondLinkViewID:              model.NewID(),
	}
	if _, err := w.CreateTableField(p); err != nil {
		t.Fatal(err)
	}

	var tableLabel string
	if err := testDB.QueryRow("SELECT display_name FROM workspace_tables WHERE id = ?", tableID).Scan(&tableLabel); err != nil {
		t.Fatal(err)
	}

	if tableLabel != label {
		t.Errorf("table is called %q, want %q", tableLabel, label)
	}
	for _, id := range []string{fieldID, secondID} {
		var fieldLabel string
		if err := testDB.QueryRow("SELECT field_display_name FROM workspace_fields WHERE id = ?", id).Scan(&fieldLabel); err != nil {
			t.Fatal(err)
		}

		if fieldLabel != label {
			t.Errorf("field %s is called %q, want %q", id, fieldLabel, label)
		}
	}

	var tables int
	if err := testDB.QueryRow("SELECT COUNT(*) FROM workspace_tables").Scan(&tables); err != nil || tables == 0 {
		t.Fatalf("workspace_tables is gone or empty: %v", err)
	}

	if _, err := w.DeleteTable("w1", tableID, 100); err != nil {
		t.Fatal(err)
	}
}

func TestDeleteTableTakesTheLinkColumnsOfOtherTablesWithIt(t *testing.T) {
	w := seedLinkWorkspace(t)

	target := createLinkTestTable(t, w, "target", "")
	oneWay := createLinkTestTable(t, w, "one_way", "")
	twoWay := createLinkTestTable(t, w, "two_way", "")
	unrelated := createLinkTestTable(t, w, "unrelated", "")

	notes, _ := createLinkTestField(t, w, target, "notes", "text", "", false)
	outgoing, _ := createLinkTestField(t, w, target, "to_unrelated", "link", unrelated, false)
	incoming, _ := createLinkTestField(t, w, oneWay, "to_target", "link", target, false)
	pair, mirror := createLinkTestField(t, w, twoWay, "with_target", "link", target, true)
	kept, _ := createLinkTestField(t, w, unrelated, "amount", "number", "", false)

	result, err := w.DeleteTable("w1", target, 100)
	if err != nil {
		t.Fatal(err)
	}

	for id, want := range map[string]bool{
		notes:    true,
		outgoing: true,
		mirror:   true,
		incoming: true,
		pair:     true,
		kept:     false,
	} {
		if got := fieldDeleted(t, id); got != want {
			t.Errorf("field %s deleted = %t, want %t", id, got, want)
		}
	}

	gone := linkedFieldsOf(result)
	want := map[string]string{incoming: oneWay, pair: twoWay}
	if len(gone) != len(want) {
		t.Errorf("reported fields = %v, want %v", gone, want)
	}
	for id, table := range want {
		if gone[id] != table {
			t.Errorf("field %s reported under %q, want %q", id, gone[id], table)
		}
	}
}

func TestDeleteTableLeavesALinkFieldThatReusedTheNameOfAnOldOne(t *testing.T) {
	w := seedLinkWorkspace(t)

	target := createLinkTestTable(t, w, "target", "")
	other := createLinkTestTable(t, w, "other", "")
	source := createLinkTestTable(t, w, "source", "")

	old, _ := createLinkTestField(t, w, source, "link_field", "link", target, false)
	mustExec(t, "UPDATE workspace_fields SET deleted_at = 1 WHERE id = ?", old)
	mustExec(t, "UPDATE workspace_relationships SET created_at = created_at - 60 WHERE table_id = ? AND table_name = 'link_field'", source)
	mustExec(t, "ALTER TABLE "+tableNameOf(t, source)+" DROP COLUMN link_field")
	reused, _ := createLinkTestField(t, w, source, "link_field", "link", other, false)

	if _, err := w.DeleteTable("w1", target, 100); err != nil {
		t.Fatal(err)
	}

	if fieldDeleted(t, reused) {
		t.Error("a link to another table went with the table an old field of the same name pointed at")
	}

	if _, err := w.DeleteTable("w1", other, 100); err != nil {
		t.Fatal(err)
	}

	if !fieldDeleted(t, reused) {
		t.Error("the link to the deleted table stayed")
	}
}

func tableNameOf(t *testing.T, tableID string) string {
	t.Helper()
	var name string
	if err := testDB.QueryRow("SELECT name FROM workspace_tables WHERE id = ?", tableID).Scan(&name); err != nil {
		t.Fatal(err)
	}

	return name
}

func TestGetLinkingTablesNamesTheTablesOutsideThatLinkIn(t *testing.T) {
	w := seedLinkWorkspace(t)
	ctx := context.Background()

	inside := createLinkTestTable(t, w, "inside", "folder")
	alsoInside := createLinkTestTable(t, w, "also_inside", "folder")
	oneWay := createLinkTestTable(t, w, "clients", "")
	twoWay := createLinkTestTable(t, w, "invoices", "")
	pointedAt := createLinkTestTable(t, w, "pointed_at", "")

	createLinkTestField(t, w, alsoInside, "to_inside", "link", inside, false)
	createLinkTestField(t, w, oneWay, "to_inside", "link", inside, false)
	createLinkTestField(t, w, twoWay, "with_inside", "link", alsoInside, true)
	createLinkTestField(t, w, inside, "to_pointed_at", "link", pointedAt, false)

	folderTables, err := w.GetFolderTableIDs(ctx, "w1", "folder")
	if err != nil {
		t.Fatal(err)
	}

	if len(folderTables) != 2 {
		t.Fatalf("folder tables = %q, want the two inside", folderTables)
	}

	for _, tc := range []struct {
		name    string
		targets []string
		want    []string
	}{
		{"folder", folderTables, []string{"clients", "invoices"}},
		{"one table", []string{inside}, []string{"also_inside", "clients"}},
		{"linked to nothing", []string{oneWay}, nil},
	} {
		tables, err := w.GetLinkingTables(ctx, "w1", tc.targets)
		if err != nil {
			t.Fatal(err)
		}

		var names []string
		for _, table := range tables {
			names = append(names, table.Name)
		}
		if strings.Join(names, ",") != strings.Join(tc.want, ",") {
			t.Errorf("%s: linking tables = %q, want %q", tc.name, names, tc.want)
		}
	}
}

func TestDeleteFolderReportsOnlyTheLinksOfTablesLeftStanding(t *testing.T) {
	w := seedLinkWorkspace(t)
	ctx := context.Background()

	inside := createLinkTestTable(t, w, "inside", "folder")
	alsoInside := createLinkTestTable(t, w, "also_inside", "folder")
	outside := createLinkTestTable(t, w, "outside", "")

	within, _ := createLinkTestField(t, w, alsoInside, "to_inside", "link", inside, false)
	fromOutside, _ := createLinkTestField(t, w, outside, "to_inside", "link", inside, false)

	tables, linked, err := w.DeleteFolder(ctx, "w1", "folder", true)
	if err != nil {
		t.Fatal(err)
	}

	if len(tables) != 2 {
		t.Errorf("deleted tables = %q, want the two inside", tables)
	}
	if !fieldDeleted(t, within) || !fieldDeleted(t, fromOutside) {
		t.Error("a link field into the folder stayed")
	}

	gone := linkedFieldsOf(linked)
	if len(gone) != 1 || gone[fromOutside] != outside {
		t.Errorf("reported fields = %v, want only %s of the table outside", gone, fromOutside)
	}
}

func linkTableOf(t *testing.T, tableID, fieldName string) string {
	t.Helper()
	var id string
	if err := testDB.QueryRow(`SELECT linked_table_id FROM workspace_relationships WHERE table_id = ? AND table_name = ?`,
		tableID, fieldName).Scan(&id); err != nil {
		t.Fatal(err)
	}
	return id
}

func TestOnlyATablesOwnOptionTableCountsAsItsOptions(t *testing.T) {
	w := seedLinkWorkspace(t)
	ctx := context.Background()

	own := createLinkTestTable(t, w, "own", "")
	other := createLinkTestTable(t, w, "other", "")
	createLinkTestField(t, w, own, "to_other", "link", other, false)

	for _, tc := range []struct {
		name, table, candidate string
		want                   bool
	}{
		{"its own status options", own, linkTableOf(t, own, "status"), true},
		{"another table's status options", own, linkTableOf(t, other, "status"), false},
		{"its link table, which holds no options", own, linkTableOf(t, own, "to_other"), false},
		{"a table that does not exist", own, "nothing", false},
	} {
		got, err := w.IsOptionTableOf(ctx, tc.table, tc.candidate)
		if err != nil {
			t.Fatal(err)
		}

		if got != tc.want {
			t.Errorf("%s: IsOptionTableOf = %t, want %t", tc.name, got, tc.want)
		}
	}
}
