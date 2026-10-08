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

// Trash and permanent-delete must refuse when a descendant is locked (mid-edit
// or held by a running job), not skip it in the cascade.

type fakeTrashFileStore struct {
	store.FileStore
	byID    map[string]*model.File
	subtree map[string][]model.File
}

func (f *fakeTrashFileStore) Get(id string) (*model.File, error) { return f.byID[id], nil }

func (f *fakeTrashFileStore) GetByIDs(_ model.User, ids []string) ([]model.File, error) {
	out := []model.File{}
	for _, id := range ids {
		if v := f.byID[id]; v != nil {
			out = append(out, *v)
		}
	}
	return out, nil
}

func (f *fakeTrashFileStore) GetSubFilesByFolder(folderID, _ string) ([]model.File, error) {
	subs := f.subtree[folderID]
	return subs, nil
}

func lockedFile(id, parent, owner string) model.File {
	lockOwner := "someone-else"
	expires := time.Now().Unix() + 3600
	return model.File{ID: id, Parent: parent, Owner: owner, Storage: "s1", LockOwner: &lockOwner, LockExpiresAt: &expires}
}

func trashApp(byID map[string]*model.File, subtree map[string][]model.File) *App {
	return &App{Store: store.Store{
		File: &fakeTrashFileStore{byID: byID, subtree: subtree},
	}}
}

func TestTrashFile_RejectsWhenDescendantLocked(t *testing.T) {
	folder := &model.File{ID: "folder-1", Owner: "user-1", IsFolder: true, Storage: "s1"}
	child := lockedFile("child-1", "folder-1", "user-1")
	a := trashApp(
		map[string]*model.File{"folder-1": folder},
		map[string][]model.File{"folder-1": {*folder, child}},
	)

	appErr := a.Trash(model.User{ID: "user-1"}, []string{"folder-1"})
	requireAppErr(t, appErr, "file.locked", http.StatusForbidden)
}

func TestTrashFile_RejectsDirectlyLockedFile(t *testing.T) {
	locked := lockedFile("file-1", "root", "user-1")
	a := trashApp(map[string]*model.File{"file-1": &locked}, nil)

	appErr := a.Trash(model.User{ID: "user-1"}, []string{"file-1"})
	requireAppErr(t, appErr, "file.locked", http.StatusForbidden)
}

func TestEnqueueMoveFileJob_RejectsWhenDescendantLocked(t *testing.T) {
	targetDir := &model.File{ID: "target-dir", Owner: "user-1", IsFolder: true, Storage: "s2"}
	folder := &model.File{ID: "folder-1", Owner: "user-1", IsFolder: true, Storage: "s1"}
	child := lockedFile("child-1", "folder-1", "user-1")
	a := trashApp(
		map[string]*model.File{"target-dir": targetDir, "folder-1": folder},
		map[string][]model.File{"folder-1": {*folder, child}},
	)

	_, appErr := a.enqueueMoveFileJob(context.Background(), model.User{ID: "user-1"}, []string{"folder-1"}, "target-dir")
	requireAppErr(t, appErr, "file.locked", http.StatusForbidden)
}

func TestDeleteFile_RejectsWhenDescendantLocked(t *testing.T) {
	folder := &model.File{ID: "folder-1", Owner: "user-1", IsFolder: true, Storage: "s1"}
	child := lockedFile("child-1", "folder-1", "user-1")
	a := trashApp(
		map[string]*model.File{"folder-1": folder},
		map[string][]model.File{"folder-1": {*folder, child}},
	)

	appErr := a.DeleteFile(model.User{ID: "user-1"}, []string{"folder-1"})
	requireAppErr(t, appErr, "file.locked", http.StatusForbidden)
}
