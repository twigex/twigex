// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package sqlstore

// Fixtures shared by the share-resolution and file-query suites.

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/twigex/twigex/model"
)

var shareRowSeq int

func insertFile(t *testing.T, db *sql.DB, id, owner, parent, storage, ftype string, isFolder bool, deletedAt int64) {
	t.Helper()
	now := time.Now().Unix()
	_, err := db.Exec(
		`INSERT INTO files
		 (id, owner, parent, storage, name, displayname, size, type, is_folder, created_at, modified_at, deleted_at)
		 VALUES (?, ?, ?, ?, ?, ?, 0, ?, ?, ?, ?, ?)`,
		id, owner, parent, storage, id, id, ftype, isFolder, now, now, deletedAt,
	)
	if err != nil {
		t.Fatalf("insertFile %s: %v", id, err)
	}
}

func seedShareFile(t *testing.T, id, owner, parent, storage, ftype string, isFolder bool, deletedAt int64) {
	t.Helper()
	insertFile(t, testDB, id, owner, parent, storage, ftype, isFolder, deletedAt)
}

func seedShareRow(t *testing.T, fileID string, shareType int, shareWith string, expiration int64, level model.AccessLevel) {
	t.Helper()
	shareRowSeq++
	id := fmt.Sprintf("share-%d", shareRowSeq)
	now := time.Now().Unix()
	_, err := testDB.Exec(
		`INSERT INTO file_share
		 (id, file_id, initiator, share_type, share_with, expiration, access_level, time_shared, created_at, updated_at)
		 VALUES (?, ?, 'initiator', ?, ?, ?, ?, ?, ?, ?)`,
		id, fileID, shareType, shareWith, expiration, level, now, now, now,
	)
	if err != nil {
		t.Fatalf("seedShareRow %s: %v", fileID, err)
	}
}

func seedFavorite(t *testing.T, userID, itemID string) {
	t.Helper()
	shareRowSeq++
	_, err := testDB.Exec(
		`INSERT INTO favorites (id, user_id, app, item_id) VALUES (?, ?, 'files', ?)`,
		fmt.Sprintf("fav-%d", shareRowSeq), userID, itemID,
	)
	if err != nil {
		t.Fatalf("seedFavorite %s/%s: %v", userID, itemID, err)
	}
}

func mustExec(t *testing.T, query string, args ...any) {
	t.Helper()
	if _, err := testDB.Exec(query, args...); err != nil {
		t.Fatalf("exec %q: %v", query, err)
	}
}

func idSet(files []model.File) map[string]bool {
	s := make(map[string]bool, len(files))
	for _, f := range files {
		s[f.ID] = true
	}
	return s
}
