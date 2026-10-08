// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

// OSS-build access-control tests: verify that without the enterprise binary
// every call returns the permissive no-op (openAccess). Enforcement tests
// live in cloud-ee alongside the implementation.

import (
	"context"
	"database/sql"
	"net/http"
	"testing"
	"time"

	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/store"
)

type fakeAccessWorkspaceStore struct {
	store.WorkspaceStore
	memberRoleName string
	roles          []model.ProjectWorkspaceRole
	tablePerms     []model.TablePermission
	// onlyMember, when set, is the one user who is a member; otherwise
	// everyone is while memberRoleName is set.
	onlyMember string
}

func (f *fakeAccessWorkspaceStore) isMember(userID string) bool {
	return f.memberRoleName != "" && (f.onlyMember == "" || f.onlyMember == userID)
}

func (f *fakeAccessWorkspaceStore) GetViewVisibility(_ context.Context, _ string) (bool, string, error) {
	return true, "", nil
}

func (f *fakeAccessWorkspaceStore) IsMember(_, userID string) (bool, error) {
	return f.isMember(userID), nil
}
func (f *fakeAccessWorkspaceStore) UserHasAnyGroupAccess(_, _ string) (bool, error) {
	return false, nil
}
func (f *fakeAccessWorkspaceStore) GetUserByUserID(userID, _ string) (*model.WorkspaceMember, error) {
	if !f.isMember(userID) {
		return nil, nil
	}
	return &model.WorkspaceMember{Role: f.memberRoleName}, nil
}
func (f *fakeAccessWorkspaceStore) GetGroupRolesForUser(_, _ string) ([]string, error) {
	return nil, nil
}
func (f *fakeAccessWorkspaceStore) GetRolesByName(names []string, _ string) ([]model.ProjectWorkspaceRole, error) {
	var out []model.ProjectWorkspaceRole
	for _, r := range f.roles {
		for _, n := range names {
			if r.Name == n {
				out = append(out, r)
			}
		}
	}
	return out, nil
}
func (f *fakeAccessWorkspaceStore) GetTablePermissionsForRoles(_ []string, _ string) ([]model.TablePermission, error) {
	return f.tablePerms, nil
}
func (f *fakeAccessWorkspaceStore) GetAllTablesBasic(_ string) ([]model.WorkspaceTable, error) {
	return nil, nil
}
func (f *fakeAccessWorkspaceStore) GetSharedViewIDsForUser(_, _ string) ([]string, error) {
	return nil, nil
}

// workspaceAdmin is a member holding the workspace's admin role, which grants
// every workspace permission.
var workspaceAdmin = model.User{ID: "admin", Role: model.SystemUserRoleId}

func workspaceAdminAccess() fakeAccessWorkspaceStore {
	role := *model.MakeDefaultProjectWorkspaceRoles()[model.ProjectWorkspaceAdminRoleID]
	return fakeAccessWorkspaceStore{memberRoleName: role.Name, roles: []model.ProjectWorkspaceRole{role}, onlyMember: workspaceAdmin.ID}
}

func accessApp(ws *fakeAccessWorkspaceStore) *App {
	return &App{
		Store: store.Store{Workspace: ws},
		Server: Server{
			License: &model.License{
				StartsAt:  time.Now().Add(-time.Hour).Unix(),
				ExpiresAt: time.Now().Add(time.Hour).Unix(),
				Features:  &model.Features{WorkspaceRoles: boolPtr(true)},
			},
		},
	}
}

const (
	wsID = "workspace-1"
	tbl1 = "table-1"
	tbl2 = "table-2"
)

func accessUser() model.User { return model.User{ID: "user-1", Role: "user"} }

func workspaceRole(name string) model.ProjectWorkspaceRole {
	return model.ProjectWorkspaceRole{
		ID: name + "-id", Name: name,
		PerTableMode: false,
		Permissions:  []string{"create_task", "update_task"},
	}
}

func perTableRole(name string) model.ProjectWorkspaceRole {
	return model.ProjectWorkspaceRole{
		ID: name + "-id", Name: name,
		PerTableMode: true,
		Permissions:  []string{},
	}
}

func twoTables() []model.WorkspaceTable {
	return []model.WorkspaceTable{{ID: tbl1}, {ID: tbl2}}
}

// One fixture with every restriction the enterprise resolver would honour:
// a per-table role, a hidden table, and an assigned-only table. Without the
// enterprise binary all of it must be ignored.
func restrictedApp() *App {
	return accessApp(&fakeAccessWorkspaceStore{
		memberRoleName: "user restricted",
		roles: []model.ProjectWorkspaceRole{
			workspaceRole("user"),
			perTableRole("restricted"),
		},
		tablePerms: []model.TablePermission{
			{TableID: tbl1, Action: "show_assigned_tasks_only"},
			{TableID: tbl2, Action: "hidden"},
		},
	})
}

func TestOSSAccessIsAlwaysPermissive(t *testing.T) {
	a := restrictedApp()
	user := accessUser()

	in := a.buildAccessInput(user, wsID)

	if got := a.filterTablesByPerTableModeFromInput(in, twoTables()); len(got) != 2 {
		t.Errorf("tables: want all 2 visible, got %d", len(got))
	}

	if assigned := a.buildAssignedOnlyMapFromInput(user, wsID, in); assigned["*"] || assigned[tbl1] {
		t.Error("assigned-only: want no restriction")
	}

	acc := a.resolveAccess(in)
	if !acc.CanPerformRowAction(tbl1, "create_task", "create_task") {
		t.Error("row actions: want allowed")
	}
	if !acc.TableHasPermission(tbl1, "manage_fields") {
		t.Error("table permissions: want allowed")
	}

	va := a.buildViewAccessStateFromInput(user, in)
	if va.AssignedOnly(tbl1) {
		t.Error("view access: want no assigned-only restriction")
	}
	if got := va.VisibleTables(twoTables()); len(got) != 2 {
		t.Errorf("view access tables: want all 2 visible, got %d", len(got))
	}
}

func TestPrivateViewsShowOnlyToTheirCreatorAndThoseSharedWith(t *testing.T) {
	shared := map[string]bool{"v-shared": true}

	for _, tc := range []struct {
		name string
		view model.WorkspaceView
		want bool
	}{
		{"public view", model.WorkspaceView{ID: "v-public", IsPublic: true, CreatedBy: "someone"}, true},
		{"own private view", model.WorkspaceView{ID: "v-mine", CreatedBy: "me"}, true},
		{"private view shared with the user", model.WorkspaceView{ID: "v-shared", CreatedBy: "someone"}, true},
		{"someone else's private view", model.WorkspaceView{ID: "v-other", CreatedBy: "someone"}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := viewVisibleTo(tc.view, "me", shared); got != tc.want {
				t.Errorf("viewVisibleTo = %v, want %v", got, tc.want)
			}
		})
	}
}

type fakeViewAccessStore struct {
	store.WorkspaceStore
	public  bool
	missing bool
	shared  []string
}

func (f fakeViewAccessStore) GetViewVisibility(_ context.Context, _ string) (bool, string, error) {
	if f.missing {
		return false, "", sql.ErrNoRows
	}
	return f.public, "creator", nil
}

func (f fakeViewAccessStore) GetViewSharedUserIDs(_ string) ([]string, error) {
	return f.shared, nil
}

func TestAPrivateViewOpensOnlyForItsCreatorAndThoseSharedWith(t *testing.T) {
	for _, tc := range []struct {
		name  string
		store fakeViewAccessStore
		user  string
		want  int
	}{
		{"public view", fakeViewAccessStore{public: true}, "someone", 0},
		{"own private view", fakeViewAccessStore{}, "creator", 0},
		{"private view shared with the user", fakeViewAccessStore{shared: []string{"someone"}}, "someone", 0},
		{"someone else's private view", fakeViewAccessStore{shared: []string{"another"}}, "someone", http.StatusNotFound},
		{"view that does not exist", fakeViewAccessStore{missing: true}, "someone", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a := &App{Store: store.Store{Workspace: tc.store}}

			got := 0
			if appErr := a.requireViewAccess(context.Background(), tc.user, "v"); appErr != nil {
				got = appErr.Status
			}
			if got != tc.want {
				t.Errorf("status = %d, want %d", got, tc.want)
			}
		})
	}
}
