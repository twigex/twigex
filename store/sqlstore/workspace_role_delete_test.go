// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package sqlstore

import (
	"context"
	"testing"
)

func TestDeleteRoleUnassignsItFromMembersAndGroups(t *testing.T) {
	requireDB(t)
	cleanTables(t, "workspace_roles", "workspace_members", "workspace_groups")

	mustExec(t, `INSERT INTO workspace_roles (id, name, displayname, description, permissions, workspace_id, per_table_mode, created_at, updated_at)
		VALUES ('r1', 'editor', 'Editor', '', '', 'w1', 0, 0, 0)`)
	mustExec(t, `INSERT INTO workspace_members (id, user_id, workspace_id, role, date_joined)
		VALUES ('m1', 'only', 'w1', 'editor', 0),
		       ('m2', 'several', 'w1', 'admin editor', 0),
		       ('m3', 'other', 'w1', 'user', 0),
		       ('m4', 'elsewhere', 'w2', 'editor', 0)`)
	mustExec(t, `INSERT INTO workspace_groups (id, workspace_id, group_id, roles, added_by, added_at)
		VALUES ('g1', 'w1', 'only-group', 'editor', 'u1', 0),
		       ('g2', 'w1', 'mixed-group', 'editor user', 'u1', 0)`)

	repo := &workspaceRepository{Db: testDB}

	if err := repo.DeleteRole(context.Background(), "w1", "r1", "editor", "user"); err != nil {
		t.Fatalf("DeleteRole: %v", err)
	}

	var roles int
	if err := testDB.QueryRow(`SELECT COUNT(*) FROM workspace_roles WHERE id = 'r1'`).Scan(&roles); err != nil {
		t.Fatalf("count roles: %v", err)
	}

	if roles != 0 {
		t.Errorf("expected the role to be deleted")
	}

	members := map[string]string{
		"only":      "user",
		"several":   "admin",
		"other":     "user",
		"elsewhere": "editor",
	}

	for userID, want := range members {
		var got string
		if err := testDB.QueryRow(`SELECT role FROM workspace_members WHERE user_id = ?`, userID).Scan(&got); err != nil {
			t.Fatalf("read member %s: %v", userID, err)
		}

		if got != want {
			t.Errorf("member %s: got role %q, want %q", userID, got, want)
		}
	}

	groups := map[string]string{
		"only-group":  "user",
		"mixed-group": "user",
	}

	for groupID, want := range groups {
		var got string
		if err := testDB.QueryRow(`SELECT roles FROM workspace_groups WHERE group_id = ?`, groupID).Scan(&got); err != nil {
			t.Fatalf("read group %s: %v", groupID, err)
		}

		if got != want {
			t.Errorf("group %s: got roles %q, want %q", groupID, got, want)
		}
	}
}
