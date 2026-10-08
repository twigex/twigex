// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"context"
	"net/http"
	"testing"

	"github.com/twigex/twigex/model"
)

type fakeRoleDeleteWorkspaceStore struct {
	fakeAccessWorkspaceStore
	role     model.ProjectWorkspaceRole
	deleted  string
	fallback string
}

func (f *fakeRoleDeleteWorkspaceStore) GetRoleByID(_, _ string) (*model.ProjectWorkspaceRole, error) {
	return &f.role, nil
}

func (f *fakeRoleDeleteWorkspaceStore) DeleteRole(_ context.Context, _, _, roleName, fallback string) error {
	f.deleted = roleName
	f.fallback = fallback

	return nil
}

func TestDeleteProjectWorkspaceRoleFallsBackToUser(t *testing.T) {
	ws := &fakeRoleDeleteWorkspaceStore{
		fakeAccessWorkspaceStore: workspaceAdminAccess(),
		role:                     model.ProjectWorkspaceRole{ID: "r1", Name: "editor"},
	}
	a := accessApp(&ws.fakeAccessWorkspaceStore)
	a.Store.Workspace = ws
	a.ConfigStore = photoTestConfig("")

	if _, appErr := a.DeleteProjectWorkspaceRole(context.Background(), workspaceAdmin, "w1", "r1"); appErr != nil {
		t.Fatalf("DeleteProjectWorkspaceRole: %v", appErr)
	}

	if ws.deleted != "editor" || ws.fallback != model.ProjectWorkspaceUserRoleID {
		t.Errorf("expected editor deleted with the user fallback, got %q and %q", ws.deleted, ws.fallback)
	}
}

func TestDeleteProjectWorkspaceRoleKeepsBuiltInRoles(t *testing.T) {
	for _, name := range []string{model.ProjectWorkspaceAdminRoleID, model.ProjectWorkspaceUserRoleID} {
		ws := &fakeRoleDeleteWorkspaceStore{
			fakeAccessWorkspaceStore: workspaceAdminAccess(),
			role:                     model.ProjectWorkspaceRole{ID: "r1", Name: name},
		}
		a := accessApp(&ws.fakeAccessWorkspaceStore)
		a.Store.Workspace = ws

		_, appErr := a.DeleteProjectWorkspaceRole(context.Background(), workspaceAdmin, "w1", "r1")
		if appErr == nil || appErr.Status != http.StatusConflict {
			t.Errorf("deleting %s: want 409, got %v", name, appErr)
		}

		if ws.deleted != "" {
			t.Errorf("expected %s to stay", name)
		}
	}
}
