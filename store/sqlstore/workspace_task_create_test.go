// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package sqlstore

import (
	"testing"
)

func TestCreateTaskWritesItsFieldsWithIt(t *testing.T) {
	requireDB(t)
	cleanTables(t, "workspace_tables")
	w := &workspaceRepository{Db: testDB}

	const table = "zz_create_task"
	mustExec(t, "DROP TABLE IF EXISTS `"+table+"`")
	if err := createTaskTable(w, table); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _, _ = testDB.Exec("DROP TABLE IF EXISTS `" + table + "`") })
	mustExec(t, `INSERT INTO workspace_tables (id, workspace_id, name, display_name, linked, single_select, parent_table_id, both_direction_link, second_table_id, folder_id, deleted_at)
		VALUES ('t1', 'w1', ?, '', false, false, '', false, '', '', 0)`, table)

	task, err := w.CreateTask("w1", "t1", "Launch", "u1", "", "task-1", 1000, map[string]any{
		"status":      "todo",
		"assignee":    "u2",
		"start_date":  int64(100),
		"due_date":    int64(200),
		"description": "Notes",
	})
	if err != nil {
		t.Fatal(err)
	}

	for column, want := range map[string]string{
		"name":        "Launch",
		"status":      "todo",
		"assignee":    "u2",
		"start_date":  "100",
		"due_date":    "200",
		"description": "Notes",
		"created_by":  "u1",
	} {
		var got string
		if err := testDB.QueryRow("SELECT CAST(`" + column + "` AS CHAR) FROM `" + table + "` WHERE id = 'task-1'").Scan(&got); err != nil {
			t.Fatal(err)
		}

		if got != want {
			t.Errorf("%s = %q, want %q", column, got, want)
		}
	}
	if (*task)["id"] != "task-1" {
		t.Errorf("returned %v, want the new task", *task)
	}

	// Bytes would reach the browser as base64.
	for _, column := range []string{"name", "description", "status"} {
		if _, ok := (*task)[column].(string); !ok {
			t.Errorf("returned %s as %T, want a string", column, (*task)[column])
		}
	}

	if _, err := w.CreateTask("w1", "t1", "Plain", "u1", "", "task-2", 1000, nil); err != nil {
		t.Fatalf("a task without fields: %v", err)
	}
}
