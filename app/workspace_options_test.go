// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"context"
	"net/http"
	"testing"

	"github.com/twigex/twigex/model"
)

type fakeOptionStore struct {
	*fakeTaskWriteStore
}

func (f fakeOptionStore) IsTableSingleSelect(tableID string) (bool, error) {
	return tableID == "t-priority" || tableID == "t-status" || tableID == "t-other-options", nil
}

func (f fakeOptionStore) GetOptionTableField(_ context.Context, tableID string) (string, string, error) {
	switch tableID {
	case "t-priority":
		return "t-a", "priority", nil
	case "t-status":
		return "t-a", "status", nil
	}

	return "", "", nil
}

var workspaceUser = model.User{ID: "user", Role: model.SystemUserRoleId}

func optionApp(roleID string) (*App, *fakeTaskWriteStore) {
	a, ws := taskWriteApp()

	role := *model.MakeDefaultProjectWorkspaceRoles()[roleID]
	ws.fakeAccessWorkspaceStore = fakeAccessWorkspaceStore{memberRoleName: role.Name, roles: []model.ProjectWorkspaceRole{role}}
	ws.options["t-priority"] = true

	a.Store.Workspace = fakeOptionStore{fakeTaskWriteStore: ws}

	return a, ws
}

func TestUserRoleCannotChangeAChoice(t *testing.T) {
	a, _ := optionApp(model.ProjectWorkspaceUserRoleID)

	_, appErr := a.UpdateWorkspaceTask(context.Background(), "ws", "t-priority", "high", "name", "Urgent", workspaceUser, "")
	if appErr == nil || appErr.Status != http.StatusForbidden {
		t.Fatalf("want 403, got %v", appErr)
	}
}

func TestUserRoleCannotDeleteAChoice(t *testing.T) {
	a, _ := optionApp(model.ProjectWorkspaceUserRoleID)

	_, appErr := a.DeleteWorkspaceTask(context.Background(), "ws", "t-priority", "high", workspaceUser)
	if appErr == nil || appErr.Status != http.StatusForbidden {
		t.Fatalf("want 403, got %v", appErr)
	}
}

func TestUserRoleCannotRenameAChoiceFromItsHeader(t *testing.T) {
	a, _ := optionApp(model.ProjectWorkspaceUserRoleID)

	_, appErr := a.UpdateWorkspaceSingleSelectName(context.Background(), "ws", "t-a", "high", "name", "Urgent", workspaceUser, "t-priority")
	if appErr == nil || appErr.Status != http.StatusForbidden {
		t.Fatalf("want 403, got %v", appErr)
	}
}

func TestUserRoleCanStillPickAChoiceForATask(t *testing.T) {
	a, ws := optionApp(model.ProjectWorkspaceUserRoleID)

	_, appErr := a.UpdateWorkspaceTask(context.Background(), "ws", "t-a", "task-1", "priority", "high", workspaceUser, "")
	if appErr != nil {
		t.Fatalf("want the task updated, got %v", appErr)
	}

	if ws.writes != 1 {
		t.Errorf("want one write, got %d", ws.writes)
	}
}

func TestAdminCannotDeleteTheCompletedStatus(t *testing.T) {
	a, ws := optionApp(model.ProjectWorkspaceAdminRoleID)
	ws.stored = map[string]string{"name": "Completed"}

	_, appErr := a.DeleteWorkspaceTask(context.Background(), "ws", "t-status", "done", workspaceUser)
	requireAppErr(t, appErr, "workspace.task_delete_protected", http.StatusBadRequest)
}

func TestAdminCannotRenameTheCompletedStatus(t *testing.T) {
	a, ws := optionApp(model.ProjectWorkspaceAdminRoleID)
	ws.stored = map[string]string{"name": "Completed"}

	_, appErr := a.UpdateWorkspaceTask(context.Background(), "ws", "t-status", "done", "name", "Finished", workspaceUser, "")
	requireAppErr(t, appErr, "workspace.task_update_protected", http.StatusBadRequest)
}

func TestAdminCanDeleteAChoiceCalledCompletedInAnotherField(t *testing.T) {
	a, ws := optionApp(model.ProjectWorkspaceAdminRoleID)
	ws.stored = map[string]string{"name": "Completed"}

	_, appErr := a.DeleteWorkspaceTask(context.Background(), "ws", "t-priority", "done", workspaceUser)
	requireAppErr(t, appErr, "workspace.task_delete_failed", http.StatusInternalServerError)
}

func TestAdminCanRenameAChoiceCalledCompletedInAnotherField(t *testing.T) {
	a, ws := optionApp(model.ProjectWorkspaceAdminRoleID)
	ws.stored = map[string]string{"name": "Completed"}

	_, appErr := a.UpdateWorkspaceTask(context.Background(), "ws", "t-priority", "done", "name", "Finished", workspaceUser, "")
	if appErr != nil {
		t.Fatalf("want the choice renamed, got %v", appErr)
	}

	if ws.writes != 1 {
		t.Errorf("want one write, got %d", ws.writes)
	}
}

func TestATaskCalledCompletedCanBeRenamed(t *testing.T) {
	a, ws := optionApp(model.ProjectWorkspaceAdminRoleID)
	ws.stored = map[string]string{"name": "Completed"}

	_, appErr := a.UpdateWorkspaceTask(context.Background(), "ws", "t-a", "task-1", "name", "Write the report", workspaceUser, "")
	if appErr != nil {
		t.Fatalf("want the task renamed, got %v", appErr)
	}

	if ws.writes != 1 {
		t.Errorf("want one write, got %d", ws.writes)
	}
}

func TestRenamingAChoiceOfAnotherTablesFieldIsForbidden(t *testing.T) {
	a, _ := optionApp(model.ProjectWorkspaceAdminRoleID)

	_, appErr := a.UpdateWorkspaceSingleSelectName(context.Background(), "ws", "t-a", "high", "name", "Urgent", workspaceUser, "t-other-options")
	if appErr == nil || appErr.Status != http.StatusForbidden {
		t.Fatalf("want 403, got %v", appErr)
	}
}

func TestUserRoleCannotAddAChoice(t *testing.T) {
	a, _ := optionApp(model.ProjectWorkspaceUserRoleID)

	_, appErr := a.AddFieldValue(context.Background(), "ws", "t-a", "Urgent", "priority", workspaceUser)
	if appErr == nil || appErr.Status != http.StatusForbidden {
		t.Fatalf("want 403, got %v", appErr)
	}
}

func TestUserRoleCannotAddAChoiceAsATask(t *testing.T) {
	a, ws := optionApp(model.ProjectWorkspaceUserRoleID)

	_, appErr := a.CreateWorkspaceTask(context.Background(), "ws", "t-priority", "Urgent", "", "", model.NewTaskFields{}, workspaceUser, "")
	if appErr == nil || appErr.Status != http.StatusForbidden {
		t.Fatalf("want 403, got %v", appErr)
	}

	if ws.created != nil {
		t.Errorf("want nothing created, got %v", ws.created)
	}
}
