// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/twigex/twigex/config"
	"github.com/twigex/twigex/internal/filestore"
	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/store"
)

type fakePhotoStorageStore struct {
	store.StorageStore
	storage  *model.Storage
	storages map[string]*model.Storage
	err      error
}

func photoTestConfig(directory string) config.ConfigStore {
	return config.ConfigStore{Config: &model.ServerConfig{
		FileSettings: model.FileSettings{Directory: model.NewString(directory)},
		SqlSettings:  model.SqlSettings{QueryTimeout: model.NewInt(25)},
	}}
}

func (f *fakePhotoStorageStore) GetPrimary() (*model.Storage, error) {
	return f.storage, f.err
}

type fakeUserPhotoStore struct {
	store.UserPhotoStore
	photo     *model.UserPhoto
	deleteErr error
	deleted   bool
}

func (f *fakeUserPhotoStore) Get(context.Context, string) (*model.UserPhoto, error) {
	return f.photo, nil
}

func (f *fakeUserPhotoStore) Delete(context.Context, string, string) (bool, error) {
	if f.deleteErr != nil {
		return false, f.deleteErr
	}

	f.deleted = true
	f.photo = nil
	return true, nil
}

func (f *fakeUserPhotoStore) Replace(_ context.Context, photo model.UserPhoto) (*model.UserPhoto, error) {
	old := f.photo
	f.photo = &photo
	return old, nil
}

func (f *fakePhotoStorageStore) GetByID(id string) (*model.Storage, error) {
	if f.storages != nil {
		return f.storages[id], f.err
	}

	return f.storage, f.err
}

type fakePhotoBackend struct {
	filestore.FileBackend
	removeErr   error
	removedPath string
	writtenPath string
}

func (f *fakePhotoBackend) WriteFile(_ context.Context, photoPath string, _ io.Reader, _ int64) error {
	f.writtenPath = photoPath
	return nil
}

func (f *fakePhotoBackend) RemoveFile(_ context.Context, photoPath string) error {
	f.removedPath = photoPath
	return f.removeErr
}

func TestUploadPhotoRecordsStorageAndPhotoID(t *testing.T) {
	backend := &fakePhotoBackend{}
	photos := &fakeUserPhotoStore{}
	a := &App{
		ConfigStore: photoTestConfig(""),
		Store: store.Store{
			UserPhoto: photos,
			Storage:   &fakePhotoStorageStore{storage: &model.Storage{ID: "s1", Directory: "data"}},
		},
		FileStorageObjects: map[string]filestore.FileBackend{"s1": backend},
	}

	photoID, appErr := a.UploadPhoto(context.Background(), model.User{ID: "u1"}, strings.NewReader("photo"))
	if appErr != nil {
		t.Fatalf("UploadPhoto: %v", appErr)
	}

	if photos.photo == nil || photos.photo.PhotoID != photoID || photos.photo.StorageID != "s1" {
		t.Fatalf("photo record = %#v, returned ID = %q", photos.photo, photoID)
	}
	if backend.writtenPath != "data/photos/"+photoID {
		t.Fatalf("written path = %q", backend.writtenPath)
	}
}

func TestUploadPhotoRemovesReplacedFileFromOriginalStorage(t *testing.T) {
	oldBackend := &fakePhotoBackend{}
	newBackend := &fakePhotoBackend{}
	photos := &fakeUserPhotoStore{photo: &model.UserPhoto{UserID: "u1", PhotoID: "old-photo", StorageID: "old-storage"}}
	a := &App{
		ConfigStore: photoTestConfig(""),
		Store: store.Store{
			UserPhoto: photos,
			Storage: &fakePhotoStorageStore{
				storage: &model.Storage{ID: "new-storage", Directory: "new-data"},
				storages: map[string]*model.Storage{
					"old-storage": {ID: "old-storage", Directory: "old-data"},
				},
			},
		},
		FileStorageObjects: map[string]filestore.FileBackend{
			"old-storage": oldBackend,
			"new-storage": newBackend,
		},
	}

	photoID, appErr := a.UploadPhoto(context.Background(), model.User{ID: "u1"}, strings.NewReader("new photo"))
	if appErr != nil {
		t.Fatalf("UploadPhoto: %v", appErr)
	}

	if oldBackend.removedPath != "old-data/photos/old-photo" {
		t.Errorf("removed path = %q", oldBackend.removedPath)
	}
	if newBackend.writtenPath != "new-data/photos/"+photoID {
		t.Errorf("written path = %q", newBackend.writtenPath)
	}
}

func TestDeletePhotoUsesPhotoRecordStorage(t *testing.T) {
	backend := &fakePhotoBackend{}
	photos := &fakeUserPhotoStore{photo: &model.UserPhoto{
		UserID: "u1", PhotoID: "p1", StorageID: "s1",
	}}
	a := &App{
		ConfigStore: photoTestConfig(""),
		Store: store.Store{
			UserPhoto: photos,
			Storage:   &fakePhotoStorageStore{storage: &model.Storage{ID: "s1", Directory: "data"}},
		},
		FileStorageObjects: map[string]filestore.FileBackend{"s1": backend},
	}

	if appErr := a.DeletePhoto(context.Background(), model.User{ID: "u1", Photo: "p1"}); appErr != nil {
		t.Fatalf("DeletePhoto: %v", appErr)
	}

	if backend.removedPath != "data/photos/p1" || !photos.deleted {
		t.Fatalf("removed path = %q, record deleted = %v", backend.removedPath, photos.deleted)
	}
}

func TestDeletePhotoRemovesStoredFile(t *testing.T) {
	photos := &fakeUserPhotoStore{photo: &model.UserPhoto{UserID: "u1", PhotoID: "p1", StorageID: "s1"}}
	backend := &fakePhotoBackend{}
	a := &App{
		ConfigStore: photoTestConfig(""),
		Store: store.Store{
			UserPhoto: photos,
			Storage:   &fakePhotoStorageStore{storage: &model.Storage{ID: "s1", Directory: "data"}},
		},
		FileStorageObjects: map[string]filestore.FileBackend{"s1": backend},
	}

	if appErr := a.DeletePhoto(context.Background(), model.User{ID: "u1"}); appErr != nil {
		t.Fatalf("DeletePhoto: %v", appErr)
	}

	if backend.removedPath != "data/photos/p1" || !photos.deleted {
		t.Errorf("removed path = %q, record deleted = %v", backend.removedPath, photos.deleted)
	}
}

func TestDeletePhotoClearsRecordWhenStoredFileRemovalFails(t *testing.T) {
	photos := &fakeUserPhotoStore{photo: &model.UserPhoto{UserID: "u1", PhotoID: "p1", StorageID: "s1"}}
	backend := &fakePhotoBackend{removeErr: errors.New("storage unavailable")}
	a := &App{
		ConfigStore: photoTestConfig(""),
		Store: store.Store{
			UserPhoto: photos,
			Storage:   &fakePhotoStorageStore{storage: &model.Storage{ID: "s1", Directory: "data"}},
		},
		FileStorageObjects: map[string]filestore.FileBackend{"s1": backend},
	}

	appErr := a.DeletePhoto(context.Background(), model.User{ID: "u1"})
	if appErr != nil {
		t.Fatalf("DeletePhoto: %v", appErr)
	}

	if !photos.deleted || photos.photo != nil {
		t.Error("photo record was not deleted after file removal failed")
	}
}

func TestDeletePhotoClearsRecordWhenStorageUnavailable(t *testing.T) {
	for _, tc := range []struct {
		name    string
		storage *model.Storage
		err     error
	}{
		{name: "lookup failed", err: errors.New("database unavailable")},
		{name: "storage missing"},
		{name: "backend missing", storage: &model.Storage{ID: "s1"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			photos := &fakeUserPhotoStore{photo: &model.UserPhoto{UserID: "u1", PhotoID: "p1", StorageID: "s1"}}
			a := &App{
				ConfigStore: photoTestConfig(""),
				Store: store.Store{
					UserPhoto: photos,
					Storage:   &fakePhotoStorageStore{storage: tc.storage, err: tc.err},
				},
			}

			appErr := a.DeletePhoto(context.Background(), model.User{ID: "u1"})
			if appErr != nil {
				t.Fatalf("DeletePhoto: %v", appErr)
			}

			if !photos.deleted || photos.photo != nil {
				t.Error("photo record was not deleted without storage access")
			}
		})
	}
}

func TestDeletePhotoKeepsFileWhenRecordDeletionFails(t *testing.T) {
	photos := &fakeUserPhotoStore{
		photo:     &model.UserPhoto{UserID: "u1", PhotoID: "p1", StorageID: "s1"},
		deleteErr: errors.New("database unavailable"),
	}
	backend := &fakePhotoBackend{}
	a := &App{
		ConfigStore: photoTestConfig(""),
		Store: store.Store{
			UserPhoto: photos,
		},
		FileStorageObjects: map[string]filestore.FileBackend{"s1": backend},
	}

	appErr := a.DeletePhoto(context.Background(), model.User{ID: "u1"})
	if appErr == nil || appErr.Status != http.StatusInternalServerError {
		t.Fatalf("expected deletion failure, got %v", appErr)
	}

	if photos.deleted || photos.photo == nil || backend.removedPath != "" {
		t.Error("photo file or record changed after database deletion failed")
	}
}
