// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package sqlstore

import (
	"database/sql"
	"testing"
	"time"
)

func seedMetadata(t *testing.T, db *sql.DB, id, fields string) {
	t.Helper()
	now := time.Now().Unix()
	_, err := db.Exec(
		`INSERT INTO file_metadata (id, type, title, fields, created_at, updated_at, created_by)
		 VALUES (?, 'text', 'Title', ?, ?, ?, 'owner-1')`,
		id, fields, now, now,
	)
	if err != nil {
		t.Fatalf("seedMetadata %s: %v", id, err)
	}
}

func seedMetadataEntry(t *testing.T, db *sql.DB, id, fileID, metadataID, value string) {
	t.Helper()
	now := time.Now().Unix()
	_, err := db.Exec(
		`INSERT INTO file_metadata_entries (id, file_id, metadata_id, value, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		id, fileID, metadataID, value, now, now,
	)
	if err != nil {
		t.Fatalf("seedMetadataEntry %s: %v", id, err)
	}
}

// The scan target used to be declared once outside the loop, so a failed
// unmarshal left the last good value in place and it was appended under the
// next file's id.
func TestMetadataEntriesDoNotBleedBetweenRows(t *testing.T) {
	db := requireDB(t)
	cleanTables(t, "file_metadata_entries", "file_metadata", "files")
	repo := fileRepository{Db: db}

	seedMetadata(t, db, "meta-1", `[{"name":"field-a"}]`)
	// Named so the parseable row is read first: rows come back in file_id order,
	// and the bleed only shows when a good row precedes a bad one.
	seedMetadataEntry(t, db, "entry-1", "file-a-good", "meta-1", `{"kept":"first"}`)
	seedMetadataEntry(t, db, "entry-2", "file-b-bad", "meta-1", `{not valid json`)

	entries, err := repo.GetMetadataEntriesForFiles([]string{"file-a-good", "file-b-bad"})
	if err != nil {
		t.Fatalf("GetMetadataEntriesForFiles: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("got %d entries, want 2", len(entries))
	}

	byFile := make(map[string]any, len(entries))
	for _, e := range entries {
		byFile[e.FileID] = e.Value
	}

	if byFile["file-a-good"] == nil {
		t.Error("the parseable row lost its value")
	}
	if v := byFile["file-b-bad"]; v != nil && v != "" {
		t.Errorf("file-b-bad carries %#v, which belongs to another row", v)
	}
}
