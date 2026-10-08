// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package sqlstore

import (
	"context"
	"reflect"
	"slices"
	"testing"

	"github.com/twigex/twigex/model"
)

func setupTaskReportTables(t *testing.T) []model.TaskSource {
	t.Helper()
	requireDB(t)
	createDynamicTable(t, "zz_test_report_a", createTaskTable)
	createDynamicTable(t, "zz_test_report_b", createTaskTable)
	mustExec(t, "ALTER TABLE `zz_test_report_a` ADD COLUMN priority INT")

	for _, r := range []struct {
		table, id, name string
		due, created    int
		priority        any
	}{
		{"zz_test_report_a", "a1", "apple", 300, 1, 2},
		{"zz_test_report_a", "a2", "cherry", 0, 150, 1},
		{"zz_test_report_b", "b1", "banana", 100, 2, nil},
		{"zz_test_report_b", "b2", "date", 200, 3, nil},
	} {
		if r.table == "zz_test_report_a" {
			mustExec(t, "INSERT INTO `zz_test_report_a` (id, name, due_date, created_at, priority, deleted_at) VALUES (?, ?, ?, ?, ?, 0)", r.id, r.name, r.due, r.created, r.priority)
		} else {
			mustExec(t, "INSERT INTO `zz_test_report_b` (id, name, due_date, created_at, deleted_at) VALUES (?, ?, ?, ?, 0)", r.id, r.name, r.due, r.created)
		}
	}
	base := map[string]bool{"id": true, "name": true, "due_date": true, "created_at": true}
	withPriority := map[string]bool{"id": true, "name": true, "due_date": true, "created_at": true, "priority": true}
	return []model.TaskSource{
		{TableID: "table-a", TableName: "zz_test_report_a", Columns: withPriority},
		{TableID: "table-b", TableName: "zz_test_report_b", Columns: base},
	}
}

func keyIDs(keys []model.TaskKey) []string {
	ids := make([]string, 0, len(keys))
	for _, k := range keys {
		ids = append(ids, k.ID)
	}
	return ids
}

func TestGetTaskPageKeysSortsAcrossTables(t *testing.T) {
	sources := setupTaskReportTables(t)
	w := &workspaceRepository{Db: testDB}

	for _, tc := range []struct {
		name          string
		sort          []model.SortParam
		limit, offset int
		want          []string
	}{
		{"dated by due date, then the rest by creation", nil, 10, 0, []string{"b1", "b2", "a1", "a2"}},
		{"second page", nil, 2, 2, []string{"a1", "a2"}},
		{"past the end", nil, 2, 4, []string{}},
		{"by name descending", []model.SortParam{{Field: "name", Direction: "desc"}}, 10, 0, []string{"b2", "a2", "b1", "a1"}},
		{"by a column only one table has", []model.SortParam{{Field: "priority", Direction: "asc"}}, 10, 0, []string{"b1", "b2", "a2", "a1"}},
		{"an unsafe field is ignored", []model.SortParam{{Field: "name; DROP TABLE x", Direction: "asc"}}, 10, 0, []string{"b1", "b2", "a1", "a2"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			keys, err := w.GetTaskPageKeys(context.Background(), sources, tc.sort, tc.limit, tc.offset)
			if err != nil {
				t.Fatal(err)
			}

			if got := keyIDs(keys); !slices.Equal(got, tc.want) {
				t.Errorf("keys = %v, want %v", got, tc.want)
			}
		})
	}

	keys, err := w.GetTaskPageKeys(context.Background(), sources, nil, 1, 0)
	if err != nil || len(keys) != 1 || keys[0] != (model.TaskKey{TableID: "table-b", ID: "b1"}) {
		t.Errorf("first key = %v, %v; want b1 of table-b", keys, err)
	}
}

func TestGetAttachmentNamesListsEachTasksFilesNewestFirst(t *testing.T) {
	requireDB(t)
	cleanTables(t, "workspace_attachments")
	for _, f := range []struct {
		id, task, field, name string
		created, deleted      int
	}{
		{"f1", "t1", "docs", "old.pdf", 1, 0},
		{"f2", "t1", "docs", "new.pdf", 2, 0},
		{"f3", "t1", "docs", "gone.pdf", 3, 1},
		{"f4", "t2", "docs", "other.png", 1, 0},
		{"f5", "t1", "photos", "elsewhere.jpg", 1, 0},
	} {
		mustExec(t, `INSERT INTO workspace_attachments (id, workspace_id, table_id, field_id, task_id, user_id, name, size, mime_type, width, height, created_at, updated_at, deleted_at)
			VALUES (?, 'w1', 'tbl', ?, ?, 'u1', ?, 1, 'x', 0, 0, ?, 0, ?)`, f.id, f.field, f.task, f.name, f.created, f.deleted)
	}
	w := &workspaceRepository{Db: testDB}

	names, err := w.GetAttachmentNames(context.Background(), "tbl", "docs", []string{"t1", "t2", "t3"})
	if err != nil {
		t.Fatal(err)
	}

	if !slices.Equal(names["t1"], []string{"new.pdf", "old.pdf"}) || !slices.Equal(names["t2"], []string{"other.png"}) || names["t3"] != nil {
		t.Errorf("names = %v", names)
	}
}

func TestGetAttachmentsForTasksListsEachTasksFilesNewestFirst(t *testing.T) {
	requireDB(t)
	cleanTables(t, "workspace_attachments")
	for _, f := range []struct {
		id, task, field, name string
		created, deleted      int
	}{
		{"f1", "t1", "docs", "old.pdf", 1, 0},
		{"f2", "t1", "docs", "new.pdf", 2, 0},
		{"f3", "t1", "docs", "gone.pdf", 3, 1},
		{"f4", "t2", "docs", "other.png", 1, 0},
		{"f5", "t1", "photos", "elsewhere.jpg", 1, 0},
	} {
		mustExec(t, `INSERT INTO workspace_attachments (id, workspace_id, table_id, field_id, task_id, user_id, name, size, mime_type, width, height, created_at, updated_at, deleted_at)
			VALUES (?, 'w1', 'tbl', ?, ?, 'u1', ?, 7, 'application/pdf', 3, 4, ?, 0, ?)`, f.id, f.field, f.task, f.name, f.created, f.deleted)
	}
	w := &workspaceRepository{Db: testDB}

	tasks := []string{"t1", "t2", "t3"}
	files, err := w.GetAttachmentsForTasks(context.Background(), "tbl", "docs", tasks)
	if err != nil {
		t.Fatal(err)
	}

	file := func(id, name string, created int64) map[string]interface{} {
		return map[string]interface{}{
			"id":         id,
			"name":       name,
			"size":       int64(7),
			"type":       "application/pdf",
			"width":      3,
			"height":     4,
			"created_at": created,
		}
	}
	want := map[string][]map[string]interface{}{
		"t1": {file("f2", "new.pdf", 2), file("f1", "old.pdf", 1)},
		"t2": {file("f4", "other.png", 1)},
	}
	if !reflect.DeepEqual(files, want) {
		t.Errorf("files = %v, want %v", files, want)
	}

	if files, err := w.GetAttachmentsForTasks(context.Background(), "tbl", "docs", nil); err != nil || len(files) != 0 {
		t.Errorf("no tasks = %v, %v; want none", files, err)
	}
}
