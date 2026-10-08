// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"
	"time"

	"github.com/twigex/twigex/interfaces"
	"github.com/twigex/twigex/model"
)

// assignedOnlyRoles stands in for the enterprise roles: the listed tables are
// assigned-only for everyone it resolves.
type assignedOnlyRoles struct{ tables map[string]bool }

func (r assignedOnlyRoles) Resolve(interfaces.AccessInput) interfaces.WorkspaceAccess {
	return assignedOnlyAccess{tables: r.tables}
}

type assignedOnlyAccess struct {
	openAccess
	tables map[string]bool
}

func (a assignedOnlyAccess) AssignedOnly(tableID string) bool { return a.tables[tableID] }

type fakeAssignedOnlyStore struct {
	fakeReadAccessStore
	gotFilters map[string]model.SQLFilter
	gridRows   []map[string]interface{}
	inline     model.SQLFilter
	viewTable  string
	deleted    []string
}

// kanbanCards is the condition the board adds so only top-level tasks become
// cards.
const kanbanCards = model.TopLevelTaskSQL

func (f *fakeAssignedOnlyStore) GetView(_ context.Context, viewID string) (*model.WorkspaceView, error) {
	return &model.WorkspaceView{ID: viewID, TableID: f.viewTable, TaskOrder: `{"section":"status","fields":[{"id":"0"},{"id":"s1"}]}`}, nil
}

func (f *fakeAssignedOnlyStore) TableHasColumn(_, column string) (bool, error) {
	return column == "status", nil
}

func (f *fakeAssignedOnlyStore) CountTasksAcrossTables(_ context.Context, sources []model.TaskSource, groups []model.SQLFilter) ([]int, error) {
	f.gotFilters["kanban"] = sources[0].Filter
	return make([]int, len(groups)), nil
}

func (f *fakeAssignedOnlyStore) GetTasksByDateRange(_ context.Context, _ string, _, _ int64, filter model.SQLFilter, _ int) ([]map[string]interface{}, int, error) {
	f.gotFilters["gantt"] = filter
	return nil, 0, nil
}

func (f *fakeAssignedOnlyStore) GetTasksByCalendarRange(_ context.Context, _ string, _, _ int64, filter model.SQLFilter) ([]map[string]interface{}, error) {
	f.gotFilters["calendar"] = filter
	return nil, nil
}

func (f *fakeAssignedOnlyStore) GetRootTasksPagedWithSubtasks(_ context.Context, _ string, access, filter model.SQLFilter, _ string, _, _ int, _ bool) ([]map[string]interface{}, int, int, error) {
	f.gotFilters["grid"] = access.And(filter)
	return f.gridRows, len(f.gridRows), len(f.gridRows), nil
}

func (f *fakeAssignedOnlyStore) GetTableColumnTypes(string) ([]*sql.ColumnType, error) {
	return nil, nil
}

func (f *fakeAssignedOnlyStore) GetTableNamesByIDs([]string) (map[string]string, error) {
	return map[string]string{}, nil
}

func (f *fakeAssignedOnlyStore) BuildFilterSQLInline(model.FilterPayload, []model.WorkspaceHeaders, *time.Location, map[string]string) model.SQLFilter {
	return f.inline
}

func assignedOnlyApp(tables map[string]bool) (*App, *fakeAssignedOnlyStore) {
	ws := &fakeAssignedOnlyStore{
		fakeReadAccessStore: fakeReadAccessStore{
			fakeAccessWorkspaceStore: fakeAccessWorkspaceStore{
				memberRoleName: "member",
				roles: []model.ProjectWorkspaceRole{{Name: "member", Permissions: []string{
					model.PermissionUpdateTask, model.PermissionViewGantt, model.PermissionViewCalendar,
				}}},
			},
			tableOwner: map[string]string{"t-mine": "ws-a", "t-open": "ws-a"},
		},
		gotFilters: map[string]model.SQLFilter{},
		gridRows: []map[string]interface{}{
			{"id": "own", "assignee": "anna"},
			{"id": "subtask-of-own", "assignee": "bob"},
		},
	}
	a := readAccessApp(&ws.fakeReadAccessStore)
	a.Store.Workspace = ws
	a.WorkspaceRoles = assignedOnlyRoles{tables: tables}
	return a, ws
}

func loadAllViews(t *testing.T, a *App, tableID string, user model.User) []map[string]interface{} {
	t.Helper()
	ctx := context.Background()
	a.Store.Workspace.(*fakeAssignedOnlyStore).viewTable = tableID
	if _, appErr := a.GetKanbanData(ctx, "ws-a", tableID, "v-kanban", "", "", 50, model.FilterPayload{}, user); appErr != nil {
		t.Fatalf("kanban: %v", appErr)
	}
	if _, appErr := a.GetTasksByDateRange(ctx, "ws-a", tableID, 0, 1, model.FilterPayload{}, user); appErr != nil {
		t.Fatalf("gantt: %v", appErr)
	}
	if _, appErr := a.GetTasksByCalendarRange(ctx, "ws-a", tableID, 0, 1, model.FilterPayload{}, user); appErr != nil {
		t.Fatalf("calendar: %v", appErr)
	}
	grid, appErr := a.GetFilteredTableData(ctx, "ws-a", tableID, "", model.FilterPayload{Timezone: "UTC", Limit: 100}, user, true)
	if appErr != nil {
		t.Fatalf("grid: %v", appErr)
	}
	return grid.DataBase
}

func TestAssignedOnlyLimitsGridKanbanGanttAndCalendar(t *testing.T) {
	anna := model.User{ID: "anna", Role: model.SystemUserRoleId}

	a, ws := assignedOnlyApp(map[string]bool{"t-mine": true})
	grid := loadAllViews(t, a, "t-mine", anna)
	for view, filter := range ws.gotFilters {
		want := "`assignee` = ?"
		if view == "kanban" {
			want += " AND " + kanbanCards
		}
		if filter.SQL != want || len(filter.Args) != 1 || filter.Args[0] != "anna" {
			t.Errorf("%s filter = %+v, want only anna's tasks", view, filter)
		}
	}
	if len(ws.gotFilters) != 4 {
		t.Errorf("filters reached %d views, want 4: %v", len(ws.gotFilters), ws.gotFilters)
	}
	if len(grid) != 1 || grid[0]["id"] != "own" {
		t.Errorf("grid rows = %v, want only anna's task without bob's subtask", grid)
	}

	a, ws = assignedOnlyApp(map[string]bool{"t-mine": true})
	loadAllViews(t, a, "t-open", anna)
	for view, filter := range ws.gotFilters {
		want := ""
		if view == "kanban" {
			want = kanbanCards
		}
		if filter.SQL != want {
			t.Errorf("%s on a table without the rule got filter %+v", view, filter)
		}
	}
}

func TestKanbanAppliesTheChosenFilterAlongsideAssignedOnly(t *testing.T) {
	anna := model.User{ID: "anna", Role: model.SystemUserRoleId}
	filters := model.FilterPayload{Timezone: "UTC", FlatFilters: []model.Filter{{Field: "status", Operator: "is", Value: "s1"}}}

	for _, tc := range []struct {
		name         string
		assignedOnly map[string]bool
		wantSQL      string
		wantArgs     []any
	}{
		{"open table", nil, "(status = ?) AND " + kanbanCards, []any{"s1"}},
		{"assigned-only table", map[string]bool{"t-mine": true}, "(status = ?) AND `assignee` = ? AND " + kanbanCards, []any{"s1", "anna"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a, ws := assignedOnlyApp(tc.assignedOnly)
			ws.inline = model.SQLFilter{SQL: "(status = ?)", Args: []any{"s1"}}
			ws.viewTable = "t-mine"
			if _, appErr := a.GetKanbanData(context.Background(), "ws-a", "t-mine", "v-kanban", "", "", 50, filters, anna); appErr != nil {
				t.Fatalf("kanban: %v", appErr)
			}
			got := ws.gotFilters["kanban"]
			if got.SQL != tc.wantSQL || !slices.Equal(got.Args, tc.wantArgs) {
				t.Errorf("kanban filter = %+v, want %q %v", got, tc.wantSQL, tc.wantArgs)
			}
		})
	}
}

func (f *fakeAssignedOnlyStore) GetTaskAssigneeID(_, taskID string) (string, error) {
	owners := map[string]string{"own": "anna", "bobs": "bob"}
	if assignee, ok := owners[taskID]; ok {
		return assignee, nil
	}
	return "", sql.ErrNoRows
}

func TestAssignedOnlyRefusesAnotherUsersTaskByID(t *testing.T) {
	anna := model.User{ID: "anna", Role: model.SystemUserRoleId}
	a, ws := assignedOnlyApp(map[string]bool{"t-mine": true})
	ws.viewTable = "t-mine"
	ctx := context.Background()

	checks := map[string]*model.AppError{}
	_, _, checks["open the task"] = a.GetItemForTableByID(ctx, "ws-a", "t-mine", "bobs", anna)
	_, checks["read its comments"] = a.GetTaskComments(ctx, "ws-a", "t-mine", "bobs", anna)
	_, checks["comment on it"] = a.AddTaskComment(ctx, "ws-a", "t-mine", "bobs", "hi", anna, nil)
	_, checks["read its subtasks"] = a.GetSubtasks(ctx, "ws-a", "t-mine", "bobs", anna)
	_, checks["check its completion"] = a.GetTaskCompletion(ctx, "ws-a", "t-mine", "bobs", anna)
	_, checks["complete its subtasks"] = a.CompleteSubtasks(ctx, "ws-a", "t-mine", "bobs", "done", anna)
	_, checks["change a field"] = a.UpdateWorkspaceTask(ctx, "ws-a", "t-mine", "bobs", "name", "mine now", anna, "")
	_, checks["change its links"] = a.ChangeTaskLinks(ctx, "ws-a", "t-mine", "bobs", "links", []string{"x"}, nil, anna)
	_, checks["assign someone"] = a.AddMemberToTask(ctx, "ws-a", "t-mine", "bobs", "anna", "assignee", anna)
	checks["move its card"] = a.MoveKanbanCard(ctx, "ws-a", "t-mine", "v-kanban", "bobs", "s1", "", anna)

	for action, appErr := range checks {
		if appErr == nil || appErr.Status != 403 {
			t.Errorf("%s of bob's task on an assigned-only table: got %v, want 403", action, appErr)
		}
	}
}

func TestRequireTaskAccessOnlyLimitsAssignedOnlyTables(t *testing.T) {
	anna := model.User{ID: "anna", Role: model.SystemUserRoleId}
	a, _ := assignedOnlyApp(map[string]bool{"t-mine": true})

	if appErr := a.requireTaskAccess(anna, "ws-a", "t-mine", "own"); appErr != nil {
		t.Errorf("her own task: %v, want allowed", appErr)
	}
	for _, task := range []string{"bobs", "missing"} {
		if appErr := a.requireTaskAccess(anna, "ws-a", "t-mine", task); appErr == nil || appErr.Status != 403 {
			t.Errorf("%s on an assigned-only table: %v, want 403", task, appErr)
		}
	}
	if appErr := a.requireTaskAccess(anna, "ws-a", "t-open", "bobs"); appErr != nil {
		t.Errorf("bob's task on an open table: %v, want allowed", appErr)
	}
}

func (f *fakeAssignedOnlyStore) IsTableSingleSelect(tableID string) (bool, error) {
	return tableID == "t-statuses", nil
}

func TestAssignedOnlyLeavesOptionTablesWhole(t *testing.T) {
	anna := model.User{ID: "anna", Role: model.SystemUserRoleId}
	a, _ := assignedOnlyApp(map[string]bool{"t-mine": true, "t-statuses": true})

	if got := a.assignedOnlyFilter(anna, "ws-a", "t-statuses"); got.SQL != "" {
		t.Errorf("status options filtered by %q, want every option shown", got.SQL)
	}
	if got := a.assignedOnlyFilter(anna, "ws-a", "t-mine"); got.SQL != "`assignee` = ?" {
		t.Errorf("task table filter = %q, want her tasks only", got.SQL)
	}
}

func (f *fakeAssignedOnlyStore) GetTaskAssignees(_ context.Context, _ string, ids []string) (map[string]string, error) {
	owners := map[string]string{}
	for _, id := range ids {
		if assignee, err := f.GetTaskAssigneeID("", id); err == nil {
			owners[id] = assignee
		}
	}
	return owners, nil
}

func TestLinksToOthersTasksLoseTheirNames(t *testing.T) {
	anna := model.User{ID: "anna", Role: model.SystemUserRoleId}
	a, _ := assignedOnlyApp(map[string]bool{"t-mine": true})
	headers := []model.WorkspaceHeaders{
		{Name: "into_mine", LinkedID: "j-mine", ParentTableID: "t-mine"},
		{Name: "into_open", LinkedID: "j-open", ParentTableID: "t-open"},
	}
	links := func() []model.LinkedItem {
		return []model.LinkedItem{{ID: "own", Name: "Anna's"}, {ID: "bobs", Name: "Bob's"}}
	}
	rows := []map[string]interface{}{{"id": "row", "into_mine": links(), "into_open": links()}}

	a.restrictLinkedItems(context.Background(), rows, headers, anna, "ws-a")

	if got, want := rows[0]["into_mine"], []model.LinkedItem{{ID: "own", Name: "Anna's"}, {ID: "bobs", Restricted: true}}; !slices.Equal(got.([]model.LinkedItem), want) {
		t.Errorf("links into her own-tasks table = %v, want %v", got, want)
	}
	if got := rows[0]["into_open"]; !slices.Equal(got.([]model.LinkedItem), links()) {
		t.Errorf("links into an open table = %v, want both names kept", got)
	}
}

func (f *fakeAssignedOnlyStore) GetAttachment(fileID, workspaceID string) (*model.WorkspaceAttachment, error) {
	return f.GetUserAttachment(fileID)
}

func (f *fakeAssignedOnlyStore) GetUserAttachment(fileID string) (*model.WorkspaceAttachment, error) {
	tasks := map[string]string{"f-own": "own", "f-bobs": "bobs"}
	return &model.WorkspaceAttachment{ID: fileID, WorkspaceID: "ws-a", TableID: "t-mine", TaskID: tasks[fileID]}, nil
}

func (f *fakeAssignedOnlyStore) DeleteAttachment(fileID, _ string) error {
	f.deleted = append(f.deleted, fileID)
	return nil
}

func TestAssignedOnlyRefusesAttachmentsOnAnotherUsersTask(t *testing.T) {
	anna := model.User{ID: "anna", Role: model.SystemUserRoleId}
	a, ws := assignedOnlyApp(map[string]bool{"t-mine": true})
	ctx := context.Background()

	if _, appErr := a.GetWorkspaceAttachment("f-own", "ws-a", anna); appErr != nil {
		t.Errorf("a file on her own task: %v, want allowed", appErr)
	}

	upload := httptest.NewRequest(http.MethodPost, "/", nil)
	checks := map[string]*model.AppError{}
	_, checks["open it"] = a.GetWorkspaceAttachment("f-bobs", "ws-a", anna)
	_, checks["open it in the editor"] = a.GetUserWorkspaceAttachment("f-bobs", anna)
	checks["delete it"] = a.DeleteWorkspaceAttachment(ctx, "f-bobs", "ws-a", anna)
	_, checks["upload to the task"] = a.UploadWorkspaceFiles("ws-a", "t-mine", "bobs", "files", "", anna, upload)
	_, checks["upload to another workspace's table"] = a.UploadWorkspaceFiles("ws-a", "t-elsewhere", "own", "files", "", anna, upload)

	for action, appErr := range checks {
		if appErr == nil || appErr.Status != 403 {
			t.Errorf("%s: got %v, want 403", action, appErr)
		}
	}
	if len(ws.deleted) != 0 {
		t.Errorf("deleted %v, want nothing deleted", ws.deleted)
	}
}

type readOnlyRoles struct{}

func (readOnlyRoles) Resolve(interfaces.AccessInput) interfaces.WorkspaceAccess {
	return readOnlyAccess{}
}

type readOnlyAccess struct{ openAccess }

func (readOnlyAccess) CanPerformRowAction(_, _, _ string) bool { return false }

func TestAttachmentChangesNeedTheUpdateTaskPermission(t *testing.T) {
	anna := model.User{ID: "anna", Role: model.SystemUserRoleId}
	a, ws := assignedOnlyApp(nil)
	a.WorkspaceRoles = readOnlyRoles{}

	if _, appErr := a.GetWorkspaceAttachment("f-own", "ws-a", anna); appErr != nil {
		t.Errorf("opening a file on her own task: %v, want allowed", appErr)
	}

	upload := httptest.NewRequest(http.MethodPost, "/", nil)
	checks := map[string]*model.AppError{}
	checks["delete"] = a.DeleteWorkspaceAttachment(context.Background(), "f-own", "ws-a", anna)
	_, checks["upload"] = a.UploadWorkspaceFiles("ws-a", "t-mine", "own", "files", "", anna, upload)

	for action, appErr := range checks {
		if appErr == nil || appErr.Status != 403 {
			t.Errorf("%s on her own task without update_task: got %v, want 403", action, appErr)
		}
	}
	if len(ws.deleted) != 0 {
		t.Errorf("deleted %v, want nothing deleted", ws.deleted)
	}
}
