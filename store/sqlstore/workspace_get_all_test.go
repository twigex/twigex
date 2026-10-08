// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package sqlstore

import (
	"context"
	"testing"
)

func TestWorkspaceGetAllListsLiveWorkspacesByTitle(t *testing.T) {
	requireDB(t)
	cleanTables(t, "workspaces")

	mustExec(t, `INSERT INTO workspaces (id, title, description, start_date, end_date, pre_fix, created_at, updated_at, deleted_at, created_by)
		VALUES ('w1', 'Zeta', '', 0, 0, 'za_', 0, 0, 0, 'u1'),
		       ('w2', 'Alpha', '', 0, 0, 'al_', 0, 0, 0, 'u1'),
		       ('w3', 'Beta', '', 0, 0, 'be_', 0, 0, 123, 'u1')`)

	repo := &workspaceRepository{Db: testDB}

	workspaces, err := repo.GetAll(context.Background())
	if err != nil {
		t.Fatalf("GetAll: %v", err)
	}

	if len(workspaces) != 2 || workspaces[0].ID != "w2" || workspaces[1].ID != "w1" {
		t.Errorf("expected Alpha then Zeta without the deleted Beta, got %+v", workspaces)
	}
}
