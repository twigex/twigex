// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package sqlstore

import (
	"context"
	"testing"
)

func seedShareLink(t *testing.T, token string, maxDownloads int) {
	t.Helper()

	mustExec(t,
		`INSERT INTO external_share (id, owner, file_id, password, share_token, active, created_at, updated_at, max_downloads)
		 VALUES (?, 'alice', 'file-1', '', ?, 1, 0, 0, ?)`,
		"es-"+token, token, maxDownloads,
	)
}

func TestCountDownloadStopsAtTheLimit(t *testing.T) {
	db := requireDB(t)
	cleanTables(t, "external_share")

	repo := &shareRepository{Db: db}
	ctx := context.Background()

	seedShareLink(t, "limited", 2)

	for attempt := 1; attempt <= 3; attempt++ {
		counted, err := repo.CountDownloadIfAllowed(ctx, "limited")
		if err != nil {
			t.Fatalf("attempt %d: %v", attempt, err)
		}

		wantCounted := attempt <= 2
		if counted != wantCounted {
			t.Errorf("attempt %d: want counted %v, got %v", attempt, wantCounted, counted)
		}
	}

	if n := countWhere(t, "SELECT downloaded FROM external_share WHERE share_token = 'limited'"); n != 2 {
		t.Errorf("want 2 downloads recorded, got %d", n)
	}
}

func TestCountDownloadWithoutALimitStartsFromAnEmptyCount(t *testing.T) {
	db := requireDB(t)
	cleanTables(t, "external_share")

	repo := &shareRepository{Db: db}

	seedShareLink(t, "open", 0)

	counted, err := repo.CountDownloadIfAllowed(context.Background(), "open")
	if err != nil || !counted {
		t.Fatalf("want the download counted, got %v, %v", counted, err)
	}

	if n := countWhere(t, "SELECT downloaded FROM external_share WHERE share_token = 'open'"); n != 1 {
		t.Errorf("want a count of 1 after starting from none, got %d", n)
	}
}
