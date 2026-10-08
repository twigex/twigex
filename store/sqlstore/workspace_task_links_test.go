// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package sqlstore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"sync"
	"testing"

	"github.com/twigex/twigex/model"
)

// setupTaskLinks builds tasks a1..a2 and b1..b9 in two task tables, with a
// "related" field on A linking to B both ways.
func setupTaskLinks(t *testing.T) *workspaceRepository {
	t.Helper()
	requireDB(t)
	cleanTables(t, "workspace_tables", "workspace_relationships")

	createDynamicTable(t, "zz_links_a", createTaskTable)
	createDynamicTable(t, "zz_links_b", createTaskTable)
	for _, name := range []string{"zz_links_ab", "zz_links_ba"} {
		createDynamicTable(t, name, func(w *workspaceRepository, name string) error {
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
	}

	for _, tbl := range []struct {
		id, name, parent, second string
		linked                   bool
	}{
		{"t-a", "zz_links_a", "", "", false},
		{"t-b", "zz_links_b", "", "", false},
		{"l-ab", "zz_links_ab", "t-b", "l-ba", true},
		{"l-ba", "zz_links_ba", "t-a", "l-ab", true},
	} {
		mustExec(t, `INSERT INTO workspace_tables
			(id, workspace_id, name, display_name, linked, single_select, parent_table_id, both_direction_link, second_table_id, folder_id, deleted_at)
			VALUES (?, 'ws', ?, '', ?, 0, ?, ?, ?, '', 0)`,
			tbl.id, tbl.name, tbl.linked, tbl.parent, tbl.linked, tbl.second)
	}
	mustExec(t, `INSERT INTO workspace_relationships (id, workspace_id, table_id, linked_table_id, table_name, created_at, deleted_at)
		VALUES ('r1', 'ws', 't-a', 'l-ab', 'related', 1, 0), ('r2', 'ws', 't-b', 'l-ba', 'related_back', 1, 0)`)

	for _, id := range []string{"a1", "a2"} {
		mustExec(t, "INSERT INTO zz_links_a (id, name, deleted_at) VALUES (?, ?, 0)", id, "task "+id)
	}
	for i := 1; i <= 9; i++ {
		id := fmt.Sprintf("b%d", i)
		mustExec(t, "INSERT INTO zz_links_b (id, name, deleted_at) VALUES (?, ?, 0)", id, "task "+id)
	}
	return &workspaceRepository{Db: testDB}
}

// linksOf lists what a1 links to, and checks each link has its reverse.
func linksOf(t *testing.T, task string) []string {
	t.Helper()
	rows, err := testDB.Query("SELECT parent_table_item_id FROM zz_links_ab WHERE table_item_id = ? ORDER BY parent_table_item_id", task)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			t.Fatal(err)
		}

		ids = append(ids, id)
	}

	var reverse []string
	back, err := testDB.Query("SELECT table_item_id FROM zz_links_ba WHERE parent_table_item_id = ? ORDER BY table_item_id", task)
	if err != nil {
		t.Fatal(err)
	}
	defer back.Close()
	for back.Next() {
		var id string
		if err := back.Scan(&id); err != nil {
			t.Fatal(err)
		}

		reverse = append(reverse, id)
	}
	if !slices.Equal(ids, reverse) {
		t.Errorf("%s links %v, but is linked back from %v", task, ids, reverse)
	}
	return ids
}

func TestChangeTaskLinksLeavesLinksItIsNotGiven(t *testing.T) {
	w := setupTaskLinks(t)
	ctx := context.Background()
	change := func(add, remove []string) error {
		_, err := w.ChangeTaskLinks(ctx, "ws", "t-a", "a1", "related", add, remove)
		return err
	}

	if err := change([]string{"b1", "b2", "b3", "b4", "b5"}, nil); err != nil {
		t.Fatal(err)
	}

	if err := change([]string{"b6"}, nil); err != nil {
		t.Fatal(err)
	}

	if got, want := linksOf(t, "a1"), []string{"b1", "b2", "b3", "b4", "b5", "b6"}; !slices.Equal(got, want) {
		t.Fatalf("after adding b6, a1 links %v, want %v", got, want)
	}

	if err := change(nil, []string{"b2"}); err != nil {
		t.Fatal(err)
	}

	if got, want := linksOf(t, "a1"), []string{"b1", "b3", "b4", "b5", "b6"}; !slices.Equal(got, want) {
		t.Fatalf("after removing b2, a1 links %v, want %v", got, want)
	}

	if err := change([]string{"b1", "b1"}, []string{"b9"}); err != nil {
		t.Fatal(err)
	}

	var rows int
	if err := testDB.QueryRow("SELECT COUNT(*) FROM zz_links_ab WHERE table_item_id = 'a1'").Scan(&rows); err != nil {
		t.Fatal(err)
	}

	if rows != 5 {
		t.Errorf("re-adding a link and removing one that is not there left %d rows, want 5", rows)
	}
	if got := linksOf(t, "a2"); len(got) != 0 {
		t.Errorf("a2 links %v, want nothing", got)
	}
}

func TestChangeTaskLinksRefusesWhatItCannotCheck(t *testing.T) {
	w := setupTaskLinks(t)
	ctx := context.Background()
	if _, err := w.ChangeTaskLinks(ctx, "ws", "t-a", "a1", "related", []string{"b1"}, nil); err != nil {
		t.Fatal(err)
	}

	mustExec(t, "UPDATE zz_links_b SET deleted_at = 1 WHERE id = 'b8'")
	mustExec(t, "UPDATE zz_links_a SET deleted_at = 1 WHERE id = 'a2'")

	for _, tc := range []struct {
		name                   string
		workspace, table, task string
		field                  string
		add, remove            []string
		want                   error
	}{
		{"a row that does not exist", "ws", "t-a", "a1", "related", []string{"b7", "nope"}, []string{"b1"}, model.ErrLinkTargetMissing},
		{"a deleted row", "ws", "t-a", "a1", "related", []string{"b8"}, nil, model.ErrLinkTargetMissing},
		{"another workspace", "other", "t-a", "a1", "related", []string{"b7"}, nil, sql.ErrNoRows},
		{"a field of another table", "ws", "t-a", "a1", "related_back", []string{"b7"}, nil, sql.ErrNoRows},
		{"an unknown field", "ws", "t-a", "a1", "missing", []string{"b7"}, nil, sql.ErrNoRows},
		{"a deleted task", "ws", "t-a", "a2", "related", []string{"b7"}, nil, sql.ErrNoRows},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := w.ChangeTaskLinks(ctx, tc.workspace, tc.table, tc.task, tc.field, tc.add, tc.remove)
			if !errors.Is(err, tc.want) {
				t.Errorf("error = %v, want %v", err, tc.want)
			}
		})
	}

	if got := linksOf(t, "a1"); !slices.Equal(got, []string{"b1"}) {
		t.Errorf("after refused changes a1 links %v, want only b1", got)
	}
}

func TestChangeTaskLinksKeepsEveryConcurrentChange(t *testing.T) {
	w := setupTaskLinks(t)

	var wg sync.WaitGroup
	errs := make(chan error, 12)
	for i := 1; i <= 6; i++ {
		for _, id := range []string{fmt.Sprintf("b%d", i), "b9"} {
			wg.Add(1)
			go func(id string) {
				defer wg.Done()
				_, err := w.ChangeTaskLinks(context.Background(), "ws", "t-a", "a1", "related", []string{id}, nil)
				errs <- err
			}(id)
		}
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}

	if got, want := linksOf(t, "a1"), []string{"b1", "b2", "b3", "b4", "b5", "b6", "b9"}; !slices.Equal(got, want) {
		t.Errorf("after saves at the same time a1 links %v, want %v, each once", got, want)
	}
}

func TestLinkingFromBothEndsAtOnceStoresTheLinkOnce(t *testing.T) {
	w := setupTaskLinks(t)

	var wg sync.WaitGroup
	errs := make(chan error, 40)
	for range 10 {
		for _, change := range []struct{ table, task, field, target string }{
			{"t-a", "a1", "related", "b1"},
			{"t-b", "b1", "related_back", "a1"},
		} {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, err := w.ChangeTaskLinks(context.Background(), "ws", change.table, change.task, change.field, []string{change.target}, nil)
				errs <- err
			}()
		}
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}

	for _, table := range []string{"zz_links_ab", "zz_links_ba"} {
		var rows int
		if err := testDB.QueryRow("SELECT COUNT(*) FROM `" + table + "`").Scan(&rows); err != nil {
			t.Fatal(err)
		}

		if rows != 1 {
			t.Errorf("%s holds %d rows for the one link, want 1", table, rows)
		}
	}
}

func TestLinkedRecordsLitePagesAndSearchesOnTheServer(t *testing.T) {
	requireDB(t)
	createDynamicTable(t, "zz_lite_tasks", createTaskTable)
	for i, r := range []struct {
		id, name string
		deleted  int
	}{
		{"t1", "Design review", 0},
		{"t2", "Write docs", 0},
		{"t3", "Review 100% done", 0},
		{"t4", "Deleted review", 1},
		{"t5", "Plan", 0},
	} {
		mustExec(t, "INSERT INTO zz_lite_tasks (id, name, created_at, deleted_at) VALUES (?, ?, ?, ?)", r.id, r.name, i+1, r.deleted)
	}
	mustExec(t, "DROP TABLE IF EXISTS zz_lite_options")
	t.Cleanup(func() { _, _ = testDB.Exec("DROP TABLE IF EXISTS zz_lite_options") })
	mustExec(t, "CREATE TABLE zz_lite_options (id VARCHAR(36) PRIMARY KEY, name VARCHAR(255), color VARCHAR(255))")
	mustExec(t, "INSERT INTO zz_lite_options VALUES ('o1', 'Low', ''), ('o2', 'High', ''), ('o3', 'Medium', '')")

	w := &workspaceRepository{Db: testDB}
	ctx := context.Background()
	for _, tc := range []struct {
		name, table, search string
		ids                 []string
		limit, offset       int
		want                []string
	}{
		{"newest first, deleted left out", "zz_lite_tasks", "", nil, 10, 0, []string{"t5", "t3", "t2", "t1"}},
		{"one page", "zz_lite_tasks", "", nil, 2, 0, []string{"t5", "t3"}},
		{"the next page", "zz_lite_tasks", "", nil, 2, 2, []string{"t2", "t1"}},
		{"a search, by name", "zz_lite_tasks", "review", nil, 10, 0, []string{"t1", "t3"}},
		{"a search for a percent sign", "zz_lite_tasks", "100%", nil, 10, 0, []string{"t3"}},
		{"a percent sign alone matches it only", "zz_lite_tasks", "%", nil, 10, 0, []string{"t3"}},
		{"by id", "zz_lite_tasks", "", []string{"t1", "t4", "missing"}, 10, 0, []string{"t1"}},
		{"options by name", "zz_lite_options", "", nil, 10, 0, []string{"o2", "o1", "o3"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rows, err := w.GetLinkedRecordsLite(ctx, tc.table, tc.search, tc.ids, tc.limit, tc.offset, model.SQLFilter{})
			if err != nil {
				t.Fatal(err)
			}

			if got := orderedIDs(rows); !slices.Equal(got, tc.want) {
				t.Errorf("rows = %v, want %v", got, tc.want)
			}
		})
	}
}
