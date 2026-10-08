// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package sqlstore

import (
	"context"
	"testing"

	"github.com/twigex/twigex/model"
)

// Asserts nothing about results, only that each statement is valid SQL naming
// columns the table has. GetSubFiles and GetStarred sat broken for years
// because their columns were dropped and nothing ran them.
func TestFileReadsExecute(t *testing.T) {
	db := requireDB(t)
	cleanTables(t, "file_share", "external_share", "files")
	repo := fileRepository{Db: db}
	ctx := context.Background()

	const (
		storage = "s1"
		owner   = "owner-1"
	)
	insertFile(t, db, "root-1", owner, "", storage, "cloud#drive", true, 0)
	insertFile(t, db, "folder-1", owner, "root-1", storage, "inode/directory", true, 0)
	insertFile(t, db, "file-1", owner, "folder-1", storage, "text/plain", false, 0)
	insertFile(t, db, "trashed-1", owner, "folder-1", storage, "text/plain", false, 123)

	folder := model.File{ID: "folder-1", Owner: owner, Storage: storage}
	user := model.User{ID: owner}

	reads := map[string]func() error{
		"Get":                   func() error { _, err := repo.Get("file-1"); return err },
		"GetByIDs":              func() error { _, err := repo.GetByIDs(user, []string{"file-1"}); return err },
		"GetByParent":           func() error { _, err := repo.GetByParent(folder, owner); return err },
		"GetChildren":           func() error { _, err := repo.GetChildren(ctx, "folder-1"); return err },
		"GetBreadcrumbs":        func() error { _, err := repo.GetBreadcrumbs(ctx, owner, "file-1"); return err },
		"GetSubFilesByFolder":   func() error { _, err := repo.GetSubFilesByFolder("folder-1", storage); return err },
		"GetUserRoot":           func() error { _, err := repo.GetUserRoot(ctx, owner, storage); return err },
		"GetDeleted":            func() error { _, err := repo.GetDeleted(owner); return err },
		"GetFavourites":         func() error { _, err := repo.GetFavourites(ctx, owner); return err },
		"GetFavouriteIDs":       func() error { _, err := repo.GetFavouriteIDs(owner, []string{"file-1"}); return err },
		"ResolveEffectiveShare": func() error { _, err := repo.ResolveEffectiveShare(ctx, owner, "file-1"); return err },
		"GetRecents":            func() error { _, err := repo.GetRecents(ctx, owner); return err },
		"GetStorageUsed":        func() error { _, err := repo.GetStorageUsed("user-1"); return err },
		"GetTotalStorage":       func() error { _, err := repo.GetTotalStorage(); return err },
		"GetGrantingRoot":       func() error { _, err := repo.GetGrantingRoot(ctx, "file-1", model.SHARE_TYPE_USER, owner); return err },
		"InheritedAccessLevel": func() error {
			_, err := repo.InheritedAccessLevel(ctx, "file-1", model.SHARE_TYPE_USER, owner)
			return err
		},
		"CanMoveFolder":           func() error { _, err := repo.CanMoveFolder("folder-1", "root-1"); return err },
		"MoveAllowedWithinShare":  func() error { _, err := repo.MoveAllowedWithinShare(ctx, owner, "file-1", "root-1"); return err },
		"GetShared":               func() error { _, err := repo.GetShared(owner); return err },
		"GetFavourite":            func() error { _, err := repo.GetFavourite(owner, "file-1"); return err },
		"GetMetadataIDForFile":    func() error { _, err := repo.GetMetadataIDForFile("file-1"); return err },
		"GetAllMetadata":          func() error { _, err := repo.GetAllMetadata(); return err },
		"GetMetadataByID":         func() error { _, err := repo.GetMetadataByID("meta-1"); return err },
		"GetMetadataEntriesFiles": func() error { _, err := repo.GetMetadataEntriesForFiles([]string{"file-1"}); return err },
	}

	for name, run := range reads {
		t.Run(name, func(t *testing.T) {
			if err := run(); err != nil {
				t.Errorf("%s: %v", name, err)
			}
		})
	}
}
