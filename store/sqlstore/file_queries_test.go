// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package sqlstore

// File listing and storage queries that do not depend on share resolution.

import (
	"context"
	"testing"
	"time"

	"github.com/twigex/twigex/model"
)

// TestFindChildren covers the inherited-listing query: every live, non-drive
// child of a folder regardless of owner or per-child share rows, direct
// children only.
func TestFindChildren(t *testing.T) {
	requireDB(t)
	repo := fileRepository{Db: testDB}
	ctx := context.Background()
	const (
		storage = "s1"
		owner   = "user-owner"
		other   = "user-other"
	)

	t.Run("returns all live non-drive direct children regardless of owner", func(t *testing.T) {
		cleanTables(t, "files", "file_share", "group_members")
		seedShareFile(t, "folder", owner, "root", storage, "inode/directory", true, 0)
		seedShareFile(t, "c-owned", owner, "folder", storage, "text/plain", false, 0)
		seedShareFile(t, "c-other", other, "folder", storage, "text/plain", false, 0)
		seedShareFile(t, "c-sub", owner, "folder", storage, "inode/directory", true, 0)
		seedShareFile(t, "c-deleted", owner, "folder", storage, "text/plain", false, 555)
		seedShareFile(t, "c-drive", owner, "folder", storage, "cloud#drive", true, 0)
		// A grandchild must not appear: GetChildren is direct children only.
		seedShareFile(t, "nested", owner, "c-sub", storage, "text/plain", false, 0)

		got, err := repo.GetChildren(ctx, "folder")
		if err != nil {
			t.Fatalf("GetChildren: %v", err)
		}
		ids := idSet(got)

		for _, id := range []string{"c-owned", "c-other", "c-sub"} {
			if !ids[id] {
				t.Errorf("expected %q among children (%v)", id, ids)
			}
		}
		for _, id := range []string{"c-deleted", "c-drive", "nested", "folder"} {
			if ids[id] {
				t.Errorf("did not expect %q among children (%v)", id, ids)
			}
		}
	})
}

// TestGetRecents covers the recents feed: what the viewer owns plus what
// arrived through a reachable share root, excluding deleted/expired/unrelated.
func TestGetRecents(t *testing.T) {
	requireDB(t)
	repo := fileRepository{Db: testDB}
	ctx := context.Background()
	const (
		storage = "s1"
		viewer  = "user-viewer"
		owner   = "user-owner"
	)

	cleanTables(t, "files", "file_share", "group_members")
	now := time.Now().Unix()
	shareTime := now - 1000

	seedShareFile(t, "own", viewer, "root", storage, "text/plain", false, 0)
	seedShareFile(t, "folder", owner, "root", storage, "inode/directory", true, 0)
	seedShareFile(t, "old", owner, "folder", storage, "text/plain", false, 0)
	seedShareFile(t, "new", owner, "folder", storage, "text/plain", false, 0)
	seedShareRow(t, "folder", model.SHARE_TYPE_USER, viewer, 0, model.AccessEditor)

	// The share happened at shareTime; "old" predates it, "new" was added after.
	mustExec(t, `UPDATE file_share SET time_shared=? WHERE file_id='folder'`, shareTime)
	mustExec(t, `UPDATE files SET created_at=? WHERE id='old'`, shareTime-500)
	mustExec(t, `UPDATE files SET created_at=? WHERE id='new'`, shareTime+500)

	// Expired shared subtree: nothing under it is accessible.
	seedShareFile(t, "expfolder", owner, "root", storage, "inode/directory", true, 0)
	seedShareFile(t, "expchild", owner, "expfolder", storage, "text/plain", false, 0)
	seedShareRow(t, "expfolder", model.SHARE_TYPE_USER, viewer, now-3600, model.AccessEditor)
	seedShareFile(t, "deleted", viewer, "root", storage, "text/plain", false, 777)
	seedShareFile(t, "unrelated", owner, "root", storage, "text/plain", false, 0)

	got, err := repo.GetRecents(ctx, viewer)
	if err != nil {
		t.Fatalf("GetRecents: %v", err)
	}
	ids := idSet(got)

	// owned file, the share root itself, and a file added after sharing.
	for _, id := range []string{"own", "folder", "new"} {
		if !ids[id] {
			t.Errorf("expected %q in recents (%v)", id, ids)
		}
	}
	// pre-existing contents of the shared folder must not flood recents.
	for _, id := range []string{"old", "deleted", "unrelated", "expfolder", "expchild"} {
		if ids[id] {
			t.Errorf("did not expect %q in recents (%v)", id, ids)
		}
	}
}

// TestGetFavorites covers subtree-aware favourites: the user's favourited files
// they can access (owned or inherited), never inaccessible files nor another
// user's favourites.
func TestGetFavorites(t *testing.T) {
	requireDB(t)
	repo := fileRepository{Db: testDB}
	ctx := context.Background()
	const (
		storage = "s1"
		viewer  = "user-viewer"
		owner   = "user-owner"
	)

	cleanTables(t, "files", "file_share", "group_members", "favorites")
	seedShareFile(t, "own-fav", viewer, "root", storage, "text/plain", false, 0)
	seedShareFile(t, "folder", owner, "root", storage, "inode/directory", true, 0)
	seedShareFile(t, "inherited-fav", owner, "folder", storage, "text/plain", false, 0)
	seedShareRow(t, "folder", model.SHARE_TYPE_USER, viewer, 0, model.AccessEditor)
	seedShareFile(t, "inaccessible-fav", owner, "root", storage, "text/plain", false, 0)
	seedShareFile(t, "accessible-not-fav", owner, "folder", storage, "text/plain", false, 0)

	seedFavorite(t, viewer, "own-fav")
	seedFavorite(t, viewer, "inherited-fav")
	seedFavorite(t, viewer, "inaccessible-fav")
	// Another user's favourite of a file the viewer can access must not surface.
	seedFavorite(t, owner, "inherited-fav")

	got, err := repo.GetFavourites(ctx, viewer)
	if err != nil {
		t.Fatalf("GetFavourites: %v", err)
	}
	ids := idSet(got)

	for _, id := range []string{"own-fav", "inherited-fav"} {
		if !ids[id] {
			t.Errorf("expected %q in favourites (%v)", id, ids)
		}
	}
	for _, id := range []string{"inaccessible-fav", "accessible-not-fav"} {
		if ids[id] {
			t.Errorf("did not expect %q in favourites (%v)", id, ids)
		}
	}
}

func TestGetUserRoot(t *testing.T) {
	requireDB(t)
	repo := fileRepository{Db: testDB}
	ctx := context.Background()

	cleanTables(t, "files")
	seedShareFile(t, "root-a", "user-a", "", "s1", "cloud#drive", false, 0)

	got, err := repo.GetUserRoot(ctx, "user-a", "s1")
	if err != nil {
		t.Fatalf("GetUserRoot: %v", err)
	}
	if got == nil || got.ID != "root-a" {
		t.Fatalf("want root-a, got %v", got)
	}

	for _, c := range []struct{ user, storage string }{{"user-b", "s1"}, {"user-a", "s2"}} {
		got, err := repo.GetUserRoot(ctx, c.user, c.storage)
		if err != nil {
			t.Fatalf("GetUserRoot(%s,%s): %v", c.user, c.storage, err)
		}
		if got != nil {
			t.Errorf("want nil root for %s/%s, got %v", c.user, c.storage, got)
		}
	}
}

func TestRenameStorageDrives(t *testing.T) {
	requireDB(t)
	repo := fileRepository{Db: testDB}
	ctx := context.Background()

	cleanTables(t, "files")
	seedShareFile(t, "anchor-s1", "", "", "s1", "cloud#drive", false, 0)
	seedShareFile(t, "a-root-s1", "user-a", "", "s1", "cloud#drive", false, 0)
	seedShareFile(t, "b-root-s1", "user-b", "", "s1", "cloud#drive", false, 0)
	seedShareFile(t, "root-s2", "user-a", "", "s2", "cloud#drive", false, 0)

	if err := repo.RenameStorageDrives(ctx, "s1", "New Name"); err != nil {
		t.Fatalf("RenameStorageDrives: %v", err)
	}

	nameOf := func(id string) string {
		var n string
		if err := testDB.QueryRow("SELECT displayname FROM files WHERE id=?", id).Scan(&n); err != nil {
			t.Fatalf("name of %s: %v", id, err)
		}
		return n
	}
	for _, id := range []string{"anchor-s1", "a-root-s1", "b-root-s1"} {
		if nameOf(id) != "New Name" {
			t.Errorf("%s should be renamed, got %q", id, nameOf(id))
		}
	}
	if nameOf("root-s2") == "New Name" {
		t.Errorf("drive on a different storage must not be renamed")
	}
}

// A listing asks about the files on screen, not every favourite the user has.
func TestGetFavouriteIDsScopesToTheFilesAsked(t *testing.T) {
	db := requireDB(t)
	cleanTables(t, "files", "favorites")
	repo := fileRepository{Db: db}

	const viewer, other = "viewer-1", "other-1"
	insertFile(t, db, "on-screen", viewer, "root", "s1", "text/plain", false, 0)
	insertFile(t, db, "off-screen", viewer, "root", "s1", "text/plain", false, 0)

	seedFavorite(t, viewer, "on-screen")
	seedFavorite(t, viewer, "off-screen")
	seedFavorite(t, other, "on-screen")

	got, err := repo.GetFavouriteIDs(viewer, []string{"on-screen"})
	if err != nil {
		t.Fatalf("GetFavouriteIDs: %v", err)
	}
	if len(got) != 1 || got[0] != "on-screen" {
		t.Errorf("want [on-screen], got %v", got)
	}

	empty, err := repo.GetFavouriteIDs(viewer, nil)
	if err != nil {
		t.Fatalf("GetFavouriteIDs(nil): %v", err)
	}
	if len(empty) != 0 {
		t.Errorf("want no rows for an empty listing, got %v", empty)
	}
}
