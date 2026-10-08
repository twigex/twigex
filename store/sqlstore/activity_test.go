// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package sqlstore

import (
	"testing"

	"github.com/twigex/twigex/model"
)

func seedActivity(t *testing.T, repo *activityRepository, userID, activityType, parentID, itemID string) {
	t.Helper()
	if err := repo.Create(userID, "files", activityType, parentID, itemID, "{}"); err != nil {
		t.Fatalf("create activity: %v", err)
	}
}

func TestGetForFileFolderReturnsChildAndOwnActivity(t *testing.T) {
	db := requireDB(t)
	cleanTables(t, "activity", "users")
	repo := &activityRepository{Db: db}
	userID := seedUser(t, db)

	seedActivity(t, repo, userID, "upload", "folder-1", "child-1")
	seedActivity(t, repo, userID, "rename", "root", "folder-1")
	seedActivity(t, repo, userID, "upload", "folder-2", "child-2")

	got, err := repo.GetForFile(model.File{ID: "folder-1", IsFolder: true})
	if err != nil {
		t.Fatalf("GetForFile: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d activities, want 2", len(got))
	}
}

func TestGetForFileTreatsQuotedIDAsData(t *testing.T) {
	db := requireDB(t)
	cleanTables(t, "activity", "users")
	repo := &activityRepository{Db: db}
	userID := seedUser(t, db)

	seedActivity(t, repo, userID, "upload", "folder-1", "child-1")

	got, err := repo.GetForFile(model.File{ID: `x" OR "1"="1`, IsFolder: true})
	if err != nil {
		t.Fatalf("GetForFile: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("got %d activities, want 0", len(got))
	}
}
