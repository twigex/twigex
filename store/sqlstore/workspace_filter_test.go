// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package sqlstore

import (
	"context"
	"slices"
	"sort"
	"testing"
	"time"

	"github.com/twigex/twigex/model"
)

const (
	filterTaskTable   = "zz_test_filter_task"
	filterLinkTable   = "zz_test_filter_link"
	filterParentTable = "zz_test_filter_parent"
)

var filterHeaders = []model.WorkspaceHeaders{
	{Name: "id"},
	{Name: "name"},
	{Name: "status"},
	{Name: "start_date"},
	{Name: "project", LinkedID: "link-table", ParentTableID: "parent-table"},
}

var filterTableNames = map[string]string{
	"link-table":   filterLinkTable,
	"parent-table": filterParentTable,
}

func setupFilterTables(t *testing.T) *workspaceRepository {
	t.Helper()
	requireDB(t)
	for name, create := range map[string]func(*workspaceRepository, string) error{
		filterTaskTable:   createTaskTable,
		filterParentTable: createTaskTable,
		filterLinkTable: func(w *workspaceRepository, name string) error {
			tx, err := w.Db.Begin()
			if err != nil {
				return err
			}
			defer tx.Rollback()
			if err := w.createLinkTableDDLTx(tx, name); err != nil {
				return err
			}
			return tx.Commit()
		},
	} {
		createDynamicTable(t, name, create)
	}

	for _, r := range []struct{ id, name, status, parent string }{
		{"t1", "O'Brien", "open", ""},
		{"t2", `C:\temp\`, "it's", ""},
		{"t3", "plain", "done", ""},
		{"t4", "child", "open", "t3"},
	} {
		mustExec(t, "INSERT INTO `"+filterTaskTable+"` (id, name, status, start_date, parent_task_id, deleted_at) VALUES (?, ?, ?, 100, NULLIF(?, ''), 0)",
			r.id, r.name, r.status, r.parent)
	}
	mustExec(t, "INSERT INTO `"+filterParentTable+"` (id, name, deleted_at) VALUES ('p1', 'Project', 0)")
	mustExec(t, "INSERT INTO `"+filterLinkTable+"` (id, table_id, table_item_id, parent_table_id, parent_table_item_id) VALUES ('l1', 'task-table', 't1', 'parent-table', 'p1')")
	return &workspaceRepository{Db: testDB}
}

func createTaskTable(w *workspaceRepository, name string) error {
	tx, err := w.Db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := w.createTaskTableDDLTx(tx, name); err != nil {
		return err
	}
	return tx.Commit()
}

func rowIDs(rows []map[string]any) []string {
	ids := make([]string, 0, len(rows))
	for _, r := range rows {
		ids = append(ids, r["id"].(string))
	}
	sort.Strings(ids)
	return ids
}

func flat(filters ...model.Filter) model.FilterPayload {
	return model.FilterPayload{FlatFilters: filters}
}

func is(field, value string) model.Filter {
	return model.Filter{Field: field, Operator: "is", Value: value, OperatorBetween: "AND"}
}

func TestInlineFilterBindsValuesAndIgnoresUnknownFields(t *testing.T) {
	w := setupFilterTables(t)
	orInjection := "OR 1=1 OR"

	cases := []struct {
		name    string
		filters model.FilterPayload
		want    []string
	}{
		{"quote in value", flat(is("name", "O'Brien")), []string{"t1"}},
		{"backslashes in value", flat(is("name", `C:\temp\`)), []string{"t2"}},
		{"backslash before quote", flat(is("name", `\' OR 1=1 -- `)), []string{}},
		{"unknown field", flat(is("nope", "x")), []string{"t1", "t2", "t3", "t4"}},
		{"unknown field beside a known one", flat(is("nope", "x"), is("name", "plain")), []string{"t3"}},
		{"linked parent", flat(is("project", "p1")), []string{"t1"}},
		{"linked parent injection", flat(is("project", `\' OR 1=1 -- `)), []string{}},
		{"relation to next", model.FilterPayload{Groups: []model.FilterGroup{
			{Filters: []model.Filter{is("name", "O'Brien")}, RelationToNext: &orInjection},
			{Filters: []model.Filter{is("name", "plain")}},
		}}, []string{}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rows, err := w.GetTableRowsFiltered(context.Background(), filterTaskTable,
				w.BuildFilterSQLInline(tc.filters, filterHeaders, nil, filterTableNames))
			if err != nil {
				t.Fatalf("query: %v", err)
			}
			if got := rowIDs(rows); !slices.Equal(got, tc.want) {
				t.Errorf("rows = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestStatusFilterBindsValuesAndIgnoresUnknownFields(t *testing.T) {
	w := setupFilterTables(t)

	cases := []struct {
		name    string
		filters model.FilterPayload
		want    []string
	}{
		{"backslashes in value", flat(is("name", `C:\temp\`)), []string{"t2"}},
		{"status list", flat(model.Filter{Field: "status", Operator: "is", Values: []string{"it's", "done"}}), []string{"t2", "t3"}},
		{"unmatched date value", flat(is("start_date", "x' OR '1'='1")), []string{}},
		{"unknown field beside a known one", flat(is("nope", "x"), is("name", "plain")), []string{"t3"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rows, err := w.GetTableRowsFiltered(context.Background(), filterTaskTable,
				w.BuildFilterSQLInline(tc.filters, filterHeaders, nil, filterTableNames))
			if err != nil {
				t.Fatalf("query: %v", err)
			}
			if got := rowIDs(rows); !slices.Equal(got, tc.want) {
				t.Errorf("rows = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestStatusFilterOnLinkedField(t *testing.T) {
	w := setupFilterTables(t)

	rows, err := w.GetTableRowsFiltered(context.Background(), filterTaskTable,
		w.BuildFilterSQLInline(flat(is("project", "p1")), filterHeaders, nil, filterTableNames))
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if got := rowIDs(rows); !slices.Equal(got, []string{"t1"}) {
		t.Errorf("rows = %v, want [t1]", got)
	}
}

func TestLaterThanTodayMatchesFutureDates(t *testing.T) {
	w := setupFilterTables(t)
	mustExec(t, "UPDATE `"+filterTaskTable+"` SET start_date = ? WHERE id = 't2'", time.Now().AddDate(0, 0, 10).Unix())
	filters := flat(is("start_date", "Later than Today"))

	rows, err := w.GetTableRowsFiltered(context.Background(), filterTaskTable,
		w.BuildFilterSQLInline(filters, filterHeaders, nil, filterTableNames))
	if err != nil {
		t.Fatalf("inline query: %v", err)
	}
	if got := rowIDs(rows); !slices.Equal(got, []string{"t2"}) {
		t.Errorf("inline rows = %v, want [t2]", got)
	}
}

func TestEffectiveRootsApplyTheFilterToTheParent(t *testing.T) {
	w := setupFilterTables(t)
	mustExec(t, "INSERT INTO `"+filterLinkTable+"` (id, table_id, table_item_id, parent_table_id, parent_table_item_id) VALUES ('l2', 'task-table', 't3', 'parent-table', 'p1')")
	const byID = "ORDER BY main.id ASC"

	ctx := context.Background()
	none := model.SQLFilter{}

	notP1 := w.BuildFilterSQLInline(flat(model.Filter{Field: "project", Operator: "is_not", Value: "p1"}), filterHeaders, nil, filterTableNames)
	if pos, err := w.GetBranchPosition(ctx, filterTaskTable, "t2", none, notP1, byID); err != nil || pos != 1 {
		t.Errorf("position under a linked filter = %d, %v; want 1", pos, err)
	}

	rows, total, _, err := w.GetRootTasksPagedWithSubtasks(ctx, filterTaskTable, none, notP1, byID, 10, 0, true)
	if got := rowIDs(rows); err != nil || total != 1 || !slices.Equal(got, []string{"t2"}) {
		t.Errorf("paged under a linked filter = %v total %d, %v; want [t2] total 1, t4 hidden with t3", got, total, err)
	}

	open := model.SQLFilter{SQL: "(main.status IS NULL OR main.status NOT IN (?))", Args: []any{"done"}}
	rows, total, _, err = w.GetRootTasksPagedWithSubtasks(ctx, filterTaskTable, none, open, byID, 10, 0, true)
	if got := rowIDs(rows); err != nil || total != 2 || !slices.Equal(got, []string{"t1", "t2"}) {
		t.Errorf("paged open tasks = %v total %d, %v; want [t1 t2] total 2, the open t4 hidden with its done parent", got, total, err)
	}
	if pos, err := w.GetBranchPosition(ctx, filterTaskTable, "t4", none, open, byID); err != nil || pos != 0 {
		t.Errorf("position of a subtask hidden with its parent = %d, %v; want 0", pos, err)
	}

	rows, total, _, err = w.GetRootTasksPagedWithSubtasks(ctx, filterTaskTable, none, none, byID, 10, 0, true)
	if got := rowIDs(rows); err != nil || total != 3 || !slices.Equal(got, []string{"t1", "t2", "t3", "t4"}) {
		t.Errorf("unfiltered = %v total %d, %v; want every task in 3 branches", got, total, err)
	}
	if pos, err := w.GetBranchPosition(ctx, filterTaskTable, "t4", none, none, byID); err != nil || pos != 3 {
		t.Errorf("position of t4 unfiltered = %d, %v; want its parent's 3", pos, err)
	}
}

func TestAssignedOnlySubtaskUnderAnotherUsersTaskIsItsOwnBranch(t *testing.T) {
	w := setupFilterTables(t)
	mustExec(t, "UPDATE `"+filterTaskTable+"` SET assignee = CASE id WHEN 't4' THEN 'anna' ELSE 'bob' END")
	const byID = "ORDER BY main.id ASC"
	ctx := context.Background()
	anna := model.SQLFilter{SQL: "`assignee` = ?", Args: []any{"anna"}}
	open := model.SQLFilter{SQL: "(main.status IS NULL OR main.status NOT IN (?))", Args: []any{"done"}}

	rows, total, _, err := w.GetRootTasksPagedWithSubtasks(ctx, filterTaskTable, anna, open, byID, 10, 0, true)
	if got := rowIDs(rows); err != nil || total != 1 || !slices.Equal(got, []string{"t4"}) {
		t.Errorf("anna's open tasks = %v total %d, %v; want [t4], bob's t3 never shown", got, total, err)
	}
	if pos, err := w.GetBranchPosition(ctx, filterTaskTable, "t4", anna, open, byID); err != nil || pos != 1 {
		t.Errorf("position of t4 = %d, %v; want 1", pos, err)
	}
}

func TestOpenSubtasksAtEveryDepth(t *testing.T) {
	w := setupFilterTables(t)
	mustExec(t, "INSERT INTO `"+filterTaskTable+"` (id, name, status, start_date, parent_task_id, deleted_at) VALUES ('t5', 'grandchild', 'done', 100, 't4', 0), ('t6', 'deep', NULL, 100, 't5', 0), ('t7', 'gone', 'open', 100, 't3', 1)")
	mustExec(t, "UPDATE `"+filterTaskTable+"` SET assignee = CASE id WHEN 't6' THEN 'anna' ELSE 'bob' END")
	ctx := context.Background()
	done := []string{"done"}

	ids, err := w.GetSubtaskIDs(ctx, filterTaskTable, "t3", done, model.SQLFilter{})
	slices.Sort(ids)
	if err != nil || !slices.Equal(ids, []string{"t4", "t6"}) {
		t.Errorf("open subtasks of t3 = %v, %v; want [t4 t6], not the done t5 or the deleted t7", ids, err)
	}

	ids, err = w.GetSubtaskIDs(ctx, filterTaskTable, "t3", nil, model.SQLFilter{})
	slices.Sort(ids)
	if err != nil || !slices.Equal(ids, []string{"t4", "t5", "t6"}) {
		t.Errorf("every live subtask of t3 = %v, %v; want [t4 t5 t6]", ids, err)
	}

	anna := model.SQLFilter{SQL: "`assignee` = ?", Args: []any{"anna"}}
	if ids, err := w.GetSubtaskIDs(ctx, filterTaskTable, "t3", done, anna); err != nil || !slices.Equal(ids, []string{"t6"}) {
		t.Errorf("anna's open subtasks of t3 = %v, %v; want [t6]", ids, err)
	}

	mustExec(t, "UPDATE `"+filterTaskTable+"` SET parent_task_id = 't6' WHERE id = 't3'")
	ids, err = w.GetSubtaskIDs(ctx, filterTaskTable, "t3", done, model.SQLFilter{})
	slices.Sort(ids)
	if err != nil || !slices.Equal(ids, []string{"t4", "t6"}) {
		t.Errorf("open subtasks in a parent cycle = %v, %v; want [t4 t6] without t3 itself", ids, err)
	}
}

func TestPositionsFollowTheSortOrder(t *testing.T) {
	w := setupFilterTables(t)
	mustExec(t, "UPDATE `"+filterTaskTable+"` SET start_date = CASE id WHEN 't1' THEN 300 WHEN 't2' THEN NULL WHEN 't4' THEN 300 ELSE 50 END")
	open := model.SQLFilter{SQL: "(main.status IS NULL OR main.status NOT IN (?))", Args: []any{"done"}}
	const byStartDesc = "ORDER BY main.start_date DESC, main.id ASC"

	ctx := context.Background()
	none := model.SQLFilter{}

	for _, c := range []struct {
		filter model.SQLFilter
		want   []string
	}{
		{open, []string{"t1", "t2"}},
		{none, []string{"t1", "t3", "t2"}},
	} {
		rows, total, _, err := w.GetRootTasksPagedWithSubtasks(ctx, filterTaskTable, none, c.filter, byStartDesc, 0, 0, true)
		if err != nil {
			t.Fatalf("paged: %v", err)
		}
		var order []string
		for _, r := range rows[:total] {
			order = append(order, r["id"].(string))
		}
		if !slices.Equal(order, c.want) {
			t.Fatalf("branch order under %q = %v, want %v", c.filter.SQL, order, c.want)
		}
		for i, id := range order {
			if pos, err := w.GetBranchPosition(ctx, filterTaskTable, id, none, c.filter, byStartDesc); err != nil || pos != i+1 {
				t.Errorf("position of %s under %q = %d, %v; want %d", id, c.filter.SQL, pos, err, i+1)
			}
		}
		if pos, err := w.GetBranchPosition(ctx, filterTaskTable, "missing", none, c.filter, byStartDesc); err != nil || pos != 0 {
			t.Errorf("position of a missing task under %q = %d, %v; want 0", c.filter.SQL, pos, err)
		}
	}

	if pos, err := w.GetBranchPosition(ctx, filterTaskTable, "t4", none, none, byStartDesc); err != nil || pos != 2 {
		t.Errorf("position of the subtask t4 = %d, %v; want its parent's 2", pos, err)
	}
}

func TestAssignedOnlyFilterInEveryTaskQuery(t *testing.T) {
	w := setupFilterTables(t)
	mustExec(t, "UPDATE `"+filterTaskTable+"` SET assignee = CASE id WHEN 't1' THEN 'anna' WHEN 't4' THEN 'anna' ELSE 'bob' END")
	anna := model.SQLFilter{SQL: "`assignee` = ?", Args: []any{"anna"}}
	const byID = "ORDER BY main.id ASC"
	ctx := context.Background()

	rows, total, _, err := w.GetRootTasksPagedWithSubtasks(ctx, filterTaskTable, anna, model.SQLFilter{}, byID, 10, 0, true)
	if got := rowIDs(rows); err != nil || total != 2 || !slices.Equal(got, []string{"t1", "t4"}) {
		t.Errorf("grid page = %v total %d, %v; want [t1 t4] total 2, t4 a root under bob's t3", got, total, err)
	}
	if pos, err := w.GetBranchPosition(ctx, filterTaskTable, "t4", anna, model.SQLFilter{}, byID); err != nil || pos != 2 {
		t.Errorf("position of t4 = %d, %v; want 2", pos, err)
	}
	if pos, err := w.GetBranchPosition(ctx, filterTaskTable, "t1", anna, model.SQLFilter{}, byID); err != nil || pos != 1 {
		t.Errorf("position of t1 = %d, %v; want 1", pos, err)
	}

	kanban, err := w.GetTableRowsFiltered(ctx, filterTaskTable, anna)
	if got := rowIDs(kanban); err != nil || !slices.Equal(got, []string{"t1", "t4"}) {
		t.Errorf("kanban rows = %v, %v; want [t1 t4]", got, err)
	}

	gantt, total, err := w.GetTasksByDateRange(ctx, filterTaskTable, 0, 200, anna, 0)
	if got := rowIDs(gantt); err != nil || total != 2 || !slices.Equal(got, []string{"t1", "t4"}) {
		t.Errorf("gantt rows = %v of %d, %v; want [t1 t4] of 2", got, total, err)
	}

	first, total, err := w.GetTasksByDateRange(ctx, filterTaskTable, 0, 200, anna, 1)
	if err != nil || total != 2 || len(first) != 1 {
		t.Errorf("limited gantt rows = %v of %d, %v; want one row of 2", first, total, err)
	}

	calendar, err := w.GetTasksByCalendarRange(ctx, filterTaskTable, 0, 200, anna)
	if got := rowIDs(calendar); err != nil || !slices.Equal(got, []string{"t1", "t4"}) {
		t.Errorf("calendar rows = %v, %v; want [t1 t4]", got, err)
	}
}

func TestFilteredPagingBindsFilterInEverySubquery(t *testing.T) {
	w := setupFilterTables(t)
	notPlain := w.BuildFilterSQLInline(flat(model.Filter{Field: "name", Operator: "is_not", Value: "plain"}), filterHeaders, nil, filterTableNames)
	const byID = "ORDER BY main.id ASC"

	ctx := context.Background()
	none := model.SQLFilter{}

	rows, total, _, err := w.GetRootTasksPagedWithSubtasks(ctx, filterTaskTable, none, notPlain, byID, 10, 0, true)
	if err != nil {
		t.Fatalf("paged: %v", err)
	}
	if got := rowIDs(rows); total != 2 || !slices.Equal(got, []string{"t1", "t2"}) {
		t.Errorf("paged rows = %v total %d, want [t1 t2] total 2", got, total)
	}

	uncounted, total, _, err := w.GetRootTasksPagedWithSubtasks(ctx, filterTaskTable, none, notPlain, byID, 10, 0, false)
	if got := rowIDs(uncounted); err != nil || total != 0 || !slices.Equal(got, []string{"t1", "t2"}) {
		t.Errorf("uncounted page = %v total %d, %v; want the same rows and no count", got, total, err)
	}
	if count, err := w.CountBranches(ctx, filterTaskTable, none, notPlain); err != nil || count != 2 {
		t.Errorf("CountBranches = %d, %v; want 2", count, err)
	}

	plain := w.BuildFilterSQLInline(flat(is("name", "plain")), filterHeaders, nil, filterTableNames)
	rows, total, _, err = w.GetRootTasksPagedWithSubtasks(ctx, filterTaskTable, none, plain, byID, 10, 0, true)
	if err != nil {
		t.Fatalf("paged with subtasks: %v", err)
	}
	if got := rowIDs(rows); total != 1 || !slices.Equal(got, []string{"t3", "t4"}) {
		t.Errorf("paged rows = %v total %d, want [t3 t4] total 1", got, total)
	}

	if pos, err := w.GetBranchPosition(ctx, filterTaskTable, "t4", none, plain, byID); err != nil || pos != 1 {
		t.Errorf("position of t4 under its matching parent = %d, %v; want 1", pos, err)
	}
	if pos, err := w.GetBranchPosition(ctx, filterTaskTable, "t2", none, notPlain, byID); err != nil || pos != 2 {
		t.Errorf("position of t2 = %d, %v; want 2", pos, err)
	}
}

func TestATopLevelTaskIsStoredWithoutAParent(t *testing.T) {
	w := setupFilterTables(t)
	tx, err := w.Db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	for _, task := range []struct{ id, parent string }{{"new-top", ""}, {"new-self", "new-self"}, {"new-sub", "t1"}} {
		if err := w.InsertTaskTx(tx, filterTaskTable, task.id, task.id, "u1", task.parent, 100); err != nil {
			t.Fatal(err)
		}
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}

	rows, err := w.GetTableRowsFiltered(context.Background(), filterTaskTable, model.SQLFilter{SQL: "main.id LIKE 'new-%' AND main.parent_task_id IS NULL"})
	if got := rowIDs(rows); err != nil || !slices.Equal(got, []string{"new-self", "new-top"}) {
		t.Errorf("new tasks without a parent = %v, %v; want new-self and new-top", got, err)
	}
}

func TestDeleteTaskTxDeletesTheTasksTogether(t *testing.T) {
	w := setupFilterTables(t)

	if _, err := w.DeleteTaskTx(filterTaskTable, false, []string{"t3", "t4"}, nil); err != nil {
		t.Fatal(err)
	}

	live, err := w.GetTableRowsFiltered(context.Background(), filterTaskTable, model.SQLFilter{})
	if got := rowIDs(live); err != nil || !slices.Equal(got, []string{"t1", "t2"}) {
		t.Errorf("live tasks = %v, %v; want t3 and its subtask t4 deleted", got, err)
	}
}

func TestTaskAssigneesAndAPickerLimitedToOwnTasks(t *testing.T) {
	w := setupFilterTables(t)
	mustExec(t, "UPDATE `"+filterTaskTable+"` SET assignee = CASE id WHEN 't1' THEN 'anna' WHEN 't2' THEN 'bob' ELSE NULL END")
	ctx := context.Background()

	owners, err := w.GetTaskAssignees(ctx, filterTaskTable, []string{"t1", "t2", "t3", "missing"})
	if err != nil || owners["t1"] != "anna" || owners["t2"] != "bob" || owners["t3"] != "" || len(owners) != 3 {
		t.Errorf("assignees = %v, %v; want t1 anna, t2 bob, t3 no one", owners, err)
	}

	anna := model.SQLFilter{SQL: "`assignee` = ?", Args: []any{"anna"}}
	rows, err := w.GetLinkedRecordsLite(ctx, filterTaskTable, "", nil, 10, 0, anna)
	if got := rowIDs(rows); err != nil || !slices.Equal(got, []string{"t1"}) {
		t.Errorf("anna's picker = %v, %v; want only her t1", got, err)
	}
}
