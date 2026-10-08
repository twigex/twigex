// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"context"
	"net/http"
	"testing"

	"github.com/twigex/twigex/config"
	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/store"
)

type fakeReadAccessStore struct {
	fakeAccessWorkspaceStore
	groupMember bool
	tableOwner  map[string]string
	tableRows   map[string]bool
}

func (f *fakeReadAccessStore) UserHasAnyGroupAccess(_, _ string) (bool, error) {
	return f.groupMember, nil
}

func (f *fakeReadAccessStore) GetTableWorkspaceID(_ context.Context, tableID string) (string, error) {
	return f.tableOwner[tableID], nil
}

func (f *fakeReadAccessStore) GetTableName(tableID string) (string, error) {
	return "tbl_" + tableID, nil
}

func (f *fakeReadAccessStore) GetTableRowByID(_, itemID string) (map[string]interface{}, error) {
	if !f.tableRows[itemID] {
		return nil, nil
	}
	return map[string]interface{}{"id": itemID}, nil
}

func (f *fakeReadAccessStore) GetComments(string) ([]model.TaskComment, error) {
	return []model.TaskComment{{ID: "c1"}}, nil
}

func (f *fakeReadAccessStore) GetTaskActivities(string) ([]model.Activity, error) {
	return nil, nil
}

func readAccessApp(ws *fakeReadAccessStore) *App {
	return &App{
		Store: store.Store{Workspace: ws},
		ConfigStore: config.ConfigStore{Config: &model.ServerConfig{
			SqlSettings: model.SqlSettings{QueryTimeout: model.NewInt(25)},
		}},
	}
}

func TestCanReadTableRequiresMembershipAndTheTableInTheWorkspace(t *testing.T) {
	owners := map[string]string{"t-a": "ws-a", "t-b": "ws-b"}
	member := model.User{ID: "u1", Role: model.SystemUserRoleId}
	admin := model.User{ID: "admin", Role: model.SystemAdminRoleId}

	cases := []struct {
		name        string
		user        model.User
		memberRole  string
		groupMember bool
		workspace   string
		table       string
		want        bool
	}{
		{"direct member, own table", member, "member", false, "ws-a", "t-a", true},
		{"group member, own table", member, "", true, "ws-a", "t-a", true},
		{"not a member", member, "", false, "ws-a", "t-a", false},
		{"member, table from another workspace", member, "member", false, "ws-a", "t-b", false},
		{"member, no such table", member, "member", false, "ws-a", "t-missing", false},
		{"system admin outside the workspace", admin, "", false, "ws-a", "t-a", false},
		{"system admin who is a member", admin, "member", false, "ws-a", "t-a", true},
		{"admin, table from another workspace", admin, "member", false, "ws-a", "t-b", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ws := &fakeReadAccessStore{
				fakeAccessWorkspaceStore: fakeAccessWorkspaceStore{memberRoleName: tc.memberRole},
				groupMember:              tc.groupMember,
				tableOwner:               owners,
			}
			if got := readAccessApp(ws).canReadTable(context.Background(), tc.user, tc.workspace, tc.table); got != tc.want {
				t.Errorf("canReadTable = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestCheckTableInWorkspace(t *testing.T) {
	a := readAccessApp(&fakeReadAccessStore{tableOwner: map[string]string{"t-a": "ws-a", "t-b": "ws-b"}})
	for _, tc := range []struct {
		workspace, table string
		want             int
	}{
		{"ws-a", "t-a", 0},
		{"ws-a", "t-b", http.StatusForbidden},
		{"ws-a", "t-missing", http.StatusForbidden},
		{"all", "detailed-task-report", 0},
		{"all", "t-a", http.StatusForbidden},
	} {
		appErr := a.CheckTableInWorkspace(context.Background(), tc.workspace, tc.table)
		got := 0
		if appErr != nil {
			got = appErr.Status
		}
		if got != tc.want {
			t.Errorf("workspace %s, table %s: status %d, want %d", tc.workspace, tc.table, got, tc.want)
		}
	}
}

func TestTaskReadsRejectNonMembers(t *testing.T) {
	ws := &fakeReadAccessStore{tableOwner: map[string]string{"t-a": "ws-a"}, tableRows: map[string]bool{"task-1": true}}
	a := readAccessApp(ws)
	outsider := model.User{ID: "outsider", Role: model.SystemUserRoleId}
	ctx := context.Background()

	checks := map[string]*model.AppError{}
	_, _, checks["GetItemForTableByID"] = a.GetItemForTableByID(ctx, "ws-a", "t-a", "task-1", outsider)
	_, checks["GetSubtasks"] = a.GetSubtasks(ctx, "ws-a", "t-a", "task-1", outsider)
	_, checks["GetTaskCompletion"] = a.GetTaskCompletion(ctx, "ws-a", "t-a", "task-1", outsider)
	_, checks["CompleteSubtasks"] = a.CompleteSubtasks(ctx, "ws-a", "t-a", "task-1", "done", outsider)
	_, checks["GetTaskComments"] = a.GetTaskComments(ctx, "ws-a", "t-a", "task-1", outsider)
	_, checks["GetFilteredTableData"] = a.GetFilteredTableData(ctx, "ws-a", "t-a", "", model.FilterPayload{}, outsider, true)
	_, checks["GetFilteredTableCount"] = a.GetFilteredTableCount(ctx, "ws-a", "t-a", "", model.FilterPayload{}, outsider)
	_, checks["GetTaskPageNumber"] = a.GetTaskPageNumber(ctx, "ws-a", "t-a", "task-1", 100, model.FilterPayload{}, outsider)
	_, _, checks["GetTableStatusTypes"] = a.GetTableStatusTypes(ctx, "ws-a", "t-a", outsider)
	for name, appErr := range checks {
		if appErr == nil || appErr.Status != http.StatusForbidden {
			t.Errorf("%s for a non-member: got %v, want 403", name, appErr)
		}
	}
}

func TestTaskCommentsNeedTheTaskInTheTable(t *testing.T) {
	ws := &fakeReadAccessStore{
		fakeAccessWorkspaceStore: fakeAccessWorkspaceStore{memberRoleName: "member"},
		tableOwner:               map[string]string{"t-a": "ws-a"},
		tableRows:                map[string]bool{"task-in-a": true},
	}
	a := readAccessApp(ws)
	member := model.User{ID: "u1", Role: model.SystemUserRoleId}

	if _, appErr := a.GetTaskComments(context.Background(), "ws-a", "t-a", "task-elsewhere", member); appErr == nil || appErr.Status != http.StatusForbidden {
		t.Errorf("comments of a task outside the table: got %v, want 403", appErr)
	}
	comments, appErr := a.GetTaskComments(context.Background(), "ws-a", "t-a", "task-in-a", member)
	if appErr != nil || len(comments) != 1 {
		t.Errorf("comments of a task in the table = %v, %v; want the one comment", comments, appErr)
	}
}

type fakeWriteAccessStore struct {
	fakeReadAccessStore
}

func (f *fakeWriteAccessStore) IsTableSingleSelect(string) (bool, error) { return false, nil }

func (f *fakeWriteAccessStore) GetTableParentID(string) string { return "" }

func TestTaskWritesRejectNonMembers(t *testing.T) {
	ws := &fakeWriteAccessStore{fakeReadAccessStore{tableOwner: map[string]string{"t-a": "ws-a"}}}
	a := readAccessApp(&ws.fakeReadAccessStore)
	a.Store.Workspace = ws
	outsider := model.User{ID: "outsider", Role: model.SystemUserRoleId}
	ctx := context.Background()

	checks := map[string]*model.AppError{}
	_, checks["UpdateWorkspaceTask"] = a.UpdateWorkspaceTask(ctx, "ws-a", "t-a", "task-1", "name", "mine", outsider, "")
	_, checks["DeleteWorkspaceTask"] = a.DeleteWorkspaceTask(ctx, "ws-a", "t-a", "task-1", outsider)
	_, checks["UpdateWorkspaceTable"] = a.UpdateWorkspaceTable(ctx, "ws-a", "t-a", "renamed", outsider)
	_, checks["CreateView"] = a.CreateView(ctx, "ws-a", "t-a", "", "mine", "kanban", true, outsider, "", "")
	for name, appErr := range checks {
		if appErr == nil || appErr.Status != http.StatusForbidden {
			t.Errorf("%s for a non-member: got %v, want 403", name, appErr)
		}
	}
}

func TestRowActionsNeedMembershipWithoutEnterpriseRoles(t *testing.T) {
	for _, tc := range []struct {
		name        string
		user        model.User
		memberRole  string
		groupMember bool
		want        bool
	}{
		{"direct member", model.User{ID: "u1", Role: model.SystemUserRoleId}, "member", false, true},
		{"group member", model.User{ID: "u1", Role: model.SystemUserRoleId}, "", true, true},
		{"not a member", model.User{ID: "u1", Role: model.SystemUserRoleId}, "", false, false},
		{"system admin outside the workspace", model.User{ID: "admin", Role: model.SystemAdminRoleId}, "", false, false},
		{"system admin who is a member", model.User{ID: "admin", Role: model.SystemAdminRoleId}, "member", false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ws := &fakeReadAccessStore{
				fakeAccessWorkspaceStore: fakeAccessWorkspaceStore{memberRoleName: tc.memberRole},
				groupMember:              tc.groupMember,
			}
			a := readAccessApp(ws)

			if got := a.CanPerformRowAction(tc.user, "ws-a", "t-a", model.PermissionDeleteTask, model.PermissionDeleteTask); got != tc.want {
				t.Errorf("CanPerformRowAction = %v, want %v", got, tc.want)
			}
			if got := a.ProjectWorkspaceTableHasPermission(tc.user, "ws-a", "t-a", "manage_views"); got != tc.want {
				t.Errorf("ProjectWorkspaceTableHasPermission = %v, want %v", got, tc.want)
			}
		})
	}
}
