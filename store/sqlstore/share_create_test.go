// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package sqlstore

import (
	"testing"

	"github.com/twigex/twigex/model"
)

func TestCreateShareReturnsStoredRow(t *testing.T) {
	db := requireDB(t)
	cleanTables(t, "file_share")
	repo := &fileRepository{Db: db}

	got, err := repo.CreateShare(model.File{ID: "file-1"}, model.SHARE_TYPE_USER, "alice", "bob", 0, model.AccessViewer)
	if err != nil {
		t.Fatalf("CreateShare: %v", err)
	}

	var storedID string
	if err := db.QueryRow(`SELECT id FROM file_share WHERE file_id = 'file-1'`).Scan(&storedID); err != nil {
		t.Fatalf("read share: %v", err)
	}
	if got.ID != storedID {
		t.Fatalf("returned id %s, stored id %s", got.ID, storedID)
	}
}

func TestCreateShareAgainUpdatesExistingShare(t *testing.T) {
	db := requireDB(t)
	cleanTables(t, "file_share")
	repo := &fileRepository{Db: db}
	file := model.File{ID: "file-1"}

	first, err := repo.CreateShare(file, model.SHARE_TYPE_USER, "alice", "bob", 0, model.AccessViewer)
	if err != nil {
		t.Fatalf("first CreateShare: %v", err)
	}
	second, err := repo.CreateShare(file, model.SHARE_TYPE_USER, "carol", "bob", 1999999999, model.AccessEditor)
	if err != nil {
		t.Fatalf("second CreateShare: %v", err)
	}

	var rows int
	if err := db.QueryRow(`SELECT COUNT(*) FROM file_share WHERE file_id = 'file-1'`).Scan(&rows); err != nil {
		t.Fatalf("count shares: %v", err)
	}
	if rows != 1 {
		t.Fatalf("got %d share rows, want 1", rows)
	}
	if second.ID != first.ID {
		t.Errorf("id changed from %s to %s", first.ID, second.ID)
	}
	if second.AccessLevel != model.AccessEditor || second.Expiration != 1999999999 {
		t.Errorf("level %d expiration %d, want %d and 1999999999", second.AccessLevel, second.Expiration, model.AccessEditor)
	}
	if second.Initiator != "alice" {
		t.Errorf("initiator = %s, want the original sharer alice", second.Initiator)
	}
}
