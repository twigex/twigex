// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package sqlstore

// Tests for Collimato workspace-group attachment and the effective-role
// resolver. Behavior lives in SQL (joins, soft-delete filters, ON DUPLICATE KEY
// UPDATE), so these run against real MySQL. See testhelper_test.go for setup.
// seedGroup / seedGroupMember are shared with channel_groups_helpers_test.go.

import (
	"context"
	"database/sql"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/twigex/twigex/model"
)

func resetCollimatoGroupTables(t *testing.T) {
	cleanTables(t,
		"collimato_workspace_groups",
		"collimato_workspace_users",
		"collimato_workspaces",
		"group_members",
		"user_groups",
		"users",
	)
}

// userRow is the minimal users row the suite seeds (the applied schema after
// migrations 38 and 59 has 17 columns, no `storage`). timezone is valid JSON so
// the scan's json.Unmarshal succeeds.
type userRow struct {
	name          string
	lastname      string
	email         string
	username      string
	authService   string
	authData      string
	deactivatedAt int64
}

func insertUser(t *testing.T, db *sql.DB, u userRow) string {
	t.Helper()
	id := model.NewID()
	now := time.Now().Unix()
	_, err := db.Exec(
		`INSERT INTO users
		 (id, email, role, password, auth_service, username, name, lastname,
		  storage_limit, timezone, mfa_active, mfa_secret, created_at, updated_at, deactivated_at, auth_data)
		 VALUES (?, ?, 'system_user', '', ?, ?, ?, ?, 0, 'null', 0, '', ?, ?, ?, ?)`,
		id, u.email, u.authService, u.username, u.name, u.lastname,
		now, now, u.deactivatedAt, u.authData,
	)
	if err != nil {
		t.Fatalf("insertUser %s: %v", u.username, err)
	}
	return id
}

func seedUser(t *testing.T, db *sql.DB) string {
	t.Helper()
	id := model.NewID()
	return insertUser(t, db, userRow{
		name:     "Test",
		lastname: "User",
		email:    id[:8] + "@example.com",
		username: "u-" + id[:8],
	})
}

func seedCollimatoWorkspace(t *testing.T, db *sql.DB) string {
	t.Helper()
	id := model.NewID()
	now := time.Now().Unix()
	_, err := db.Exec(
		`INSERT INTO collimato_workspaces
		 (id, name, description, status, server_id, secret, created_by, created_at, updated_at, deleted_at)
		 VALUES (?, ?, '', 'active', ?, '', ?, ?, ?, 0)`,
		id, "ws-"+id[:8], model.NewID(), model.NewID(), now, now,
	)
	if err != nil {
		t.Fatalf("seedCollimatoWorkspace: %v", err)
	}
	return id
}

func seedCollimatoWorkspaceUser(t *testing.T, db *sql.DB, workspaceID, userID, role string) {
	t.Helper()
	now := time.Now().Unix()
	_, err := db.Exec(
		`INSERT INTO collimato_workspace_users
		 (id, workspace_id, user_id, role, created_at, updated_at, deleted_at)
		 VALUES (?, ?, ?, ?, ?, ?, 0)`,
		model.NewID(), workspaceID, userID, role, now, now,
	)
	if err != nil {
		t.Fatalf("seedCollimatoWorkspaceUser: %v", err)
	}
}

func activeWorkspaceGroupCount(t *testing.T, db *sql.DB, workspaceID string) int {
	t.Helper()
	var n int
	if err := db.QueryRow(
		`SELECT COUNT(*) FROM collimato_workspace_groups WHERE workspace_id = ? AND deleted_at = 0`,
		workspaceID,
	).Scan(&n); err != nil {
		t.Fatalf("activeWorkspaceGroupCount: %v", err)
	}
	return n
}

func TestAddWorkspaceGroups_AttachesBatch(t *testing.T) {
	db := requireDB(t)
	resetCollimatoGroupTables(t)

	repo := &collimatoRepository{Db: db}
	ctx := context.Background()

	ws := seedCollimatoWorkspace(t, db)
	g1, g2 := seedGroup(t, db, false), seedGroup(t, db, false)

	if err := repo.AddWorkspaceGroups(ctx, ws, []string{g1, g2}, []string{"workspace_user"}, model.NewID()); err != nil {
		t.Fatalf("AddWorkspaceGroups: %v", err)
	}
	if got := activeWorkspaceGroupCount(t, db, ws); got != 2 {
		t.Errorf("expected 2 attachments, got %d", got)
	}
}

func TestAddWorkspaceGroups_IdempotentUpdatesRolesAndReactivates(t *testing.T) {
	db := requireDB(t)
	resetCollimatoGroupTables(t)

	repo := &collimatoRepository{Db: db}
	ctx := context.Background()

	ws := seedCollimatoWorkspace(t, db)
	g := seedGroup(t, db, false)
	adder := model.NewID()

	if err := repo.AddWorkspaceGroups(ctx, ws, []string{g}, []string{"viewer"}, adder); err != nil {
		t.Fatalf("first attach: %v", err)
	}
	// Soft-delete it, then re-attach with new roles: must reactivate, not dup.
	if err := repo.RemoveWorkspaceGroup(ctx, ws, g); err != nil {
		t.Fatalf("remove: %v", err)
	}
	if err := repo.AddWorkspaceGroups(ctx, ws, []string{g}, []string{"editor", "admin"}, adder); err != nil {
		t.Fatalf("re-attach: %v", err)
	}

	if got := activeWorkspaceGroupCount(t, db, ws); got != 1 {
		t.Fatalf("expected 1 active attachment after reactivate, got %d", got)
	}
	groups, err := repo.GetWorkspaceGroups(ctx, ws)
	if err != nil {
		t.Fatalf("GetWorkspaceGroups: %v", err)
	}
	if len(groups) != 1 {
		t.Fatalf("expected 1 listed group, got %d", len(groups))
	}
	gotRoles := append([]string{}, groups[0].Roles...)
	sort.Strings(gotRoles)
	if len(gotRoles) != 2 || gotRoles[0] != "admin" || gotRoles[1] != "editor" {
		t.Errorf("expected roles [admin editor] after re-attach, got %v", groups[0].Roles)
	}
}

func TestGetWorkspaceGroups_EnrichesAndSkipsDeleted(t *testing.T) {
	db := requireDB(t)
	resetCollimatoGroupTables(t)

	repo := &collimatoRepository{Db: db}
	ctx := context.Background()

	ws := seedCollimatoWorkspace(t, db)
	g := seedGroup(t, db, false)
	seedGroupMember(t, db, g, model.NewID())
	seedGroupMember(t, db, g, model.NewID())
	gone := seedGroup(t, db, false)

	if err := repo.AddWorkspaceGroups(ctx, ws, []string{g, gone}, []string{"workspace_user"}, model.NewID()); err != nil {
		t.Fatalf("AddWorkspaceGroups: %v", err)
	}
	if err := repo.RemoveWorkspaceGroup(ctx, ws, gone); err != nil {
		t.Fatalf("RemoveWorkspaceGroup: %v", err)
	}

	groups, err := repo.GetWorkspaceGroups(ctx, ws)
	if err != nil {
		t.Fatalf("GetWorkspaceGroups: %v", err)
	}
	if len(groups) != 1 {
		t.Fatalf("expected only the active group, got %d", len(groups))
	}
	if groups[0].GroupID != g {
		t.Errorf("unexpected group listed: %+v", groups[0])
	}
	if groups[0].Name == "" {
		t.Error("expected enriched group name")
	}
	if groups[0].MemberCount != 2 {
		t.Errorf("expected member_count 2, got %d", groups[0].MemberCount)
	}
}

func TestUpdateWorkspaceGroupRoles(t *testing.T) {
	db := requireDB(t)
	resetCollimatoGroupTables(t)

	repo := &collimatoRepository{Db: db}
	ctx := context.Background()

	ws := seedCollimatoWorkspace(t, db)
	g := seedGroup(t, db, false)
	if err := repo.AddWorkspaceGroups(ctx, ws, []string{g}, []string{"viewer"}, model.NewID()); err != nil {
		t.Fatalf("AddWorkspaceGroups: %v", err)
	}
	if err := repo.UpdateWorkspaceGroupRoles(ctx, ws, g, []string{"admin"}); err != nil {
		t.Fatalf("UpdateWorkspaceGroupRoles: %v", err)
	}

	groups, err := repo.GetWorkspaceGroups(ctx, ws)
	if err != nil {
		t.Fatalf("GetWorkspaceGroups: %v", err)
	}
	if len(groups) != 1 || len(groups[0].Roles) != 1 || groups[0].Roles[0] != "admin" {
		t.Errorf("expected roles [admin], got %+v", groups)
	}
}

func sortedRoles(in []string) []string {
	out := append([]string{}, in...)
	sort.Strings(out)
	return out
}

func TestGetEffectiveRolesForUser_UnionsAndDedupes(t *testing.T) {
	db := requireDB(t)
	resetCollimatoGroupTables(t)

	repo := &collimatoRepository{Db: db}
	ctx := context.Background()

	ws := seedCollimatoWorkspace(t, db)
	user := model.NewID()

	seedCollimatoWorkspaceUser(t, db, ws, user, "viewer,editor")
	// Group the user belongs to grants editor,admin (editor overlaps -> dedupe)
	g := seedGroup(t, db, false)
	seedGroupMember(t, db, g, user)
	if err := repo.AddWorkspaceGroups(ctx, ws, []string{g}, []string{"editor", "admin"}, model.NewID()); err != nil {
		t.Fatalf("AddWorkspaceGroups: %v", err)
	}

	roles, err := repo.GetEffectiveRolesForUser(ctx, ws, user)
	if err != nil {
		t.Fatalf("GetEffectiveRolesForUser: %v", err)
	}
	got := sortedRoles(roles)
	want := []string{"admin", "editor", "viewer"}
	if len(got) != 3 || got[0] != want[0] || got[1] != want[1] || got[2] != want[2] {
		t.Errorf("expected %v, got %v", want, roles)
	}
}

func TestGetEffectiveRolesForUser_GroupOnlyMember(t *testing.T) {
	db := requireDB(t)
	resetCollimatoGroupTables(t)

	repo := &collimatoRepository{Db: db}
	ctx := context.Background()

	ws := seedCollimatoWorkspace(t, db)
	user := model.NewID()
	g := seedGroup(t, db, false)
	seedGroupMember(t, db, g, user)
	if err := repo.AddWorkspaceGroups(ctx, ws, []string{g}, []string{"workspace_user"}, model.NewID()); err != nil {
		t.Fatalf("AddWorkspaceGroups: %v", err)
	}

	roles, err := repo.GetEffectiveRolesForUser(ctx, ws, user)
	if err != nil {
		t.Fatalf("GetEffectiveRolesForUser: %v", err)
	}
	if len(roles) != 1 || roles[0] != "workspace_user" {
		t.Errorf("group-only member should get the group's role, got %v", roles)
	}
}

func TestGetEffectiveRolesForUser_ExcludesDeletedGroupAndNonMembers(t *testing.T) {
	db := requireDB(t)
	resetCollimatoGroupTables(t)

	repo := &collimatoRepository{Db: db}
	ctx := context.Background()

	ws := seedCollimatoWorkspace(t, db)
	user := model.NewID()
	other := model.NewID()

	// Soft-deleted user_group: must contribute nothing even if attached.
	deletedGroup := seedGroup(t, db, true)
	seedGroupMember(t, db, deletedGroup, user)
	// A live group the user is NOT in.
	otherGroup := seedGroup(t, db, false)
	seedGroupMember(t, db, otherGroup, other)

	if err := repo.AddWorkspaceGroups(ctx, ws, []string{deletedGroup, otherGroup}, []string{"admin"}, model.NewID()); err != nil {
		t.Fatalf("AddWorkspaceGroups: %v", err)
	}

	roles, err := repo.GetEffectiveRolesForUser(ctx, ws, user)
	if err != nil {
		t.Fatalf("GetEffectiveRolesForUser: %v", err)
	}
	if len(roles) != 0 {
		t.Errorf("expected no roles (deleted group + not a member of the other), got %v", roles)
	}
}

func TestGetWorkspacesForUser_IncludesGroupAttached(t *testing.T) {
	db := requireDB(t)
	resetCollimatoGroupTables(t)

	repo := &collimatoRepository{Db: db}
	ctx := context.Background()

	ws := seedCollimatoWorkspace(t, db)
	user := seedUser(t, db)
	g := seedGroup(t, db, false)
	seedGroupMember(t, db, g, user)
	if err := repo.AddWorkspaceGroups(ctx, ws, []string{g}, []string{"workspace_user"}, model.NewID()); err != nil {
		t.Fatalf("AddWorkspaceGroups: %v", err)
	}

	wss, err := repo.GetWorkspacesForUser(user)
	if err != nil {
		t.Fatalf("GetWorkspacesForUser: %v", err)
	}
	if len(wss) != 1 || wss[0].ID != ws {
		t.Fatalf("group-only member should see the workspace, got %d (%+v)", len(wss), wss)
	}
}

func TestGetWorkspacesForUser_DirectAndGroupNotDuplicated(t *testing.T) {
	db := requireDB(t)
	resetCollimatoGroupTables(t)

	repo := &collimatoRepository{Db: db}
	ctx := context.Background()

	ws := seedCollimatoWorkspace(t, db)
	user := seedUser(t, db)
	seedCollimatoWorkspaceUser(t, db, ws, user, "admin")
	g := seedGroup(t, db, false)
	seedGroupMember(t, db, g, user)
	if err := repo.AddWorkspaceGroups(ctx, ws, []string{g}, []string{"viewer"}, model.NewID()); err != nil {
		t.Fatalf("AddWorkspaceGroups: %v", err)
	}

	wss, err := repo.GetWorkspacesForUser(user)
	if err != nil {
		t.Fatalf("GetWorkspacesForUser: %v", err)
	}
	if len(wss) != 1 {
		t.Errorf("UNION should de-dupe a user with both paths, got %d", len(wss))
	}
}

func TestGetWorkspacesForUser_NoAccessSeesNothing(t *testing.T) {
	db := requireDB(t)
	resetCollimatoGroupTables(t)

	repo := &collimatoRepository{Db: db}
	ctx := context.Background()

	ws := seedCollimatoWorkspace(t, db)
	g := seedGroup(t, db, false)
	// A group attached to the workspace, but our user is not in it.
	if err := repo.AddWorkspaceGroups(ctx, ws, []string{g}, []string{"viewer"}, model.NewID()); err != nil {
		t.Fatalf("AddWorkspaceGroups: %v", err)
	}
	stranger := seedUser(t, db)

	wss, err := repo.GetWorkspacesForUser(stranger)
	if err != nil {
		t.Fatalf("GetWorkspacesForUser: %v", err)
	}
	if len(wss) != 0 {
		t.Errorf("non-member should see no workspaces, got %d", len(wss))
	}
}

func TestGetWorkspaceGroupMembers_ExcludesDirectAndUnionsRoles(t *testing.T) {
	db := requireDB(t)
	resetCollimatoGroupTables(t)

	repo := &collimatoRepository{Db: db}
	ctx := context.Background()

	ws := seedCollimatoWorkspace(t, db)

	directUser := seedUser(t, db)
	seedCollimatoWorkspaceUser(t, db, ws, directUser, "admin")

	groupUser := seedUser(t, db)
	g1, g2 := seedGroup(t, db, false), seedGroup(t, db, false)
	seedGroupMember(t, db, g1, groupUser)
	seedGroupMember(t, db, g2, groupUser)
	// directUser is also in g1 but must be excluded (already a direct member).
	seedGroupMember(t, db, g1, directUser)

	if err := repo.AddWorkspaceGroups(ctx, ws, []string{g1}, []string{"viewer"}, model.NewID()); err != nil {
		t.Fatalf("AddWorkspaceGroups g1: %v", err)
	}
	if err := repo.AddWorkspaceGroups(ctx, ws, []string{g2}, []string{"editor"}, model.NewID()); err != nil {
		t.Fatalf("AddWorkspaceGroups g2: %v", err)
	}

	members, err := repo.GetWorkspaceGroupMembers(ctx, ws)
	if err != nil {
		t.Fatalf("GetWorkspaceGroupMembers: %v", err)
	}
	if len(members) != 1 {
		t.Fatalf("expected only the group-only user, got %d (%+v)", len(members), members)
	}
	m := members[0]
	if m.UserID != groupUser {
		t.Errorf("expected groupUser, got %s", m.UserID)
	}
	if !m.ViaGroup {
		t.Error("expected ViaGroup=true")
	}
	if !strings.Contains(m.Role, "viewer") || !strings.Contains(m.Role, "editor") {
		t.Errorf("expected union of roles viewer+editor, got %q", m.Role)
	}
	if m.UserInfo.ID != groupUser {
		t.Errorf("expected hydrated UserInfo, got %+v", m.UserInfo)
	}
}

func TestGetWorkspaceGroupMembers_SkipsDetachedAndDeletedGroups(t *testing.T) {
	db := requireDB(t)
	resetCollimatoGroupTables(t)

	repo := &collimatoRepository{Db: db}
	ctx := context.Background()

	ws := seedCollimatoWorkspace(t, db)
	u := seedUser(t, db)
	g := seedGroup(t, db, false)
	seedGroupMember(t, db, g, u)
	if err := repo.AddWorkspaceGroups(ctx, ws, []string{g}, []string{"viewer"}, model.NewID()); err != nil {
		t.Fatalf("AddWorkspaceGroups: %v", err)
	}
	if err := repo.RemoveWorkspaceGroup(ctx, ws, g); err != nil {
		t.Fatalf("RemoveWorkspaceGroup: %v", err)
	}

	members, err := repo.GetWorkspaceGroupMembers(ctx, ws)
	if err != nil {
		t.Fatalf("GetWorkspaceGroupMembers: %v", err)
	}
	if len(members) != 0 {
		t.Errorf("detached group should contribute no members, got %d", len(members))
	}
}

func TestGetEffectiveRolesForUser_DetachedGroupDropsRoles(t *testing.T) {
	db := requireDB(t)
	resetCollimatoGroupTables(t)

	repo := &collimatoRepository{Db: db}
	ctx := context.Background()

	ws := seedCollimatoWorkspace(t, db)
	user := model.NewID()
	g := seedGroup(t, db, false)
	seedGroupMember(t, db, g, user)
	if err := repo.AddWorkspaceGroups(ctx, ws, []string{g}, []string{"admin"}, model.NewID()); err != nil {
		t.Fatalf("AddWorkspaceGroups: %v", err)
	}

	// Detaching the group must drop the user's roles immediately (live resolution).
	if err := repo.RemoveWorkspaceGroup(ctx, ws, g); err != nil {
		t.Fatalf("RemoveWorkspaceGroup: %v", err)
	}
	roles, err := repo.GetEffectiveRolesForUser(ctx, ws, user)
	if err != nil {
		t.Fatalf("GetEffectiveRolesForUser: %v", err)
	}
	if len(roles) != 0 {
		t.Errorf("expected no roles after detach, got %v", roles)
	}
}

func TestGetWorkspaceMemberIDs_DirectAndGroupMembersPerWorkspace(t *testing.T) {
	db := requireDB(t)
	resetCollimatoGroupTables(t)

	repo := &collimatoRepository{Db: db}
	ctx := context.Background()

	ws1 := seedCollimatoWorkspace(t, db)
	ws2 := seedCollimatoWorkspace(t, db)
	empty := seedCollimatoWorkspace(t, db)

	direct := seedUser(t, db)
	viaGroup := seedUser(t, db)
	both := seedUser(t, db)
	other := seedUser(t, db)

	seedCollimatoWorkspaceUser(t, db, ws1, direct, "admin")
	seedCollimatoWorkspaceUser(t, db, ws1, both, "viewer")
	seedCollimatoWorkspaceUser(t, db, ws2, other, "viewer")

	g := seedGroup(t, db, false)
	seedGroupMember(t, db, g, viaGroup)
	seedGroupMember(t, db, g, both)
	if err := repo.AddWorkspaceGroups(ctx, ws1, []string{g}, []string{"viewer"}, model.NewID()); err != nil {
		t.Fatalf("AddWorkspaceGroups: %v", err)
	}

	members, err := repo.GetWorkspaceMemberIDs(ctx, []string{ws1, ws2, empty})
	if err != nil {
		t.Fatalf("GetWorkspaceMemberIDs: %v", err)
	}

	got := append([]string(nil), members[ws1]...)
	sort.Strings(got)

	want := []string{direct, viaGroup, both}
	sort.Strings(want)

	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("ws1 members = %v, want %v (direct, group-only and both, once each)", got, want)
	}

	if len(members[ws2]) != 1 || members[ws2][0] != other {
		t.Errorf("ws2 members = %v, want [%s]", members[ws2], other)
	}

	if len(members[empty]) != 0 {
		t.Errorf("workspace with no members = %v, want none", members[empty])
	}
}

func TestGetWorkspaceMemberIDs_SkipsDetachedAndDeletedGroups(t *testing.T) {
	db := requireDB(t)
	resetCollimatoGroupTables(t)

	repo := &collimatoRepository{Db: db}
	ctx := context.Background()

	ws := seedCollimatoWorkspace(t, db)

	detachedMember := seedUser(t, db)
	detached := seedGroup(t, db, false)
	seedGroupMember(t, db, detached, detachedMember)
	if err := repo.AddWorkspaceGroups(ctx, ws, []string{detached}, []string{"viewer"}, model.NewID()); err != nil {
		t.Fatalf("AddWorkspaceGroups: %v", err)
	}

	if err := repo.RemoveWorkspaceGroup(ctx, ws, detached); err != nil {
		t.Fatalf("RemoveWorkspaceGroup: %v", err)
	}

	deletedMember := seedUser(t, db)
	deleted := seedGroup(t, db, true)
	seedGroupMember(t, db, deleted, deletedMember)
	if err := repo.AddWorkspaceGroups(ctx, ws, []string{deleted}, []string{"viewer"}, model.NewID()); err != nil {
		t.Fatalf("AddWorkspaceGroups: %v", err)
	}

	members, err := repo.GetWorkspaceMemberIDs(ctx, []string{ws})
	if err != nil {
		t.Fatalf("GetWorkspaceMemberIDs: %v", err)
	}

	if len(members[ws]) != 0 {
		t.Errorf("expected no members through a detached or deleted group, got %v", members[ws])
	}
}

func TestGetWorkspaceMemberIDs_NoWorkspaces(t *testing.T) {
	db := requireDB(t)

	repo := &collimatoRepository{Db: db}

	members, err := repo.GetWorkspaceMemberIDs(context.Background(), nil)
	if err != nil {
		t.Fatalf("GetWorkspaceMemberIDs: %v", err)
	}

	if members == nil || len(members) != 0 {
		t.Errorf("expected an empty, non-nil map, got %#v", members)
	}
}
