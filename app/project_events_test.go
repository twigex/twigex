// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"context"
	"database/sql"
	"encoding/json"
	"slices"
	"testing"

	"github.com/twigex/twigex/interfaces"
	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/store"
)

type fakeChangeUserStore struct {
	store.UserStore
	users map[string]model.User
}

func (f fakeChangeUserStore) GetByIDs(ids []string) ([]model.User, error) {
	out := make([]model.User, 0, len(ids))
	for _, id := range ids {
		if u, ok := f.users[id]; ok {
			out = append(out, u)
		}
	}
	return out, nil
}

type fakeChangeWorkspaceStore struct {
	fakeAccessWorkspaceStore
	members map[string]bool
}

func (f *fakeChangeWorkspaceStore) IsMember(_, userID string) (bool, error) {
	return f.members[userID], nil
}

func (f *fakeChangeWorkspaceStore) GetMemberUserIDs(_ context.Context, _ string, ids []string) ([]string, error) {
	var out []string
	for _, id := range ids {
		if f.members[id] {
			out = append(out, id)
		}
	}
	return out, nil
}

// changeRoles hides tables from, or makes them assigned-only for, the users
// named in each map.
type changeRoles struct {
	hidden       map[string]bool
	assignedOnly map[string]bool
}

func (r changeRoles) Resolve(in interfaces.AccessInput) interfaces.WorkspaceAccess {
	return changeAccess{roles: r, user: in.Roles}
}

type changeAccess struct {
	openAccess
	roles changeRoles
	user  []model.ProjectWorkspaceRole
}

func (a changeAccess) name() string {
	if len(a.user) == 0 {
		return ""
	}
	return a.user[0].DisplayName
}

func (a changeAccess) VisibleTables(tables []model.WorkspaceTable) []model.WorkspaceTable {
	if a.roles.hidden[a.name()] {
		return nil
	}
	return tables
}

func (a changeAccess) AssignedOnly(string) bool { return a.roles.assignedOnly[a.name()] }

// perUserRoleStore gives each user a role whose display name is their id, so
// changeAccess can tell them apart.
type perUserRoleStore struct {
	*fakeChangeWorkspaceStore
}

func (s perUserRoleStore) GetUserByUserID(userID, _ string) (*model.WorkspaceMember, error) {
	return &model.WorkspaceMember{Role: "role-" + userID}, nil
}

func (s perUserRoleStore) GetRoleNamesForUsers(_ context.Context, _ string, userIDs []string) (map[string][]string, error) {
	out := make(map[string][]string, len(userIDs))
	for _, id := range userIDs {
		out[id] = []string{"role-" + id}
	}
	return out, nil
}

func (s perUserRoleStore) GetRolesByName(names []string, _ string) ([]model.ProjectWorkspaceRole, error) {
	out := make([]model.ProjectWorkspaceRole, 0, len(names))
	for _, n := range names {
		role := model.ProjectWorkspaceRole{Name: n, DisplayName: n[len("role-"):]}
		if role.DisplayName == workspaceAdmin.ID {
			role.Permissions = model.MakeDefaultProjectWorkspaceRoles()[model.ProjectWorkspaceAdminRoleID].Permissions
		}
		out = append(out, role)
	}
	return out, nil
}

func changeApp(connected []string, roles changeRoles) (*App, map[string]*Client) {
	users := map[string]model.User{
		"admin":    workspaceAdmin,
		"member":   {ID: "member", Role: model.SystemUserRoleId},
		"hidden":   {ID: "hidden", Role: model.SystemUserRoleId},
		"assigned": {ID: "assigned", Role: model.SystemUserRoleId},
		"outsider": {ID: "outsider", Role: model.SystemUserRoleId},
	}
	ws := &fakeChangeWorkspaceStore{members: map[string]bool{"admin": true, "member": true, "hidden": true, "assigned": true}}
	hub := &Hub{clients: map[*Client]bool{}}
	clients := map[string]*Client{}
	for _, id := range connected {
		c := &Client{hub: hub, user: id, send: make(chan []byte, 4)}
		hub.clients[c] = true
		clients[id] = c
	}
	a := readAccessApp(&fakeReadAccessStore{})
	a.Store = store.Store{Workspace: perUserRoleStore{ws}, User: fakeChangeUserStore{users: users}}
	a.Server.NotificationHub = hub
	a.WorkspaceRoles = roles
	return a, clients
}

func received(c *Client) map[string]any {
	select {
	case msg := <-c.send:
		var event model.WebsocketEvent
		_ = json.Unmarshal(msg, &event)
		return event.Data
	default:
		return nil
	}
}

type fakeTaskRowStore struct {
	perUserRoleStore
	singleSelect map[string]bool
	rows         map[string]map[string]interface{}
}

func (f fakeTaskRowStore) GetTableMeta(tableID string) (string, bool, error) {
	return "tbl_" + tableID, f.singleSelect[tableID], nil
}

func (f fakeTaskRowStore) GetTableRowByID(_, itemID string) (map[string]interface{}, error) {
	return f.rows[itemID], nil
}

func (f fakeTaskRowStore) GetTableColumnTypes(string) ([]*sql.ColumnType, error) {
	return nil, nil
}

func TestPublishTaskRowSendsTasksAndOptionsToTheRightPeople(t *testing.T) {
	everyone := []string{"member", "assigned"}
	roles := changeRoles{assignedOnly: map[string]bool{"assigned": true}}

	for _, tc := range []struct {
		name         string
		table        string
		singleSelect bool
		wantAssigned bool
		wantKey      string
	}{
		{"task in a task table", "t-tasks", false, false, "task"},
		{"option in a status table", "t-status", true, true, "row"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a, clients := changeApp(everyone, roles)
			ws := a.Store.Workspace.(perUserRoleStore)
			a.Store.Workspace = fakeTaskRowStore{
				perUserRoleStore: ws,
				singleSelect:     map[string]bool{tc.table: tc.singleSelect},
				rows:             map[string]map[string]interface{}{"row-1": {"id": "row-1", "assignee": "member", "name": "Row"}},
			}
			a.publishTaskRow(context.Background(), "task_updated", "ws-a", tc.table, "row-1")

			data := received(clients["member"])
			if data == nil {
				t.Fatal("member received nothing")
			}
			if row, _ := data[tc.wantKey].(map[string]any); row["id"] != "row-1" {
				t.Errorf("member data = %v, want the row under %q", data, tc.wantKey)
			}
			if got := received(clients["assigned"]) != nil; got != tc.wantAssigned {
				t.Errorf("assigned-only user received = %v, want %v", got, tc.wantAssigned)
			}
		})
	}
}

func TestReassigningATaskMovesItBetweenAssignedOnlyUsers(t *testing.T) {
	roles := changeRoles{assignedOnly: map[string]bool{"assigned": true, "hidden": true}}
	a, clients := changeApp([]string{"member", "assigned", "hidden"}, roles)
	a.Store.Workspace = fakeTaskRowStore{
		perUserRoleStore: a.Store.Workspace.(perUserRoleStore),
		rows:             map[string]map[string]interface{}{"row-1": {"id": "row-1", "assignee": "hidden"}},
	}

	a.publishTaskAssignment(context.Background(), "ws-a", "t-tasks", "row-1", "assigned")

	kinds := func(c *Client) []any {
		var out []any
		for data := received(c); data != nil; data = received(c) {
			out = append(out, data["type"])
			if data["type"] == "task_unassigned" && data["task_id"] != "row-1" {
				t.Errorf("task_unassigned = %v, want task_id row-1", data)
			}
		}
		return out
	}
	for user, want := range map[string][]any{
		"member":   {"task_updated"},
		"assigned": {"task_unassigned"},
		"hidden":   {"task_updated", "task_created"},
	} {
		if got := kinds(clients[user]); !slices.Equal(got, want) {
			t.Errorf("%s received %v, want %v", user, got, want)
		}
	}
}

type fakeFieldDeleteStore struct {
	perUserRoleStore
}

func (fakeFieldDeleteStore) GetFieldParentInfo(string) (string, string, error) {
	return "field-in-other-table", "t-other", nil
}

func (fakeFieldDeleteStore) DeleteTableField(_, _, _, _ string) error { return nil }

func TestDeletingAFieldPublishesItsIDNotAnyRows(t *testing.T) {
	a, clients := changeApp([]string{"member"}, changeRoles{})
	a.Store.Workspace = fakeFieldDeleteStore{a.Store.Workspace.(perUserRoleStore)}
	admin := workspaceAdmin

	if _, appErr := a.DeleteWorkspaceTableField(context.Background(), "ws-a", "t-a", "field-1", true, "", admin); appErr != nil {
		t.Fatalf("DeleteWorkspaceTableField: %v", appErr)
	}
	data := received(clients["member"])
	linked, _ := data["linked_data"].(map[string]any)
	if data["type"] != "DELETE_FIELD" || data["field_id"] != "field-1" || linked["parent_field_id"] != "field-in-other-table" || linked["table_id"] != "t-other" {
		t.Errorf("published %v", data)
	}
	if _, carriesRows := data["data"]; carriesRows {
		t.Errorf("the change carries data it should not: %v", data["data"])
	}
}

type fakeFormulaStore struct{ perUserRoleStore }

func (fakeFormulaStore) UpdateFieldFormula(context.Context, string, string, string, *model.FormulaSpec) error {
	return nil
}

func TestChangingAFormulaReachesThoseWhoSeeTheTable(t *testing.T) {
	a, clients := changeApp([]string{"member", "hidden"}, changeRoles{hidden: map[string]bool{"hidden": true}})
	a.Store.Workspace = fakeFormulaStore{a.Store.Workspace.(perUserRoleStore)}
	admin := workspaceAdmin

	formula := &model.FormulaSpec{Name: "MAX"}
	if appErr := a.UpdateCalculationsField(context.Background(), "ws-a", "t-a", "c1", formula, admin); appErr != nil {
		t.Fatalf("UpdateCalculationsField: %v", appErr)
	}

	data := received(clients["member"])
	if data["type"] != "field_formula_updated" || data["field_name"] != "c1" || data["formula"] == nil {
		t.Errorf("published %v", data)
	}

	if leaked := received(clients["hidden"]); leaked != nil {
		t.Errorf("someone who cannot see the table was sent %v", leaked)
	}
}

type fakeViewVisibilityStore struct {
	perUserRoleStore
	public map[string]bool
	shared map[string][]string
}

func (f fakeViewVisibilityStore) GetViewVisibility(_ context.Context, viewID string) (bool, string, error) {
	return f.public[viewID], "member", nil
}

func (f fakeViewVisibilityStore) GetViewSharedUserIDs(viewID string) ([]string, error) {
	return f.shared[viewID], nil
}

func (f fakeViewVisibilityStore) GetView(_ context.Context, viewID string) (*model.WorkspaceView, error) {
	return &model.WorkspaceView{ID: viewID, ViewType: "kanban", CreatedBy: "member"}, nil
}

func TestChangingWhoSeesAViewAddsOrRemovesItForThem(t *testing.T) {
	everyone := []string{"member", "assigned", "hidden"}
	ownerOnly := map[string]bool{"member": true}

	for _, tc := range []struct {
		name   string
		public bool
		shared []string
		before map[string]bool
		want   map[string]any
	}{
		{"shared with another member", false, []string{"assigned"}, ownerOnly, map[string]any{"member": "VIEW_VISIBILITY", "assigned": "VIEW_CREATED"}},
		{"share removed", false, nil, map[string]bool{"member": true, "assigned": true}, map[string]any{"member": "VIEW_VISIBILITY", "assigned": "DELETE_VIEW"}},
		{"made public", true, nil, ownerOnly, map[string]any{"member": "VIEW_VISIBILITY", "assigned": "VIEW_CREATED", "hidden": "VIEW_CREATED"}},
		{"made private", false, nil, nil, map[string]any{"member": "VIEW_VISIBILITY", "assigned": "DELETE_VIEW", "hidden": "DELETE_VIEW"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a, clients := changeApp(everyone, changeRoles{})
			a.Store.Workspace = fakeViewVisibilityStore{a.Store.Workspace.(perUserRoleStore),
				map[string]bool{"v": tc.public}, map[string][]string{"v": tc.shared}}

			a.publishViewAccessChange(context.Background(), "ws-a", "t-a", "v", tc.before)

			for _, id := range everyone {
				data := received(clients[id])
				if got := data["type"]; got != tc.want[id] {
					t.Errorf("%s received %v, want %v", id, got, tc.want[id])
				}
				if data["type"] == "VIEW_VISIBILITY" {
					if data["view_id"] != "v" || data["is_public"] != tc.public {
						t.Errorf("%s got %v, want view v public=%v", id, data, tc.public)
					}
					continue
				}
				if data != nil && data["delete_id"] != "v" {
					if view, _ := data["data"].(map[string]any); view["id"] != "v" {
						t.Errorf("%s got %v, want view v", id, data)
					}
				}
			}
		})
	}
}

func TestPrivateViewsAndFiltersReachOnlyThoseWhoMaySeeThem(t *testing.T) {
	everyone := []string{"member", "assigned"}

	for _, tc := range []struct {
		name    string
		publish func(a *App)
		shared  bool
	}{
		{"private saved filter", func(a *App) {
			a.publishSavedFilter(context.Background(), "FILTER_CREATED", "ws-a", "t-a", "v-private", "s1", "f1", "Mine", model.FilterPayload{}, true, false, "member")
		}, false},
		{"public saved filter", func(a *App) {
			a.publishSavedFilter(context.Background(), "FILTER_CREATED", "ws-a", "t-a", "v-public", "s1", "f1", "Ours", model.FilterPayload{}, false, false, "member")
		}, true},
		{"private view", func(a *App) {
			a.publishKanbanColumns(context.Background(), "ws-a", "t-a", "v-private", "{}")
		}, false},
		{"public view", func(a *App) {
			a.publishKanbanColumns(context.Background(), "ws-a", "t-a", "v-public", "{}")
		}, true},
		{"private view shared with the other member", func(a *App) {
			a.publishKanbanColumns(context.Background(), "ws-a", "t-a", "v-shared", "{}")
		}, true},
		{"public filter on a private view", func(a *App) {
			a.publishSavedFilter(context.Background(), "FILTER_CREATED", "ws-a", "t-a", "v-private", "s1", "f1", "Ours", model.FilterPayload{}, false, false, "member")
		}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a, clients := changeApp(everyone, changeRoles{})
			a.Store.Workspace = fakeViewVisibilityStore{a.Store.Workspace.(perUserRoleStore), map[string]bool{"v-public": true},
				map[string][]string{"v-shared": {"assigned"}}}
			tc.publish(a)

			if received(clients["member"]) == nil {
				t.Error("the owner received nothing")
			}
			if got := received(clients["assigned"]) != nil; got != tc.shared {
				t.Errorf("another member received = %v, want %v", got, tc.shared)
			}
		})
	}
}

type fakeRenameStore struct {
	perUserRoleStore
}

func (fakeRenameStore) UpdateTable(_, _, name string) (*string, error) { return &name, nil }

func (fakeRenameStore) UpdateFolder(_, _, _ string) (bool, error) { return true, nil }

func TestTableChangesFollowTableVisibilityAndFolderChangesDoNot(t *testing.T) {
	everyone := []string{"member", "hidden"}
	roles := changeRoles{hidden: map[string]bool{"hidden": true}}
	admin := workspaceAdmin

	for _, tc := range []struct {
		name       string
		rename     func(a *App) *model.AppError
		kind       string
		wantHidden bool
	}{
		{"table rename", func(a *App) *model.AppError {
			_, err := a.UpdateWorkspaceTable(context.Background(), "ws-a", "t-a", "Renamed", admin)
			return err
		}, "UPDATE_TABLE_NAME", false},
		{"folder rename", func(a *App) *model.AppError {
			_, err := a.UpdateWorkspaceFolder(context.Background(), "ws-a", "folder-1", "Renamed", admin)
			return err
		}, "UPDATE_FOLDER_NAME", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a, clients := changeApp(everyone, roles)
			a.Store.Workspace = fakeRenameStore{a.Store.Workspace.(perUserRoleStore)}
			if appErr := tc.rename(a); appErr != nil {
				t.Fatalf("rename: %v", appErr)
			}

			data := received(clients["member"])
			renamed, _ := data["data"].(map[string]any)
			if data["type"] != tc.kind || renamed["name"] != "Renamed" {
				t.Errorf("member got %v", data)
			}
			if got := received(clients["hidden"]) != nil; got != tc.wantHidden {
				t.Errorf("user who cannot see the table received = %v, want %v", got, tc.wantHidden)
			}
		})
	}
}

func TestProjectChangesReachOnlyThoseWhoMaySeeThem(t *testing.T) {
	everyone := []string{"admin", "member", "hidden", "assigned", "outsider"}
	roles := changeRoles{hidden: map[string]bool{"hidden": true}, assignedOnly: map[string]bool{"assigned": true}}

	for _, tc := range []struct {
		name string
		task map[string]any
		want map[string]bool
	}{
		{"change to someone else's task", map[string]any{"id": "t1", "assignee": "member"}, map[string]bool{"admin": true, "member": true}},
		{"change to the assigned-only user's task", map[string]any{"id": "t1", "assignee": "assigned"}, map[string]bool{"admin": true, "member": true, "assigned": true}},
		{"change that is not about one task", nil, map[string]bool{"admin": true, "member": true, "assigned": true}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a, clients := changeApp(everyone, roles)
			ctx := WithOriginClient(context.Background(), "tab-1")
			a.publishProjectChange(ctx, projectChange{kind: "task_updated", workspaceID: "ws-a", tableID: "t-a", task: tc.task})

			for _, id := range everyone {
				data := received(clients[id])
				if got := data != nil; got != tc.want[id] {
					t.Errorf("%s received = %v, want %v", id, got, tc.want[id])
					continue
				}
				if data != nil && (data["type"] != "task_updated" || data["workspace_id"] != "ws-a" || data["table_id"] != "t-a" || data["origin_client_id"] != "tab-1") {
					t.Errorf("%s got %v", id, data)
				}
			}
		})
	}
}

type fakeLinkOwnersStore struct {
	perUserRoleStore
}

func (fakeLinkOwnersStore) GetTableName(tableID string) (string, error) {
	return "tbl_" + tableID, nil
}

func (fakeLinkOwnersStore) GetTaskAssignees(_ context.Context, _ string, ids []string) (map[string]string, error) {
	owners := map[string]string{}
	for _, id := range ids {
		if id == "theirs" {
			owners[id] = "assigned"
		} else {
			owners[id] = "member"
		}
	}
	return owners, nil
}

func TestATaskChangeHidesLinksAUserMayNotSee(t *testing.T) {
	a, clients := changeApp([]string{"member", "assigned"}, changeRoles{assignedOnly: map[string]bool{"assigned": true}})
	a.Store.Workspace = fakeLinkOwnersStore{a.Store.Workspace.(perUserRoleStore)}

	a.publishProjectChange(context.Background(), projectChange{
		kind:        "task_updated",
		workspaceID: "ws-a",
		tableID:     "t-tasks",
		task: map[string]any{
			"id":       "task-1",
			"assignee": "assigned",
			"links":    []model.LinkedItem{{ID: "theirs", Name: "Theirs"}, {ID: "others", Name: "Someone else's"}},
		},
		headers: []model.WorkspaceHeaders{{Name: "links", LinkedID: "j", ParentTableID: "t-linked"}},
	})

	names := func(data map[string]any) []any {
		task, _ := data["task"].(map[string]any)
		var out []any
		for _, link := range task["links"].([]any) {
			out = append(out, link.(map[string]any)["name"])
		}
		return out
	}
	if got := names(received(clients["member"])); !slices.Equal(got, []any{"Theirs", "Someone else's"}) {
		t.Errorf("member saw links %v, want both names", got)
	}
	if got := names(received(clients["assigned"])); !slices.Equal(got, []any{"Theirs", ""}) {
		t.Errorf("assigned-only user saw links %v, want only their own task named", got)
	}
}
