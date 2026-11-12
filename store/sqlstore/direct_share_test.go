// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package sqlstore

import (
	"context"
	"slices"
	"testing"

	"github.com/twigex/twigex/model"
)

func TestGetDirectShareTypesSeesOnlyTheItemItself(t *testing.T) {
	db := requireDB(t)
	cleanTables(t, "files", "file_share", "group_members")

	repo := &fileRepository{Db: db}
	ctx := context.Background()

	seedShareFile(t, "folder", "anna", "drive", "storage", "folder", true, 0)
	seedShareFile(t, "report", "anna", "folder", "storage", "docx", false, 0)
	seedShareFile(t, "notes", "anna", "folder", "storage", "md", false, 0)
	seedShareRow(t, "folder", model.SHARE_TYPE_USER, "ben", 0, model.AccessEditor)
	seedShareRow(t, "notes", model.SHARE_TYPE_GROUP, "group-1", 0, model.AccessViewer)
	seedGroupMember(t, db, "group-1", "ben")

	folderTypes, err := repo.GetDirectShareTypes(ctx, "ben", "folder")
	if err != nil {
		t.Fatalf("GetDirectShareTypes folder: %v", err)
	}

	if !slices.Equal(folderTypes, []int{model.SHARE_TYPE_USER}) {
		t.Errorf("want a direct user share on the folder, got %v", folderTypes)
	}

	reportTypes, err := repo.GetDirectShareTypes(ctx, "ben", "report")
	if err != nil {
		t.Fatalf("GetDirectShareTypes report: %v", err)
	}

	if len(reportTypes) != 0 {
		t.Errorf("want no direct share on a file reached through its folder, got %v", reportTypes)
	}

	notesTypes, err := repo.GetDirectShareTypes(ctx, "ben", "notes")
	if err != nil {
		t.Fatalf("GetDirectShareTypes notes: %v", err)
	}

	if !slices.Equal(notesTypes, []int{model.SHARE_TYPE_GROUP}) {
		t.Errorf("want a direct group share on the notes, got %v", notesTypes)
	}
}
