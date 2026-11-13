// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/twigex/twigex/config"
	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/store"
)

type fakeAccessFileStore struct {
	store.FileStore
	files          map[string]model.File
	shares         map[string]model.AccessLevel
	directShares   map[string][]int
	trashedByOwner map[string][]string
	unshared       []string
}

func (f *fakeAccessFileStore) GetDirectShareTypes(_ context.Context, userID string, fileID string) ([]int, error) {
	return f.directShares[userID+"/"+fileID], nil
}

func (f *fakeAccessFileStore) GetSubFilesByFolder(string, string) ([]model.File, error) {
	return []model.File{}, nil
}

func (f *fakeAccessFileStore) Trash(owner string, files, _ []model.File) error {
	for _, file := range files {
		f.trashedByOwner[owner] = append(f.trashedByOwner[owner], file.ID)
	}

	return nil
}

func (f *fakeAccessFileStore) RemoveUserFromShare(fileID string, _ string) error {
	f.unshared = append(f.unshared, fileID)

	return nil
}

type fakeAccessActivityStore struct {
	store.ActivityStore
}

func (f fakeAccessActivityStore) Create(string, string, string, string, string, string) error {
	return nil
}

type fakeAccessUserStore struct {
	store.UserStore
}

func (f fakeAccessUserStore) GetSharedUsers(string) ([]model.User, error) {
	return nil, errors.New("not needed in these tests")
}

func (f *fakeAccessFileStore) Get(id string) (*model.File, error) {
	file, found := f.files[id]
	if !found {
		return nil, nil
	}

	return &file, nil
}

func (f *fakeAccessFileStore) ResolveEffectiveShare(_ context.Context, userID string, fileID string) ([]model.File, error) {
	level, shared := f.shares[userID+"/"+fileID]
	if !shared {
		return []model.File{}, nil
	}

	return []model.File{{AccessLevel: level}}, nil
}

const (
	annaID = "anna"
	benID  = "ben"
)

func fileAccessApp() *App {
	files := &fakeAccessFileStore{
		files: map[string]model.File{
			"anna-drive":   {ID: "anna-drive", Owner: annaID, Type: "cloud#drive"},
			"ben-drive":    {ID: "ben-drive", Owner: benID, Type: "cloud#drive"},
			"anna-shared":  {ID: "anna-shared", Owner: annaID, Parent: "anna-drive", IsFolder: true},
			"anna-private": {ID: "anna-private", Owner: annaID, Parent: "anna-drive", IsFolder: true},
			"anna-report":  {ID: "anna-report", Owner: annaID, Parent: "anna-shared", Type: "docx"},
		},
		shares: map[string]model.AccessLevel{
			benID + "/anna-shared": model.AccessViewer,
			benID + "/anna-report": model.AccessViewer,
		},
		directShares:   map[string][]int{},
		trashedByOwner: map[string][]string{},
	}

	return &App{
		Store: store.Store{
			File:     files,
			Activity: fakeAccessActivityStore{},
			User:     fakeAccessUserStore{},
		},
		ConfigStore: config.ConfigStore{
			Config: &model.ServerConfig{
				SqlSettings: model.SqlSettings{QueryTimeout: model.NewInt(25)},
			},
		},
	}
}

func plainUser(id string) model.User {
	return model.User{ID: id, Role: model.SystemUserRoleId}
}

func TestDriveOwnerGetsFullAccess(t *testing.T) {
	a := fileAccessApp()

	drive, appErr := a.HasPermission("anna-drive", plainUser(annaID))
	if appErr != nil {
		t.Fatalf("HasPermission: %v", appErr)
	}

	if drive.AccessLevel != model.AccessManager {
		t.Errorf("want manager access on your own drive, got %v", drive.AccessLevel)
	}
}

func TestSomeoneElsesDriveIsForbidden(t *testing.T) {
	a := fileAccessApp()

	_, appErr := a.HasPermission("anna-drive", plainUser(benID))
	requireAppErr(t, appErr, "file.forbidden", http.StatusForbidden)
}

func TestListingSomeoneElsesDriveIsForbidden(t *testing.T) {
	a := fileAccessApp()

	_, appErr := a.GetUserFiles(plainUser(benID), "anna-drive")
	requireAppErr(t, appErr, "file.forbidden", http.StatusForbidden)
}

func TestSharedDriveStillOpensForTheRecipient(t *testing.T) {
	a := fileAccessApp()
	files := a.Store.File.(*fakeAccessFileStore)
	files.shares[benID+"/anna-drive"] = model.AccessViewer

	drive, appErr := a.HasPermission("anna-drive", plainUser(benID))
	if appErr != nil {
		t.Fatalf("HasPermission: %v", appErr)
	}

	if drive.AccessLevel != model.AccessViewer {
		t.Errorf("want viewer access from the share, got %v", drive.AccessLevel)
	}
}

func TestCreatingFoldersNeedsEditAccess(t *testing.T) {
	cases := []struct {
		name     string
		parentID string
	}{
		{"someone else's drive", "anna-drive"},
		{"a folder shared as viewer", "anna-shared"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			a := fileAccessApp()

			_, appErr := a.CreateFolder(plainUser(benID), "New", tc.parentID)
			requireAppErr(t, appErr, "file.forbidden", http.StatusForbidden)
		})
	}
}

func TestRenamingNeedsEditAccess(t *testing.T) {
	a := fileAccessApp()

	appErr := a.Rename(plainUser(benID), "anna-report", "Renamed.docx")
	requireAppErr(t, appErr, "file.forbidden", http.StatusForbidden)
}

func TestRenamingSomeoneElsesDriveIsForbidden(t *testing.T) {
	a := fileAccessApp()

	appErr := a.Rename(plainUser(benID), "anna-drive", "Mine now")
	requireAppErr(t, appErr, "file.forbidden", http.StatusForbidden)
}

func TestDeletingYourOwnDriveIsRefused(t *testing.T) {
	a := fileAccessApp()

	appErr := a.DeleteFile(plainUser(annaID), []string{"anna-drive"})
	requireAppErr(t, appErr, "file.forbidden", http.StatusForbidden)
}

func TestTrashingYourOwnDriveIsRefused(t *testing.T) {
	a := fileAccessApp()

	appErr := a.Trash(plainUser(annaID), []string{"anna-drive"})
	requireAppErr(t, appErr, "file.forbidden", http.StatusForbidden)

	files := a.Store.File.(*fakeAccessFileStore)
	if len(files.trashedByOwner) != 0 {
		t.Errorf("want the drive left alone, got %v", files.trashedByOwner)
	}
}

func TestSharingADriveIsRefused(t *testing.T) {
	a := fileAccessApp()
	owner := model.User{ID: annaID, Role: model.SystemAdminRoleId}

	_, appErr := a.Share(owner, model.FileShare{FileID: "anna-drive", Users: []string{benID}})
	requireAppErr(t, appErr, "file.share_forbidden", http.StatusForbidden)
}

func TestLinkingADriveIsRefused(t *testing.T) {
	a := fileAccessApp()
	owner := model.User{ID: annaID, Role: model.SystemAdminRoleId}

	_, appErr := a.CreateLink(context.Background(), owner, model.Request{Item: "anna-drive", AllowView: true})
	requireAppErr(t, appErr, "file.share_forbidden", http.StatusForbidden)
}

func TestMovingIntoAFolderYouCannotEditIsForbidden(t *testing.T) {
	cases := []struct {
		name   string
		target string
	}{
		{"someone else's drive", "anna-drive"},
		{"someone else's private folder", "anna-private"},
		{"a folder shared as viewer", "anna-shared"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			a := fileAccessApp()

			_, appErr := a.MoveFile(context.Background(), plainUser(benID), []string{"ben-drive"}, tc.target)
			requireAppErr(t, appErr, "file.forbidden", http.StatusForbidden)
		})
	}
}

func TestOfficeOnlyOpensDocuments(t *testing.T) {
	cases := []struct {
		name string
		file *model.File
		want bool
	}{
		{"document", &model.File{Type: "docx"}, true},
		{"folder", &model.File{IsFolder: true}, false},
		{"drive", &model.File{Type: "cloud#drive"}, false},
		{"missing", nil, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := isOfficeDocument(tc.file)

			if got != tc.want {
				t.Errorf("want %v, got %v", tc.want, got)
			}
		})
	}
}

func shareAs(a *App, userID string, fileID string, level model.AccessLevel) {
	files := a.Store.File.(*fakeAccessFileStore)
	files.shares[userID+"/"+fileID] = level
}

func TestEditorDeleteGoesToTheOwnersTrash(t *testing.T) {
	a := fileAccessApp()
	shareAs(a, benID, "anna-report", model.AccessEditor)

	appErr := a.Trash(plainUser(benID), []string{"anna-report"})
	if appErr != nil {
		t.Fatalf("Trash: %v", appErr)
	}

	files := a.Store.File.(*fakeAccessFileStore)
	trashed := files.trashedByOwner[annaID]

	if len(trashed) != 1 || trashed[0] != "anna-report" {
		t.Errorf("want the file in Anna's trash, got %v", files.trashedByOwner)
	}
}

func TestTrashingAFileYouCannotReachIsForbidden(t *testing.T) {
	a := fileAccessApp()

	appErr := a.Trash(plainUser(benID), []string{"anna-private"})
	requireAppErr(t, appErr, "file.forbidden", http.StatusForbidden)

	files := a.Store.File.(*fakeAccessFileStore)
	if len(files.trashedByOwner) != 0 {
		t.Errorf("want nothing trashed, got %v", files.trashedByOwner)
	}
}

func TestTrashingAnItemAlreadyInTheTrashIsNotFound(t *testing.T) {
	a := fileAccessApp()
	files := a.Store.File.(*fakeAccessFileStore)

	report := files.files["anna-report"]
	report.DeletedAt = 1
	files.files["anna-report"] = report

	appErr := a.Trash(plainUser(annaID), []string{"anna-report"})
	requireAppErr(t, appErr, "file.not_found", http.StatusNotFound)
}

func TestViewerDeleteIsRefused(t *testing.T) {
	a := fileAccessApp()

	appErr := a.Trash(plainUser(benID), []string{"anna-report"})
	requireAppErr(t, appErr, "file.forbidden", http.StatusForbidden)

	files := a.Store.File.(*fakeAccessFileStore)
	if len(files.trashedByOwner) != 0 || len(files.unshared) != 0 {
		t.Errorf("want nothing changed, got trash %v and unshared %v", files.trashedByOwner, files.unshared)
	}
}

func TestDeletingADirectShareRemovesItFromSharedItems(t *testing.T) {
	a := fileAccessApp()
	files := a.Store.File.(*fakeAccessFileStore)
	files.directShares[benID+"/anna-shared"] = []int{model.SHARE_TYPE_USER}

	appErr := a.Trash(plainUser(benID), []string{"anna-shared"})
	if appErr != nil {
		t.Fatalf("Trash: %v", appErr)
	}

	if len(files.unshared) != 1 || files.unshared[0] != "anna-shared" {
		t.Errorf("want Ben taken off the share, got %v", files.unshared)
	}

	if len(files.trashedByOwner) != 0 {
		t.Errorf("want Anna's folder left alone, got %v", files.trashedByOwner)
	}
}

func TestDeletingAGroupShareIsRefused(t *testing.T) {
	a := fileAccessApp()
	files := a.Store.File.(*fakeAccessFileStore)
	files.directShares[benID+"/anna-shared"] = []int{model.SHARE_TYPE_GROUP}

	appErr := a.Trash(plainUser(benID), []string{"anna-shared"})
	requireAppErr(t, appErr, "file.forbidden", http.StatusForbidden)
}

func TestEditorCannotDeletePermanently(t *testing.T) {
	a := fileAccessApp()
	shareAs(a, benID, "anna-report", model.AccessEditor)

	appErr := a.DeleteFile(plainUser(benID), []string{"anna-report"})
	requireAppErr(t, appErr, "file.forbidden", http.StatusForbidden)
}

func (f *fakeAccessFileStore) Restore(string, []string) error {
	return nil
}

type countingActivityStore struct {
	store.ActivityStore
	recorded int
}

func (c *countingActivityStore) Create(string, string, string, string, string, string) error {
	c.recorded++

	return nil
}

func TestRestoringSomeoneElsesFileRecordsNothing(t *testing.T) {
	a := fileAccessApp()
	activity := &countingActivityStore{}
	a.Store.Activity = activity

	appErr := a.Restore(plainUser(benID), []string{"anna-report"})
	if appErr != nil {
		t.Fatalf("Restore: %v", appErr)
	}

	if activity.recorded != 0 {
		t.Errorf("want no activity for a file Ben does not own, got %d", activity.recorded)
	}
}
