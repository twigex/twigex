// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package sqlstore

import (
	"context"
	"errors"
	"slices"
	"strconv"
	"testing"

	"github.com/twigex/twigex/model"
)

func setupTasksAcrossTables(t *testing.T) []model.TaskSource {
	t.Helper()
	requireDB(t)
	createDynamicTable(t, "zz_test_across_a", createTaskTable)
	createDynamicTable(t, "zz_test_across_b", createTaskTable)

	for _, r := range []struct {
		table, id, assignee string
		due                 any
		deleted             int
	}{
		{"zz_test_across_a", "a1", "me", 100, 0},
		{"zz_test_across_a", "a2", "me", 300, 0},
		{"zz_test_across_a", "a3", "me", nil, 0},
		{"zz_test_across_a", "a4", "someone", 150, 0},
		{"zz_test_across_a", "a5", "me", 120, 1},
		{"zz_test_across_b", "b1", "me", 200, 0},
		{"zz_test_across_b", "b2", "me", 50, 0},
		{"zz_test_across_b", "b3", "me", 0, 0},
	} {
		mustExec(t, "INSERT INTO `"+r.table+"` (id, name, assignee, due_date, deleted_at) VALUES (?, ?, ?, ?, ?)",
			r.id, "task "+r.id, r.assignee, r.due, r.deleted)
	}

	mine := model.SQLFilter{SQL: "main.assignee = ?", Args: []any{"me"}}
	return []model.TaskSource{
		{TableID: "table-a", TableName: "zz_test_across_a", Filter: mine},
		{TableID: "table-b", TableName: "zz_test_across_b", Filter: mine},
	}
}

func orderedIDs(rows []map[string]any) []string {
	ids := make([]string, 0, len(rows))
	for _, r := range rows {
		ids = append(ids, r["id"].(string))
	}
	return ids
}

// rowValue reads a whole number column from a scanned row, as a cursor would.
func rowValue(t *testing.T, v any) *int64 {
	t.Helper()
	switch v := v.(type) {
	case nil:
		return nil
	case int64:
		return &v
	case string:
		n, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			t.Fatalf("value %q: %v", v, err)
		}

		return &n
	}
	t.Fatalf("value %v of type %T", v, v)
	return nil
}

func TestGetTasksAcrossTablesPagesInDueDateOrder(t *testing.T) {
	sources := setupTasksAcrossTables(t)
	w := &workspaceRepository{Db: testDB}

	rows, err := w.GetTasksAcrossTables(context.Background(), sources, 10)
	if err != nil {
		t.Fatal(err)
	}

	all := []string{"a3", "b3", "b2", "a1", "b1", "a2"}
	if got := orderedIDs(rows); !slices.Equal(got, all) {
		t.Fatalf("all tasks = %v, want %v", got, all)
	}
	if rows[2]["table_id"] != "table-b" || rows[2]["name"] != "task b2" {
		t.Errorf("row = %v, want b2 from table-b with its name", rows[2])
	}

	for _, limit := range []int{1, 2, 4} {
		var got []string
		page := sources
		for range len(all) + 1 {
			rows, err := w.GetTasksAcrossTables(context.Background(), page, limit)
			if err != nil {
				t.Fatal(err)
			}

			got = append(got, orderedIDs(rows)...)
			if len(rows) < limit {
				break
			}

			last := rows[len(rows)-1]
			after := model.RowsAfter("main.due_date", rowValue(t, last["due_date"]), last["id"].(string))
			page = make([]model.TaskSource, len(sources))
			for i, s := range sources {
				s.Filter = s.Filter.And(after)
				page[i] = s
			}
		}
		if !slices.Equal(got, all) {
			t.Errorf("pages of %d = %v, want %v", limit, got, all)
		}
	}
}

func TestCountTasksAcrossTablesCountsEachTaskInItsFirstGroup(t *testing.T) {
	sources := setupTasksAcrossTables(t)
	w := &workspaceRepository{Db: testDB}

	counts, err := w.CountTasksAcrossTables(context.Background(), sources, []model.SQLFilter{
		{SQL: "main.due_date > 0 AND main.due_date < ?", Args: []any{150}},
		{SQL: "main.due_date >= ?", Args: []any{100}},
		{SQL: "(main.due_date IS NULL OR main.due_date = 0)"},
	})
	if err != nil {
		t.Fatal(err)
	}

	if want := []int{2, 2, 2}; !slices.Equal(counts, want) {
		t.Errorf("counts = %v, want %v", counts, want)
	}
}

func TestTasksAcrossTablesWithNoTables(t *testing.T) {
	requireDB(t)
	w := &workspaceRepository{Db: testDB}
	rows, err := w.GetTasksAcrossTables(context.Background(), nil, 10)
	if err != nil || len(rows) != 0 {
		t.Errorf("rows = %v, %v; want none", rows, err)
	}

	counts, err := w.CountTasksAcrossTables(context.Background(), nil, []model.SQLFilter{{SQL: "1 = 1"}})
	if err != nil || !slices.Equal(counts, []int{0}) {
		t.Errorf("counts = %v, %v; want [0]", counts, err)
	}
}

func TestGetTableRowsPagePutsListedRowsFirstThenByCreation(t *testing.T) {
	requireDB(t)
	createDynamicTable(t, "zz_test_rows_page", createTaskTable)
	for _, r := range []struct {
		id, status string
		created    any
	}{
		{"old", "s1", 1}, {"mid", "s1", 2}, {"new", "s1", 3}, {"other", "s2", 0}, {"gone", "s1", 4},
		{"unknown-b", "s1", nil}, {"unknown-a", "s1", nil}, {"same-b", "s1", 2}, {"same-a", "s1", 2},
	} {
		deleted := 0
		if r.id == "gone" {
			deleted = 1
		}
		mustExec(t, "INSERT INTO `zz_test_rows_page` (id, name, status, created_at, deleted_at) VALUES (?, ?, ?, ?, ?)", r.id, r.id, r.status, r.created, deleted)
	}
	w := &workspaceRepository{Db: testDB}
	inS1 := model.SQLFilter{SQL: "main.status = ?", Args: []any{"s1"}}
	byCreation := []string{"unknown-a", "unknown-b", "old", "mid", "same-a", "same-b", "new"}

	for _, tc := range []struct {
		first []string
		limit int
		want  []string
	}{
		{nil, 10, byCreation},
		{[]string{"new", "missing", "old"}, 10, []string{"new", "old", "unknown-a", "unknown-b", "mid", "same-a", "same-b"}},
		{[]string{"new", "old"}, 1, []string{"new"}},
	} {
		rows, err := w.GetTableRowsPage(context.Background(), "zz_test_rows_page", inS1, tc.first, tc.limit)
		if err != nil {
			t.Fatalf("%v: %v", tc.first, err)
		}

		if got := orderedIDs(rows); !slices.Equal(got, tc.want) {
			t.Errorf("first %v limit %d = %v, want %v", tc.first, tc.limit, got, tc.want)
		}
	}

	for _, limit := range []int{1, 2, 3} {
		var got []string
		filter := inS1
		for range len(byCreation) + 1 {
			rows, err := w.GetTableRowsPage(context.Background(), "zz_test_rows_page", filter, nil, limit)
			if err != nil {
				t.Fatal(err)
			}

			got = append(got, orderedIDs(rows)...)
			if len(rows) < limit {
				break
			}
			last := rows[len(rows)-1]
			filter = inS1.And(model.RowsAfter("main.created_at", rowValue(t, last["created_at"]), last["id"].(string)))
		}
		if !slices.Equal(got, byCreation) {
			t.Errorf("pages of %d = %v, want %v", limit, got, byCreation)
		}
	}
}

func TestTaskOrParentMatchesFollowsWhatTheGridShows(t *testing.T) {
	requireDB(t)
	createDynamicTable(t, "zz_test_shown", createTaskTable)
	for _, r := range []struct {
		id, status, parent string
		deleted            int
	}{
		{"open-root", "open", "", 0},
		{"done-child", "done", "open-root", 0},
		{"done-grandchild", "done", "done-child", 0},
		{"done-root", "done", "", 0},
		{"open-deleted", "open", "", 1},
		{"loop-a", "done", "loop-b", 0},
		{"loop-b", "done", "loop-a", 0},
	} {
		mustExec(t, "INSERT INTO `zz_test_shown` (id, name, status, parent_task_id, deleted_at) VALUES (?, ?, ?, NULLIF(?, ''), ?)", r.id, r.id, r.status, r.parent, r.deleted)
	}
	w := &workspaceRepository{Db: testDB}
	open := model.SQLFilter{SQL: "main.status = ?", Args: []any{"open"}}

	for id, want := range map[string]bool{
		"open-root":       true,
		"done-child":      true,
		"done-grandchild": true,
		"done-root":       false,
		"open-deleted":    false,
		"missing":         false,
		"loop-a":          false,
	} {
		got, err := w.TaskOrParentMatches(context.Background(), "zz_test_shown", id, open)
		if err != nil {
			t.Fatalf("%s: %v", id, err)
		}

		if got != want {
			t.Errorf("%s shown = %v, want %v", id, got, want)
		}
	}
	if got, err := w.TaskOrParentMatches(context.Background(), "zz_test_shown", "done-root", model.SQLFilter{}); err != nil || !got {
		t.Errorf("with no filter = %v, %v; want shown", got, err)
	}
}

func TestStreamTableRowsHandsOverMatchingRowsInOrder(t *testing.T) {
	requireDB(t)
	createDynamicTable(t, "zz_test_stream", createTaskTable)
	for _, r := range []struct {
		id, status   string
		due, deleted int
	}{
		{"late", "open", 300, 0}, {"early", "open", 100, 0}, {"done", "done", 200, 0}, {"gone", "open", 50, 1},
	} {
		mustExec(t, "INSERT INTO `zz_test_stream` (id, name, status, due_date, deleted_at) VALUES (?, ?, ?, ?, ?)", r.id, r.id, r.status, r.due, r.deleted)
	}
	w := &workspaceRepository{Db: testDB}
	open := model.SQLFilter{SQL: "main.status = ?", Args: []any{"open"}}

	var got []string
	err := w.StreamTableRows(context.Background(), "zz_test_stream", open, "ORDER BY main.due_date", func(row map[string]interface{}) error {
		got = append(got, row["id"].(string))
		return nil
	})
	if err != nil || !slices.Equal(got, []string{"early", "late"}) {
		t.Errorf("rows = %v, %v; want early then late", got, err)
	}

	stop := errors.New("stop")
	calls := 0
	err = w.StreamTableRows(context.Background(), "zz_test_stream", model.SQLFilter{}, "ORDER BY main.due_date", func(map[string]interface{}) error {
		calls++
		return stop
	})
	if !errors.Is(err, stop) || calls != 1 {
		t.Errorf("after an error from fn = %v with %d calls, want it returned after one", err, calls)
	}
}
