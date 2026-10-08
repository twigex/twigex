// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package sqlstore

import (
	"context"
	"sort"
	"strings"
	"testing"
)

func TestUpdateKeepsTheDescriptionUnlessGivenOne(t *testing.T) {
	db := requireDB(t)
	resetWorkspaceGroupTables(t)
	repo := &workspaceRepository{Db: db}
	ws := seedWorkspace(t, db)

	description := "Launch plan"
	if _, err := repo.Update(ws, "Renamed", &description); err != nil {
		t.Fatal(err)
	}

	if _, err := repo.Update(ws, "Renamed again", nil); err != nil {
		t.Fatal(err)
	}

	var title, got string
	if err := db.QueryRow("SELECT title, description FROM workspaces WHERE id = ?", ws).Scan(&title, &got); err != nil {
		t.Fatal(err)
	}

	if title != "Renamed again" || got != description {
		t.Errorf("title %q description %q, want the new title and the description kept", title, got)
	}
}

// The workspace list reads every workspace at once; each read has to give
// what the per-workspace reads it replaced give.
func TestWorkspaceListBatchesMatchThePerWorkspaceReads(t *testing.T) {
	db := requireDB(t)
	resetWorkspaceGroupTables(t)
	cleanTables(t, "workspace_members", "workspace_roles", "workspace_table_permissions", "workspace_tables")
	repo := &workspaceRepository{Db: db}
	ctx := context.Background()

	ws1, ws2, ws3 := seedWorkspace(t, db), seedWorkspace(t, db), seedWorkspace(t, db)
	group := seedUserGroup(t, db)
	deletedGroup := seedUserGroup(t, db)
	mustExec(t, "UPDATE user_groups SET deleted_at = 1 WHERE id = ?", deletedGroup)
	if err := repo.AddGroups(ctx, ws1, []string{group}, []string{"editor"}, "adder"); err != nil {
		t.Fatal(err)
	}

	if err := repo.AddGroups(ctx, ws3, []string{deletedGroup}, []string{"owner"}, "adder"); err != nil {
		t.Fatal(err)
	}

	me, other := seedUser(t, db), seedUser(t, db)
	mustExec(t, `INSERT INTO workspace_members (id, user_id, workspace_id, role, date_joined) VALUES
		('m1', ?, ?, 'viewer', 1), ('m2', ?, ?, 'owner member', 1), ('m3', ?, ?, 'viewer', 1)`,
		me, ws1, me, ws2, other, ws1)
	mustExec(t, `INSERT INTO group_members (id, group_id, user_id, role, joined_at) VALUES
		('g1', ?, ?, 'member', 1), ('g2', ?, ?, 'member', 1)`, group, me, deletedGroup, me)
	for _, r := range [][4]string{
		{"r1", "viewer", ws1, "1"},
		{"r2", "editor", ws1, "0"},
		{"r3", "owner", ws2, "0"},
		{"r4", "member", ws2, "1"},
		{"r5", "owner", ws3, "0"},
	} {
		mustExec(t, `INSERT INTO workspace_roles (id, name, displayname, description, permissions, workspace_id, per_table_mode, created_at, updated_at)
			VALUES (?, ?, ?, '', '["view_workspace"]', ?, ?, 0, 0)`, r[0], r[1], r[1], r[2], r[3])
	}
	for _, tb := range [][2]string{{"t1", ws1}, {"t2", ws2}, {"t3", ws2}} {
		mustExec(t, `INSERT INTO workspace_tables (id, workspace_id, name, display_name, linked, single_select, parent_table_id, both_direction_link, second_table_id, folder_id, deleted_at)
			VALUES (?, ?, ?, '', false, false, '', false, '', '', 0)`, tb[0], tb[1], tb[0])
	}
	mustExec(t, `INSERT INTO workspace_table_permissions (id, role_id, workspace_id, table_id, action, created_at, updated_at) VALUES
		('p1', 'r1', ?, 't1', 'view', 0, 0), ('p2', 'r4', ?, 't2', 'view', 0, 0), ('p3', 'r3', ?, 't3', 'view', 0, 0)`,
		ws1, ws2, ws2)

	ids := []string{ws1, ws2, ws3}
	roleNames, err := repo.GetUserRoleNamesByWorkspace(ctx, me, ids)
	if err != nil {
		t.Fatal(err)
	}

	roles, err := repo.GetRolesForWorkspaces(ctx, ids)
	if err != nil {
		t.Fatal(err)
	}

	members, err := repo.GetMembersForWorkspaces(ctx, ids)
	if err != nil {
		t.Fatal(err)
	}

	groupCounts, err := repo.GetGroupCounts(ctx, ids)
	if err != nil {
		t.Fatal(err)
	}

	for _, ws := range ids {
		var want []string
		if direct, _ := repo.GetUserByUserID(me, ws); direct != nil {
			want = strings.Fields(direct.Role)
		}
		viaGroups, _ := repo.GetGroupRolesForUser(me, ws)
		want = append(want, viaGroups...)
		sort.Strings(want)
		got := append([]string(nil), roleNames[ws]...)
		sort.Strings(got)
		if strings.Join(got, ",") != strings.Join(want, ",") {
			t.Errorf("roles held in %s = %v, want %v", ws, got, want)
		}

		one, err := repo.GetRolesByName([]string{"viewer", "editor", "owner", "member"}, ws)
		if err != nil {
			t.Fatal(err)
		}

		if len(roles[ws]) != len(one) {
			t.Errorf("%s has %d roles, want %d", ws, len(roles[ws]), len(one))
		}

		perWorkspace, _ := repo.GetMembers(ws)
		if len(members[ws]) != len(perWorkspace) {
			t.Errorf("%s has %d members, want %d", ws, len(members[ws]), len(perWorkspace))
		}

		groups, _ := repo.GetGroups(ctx, ws)
		if groupCounts[ws] != len(groups) {
			t.Errorf("%s is shared with %d groups, want %d", ws, groupCounts[ws], len(groups))
		}
	}

	perms, err := repo.GetTablePermissionsForRoleIDs(ctx, []string{"r1", "r4"})
	if err != nil {
		t.Fatal(err)
	}

	var tables []string
	for _, p := range perms {
		tables = append(tables, p.TableID)
	}
	sort.Strings(tables)
	if strings.Join(tables, ",") != "t1,t2" {
		t.Errorf("table permissions of r1 and r4 = %v, want t1 and t2 only", tables)
	}
}
