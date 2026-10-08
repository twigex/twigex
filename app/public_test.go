// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/store"
)

type fakePublicShareStore struct {
	store.ShareStore
	link *model.ExternalShare
}

func (f *fakePublicShareStore) GetByToken(_ context.Context, _ string) (*model.ExternalShare, error) {
	return f.link, nil
}

func (f *fakePublicShareStore) IncrementDownloaded(_ context.Context, _ string) error { return nil }

type fakePublicFileStore struct {
	store.FileStore
	files   map[string]*model.File  // id -> file, backs Find
	subtree map[string][]model.File // folderID -> descendants, backs GetSubFilesByFolder
}

func (f *fakePublicFileStore) Get(id string) (*model.File, error) {
	return f.files[id], nil
}

func (f *fakePublicFileStore) GetSubFilesByFolder(folderID, _ string) ([]model.File, error) {
	subs := f.subtree[folderID]
	return subs, nil
}

type fakePublicUserStore struct {
	store.UserStore
	user *model.User
}

func (f *fakePublicUserStore) Get(_ string) (*model.User, error) {
	return f.user, nil
}

// newPublicApp wires an App with fakes for a single share. owner may be nil to
// skip the async download notification (its goroutine would otherwise reach
// unfaked stores).
func newPublicApp(link *model.ExternalShare, root *model.File, subtree []model.File, owner *model.User) *App {
	files := map[string]*model.File{}
	subMap := map[string][]model.File{}
	if root != nil {
		files[root.ID] = root
		subMap[root.ID] = subtree
	}
	return &App{Store: store.Store{
		Share: &fakePublicShareStore{link: link},
		File:  &fakePublicFileStore{files: files, subtree: subMap},
		User:  &fakePublicUserStore{user: owner},
	}}
}

func requireAppErr(t *testing.T, appErr *model.AppError, wantID string, wantStatus int) {
	t.Helper()
	if appErr == nil {
		t.Fatalf("expected error %q, got nil", wantID)
	}
	if appErr.ID != wantID {
		t.Fatalf("expected error ID %q, got %q", wantID, appErr.ID)
	}
	if appErr.Status != wantStatus {
		t.Fatalf("expected status %d for %q, got %d", wantStatus, wantID, appErr.Status)
	}
}

// folderShareFixture builds a folder share "tok" -> folder "folder-1" whose live
// subtree contains a direct child, a nested descendant, and a trashed child.
func folderShareFixture() (*model.ExternalShare, *model.File, []model.File) {
	link := &model.ExternalShare{
		ShareToken:    "tok",
		FileID:        "folder-1",
		Owner:         "owner-1",
		Active:        true,
		AllowView:     true,
		AllowDownload: true,
	}
	root := &model.File{ID: "folder-1", IsFolder: true, Storage: "s1", Owner: "owner-1"}
	subtree := []model.File{
		{ID: "child-1", Parent: "folder-1", Storage: "s1", DisplayName: "a.txt"},
		{ID: "sub-1", Parent: "folder-1", Storage: "s1", IsFolder: true},
		{ID: "nested-1", Parent: "sub-1", Storage: "s1", DisplayName: "deep.txt"},
		{ID: "trashed-1", Parent: "folder-1", Storage: "s1", DisplayName: "gone.txt", DeletedAt: 123},
	}
	return link, root, subtree
}

// publicTarget is the shared gate for view, download and office.

func TestPublicTarget_RejectsFileOutsideSharedFolder(t *testing.T) {
	link, root, subtree := folderShareFixture()
	a := newPublicApp(link, root, subtree, nil)

	// "outsider" is a real file elsewhere, not part of this share's subtree.
	_, _, appErr := a.publicTarget(context.Background(), "tok", "outsider", "")
	requireAppErr(t, appErr, "link.not_found", http.StatusNotFound)
}

func TestPublicTarget_AllowsDirectChild(t *testing.T) {
	link, root, subtree := folderShareFixture()
	a := newPublicApp(link, root, subtree, nil)

	_, target, appErr := a.publicTarget(context.Background(), "tok", "child-1", "")
	if appErr != nil {
		t.Fatalf("expected access to child-1, got %v", appErr)
	}
	if target.ID != "child-1" {
		t.Fatalf("expected target child-1, got %q", target.ID)
	}
}

func TestPublicTarget_AllowsNestedDescendant(t *testing.T) {
	link, root, subtree := folderShareFixture()
	a := newPublicApp(link, root, subtree, nil)

	_, target, appErr := a.publicTarget(context.Background(), "tok", "nested-1", "")
	if appErr != nil {
		t.Fatalf("expected access to nested-1, got %v", appErr)
	}
	if target.ID != "nested-1" {
		t.Fatalf("expected target nested-1, got %q", target.ID)
	}
}

func TestPublicTarget_RejectsTrashedChild(t *testing.T) {
	link, root, subtree := folderShareFixture()
	a := newPublicApp(link, root, subtree, nil)

	// trashed-1 is inside the subtree but soft-deleted; must not be reachable.
	_, _, appErr := a.publicTarget(context.Background(), "tok", "trashed-1", "")
	requireAppErr(t, appErr, "link.not_found", http.StatusNotFound)
}

func TestPublicTarget_RejectsLiveFileUnderTrashedFolder(t *testing.T) {
	link := &model.ExternalShare{
		ShareToken: "tok", FileID: "folder-1", Owner: "owner-1",
		Active: true, AllowView: true, AllowDownload: true,
	}
	root := &model.File{ID: "folder-1", IsFolder: true, Storage: "s1", Owner: "owner-1"}
	// sub-trash is trashed but stranded is still live under it. R1 prevents this
	// shape at the source; this verifies the resolver rejects it as a backstop.
	subtree := []model.File{
		{ID: "sub-trash", Parent: "folder-1", Storage: "s1", IsFolder: true, DeletedAt: 123},
		{ID: "stranded", Parent: "sub-trash", Storage: "s1"},
	}
	a := newPublicApp(link, root, subtree, nil)

	_, _, appErr := a.publicTarget(context.Background(), "tok", "stranded", "")
	requireAppErr(t, appErr, "link.not_found", http.StatusNotFound)
}

func TestPublicTarget_SingleFileShare_RejectsForeignChild(t *testing.T) {
	link := &model.ExternalShare{
		ShareToken: "tok", FileID: "file-1", Owner: "owner-1",
		Active: true, AllowView: true, AllowDownload: true,
	}
	root := &model.File{ID: "file-1", IsFolder: false, Storage: "s1", Owner: "owner-1"}
	a := newPublicApp(link, root, nil, nil)

	// A single-file share must not expose any other file by id.
	_, _, appErr := a.publicTarget(context.Background(), "tok", "some-other-file", "")
	requireAppErr(t, appErr, "link.not_found", http.StatusNotFound)
}

func TestPublicTarget_SingleFileShare_AllowsTheFile(t *testing.T) {
	link := &model.ExternalShare{
		ShareToken: "tok", FileID: "file-1", Owner: "owner-1",
		Active: true, AllowView: true, AllowDownload: true,
	}
	root := &model.File{ID: "file-1", IsFolder: false, Storage: "s1", Owner: "owner-1"}
	a := newPublicApp(link, root, nil, nil)

	for _, childID := range []string{"", "file-1"} {
		_, target, appErr := a.publicTarget(context.Background(), "tok", childID, "")
		if appErr != nil {
			t.Fatalf("childID=%q: expected the shared file, got %v", childID, appErr)
		}
		if target.ID != "file-1" {
			t.Fatalf("childID=%q: expected file-1, got %q", childID, target.ID)
		}
	}
}

func TestPublicTarget_PasswordRequiredWithoutToken(t *testing.T) {
	link, root, subtree := folderShareFixture()
	link.PasswordProtected = true
	a := newPublicApp(link, root, subtree, nil)

	_, _, appErr := a.publicTarget(context.Background(), "tok", "child-1", "")
	requireAppErr(t, appErr, "link.password_required", http.StatusUnauthorized)
}

func TestPublicTarget_ExpiredLink(t *testing.T) {
	link, root, subtree := folderShareFixture()
	link.Expiration = time.Now().Add(-time.Hour).Unix()
	a := newPublicApp(link, root, subtree, nil)

	_, _, appErr := a.publicTarget(context.Background(), "tok", "child-1", "")
	requireAppErr(t, appErr, "link.expired", http.StatusNotFound)
}

func TestPublicDownloadTarget_RejectsFileOutsideSharedFolder(t *testing.T) {
	link, root, subtree := folderShareFixture()
	a := newPublicApp(link, root, subtree, nil)

	// The core abuse case: guess another file's id and try to download it.
	_, appErr := a.PublicDownloadTarget(context.Background(), "tok", "outsider", "")
	requireAppErr(t, appErr, "link.not_found", http.StatusNotFound)
}

func TestPublicDownloadTarget_AllowsSubtreeChild(t *testing.T) {
	link, root, subtree := folderShareFixture()
	a := newPublicApp(link, root, subtree, nil)

	target, appErr := a.PublicDownloadTarget(context.Background(), "tok", "child-1", "")
	if appErr != nil {
		t.Fatalf("expected download of child-1, got %v", appErr)
	}
	if target.ID != "child-1" {
		t.Fatalf("expected target child-1, got %q", target.ID)
	}
}

func TestPublicDownloadTarget_ForbiddenWhenDownloadDisabled(t *testing.T) {
	link, root, subtree := folderShareFixture()
	link.AllowDownload = false // view-only link
	a := newPublicApp(link, root, subtree, nil)

	// Even a legitimate subtree file must not be downloadable on a view-only link.
	_, appErr := a.PublicDownloadTarget(context.Background(), "tok", "child-1", "")
	requireAppErr(t, appErr, "link.download_forbidden", http.StatusForbidden)
}

func TestPublicDownloadTarget_DownloadLimitReached(t *testing.T) {
	link, root, subtree := folderShareFixture()
	link.MaxDownloads = 2
	link.Downloaded = 2
	a := newPublicApp(link, root, subtree, nil)

	_, appErr := a.PublicDownloadTarget(context.Background(), "tok", "child-1", "")
	requireAppErr(t, appErr, "link.download_limit", http.StatusForbidden)
}
