// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"net/http"
	"testing"
	"time"

	"github.com/twigex/twigex/model"
)

// validateGuestOfficeEdit is the write gate for anonymous office editing. It
// must never let a guest save a file that (a) the token is not bound to,
// (b) the link no longer permits editing, or (c) lives outside the shared
// folder's live subtree. Reuses the fakes/helpers from public_test.go.

func editableFolderShare() (*model.ExternalShare, *model.File, []model.File) {
	link, root, subtree := folderShareFixture()
	link.AllowEdit = true
	return link, root, subtree
}

func TestValidateGuestOfficeEdit_RejectsTokenBoundToDifferentFile(t *testing.T) {
	link, root, subtree := editableFolderShare()
	a := newPublicApp(link, root, subtree, nil)

	// Token minted for child-1 replayed against child-2's WOPI endpoint.
	_, appErr := a.validateGuestOfficeEdit("tok", "child-1", "child-2")
	requireAppErr(t, appErr, "office.token_invalid", http.StatusForbidden)
}

func TestValidateGuestOfficeEdit_RejectsWhenEditNotAllowed(t *testing.T) {
	link, root, subtree := folderShareFixture() // AllowEdit defaults false
	a := newPublicApp(link, root, subtree, nil)

	// View-only (or edit revoked) link: a save must be rejected even for a
	// legitimate subtree file. Re-checked live, so revocation takes effect.
	_, appErr := a.validateGuestOfficeEdit("tok", "child-1", "child-1")
	requireAppErr(t, appErr, "office.readonly", http.StatusForbidden)
}

func TestValidateGuestOfficeEdit_RejectsWhenViewNotAllowed(t *testing.T) {
	link, root, subtree := editableFolderShare()
	link.AllowView = false
	a := newPublicApp(link, root, subtree, nil)

	_, appErr := a.validateGuestOfficeEdit("tok", "child-1", "child-1")
	requireAppErr(t, appErr, "office.readonly", http.StatusForbidden)
}

func TestValidateGuestOfficeEdit_AllowsEditableDescendant(t *testing.T) {
	link, root, subtree := editableFolderShare()
	a := newPublicApp(link, root, subtree, nil)

	file, appErr := a.validateGuestOfficeEdit("tok", "child-1", "child-1")
	if appErr != nil {
		t.Fatalf("expected edit of child-1, got %v", appErr)
	}
	if file.ID != "child-1" {
		t.Fatalf("expected file child-1, got %q", file.ID)
	}
}

func TestValidateGuestOfficeEdit_AllowsEditableNestedDescendant(t *testing.T) {
	link, root, subtree := editableFolderShare()
	a := newPublicApp(link, root, subtree, nil)

	file, appErr := a.validateGuestOfficeEdit("tok", "nested-1", "nested-1")
	if appErr != nil {
		t.Fatalf("expected edit of nested-1, got %v", appErr)
	}
	if file.ID != "nested-1" {
		t.Fatalf("expected file nested-1, got %q", file.ID)
	}
}

func TestValidateGuestOfficeEdit_RejectsFileOutsideSharedFolder(t *testing.T) {
	link, root, subtree := editableFolderShare()
	a := newPublicApp(link, root, subtree, nil)

	// A file not in the live subtree (e.g. moved out since the token was minted)
	// must not be writable, even though bound==fileID here.
	_, appErr := a.validateGuestOfficeEdit("tok", "outsider", "outsider")
	requireAppErr(t, appErr, "link.not_found", http.StatusNotFound)
}

func TestValidateGuestOfficeEdit_RejectsFolderItself(t *testing.T) {
	link, root, subtree := editableFolderShare()
	a := newPublicApp(link, root, subtree, nil)

	// The shared folder is not an editable document.
	_, appErr := a.validateGuestOfficeEdit("tok", "folder-1", "folder-1")
	requireAppErr(t, appErr, "office.token_invalid", http.StatusForbidden)
}

func TestValidateGuestOfficeEdit_RejectsTrashedDescendant(t *testing.T) {
	link, root, subtree := editableFolderShare()
	a := newPublicApp(link, root, subtree, nil)

	_, appErr := a.validateGuestOfficeEdit("tok", "trashed-1", "trashed-1")
	requireAppErr(t, appErr, "link.not_found", http.StatusNotFound)
}

func TestValidateGuestOfficeEdit_SingleFileShareAllowsTheFile(t *testing.T) {
	link := &model.ExternalShare{
		ShareToken: "tok", FileID: "file-1", Owner: "owner-1",
		Active: true, AllowView: true, AllowEdit: true,
	}
	root := &model.File{ID: "file-1", IsFolder: false, Storage: "s1", Owner: "owner-1"}
	a := newPublicApp(link, root, nil, nil)

	file, appErr := a.validateGuestOfficeEdit("tok", "file-1", "file-1")
	if appErr != nil {
		t.Fatalf("expected edit of file-1, got %v", appErr)
	}
	if file.ID != "file-1" {
		t.Fatalf("expected file-1, got %q", file.ID)
	}
}

func TestValidateGuestOfficeEdit_RejectsExpiredLink(t *testing.T) {
	link, root, subtree := editableFolderShare()
	link.Expiration = time.Now().Add(-time.Hour).Unix()
	a := newPublicApp(link, root, subtree, nil)

	_, appErr := a.validateGuestOfficeEdit("tok", "child-1", "child-1")
	requireAppErr(t, appErr, "link.expired", http.StatusNotFound)
}
