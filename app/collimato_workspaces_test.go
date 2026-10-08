// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"context"
	"net/http"
	"testing"

	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/store"
)

type fakeCollimatoListStore struct {
	store.CollimatoStore
	all       []model.CollimatoWorkspace
	memberOf  map[string][]model.CollimatoWorkspace
	roles     map[string][]string
	listedAll bool
}

func (f *fakeCollimatoListStore) GetWorkspaces() ([]model.CollimatoWorkspace, error) {
	f.listedAll = true
	return f.all, nil
}

func (f *fakeCollimatoListStore) GetWorkspacesForUser(userID string) ([]model.CollimatoWorkspace, error) {
	return f.memberOf[userID], nil
}

func (f *fakeCollimatoListStore) GetWorkspaceMemberIDs(context.Context, []string) (map[string][]string, error) {
	return map[string][]string{}, nil
}

func (f *fakeCollimatoListStore) GetEffectiveRolesForUser(_ context.Context, workspaceID, _ string) ([]string, error) {
	return f.roles[workspaceID], nil
}

type fakeCollimatoAdminStore struct {
	store.CollimatoStore
	workspace *model.CollimatoWorkspace
}

func (f *fakeCollimatoAdminStore) GetWorkspaceByID(string) (*model.CollimatoWorkspace, error) {
	return f.workspace, nil
}

func (f *fakeCollimatoAdminStore) GetEffectiveRolesForUser(context.Context, string, string) ([]string, error) {
	return nil, nil
}

func TestCollimatoAdminsGetNoShortcutIntoOtherWorkspaces(t *testing.T) {
	a := &App{Store: store.Store{Collimato: &fakeCollimatoAdminStore{
		workspace: &model.CollimatoWorkspace{ID: "ws", CreatedBy: "owner"},
	}}}
	admin := model.User{ID: "admin", Role: model.SystemAdminRoleId}

	if appErr := a.DeleteWorkspace(admin, "ws"); appErr == nil || appErr.Status != http.StatusForbidden {
		t.Errorf("DeleteWorkspace: want 403, got %v", appErr)
	}

	if _, appErr := a.FinishWorkspaceCreation(admin, "ws", model.CollimatoWorkspaceStatusDraft); appErr == nil || appErr.Status != http.StatusForbidden {
		t.Errorf("FinishWorkspaceCreation: want 403, got %v", appErr)
	}

	if _, appErr := a.GetWorkspaceUsers(admin, "ws"); appErr == nil || appErr.Status != http.StatusForbidden {
		t.Errorf("GetWorkspaceUsers: want 403, got %v", appErr)
	}
}

func TestGetWorkspacesShowsAdminsOnlyTheirOwn(t *testing.T) {
	mine := model.CollimatoWorkspace{ID: "mine"}
	theirs := model.CollimatoWorkspace{ID: "theirs"}

	collimato := &fakeCollimatoListStore{
		all:      []model.CollimatoWorkspace{mine, theirs},
		memberOf: map[string][]model.CollimatoWorkspace{"admin": {mine}},
	}

	a := &App{
		Store:       store.Store{Collimato: collimato},
		ConfigStore: photoTestConfig(t.TempDir()),
	}
	admin := model.User{ID: "admin", Role: model.SystemAdminRoleId}

	got, appErr := a.GetWorkspaces(context.Background(), admin)
	if appErr != nil {
		t.Fatalf("GetWorkspaces: %v", appErr)
	}

	if len(got) != 1 || got[0].ID != "mine" {
		t.Fatalf("expected only the workspace the admin belongs to, got %v", got)
	}

	if collimato.listedAll {
		t.Error("expected the admin list not to read every workspace")
	}
}

func TestGetWorkspacesOffersDeleteToWorkspaceAdmins(t *testing.T) {
	led := model.CollimatoWorkspace{ID: "led", CreatedBy: "someone"}
	joined := model.CollimatoWorkspace{ID: "joined", CreatedBy: "u1"}

	collimato := &fakeCollimatoListStore{
		memberOf: map[string][]model.CollimatoWorkspace{"u1": {led, joined}},
		roles: map[string][]string{
			"led":    {model.CollimatoWorkspaceAdminRoleId},
			"joined": {model.CollimatoWorkspaceUserRoleId},
		},
	}

	a := &App{
		Store:       store.Store{Collimato: collimato},
		ConfigStore: photoTestConfig(t.TempDir()),
	}

	got, appErr := a.GetWorkspaces(context.Background(), model.User{ID: "u1", Role: model.SystemAdminRoleId})
	if appErr != nil {
		t.Fatalf("GetWorkspaces: %v", appErr)
	}

	if len(got) != 2 || !got[0].CanDelete || got[1].CanDelete {
		t.Errorf("expected delete only where the user is a workspace admin, got %+v", got)
	}
}
