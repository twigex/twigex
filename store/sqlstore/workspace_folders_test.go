// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package sqlstore

import (
	"context"
	"database/sql"
	"errors"
	"sync"
	"testing"

	"github.com/twigex/twigex/model"
)

// seedFolderTree gives workspace w1 the folders a > b > c and d, a table in c
// and a table at the root, and workspace w2 a folder of its own.
func seedFolderTree(t *testing.T) *workspaceRepository {
	t.Helper()
	requireDB(t)
	cleanTables(t, "workspaces", "workspace_folders", "workspace_tables")

	for _, id := range []string{"w1", "w2"} {
		mustExec(t, `INSERT INTO workspaces (id, title, description, start_date, end_date, pre_fix, created_at, updated_at, deleted_at, created_by)
			VALUES (?, ?, '', 0, 0, '', 0, 0, 0, 'u1')`, id, id)
	}
	for _, f := range [][3]string{
		{"a", "w1", ""},
		{"b", "w1", "a"},
		{"c", "w1", "b"},
		{"d", "w1", ""},
		{"other", "w2", ""},
	} {
		mustExec(t, `INSERT INTO workspace_folders (id, workspace_id, parent_folder_id, name, description, created_at, updated_at, deleted_at)
			VALUES (?, ?, ?, ?, '', 0, 0, 0)`, f[0], f[1], f[2], f[0])
	}
	for _, tb := range [][2]string{
		{"in-c", "c"},
		{"at-root", ""},
	} {
		mustExec(t, `INSERT INTO workspace_tables (id, workspace_id, name, display_name, linked, single_select, parent_table_id, both_direction_link, second_table_id, folder_id, deleted_at)
			VALUES (?, 'w1', ?, '', false, false, '', false, '', ?, 0)`, tb[0], tb[0], tb[1])
	}
	return &workspaceRepository{Db: testDB}
}

func folderOf(t *testing.T, table string) string {
	t.Helper()
	var folder string
	if err := testDB.QueryRow("SELECT folder_id FROM workspace_tables WHERE id = ?", table).Scan(&folder); err != nil {
		t.Fatal(err)
	}

	return folder
}

func parentOf(t *testing.T, folder string) string {
	t.Helper()
	var parent string
	if err := testDB.QueryRow("SELECT parent_folder_id FROM workspace_folders WHERE id = ?", folder).Scan(&parent); err != nil {
		t.Fatal(err)
	}

	return parent
}

func TestMoveTableIntoAFolderAndBackToTheRoot(t *testing.T) {
	w := seedFolderTree(t)
	ctx := context.Background()

	if err := w.MoveTable(ctx, "w1", "at-root", "b"); err != nil {
		t.Fatal(err)
	}

	if got := folderOf(t, "at-root"); got != "b" {
		t.Errorf("folder = %q, want b", got)
	}

	if err := w.MoveTable(ctx, "w1", "at-root", ""); err != nil {
		t.Fatal(err)
	}

	if got := folderOf(t, "at-root"); got != "" {
		t.Errorf("folder = %q, want the root", got)
	}
}

func TestMoveTableRefusesWhatIsNotLiveInTheWorkspace(t *testing.T) {
	w := seedFolderTree(t)
	ctx := context.Background()
	mustExec(t, "UPDATE workspace_folders SET deleted_at = 1 WHERE id = 'd'")

	for _, tc := range []struct {
		name, workspace, table, folder string
	}{
		{"a folder of another workspace", "w1", "in-c", "other"},
		{"a deleted folder", "w1", "in-c", "d"},
		{"a folder that does not exist", "w1", "in-c", "missing"},
		{"a table of another workspace", "w2", "in-c", ""},
	} {
		if err := w.MoveTable(ctx, tc.workspace, tc.table, tc.folder); !errors.Is(err, sql.ErrNoRows) {
			t.Errorf("%s: %v, want no rows", tc.name, err)
		}
	}
	if got := folderOf(t, "in-c"); got != "c" {
		t.Errorf("a refused move still moved the table to %q", got)
	}
}

func TestMoveFolderRefusesToGoInsideItself(t *testing.T) {
	w := seedFolderTree(t)
	ctx := context.Background()

	for _, target := range []string{"a", "b", "c"} {
		if err := w.MoveFolder(ctx, "w1", "a", target); !errors.Is(err, model.ErrFolderCycle) {
			t.Errorf("a into %s: %v, want a cycle", target, err)
		}
	}
	if err := w.MoveFolder(ctx, "w1", "a", "other"); !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("into another workspace's folder: %v, want no rows", err)
	}
	if got := parentOf(t, "a"); got != "" {
		t.Errorf("a refused move still moved a under %q", got)
	}

	if err := w.MoveFolder(ctx, "w1", "c", "d"); err != nil {
		t.Fatal(err)
	}

	if err := w.MoveFolder(ctx, "w1", "b", ""); err != nil {
		t.Fatal(err)
	}

	if got := parentOf(t, "c"); got != "d" {
		t.Errorf("c is under %q, want d", got)
	}
	if got := parentOf(t, "b"); got != "" {
		t.Errorf("b is under %q, want the root", got)
	}
}

func TestMovingTwoFoldersIntoEachOtherAtOnceLetsOnlyOneThrough(t *testing.T) {
	w := seedFolderTree(t)
	ctx := context.Background()

	for range 10 {
		mustExec(t, "UPDATE workspace_folders SET parent_folder_id = '' WHERE id IN ('a', 'd')")

		var wg sync.WaitGroup
		errs := make([]error, 2)
		for i, move := range [][2]string{{"a", "d"}, {"d", "a"}} {
			wg.Add(1)
			go func() {
				defer wg.Done()
				errs[i] = w.MoveFolder(ctx, "w1", move[0], move[1])
			}()
		}
		wg.Wait()

		cycles := 0
		for _, err := range errs {
			switch {
			case errors.Is(err, model.ErrFolderCycle):
				cycles++
			case err != nil:
				t.Fatal(err)
			}
		}
		if cycles != 1 {
			t.Fatalf("%d of the two moves were refused, want one; a under %q, d under %q",
				cycles, parentOf(t, "a"), parentOf(t, "d"))
		}
	}
}

func TestCreateFolderOnlyInsideALiveFolderOfTheWorkspace(t *testing.T) {
	w := seedFolderTree(t)
	ctx := context.Background()
	mustExec(t, "UPDATE workspace_folders SET deleted_at = 1 WHERE id = 'd'")

	for _, tc := range []struct {
		name, parent string
	}{
		{"a folder of another workspace", "other"},
		{"a deleted folder", "d"},
		{"a folder that does not exist", "missing"},
	} {
		if _, err := w.CreateFolder(ctx, "w1", "New", tc.parent, "new-"+tc.parent); !errors.Is(err, sql.ErrNoRows) {
			t.Errorf("inside %s: %v, want no rows", tc.name, err)
		}
	}

	var refused int
	if err := testDB.QueryRow("SELECT COUNT(*) FROM workspace_folders WHERE id LIKE 'new-%'").Scan(&refused); err != nil {
		t.Fatal(err)
	}

	if refused != 0 {
		t.Errorf("%d refused folders were still created", refused)
	}

	for _, parent := range []string{"b", ""} {
		folder, err := w.CreateFolder(ctx, "w1", "New", parent, "made-in-"+parent)
		if err != nil {
			t.Fatal(err)
		}

		if folder.ParentFolderID != parent || parentOf(t, folder.ID) != parent {
			t.Errorf("created under %q, want %q", folder.ParentFolderID, parent)
		}
	}
}

func TestDeleteFolderRefusesATableAnywhereInsideUnlessTablesMayGo(t *testing.T) {
	w := seedFolderTree(t)
	ctx := context.Background()

	for _, folder := range []string{"a", "b", "c"} {
		if _, _, err := w.DeleteFolder(ctx, "w1", folder, false); !errors.Is(err, model.ErrFolderHasTables) {
			t.Errorf("deleting %s: %v, want it refused", folder, err)
		}
	}

	if err := w.MoveTable(ctx, "w1", "in-c", ""); err != nil {
		t.Fatal(err)
	}

	if _, _, err := w.DeleteFolder(ctx, "w1", "a", false); err != nil {
		t.Fatal(err)
	}

	var live int
	if err := testDB.QueryRow(`SELECT COUNT(*) FROM workspace_folders
		WHERE workspace_id = 'w1' AND deleted_at = 0`).Scan(&live); err != nil {
		t.Fatal(err)
	}
	if live != 1 {
		t.Errorf("%d folders are left, want only d: the folders inside a go with it", live)
	}
	if _, _, err := w.DeleteFolder(ctx, "w1", "b", true); !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("deleting a folder already gone: %v, want no rows", err)
	}
	if _, _, err := w.DeleteFolder(ctx, "w2", "d", true); !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("deleting another workspace's folder: %v, want no rows", err)
	}
}

func TestDeleteFolderWithTablesDeletesTheTablesInsideAndTheirFields(t *testing.T) {
	w := seedFolderTree(t)
	ctx := context.Background()
	cleanTables(t, "workspace_fields")

	for _, f := range [][2]string{
		{"f-in-c", "in-c"},
		{"f-at-root", "at-root"},
	} {
		mustExec(t, `INSERT INTO workspace_fields (id, workspace_id, table_id, field_name, field_display_name, field_type, parent_field_id, created_at, updated_at, deleted_at)
			VALUES (?, 'w1', ?, ?, ?, 'text', NULL, 1, 0, 0)`, f[0], f[1], f[0], f[0])
	}

	tables, _, err := w.DeleteFolder(ctx, "w1", "a", true)
	if err != nil {
		t.Fatal(err)
	}

	if len(tables) != 1 || tables[0] != "in-c" {
		t.Errorf("deleted tables = %q, want only in-c, the one inside", tables)
	}

	deleted := func(query, id string) bool {
		t.Helper()
		var at int64
		if err := testDB.QueryRow(query, id).Scan(&at); err != nil {
			t.Fatal(err)
		}

		return at != 0
	}
	for id, want := range map[string]bool{"in-c": true, "at-root": false} {
		if got := deleted("SELECT deleted_at FROM workspace_tables WHERE id = ?", id); got != want {
			t.Errorf("table %s deleted = %t, want %t", id, got, want)
		}
	}
	for id, want := range map[string]bool{"f-in-c": true, "f-at-root": false} {
		if got := deleted("SELECT deleted_at FROM workspace_fields WHERE id = ?", id); got != want {
			t.Errorf("field %s deleted = %t, want %t", id, got, want)
		}
	}
	for id, want := range map[string]bool{"a": true, "b": true, "c": true, "d": false} {
		if got := deleted("SELECT deleted_at FROM workspace_folders WHERE id = ?", id); got != want {
			t.Errorf("folder %s deleted = %t, want %t", id, got, want)
		}
	}
}
