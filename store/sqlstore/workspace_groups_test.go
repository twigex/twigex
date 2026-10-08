// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package sqlstore

// Tests for workspace-group attachment: add, list, remove, update roles, member
// resolution. Behavior lives in SQL (ON DUPLICATE KEY, JOINs, hard-delete), so
// these run against real MySQL. See testhelper_test.go for harness details.
// seedUser is shared with collimato_groups_test.go (same package).

import (
	"context"
	"database/sql"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/twigex/twigex/model"
)

func resetWorkspaceGroupTables(t *testing.T) {
	t.Helper()
	cleanTables(t,
		"workspace_groups",
		"group_members",
		"user_groups",
		"workspaces",
		"users",
	)
}

func seedWorkspace(t *testing.T, db *sql.DB) string {
	t.Helper()
	id := model.NewID()
	now := time.Now().Unix()
	_, err := db.Exec(
		`INSERT INTO workspaces (id, title, description, start_date, end_date, pre_fix, created_at, updated_at, deleted_at, created_by)
		 VALUES (?, ?, '', 0, 0, '', ?, ?, 0, ?)`,
		id, "ws-"+id[:8], now, now, model.NewID(),
	)
	if err != nil {
		t.Fatalf("seedWorkspace: %v", err)
	}
	return id
}

func seedUserGroup(t *testing.T, db *sql.DB) string {
	t.Helper()
	id := model.NewID()
	now := time.Now().Unix()
	_, err := db.Exec(
		`INSERT INTO user_groups (id, name, description, owner_id, created_at, deleted_at)
		 VALUES (?, ?, '', ?, ?, 0)`,
		id, "group-"+id[:8], model.NewID(), now,
	)
	if err != nil {
		t.Fatalf("seedUserGroup: %v", err)
	}
	return id
}

func TestGetMemberUserIDs(t *testing.T) {
	db := requireDB(t)
	resetWorkspaceGroupTables(t)
	cleanTables(t, "workspace_members")
	repo := &workspaceRepository{Db: db}
	ctx := context.Background()

	wsID := seedWorkspace(t, db)
	otherWS := seedWorkspace(t, db)
	group := seedUserGroup(t, db)
	deletedGroup := seedUserGroup(t, db)
	mustExec(t, "UPDATE user_groups SET deleted_at = 1 WHERE id = ?", deletedGroup)
	if err := repo.AddGroups(ctx, wsID, []string{group, deletedGroup}, []string{"member"}, "adder"); err != nil {
		t.Fatalf("AddGroups: %v", err)
	}

	mustExec(t, `INSERT INTO workspace_members (id, user_id, workspace_id, role, date_joined) VALUES
		('m1', 'direct', ?, 'member', 1), ('m2', 'elsewhere', ?, 'member', 1), ('m3', 'both', ?, 'member', 1)`, wsID, otherWS, wsID)
	mustExec(t, `INSERT INTO group_members (id, group_id, user_id, role, joined_at) VALUES
		('g1', ?, 'via-group', 'member', 1), ('g2', ?, 'both', 'member', 1), ('g3', ?, 'via-deleted-group', 'member', 1)`, group, group, deletedGroup)

	got, err := repo.GetMemberUserIDs(ctx, wsID, []string{"direct", "via-group", "both", "elsewhere", "via-deleted-group", "stranger"})
	if err != nil {
		t.Fatalf("GetMemberUserIDs: %v", err)
	}

	sort.Strings(got)
	if want := []string{"both", "direct", "via-group"}; strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("members = %v, want %v", got, want)
	}
	if got, err := repo.GetMemberUserIDs(ctx, wsID, nil); err != nil || len(got) != 0 {
		t.Errorf("no candidates = %v, %v; want none", got, err)
	}
}

func TestBatchedRoleNamesCoverDirectAndGroupMembers(t *testing.T) {
	db := requireDB(t)
	resetWorkspaceGroupTables(t)
	cleanTables(t, "workspace_members")
	repo := &workspaceRepository{Db: db}
	ctx := context.Background()

	wsID := seedWorkspace(t, db)
	otherWS := seedWorkspace(t, db)
	group := seedUserGroup(t, db)
	deletedGroup := seedUserGroup(t, db)
	mustExec(t, "UPDATE user_groups SET deleted_at = 1 WHERE id = ?", deletedGroup)
	if err := repo.AddGroups(ctx, wsID, []string{group}, []string{"editor", "viewer"}, "adder"); err != nil {
		t.Fatalf("AddGroups: %v", err)
	}

	if err := repo.AddGroups(ctx, wsID, []string{deletedGroup}, []string{"owner"}, "adder"); err != nil {
		t.Fatalf("AddGroups: %v", err)
	}

	mustExec(t, `INSERT INTO workspace_members (id, user_id, workspace_id, role, date_joined) VALUES
		('m1', 'direct', ?, 'member viewer', 1), ('m2', 'both', ?, 'viewer', 1), ('m3', 'elsewhere', ?, 'owner', 1)`,
		wsID, wsID, otherWS)
	mustExec(t, `INSERT INTO group_members (id, group_id, user_id, role, joined_at) VALUES
		('g1', ?, 'via-group', 'member', 1), ('g2', ?, 'both', 'member', 1), ('g3', ?, 'direct', 'member', 1)`,
		group, group, deletedGroup)

	users := []string{"direct", "via-group", "both", "elsewhere", "stranger"}
	names, err := repo.GetRoleNamesForUsers(ctx, wsID, users)
	if err != nil {
		t.Fatalf("GetRoleNamesForUsers: %v", err)
	}

	want := map[string]string{"direct": "member,viewer", "via-group": "editor,viewer", "both": "viewer,editor"}
	for _, u := range users {
		if got := strings.Join(names[u], ","); got != want[u] {
			t.Errorf("roles of %s = %q, want %q", u, got, want[u])
		}
	}
}

func TestAddWorkspaceGroups_ListsBack(t *testing.T) {
	db := requireDB(t)
	resetWorkspaceGroupTables(t)

	repo := &workspaceRepository{Db: db}
	ctx := context.Background()

	wsID := seedWorkspace(t, db)
	gID := seedUserGroup(t, db)
	adder := model.NewID()

	if err := repo.AddGroups(ctx, wsID, []string{gID}, []string{"member", "editor"}, adder); err != nil {
		t.Fatalf("AddGroups: %v", err)
	}

	groups, err := repo.GetGroups(ctx, wsID)
	if err != nil {
		t.Fatalf("GetGroups: %v", err)
	}
	if len(groups) != 1 {
		t.Fatalf("want 1 group, got %d", len(groups))
	}
	if groups[0].GroupID != gID {
		t.Errorf("group_id = %s, want %s", groups[0].GroupID, gID)
	}

	got := groups[0].Roles
	sort.Strings(got)
	want := []string{"editor", "member"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("roles = %v, want %v", got, want)
	}
}

func TestAddWorkspaceGroups_DuplicateUpdatesRoles(t *testing.T) {
	db := requireDB(t)
	resetWorkspaceGroupTables(t)

	repo := &workspaceRepository{Db: db}
	ctx := context.Background()

	wsID := seedWorkspace(t, db)
	gID := seedUserGroup(t, db)
	adder := model.NewID()

	if err := repo.AddGroups(ctx, wsID, []string{gID}, []string{"member"}, adder); err != nil {
		t.Fatalf("first AddWorkspaceGroups: %v", err)
	}
	// Re-attach with different roles: must not create a duplicate row.
	if err := repo.AddGroups(ctx, wsID, []string{gID}, []string{"admin"}, adder); err != nil {
		t.Fatalf("second AddWorkspaceGroups: %v", err)
	}

	groups, err := repo.GetGroups(ctx, wsID)
	if err != nil {
		t.Fatalf("GetGroups: %v", err)
	}
	if len(groups) != 1 {
		t.Errorf("want exactly 1 row (no duplicate), got %d", len(groups))
	}
	if len(groups) > 0 && (len(groups[0].Roles) != 1 || groups[0].Roles[0] != "admin") {
		t.Errorf("roles after update = %v, want [admin]", groups[0].Roles)
	}
}

func TestRemoveWorkspaceGroup(t *testing.T) {
	db := requireDB(t)
	resetWorkspaceGroupTables(t)

	repo := &workspaceRepository{Db: db}
	ctx := context.Background()

	wsID := seedWorkspace(t, db)
	gID := seedUserGroup(t, db)

	if err := repo.AddGroups(ctx, wsID, []string{gID}, []string{"member"}, model.NewID()); err != nil {
		t.Fatalf("AddGroups: %v", err)
	}
	if err := repo.RemoveGroup(ctx, wsID, gID); err != nil {
		t.Fatalf("RemoveWorkspaceGroup: %v", err)
	}

	groups, err := repo.GetGroups(ctx, wsID)
	if err != nil {
		t.Fatalf("GetGroups: %v", err)
	}
	if len(groups) != 0 {
		t.Errorf("want 0 groups after removal, got %d", len(groups))
	}
}

func TestUpdateWorkspaceGroupRoles_ReplacesRoles(t *testing.T) {
	db := requireDB(t)
	resetWorkspaceGroupTables(t)

	repo := &workspaceRepository{Db: db}
	ctx := context.Background()

	wsID := seedWorkspace(t, db)
	gID := seedUserGroup(t, db)

	if err := repo.AddGroups(ctx, wsID, []string{gID}, []string{"member"}, model.NewID()); err != nil {
		t.Fatalf("AddGroups: %v", err)
	}
	if err := repo.UpdateGroupRoles(ctx, wsID, gID, []string{"admin", "editor"}); err != nil {
		t.Fatalf("UpdateWorkspaceGroupRoles: %v", err)
	}

	groups, err := repo.GetGroups(ctx, wsID)
	if err != nil {
		t.Fatalf("GetGroups: %v", err)
	}
	if len(groups) == 0 {
		t.Fatal("no groups returned")
	}
	got := groups[0].Roles
	sort.Strings(got)
	want := []string{"admin", "editor"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("roles = %v, want %v", got, want)
	}
}

func TestAddWorkspaceGroups_EmptyGroupIDs_NoOp(t *testing.T) {
	db := requireDB(t)
	resetWorkspaceGroupTables(t)

	repo := &workspaceRepository{Db: db}
	ctx := context.Background()

	wsID := seedWorkspace(t, db)

	// Empty slice: must not error and must not insert anything.
	if err := repo.AddGroups(ctx, wsID, []string{}, []string{"member"}, model.NewID()); err != nil {
		t.Fatalf("AddGroups with empty IDs: %v", err)
	}

	groups, err := repo.GetGroups(ctx, wsID)
	if err != nil {
		t.Fatalf("GetGroups: %v", err)
	}
	if len(groups) != 0 {
		t.Errorf("want 0 groups, got %d", len(groups))
	}
}

func TestSearchMembersCoversTheChosenWorkspacesOrEveryLiveOne(t *testing.T) {
	db := requireDB(t)
	resetWorkspaceGroupTables(t)
	cleanTables(t, "workspace_members")
	repo := &workspaceRepository{Db: db}
	ctx := context.Background()

	a := seedWorkspace(t, db)
	b := seedWorkspace(t, db)
	gone := seedWorkspace(t, db)
	mustExec(t, "UPDATE workspaces SET deleted_at = 1 WHERE id = ?", gone)

	anna := insertUser(t, db, userRow{name: "Anna", lastname: "Direct", email: "anna@example.com", username: "anna"})
	ben := insertUser(t, db, userRow{name: "Ben", lastname: "Grouped", email: "ben@example.com", username: "ben"})
	dora := insertUser(t, db, userRow{name: "Dora", lastname: "Gone", email: "dora@example.com", username: "dora"})
	insertUser(t, db, userRow{name: "Nora", lastname: "Nowhere", email: "nora@example.com", username: "nora"})
	olga := insertUser(t, db, userRow{name: "Olga", lastname: "Left", email: "olga@example.com", username: "olga", deactivatedAt: 1})

	mustExec(t, `INSERT INTO workspace_members (id, user_id, workspace_id, role, date_joined) VALUES
		('m1', ?, ?, 'member', 1), ('m2', ?, ?, 'member', 1), ('m3', ?, ?, 'member', 1), ('m4', ?, ?, 'member', 1)`,
		anna, a, dora, gone, olga, a, anna, b)
	group := seedUserGroup(t, db)
	if err := repo.AddGroups(ctx, b, []string{group}, []string{"member"}, "adder"); err != nil {
		t.Fatal(err)
	}

	seedGroupMember(t, db, group, ben)
	seedGroupMember(t, db, group, anna)

	for _, tc := range []struct {
		name       string
		workspaces []string
		query      string
		want       []string
	}{
		{"one workspace", []string{a}, "", []string{"Anna"}},
		{"through a group", []string{b}, "", []string{"Anna", "Ben"}},
		{"several", []string{a, b}, "", []string{"Anna", "Ben"}},
		{"every live workspace", nil, "", []string{"Anna", "Ben"}},
		{"searched", nil, "be", []string{"Ben"}},
	} {
		users, err := repo.SearchMembers(ctx, tc.workspaces, tc.query, 50, 0)
		if err != nil {
			t.Fatal(err)
		}

		var names []string
		for _, u := range users {
			names = append(names, u.Name)
		}
		if strings.Join(names, ",") != strings.Join(tc.want, ",") {
			t.Errorf("%s: members = %q, want %q", tc.name, names, tc.want)
		}
	}
}
