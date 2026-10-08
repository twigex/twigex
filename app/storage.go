// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"context"
	"net/http"
	"os"
	"path"

	"github.com/twigex/twigex/crypto"
	"github.com/twigex/twigex/internal/filestore"
	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/tlog"
)

func (a *App) GetStorages(user model.User) ([]model.Storage, *model.AppError) {
	if user.Role != model.SystemAdminRoleId {
		return nil, model.NewAppError("session.not_authorized", http.StatusForbidden)
	}

	storages, err := a.Store.Storage.GetAll()
	if err != nil {
		return nil, model.NewAppError("storage.get_all_failed", http.StatusInternalServerError)
	}

	for i := range storages {
		storages[i].SecretKey = ""
	}

	return storages, nil
}

func (a *App) CreateStorage(user model.User, storage model.Storage) (*model.Storage, *model.AppError) {
	if user.Role != model.SystemAdminRoleId {
		return nil, model.NewAppError("session.not_authorized", http.StatusForbidden)
	}

	if !a.canCreateNewStorage() {
		return nil, model.NewAppError("storage.license_required", http.StatusForbidden)
	}

	if storage.Type == model.S3Storage && storage.SecretKey != "" {
		encryptKey := *a.ConfigStore.Config.ServerSettings.AtRestEncryptKey
		encrypted, err := crypto.Encrypt(encryptKey, storage.SecretKey)
		if err != nil {
			return nil, model.NewAppError("storage.encrypt_failed", http.StatusInternalServerError)
		}

		storage.SecretKey = encrypted
	}

	created, err := a.Store.Storage.Create(&storage)
	if err != nil {
		return nil, model.NewAppError("storage.create_failed", http.StatusInternalServerError)
	}

	created.Directory = "data/storage/" + created.ID
	if err := a.Store.Storage.UpdateCredentials(created.ID, created); err != nil {
		return nil, model.NewAppError("storage.update_directory_failed", http.StatusInternalServerError)
	}

	file := model.File{
		ID:          model.NewID(),
		Owner:       "",
		Parent:      "",
		Storage:     created.ID,
		Name:        created.Label,
		DisplayName: created.Label,
		Shared:      false,
		Size:        0,
		Type:        "cloud#drive",
		IsFolder:    false,
	}

	if _, err := a.Store.File.Create(file); err != nil {
		return nil, model.NewAppError("storage.drive_create_failed", http.StatusInternalServerError)
	}

	if created.Type == model.LocalStorage {
		if err := os.MkdirAll(created.Directory, 0o700); err != nil {
			tlog.Errorw("Failed to create storage directory",
				"storage", created.ID,
				"error", err,
			)
			return nil, model.NewAppError("storage.create_failed", http.StatusInternalServerError)
		}
	}

	backend, err := filestore.NewFileBackend(filestore.FileBackendSettings{
		DriverName:        storage.DriverName(),
		Directory:         storage.Directory,
		S3AccessKeyId:     storage.AccessKey,
		S3SecretAccessKey: storage.SecretKey,
		S3Bucket:          storage.Bucket,
		S3Endpoint:        storage.Endpoint,
		S3SSL:             storage.SSL,
	})
	if err != nil {
		return nil, model.NewAppError("storage.drive_create_failed", http.StatusInternalServerError)
	}

	// FileStorageObjects is only initialized when the app is running as a server.
	// In CLI context (CreateStorage called from cmdline or Create from env), it will be nil.
	if a.FileStorageObjects != nil {
		a.FileStorageObjects[created.ID] = backend
	}

	created.SecretKey = ""

	return created, nil
}

func (a *App) UpdateStorage(ctx context.Context, user model.User, id string, req model.Storage) (*model.Storage, *model.AppError) {
	if user.Role != model.SystemAdminRoleId {
		return nil, model.NewAppError("session.not_authorized", http.StatusForbidden)
	}

	existing, err := a.Store.Storage.GetByID(id)
	if err != nil {
		return nil, model.NewAppError("storage.get_failed", http.StatusInternalServerError)
	}

	if existing == nil {
		return nil, model.NewAppError("storage.not_found", http.StatusNotFound)
	}

	existing.Label = req.Label
	existing.Description = req.Description
	existing.Endpoint = req.Endpoint
	existing.Bucket = req.Bucket
	existing.AccessKey = req.AccessKey
	existing.SSL = req.SSL

	encryptKey := *a.ConfigStore.Config.ServerSettings.AtRestEncryptKey

	if req.SecretKey != "" {
		encrypted, err := crypto.Encrypt(encryptKey, req.SecretKey)
		if err != nil {
			return nil, model.NewAppError("storage.encrypt_failed", http.StatusInternalServerError)
		}

		existing.SecretKey = encrypted
	}

	if err := a.Store.Storage.Update(id, existing); err != nil {
		return nil, model.NewAppError("storage.update_failed", http.StatusInternalServerError)
	}

	renameCtx, cancel := a.dbCtx(ctx)
	defer cancel()

	if err := a.Store.File.RenameStorageDrives(renameCtx, existing.ID, req.Label); err != nil {
		tlog.Errorw("Failed to rename storage drives", "storage", existing.ID, "error", err)
		return nil, model.NewAppError("storage.update_failed", http.StatusInternalServerError)
	}

	decrypted := ""

	if existing.Type != model.LocalStorage {
		decrypted, err = crypto.Decrypt(encryptKey, existing.SecretKey)
		if err != nil {
			tlog.Errorw("Failed to decrypt storage secret key", "storage", existing.ID, "error", err)
			return nil, model.NewAppError("storage.update_failed", http.StatusInternalServerError)
		}
	}

	backend, err := filestore.NewFileBackend(filestore.FileBackendSettings{
		DriverName:        existing.DriverName(),
		Directory:         existing.Directory,
		S3AccessKeyId:     existing.AccessKey,
		S3SecretAccessKey: decrypted,
		S3Bucket:          existing.Bucket,
		S3Endpoint:        existing.Endpoint,
		S3SSL:             existing.SSL,
	})
	if err != nil {
		return nil, model.NewAppError("storage.update_failed", http.StatusInternalServerError)
	}

	a.FileStorageObjects[existing.ID] = backend

	existing.SecretKey = ""
	return existing, nil
}

func (a *App) DeleteStorage(ctx context.Context, user model.User, id string) *model.AppError {
	if user.Role != model.SystemAdminRoleId {
		return model.NewAppError("session.not_authorized", http.StatusForbidden)
	}

	storage, err := a.Store.Storage.GetByID(id)
	if err != nil {
		return model.NewAppError("storage.get_failed", http.StatusInternalServerError)
	}

	if storage == nil {
		return model.NewAppError("storage.not_found", http.StatusNotFound)
	}

	dbCtx, cancel := a.dbCtx(ctx)
	defer cancel()

	photoCount, err := a.Store.UserPhoto.CountByStorage(dbCtx, id)
	if err != nil {
		tlog.Errorw("Failed to count photos in storage", "storage_id", id, "error", err)
		return model.NewAppError("storage.delete_failed", http.StatusInternalServerError)
	}

	if photoCount > 0 {
		return model.NewAppError("storage.delete_failed", http.StatusConflict)
	}

	if err := a.Store.File.DeleteDrive(id); err != nil {
		return model.NewAppError("storage.drive_delete_failed", http.StatusInternalServerError)
	}

	if err := a.Store.Storage.Delete(id); err != nil {
		return model.NewAppError("storage.delete_failed", http.StatusInternalServerError)
	}

	return nil
}

func (a *App) SetPrimaryStorage(user model.User, id string) *model.AppError {
	if user.Role != model.SystemAdminRoleId {
		return model.NewAppError("session.not_authorized", http.StatusForbidden)
	}

	storage, err := a.Store.Storage.GetByID(id)
	if err != nil {
		return model.NewAppError("storage.get_failed", http.StatusInternalServerError)
	}

	if storage == nil {
		return model.NewAppError("storage.not_found", http.StatusNotFound)
	}

	if err := a.Store.Storage.UpdatePrimary(id); err != nil {
		return model.NewAppError("storage.set_primary_failed", http.StatusInternalServerError)
	}

	return nil
}

func (a *App) canCreateNewStorage() bool {
	storages, err := a.Store.Storage.GetAll()
	if err != nil {
		return false
	}

	if len(storages) == 0 {
		return true
	}

	return a.Server.License.HasMultipleStorages()
}

func (a *App) BuildFilePath(storageID, id, app string) (string, error) {
	storage, err := a.Store.Storage.GetByID(storageID)
	if err != nil {
		return "", err
	}

	filePath := ""
	switch app {
	case model.AppFiles:
		filePath = path.Join(storage.Directory, "files", id)
	case model.AppChat:
		filePath = path.Join(storage.Directory, "chat", id)
	case model.AppProjects:
		filePath = path.Join(storage.Directory, "projects", id)
	}

	return filePath, nil
}

func (a *App) BuildTempFilePath(storageID, id, app string) (string, error) {
	storage, err := a.Store.Storage.GetByID(storageID)
	if err != nil {
		return "", err
	}

	filePath := ""
	switch app {
	case model.AppFiles:
		filePath = path.Join(storage.Directory, "tmp", "files", id)
	case model.AppChat:
		filePath = path.Join(storage.Directory, "tmp", "chat", id)
	case model.AppProjects:
		filePath = path.Join(storage.Directory, "tmp", "projects", id)
	}

	return filePath, nil
}
