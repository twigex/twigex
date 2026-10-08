// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"context"
	"database/sql"
	"encoding/json"
	"slices"
	"strings"
	"testing"

	"github.com/twigex/twigex/model"
)

func kanbanOrder(t *testing.T, raw string) kanbanViewOrder {
	t.Helper()
	var order kanbanViewOrder
	if err := json.Unmarshal([]byte(raw), &order); err != nil {
		t.Fatal(err)
	}

	return order
}

func TestKanbanColumnFiltersPutUnknownValuesInUnassigned(t *testing.T) {
	got := kanbanColumnFilters(kanbanOrder(t, `{"section":"status","fields":[{"id":"0"},{"id":"s1"},{"id":"s2"}]}`))
	want := []model.SQLFilter{
		{SQL: "(main.`status` IS NULL OR main.`status` = '' OR main.`status` NOT IN (?,?))", Args: []any{"s1", "s2"}},
		{SQL: "main.`status` = ?", Args: []any{"s1"}},
		{SQL: "main.`status` = ?", Args: []any{"s2"}},
	}
	for i := range want {
		if got[i].SQL != want[i].SQL || !slices.Equal(got[i].Args, want[i].Args) {
			t.Errorf("column %d = %+v, want %+v", i, got[i], want[i])
		}
	}

	only := kanbanColumnFilters(kanbanOrder(t, `{"section":"status","fields":[{"id":"0"}]}`))
	if only[0].SQL != "1 = 1" {
		t.Errorf("a board with only Unassigned = %+v, want every card", only[0])
	}

	named := kanbanColumnFilters(kanbanOrder(t, `{"section":"status","fields":[{"id":"u1","name":"Unassigned"},{"id":"s1"}]}`))
	if !strings.Contains(named[0].SQL, "NOT IN (?)") || !slices.Equal(named[0].Args, []any{"s1"}) {
		t.Errorf("a column named Unassigned = %+v, want the cards of no other column", named[0])
	}
}

type fakeKanbanPageStore struct {
	fakeReadAccessStore
	viewTable   string
	counts      int
	pages       []string
	lastPage    model.SQLFilter
	rows        []map[string]interface{}
	subtasksFor string
}

func (f *fakeKanbanPageStore) GetView(_ context.Context, viewID string) (*model.WorkspaceView, error) {
	return &model.WorkspaceView{ID: viewID, TableID: f.viewTable,
		TaskOrder: `{"section":"status","fields":[{"id":"0"},{"id":"s1","order":[{"id":"late"},{"id":"early"}]}]}`}, nil
}

func (f *fakeKanbanPageStore) TableHasColumn(_, column string) (bool, error) {
	return column == "status", nil
}

func (f *fakeKanbanPageStore) GetTableColumnTypes(string) ([]*sql.ColumnType, error) { return nil, nil }

// CountTasksAcrossTables gives s1 the count in counts and every other column
// none.
func (f *fakeKanbanPageStore) CountTasksAcrossTables(_ context.Context, _ []model.TaskSource, groups []model.SQLFilter) ([]int, error) {
	counts := make([]int, len(groups))
	for i, g := range groups {
		if g.SQL == "main.`status` = ?" {
			counts[i] = f.counts
		}
	}
	return counts, nil
}

func (f *fakeKanbanPageStore) GetTableRowsPage(_ context.Context, _ string, filter model.SQLFilter, firstIDs []string, limit int) ([]map[string]interface{}, error) {
	f.pages = append(f.pages, strings.Join(firstIDs, ","))
	f.lastPage = filter
	if f.rows != nil {
		return f.rows[:min(limit, len(f.rows))], nil
	}
	return []map[string]interface{}{{"id": "card-" + filter.Args[len(filter.Args)-1].(string), "description": "long"}}, nil
}

func (f *fakeKanbanPageStore) GetTableRowsFiltered(_ context.Context, _ string, filter model.SQLFilter) ([]map[string]interface{}, error) {
	f.subtasksFor = strings.TrimSpace(filter.SQL)
	return []map[string]interface{}{{"id": "sub-1"}}, nil
}

func kanbanPageApp(ws *fakeKanbanPageStore) *App {
	ws.tableOwner = map[string]string{"t-a": "ws-a"}
	ws.fakeAccessWorkspaceStore = workspaceAdminAccess()
	a := readAccessApp(&ws.fakeReadAccessStore)
	a.Store.Workspace = ws
	return a
}

func TestGetKanbanDataReturnsAPagePerColumnWithItsSubtasks(t *testing.T) {
	admin := workspaceAdmin
	ws := &fakeKanbanPageStore{viewTable: "t-a", counts: 120}
	a := kanbanPageApp(ws)

	page, appErr := a.GetKanbanData(context.Background(), "ws-a", "t-a", "v1", "", "", 50, model.FilterPayload{}, admin)
	if appErr != nil {
		t.Fatal(appErr)
	}
	if len(page.Columns) != 2 || page.Columns[0].Total != 0 || len(page.Columns[0].Tasks) != 0 || page.Columns[1].Total != 120 {
		t.Fatalf("columns = %+v, want an empty Unassigned and s1 with 120 cards", page.Columns)
	}
	if !slices.Equal(ws.pages, []string{"late,early"}) {
		t.Errorf("pages read = %v, want only s1's, in its saved order", ws.pages)
	}
	if card := page.Columns[1].Tasks[0]; card["id"] != "card-s1" || card["description"] != nil {
		t.Errorf("card = %v, want s1's card without its description", card)
	}
	if len(page.Subtasks) != 1 || !strings.Contains(ws.subtasksFor, "main.parent_task_id IN (?)") {
		t.Errorf("subtasks = %v read with %q, want those of the card on the page", page.Subtasks, ws.subtasksFor)
	}

	if page.Columns[1].Next != "" {
		t.Errorf("s1 has a next page, want none past its one card")
	}
}

func TestGetKanbanDataLoadsMoreAfterTheLastCardShown(t *testing.T) {
	admin := workspaceAdmin
	ws := &fakeKanbanPageStore{viewTable: "t-a", counts: 5}
	a := kanbanPageApp(ws)
	load := func(after string, rows ...map[string]interface{}) model.KanbanColumnPage {
		t.Helper()
		ws.rows = rows
		page, appErr := a.GetKanbanData(context.Background(), "ws-a", "t-a", "v1", "s1", after, 2, model.FilterPayload{}, admin)
		if appErr != nil {
			t.Fatal(appErr)
		}
		if len(page.Columns) != 1 || page.Columns[0].ID != "s1" || page.Columns[0].Total != 5 {
			t.Fatalf("columns = %+v, want s1 alone with its count", page.Columns)
		}
		return page.Columns[0]
	}
	card := func(id string, created any) map[string]interface{} {
		return map[string]interface{}{"id": id, "created_at": created}
	}

	first := load("", card("late", 9), card("early", 1), card("a", 3))
	if len(first.Tasks) != 2 || first.Next == "" || ws.pages[0] != "late,early" {
		t.Fatalf("first page = %+v read with %v, want the saved cards and more to come", first, ws.pages)
	}

	second := load(first.Next, card("a", 3), card("b", "4"), card("c", 4))
	if ws.pages[1] != "" || !strings.Contains(ws.lastPage.SQL, "main.id NOT IN (?,?)") {
		t.Errorf("second page read with first %q and %q, want past the saved order, leaving it out", ws.pages[1], ws.lastPage.SQL)
	}
	if len(second.Tasks) != 2 || second.Next == "" {
		t.Fatalf("second page = %+v, want two cards and more to come", second)
	}

	third := load(second.Next, card("c", 4))
	created := int64(4)
	after := model.RowsAfter("main.created_at", &created, "b")
	if !strings.HasSuffix(ws.lastPage.SQL, after.SQL) || !slices.Equal(ws.lastPage.Args[len(ws.lastPage.Args)-3:], after.Args) {
		t.Errorf("third page read with %+v, want it to start after b", ws.lastPage)
	}
	if len(third.Tasks) != 1 || third.Next != "" {
		t.Errorf("third page = %+v, want the last card and no page after it", third)
	}

	moved := taskCursor{ID: "dragged-away", Saved: new(int)}
	load(moved.encode(), card("early", 1))
	if ws.pages[3] != "early" {
		t.Errorf("after a card dragged out of the saved order, first = %q, want the saved cards after its old place", ws.pages[3])
	}
}

func TestGetKanbanDataRefusesAViewOfAnotherTable(t *testing.T) {
	admin := workspaceAdmin
	a := kanbanPageApp(&fakeKanbanPageStore{viewTable: "t-other"})
	if _, appErr := a.GetKanbanData(context.Background(), "ws-a", "t-a", "v1", "", "", 50, model.FilterPayload{}, admin); appErr == nil || appErr.Status != 404 {
		t.Errorf("error = %v, want not found", appErr)
	}
}
