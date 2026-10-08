// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package sqlstore

import (
	"context"
	"testing"

	"github.com/twigex/twigex/model"
)

func setupHeaderMeta(t *testing.T) {
	t.Helper()
	requireDB(t)
	cleanTables(t, "workspace_fields", "workspace_relationships", "workspace_tables")

	for _, tb := range []struct {
		id, name, parent   string
		singleSelect, both int
	}{
		{"t-tasks", "zz_tasks", "", 0, 0},
		{"t-link", "zz_link", "t-projects", 0, 1},
		{"t-status", "zz_test_meta_status", "t-tasks", 1, 0},
		{"t-projects", "zz_projects", "", 0, 0},
	} {
		mustExec(t, `INSERT INTO workspace_tables (id, workspace_id, name, display_name, linked, single_select, parent_table_id, both_direction_link, second_table_id, folder_id, deleted_at)
			VALUES (?, 'w1', ?, '', 0, ?, ?, ?, '', '', 0)`, tb.id, tb.name, tb.singleSelect, tb.parent, tb.both)
	}
	for _, f := range []struct {
		id, table, name, display, fieldType, parent string
		deleted                                     int
	}{
		{"f-project", "t-tasks", "project", "Project", "link", "", 0},
		{"f-notes", "t-tasks", "notes", "Notes", "text", "", 0},
		{"f-old", "t-tasks", "gone", "Gone", "text", "", 1},
		{"f-back", "t-projects", "tasks_back", "Tasks", "link", "f-project", 0},
	} {
		mustExec(t, `INSERT INTO workspace_fields (id, workspace_id, table_id, field_name, field_display_name, field_type, parent_field_id, created_at, updated_at, deleted_at)
			VALUES (?, 'w1', ?, ?, ?, ?, NULLIF(?, ''), 1, 0, ?)`, f.id, f.table, f.name, f.display, f.fieldType, f.parent, f.deleted)
	}
	for _, r := range []struct{ id, name, linked string }{
		{"r1", "project", "t-link"},
		{"r2", "status", "t-status"},
	} {
		mustExec(t, `INSERT INTO workspace_relationships (id, workspace_id, table_id, linked_table_id, table_name, created_at, updated_at, deleted_at)
			VALUES (?, 'w1', 't-tasks', ?, ?, 1, 0, 0)`, r.id, r.linked, r.name)
	}
}

func TestGetTableHeaderMetaReadsEveryColumnsSettings(t *testing.T) {
	setupHeaderMeta(t)
	w := &workspaceRepository{Db: testDB}

	meta, err := w.GetTableHeaderMeta("t-tasks")
	if err != nil {
		t.Fatal(err)
	}

	for name, want := range map[string]model.WorkspaceFieldData{
		"project": {ID: "f-project", DisplayName: "Project", FieldType: "link"},
		"notes":   {ID: "f-notes", DisplayName: "Notes", FieldType: "text"},
	} {
		f := meta.Fields[name]
		if f.ID != want.ID || f.DisplayName != want.DisplayName || f.FieldType != want.FieldType {
			t.Errorf("field %s = %+v, want %+v", name, f, want)
		}
	}
	if f, ok := meta.Fields["gone"]; ok {
		t.Errorf("a deleted field was read: %+v", f)
	}

	for name, want := range map[string]string{"project": "t-link", "status": "t-status", "notes": ""} {
		if got := meta.Links[name]; got != want {
			t.Errorf("link %s = %q, want %q", name, got, want)
		}
	}
	for id, want := range map[string]model.LinkedTableMeta{
		"t-link":   {ParentTableID: "t-projects", BothDirections: true},
		"t-status": {ParentTableID: "t-tasks", SingleSelect: true},
	} {
		lt := meta.LinkedTables[id]
		if lt.ParentTableID != want.ParentTableID || lt.SingleSelect != want.SingleSelect || lt.BothDirections != want.BothDirections {
			t.Errorf("linked table %s = %+v, want %+v", id, lt, want)
		}
	}
	if got := meta.LinkedFieldNames["t-projects/f-project"]; got != "tasks_back" {
		t.Errorf("linked field name = %q, want tasks_back", got)
	}
}

func TestGetSingleSelectOptionsReturnsOnlyTheAskedOptions(t *testing.T) {
	setupHeaderMeta(t)
	createDynamicTable(t, "zz_test_meta_status", func(w *workspaceRepository, name string) error {
		_, err := w.Db.Exec("CREATE TABLE `" + name + "` (id VARCHAR(36) PRIMARY KEY, name VARCHAR(255), color VARCHAR(255))")
		return err
	})
	mustExec(t, "INSERT INTO `zz_test_meta_status` (id, name, color) VALUES ('s1', 'Open', '#fff'), ('s2', 'Done', NULL), ('s3', 'Other', '#000')")
	w := &workspaceRepository{Db: testDB}

	got, err := w.GetSingleSelectOptions(context.Background(), "t-status", []string{"s1", "s2", "missing"})
	if err != nil {
		t.Fatal(err)
	}

	if s1 := got["s1"]; len(got) != 2 || s1 == nil || s1.ID != "s1" || s1.Name != "Open" || s1.Color != "#fff" || got["s2"].Name != "Done" {
		t.Errorf("options = %v", got)
	}
	if none, err := w.GetSingleSelectOptions(context.Background(), "no-such-table", []string{"s1"}); err != nil || len(none) != 0 {
		t.Errorf("unknown table = %v, %v; want none", none, err)
	}
}
