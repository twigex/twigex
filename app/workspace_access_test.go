// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

// Access without the enterprise module: the built-in roles are enforced, and
// per-table roles, which only the enterprise module can apply, give nothing.

import (
	"context"
	"database/sql"
	"net/http"
	"testing"
	"time"

	"github.com/twigex/twigex/interfaces"
	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/store"
)

type openAccess struct{}

func (openAccess) VisibleTables(tables []model.WorkspaceTable) []model.WorkspaceTable { return tables }
func (openAccess) AssignedOnlyAll() bool                                              { return false }
func (openAccess) AssignedOnly(_ string) bool                                         { return false }
func (openAccess) CanPerformRowAction(_, _, _ string) bool                            { return true }
func (openAccess) TableHasPermission(_, _ string) bool                                { return true }

type enterpriseRoles struct{}

func (enterpriseRoles) Resolve(interfaces.AccessInput) interfaces.WorkspaceAccess {
	return openAccess{}
}

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

func (f *fakeAccessWorkspaceStore) GetOptionTableField(context.Context, string) (string, string, error) {
	return "", "", nil
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

// A workspace-level role next to a per-table role with a hidden table and an
// assigned-only table.
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

func builtInRoleApp(roleName string) *App {
	role := *model.MakeDefaultProjectWorkspaceRoles()[roleName]

	return accessApp(&fakeAccessWorkspaceStore{
		memberRoleName: role.Name,
		roles:          []model.ProjectWorkspaceRole{role},
	})
}

func TestWithoutEnterpriseTheUserRoleCannotDeleteTasks(t *testing.T) {
	a := builtInRoleApp(model.ProjectWorkspaceUserRoleID)
	user := accessUser()

	if !a.CanPerformRowAction(user, wsID, tbl1, model.PermissionCreateTask, model.PermissionCreateTask) {
		t.Error("create task: want allowed for the user role")
	}

	if !a.CanPerformRowAction(user, wsID, tbl1, model.PermissionUpdateTask, model.PermissionUpdateTask) {
		t.Error("update task: want allowed for the user role")
	}

	if a.CanPerformRowAction(user, wsID, tbl1, model.PermissionDeleteTask, model.PermissionDeleteTask) {
		t.Error("delete task: want refused for the user role")
	}

	in := a.buildAccessInput(user, wsID)

	if got := a.filterTablesByPerTableModeFromInput(in, twoTables()); len(got) != 2 {
		t.Errorf("tables: want all 2 visible, got %d", len(got))
	}

	if assigned := a.buildAssignedOnlyMapFromInput(user, wsID, in); assigned["*"] || assigned[tbl1] {
		t.Error("assigned-only: want no restriction")
	}
}

func TestWithoutEnterpriseTheAdminRoleCanDeleteTasks(t *testing.T) {
	a := builtInRoleApp(model.ProjectWorkspaceAdminRoleID)

	if !a.CanPerformRowAction(accessUser(), wsID, tbl1, model.PermissionDeleteTask, model.PermissionDeleteTask) {
		t.Error("delete task: want allowed for the admin role")
	}
}

func perTableOnlyApp() *App {
	return accessApp(&fakeAccessWorkspaceStore{
		memberRoleName: "restricted",
		roles:          []model.ProjectWorkspaceRole{perTableRole("restricted")},
		tablePerms: []model.TablePermission{
			{TableID: tbl1, Action: "create_task"},
			{TableID: tbl1, Action: "view"},
		},
	})
}

func TestWithoutEnterpriseOnlyAPerTableRoleGivesNothing(t *testing.T) {
	a := perTableOnlyApp()
	user := accessUser()

	if a.CanPerformRowAction(user, wsID, tbl1, model.PermissionCreateTask, model.PermissionCreateTask) {
		t.Error("row actions: want refused")
	}

	if a.ProjectWorkspaceTableHasPermission(user, wsID, tbl1, "manage_views") {
		t.Error("table permissions: want refused")
	}

	in := a.buildAccessInput(user, wsID)

	if got := a.filterTablesByPerTableModeFromInput(in, twoTables()); len(got) != 0 {
		t.Errorf("tables: want none visible, got %d", len(got))
	}
}

func TestWithoutEnterpriseAWorkspaceRoleNextToAPerTableRoleStillApplies(t *testing.T) {
	a := restrictedApp()
	user := accessUser()

	if !a.CanPerformRowAction(user, wsID, tbl2, model.PermissionCreateTask, model.PermissionCreateTask) {
		t.Error("create task: want allowed by the workspace-level role")
	}

	if a.CanPerformRowAction(user, wsID, tbl2, model.PermissionDeleteTask, model.PermissionDeleteTask) {
		t.Error("delete task: want refused, no role allows it")
	}

	in := a.buildAccessInput(user, wsID)

	if got := a.filterTablesByPerTableModeFromInput(in, twoTables()); len(got) != 2 {
		t.Errorf("tables: want both visible, got %d", len(got))
	}

	access := a.resolveAccess(in)
	if !access.AssignedOnly(tbl1) || access.AssignedOnly(tbl2) {
		t.Errorf("assigned-only: want only %s restricted", tbl1)
	}
}

func TestWithoutEnterpriseShowAssignedTasksOnlyStillApplies(t *testing.T) {
	role := workspaceRole("limited")
	role.Permissions = append(role.Permissions, model.PermissionShowAssignedTasksOnly)

	a := accessApp(&fakeAccessWorkspaceStore{
		memberRoleName: role.Name,
		roles:          []model.ProjectWorkspaceRole{role},
	})
	user := accessUser()

	in := a.buildAccessInput(user, wsID)

	if assigned := a.buildAssignedOnlyMapFromInput(user, wsID, in); !assigned["*"] {
		t.Error("assigned-only: want every table restricted")
	}
}

func TestWithoutEnterpriseANonMemberIsRefused(t *testing.T) {
	a := accessApp(&fakeAccessWorkspaceStore{})

	if a.CanPerformRowAction(accessUser(), wsID, tbl1, model.PermissionCreateTask, model.PermissionCreateTask) {
		t.Error("row actions: want refused for a non-member")
	}
}

func TestManagingRolesNeedsTheLicenceAndTheEnterpriseModule(t *testing.T) {
	a := accessApp(&fakeAccessWorkspaceStore{})

	if appErr := a.RequireWorkspaceRoles(); appErr == nil || appErr.Status != http.StatusNotImplemented {
		t.Errorf("licence without the module: want 501, got %v", appErr)
	}

	a.WorkspaceRoles = enterpriseRoles{}

	if appErr := a.RequireWorkspaceRoles(); appErr != nil {
		t.Errorf("licence and module: want allowed, got %v", appErr)
	}

	a.Server.License = nil

	if appErr := a.RequireWorkspaceRoles(); appErr == nil || appErr.Status != http.StatusPaymentRequired {
		t.Errorf("module without a licence: want 402, got %v", appErr)
	}
}

func TestWithoutEnterpriseRolesCannotBeCreatedOrEdited(t *testing.T) {
	a := accessApp(&fakeAccessWorkspaceStore{})
	user := accessUser()

	if _, appErr := a.CreateProjectWorkspaceRole(user, wsID, model.ProjectWorkspaceRole{Name: "editor"}); appErr == nil || appErr.Status != http.StatusNotImplemented {
		t.Errorf("create: want 501, got %v", appErr)
	}

	if _, appErr := a.UpdateProjectWorkspaceRole(user, wsID, "editor-id", model.ProjectWorkspaceRolePatch{}); appErr == nil || appErr.Status != http.StatusNotImplemented {
		t.Errorf("update: want 501, got %v", appErr)
	}

	if appErr := a.UpdateRoleTablePermissions(user, wsID, "editor-id", nil); appErr == nil || appErr.Status != http.StatusNotImplemented {
		t.Errorf("table permissions: want 501, got %v", appErr)
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

func TestTheLicenceMapOffersWorkspaceRolesOnlyWithTheEnterpriseModule(t *testing.T) {
	a := accessApp(&fakeAccessWorkspaceStore{})

	if got := (*a.GetLicenseMap(false))["workspace_roles"]; got != false {
		t.Errorf("licence without the module: want false, got %v", got)
	}

	a.WorkspaceRoles = enterpriseRoles{}

	if got := (*a.GetLicenseMap(false))["workspace_roles"]; got != true {
		t.Errorf("licence and module: want true, got %v", got)
	}
}
