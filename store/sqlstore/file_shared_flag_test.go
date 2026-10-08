// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package sqlstore

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/twigex/twigex/model"
)

var extShareSeq int

func seedExternalShare(t *testing.T, db *sql.DB, fileID string, active bool) {
	t.Helper()
	extShareSeq++
	now := time.Now().Unix()
	_, err := db.Exec(
		`INSERT INTO external_share
		 (id, owner, file_id, password_protected, password, share_token, active, created_at, updated_at)
		 VALUES (?, 'owner-1', ?, 0, '', ?, ?, ?, ?)`,
		fmt.Sprintf("ext-%d", extShareSeq), fileID,
		fmt.Sprintf("token-%d", extShareSeq), active, now, now,
	)
	if err != nil {
		t.Fatalf("seedExternalShare %s: %v", fileID, err)
	}
}

// A file carries the flag when it holds a live share of its own, not when it
// sits inside one.
func TestSharedFlagFollowsTheShareTables(t *testing.T) {
	db := requireDB(t)
	cleanTables(t, "file_share", "external_share", "files")
	repo := fileRepository{Db: db}

	const (
		storage = "s1"
		owner   = "owner-1"
		viewer  = "viewer-1"
	)
	insertFile(t, db, "plain", owner, "root", storage, "text/plain", false, 0)
	insertFile(t, db, "user-shared", owner, "root", storage, "text/plain", false, 0)
	insertFile(t, db, "shared-folder", owner, "root", storage, "inode/directory", true, 0)
	insertFile(t, db, "child-of-shared", owner, "shared-folder", storage, "text/plain", false, 0)
	insertFile(t, db, "expired-share", owner, "root", storage, "text/plain", false, 0)
	insertFile(t, db, "public-link", owner, "root", storage, "text/plain", false, 0)
	insertFile(t, db, "dead-link", owner, "root", storage, "text/plain", false, 0)

	seedShareRow(t, "user-shared", model.SHARE_TYPE_USER, viewer, 0, model.AccessEditor)
	seedShareRow(t, "shared-folder", model.SHARE_TYPE_USER, viewer, 0, model.AccessEditor)
	seedShareRow(t, "expired-share", model.SHARE_TYPE_USER, viewer, time.Now().Unix()-3600, model.AccessEditor)
	seedExternalShare(t, db, "public-link", true)
	seedExternalShare(t, db, "dead-link", false)

	want := map[string]bool{
		"plain":           false,
		"user-shared":     true,
		"shared-folder":   true,
		"child-of-shared": false, // inherits access, carries no share of its own
		"expired-share":   false, // the row is still there, the access is not
		"public-link":     true,
		"dead-link":       false, // deactivated link grants nothing
	}

	for id, expected := range want {
		t.Run(id, func(t *testing.T) {
			f, err := repo.Get(id)
			if err != nil {
				t.Fatalf("Get: %v", err)
			}
			if f == nil {
				t.Fatal("Get returned nothing")
			}
			if f.Shared != expected {
				t.Errorf("shared = %v, want %v", f.Shared, expected)
			}
		})
	}
}
