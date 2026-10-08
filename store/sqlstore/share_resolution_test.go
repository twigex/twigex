// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package sqlstore

// The share read paths, over both shapes a grant takes in file_share: a row on
// every descendant (fanned out), or a single row on the share root that
// descendants inherit (collapsed). Both must resolve to the same permission, so
// the cases below cover each shape rather than one of them.
//
// Covered here: GetFilePermissions (single-file effective permission, the core
// of HasPermission), GetByParent (folder listing), GetShared (shared-with-me)
// and GetBreadcrumbs.

import (
	"context"
	"slices"
	"testing"
	"time"

	"github.com/twigex/twigex/model"
)

func TestFindByParentID_CurrentContract(t *testing.T) {
	requireDB(t)
	repo := fileRepository{Db: testDB}
	const (
		storage = "s1"
		viewer  = "user-viewer"
		owner   = "user-owner"
	)

	t.Run("owned children and per-child shared rows; excludes deleted and expired", func(t *testing.T) {
		cleanTables(t, "files", "file_share", "group_members")
		seedShareFile(t, "folder", owner, "root", storage, "inode/directory", true, 0)
		seedShareFile(t, "shared-child", owner, "folder", storage, "text/plain", false, 0)
		seedShareFile(t, "deleted-child", owner, "folder", storage, "text/plain", false, 123)
		seedShareFile(t, "expired-child", owner, "folder", storage, "text/plain", false, 0)
		seedShareFile(t, "owned-child", viewer, "folder", storage, "text/plain", false, 0)

		// Eager fan-out: one user share row per descendant, parent = immediate parent.
		seedShareRow(t, "shared-child", model.SHARE_TYPE_USER, viewer, 0, model.AccessEditor)
		seedShareRow(t, "deleted-child", model.SHARE_TYPE_USER, viewer, 0, model.AccessEditor)
		seedShareRow(t, "expired-child", model.SHARE_TYPE_USER, viewer, time.Now().Unix()-3600, model.AccessEditor)

		folder := model.File{ID: "folder", Type: "inode/directory"}
		got, err := repo.GetByParent(folder, viewer)
		if err != nil {
			t.Fatalf("GetByParent: %v", err)
		}
		ids := idSet(got)

		for _, id := range []string{"owned-child", "shared-child"} {
			if !ids[id] {
				t.Errorf("expected %q in listing, missing (%v)", id, ids)
			}
		}
		for _, id := range []string{"deleted-child", "expired-child", "folder"} {
			if ids[id] {
				t.Errorf("did not expect %q in listing (%v)", id, ids)
			}
		}
	})

	t.Run("an empty folder returns a non-nil slice (marshals as [], not null)", func(t *testing.T) {
		cleanTables(t, "files", "file_share", "group_members")
		seedShareFile(t, "empty", owner, "root", storage, "inode/directory", true, 0)

		got, err := repo.GetByParent(model.File{ID: "empty", Type: "inode/directory"}, owner)
		if err != nil {
			t.Fatalf("GetByParent: %v", err)
		}
		if got == nil {
			t.Fatalf("empty folder must return a non-nil empty slice, got %v", got)
		}
	})
}

// TestGetShared covers the ACL shared-with-me view: every share row reachable
// by the user (directly or via a group), independent of parent. In the
// collapsed model only share roots carry rows, so those are exactly what
// surfaces.
func TestGetShared(t *testing.T) {
	requireDB(t)
	repo := fileRepository{Db: testDB}
	const (
		storage = "s1"
		viewer  = "user-viewer"
		owner   = "user-owner"
	)

	t.Run("returns the user's share roots, not unshared or others' files", func(t *testing.T) {
		cleanTables(t, "files", "file_share", "group_members", "favorites", "file_metadata_entries")
		seedShareFile(t, "folder", owner, "root", storage, "inode/directory", true, 0)
		seedShareFile(t, "deepfile", owner, "otherFolder", storage, "text/plain", false, 0)
		seedShareFile(t, "unshared", owner, "root", storage, "text/plain", false, 0)
		seedShareFile(t, "groupfile", owner, "root", storage, "text/plain", false, 0)

		seedShareRow(t, "folder", model.SHARE_TYPE_USER, viewer, 0, model.AccessEditor)
		seedShareRow(t, "deepfile", model.SHARE_TYPE_USER, viewer, 0, model.AccessEditor)
		seedGroupMember(t, testDB, "g1", viewer)
		seedShareRow(t, "groupfile", model.SHARE_TYPE_GROUP, "g1", 0, model.AccessEditor)

		got, err := repo.GetShared(viewer)
		if err != nil {
			t.Fatalf("GetShared: %v", err)
		}
		ids := idSet(got)
		for _, id := range []string{"folder", "deepfile", "groupfile"} {
			if !ids[id] {
				t.Errorf("expected %q in shared-with-me (%v)", id, ids)
			}
		}
		if ids["unshared"] {
			t.Errorf("unshared file must not appear (%v)", ids)
		}
	})

	t.Run("deleted and expired shares are excluded", func(t *testing.T) {
		cleanTables(t, "files", "file_share", "group_members", "favorites", "file_metadata_entries")
		seedShareFile(t, "deleted", owner, "root", storage, "text/plain", false, 999)
		seedShareFile(t, "expired", owner, "root", storage, "text/plain", false, 0)
		seedShareRow(t, "deleted", model.SHARE_TYPE_USER, viewer, 0, model.AccessEditor)
		seedShareRow(t, "expired", model.SHARE_TYPE_USER, viewer, time.Now().Unix()-3600, model.AccessEditor)

		got, err := repo.GetShared(viewer)
		if err != nil {
			t.Fatalf("GetShared: %v", err)
		}
		if len(got) != 0 {
			t.Errorf("deleted/expired shares should be excluded, got %v", idSet(got))
		}
	})

	t.Run("shared directly and via a group appears once, most-permissive", func(t *testing.T) {
		cleanTables(t, "files", "file_share", "group_members", "favorites", "file_metadata_entries")
		seedShareFile(t, "dup", owner, "root", storage, "text/plain", false, 0)
		seedGroupMember(t, testDB, "g1", viewer)
		seedGroupMember(t, testDB, "g2", viewer)
		seedShareRow(t, "dup", model.SHARE_TYPE_USER, viewer, 0, model.AccessViewer)
		seedShareRow(t, "dup", model.SHARE_TYPE_GROUP, "g1", 0, model.AccessEditor)
		seedShareRow(t, "dup", model.SHARE_TYPE_GROUP, "g2", 0, model.AccessManager)

		got, err := repo.GetShared(viewer)
		if err != nil {
			t.Fatalf("GetShared: %v", err)
		}
		if len(got) != 1 {
			t.Fatalf("want one entry across direct + two groups, got %d (%v)", len(got), idSet(got))
		}
		if got[0].AccessLevel != model.AccessManager {
			t.Errorf("want most-permissive (manager), got %v", got[0].AccessLevel)
		}
	})

	t.Run("most-permissive expiration: never (0) wins over a future expiry", func(t *testing.T) {
		cleanTables(t, "files", "file_share", "group_members", "favorites", "file_metadata_entries")
		seedShareFile(t, "f", owner, "root", storage, "text/plain", false, 0)
		seedGroupMember(t, testDB, "g1", viewer)
		seedShareRow(t, "f", model.SHARE_TYPE_USER, viewer, time.Now().Unix()+3600, model.AccessViewer)
		seedShareRow(t, "f", model.SHARE_TYPE_GROUP, "g1", 0, model.AccessViewer)

		got, err := repo.GetShared(viewer)
		if err != nil {
			t.Fatalf("GetShared: %v", err)
		}
		if len(got) != 1 || got[0].Expiration != 0 {
			t.Fatalf("a never-expiring grant should win; got %+v", got)
		}
	})
}

// ResolveEffectiveShare must agree with GetFilePermissions when the grant sits
// on the file itself, and resolve by walking up when the grant is only on an
// ancestor, which is the collapsed shape where descendants carry no row.
func TestResolveEffectiveShare(t *testing.T) {
	requireDB(t)
	repo := fileRepository{Db: testDB}
	ctx := context.Background()
	const (
		storage = "s1"
		viewer  = "user-viewer"
		owner   = "user-owner"
	)

	t.Run("grant on the file itself resolves (parity with GetFilePermissions)", func(t *testing.T) {
		cleanTables(t, "files", "file_share", "group_members")
		seedShareFile(t, "f1", owner, "root", storage, "text/plain", false, 0)
		seedShareRow(t, "f1", model.SHARE_TYPE_USER, viewer, 0, model.AccessEditor)

		perms, err := repo.ResolveEffectiveShare(ctx, viewer, "f1")
		if err != nil {
			t.Fatalf("ResolveEffectiveShare: %v", err)
		}
		if len(perms) != 1 {
			t.Fatalf("want 1 permission row, got %d", len(perms))
		}
		p := perms[0]
		if !p.AccessLevel.CanEdit() {
			t.Errorf("want editor, got %v", p.AccessLevel)
		}
	})

	t.Run("no share anywhere yields none", func(t *testing.T) {
		cleanTables(t, "files", "file_share", "group_members")
		seedShareFile(t, "folder", owner, "root", storage, "inode/directory", true, 0)
		seedShareFile(t, "f1", owner, "folder", storage, "text/plain", false, 0)

		perms, err := repo.ResolveEffectiveShare(ctx, viewer, "f1")
		if err != nil {
			t.Fatalf("ResolveEffectiveShare: %v", err)
		}
		if len(perms) != 0 {
			t.Fatalf("want 0 permission rows, got %d", len(perms))
		}
	})

	t.Run("inherits from parent folder when the file has no row of its own", func(t *testing.T) {
		cleanTables(t, "files", "file_share", "group_members")
		seedShareFile(t, "folder", owner, "root", storage, "inode/directory", true, 0)
		seedShareFile(t, "f1", owner, "folder", storage, "text/plain", false, 0)
		// Collapsed model: only the folder (share root) carries a row.
		seedShareRow(t, "folder", model.SHARE_TYPE_USER, viewer, 0, model.AccessEditor)

		perms, err := repo.ResolveEffectiveShare(ctx, viewer, "f1")
		if err != nil {
			t.Fatalf("ResolveEffectiveShare: %v", err)
		}
		if len(perms) != 1 || !perms[0].AccessLevel.CanEdit() {
			t.Fatalf("file should inherit edit from its shared parent, got %+v", perms)
		}
	})

	t.Run("inherits from a grandparent several levels up", func(t *testing.T) {
		cleanTables(t, "files", "file_share", "group_members")
		seedShareFile(t, "gp", owner, "root", storage, "inode/directory", true, 0)
		seedShareFile(t, "p", owner, "gp", storage, "inode/directory", true, 0)
		seedShareFile(t, "f1", owner, "p", storage, "text/plain", false, 0)
		seedShareRow(t, "gp", model.SHARE_TYPE_USER, viewer, 0, model.AccessEditor)

		perms, err := repo.ResolveEffectiveShare(ctx, viewer, "f1")
		if err != nil {
			t.Fatalf("ResolveEffectiveShare: %v", err)
		}
		if len(perms) != 1 || !perms[0].AccessLevel.CanEdit() {
			t.Fatalf("file should inherit from grandparent, got %+v", perms)
		}
	})

	t.Run("a deleted intermediate ancestor blocks inheritance", func(t *testing.T) {
		cleanTables(t, "files", "file_share", "group_members")
		seedShareFile(t, "gp", owner, "root", storage, "inode/directory", true, 0)
		seedShareFile(t, "p", owner, "gp", storage, "inode/directory", true, 999) // trashed
		seedShareFile(t, "f1", owner, "p", storage, "text/plain", false, 0)
		seedShareRow(t, "gp", model.SHARE_TYPE_USER, viewer, 0, model.AccessEditor)

		perms, err := repo.ResolveEffectiveShare(ctx, viewer, "f1")
		if err != nil {
			t.Fatalf("ResolveEffectiveShare: %v", err)
		}
		if len(perms) != 0 {
			t.Fatalf("inheritance must not cross a trashed ancestor, got %+v", perms)
		}
	})

	t.Run("an expired ancestor grant is filtered out", func(t *testing.T) {
		cleanTables(t, "files", "file_share", "group_members")
		seedShareFile(t, "folder", owner, "root", storage, "inode/directory", true, 0)
		seedShareFile(t, "f1", owner, "folder", storage, "text/plain", false, 0)
		seedShareRow(t, "folder", model.SHARE_TYPE_USER, viewer, time.Now().Unix()-3600, model.AccessEditor)

		perms, err := repo.ResolveEffectiveShare(ctx, viewer, "f1")
		if err != nil {
			t.Fatalf("ResolveEffectiveShare: %v", err)
		}
		if len(perms) != 0 {
			t.Fatalf("expired inherited grant should yield none, got %+v", perms)
		}
	})

	t.Run("inherits a group share on an ancestor", func(t *testing.T) {
		cleanTables(t, "files", "file_share", "group_members")
		seedShareFile(t, "folder", owner, "root", storage, "inode/directory", true, 0)
		seedShareFile(t, "f1", owner, "folder", storage, "text/plain", false, 0)
		seedGroupMember(t, testDB, "g1", viewer)
		seedShareRow(t, "folder", model.SHARE_TYPE_GROUP, "g1", 0, model.AccessEditor)

		perms, err := repo.ResolveEffectiveShare(ctx, viewer, "f1")
		if err != nil {
			t.Fatalf("ResolveEffectiveShare: %v", err)
		}
		if len(perms) != 1 || !perms[0].AccessLevel.CanEdit() {
			t.Fatalf("file should inherit edit from an ancestor group share, got %+v", perms)
		}
	})

	t.Run("a never-expiring grant keeps access despite an expired one", func(t *testing.T) {
		cleanTables(t, "files", "file_share", "group_members")
		seedShareFile(t, "f1", owner, "root", storage, "text/plain", false, 0)
		seedGroupMember(t, testDB, "g1", viewer)
		seedShareRow(t, "f1", model.SHARE_TYPE_USER, viewer, 0, model.AccessEditor)
		seedShareRow(t, "f1", model.SHARE_TYPE_GROUP, "g1", time.Now().Unix()-3600, model.AccessViewer)

		perms, err := repo.ResolveEffectiveShare(ctx, viewer, "f1")
		if err != nil {
			t.Fatalf("ResolveEffectiveShare: %v", err)
		}
		if len(perms) != 1 {
			t.Fatalf("the never-expiring grant should keep access, got %+v", perms)
		}
	})

	t.Run("a trashed target yields none even with a direct share", func(t *testing.T) {
		cleanTables(t, "files", "file_share", "group_members")
		seedShareFile(t, "f1", owner, "root", storage, "text/plain", false, 999) // trashed
		seedShareRow(t, "f1", model.SHARE_TYPE_USER, viewer, 0, model.AccessEditor)

		perms, err := repo.ResolveEffectiveShare(ctx, viewer, "f1")
		if err != nil {
			t.Fatalf("ResolveEffectiveShare: %v", err)
		}
		if len(perms) != 0 {
			t.Fatalf("a trashed file must not be accessible to collaborators, got %+v", perms)
		}
	})

	t.Run("a trashed target yields none even with an inherited share", func(t *testing.T) {
		cleanTables(t, "files", "file_share", "group_members")
		seedShareFile(t, "folder", owner, "root", storage, "inode/directory", true, 0)
		seedShareFile(t, "f1", owner, "folder", storage, "text/plain", false, 999) // trashed
		seedShareRow(t, "folder", model.SHARE_TYPE_USER, viewer, 0, model.AccessEditor)

		perms, err := repo.ResolveEffectiveShare(ctx, viewer, "f1")
		if err != nil {
			t.Fatalf("ResolveEffectiveShare: %v", err)
		}
		if len(perms) != 0 {
			t.Fatalf("a trashed file must not inherit access, got %+v", perms)
		}
	})

	t.Run("most-permissive across own share and an ancestor group share", func(t *testing.T) {
		cleanTables(t, "files", "file_share", "group_members")
		seedShareFile(t, "folder", owner, "root", storage, "inode/directory", true, 0)
		seedShareFile(t, "f1", owner, "folder", storage, "text/plain", false, 0)
		// Own share grants nothing; ancestor group share grants edit+reshare.
		seedShareRow(t, "f1", model.SHARE_TYPE_USER, viewer, 0, model.AccessViewer)
		seedGroupMember(t, testDB, "g1", viewer)
		seedShareRow(t, "folder", model.SHARE_TYPE_GROUP, "g1", 0, model.AccessEditor)

		perms, err := repo.ResolveEffectiveShare(ctx, viewer, "f1")
		if err != nil {
			t.Fatalf("ResolveEffectiveShare: %v", err)
		}
		if len(perms) != 1 {
			t.Fatalf("want 1 aggregated row, got %d", len(perms))
		}
		p := perms[0]
		if !p.AccessLevel.CanEdit() {
			t.Errorf("union of own+inherited should grant editor; got %v", p.AccessLevel)
		}
	})
}

// TestGetBreadcrumbs covers the breadcrumb ancestor chain: the contiguous
// accessible prefix from a file up to its access origin, excluding folders
// above a share root (which the user was never shared) and stopping at a
// deleted ancestor. The caller appends the storage root separately.
func TestGetBreadcrumbs(t *testing.T) {
	requireDB(t)
	repo := fileRepository{Db: testDB}
	ctx := context.Background()
	const (
		storage = "s1"
		viewer  = "user-viewer"
		owner   = "user-owner"
		driveID = "drive-owner"
	)

	ids := func(files []model.File) []string {
		out := make([]string, 0, len(files))
		for _, f := range files {
			out = append(out, f.ID)
		}
		return out
	}

	t.Run("owned chain stops before the drive root", func(t *testing.T) {
		cleanTables(t, "files", "file_share", "group_members")
		seedShareFile(t, driveID, "", "0", storage, "cloud#drive", true, 0)
		seedShareFile(t, "myFolder", viewer, driveID, storage, "inode/directory", true, 0)
		seedShareFile(t, "myFile", viewer, "myFolder", storage, "text/plain", false, 0)

		got, err := repo.GetBreadcrumbs(ctx, viewer, "myFile")
		if err != nil {
			t.Fatalf("GetBreadcrumbs: %v", err)
		}
		if want := []string{"myFile", "myFolder"}; !slices.Equal(ids(got), want) {
			t.Errorf("owned chain = %v, want %v", ids(got), want)
		}
	})

	t.Run("shared chain stops at the share root (no leak of folders above)", func(t *testing.T) {
		cleanTables(t, "files", "file_share", "group_members")
		seedShareFile(t, driveID, "", "0", storage, "cloud#drive", true, 0)
		seedShareFile(t, "ownerTop", owner, driveID, storage, "inode/directory", true, 0)
		seedShareFile(t, "shareRoot", owner, "ownerTop", storage, "inode/directory", true, 0)
		seedShareFile(t, "mid", owner, "shareRoot", storage, "inode/directory", true, 0)
		seedShareFile(t, "file", owner, "mid", storage, "text/plain", false, 0)
		// Collapsed model: only the share root carries a row.
		seedShareRow(t, "shareRoot", model.SHARE_TYPE_USER, viewer, 0, model.AccessEditor)

		got, err := repo.GetBreadcrumbs(ctx, viewer, "file")
		if err != nil {
			t.Fatalf("GetBreadcrumbs: %v", err)
		}
		if want := []string{"file", "mid", "shareRoot"}; !slices.Equal(ids(got), want) {
			t.Errorf("shared chain = %v, want %v (must exclude ownerTop and drive)", ids(got), want)
		}
	})

	t.Run("fanned-out shares give the same chain as collapsed", func(t *testing.T) {
		cleanTables(t, "files", "file_share", "group_members")
		seedShareFile(t, driveID, "", "0", storage, "cloud#drive", true, 0)
		seedShareFile(t, "ownerTop", owner, driveID, storage, "inode/directory", true, 0)
		seedShareFile(t, "shareRoot", owner, "ownerTop", storage, "inode/directory", true, 0)
		seedShareFile(t, "mid", owner, "shareRoot", storage, "inode/directory", true, 0)
		seedShareFile(t, "file", owner, "mid", storage, "text/plain", false, 0)
		// Eager fan-out: every descendant of the share root carries its own row.
		seedShareRow(t, "shareRoot", model.SHARE_TYPE_USER, viewer, 0, model.AccessEditor)
		seedShareRow(t, "mid", model.SHARE_TYPE_USER, viewer, 0, model.AccessEditor)
		seedShareRow(t, "file", model.SHARE_TYPE_USER, viewer, 0, model.AccessEditor)

		got, err := repo.GetBreadcrumbs(ctx, viewer, "file")
		if err != nil {
			t.Fatalf("GetBreadcrumbs: %v", err)
		}
		if want := []string{"file", "mid", "shareRoot"}; !slices.Equal(ids(got), want) {
			t.Errorf("fanned-out chain = %v, want %v", ids(got), want)
		}
	})

	t.Run("a deleted ancestor blocks the chain", func(t *testing.T) {
		cleanTables(t, "files", "file_share", "group_members")
		seedShareFile(t, driveID, "", "0", storage, "cloud#drive", true, 0)
		seedShareFile(t, "shareRoot", owner, driveID, storage, "inode/directory", true, 0)
		seedShareFile(t, "deletedMid", owner, "shareRoot", storage, "inode/directory", true, 999)
		seedShareFile(t, "file", owner, "deletedMid", storage, "text/plain", false, 0)
		seedShareRow(t, "shareRoot", model.SHARE_TYPE_USER, viewer, 0, model.AccessEditor)

		got, err := repo.GetBreadcrumbs(ctx, viewer, "file")
		if err != nil {
			t.Fatalf("GetBreadcrumbs: %v", err)
		}
		if len(got) != 0 {
			t.Errorf("deleted ancestor should block inheritance, got %v", ids(got))
		}
	})
}

func TestMoveAllowedWithinShare(t *testing.T) {
	requireDB(t)
	repo := fileRepository{Db: testDB}
	ctx := context.Background()
	const (
		storage = "s1"
		viewer  = "user-viewer"
		owner   = "user-owner"
	)

	cleanTables(t, "files", "file_share", "group_members")
	seedShareFile(t, "shareRoot", owner, "root", storage, "inode/directory", true, 0)
	seedShareFile(t, "sub", owner, "shareRoot", storage, "inode/directory", true, 0)
	seedShareFile(t, "file", owner, "sub", storage, "text/plain", false, 0)
	seedShareFile(t, "outside", owner, "root", storage, "inode/directory", true, 0)
	seedShareRow(t, "shareRoot", model.SHARE_TYPE_USER, viewer, 0, model.AccessEditor)

	cases := []struct {
		name         string
		user, target string
		wantAllowed  bool
	}{
		{"into a folder inside the shared subtree", viewer, "sub", true},
		{"onto the share root itself", viewer, "shareRoot", true},
		{"outside the shared subtree is denied", viewer, "outside", false},
		{"user without the share is denied", "user-nobody", "sub", false},
	}
	for _, c := range cases {
		got, err := repo.MoveAllowedWithinShare(ctx, c.user, "file", c.target)
		if err != nil {
			t.Fatalf("%s: %v", c.name, err)
		}
		if got != c.wantAllowed {
			t.Errorf("%s: allowed=%v, want %v", c.name, got, c.wantAllowed)
		}
	}
}

func TestFindGrantingRoot(t *testing.T) {
	requireDB(t)
	repo := fileRepository{Db: testDB}
	ctx := context.Background()
	const (
		storage = "s1"
		viewer  = "user-viewer"
		owner   = "user-owner"
	)

	cleanTables(t, "files", "file_share", "group_members")
	seedShareFile(t, "shareRoot", owner, "root", storage, "inode/directory", true, 0)
	seedShareFile(t, "sub", owner, "shareRoot", storage, "inode/directory", true, 0)
	seedShareFile(t, "file", owner, "sub", storage, "text/plain", false, 0)
	seedShareRow(t, "shareRoot", model.SHARE_TYPE_USER, viewer, 0, model.AccessEditor)
	seedShareRow(t, "shareRoot", model.SHARE_TYPE_GROUP, "g1", 0, model.AccessEditor)

	t.Run("inherited file resolves up to the granting root", func(t *testing.T) {
		got, err := repo.GetGrantingRoot(ctx, "file", model.SHARE_TYPE_USER, viewer)
		if err != nil || got != "shareRoot" {
			t.Fatalf("want shareRoot, got %q err=%v", got, err)
		}
	})
	t.Run("a direct share resolves to itself", func(t *testing.T) {
		got, _ := repo.GetGrantingRoot(ctx, "shareRoot", model.SHARE_TYPE_USER, viewer)
		if got != "shareRoot" {
			t.Errorf("want shareRoot (self), got %q", got)
		}
	})
	t.Run("group grant resolves too", func(t *testing.T) {
		got, _ := repo.GetGrantingRoot(ctx, "file", model.SHARE_TYPE_GROUP, "g1")
		if got != "shareRoot" {
			t.Errorf("want shareRoot for group, got %q", got)
		}
	})
	t.Run("no grant returns empty", func(t *testing.T) {
		got, _ := repo.GetGrantingRoot(ctx, "file", model.SHARE_TYPE_USER, "user-nobody")
		if got != "" {
			t.Errorf("want empty, got %q", got)
		}
	})
	t.Run("nearest grant wins when file is itself re-shared", func(t *testing.T) {
		seedShareRow(t, "file", model.SHARE_TYPE_USER, viewer, 0, model.AccessEditor)
		got, _ := repo.GetGrantingRoot(ctx, "file", model.SHARE_TYPE_USER, viewer)
		if got != "file" {
			t.Errorf("want file (nearest), got %q", got)
		}
	})
}

func TestInheritedAccessLevel(t *testing.T) {
	requireDB(t)
	repo := fileRepository{Db: testDB}
	ctx := context.Background()
	const (
		storage = "s1"
		viewer  = "user-viewer"
		owner   = "user-owner"
	)

	cleanTables(t, "files", "file_share", "group_members")
	seedShareFile(t, "gp", owner, "root", storage, "inode/directory", true, 0)
	seedShareFile(t, "sub", owner, "gp", storage, "inode/directory", true, 0)
	seedShareRow(t, "gp", model.SHARE_TYPE_USER, viewer, 0, model.AccessEditor)

	t.Run("inherits an ancestor's level", func(t *testing.T) {
		got, err := repo.InheritedAccessLevel(ctx, "sub", model.SHARE_TYPE_USER, viewer)
		if err != nil {
			t.Fatalf("InheritedAccessLevel: %v", err)
		}
		if got != model.AccessEditor {
			t.Errorf("want editor inherited from gp, got %v", got)
		}
	})

	t.Run("a node's own grant is not counted as inherited", func(t *testing.T) {
		got, _ := repo.InheritedAccessLevel(ctx, "gp", model.SHARE_TYPE_USER, viewer)
		if got != 0 {
			t.Errorf("gp's own grant must not count as inherited, got %v", got)
		}
	})

	t.Run("no grant yields zero", func(t *testing.T) {
		got, _ := repo.InheritedAccessLevel(ctx, "sub", model.SHARE_TYPE_USER, "nobody")
		if got != 0 {
			t.Errorf("want 0 for no grant, got %v", got)
		}
	})
}

func TestGetSharedUsersForFile_Inherited(t *testing.T) {
	requireDB(t)
	repo := userRepository{Db: testDB}
	ctx := context.Background()
	const (
		storage = "s1"
		owner   = "user-owner"
	)

	cleanTables(t, "files", "file_share", "users")
	uid := seedUser(t, testDB)
	seedShareFile(t, "folder", owner, "root", storage, "inode/directory", true, 0)
	seedShareFile(t, "file", owner, "folder", storage, "text/plain", false, 0)
	seedShareRow(t, "folder", model.SHARE_TYPE_USER, uid, 0, model.AccessEditor)

	t.Run("inherited file lists the user, flagged inherited from the folder", func(t *testing.T) {
		got, err := repo.GetSharedUsersForFile(ctx, model.File{ID: "file", Owner: owner})
		if err != nil {
			t.Fatalf("GetSharedUsersForFile: %v", err)
		}
		if len(got) != 1 || got[0].ID != uid {
			t.Fatalf("want [%s], got %+v", uid, got)
		}
		su := got[0]
		if !su.Inherited || su.GrantedBy != "folder" {
			t.Errorf("want inherited from \"folder\"; got inherited=%v grantedBy=%q", su.Inherited, su.GrantedBy)
		}
	})

	t.Run("the share root lists the user as a direct (not inherited) share", func(t *testing.T) {
		got, err := repo.GetSharedUsersForFile(ctx, model.File{ID: "folder", Owner: owner})
		if err != nil {
			t.Fatalf("GetSharedUsersForFile: %v", err)
		}
		if len(got) != 1 || got[0].Inherited {
			t.Errorf("direct share must not be inherited; got %+v", got)
		}
	})
}
