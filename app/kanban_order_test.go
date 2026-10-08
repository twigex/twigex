// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"testing"

	"github.com/twigex/twigex/model"
)

const kanbanBoard = `{"section":"status","fieldsVisible":["name"],"fields":[
	{"id":"todo","name":"To do","width":"300","color":"#fff","order":[{"id":"a","status":"todo"},{"id":"b","status":"todo"},{"id":"c","status":"todo"}]},
	{"id":"done","name":"Done","width":"250","order":[{"id":"d","status":"done"}]}
]}`

func placedIn(t *testing.T, orderJSON, column string) []string {
	t.Helper()
	var order map[string]any
	if err := json.Unmarshal([]byte(orderJSON), &order); err != nil {
		t.Fatal(err)
	}

	for _, c := range kanbanColumns(order) {
		if kanbanOrderID(c["id"]) == column {
			return placedIDs(c)
		}
	}
	t.Fatalf("no column %s in %s", column, orderJSON)
	return nil
}

func noRest(t *testing.T) func([]string) ([]string, error) {
	return func([]string) ([]string, error) {
		t.Helper()
		t.Fatal("the rest of the column was read for a card that has a place")
		return nil, nil
	}
}

func TestMoveKanbanCardPlacesOneCardAndKeepsTheRest(t *testing.T) {
	for _, tc := range []struct {
		name, task, column, after string
		todo, done                []string
	}{
		{"down the column", "a", "todo", "b", []string{"b", "a", "c"}, []string{"d"}},
		{"to the top", "c", "todo", "", []string{"c", "a", "b"}, []string{"d"}},
		{"to another column", "b", "done", "d", []string{"a", "c"}, []string{"d", "b"}},
		{"to the top of another column", "d", "todo", "", []string{"d", "a", "b", "c"}, []string{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			out, err := moveKanbanCard(kanbanBoard, tc.task, tc.column, tc.after, noRest(t))
			if err != nil {
				t.Fatal(err)
			}

			if got := placedIn(t, out, "todo"); !slices.Equal(got, tc.todo) {
				t.Errorf("to do = %v, want %v", got, tc.todo)
			}
			if got := placedIn(t, out, "done"); !slices.Equal(got, tc.done) {
				t.Errorf("done = %v, want %v", got, tc.done)
			}
		})
	}

	out, _ := moveKanbanCard(kanbanBoard, "a", "done", "", noRest(t))
	var order map[string]any
	_ = json.Unmarshal([]byte(out), &order)
	done := kanbanColumns(order)[1]
	if done["name"] != "Done" || done["width"] != "250" || order["section"] != "status" || order["fieldsVisible"] == nil {
		t.Errorf("the move lost the columns' settings: %s", out)
	}
	if card := placedCards(done)[0].(map[string]any); card["status"] != "done" {
		t.Errorf("moved card = %v, want its column under the grouped field", card)
	}
}

func TestMoveKanbanCardPlacesTheCardsAboveOneWithNoPlace(t *testing.T) {
	var asked []string
	rest := func(placed []string) ([]string, error) {
		asked = placed
		return []string{"x", "a2", "y"}, nil
	}

	out, err := moveKanbanCard(kanbanBoard, "a", "todo", "y", rest)
	if err != nil {
		t.Fatal(err)
	}

	if !slices.Equal(asked, []string{"b", "c"}) {
		t.Errorf("rest was asked to leave out %v, want the placed cards b, c", asked)
	}
	if got, want := placedIn(t, out, "todo"), []string{"b", "c", "x", "a2", "y", "a"}; !slices.Equal(got, want) {
		t.Errorf("to do = %v, want %v", got, want)
	}
}

func TestMoveKanbanCardAfterACardNotInTheColumnPlacesOnlyTheCard(t *testing.T) {
	rest := func([]string) ([]string, error) {
		return []string{"x", "y"}, nil
	}

	out, err := moveKanbanCard(kanbanBoard, "a", "todo", "gone", rest)
	if err != nil {
		t.Fatal(err)
	}

	if got, want := placedIn(t, out, "todo"), []string{"b", "c", "a"}; !slices.Equal(got, want) {
		t.Errorf("to do = %v, want %v", got, want)
	}
}

func TestMoveKanbanCardPastTheCapKeepsTheMovedCard(t *testing.T) {
	cards := make([]any, kanbanSavedOrderLimit)
	for i := range cards {
		cards[i] = map[string]any{"id": fmt.Sprintf("t%d", i)}
	}
	board, _ := json.Marshal(map[string]any{"section": "status", "fields": []any{map[string]any{"id": "todo", "order": cards}}})

	last := fmt.Sprintf("t%d", kanbanSavedOrderLimit-1)
	out, err := moveKanbanCard(string(board), "new", "todo", last, noRest(t))
	if err != nil {
		t.Fatal(err)
	}

	got := placedIn(t, out, "todo")
	if len(got) != kanbanSavedOrderLimit || got[len(got)-1] != "new" {
		t.Errorf("placed %d cards ending %v, want %d ending with the moved one", len(got), got[len(got)-1:], kanbanSavedOrderLimit)
	}
}

func TestMoveKanbanCardRefusesAColumnTheViewHasNot(t *testing.T) {
	if _, err := moveKanbanCard(kanbanBoard, "a", "gone", "", noRest(t)); !errors.Is(err, errUnknownKanbanColumn) {
		t.Errorf("error = %v, want an unknown column", err)
	}
}

func TestMoveKanbanCardKeepsTheCapOnPlacedCards(t *testing.T) {
	cards := make([]any, kanbanSavedOrderLimit)
	for i := range cards {
		cards[i] = map[string]any{"id": fmt.Sprintf("t%d", i)}
	}
	board, _ := json.Marshal(map[string]any{"section": "status", "fields": []any{map[string]any{"id": "todo", "order": cards}}})

	out, err := moveKanbanCard(string(board), "new", "todo", "", noRest(t))
	if err != nil {
		t.Fatal(err)
	}

	got := placedIn(t, out, "todo")
	if len(got) != kanbanSavedOrderLimit || got[0] != "new" {
		t.Errorf("placed %d cards starting %v, want %d starting with the moved one", len(got), got[:1], kanbanSavedOrderLimit)
	}
}

func TestMergeKanbanColumnsKeepsWhereCardsSit(t *testing.T) {
	incoming := `{"section":"status","fieldsVisible":["name","due_date"],"fields":[
		{"id":"done","name":"Finished","width":"400","order":[{"id":"z"}]},
		{"id":"todo","name":"To do","width":"300"},
		{"id":"new","name":"New column"}
	]}`
	out, err := mergeKanbanColumns(kanbanBoard, incoming)
	if err != nil {
		t.Fatal(err)
	}

	var order map[string]any
	_ = json.Unmarshal([]byte(out), &order)
	columns := kanbanColumns(order)
	if len(columns) != 3 || columns[0]["name"] != "Finished" || columns[0]["width"] != "400" {
		t.Fatalf("columns = %v, want the client's names, widths and order", columns)
	}
	if got := placedIn(t, out, "done"); !slices.Equal(got, []string{"d"}) {
		t.Errorf("done = %v, want the stored d, not the client's z", got)
	}
	if got := placedIn(t, out, "todo"); !slices.Equal(got, []string{"a", "b", "c"}) {
		t.Errorf("to do = %v, want the stored order", got)
	}
	if got := placedIn(t, out, "new"); len(got) != 0 {
		t.Errorf("new column = %v, want empty", got)
	}
	if visible, _ := order["fieldsVisible"].([]any); len(visible) != 2 {
		t.Errorf("fields shown on cards = %v, want the client's", order["fieldsVisible"])
	}
}

func TestWithoutPlacedCardsKeepsTheColumns(t *testing.T) {
	out := withoutPlacedCards(kanbanBoard)
	var order map[string]any
	if err := json.Unmarshal([]byte(out), &order); err != nil {
		t.Fatal(err)
	}

	for _, c := range kanbanColumns(order) {
		if _, has := c["order"]; has {
			t.Errorf("column %v still lists its cards", c["id"])
		}
	}
	if columns := kanbanColumns(order); len(columns) != 2 || columns[0]["width"] != "300" || order["section"] != "status" {
		t.Errorf("stripped order = %s, want the columns kept", out)
	}
	if got := withoutPlacedCards("not json"); got != "not json" {
		t.Errorf("unreadable order = %q, want it as it was", got)
	}
}

type fakeKanbanOrderStore struct {
	fakeReadAccessStore
	view      model.WorkspaceView
	restAsked model.SQLFilter
	rest      []string
	renamed   string
	replaced  string
}

func (f *fakeKanbanOrderStore) GetView(_ context.Context, id string) (*model.WorkspaceView, error) {
	if id != f.view.ID {
		return nil, sql.ErrNoRows
	}
	v := f.view
	return &v, nil
}

func (f *fakeKanbanOrderStore) GetTableName(string) (string, error) { return "pre_tasks", nil }

func (f *fakeKanbanOrderStore) TableHasColumn(_, column string) (bool, error) {
	return column == "status", nil
}

func (f *fakeKanbanOrderStore) GetViewVisibility(context.Context, string) (bool, string, error) {
	return true, "", nil
}

func (f *fakeKanbanOrderStore) ChangeViewOrder(_ context.Context, _, _, _ string, change func(string) (string, error)) (string, error) {
	order, err := change(f.view.TaskOrder)
	if err == nil {
		f.view.TaskOrder = order
	}
	return order, err
}

func (f *fakeKanbanOrderStore) GetKanbanRestIDs(_ context.Context, _ string, filter model.SQLFilter, _ string, _ int) ([]string, error) {
	f.restAsked = filter
	return f.rest, nil
}

func (f *fakeKanbanOrderStore) UpdateViewName(_, _, _, name string) error {
	f.renamed = name
	return nil
}

func (f *fakeKanbanOrderStore) UpdateView(_, _, _, order, _, _ string) (*model.WorkspaceView, error) {
	f.replaced = order
	return &model.WorkspaceView{}, nil
}

func kanbanOrderApp(viewType, order string) (*App, *fakeKanbanOrderStore) {
	ws := &fakeKanbanOrderStore{view: model.WorkspaceView{ID: "v1", WorkspaceID: "ws", TableID: "t1", ViewType: viewType, Name: "Board", TaskOrder: order}}
	ws.tableOwner = map[string]string{"t1": "ws"}
	ws.fakeAccessWorkspaceStore = workspaceAdminAccess()
	a := readAccessApp(&ws.fakeReadAccessStore)
	a.Store.Workspace = ws
	return a, ws
}

func TestMoveKanbanCardRefusesSomeoneOutsideTheWorkspace(t *testing.T) {
	stranger := model.User{ID: "stranger"}
	a, ws := kanbanOrderApp("kanban", kanbanBoard)

	appErr := a.MoveKanbanCard(context.Background(), "ws", "t1", "v1", "a", "done", "", stranger)
	if appErr == nil || appErr.Status != http.StatusForbidden {
		t.Fatalf("error = %v, want forbidden", appErr)
	}

	if ws.replaced != "" {
		t.Errorf("the order was saved: %s", ws.replaced)
	}
}

func TestMoveKanbanCardAsksForTheColumnsUnplacedCards(t *testing.T) {
	admin := workspaceAdmin
	a, ws := kanbanOrderApp("kanban", kanbanBoard)
	ws.rest = []string{"e", "f"}

	if appErr := a.MoveKanbanCard(context.Background(), "ws", "t1", "v1", "a", "done", "f", admin); appErr != nil {
		t.Fatal(appErr)
	}
	if !strings.Contains(ws.restAsked.SQL, kanbanCardsSQL) || !strings.Contains(ws.restAsked.SQL, "NOT IN (?)") {
		t.Errorf("unplaced cards read with %q, want the column's top-level cards without those placed", ws.restAsked.SQL)
	}
	if got, want := placedIn(t, ws.view.TaskOrder, "done"), []string{"d", "e", "f", "a"}; !slices.Equal(got, want) {
		t.Errorf("done = %v, want %v", got, want)
	}

	for _, tc := range []struct {
		name, view, task, column, after string
		want                            int
	}{
		{"a column the view has not", "v1", "a", "gone", "", http.StatusBadRequest},
		{"a card after itself", "v1", "a", "todo", "a", http.StatusBadRequest},
		{"an id that is not an id", "v1", "a'; --", "todo", "", http.StatusBadRequest},
		{"a view that is not there", "v2", "a", "todo", "", http.StatusNotFound},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a, _ := kanbanOrderApp("kanban", kanbanBoard)
			appErr := a.MoveKanbanCard(context.Background(), "ws", "t1", tc.view, tc.task, tc.column, tc.after, admin)
			if appErr == nil || appErr.Status != tc.want {
				t.Errorf("error = %v, want status %d", appErr, tc.want)
			}
		})
	}

	a, _ = kanbanOrderApp("grid", kanbanBoard)
	if appErr := a.MoveKanbanCard(context.Background(), "ws", "t1", "v1", "a", "todo", "", admin); appErr == nil || appErr.Status != http.StatusNotFound {
		t.Errorf("a grid view: %v, want not found", appErr)
	}
}

func TestUpdatingAViewKeepsWhatItDoesNotSend(t *testing.T) {
	admin := workspaceAdmin

	a, ws := kanbanOrderApp("kanban", kanbanBoard)
	view, appErr := a.UpdateView(context.Background(), "v1", "ws", "t1", "", "Renamed", admin)
	if appErr != nil {
		t.Fatal(appErr)
	}
	if ws.renamed != "Renamed" || ws.view.TaskOrder != kanbanBoard || ws.replaced != "" {
		t.Errorf("a rename changed the order: %q", ws.view.TaskOrder)
	}
	if view.Name != "Renamed" {
		t.Errorf("returned view = %+v, want the new name", view)
	}

	a, ws = kanbanOrderApp("kanban", kanbanBoard)
	incoming := `{"section":"status","fields":[{"id":"done","name":"Done","order":[]},{"id":"todo","name":"To do","width":"500"}]}`
	view, appErr = a.UpdateView(context.Background(), "v1", "ws", "t1", incoming, "Board", admin)
	if appErr != nil {
		t.Fatal(appErr)
	}
	if got := placedIn(t, ws.view.TaskOrder, "todo"); !slices.Equal(got, []string{"a", "b", "c"}) {
		t.Errorf("to do after a column save = %v, want its cards kept", got)
	}
	if strings.Contains(view.TaskOrder, `"order"`) {
		t.Errorf("returned order = %s, want the columns without the cards", view.TaskOrder)
	}

	a, ws = kanbanOrderApp("form", `{"order":[]}`)
	if _, appErr := a.UpdateView(context.Background(), "v1", "ws", "t1", `{"order":["name"]}`, "Form", admin); appErr != nil {
		t.Fatal(appErr)
	}
	if ws.replaced != `{"order":["name"]}` {
		t.Errorf("a form view was not saved as sent: %q", ws.replaced)
	}
}

func TestMoveKanbanCardRefusesABoardGroupedByAColumnTheTableLacks(t *testing.T) {
	admin := workspaceAdmin
	board := strings.Replace(kanbanBoard, `"section":"status"`, `"section":"status`+"`"+`, (SELECT 1) --"`, 1)
	a, ws := kanbanOrderApp("kanban", board)
	ws.rest = []string{"e"}

	if appErr := a.MoveKanbanCard(context.Background(), "ws", "t1", "v1", "a", "done", "", admin); appErr == nil || appErr.Status != http.StatusBadRequest {
		t.Fatalf("error = %v, want the move refused", appErr)
	}
	if ws.restAsked.SQL != "" {
		t.Errorf("cards were read with %q for a column the table does not have", ws.restAsked.SQL)
	}
}
