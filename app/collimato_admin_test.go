// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"context"
	"net/http"
	"slices"
	"strings"
	"testing"

	"github.com/twigex/twigex/interfaces"
	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/store"
)

type fakeWorkspaceManagementStore struct {
	store.CollimatoStore
	workspace *model.CollimatoWorkspace
	members   map[string]*model.CollimatoWorkspaceUser
	removed   []string
	deleted   bool
}

func (f *fakeWorkspaceManagementStore) GetWorkspaceByID(string) (*model.CollimatoWorkspace, error) {
	return f.workspace, nil
}

func (f *fakeWorkspaceManagementStore) GetWorkspaces() ([]model.CollimatoWorkspace, error) {
	return []model.CollimatoWorkspace{*f.workspace}, nil
}

func (f *fakeWorkspaceManagementStore) GetWorkspaceMemberIDs(context.Context, []string) (map[string][]string, error) {
	return map[string][]string{}, nil
}

func (f *fakeWorkspaceManagementStore) GetEffectiveRolesForUser(_ context.Context, _, userID string) ([]string, error) {
	member, ok := f.members[userID]
	if !ok {
		return nil, nil
	}

	return strings.Split(member.Role, ","), nil
}

func (f *fakeWorkspaceManagementStore) GetWorkspaceUsers(string) ([]model.CollimatoWorkspaceUser, error) {
	users := []model.CollimatoWorkspaceUser{}
	for _, member := range f.members {
		users = append(users, *member)
	}

	return users, nil
}

func (f *fakeWorkspaceManagementStore) AddWorkspaceUsers(_ string, users []string, roles []string) error {
	for _, userID := range users {
		if _, ok := f.members[userID]; ok {
			continue
		}

		f.members[userID] = &model.CollimatoWorkspaceUser{
			ID:     "row-" + userID,
			UserID: userID,
			Role:   roles[0],
		}
	}

	return nil
}

func (f *fakeWorkspaceManagementStore) GetWorkspaceUserByUserID(userID, _ string) (*model.CollimatoWorkspaceUser, error) {
	return f.members[userID], nil
}

func (f *fakeWorkspaceManagementStore) UpdateUserRoles(_ string, memberID string, roles []string) error {
	for _, member := range f.members {
		if member.ID == memberID {
			member.Role = strings.Join(roles, ",")
		}
	}

	return nil
}

func (f *fakeWorkspaceManagementStore) RemoveWorkspaceUser(_ string, userID string) error {
	f.removed = append(f.removed, userID)
	return nil
}

func (f *fakeWorkspaceManagementStore) DeleteWorkspace(string) error {
	f.deleted = true
	return nil
}

type fakeRoleAssigner struct {
	interfaces.CollimatoRoles
	assigned map[string][]string
}

func (f *fakeRoleAssigner) AssignUserRoles(_ string, memberID string, roles []string) *model.AppError {
	f.assigned[memberID] = roles
	return nil
}

func (f *fakeRoleAssigner) AssignGroupRoles(_ string, groupID string, roles []string) *model.AppError {
	f.assigned[groupID] = roles
	return nil
}

func workspaceManagementApp(collimato *fakeWorkspaceManagementStore) *App {
	return &App{
		Store: store.Store{
			Collimato: collimato,
			Roles:     &fakeRoleStore{known: map[string]bool{}},
		},
		ConfigStore:    photoTestConfig(""),
		CollimatoRoles: &fakeRoleAssigner{assigned: map[string][]string{}},
	}
}

func ownedWorkspace() *fakeWorkspaceManagementStore {
	return &fakeWorkspaceManagementStore{
		workspace: &model.CollimatoWorkspace{ID: "ws", CreatedBy: "owner"},
		members: map[string]*model.CollimatoWorkspaceUser{
			"owner":  {ID: "row-owner", UserID: "owner", Role: model.CollimatoWorkspaceAdminRoleId},
			"member": {ID: "row-member", UserID: "member", Role: model.CollimatoWorkspaceUserRoleId},
		},
	}
}

func TestWorkspaceManagementRefusesUsersWithoutPermission(t *testing.T) {
	collimato := ownedWorkspace()
	a := workspaceManagementApp(collimato)
	user := regularUser()

	if _, appErr := a.ListAllWorkspaces(context.Background(), user); appErr == nil || appErr.Status != http.StatusForbidden {
		t.Errorf("ListAllWorkspaces: want 403, got %v", appErr)
	}

	if _, appErr := a.GetWorkspaceDetails(context.Background(), user, "ws"); appErr == nil || appErr.Status != http.StatusForbidden {
		t.Errorf("GetWorkspaceDetails: want 403, got %v", appErr)
	}

	if appErr := a.DeleteManagedWorkspace(user, "ws"); appErr == nil || appErr.Status != http.StatusForbidden {
		t.Errorf("DeleteManagedWorkspace: want 403, got %v", appErr)
	}

	if appErr := a.JoinManagedWorkspace(user, "ws"); appErr == nil || appErr.Status != http.StatusForbidden {
		t.Errorf("JoinManagedWorkspace: want 403, got %v", appErr)
	}

	if collimato.deleted || collimato.members["u1"] != nil {
		t.Error("expected nothing to change for a user without the permission")
	}
}

func TestWorkspaceManagementGrantedThroughRole(t *testing.T) {
	a := workspaceManagementApp(ownedWorkspace())
	a.Store.Roles = &permFakeRoleStore{perms: map[string][]string{
		"collimato_managers": {model.CollimatoSectionPermissions.PermissionManageCollimato.Id},
	}}

	user := model.User{ID: "u1", Role: "collimato_managers"}

	got, appErr := a.ListAllWorkspaces(context.Background(), user)
	if appErr != nil {
		t.Fatalf("ListAllWorkspaces: %v", appErr)
	}

	if len(got) != 1 || got[0].ID != "ws" {
		t.Fatalf("expected every workspace, got %v", got)
	}
}

func TestDeleteManagedWorkspaceNeedsNoMembership(t *testing.T) {
	collimato := ownedWorkspace()
	a := workspaceManagementApp(collimato)

	if appErr := a.DeleteManagedWorkspace(admin(), "ws"); appErr != nil {
		t.Fatalf("DeleteManagedWorkspace: %v", appErr)
	}

	if !collimato.deleted {
		t.Error("expected the workspace to be deleted")
	}
}

func TestRemoveManagedWorkspaceUserKeepsLastAdmin(t *testing.T) {
	collimato := ownedWorkspace()
	a := workspaceManagementApp(collimato)

	if appErr := a.RemoveManagedWorkspaceUser(admin(), "ws", "owner"); appErr == nil || appErr.Status != http.StatusBadRequest {
		t.Errorf("removing the last admin: want 400, got %v", appErr)
	}

	if appErr := a.RemoveManagedWorkspaceUser(admin(), "ws", "member"); appErr != nil {
		t.Fatalf("removing a member: %v", appErr)
	}

	collimato.members["second"] = &model.CollimatoWorkspaceUser{ID: "row-second", UserID: "second", Role: model.CollimatoWorkspaceAdminRoleId}

	if appErr := a.RemoveManagedWorkspaceUser(admin(), "ws", "owner"); appErr != nil {
		t.Fatalf("removing the creator once another admin exists: %v", appErr)
	}

	if !slices.Equal(collimato.removed, []string{"member", "owner"}) {
		t.Errorf("expected the member and then the creator removed, got %v", collimato.removed)
	}
}

func TestUpdateManagedWorkspaceUserRolesKeepsLastAdmin(t *testing.T) {
	collimato := ownedWorkspace()
	collimato.members["owner"].Role = "analyst," + model.CollimatoWorkspaceAdminRoleId

	a := workspaceManagementApp(collimato)
	assigner := a.CollimatoRoles.(*fakeRoleAssigner)

	appErr := a.UpdateManagedWorkspaceUserRoles(admin(), "ws", "owner", []string{"analyst"})
	if appErr == nil || appErr.Status != http.StatusBadRequest {
		t.Errorf("dropping the last admin role: want 400, got %v", appErr)
	}

	appErr = a.UpdateManagedWorkspaceUserRoles(admin(), "ws", "member", []string{"analyst"})
	if appErr != nil {
		t.Fatalf("changing a member's roles: %v", appErr)
	}

	if _, ok := assigner.assigned["row-owner"]; ok {
		t.Error("expected the last admin's roles to stay unchanged")
	}

	if !slices.Equal(assigner.assigned["row-member"], []string{"analyst"}) {
		t.Errorf("expected the member row to get the new roles, got %v", assigner.assigned)
	}
}

func TestDeleteWorkspaceFromInsideNeedsWorkspaceAdmin(t *testing.T) {
	collimato := ownedWorkspace()
	collimato.members["owner"].Role = model.CollimatoWorkspaceUserRoleId
	collimato.members["member"].Role = model.CollimatoWorkspaceAdminRoleId

	a := workspaceManagementApp(collimato)

	if appErr := a.DeleteWorkspace(model.User{ID: "owner"}, "ws"); appErr == nil || appErr.Status != http.StatusForbidden {
		t.Errorf("creator without the admin role: want 403, got %v", appErr)
	}

	if collimato.deleted {
		t.Fatal("expected the workspace to stay")
	}

	if appErr := a.DeleteWorkspace(model.User{ID: "member"}, "ws"); appErr != nil {
		t.Fatalf("workspace admin: %v", appErr)
	}

	if !collimato.deleted {
		t.Error("expected the workspace to be deleted")
	}
}

func TestJoinManagedWorkspaceAddsAdminRoleToExistingMember(t *testing.T) {
	collimato := ownedWorkspace()
	collimato.members["admin"] = &model.CollimatoWorkspaceUser{
		ID:     "row-admin",
		UserID: "admin",
		Role:   model.CollimatoWorkspaceUserRoleId,
	}

	a := workspaceManagementApp(collimato)

	if appErr := a.JoinManagedWorkspace(admin(), "ws"); appErr != nil {
		t.Fatalf("JoinManagedWorkspace: %v", appErr)
	}

	want := model.CollimatoWorkspaceUserRoleId + "," + model.CollimatoWorkspaceAdminRoleId
	if got := collimato.members["admin"].Role; got != want {
		t.Errorf("expected roles %q, got %q", want, got)
	}
}

func TestAssigningRolesInsideAWorkspaceNeedsWorkspacePermission(t *testing.T) {
	a := workspaceManagementApp(ownedWorkspace())
	assigner := a.CollimatoRoles.(*fakeRoleAssigner)

	if appErr := a.UpdateUserRoles(admin(), "ws", "row-member", []string{"analyst"}); appErr == nil || appErr.Status != http.StatusForbidden {
		t.Errorf("UpdateUserRoles: want 403, got %v", appErr)
	}

	if appErr := a.UpdateWorkspaceGroupRoles(admin(), "ws", "g1", []string{"analyst"}); appErr == nil || appErr.Status != http.StatusForbidden {
		t.Errorf("UpdateWorkspaceGroupRoles: want 403, got %v", appErr)
	}

	if len(assigner.assigned) != 0 {
		t.Errorf("expected no roles assigned, got %v", assigner.assigned)
	}
}

func TestAddManagedWorkspaceGroupsChecksPermissionBeforeLicense(t *testing.T) {
	a := workspaceManagementApp(ownedWorkspace())

	if _, appErr := a.AddManagedWorkspaceGroups(regularUser(), "ws", []string{"g1"}, []string{"workspace_user"}); appErr == nil || appErr.Status != http.StatusForbidden {
		t.Errorf("without the permission: want 403, got %v", appErr)
	}

	if _, appErr := a.AddManagedWorkspaceGroups(admin(), "ws", []string{"g1"}, []string{"workspace_user"}); appErr == nil || appErr.Status != http.StatusPaymentRequired {
		t.Errorf("without a groups licence: want 402, got %v", appErr)
	}
}

func TestManagedRoleChangesNeedTheEnterpriseBuild(t *testing.T) {
	a := workspaceManagementApp(ownedWorkspace())
	a.CollimatoRoles = nil

	if appErr := a.UpdateManagedWorkspaceUserRoles(admin(), "ws", "member", []string{"analyst"}); appErr == nil || appErr.Status != http.StatusPaymentRequired {
		t.Errorf("UpdateManagedWorkspaceUserRoles: want 402, got %v", appErr)
	}

	if appErr := a.UpdateManagedWorkspaceGroupRoles(admin(), "ws", "g1", []string{"analyst"}); appErr == nil || appErr.Status != http.StatusPaymentRequired {
		t.Errorf("UpdateManagedWorkspaceGroupRoles: want 402, got %v", appErr)
	}
}
