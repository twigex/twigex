// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package sqlstore

import (
	"testing"
	"time"

	"github.com/twigex/twigex/model"
)

func countCollimatoRows(t *testing.T, table, workspaceID string) int {
	t.Helper()
	var n int
	if err := testDB.QueryRow("SELECT COUNT(*) FROM `"+table+"` WHERE workspace_id = ?", workspaceID).Scan(&n); err != nil {
		t.Fatalf("count %s: %v", table, err)
	}
	return n
}

func TestDeleteCollimatoWorkspaceRemovesGroupsAndFiles(t *testing.T) {
	db := requireDB(t)
	cleanTables(t, "collimato_workspaces", "collimato_workspace_groups", "collimato_workspace_files", "collimato_workspace_users")
	repo := &collimatoRepository{Db: db}

	doomed := seedCollimatoWorkspace(t, db)
	kept := seedCollimatoWorkspace(t, db)
	now := time.Now().Unix()
	for _, ws := range []string{doomed, kept} {
		mustExec(t,
			`INSERT INTO collimato_workspace_groups (id, workspace_id, group_id, roles, added_by, added_at, deleted_at)
			 VALUES (?, ?, ?, 'user', ?, ?, 0)`,
			model.NewID(), ws, model.NewID(), model.NewID(), now)
		mustExec(t,
			`INSERT INTO collimato_workspace_files (id, workspace_id, name, file_type, content, encrypted, created_at, updated_at)
			 VALUES (?, ?, 'orders.yml', 'model', 'cubes: []', 1, ?, ?)`,
			model.NewID(), ws, now, now)
	}

	if err := repo.DeleteWorkspace(doomed); err != nil {
		t.Fatalf("DeleteWorkspace: %v", err)
	}

	for _, table := range []string{"collimato_workspace_groups", "collimato_workspace_files"} {
		if n := countCollimatoRows(t, table, doomed); n != 0 {
			t.Errorf("%s still has %d rows for the deleted workspace", table, n)
		}
		if n := countCollimatoRows(t, table, kept); n != 1 {
			t.Errorf("%s has %d rows for the other workspace, want 1", table, n)
		}
	}
}

func TestAddCollimatoUsersTakesNewRoleOnlyWhenRejoining(t *testing.T) {
	db := requireDB(t)
	cleanTables(t, "collimato_workspace_users")
	repo := &collimatoRepository{Db: db}

	if err := repo.AddWorkspaceUsers("ws", []string{"removed-admin"}, []string{"admin"}); err != nil {
		t.Fatalf("add admin: %v", err)
	}
	if err := repo.AddWorkspaceUsers("ws", []string{"active-editor"}, []string{"editor"}); err != nil {
		t.Fatalf("add editor: %v", err)
	}
	if err := repo.RemoveWorkspaceUser("ws", "removed-admin"); err != nil {
		t.Fatalf("remove admin: %v", err)
	}

	if err := repo.AddWorkspaceUsers("ws", []string{"removed-admin", "active-editor"}, []string{"viewer"}); err != nil {
		t.Fatalf("re-add: %v", err)
	}

	for user, want := range map[string]string{"removed-admin": "viewer", "active-editor": "editor"} {
		var role string
		var deletedAt int64
		if err := db.QueryRow(`SELECT role, deleted_at FROM collimato_workspace_users WHERE workspace_id = 'ws' AND user_id = ?`, user).Scan(&role, &deletedAt); err != nil {
			t.Fatalf("read %s: %v", user, err)
		}
		if role != want || deletedAt != 0 {
			t.Errorf("%s: role %q deleted_at %d, want %q and active", user, role, deletedAt, want)
		}
	}
}
