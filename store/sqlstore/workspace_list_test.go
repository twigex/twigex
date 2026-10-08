// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package sqlstore

import (
	"database/sql"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/twigex/twigex/model"
)

func resetWorkspaceListTables(t *testing.T) {
	t.Helper()
	cleanTables(t,
		"workspace_members",
		"workspace_groups",
		"group_members",
		"user_groups",
		"workspaces",
		"users",
	)
}

func seedTitledWorkspace(t *testing.T, db *sql.DB, title string, deletedAt int64) string {
	t.Helper()
	id := model.NewID()
	now := time.Now().Unix()
	_, err := db.Exec(
		`INSERT INTO workspaces (id, title, description, start_date, end_date, pre_fix, created_at, updated_at, deleted_at, created_by)
		 VALUES (?, ?, '', 0, 0, '', ?, ?, ?, ?)`,
		id, title, now, now, deletedAt, model.NewID(),
	)
	if err != nil {
		t.Fatalf("seedTitledWorkspace %s: %v", title, err)
	}
	return id
}

func seedWorkspaceMember(t *testing.T, db *sql.DB, workspaceID, userID string) {
	t.Helper()
	_, err := db.Exec(
		`INSERT INTO workspace_members (id, user_id, workspace_id, role, date_joined)
		 VALUES (?, ?, ?, 'user', ?)`,
		model.NewID(), userID, workspaceID, time.Now().Unix(),
	)
	if err != nil {
		t.Fatalf("seedWorkspaceMember: %v", err)
	}
}

func seedWorkspaceGroup(t *testing.T, db *sql.DB, workspaceID, groupID string) {
	t.Helper()
	_, err := db.Exec(
		`INSERT INTO workspace_groups (id, workspace_id, group_id, roles, added_by, added_at)
		 VALUES (?, ?, ?, 'user', ?, ?)`,
		model.NewID(), workspaceID, groupID, model.NewID(), time.Now().Unix(),
	)
	if err != nil {
		t.Fatalf("seedWorkspaceGroup: %v", err)
	}
}

func workspaceTitles(workspaces []model.Workspace) string {
	titles := make([]string, len(workspaces))
	for i, w := range workspaces {
		titles[i] = w.Title
	}
	return strings.Join(titles, ",")
}

func TestGetAllForUserCombinesDirectAndGroupAccess(t *testing.T) {
	db := requireDB(t)
	resetWorkspaceListTables(t)
	repo := &workspaceRepository{Db: db}

	user := seedUser(t, db)
	other := seedUser(t, db)
	liveGroup := seedGroup(t, db, false)
	deletedGroup := seedGroup(t, db, true)
	seedGroupMember(t, db, liveGroup, user)
	seedGroupMember(t, db, deletedGroup, user)

	golf := seedTitledWorkspace(t, db, "Golf", 0)
	seedWorkspaceMember(t, db, golf, other)

	charlie := seedTitledWorkspace(t, db, "Charlie", 0)
	seedWorkspaceMember(t, db, charlie, user)
	seedWorkspaceGroup(t, db, charlie, liveGroup)

	echo := seedTitledWorkspace(t, db, "Echo", time.Now().Unix())
	seedWorkspaceMember(t, db, echo, user)

	bravo := seedTitledWorkspace(t, db, "Bravo", 0)
	seedWorkspaceGroup(t, db, bravo, liveGroup)

	delta := seedTitledWorkspace(t, db, "Delta", 0)
	seedWorkspaceGroup(t, db, delta, deletedGroup)

	alpha := seedTitledWorkspace(t, db, "Alpha", 0)
	seedWorkspaceMember(t, db, alpha, user)

	got, err := repo.GetAllForUser(user)
	if err != nil {
		t.Fatalf("GetAllForUser: %v", err)
	}
	if titles := workspaceTitles(got); titles != "Alpha,Bravo,Charlie" {
		t.Fatalf("titles = %s, want Alpha,Bravo,Charlie", titles)
	}
}

func TestGetAllForUserWithoutAccessIsEmpty(t *testing.T) {
	db := requireDB(t)
	resetWorkspaceListTables(t)
	repo := &workspaceRepository{Db: db}

	ws := seedTitledWorkspace(t, db, "Alpha", 0)
	seedWorkspaceMember(t, db, ws, seedUser(t, db))

	got, err := repo.GetAllForUser(model.NewID())
	if err != nil {
		t.Fatalf("GetAllForUser: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("got %d workspaces, want 0", len(got))
	}
}

func TestWorkspaceReadsTolerateNullColumns(t *testing.T) {
	db := requireDB(t)
	resetWorkspaceListTables(t)
	repo := &workspaceRepository{Db: db}
	user := seedUser(t, db)

	nullable := []string{"description", "start_date", "end_date", "updated_at", "deleted_at", "created_by"}
	rows := map[string][]string{"A normal": nil, "B all null": nullable}
	for i, col := range nullable {
		rows[fmt.Sprintf("C%d null %s", i, col)] = []string{col}
	}

	for title, nulls := range rows {
		id := model.NewID()
		mustExec(t, `INSERT INTO workspaces (id, title, description, start_date, end_date, pre_fix, created_at, updated_at, deleted_at, created_by)
			VALUES (?, ?, 'about', 5, 9, 'abc_', 1, 7, 0, 'someone')`, id, title)
		for _, col := range nulls {
			mustExec(t, "UPDATE workspaces SET `"+col+"` = NULL WHERE id = ?", id)
		}
		seedWorkspaceMember(t, db, id, user)

		row, err := repo.GetRow(id)
		if err != nil {
			t.Errorf("GetRow %q: %v", title, err)
			continue
		}
		if row.Title != title {
			t.Errorf("GetRow %q returned title %q", title, row.Title)
		}
		for _, col := range nulls {
			var got any
			switch col {
			case "description":
				got = row.Description
			case "start_date":
				got = row.StartDate
			case "end_date":
				got = row.EndDate
			case "updated_at":
				got = row.UpdatedAt
			case "deleted_at":
				got = row.DeletedAt
			default:
				continue
			}
			if got != "" && got != int64(0) {
				t.Errorf("GetRow %q: NULL %s read as %v, want the zero value", title, col, got)
			}
		}
	}

	list, err := repo.GetAllForUser(user)
	if err != nil {
		t.Fatalf("GetAllForUser: %v", err)
	}
	if len(list) != len(rows) {
		t.Fatalf("GetAllForUser returned %d workspaces, want %d", len(list), len(rows))
	}
	for _, w := range list {
		if w.Title == "A normal" && (w.Description != "about" || w.StartDate != 5 || w.EndDate != 9 || w.UpdatedAt != 7) {
			t.Errorf("normal row read back as %+v", w)
		}
	}
}
