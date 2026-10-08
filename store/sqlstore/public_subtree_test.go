// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package sqlstore

import (
	"database/sql"
	"testing"
)

func seedSubtreeFile(t *testing.T, db *sql.DB, id, parent, storage string, isFolder bool, deletedAt int64) {
	t.Helper()
	ftype := "text/plain"
	if isFolder {
		ftype = "inode/directory"
	}
	insertFile(t, db, id, "owner-1", parent, storage, ftype, isFolder, deletedAt)
}

// TestFindSubFilesByFolderID_Contract pins the behaviour the public-share
// subtree security relies on: recursion to any depth, storage scoping, and that
// soft-deleted rows are returned by the query (the deleted_at filtering lives in
// the app resolver, not here). If this changes, publicSubtreeFile's guard and
// the fakes in app/public_test.go must be revisited.
func TestFindSubFilesByFolderID_Contract(t *testing.T) {
	db := requireDB(t)
	cleanTables(t, "files")
	repo := fileRepository{Db: db}

	const storage = "s1"

	// Shared tree: folder-1 / { child-1, sub-1 / { nested-1, trashed-1(deleted) } }
	seedSubtreeFile(t, db, "folder-1", "root", storage, true, 0)
	seedSubtreeFile(t, db, "child-1", "folder-1", storage, false, 0)
	seedSubtreeFile(t, db, "sub-1", "folder-1", storage, true, 0)
	seedSubtreeFile(t, db, "nested-1", "sub-1", storage, false, 0)
	seedSubtreeFile(t, db, "trashed-1", "sub-1", storage, false, 123)

	// A file outside the shared folder must never appear.
	seedSubtreeFile(t, db, "other-folder", "root", storage, true, 0)
	seedSubtreeFile(t, db, "outsider", "other-folder", storage, false, 0)

	// A descendant living in a different storage must be excluded by the scope.
	seedSubtreeFile(t, db, "wrong-storage", "folder-1", "s2", false, 0)

	got, err := repo.GetSubFilesByFolder("folder-1", storage)
	if err != nil {
		t.Fatalf("GetSubFilesByFolder: %v", err)
	}

	ids := map[string]bool{}
	for _, f := range got {
		ids[f.ID] = true
	}

	want := []string{"folder-1", "child-1", "sub-1", "nested-1", "trashed-1"}
	for _, id := range want {
		if !ids[id] {
			t.Errorf("expected %q in subtree, missing", id)
		}
	}

	for _, id := range []string{"outsider", "other-folder", "wrong-storage"} {
		if ids[id] {
			t.Errorf("did not expect %q in subtree", id)
		}
	}

	if len(ids) != len(want) {
		t.Errorf("subtree returned %d files, want %d (%v)", len(ids), len(want), ids)
	}
}
