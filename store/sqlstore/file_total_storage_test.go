// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package sqlstore

import "testing"

func TestGetTotalStorageEmptyIsZero(t *testing.T) {
	db := requireDB(t)
	cleanTables(t, "files")

	got, err := (&fileRepository{Db: db}).GetTotalStorage()
	if err != nil {
		t.Fatalf("GetTotalStorage: %v", err)
	}
	if got != 0 {
		t.Fatalf("total = %d, want 0", got)
	}
}

func TestGetTotalStorageSumsEveryFile(t *testing.T) {
	db := requireDB(t)
	cleanTables(t, "files")

	insertFile(t, db, "drive", "alice", "", "s1", "cloud#drive", true, 0)
	insertFile(t, db, "big", "alice", "drive", "s1", "video/mp4", false, 0)
	insertFile(t, db, "trashed", "bob", "drive", "s2", "text/plain", false, 1)
	mustExec(t, `UPDATE files SET size = 5000000000 WHERE id = 'big'`)
	mustExec(t, `UPDATE files SET size = 123 WHERE id = 'trashed'`)

	got, err := (&fileRepository{Db: db}).GetTotalStorage()
	if err != nil {
		t.Fatalf("GetTotalStorage: %v", err)
	}
	if want := int64(5000000123); got != want {
		t.Fatalf("total = %d, want %d", got, want)
	}
}

func TestGetStorageUsedCountsEveryFileTheOwnerHas(t *testing.T) {
	db := requireDB(t)
	cleanTables(t, "files")

	insertFile(t, db, "drive", "alice", "", "s1", "cloud#drive", true, 0)
	insertFile(t, db, "folder", "alice", "drive", "s1", "folder", true, 0)
	insertFile(t, db, "top", "alice", "drive", "s1", "text/plain", false, 0)
	insertFile(t, db, "nested", "alice", "folder", "s1", "text/plain", false, 0)
	insertFile(t, db, "trashed", "alice", "folder", "s1", "text/plain", false, 1)
	insertFile(t, db, "bobs", "bob", "drive", "s1", "text/plain", false, 0)
	mustExec(t, `UPDATE files SET size = 100 WHERE id = 'drive'`)
	mustExec(t, `UPDATE files SET size = 1 WHERE id = 'top'`)
	mustExec(t, `UPDATE files SET size = 20 WHERE id = 'nested'`)
	mustExec(t, `UPDATE files SET size = 300 WHERE id = 'trashed'`)
	mustExec(t, `UPDATE files SET size = 4000 WHERE id = 'bobs'`)

	got, err := (&fileRepository{Db: db}).GetStorageUsed("alice")
	if err != nil {
		t.Fatalf("GetStorageUsed: %v", err)
	}
	if got != 321 {
		t.Fatalf("used = %d, want 321 (top, nested and trashed files, not the drive row or bob's)", got)
	}
}
