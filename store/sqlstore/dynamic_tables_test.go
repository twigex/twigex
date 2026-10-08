// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package sqlstore

import (
	"context"
	"sort"
	"strings"
	"testing"
)

func tableIndexes(t *testing.T, table string) string {
	t.Helper()
	rows, err := testDB.Query(`SELECT INDEX_NAME, GROUP_CONCAT(COLUMN_NAME ORDER BY SEQ_IN_INDEX), MIN(NON_UNIQUE) = 0
		FROM information_schema.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ?
		GROUP BY INDEX_NAME`, table)
	if err != nil {
		t.Fatalf("read indexes of %s: %v", table, err)
	}
	defer rows.Close()

	var got []string
	for rows.Next() {
		var name, cols string
		var unique bool
		if err := rows.Scan(&name, &cols, &unique); err != nil {
			t.Fatalf("scan index: %v", err)
		}

		switch {
		case name == "PRIMARY":
			got = append(got, "PRIMARY("+cols+")")
		case unique:
			got = append(got, "UNIQUE("+cols+")")
		default:
			got = append(got, "("+cols+")")
		}
	}
	sort.Strings(got)
	return strings.Join(got, " ")
}

func createDynamicTable(t *testing.T, name string, create func(*workspaceRepository, string) error) {
	t.Helper()
	mustExec(t, "DROP TABLE IF EXISTS `"+name+"`")
	t.Cleanup(func() { _, _ = testDB.Exec("DROP TABLE IF EXISTS `" + name + "`") })
	if err := create(&workspaceRepository{Db: testDB}, name); err != nil {
		t.Fatalf("create %s: %v", name, err)
	}
}

func TestDynamicTablesGetOnlyTheirIntendedIndexes(t *testing.T) {
	requireDB(t)

	createDynamicTable(t, "zz_test_link", func(w *workspaceRepository, name string) error {
		tx, err := w.Db.Begin()
		if err != nil {
			return err
		}
		defer tx.Rollback()
		if err := w.createLinkTableDDLTx(tx, name); err != nil {
			return err
		}
		return tx.Commit()
	})
	if got, want := tableIndexes(t, "zz_test_link"), "(parent_table_item_id) PRIMARY(id) UNIQUE(table_item_id,parent_table_item_id)"; got != want {
		t.Errorf("link table indexes = %s, want %s", got, want)
	}
	mustExec(t, "INSERT INTO `zz_test_link` VALUES ('l1', 't', 'a', 'p', 'b')")
	if _, err := testDB.Exec("INSERT INTO `zz_test_link` VALUES ('l2', 't', 'a', 'p', 'b')"); err == nil {
		t.Error("a link table took the same link twice")
	}

	createDynamicTable(t, "zz_test_task", func(w *workspaceRepository, name string) error {
		tx, err := w.Db.Begin()
		if err != nil {
			return err
		}
		defer tx.Rollback()
		if err := w.createTaskTableDDLTx(tx, name); err != nil {
			return err
		}
		return tx.Commit()
	})
	want := "(assignee) (created_at) (created_by) (due_date) (parent_task_id,deleted_at) (start_date) (status) (updated_at) PRIMARY(id)"
	if got := tableIndexes(t, "zz_test_task"); got != want {
		t.Errorf("task table indexes = %s, want %s", got, want)
	}
}

func registerDynamicTable(t *testing.T, name string, linked, singleSelect bool) {
	t.Helper()
	mustExec(t, `INSERT INTO workspace_tables (id, workspace_id, name, display_name, linked, single_select, parent_table_id, both_direction_link, second_table_id, folder_id, deleted_at)
		VALUES (?, 'w1', ?, '', ?, ?, '', false, '', '', 0)`, "id-"+name, name, linked, singleSelect)
}

func TestGetTableWorkspaceID(t *testing.T) {
	requireDB(t)
	cleanTables(t, "workspace_tables")
	registerDynamicTable(t, "zz_owned", false, false)
	w := &workspaceRepository{Db: testDB}

	if got, err := w.GetTableWorkspaceID(context.Background(), "id-zz_owned"); err != nil || got != "w1" {
		t.Errorf("owned table: %q, %v; want w1", got, err)
	}
	if got, err := w.GetTableWorkspaceID(context.Background(), "no-such-table"); err != nil || got != "" {
		t.Errorf("missing table: %q, %v; want empty", got, err)
	}
}

func TestCreateIndexesReportsFailures(t *testing.T) {
	requireDB(t)
	mustExec(t, "DROP TABLE IF EXISTS `zz_test_bad_index`")
	mustExec(t, "CREATE TABLE `zz_test_bad_index` (id VARCHAR(36) NOT NULL PRIMARY KEY)")
	t.Cleanup(func() { _, _ = testDB.Exec("DROP TABLE IF EXISTS `zz_test_bad_index`") })

	tx, err := testDB.Begin()
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	defer tx.Rollback()

	if err := createIndexesTx(tx, "zz_test_bad_index", dynamicTableLayout{indexes: [][]string{{"no_such_column"}}}); err == nil {
		t.Fatal("an index on a missing column was reported as created")
	}
}
