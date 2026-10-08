// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package sqlstore

import "testing"

func TestDeleteFileRemovesExternalSharesAndMetadata(t *testing.T) {
	db := requireDB(t)
	cleanTables(t, "files", "external_share", "file_metadata_entries")
	repo := &fileRepository{Db: db}

	for _, id := range []string{"doomed", "kept"} {
		insertFile(t, db, id, "alice", "drive", "s1", "text/plain", false, 0)
		mustExec(t, `INSERT INTO external_share (id, owner, file_id, password, share_token, active, created_at, updated_at)
			VALUES (?, 'alice', ?, '', ?, 0, 0, 0)`, "es-"+id, id, "tok-"+id)
		mustExec(t, `INSERT INTO file_metadata_entries (id, file_id, metadata_id, value) VALUES (?, ?, 'meta', 'v')`, "me-"+id, id)
	}

	if ok, err := repo.Delete("doomed", "s1"); err != nil || !ok {
		t.Fatalf("Delete = %v, %v", ok, err)
	}

	for _, table := range []string{"external_share", "file_metadata_entries"} {
		if n := countWhere(t, "SELECT COUNT(*) FROM `"+table+"` WHERE file_id = 'doomed'"); n != 0 {
			t.Errorf("%s still has %d rows for the deleted file", table, n)
		}
		if n := countWhere(t, "SELECT COUNT(*) FROM `"+table+"` WHERE file_id = 'kept'"); n != 1 {
			t.Errorf("%s has %d rows for the other file, want 1", table, n)
		}
	}
}

func TestDeleteChartRemovesItFromDashboards(t *testing.T) {
	db := requireDB(t)
	cleanTables(t, "collimato_charts", "collimato_dashboard_charts")
	repo := &collimatoRepository{Db: db}

	for _, id := range []string{"doomed", "kept"} {
		mustExec(t, `INSERT INTO collimato_charts (id, name, chart_type, configuration, data, owner_id, workspace_id, created_at)
			VALUES (?, 'c', 'bar', '{}', '{}', 'alice', 'ws', 0)`, id)
		mustExec(t, `INSERT INTO collimato_dashboard_charts (id, dashboard_id, chart_id, created_by, position, created_at)
			VALUES (?, 'dash', ?, 'alice', '{}', 0)`, "dc-"+id, id)
	}

	if err := repo.DeleteChart("doomed"); err != nil {
		t.Fatalf("DeleteChart: %v", err)
	}

	if n := countWhere(t, `SELECT COUNT(*) FROM collimato_dashboard_charts WHERE chart_id = 'doomed'`); n != 0 {
		t.Errorf("dashboard still links the deleted chart %d times", n)
	}
	if n := countWhere(t, `SELECT COUNT(*) FROM collimato_dashboard_charts WHERE chart_id = 'kept'`); n != 1 {
		t.Errorf("the other chart has %d dashboard links, want 1", n)
	}
	if n := countWhere(t, `SELECT COUNT(*) FROM collimato_charts WHERE id = 'doomed'`); n != 0 {
		t.Errorf("chart row still exists")
	}
}
