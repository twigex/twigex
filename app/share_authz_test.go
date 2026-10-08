// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"context"
	"net/http"
	"testing"

	"github.com/twigex/twigex/config"
	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/store"
)

// Managing other people's access requires owner or Manager; a plain Editor or
// Viewer collaborator may only remove their own access (leave).

type fakeShareFileStore struct {
	store.FileStore
	file         *model.File
	actorLevel   model.AccessLevel // the non-owner caller's resolved level
	removedUser  bool
	removedGroup bool
	updatedPatch model.FileSharePatch
}

func (f *fakeShareFileStore) Get(string) (*model.File, error) { return f.file, nil }

func (f *fakeShareFileStore) ResolveEffectiveShare(context.Context, string, string) ([]model.File, error) {
	// Non-empty means the caller has (inherited) access, so HasPermission passes.
	return []model.File{{AccessLevel: f.actorLevel}}, nil
}

func (f *fakeShareFileStore) GetGrantingRoot(context.Context, string, int, string) (string, error) {
	return "root", nil
}

func (f *fakeShareFileStore) RemoveUserFromShare(string, string) error {
	f.removedUser = true
	return nil
}

func (f *fakeShareFileStore) RemoveGroupFromShare(context.Context, string, string) error {
	f.removedGroup = true
	return nil
}

func (f *fakeShareFileStore) UpdateUserPermission(p model.FileSharePatch) error {
	f.updatedPatch = p
	return nil
}

func shareApp(file *model.File) (*App, *fakeShareFileStore) {
	fs := &fakeShareFileStore{file: file, actorLevel: model.AccessEditor}
	a := &App{
		Store: store.Store{File: fs},
		ConfigStore: config.ConfigStore{
			Config: &model.ServerConfig{
				SqlSettings: model.SqlSettings{QueryTimeout: model.NewInt(25)},
			},
		},
	}
	return a, fs
}

// admin only satisfies the coarse PermissionShareFiles session gate; it does not
// bypass the owner check, matching UpdateSharePermissions.
func actor(id string) model.User {
	return model.User{ID: id, Role: model.SystemAdminRoleId}
}

func TestUnshare_NonOwnerCannotRemoveAnother(t *testing.T) {
	a, fs := shareApp(&model.File{ID: "f1", Owner: "owner"})

	appErr := a.Unshare(actor("attacker"), "f1", "victim")

	requireAppErr(t, appErr, "file.forbidden", http.StatusForbidden)
	if fs.removedUser {
		t.Fatalf("a non-owner must not be able to remove another user's access")
	}
}

func TestUnshare_CollaboratorMayLeave(t *testing.T) {
	a, fs := shareApp(&model.File{ID: "f1", Owner: "owner"})

	if appErr := a.Unshare(actor("sharee"), "f1", "sharee"); appErr != nil {
		t.Fatalf("a collaborator must be able to remove their own access, got %v", appErr)
	}
	if !fs.removedUser {
		t.Fatalf("self-leave should remove the caller's own grant")
	}
}

func TestUnshare_OwnerMayRemoveAnother(t *testing.T) {
	a, fs := shareApp(&model.File{ID: "f1", Owner: "owner"})

	if appErr := a.Unshare(actor("owner"), "f1", "victim"); appErr != nil {
		t.Fatalf("owner must be able to remove another user, got %v", appErr)
	}
	if !fs.removedUser {
		t.Fatalf("owner removal should remove the target's grant")
	}
}

func TestUnshareGroup_NonOwnerForbidden(t *testing.T) {
	a, fs := shareApp(&model.File{ID: "f1", Owner: "owner"})

	appErr := a.UnshareGroup(context.Background(), actor("sharee"), "f1", "g1")

	requireAppErr(t, appErr, "file.forbidden", http.StatusForbidden)
	if fs.removedGroup {
		t.Fatalf("a non-owner must not be able to remove a group's access")
	}
}

func TestUnshareGroup_OwnerAllowed(t *testing.T) {
	a, fs := shareApp(&model.File{ID: "f1", Owner: "owner"})

	if appErr := a.UnshareGroup(context.Background(), actor("owner"), "f1", "g1"); appErr != nil {
		t.Fatalf("owner must be able to remove a group, got %v", appErr)
	}
	if !fs.removedGroup {
		t.Fatalf("owner group removal should remove the group grant")
	}
}

func TestUnshare_ManagerMayRemoveAnother(t *testing.T) {
	a, fs := shareApp(&model.File{ID: "f1", Owner: "owner"})
	fs.actorLevel = model.AccessManager

	if appErr := a.Unshare(actor("manager"), "f1", "victim"); appErr != nil {
		t.Fatalf("a manager must be able to remove another user, got %v", appErr)
	}
	if !fs.removedUser {
		t.Fatalf("manager removal should remove the target's grant")
	}
}

func TestUnshareGroup_ManagerAllowed(t *testing.T) {
	a, fs := shareApp(&model.File{ID: "f1", Owner: "owner"})
	fs.actorLevel = model.AccessManager

	if appErr := a.UnshareGroup(context.Background(), actor("manager"), "f1", "g1"); appErr != nil {
		t.Fatalf("a manager must be able to remove a group, got %v", appErr)
	}
	if !fs.removedGroup {
		t.Fatalf("manager group removal should remove the group grant")
	}
}

func TestUpdateSharePermissions_EditsGroupRow(t *testing.T) {
	a, fs := shareApp(&model.File{ID: "f1", Owner: "owner"})

	patch := model.FileSharePatch{
		ID:          "g1",
		FileID:      "f1",
		ShareType:   model.SHARE_TYPE_GROUP,
		AccessLevel: model.AccessEditor,
	}
	if appErr := a.UpdateSharePermissions(context.Background(), actor("owner"), patch); appErr != nil {
		t.Fatalf("owner should be able to edit a group's access, got %v", appErr)
	}
	if fs.updatedPatch.ShareType != model.SHARE_TYPE_GROUP {
		t.Fatalf("update must target the group share row, got share_type=%d", fs.updatedPatch.ShareType)
	}
}
