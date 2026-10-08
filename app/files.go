// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path"
	"slices"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/disintegration/imaging"
	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/tlog"
)

type movedFile struct {
	ID            string
	Type          string
	Storage       string
	TargetStorage string
}

func isThumbnailType(fileType string) bool {
	switch fileType {
	case "image/jpeg", "image/png", "image/gif", "image/bmp", "image/tiff":
		return true
	}

	return false
}

// To read progress on file move
type ProgressReader struct {
	r      io.ReadCloser
	onRead func(n int)
}

func (p *ProgressReader) Read(b []byte) (int, error) {
	n, err := p.r.Read(b)
	if n > 0 && p.onRead != nil {
		p.onRead(n)
	}

	return n, err
}

func (p *ProgressReader) Close() error {
	return p.r.Close()
}

func (a *App) enrichFiles(files []model.File, userID string) *model.AppError {
	if len(files) == 0 {
		return nil
	}

	ids := make([]string, len(files))
	for i, f := range files {
		ids[i] = f.ID
	}

	favoriteIDs, err := a.Store.File.GetFavouriteIDs(userID, ids)
	if err != nil {
		tlog.Errorw("Failed to retrieve favorite IDs", "user_id", userID, "error", err)
		return model.NewAppError("file.retrieval_failed", http.StatusInternalServerError)
	}

	favoriteSet := make(map[string]bool, len(favoriteIDs))
	for _, id := range favoriteIDs {
		favoriteSet[id] = true
	}

	metaByFile := make(map[string][]model.FileMetadataEntry)
	entries, err := a.Store.File.GetMetadataEntriesForFiles(ids)
	if err != nil {
		tlog.Errorw("Failed to retrieve file metadata", "error", err)
		return model.NewAppError("file.retrieval_failed", http.StatusInternalServerError)
	}

	for _, e := range entries {
		metaByFile[e.FileID] = append(metaByFile[e.FileID], e)
	}

	for i := range files {
		files[i].Favourite = favoriteSet[files[i].ID]
		files[i].Metadata = metaByFile[files[i].ID]
	}

	return nil
}

func (a *App) CreateMetadata(user model.User, name, metadataType string, fields interface{}) (*model.FileMetadata, *model.AppError) {
	if user.Role != model.SystemAdminRoleId {
		return nil, model.NewAppError("session.not_authorized", http.StatusForbidden)
	}

	meta, err := a.Store.File.CreateMetadata(name, metadataType, fields)
	if err != nil {
		tlog.Errorw("Failed to create file metadata",
			"name", name,
			"type", metadataType,
			"error", err,
		)
		return nil, model.NewAppError("file.metadata_create_failed", http.StatusInternalServerError)
	}

	return meta, nil
}

func (a *App) GetAllMetadata() ([]model.FileMetadata, *model.AppError) {

	meta, err := a.Store.File.GetAllMetadata()
	if err != nil {
		tlog.Errorw("Failed to retrieve file metadata",
			"error", err,
		)
		return nil, model.NewAppError("file.metadata_retrieval_failed", http.StatusInternalServerError)
	}

	return meta, nil
}

func (a *App) CreateMetadataEntry(user model.User, id string, metaid string, value string) (*model.FileMetadataEntry, *model.AppError) {

	file, appErr := a.HasPermission(id, user)
	if appErr != nil {
		return nil, appErr
	}

	if file.Owner != user.ID && !file.AccessLevel.CanEdit() {
		return nil, model.NewAppError("file.forbidden", http.StatusForbidden)
	}

	metadata, err := a.Store.File.GetMetadataByID(metaid)
	if err != nil {
		tlog.Errorw("Failed to retrieve file metadata",
			"file_id", id,
			"metadata_id", metaid,
			"error", err,
		)
		return nil, model.NewAppError("file.metadata_retrieval_failed", http.StatusInternalServerError)
	}

	fileMeta, err := a.Store.File.CreateMetadataEntry(file.ID, value, *metadata)
	if err != nil {
		tlog.Errorw("Failed to add metadata entry to file",
			"file_id", id,
			"metadata_id", metaid,
			"error", err,
		)
		return nil, model.NewAppError("file.metadata_add_failed", http.StatusInternalServerError)
	}

	a.RecordActivity(user.ID, model.AppFiles, model.ActivityFileMeta, file.Parent, file.ID, map[string]any{
		"name": file.DisplayName,
		"type": file.Type,
	})

	return fileMeta, nil
}

func (a *App) DeleteMetadataEntry(fileID string, id string, user model.User) *model.AppError {

	file, appErr := a.HasPermission(fileID, user)
	if appErr != nil {
		return appErr
	}

	if file.Owner != user.ID && !file.AccessLevel.CanEdit() {
		return model.NewAppError("file.forbidden", http.StatusForbidden)
	}

	if err := a.Store.File.DeleteMetadataEntry(file.ID, id); err != nil {
		tlog.Errorw("Failed to remove file metadata entry",
			"file_id", fileID,
			"metadata_id", id,
			"error", err,
		)
		return model.NewAppError("file.metadata_delete_failed", http.StatusInternalServerError)
	}

	return nil
}

func (a *App) DeleteMetadata(user model.User, id string) *model.AppError {
	if user.Role != model.SystemAdminRoleId {
		return model.NewAppError("session.not_authorized", http.StatusForbidden)
	}

	if err := a.Store.File.DeleteMetadata(id); err != nil {
		tlog.Errorw("Failed to delete file metadata",
			"metadata_id", id,
			"error", err,
		)
		return model.NewAppError("file.metadata_delete_failed", http.StatusInternalServerError)
	}

	return nil
}

func (a *App) EditFileMetadata(user model.User, id, name, metadataType string, fields interface{}) *model.AppError {
	if user.Role != model.SystemAdminRoleId {
		return model.NewAppError("session.not_authorized", http.StatusForbidden)
	}

	s, err := json.Marshal(fields)
	if err != nil {
		tlog.Errorw("Failed to marshal file metadata values",
			"metadata_id", id,
			"error", err,
		)
		return model.NewAppError("file.metadata_edit_failed", http.StatusInternalServerError)
	}

	if err = a.Store.File.UpdateMetadata(id, name, metadataType, string(s)); err != nil {
		tlog.Errorw("Failed to edit file metadata",
			"metadata_id", id,
			"error", err,
		)
		return model.NewAppError("file.metadata_edit_failed", http.StatusInternalServerError)
	}

	return nil
}

func (a *App) UpdateMetadataEntry(user model.User, id string, value interface{}) *model.AppError {

	fileID, err := a.Store.File.GetMetadataIDForFile(id)
	if err != nil {
		tlog.Errorw("Failed to retrieve file ID for metadata entry",
			"metadata_id", id,
			"error", err,
		)
		return model.NewAppError("file.metadata_retrieval_failed", http.StatusInternalServerError)
	}

	file, appErr := a.HasPermission(*fileID, user)
	if appErr != nil {
		return appErr
	}

	if file.Owner != user.ID && !file.AccessLevel.CanEdit() {
		return model.NewAppError("file.forbidden", http.StatusForbidden)
	}

	if err = a.Store.File.UpdateMetadataEntry(id, value); err != nil {
		tlog.Errorw("Failed to update file metadata entry",
			"metadata_id", id,
			"file_id", *fileID,
			"error", err,
		)
		return model.NewAppError("file.metadata_update_failed", http.StatusInternalServerError)
	}

	a.RecordActivity(user.ID, model.AppFiles, model.ActivityFileMeta, file.Parent, file.ID, map[string]any{
		"name": file.DisplayName,
		"type": file.Type,
	})

	return nil
}

// ensureUserRoot returns the user's root ("My Drive") on the storage, creating
// it on first access. Roots are per-user (owner=userID, parent=""), so a user
// only ever sees their own files under it; shared items surface separately.
func (a *App) ensureUserRoot(ctx context.Context, userID string, storageID string, name string) (*model.File, error) {
	root, err := a.Store.File.GetUserRoot(ctx, userID, storageID)
	if err != nil {
		return nil, err
	}

	if root != nil {
		return root, nil
	}

	newRoot := model.File{
		ID:          model.NewID(),
		Owner:       userID,
		Parent:      "",
		Storage:     storageID,
		Name:        name,
		DisplayName: name,
		Type:        "cloud#drive",
	}

	if _, err := a.Store.File.Create(newRoot); err != nil {
		return nil, err
	}

	return &newRoot, nil
}

func (a *App) GetUserDrives(user model.User) ([]model.File, *model.AppError) {
	storages, err := a.Store.Storage.GetAll()
	if err != nil {
		tlog.Errorw("Failed to list storages", "error", err)
		return nil, model.NewAppError("file.retrieval_failed", http.StatusInternalServerError)
	}

	ctx, cancel := a.dbCtx(context.Background())
	defer cancel()

	drives := make([]model.File, 0, len(storages))
	for _, s := range storages {
		root, err := a.ensureUserRoot(ctx, user.ID, s.ID, s.Label)
		if err != nil {
			tlog.Errorw("Failed to ensure user root",
				"user_id", user.ID,
				"storage", s.ID,
				"error", err,
			)
			return nil, model.NewAppError("file.retrieval_failed", http.StatusInternalServerError)
		}

		drives = append(drives, *root)
	}

	return drives, nil
}

// listChildren returns the children of folder visible to userID. An owner sees
// their own children (plus any child shared directly to them, via
// GetByParent); a non-owner reaches a folder through a share, so they see
// all of its contents.
func (a *App) listChildren(ctx context.Context, userID string, folder model.File) ([]model.File, error) {
	if folder.Owner == userID {
		return a.Store.File.GetByParent(folder, userID)
	}

	return a.Store.File.GetChildren(ctx, folder.ID)
}

func (a *App) GetUserFiles(user model.User, id string) ([]model.File, *model.AppError) {
	file, appErr := a.HasPermission(id, user)
	if appErr != nil {
		return nil, appErr
	}

	// No request context to inherit cancellation from, so the inherited-listing
	// branch is capped by the configured query timeout instead.
	ctx, cancel := a.dbCtx(context.Background())
	defer cancel()

	dbFiles, err := a.listChildren(ctx, user.ID, *file)
	if err != nil {
		tlog.Errorw("Failed to retrieve files by parent",
			"parent_id", id,
			"user_id", user.ID,
			"error", err,
		)
		return nil, model.NewAppError("file.retrieval_failed", http.StatusInternalServerError)
	}

	if appErr = a.enrichFiles(dbFiles, user.ID); appErr != nil {
		return nil, appErr
	}

	return dbFiles, nil
}

func (a *App) GetSharedFiles(userID string) ([]model.File, *model.AppError) {
	sharedFiles, err := a.Store.File.GetShared(userID)
	if err != nil {
		tlog.Errorw("Failed to retrieve shared files",
			"user_id", userID,
			"error", err,
		)
		return nil, model.NewAppError("file.retrieval_failed", http.StatusInternalServerError)
	}

	return sharedFiles, nil
}

func (a *App) GetRecents(user model.User) ([]model.File, *model.AppError) {
	ctx, cancel := a.dbCtx(context.Background())
	defer cancel()

	recents, err := a.Store.File.GetRecents(ctx, user.ID)
	if err != nil {
		tlog.Errorw("Failed to retrieve recent files",
			"user_id", user.ID,
			"error", err,
		)
		return nil, model.NewAppError("file.retrieval_failed", http.StatusInternalServerError)
	}

	if appErr := a.enrichFiles(recents, user.ID); appErr != nil {
		return nil, appErr
	}

	return recents, nil
}

func (a *App) GetParent(user model.User, id string) ([]model.File, *model.AppError) {
	if _, appErr := a.HasPermission(id, user); appErr != nil {
		return nil, appErr
	}

	ctx, cancel := a.dbCtx(context.Background())
	defer cancel()

	arr, err := a.Store.File.GetBreadcrumbs(ctx, user.ID, id)
	if err != nil {
		tlog.Errorw("Failed to resolve breadcrumbs",
			"file_id", id,
			"user_id", user.ID,
			"error", err,
		)
		return nil, model.NewAppError("file.retrieval_failed", http.StatusInternalServerError)
	}

	return arr, nil
}

func (a *App) AddFavorite(user model.User, fileID string) *model.AppError {
	file, appErr := a.HasPermission(fileID, user)
	if appErr != nil {
		return appErr
	}

	id, err := a.Store.File.GetFavourite(user.ID, file.ID)
	if err != nil {
		tlog.Errorw("Failed to retrieve favourite file",
			"file_id", fileID,
			"user_id", user.ID,
			"error", err,
		)
		return model.NewAppError("file.retrieval_failed", http.StatusInternalServerError)
	}

	if id != "" {
		if err = a.Store.File.RemoveFavourite(user.ID, fileID); err != nil {
			tlog.Errorw("Failed to remove favourite file",
				"file_id", fileID,
				"user_id", user.ID,
				"error", err,
			)
			return model.NewAppError("file.delete_failed", http.StatusInternalServerError)
		}

		return nil
	}

	if err = a.Store.File.AddFavourite(user.ID, fileID); err != nil {
		tlog.Errorw("Failed to add favourite file",
			"file_id", fileID,
			"user_id", user.ID,
			"error", err,
		)
		return model.NewAppError("file.create_failed", http.StatusInternalServerError)
	}

	return nil
}

func (a *App) GetFavourites(userID string) ([]model.File, *model.AppError) {
	ctx, cancel := a.dbCtx(context.Background())
	defer cancel()

	f, err := a.Store.File.GetFavourites(ctx, userID)
	if err != nil {
		tlog.Errorw("Failed to retrieve favourite files",
			"user_id", userID,
			"error", err,
		)
		return nil, model.NewAppError("file.retrieval_failed", http.StatusInternalServerError)
	}

	if appErr := a.enrichFiles(f, userID); appErr != nil {
		return nil, appErr
	}

	return f, nil
}

func (a *App) GetDetails(ctx context.Context, user model.User, fileID string) (*model.FileDetails, *model.AppError) {
	file, appErr := a.HasPermission(fileID, user)
	if appErr != nil {
		return nil, appErr
	}

	owner, err := a.Store.User.Get(file.Owner)
	if err != nil {
		tlog.Errorw("Failed to retrieve file owner",
			"file_id", fileID,
			"owner_id", file.Owner,
			"error", err,
		)
		return nil, model.NewAppError("user.not_found", http.StatusInternalServerError)
	}

	sharedUsers, err := a.Store.User.GetSharedUsersForFile(ctx, *file)
	if err != nil {
		tlog.Errorw("Failed to retrieve shared users for file",
			"file_id", fileID,
			"error", err,
		)
		return nil, model.NewAppError("file.retrieval_failed", http.StatusInternalServerError)
	}

	sharedGroups, err := a.Store.Groups.GetSharesForFile(ctx, file.ID)
	if err != nil {
		tlog.Errorw("Failed to retrieve shared groups for file",
			"file_id", fileID,
			"error", err,
		)
		return nil, model.NewAppError("file.retrieval_failed", http.StatusInternalServerError)
	}

	// Owners and managers govern all links on the file; others get none.
	links := []model.ExternalShare{}
	if file.AccessLevel.CanShare() {
		links, err = a.Store.Share.GetActiveForFile(ctx, file.ID)
		if err != nil {
			tlog.Errorw("Failed to retrieve shared links for file",
				"file_id", fileID,
				"error", err,
			)
			return nil, model.NewAppError("file.retrieval_failed", http.StatusInternalServerError)
		}
	}

	sharedLinks := make([]model.SharedLinks, 0, len(links))
	for _, link := range links {
		linkOwner, err := a.Store.User.Get(link.Owner)
		if err != nil {
			tlog.Errorw("Failed to retrieve shared link owner",
				"file_id", fileID,
				"owner_id", link.Owner,
				"error", err,
			)
			return nil, model.NewAppError("user.not_found", http.StatusInternalServerError)
		}

		if linkOwner == nil {
			return nil, model.NewAppError("user.not_found", http.StatusNotFound)
		}

		sharedLinks = append(sharedLinks, model.SharedLinks{
			ID:                link.ShareToken,
			FullName:          linkOwner.Name + " " + linkOwner.LastName,
			ShareTime:         link.ShareTime,
			Expiration:        link.Expiration,
			PasswordProtected: link.PasswordProtected,
			AllowView:         link.AllowView,
			AllowDownload:     link.AllowDownload,
			AllowUpload:       link.AllowUpload,
			AllowEdit:         link.AllowEdit,
			Message:           link.Message,
		})
	}

	return &model.FileDetails{
		ID:           file.ID,
		Type:         file.Type,
		Owner:        owner.Name + " " + owner.LastName,
		OwnerEmail:   owner.Email,
		AccessLevel:  file.AccessLevel,
		Modified:     file.Created,
		Created:      file.Created,
		SharedUsers:  sharedUsers,
		SharedGroups: sharedGroups,
		SharedLinks:  sharedLinks,
	}, nil
}

func (a *App) CreateDocument(user model.User, d model.FileCrateRequest) (*model.File, *model.AppError) {

	file, appErr := a.HasPermission(d.ID, user)
	if appErr != nil {
		return nil, appErr
	}

	owner := newFileOwner(*file, user)

	var src string
	var mimeType string
	switch d.Type {
	case "spreadsheet":
		src = "templates/excel.xlsx"
		mimeType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case "presentation":
		src = "templates/powerpoint.pptx"
		mimeType = "application/vnd.openxmlformats-officedocument.presentationml.presentation"
	case "document":
		src = "templates/word.docx"
		mimeType = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case "text":
		src = "templates/text.txt"
		mimeType = "text/plain"
	default:
		return nil, model.NewAppError("file.unknown_type", http.StatusBadRequest)
	}

	ctx := context.Background()

	newFileID := model.NewID()

	filePath, err := a.BuildFilePath(file.Storage, newFileID, model.AppFiles)
	if err != nil {
		tlog.Errorw("Failed to create file path",
			"file_id", newFileID,
			"storage", file.Storage,
			"error", err,
		)
		return nil, model.NewAppError("file.create_failed", http.StatusInternalServerError)
	}

	exists, err := a.FileStorageObjects[file.Storage].FileExists(ctx, filePath)
	if err != nil {
		tlog.Errorw("Failed to check if file exists in storage",
			"file_id", newFileID,
			"storage", file.Storage,
			"error", err,
		)
		return nil, model.NewAppError("file.create_failed", http.StatusInternalServerError)
	}

	if exists {
		tlog.Errorw("File ID collision detected in storage",
			"file_id", newFileID,
			"storage", file.Storage,
		)
		return nil, model.NewAppError("file.create_failed", http.StatusInternalServerError)
	}

	template, err := os.ReadFile(src)
	if err != nil {
		tlog.Errorw("Failed to read document template",
			"template", src,
			"error", err,
		)
		return nil, model.NewAppError("file.create_failed", http.StatusInternalServerError)
	}

	if err = a.FileStorageObjects[file.Storage].WriteFile(ctx, filePath, bytes.NewBuffer(template), int64(len(template))); err != nil {
		tlog.Errorw("Failed to write document to storage",
			"file_id", newFileID,
			"storage", file.Storage,
			"error", err,
		)
		return nil, model.NewAppError("file.create_failed", http.StatusInternalServerError)
	}

	size, err := a.FileStorageObjects[file.Storage].FileSize(filePath)
	if err != nil {
		tlog.Errorw("Failed to retrieve file size",
			"file_id", newFileID,
			"storage", file.Storage,
			"error", err,
		)
	}

	hasSpace, err := a.ownerHasSpace(owner, user, size)
	if err != nil {
		return nil, model.NewAppError("file.no_space", http.StatusInsufficientStorage)
	}

	if !hasSpace {
		if err = a.FileStorageObjects[file.Storage].RemoveFile(ctx, filePath); err != nil {
			tlog.Errorw("Failed to remove file after storage limit check",
				"file_id", newFileID,
				"storage", file.Storage,
				"error", err,
			)
		}

		return nil, model.NewAppError("file.no_space", http.StatusInsufficientStorage)
	}

	myNewFile := model.File{
		ID:          newFileID,
		Owner:       owner,
		Parent:      file.ID,
		Storage:     file.Storage,
		Size:        size,
		Name:        d.DocName + "." + d.Extension,
		DisplayName: d.DocName + "." + d.Extension,
		Type:        mimeType,
		Shared:      file.Shared,
	}

	storedID, err := a.Store.File.Create(myNewFile)
	if err != nil {
		tlog.Errorw("Failed to add file record",
			"file_id", newFileID,
			"error", err,
		)
		return nil, model.NewAppError("file.create_failed", http.StatusInternalServerError)
	}

	newFile, err := a.Store.File.Get(*storedID)
	if err != nil {
		tlog.Errorw("Failed to retrieve created file",
			"file_id", *storedID,
			"error", err,
		)
		return nil, model.NewAppError("file.create_failed", http.StatusInternalServerError)
	}

	if newFile == nil {
		return nil, model.NewAppError("file.not_found", http.StatusNotFound)
	}

	a.RecordActivity(user.ID, model.AppFiles, model.ActivityFileCreate, file.ID, newFile.ID, map[string]any{
		"type": newFile.Type,
		"name": newFile.DisplayName,
	})

	a.CreateFileNotification(*newFile, user, model.NOTIFICATION_FILE_CREATE, nil)

	return newFile, nil
}

func (a *App) CreateFolder(user model.User, name string, id string) (*model.File, *model.AppError) {
	file, appErr := a.HasPermission(id, user)
	if appErr != nil {
		return nil, appErr
	}

	owner := newFileOwner(*file, user)

	hasSpace, err := a.ownerHasSpace(owner, user, 0)
	if err != nil {
		return nil, model.NewAppError("file.no_space", http.StatusInsufficientStorage)
	}

	if !hasSpace {
		return nil, model.NewAppError("file.no_space", http.StatusInsufficientStorage)
	}

	newID, err := a.Store.File.CreateFolder(model.File{
		ID:          model.NewID(),
		Owner:       owner,
		Parent:      file.ID,
		Storage:     file.Storage,
		Name:        name,
		DisplayName: name,
		Size:        0,
		Shared:      file.Shared,
	})
	if err != nil {
		tlog.Errorw("Failed to add folder record",
			"parent_id", file.ID,
			"user_id", user.ID,
			"error", err,
		)
		return nil, model.NewAppError("file.create_failed", http.StatusInternalServerError)
	}

	newFile, err := a.Store.File.Get(*newID)
	if err != nil {
		tlog.Errorw("Failed to retrieve created folder",
			"file_id", *newID,
			"error", err,
		)
		return nil, model.NewAppError("file.create_failed", http.StatusInternalServerError)
	}

	if newFile == nil {
		return nil, model.NewAppError("file.not_found", http.StatusNotFound)
	}

	a.RecordActivity(user.ID, model.AppFiles, model.ActivityFileCreate, file.ID, newFile.ID, map[string]any{
		"type": newFile.Type,
		"name": newFile.DisplayName,
	})

	a.CreateFileNotification(*newFile, user, model.NOTIFICATION_FILE_CREATE, nil)

	return newFile, nil
}

func (a *App) GetFileToDownload(user model.User, id string) (*model.File, *model.AppError) {
	_, appErr := a.HasPermission(id, user)
	if appErr != nil {
		return nil, appErr
	}

	file, err := a.Store.File.Get(id)
	if err != nil {
		tlog.Errorw("Failed to retrieve file for download",
			"file_id", id,
			"user_id", user.ID,
			"error", err,
		)
		return nil, model.NewAppError("file.retrieval_failed", http.StatusInternalServerError)
	}

	if file == nil {
		return nil, model.NewAppError("file.not_found", http.StatusNotFound)
	}

	return file, nil
}

func (a *App) MoveFile(ctx context.Context, user model.User, fileIDS []string, target string) (*model.MoveFileResult, *model.AppError) {
	targetDir, err := a.Store.File.Get(target)
	if err != nil {
		tlog.Errorw("Failed to retrieve target directory",
			"target_id", target,
			"user_id", user.ID,
			"error", err,
		)
		return nil, model.NewAppError("file.retrieval_failed", http.StatusInternalServerError)
	}

	if targetDir == nil {
		return nil, model.NewAppError("file.not_found", http.StatusNotFound)
	}

	isDifferentDisk := false
	for _, fid := range fileIDS {
		f, appErr := a.HasPermission(fid, user)
		if appErr != nil {
			return nil, appErr
		}

		if f.IsLocked() {
			return nil, model.NewAppError("file.locked", http.StatusForbidden)
		}

		if f.Storage != targetDir.Storage {
			isDifferentDisk = true
			break
		}
	}

	if isDifferentDisk {
		job, appErr := a.enqueueMoveFileJob(ctx, user, fileIDS, target)
		if appErr != nil {
			return nil, appErr
		}

		return &model.MoveFileResult{Job: job}, nil
	}

	fileMove, appErr := a.MoveFileWithinDisk(ctx, user, fileIDS, target)
	if appErr != nil {
		return nil, appErr
	}

	return &model.MoveFileResult{FileMove: fileMove}, nil
}

func (a *App) MoveFileWithinDisk(ctx context.Context, user model.User, fileIDS []string, target string) (*model.FileMove, *model.AppError) {
	targetDir, err := a.Store.File.Get(target)
	if err != nil {
		tlog.Errorw("Failed to retrieve target directory",
			"target_id", target,
			"error", err,
		)
		return nil, model.NewAppError("file.retrieval_failed", http.StatusInternalServerError)
	}

	if targetDir == nil {
		return nil, model.NewAppError("file.not_found", http.StatusNotFound)
	}

	fileMove := model.FileMove{
		Immediate: true,
		Payload: model.FileMovePayload{
			UserID: user.ID,
			Target: targetDir.ID,
		},
	}

	id := make([]string, 0, len(fileIDS))
	for _, v := range fileIDS {
		if !slices.Contains(id, v) {
			id = append(id, v)
		}
	}

	if slices.Contains(id, targetDir.ID) {
		return nil, model.NewAppError("file.move_failed", http.StatusBadRequest)
	}

	for _, v := range id {
		ok, err := a.Store.File.CanMoveFolder(v, targetDir.ID)
		if err != nil {
			tlog.Errorw("Failed to verify file move eligibility",
				"file_id", v,
				"target_id", targetDir.ID,
				"error", err,
			)
			return nil, model.NewAppError("file.move_failed", http.StatusInternalServerError)
		}

		if !ok {
			return nil, model.NewAppError("file.move_failed", http.StatusForbidden)
		}
	}

	rootIDS := make([]string, 0)
	subfileIDS := make([]string, 0)

	for _, v := range id {
		file, appErr := a.HasPermission(v, user)
		if appErr != nil {
			return nil, appErr
		}

		if user.ID != file.Owner {
			allowed, err := a.Store.File.MoveAllowedWithinShare(ctx, user.ID, v, targetDir.ID)
			if err != nil {
				tlog.Errorw("Failed to authorize shared move",
					"file_id", v,
					"target_id", targetDir.ID,
					"error", err,
				)
				return nil, model.NewAppError("file.retrieval_failed", http.StatusInternalServerError)
			}

			if !allowed {
				return nil, model.NewAppError("file.move_failed", http.StatusForbidden)
			}
		}

		if file.Parent == targetDir.ID {
			tlog.Warnw("Skipping file move to the same folder",
				"file_id", v,
				"target_id", targetDir.ID,
			)
			return nil, nil
		}

		if targetDir.Type == "cloud#drive" {
			file.Owner = user.ID
		} else {
			file.Owner = targetDir.Owner
		}

		if file.IsFolder {
			subFiles, err := a.Store.File.GetSubFilesByFolder(file.ID, file.Storage)
			if err != nil {
				tlog.Errorw("Failed to retrieve subfiles for folder",
					"folder_id", file.ID,
					"error", err,
				)
				return nil, model.NewAppError("file.retrieval_failed", http.StatusInternalServerError)
			}

			for _, f := range subFiles {
				if f.ID != file.ID {
					subfileIDS = append(subfileIDS, f.ID)
					fileMove.Payload.SubfileIDs = append(fileMove.Payload.SubfileIDs, model.MoveItem{
						ID:       f.ID,
						IsFolder: f.IsFolder,
						Target:   targetDir.ID,
						Source:   f.Parent,
					})
				}
			}
		}

		rootIDS = append(rootIDS, file.ID)
		fileMove.Payload.RootIDs = append(fileMove.Payload.RootIDs, model.MoveItem{
			ID:       file.ID,
			IsFolder: file.IsFolder,
			Target:   targetDir.ID,
			Source:   file.Parent,
		})
	}

	if err = a.Store.File.UpdateParentID(rootIDS, subfileIDS, targetDir.ID, targetDir.Storage); err != nil {
		tlog.Errorw("Failed to update file parent IDs",
			"target_id", targetDir.ID,
			"error", err,
		)
		return nil, model.NewAppError("file.move_failed", http.StatusInternalServerError)
	}

	return &fileMove, nil
}

func (a *App) enqueueMoveFileJob(ctx context.Context, user model.User, fileIDS []string, target string) (*model.Job, *model.AppError) {
	targetDir, appErr := a.HasPermission(target, user)
	if appErr != nil {
		return nil, appErr
	}

	id := make([]string, 0, len(fileIDS))
	for _, v := range fileIDS {
		if !slices.Contains(id, v) {
			id = append(id, v)
		}
	}

	rootIDS := make([]model.MoveItem, 0)
	subfileIDS := make([]model.MoveItem, 0)
	totalBytes := int64(0)

	for _, v := range id {
		file, appErr := a.HasPermission(v, user)
		if appErr != nil {
			return nil, appErr
		}

		if file.Owner != user.ID {
			return nil, model.NewAppError("file.forbidden", http.StatusForbidden)
		}

		if file.IsFolder {
			subFiles, err := a.Store.File.GetSubFilesByFolder(file.ID, file.Storage)
			if err != nil {
				tlog.Errorw("Failed to retrieve subfiles for folder",
					"folder_id", file.ID,
					"error", err,
				)
				return nil, model.NewAppError("file.retrieval_failed", http.StatusInternalServerError)
			}

			for _, f := range subFiles {
				if f.ID == file.ID {
					continue
				}
				// Reject if a descendant is locked.
				// the cross-disk move rewrites its storage, so a skipped
				// one would be stranded on the old disk.
				if f.IsLocked() {
					return nil, model.NewAppError("file.locked", http.StatusForbidden)
				}

				if !f.IsFolder {
					totalBytes += f.Size
				}

				subfileIDS = append(subfileIDS, model.MoveItem{
					ID:       f.ID,
					IsFolder: f.IsFolder,
					Source:   f.Parent,
					Target:   target,
				})
			}
		} else {
			totalBytes += file.Size
		}

		rootIDS = append(rootIDS, model.MoveItem{
			ID:       file.ID,
			IsFolder: file.IsFolder,
			Source:   file.Parent,
			Target:   target,
		})
	}

	payloadBytes, err := json.Marshal(model.MoveFilesJobPayload{
		UserID:     user.ID,
		RootIDS:    rootIDS,
		SubfileIDs: subfileIDS,
		Target:     targetDir.ID,
		TotalBytes: totalBytes,
	})
	if err != nil {
		tlog.Errorw("Failed to marshal move file job payload",
			"user_id", user.ID,
			"error", err,
		)
		return nil, model.NewAppError("file.move_failed", http.StatusInternalServerError)
	}

	job, err := a.Store.Jobs.Create(model.Job{
		ID:        model.NewID(),
		Type:      model.JobTypeMoveFiles,
		Status:    model.JobStatusPending,
		UserID:    user.ID,
		Payload:   payloadBytes,
		CreatedAt: time.Now().Unix(),
	})
	if err != nil {
		tlog.Errorw("Failed to create move file job",
			"user_id", user.ID,
			"error", err,
		)
		return nil, model.NewAppError("file.move_failed", http.StatusInternalServerError)
	}

	return job, nil
}

func (a *App) MoveFilesAcrossDisks(ctx context.Context, jobID string) error {
	moveJob, err := a.Store.Jobs.Get(jobID)
	if err != nil {
		tlog.Errorw("Failed to retrieve move file job",
			"job_id", jobID,
			"error", err,
		)
		return err
	}

	if moveJob == nil {
		return errors.New("move job not found")
	}

	moveFileJob := model.MoveFilesJobPayload{}
	if err = json.Unmarshal([]byte(moveJob.Payload), &moveFileJob); err != nil {
		tlog.Errorw("Failed to unmarshal move file job payload",
			"job_id", jobID,
			"error", err,
		)
		return err
	}

	allFiles := make([]string, 0, len(moveFileJob.RootIDS)+len(moveFileJob.SubfileIDs))
	rootIDs := make([]string, 0, len(moveFileJob.RootIDS))
	subfileIDs := make([]string, 0, len(moveFileJob.SubfileIDs))

	for _, f := range moveFileJob.RootIDS {
		allFiles = append(allFiles, f.ID)
		rootIDs = append(rootIDs, f.ID)
	}

	for _, f := range moveFileJob.SubfileIDs {
		allFiles = append(allFiles, f.ID)
		subfileIDs = append(subfileIDs, f.ID)
	}

	targetDir, err := a.Store.File.Get(moveFileJob.Target)
	if err != nil {
		tlog.Errorw("Failed to retrieve target directory",
			"job_id", jobID,
			"target_id", moveFileJob.Target,
			"error", err,
		)
		return err
	}

	if targetDir == nil {
		return errors.New("move target not found")
	}

	var movedBytes int64
	moved := make([]movedFile, 0)

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if moveFileJob.TotalBytes == 0 {
					continue
				}

				progress := atomic.LoadInt64(&movedBytes) * 100 / moveFileJob.TotalBytes
				_ = a.Store.Jobs.UpdateProgress(moveJob.ID, int(progress))
			}
		}
	}()

	committed := false
	defer func() {
		if !committed {
			for _, f := range moved {
				if err := a.removeStoredFile(context.Background(), f.TargetStorage, f); err != nil {
					tlog.Errorw("Failed to remove file during rollback",
						"job_id", jobID,
						"file_id", f.ID,
						"error", err,
					)
				}
			}
		}
	}()

	for _, v := range allFiles {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		file, err := a.Store.File.Get(v)
		if err != nil {
			tlog.Errorw("Failed to retrieve file during move",
				"job_id", jobID,
				"file_id", v,
				"error", err,
			)
			return err
		}

		if file == nil {
			continue
		}

		if file.IsFolder {
			continue
		}

		srcPath, err := a.BuildFilePath(file.Storage, file.ID, model.AppFiles)
		if err != nil {
			tlog.Errorw("Failed to build source file path",
				"job_id", jobID,
				"file_id", file.ID,
				"error", err,
			)
			return err
		}

		object, err := a.FileStorageObjects[file.Storage].ReadFile(ctx, srcPath)
		if err != nil {
			tlog.Errorw("Failed to read file from storage",
				"job_id", jobID,
				"file_id", file.ID,
				"error", err,
			)
			return err
		}

		fileSize, err := a.FileStorageObjects[file.Storage].FileSize(srcPath)
		if err != nil {
			tlog.Errorw("Failed to retrieve file size",
				"job_id", jobID,
				"file_id", file.ID,
				"error", err,
			)
			return err
		}

		pr := &ProgressReader{
			r: object,
			onRead: func(n int) {
				atomic.AddInt64(&movedBytes, int64(n))
				if moveFileJob.TotalBytes == 0 {
					return
				}

				progress := atomic.LoadInt64(&movedBytes) * 100 / moveFileJob.TotalBytes
				if err := a.Store.Jobs.UpdateProgress(moveJob.ID, int(progress)); err != nil {
					tlog.Errorw("Failed to update move file job progress",
						"job_id", jobID,
						"error", err,
					)
				}
			},
		}

		// Recorded before the copy so a rollback also removes a partly written file.
		moved = append(moved, movedFile{
			ID:            file.ID,
			Type:          file.Type,
			Storage:       file.Storage,
			TargetStorage: targetDir.Storage,
		})

		if err = a.moveFile(ctx, *file, *targetDir, pr, fileSize); err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				tlog.Warnw("Move file job cancelled",
					"job_id", jobID,
					"file_id", file.ID,
				)
				return err
			}

			tlog.Errorw("Failed to move file",
				"job_id", jobID,
				"file_id", file.ID,
				"error", err,
			)
			return err
		}

		// A thumbnail can be rebuilt from the file, so losing one must not fail the move.
		if isThumbnailType(file.Type) {
			if err = a.copyThumbnailImageToDisk(ctx, *file, *targetDir, file.ID); err != nil {
				tlog.Warnw("Failed to copy thumbnail during move",
					"job_id", jobID,
					"file_id", file.ID,
					"error", err,
				)
			}
		}

		if moveFileJob.TotalBytes > 0 {
			progress := atomic.LoadInt64(&movedBytes) * 100 / moveFileJob.TotalBytes
			if err = a.Store.Jobs.UpdateProgress(moveJob.ID, int(progress)); err != nil {
				tlog.Errorw("Failed to update move file job progress",
					"job_id", jobID,
					"error", err,
				)
				return err
			}
		}
	}

	if err = a.Store.File.MoveToStorage(moveJob.ID, rootIDs, subfileIDs, targetDir.Storage, targetDir.ID); err != nil {
		tlog.Errorw("Failed to update file records after move",
			"job_id", jobID,
			"target_id", targetDir.ID,
			"error", err,
		)
		return err
	}

	committed = true

	for _, f := range moved {
		if err = a.removeStoredFile(context.Background(), f.Storage, f); err != nil {
			tlog.Errorw("Failed to remove source file after move",
				"job_id", jobID,
				"file_id", f.ID,
				"error", err,
			)
		}
	}

	return nil
}

func (a *App) removeStoredFile(ctx context.Context, storageID string, f movedFile) error {
	names := []string{f.ID}
	if isThumbnailType(f.Type) {
		names = append(names, "thumbnails/"+f.ID+".jpg")
	}

	var errs []error
	for _, name := range names {
		p, err := a.BuildFilePath(storageID, name, model.AppFiles)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		if err = a.FileStorageObjects[storageID].RemoveFile(ctx, p); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

func (a *App) moveFile(
	ctx context.Context,
	f model.File,
	targetDir model.File,
	r io.ReadCloser,
	fileSize int64,
) error {
	defer r.Close()

	dstPath, err := a.BuildFilePath(targetDir.Storage, f.ID, model.AppFiles)
	if err != nil {
		return err
	}

	return a.FileStorageObjects[targetDir.Storage].CopyFileFromReader(ctx, dstPath, r, fileSize)
}

func (a *App) copyThumbnailImageToDisk(ctx context.Context, f model.File, targetDir model.File, newFileID string) error {
	srcThumbPath, err := a.BuildFilePath(f.Storage, "thumbnails/"+f.ID+".jpg", model.AppFiles)
	if err != nil {
		return err
	}

	object, err := a.FileStorageObjects[f.Storage].ReadFile(ctx, srcThumbPath)
	if err != nil {
		return err
	}

	defer object.Close()

	size, err := a.FileStorageObjects[f.Storage].FileSize(srcThumbPath)
	if err != nil {
		return err
	}

	dstThumbPath, err := a.BuildFilePath(targetDir.Storage, "thumbnails/"+newFileID+".jpg", model.AppFiles)
	if err != nil {
		return err
	}

	return a.FileStorageObjects[targetDir.Storage].WriteFile(ctx, dstThumbPath, object, size)
}

func (a *App) DeleteFile(user model.User, id []string) *model.AppError {
	for _, fileID := range id {
		file, appErr := a.HasPermission(fileID, user)
		if appErr != nil {
			return appErr
		}

		if file.IsLocked() {
			return model.NewAppError("file.locked", http.StatusForbidden)
		}

		if file.IsOwner(user.ID) {
			if file.IsFolder {
				subFiles, err := a.Store.File.GetSubFilesByFolder(file.ID, file.Storage)
				if err != nil {
					tlog.Errorw("Failed to retrieve subfiles for folder",
						"file_id", file.ID,
						"error", err,
					)
					return model.NewAppError("file.retrieval_failed", http.StatusInternalServerError)
				}
				// Reject before deleting anything: permanent deletion is
				// irreversible, so a descendant held by a running job (e.g. a
				// cross-disk move) or an active edit must block the whole op.
				for _, sf := range subFiles {
					if sf.ID != file.ID && sf.IsLocked() {
						return model.NewAppError("file.locked", http.StatusForbidden)
					}
				}

				for _, sf := range subFiles {
					if sf.ID != file.ID {
						if appErr := a.permanentlyDeleteFile(sf); appErr != nil {
							return appErr
						}
					}
				}
			}

			if appErr = a.permanentlyDeleteFile(*file); appErr != nil {
				return appErr
			}
		} else {
			if err := a.Store.File.RemoveUserFromShare(file.ID, user.ID); err != nil {
				tlog.Errorw("Failed to remove user from shared file",
					"file_id", file.ID,
					"user_id", user.ID,
					"error", err,
				)
			}
		}
	}

	return nil
}

func (a *App) permanentlyDeleteFile(file model.File) *model.AppError {
	if !file.IsFolder {
		ctx := context.Background()

		filePath, err := a.BuildFilePath(file.Storage, file.ID, model.AppFiles)
		if err != nil {
			tlog.Errorw("Failed to build file path for deletion",
				"file_id", file.ID,
				"storage", file.Storage,
				"error", err,
			)
		} else {
			if err := a.FileStorageObjects[file.Storage].RemoveFile(ctx, filePath); err != nil {
				tlog.Errorw("Failed to remove file from storage",
					"file_id", file.ID,
					"storage", file.Storage,
					"error", err,
				)
			}

			thumbPath, err := a.BuildFilePath(file.Storage, "thumbnails/"+file.ID+".jpg", model.AppFiles)
			if err != nil {
				tlog.Errorw("Failed to build thumbnail path for deletion",
					"file_id", file.ID,
					"storage", file.Storage,
					"error", err,
				)
			} else {
				if err := a.FileStorageObjects[file.Storage].RemoveFile(ctx, thumbPath); err != nil {
					tlog.Errorw("Failed to remove file thumbnail from storage",
						"file_id", file.ID,
						"storage", file.Storage,
						"error", err,
					)
				}
			}
		}
	}

	if _, err := a.Store.File.Delete(file.ID, file.Storage); err != nil {
		tlog.Errorw("Failed to delete file record",
			"file_id", file.ID,
			"error", err,
		)
		return model.NewAppError("file.delete_failed", http.StatusInternalServerError)
	}

	if err := a.Store.File.DeleteShares(file.ID); err != nil {
		tlog.Errorw("Failed to delete shares for file",
			"file_id", file.ID,
			"error", err,
		)
	}

	if err := a.Store.Share.DeactivateForFile(context.Background(), file.ID); err != nil {
		tlog.Warnw("Failed to deactivate public links for deleted file",
			"file_id", file.ID,
		)
	}

	return nil
}

func (a *App) GetDeleted(user model.User) ([]model.File, *model.AppError) {
	files, err := a.Store.File.GetDeleted(user.ID)
	if err != nil {
		tlog.Errorw("Failed to retrieve deleted files",
			"user_id", user.ID,
			"error", err,
		)
		return nil, model.NewAppError("file.retrieval_failed", http.StatusInternalServerError)
	}

	if appErr := a.enrichFiles(files, user.ID); appErr != nil {
		return nil, appErr
	}

	return files, nil
}

func (a *App) Restore(user model.User, id []string) *model.AppError {
	if err := a.Store.File.Restore(user.ID, id); err != nil {
		tlog.Errorw("Failed to restore files",
			"user_id", user.ID,
			"error", err,
		)
		return model.NewAppError("file.restore_failed", http.StatusInternalServerError)
	}

	for _, fileID := range id {
		f, err := a.Store.File.Get(fileID)
		if err != nil {
			tlog.Errorw("Failed to retrieve restored file",
				"file_id", fileID,
				"user_id", user.ID,
				"error", err,
			)
			continue
		}

		if f == nil {
			continue
		}

		a.RecordActivity(user.ID, model.AppFiles, model.ActivityFileRestore, f.Parent, f.ID, map[string]any{
			"name": f.DisplayName,
			"type": f.Type,
		})

		a.CreateFileNotification(*f, user, model.NOTIFICATION_FILE_RESTORE, nil)
	}

	return nil
}

func (a *App) Trash(user model.User, IDS []string) *model.AppError {
	files, err := a.Store.File.GetByIDs(user, IDS)
	if err != nil {
		tlog.Errorw("Failed to retrieve files",
			"user_id", user.ID,
			"error", err,
		)
		return model.NewAppError("file.retrieval_failed", http.StatusInternalServerError)
	}

	if files == nil {
		return model.NewAppError("file.not_found", http.StatusNotFound)
	}

	filesToTrash := make([]model.File, 0)
	subFiles := make([]model.File, 0)
	filesToUnshare := make([]model.File, 0)

	for _, v := range files {
		if v.IsOwner(user.ID) {
			if v.IsLocked() {
				return model.NewAppError("file.locked", http.StatusForbidden)
			}

			filesToTrash = append(filesToTrash, v)
			if v.IsFolder {
				f, err := a.Store.File.GetSubFilesByFolder(v.ID, v.Storage)
				if err != nil {
					tlog.Errorw("Failed to retrieve subfiles for folder",
						"file_id", v.ID,
						"error", err,
					)
					return model.NewAppError("file.retrieval_failed", http.StatusInternalServerError)
				}

				for _, sf := range f {
					if sf.ID != v.ID { // exclude the parent folder itself
						// Reject if a descendant is locked (mid-edit or held by a
						// running job like a cross-disk move): skipping strands a
						// live file under a trashed folder, forcing it corrupts
						// the lock holder.
						if sf.IsLocked() {
							return model.NewAppError("file.locked", http.StatusForbidden)
						}

						subFiles = append(subFiles, sf)
					}
				}
			}
		} else {
			filesToUnshare = append(filesToUnshare, v)
		}
	}

	if len(filesToTrash) > 0 {
		if err = a.Store.File.Trash(user.ID, filesToTrash, subFiles); err != nil {
			tlog.Errorw("Failed to trash files",
				"user_id", user.ID,
				"error", err,
			)
			return model.NewAppError("file.delete_failed", http.StatusInternalServerError)
		}

		for _, v := range filesToTrash {
			a.RecordActivity(user.ID, model.AppFiles, model.ActivityFileDelete, v.Parent, v.ID, map[string]any{
				"name": v.DisplayName,
				"type": v.Type,
			})
			a.CreateFileNotification(v, user, model.NOTIFICATION_FILE_DELETE, nil)
		}
	}

	for _, v := range filesToUnshare {
		if err := a.Store.File.RemoveUserFromShare(v.ID, user.ID); err != nil {
			tlog.Errorw("Failed to remove user from shared file",
				"file_id", v.ID,
				"user_id", user.ID,
				"error", err,
			)
			return model.NewAppError("file.delete_failed", http.StatusInternalServerError)
		}

		if v.IsFolder {
			f, err := a.Store.File.GetSubFilesByFolder(v.ID, v.Storage)
			if err != nil {
				tlog.Errorw("Failed to retrieve subfiles for folder",
					"file_id", v.ID,
					"error", err,
				)
				return model.NewAppError("file.retrieval_failed", http.StatusInternalServerError)
			}

			for i := range f {
				if err := a.Store.File.RemoveUserFromShare(f[i].ID, user.ID); err != nil {
					tlog.Errorw("Failed to remove user from shared subfile",
						"file_id", f[i].ID,
						"user_id", user.ID,
						"error", err,
					)
					return model.NewAppError("file.delete_failed", http.StatusInternalServerError)
				}
			}
		}
	}

	return nil
}

func (a *App) StartUploadSession(
	user model.User,
	r *http.Request,
) (*model.UploadSession, *model.AppError) {
	parentID := r.FormValue("parent_id")
	if parentID == "" {
		return nil, model.NewAppError("upload.missing_destination", http.StatusBadRequest)
	}

	name := r.FormValue("name")
	if name == "" {
		return nil, model.NewAppError("upload.missing_name", http.StatusBadRequest)
	}

	mime := r.FormValue("type")
	if mime == "" {
		mime = "application/octet-stream"
	}

	var totalSize int64
	if v := r.FormValue("full_size"); v != "" {
		if parsed, err := strconv.ParseInt(v, 10, 64); err == nil {
			totalSize = parsed
		}
	}

	parentFile, appErr := a.HasPermission(parentID, user)
	if appErr != nil {
		return nil, appErr
	}

	if totalSize > 0 {
		hasSpace, err := a.ownerHasSpace(newFileOwner(*parentFile, user), user, totalSize)
		if err != nil || !hasSpace {
			return nil, model.NewAppError("file.no_space", http.StatusInsufficientStorage)
		}
	}

	now := time.Now().Unix()
	session := model.UploadSession{
		ID:            model.NewID(),
		UserID:        user.ID,
		Storage:       parentFile.Storage,
		ContextType:   "files",
		ContextID:     parentFile.ID,
		FileName:      path.Base(name),
		MimeType:      mime,
		TotalSize:     totalSize,
		UploadedSize:  0,
		UploadedParts: 0,
		Status:        model.UploadStatusPending,
		CreatedAt:     now,
		UpdatedAt:     now,
		ExpiresAt:     time.Now().Add(24 * time.Hour).Unix(),
	}

	sessionFilePath, err := a.BuildFilePath(session.Storage, session.ID, model.AppFiles)
	if err != nil {
		tlog.Errorw("Failed to build upload session file path",
			"session_id", session.ID,
			"storage", session.Storage,
			"user_id", user.ID,
			"error", err,
		)
		return nil, model.NewAppError("upload.start_failed", http.StatusInternalServerError)
	}

	uploadID, err := a.FileStorageObjects[session.Storage].StartMultipartUpload(r.Context(), sessionFilePath)
	if err != nil {
		tlog.Errorw("Failed to start multipart upload",
			"session_id", session.ID,
			"storage", session.Storage,
			"user_id", user.ID,
			"error", err,
		)
		return nil, model.NewAppError("upload.start_failed", http.StatusInternalServerError)
	}

	session.UploadID = uploadID

	created, err := a.Store.UploadSession.Create(session)
	if err != nil {
		tlog.Errorw("Failed to create upload session",
			"session_id", session.ID,
			"user_id", user.ID,
			"error", err,
		)
		return nil, model.NewAppError("upload.start_failed", http.StatusInternalServerError)
	}

	return created, nil
}

func (a *App) uploadSingleFile(
	user model.User,
	r *http.Request,
) (*model.File, *model.AppError) {
	ctx := r.Context()

	parentFile, appErr := a.HasPermission(r.FormValue("parent_id"), user)
	if appErr != nil {
		return nil, appErr
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		return nil, model.NewAppError("upload.missing_name", http.StatusBadRequest)
	}

	defer file.Close()

	name := r.FormValue("name")
	if name == "" {
		return nil, model.NewAppError("upload.missing_name", http.StatusBadRequest)
	}

	if header.Size <= 0 {
		return nil, model.NewAppError("upload.invalid_size", http.StatusBadRequest)
	}

	mime := r.FormValue("type")
	if mime == "" {
		mime = "application/octet-stream"
	}

	owner := newFileOwner(*parentFile, user)
	hasSpace, err := a.ownerHasSpace(owner, user, header.Size)
	if err != nil || !hasSpace {
		return nil, model.NewAppError("file.no_space", http.StatusInsufficientStorage)
	}

	newID := model.NewID()

	newFilePath, err := a.BuildFilePath(parentFile.Storage, newID, model.AppFiles)
	if err != nil {
		tlog.Errorw("Failed to build file path",
			"file_id", newID,
			"storage", parentFile.Storage,
			"user_id", user.ID,
			"error", err,
		)
		return nil, model.NewAppError("file.create_failed", http.StatusInternalServerError)
	}

	if err := a.FileStorageObjects[parentFile.Storage].WriteFile(ctx, newFilePath, file, header.Size); err != nil {
		tlog.Errorw("Failed to write file to storage",
			"file_id", newID,
			"storage", parentFile.Storage,
			"user_id", user.ID,
			"error", err,
		)
		return nil, model.NewAppError("file.create_failed", http.StatusInternalServerError)
	}

	newFile := model.File{
		ID:          newID,
		Owner:       owner,
		Parent:      parentFile.ID,
		Storage:     parentFile.Storage,
		Size:        header.Size,
		Name:        name,
		DisplayName: name,
		Type:        mime,
	}

	fileID, err := a.Store.File.Create(newFile)
	if err != nil {
		tlog.Errorw("Failed to add file record",
			"file_id", newID,
			"user_id", user.ID,
			"error", err,
		)
		return nil, model.NewAppError("file.create_failed", http.StatusInternalServerError)
	}

	if err = a.createImageThumbnail(ctx, newFile); err != nil {
		tlog.Errorw("Failed to create image thumbnail",
			"file_id", newID,
			"error", err,
		)
	}

	a.RecordActivity(user.ID, model.AppFiles, model.ActivityFileUpload, newFile.Parent, newFile.ID, map[string]any{
		"name": newFile.DisplayName,
		"type": newFile.Type,
		"size": strconv.FormatInt(newFile.Size, 10),
	})

	createdFile, err := a.Store.File.Get(*fileID)
	if err != nil {
		tlog.Errorw("Failed to retrieve created file",
			"file_id", *fileID,
			"error", err,
		)
		return nil, model.NewAppError("file.create_failed", http.StatusInternalServerError)
	}

	return createdFile, nil
}

func (a *App) uploadChunk(
	user model.User,
	r *http.Request,
) *model.AppError {
	ctx := r.Context()

	sessionID := r.FormValue("upload_session_id")
	if sessionID == "" {
		return model.NewAppError("upload.missing_destination", http.StatusBadRequest)
	}

	partNumberStr := r.FormValue("part_number")
	if partNumberStr == "" {
		return model.NewAppError("upload.chunk_failed", http.StatusBadRequest)
	}

	partNumber, err := strconv.Atoi(partNumberStr)
	if err != nil || partNumber < 1 {
		return model.NewAppError("upload.invalid_size", http.StatusBadRequest)
	}

	session, err := a.Store.UploadSession.Get(sessionID)
	if err != nil {
		tlog.Errorw("Failed to retrieve upload session",
			"session_id", sessionID,
			"error", err,
		)
		return model.NewAppError("upload.not_found", http.StatusInternalServerError)
	}

	if session == nil {
		return model.NewAppError("upload.not_found", http.StatusNotFound)
	}

	if session.UserID != user.ID {
		return model.NewAppError("upload.forbidden", http.StatusForbidden)
	}

	if time.Unix(session.ExpiresAt, 0).Before(time.Now()) {
		go a.cleanupExpiredSession(session)
		return model.NewAppError("upload.expired", http.StatusGone)
	}

	if session.Status != model.UploadStatusPending && session.Status != model.UploadStatusUploading {
		return model.NewAppError("upload.conflict", http.StatusConflict)
	}

	if session.UploadID == "" {
		return model.NewAppError("upload.start_failed", http.StatusBadRequest)
	}

	existingParts, err := session.GetParts()
	if err != nil {
		tlog.Errorw("Failed to parse upload session parts",
			"session_id", sessionID,
			"error", err,
		)
		existingParts = make(map[int]string)
	}

	if _, exists := existingParts[partNumber]; exists {
		tlog.Warnw("Part already uploaded for session",
			"session_id", sessionID,
			"part_number", partNumber,
		)
		return nil
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		return model.NewAppError("upload.chunk_failed", http.StatusBadRequest)
	}

	defer file.Close()

	if header.Size <= 0 {
		return model.NewAppError("upload.invalid_size", http.StatusBadRequest)
	}

	if session.TotalSize > 0 && (session.UploadedSize+header.Size) > session.TotalSize {
		return model.NewAppError("upload.invalid_size", http.StatusBadRequest)
	}

	backend, exists := a.FileStorageObjects[session.Storage]
	if !exists {
		tlog.Errorw("Storage backend not found",
			"session_id", sessionID,
			"storage", session.Storage,
		)
		return model.NewAppError("storage.not_available", http.StatusInternalServerError)
	}

	sessionPath, err := a.BuildFilePath(session.Storage, session.ID, session.ContextType)
	if err != nil {
		tlog.Errorw("Failed to build session file path",
			"session_id", sessionID,
			"error", err,
		)
		return model.NewAppError("upload.chunk_failed", http.StatusInternalServerError)
	}

	etag, err := backend.UploadPart(ctx, sessionPath, session.UploadID, partNumber, file, header.Size)
	if err != nil {
		tlog.Errorw("Failed to upload chunk",
			"session_id", sessionID,
			"part_number", partNumber,
			"error", err,
		)
		return model.NewAppError("upload.chunk_failed", http.StatusInternalServerError)
	}

	existingParts[partNumber] = etag

	jsonBytes, err := json.Marshal(existingParts)
	if err != nil {
		tlog.Errorw("Failed to marshal upload parts",
			"session_id", sessionID,
			"error", err,
		)
		return model.NewAppError("upload.chunk_failed", http.StatusInternalServerError)
	}

	partsJSON := string(jsonBytes)
	session.PartsJSON = &partsJSON

	now := time.Now().Unix()

	if err = a.Store.UploadSession.UpdateParts(
		sessionID,
		partsJSON,
		session.UploadedSize+header.Size,
		len(existingParts),
		now,
	); err != nil {
		tlog.Errorw("Failed to update upload session parts",
			"session_id", sessionID,
			"error", err,
		)
		return model.NewAppError("upload.chunk_failed", http.StatusInternalServerError)
	}

	if session.Status == model.UploadStatusPending {
		if err = a.Store.UploadSession.UpdateStatus(sessionID, model.UploadStatusUploading, now); err != nil {
			tlog.Errorw("Failed to update upload session status",
				"session_id", sessionID,
				"error", err,
			)
		}
	}

	return nil
}

func (a *App) cleanupExpiredSession(session *model.UploadSession) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if backend, exists := a.FileStorageObjects[session.Storage]; exists && session.UploadID != "" {
		if p, err := a.BuildFilePath(session.Storage, session.ID, session.ContextType); err != nil {
			tlog.Errorw("Failed to build path for expired session abort",
				"session_id", session.ID,
				"error", err,
			)
		} else if err := backend.AbortMultipartUpload(ctx, p, session.UploadID); err != nil {
			tlog.Errorw("Failed to abort expired upload session",
				"session_id", session.ID,
				"error", err,
			)
		}
	}

	if err := a.Store.UploadSession.Delete(session.ID); err != nil {
		tlog.Errorw("Failed to delete expired upload session",
			"session_id", session.ID,
			"error", err,
		)
	}
}

func (a *App) AbortUploadSession(user model.User, sessionID string) *model.AppError {
	session, err := a.Store.UploadSession.Get(sessionID)
	if err != nil {
		tlog.Errorw("Failed to retrieve upload session",
			"session_id", sessionID,
			"error", err,
		)
		return model.NewAppError("upload.not_found", http.StatusInternalServerError)
	}

	if session == nil {
		return model.NewAppError("upload.not_found", http.StatusNotFound)
	}

	if session.UserID != user.ID {
		return model.NewAppError("upload.forbidden", http.StatusForbidden)
	}

	if backend, exists := a.FileStorageObjects[session.Storage]; exists && session.UploadID != "" {
		abortPath, err := a.BuildFilePath(session.Storage, session.ID, session.ContextType)
		if err != nil {
			tlog.Errorw("Failed to build path for upload abort",
				"session_id", sessionID,
				"error", err,
			)
			return model.NewAppError("upload.abort_failed", http.StatusInternalServerError)
		}

		if err := backend.AbortMultipartUpload(context.Background(), abortPath, session.UploadID); err != nil {
			tlog.Errorw("Failed to abort multipart upload",
				"session_id", sessionID,
				"error", err,
			)
			return model.NewAppError("upload.abort_failed", http.StatusInternalServerError)
		}
	}

	if err = a.Store.UploadSession.Delete(session.ID); err != nil {
		tlog.Errorw("Failed to delete upload session",
			"session_id", sessionID,
			"error", err,
		)
		return model.NewAppError("upload.abort_failed", http.StatusInternalServerError)
	}

	return nil
}

func (a *App) finishUpload(
	user model.User,
	r *http.Request,
) (*model.File, *model.AppError) {
	sessionID := r.FormValue("upload_session_id")
	if sessionID == "" {
		return nil, model.NewAppError("upload.missing_destination", http.StatusBadRequest)
	}

	session, err := a.Store.UploadSession.Get(sessionID)
	if err != nil {
		tlog.Errorw("Failed to retrieve upload session",
			"session_id", sessionID,
			"error", err,
		)
		return nil, model.NewAppError("upload.not_found", http.StatusInternalServerError)
	}

	if session == nil {
		return nil, model.NewAppError("upload.not_found", http.StatusNotFound)
	}

	if session.UserID != user.ID {
		return nil, model.NewAppError("upload.forbidden", http.StatusForbidden)
	}

	if session.UploadedSize != session.TotalSize {
		return nil, model.NewAppError("upload.incomplete", http.StatusConflict)
	}

	parts, err := session.GetParts()
	if err != nil {
		tlog.Errorw("Failed to decode upload session parts",
			"session_id", sessionID,
			"error", err,
		)
		return nil, model.NewAppError("upload.finalize_failed", http.StatusInternalServerError)
	}

	if len(parts) == 0 {
		return nil, model.NewAppError("upload.incomplete", http.StatusBadRequest)
	}

	completePath, err := a.BuildFilePath(session.Storage, session.ID, session.ContextType)
	if err != nil {
		tlog.Errorw("Failed to build path for upload completion",
			"session_id", sessionID,
			"error", err,
		)
		return nil, model.NewAppError("upload.finalize_failed", http.StatusInternalServerError)
	}

	if err = a.FileStorageObjects[session.Storage].CompleteMultipartUpload(
		r.Context(),
		completePath,
		session.UploadID,
		parts,
	); err != nil {
		tlog.Errorw("Failed to complete multipart upload",
			"session_id", sessionID,
			"error", err,
		)
		return nil, model.NewAppError("upload.finalize_failed", http.StatusInternalServerError)
	}

	parentFile, appErr := a.HasPermission(session.ContextID, user)
	if appErr != nil {
		return nil, appErr
	}

	owner := newFileOwner(*parentFile, user)
	if hasSpace, err := a.ownerHasSpace(owner, user, session.TotalSize); err != nil || !hasSpace {
		if err := a.FileStorageObjects[session.Storage].RemoveFile(r.Context(), completePath); err != nil {
			tlog.Errorw("Failed to remove upload after storage limit check",
				"session_id", sessionID,
				"storage", session.Storage,
				"error", err,
			)
		}

		return nil, model.NewAppError("file.no_space", http.StatusInsufficientStorage)
	}

	newFile := model.File{
		ID:          session.ID,
		Owner:       owner,
		Parent:      session.ContextID,
		Storage:     session.Storage,
		Size:        session.TotalSize,
		Name:        session.FileName,
		DisplayName: session.FileName,
		Type:        session.MimeType,
	}

	fileID, err := a.Store.File.Create(newFile)
	if err != nil {
		tlog.Errorw("Failed to add file record",
			"session_id", sessionID,
			"error", err,
		)
		return nil, model.NewAppError("upload.finalize_failed", http.StatusInternalServerError)
	}

	if err = a.Store.UploadSession.Delete(sessionID); err != nil {
		tlog.Errorw("Failed to delete upload session",
			"session_id", sessionID,
			"error", err,
		)
	}

	createdFile, err := a.Store.File.Get(*fileID)
	if err != nil {
		tlog.Errorw("Failed to retrieve created file",
			"file_id", *fileID,
			"error", err,
		)
		return nil, model.NewAppError("upload.finalize_failed", http.StatusInternalServerError)
	}

	if err = a.createImageThumbnail(r.Context(), newFile); err != nil {
		tlog.Errorw("Failed to create image thumbnail",
			"file_id", newFile.ID,
			"error", err,
		)
	}

	a.RecordActivity(user.ID, model.AppFiles, model.ActivityFileUpload, newFile.Parent, newFile.ID, map[string]any{
		"name": newFile.DisplayName,
		"type": newFile.Type,
		"size": strconv.FormatInt(newFile.Size, 10),
	})

	return createdFile, nil
}

func (a *App) UploadFile(
	user model.User,
	r *http.Request,
) (*model.File, *model.AppError) {

	action := r.URL.Query().Get("action")
	switch action {
	case "upload":
		return a.uploadSingleFile(user, r)
	case "chunk":
		if err := a.uploadChunk(user, r); err != nil {
			return nil, err
		}

		return nil, nil
	case "finish":
		return a.finishUpload(user, r)
	default:
		return nil, model.NewAppError("upload.invalid_action", http.StatusBadRequest)
	}
}

func (a *App) createImageThumbnail(ctx context.Context, file model.File) error {
	if !isThumbnailType(file.Type) {
		return nil
	}

	srcPath, err := a.BuildFilePath(file.Storage, file.ID, model.AppFiles)
	if err != nil {
		return err
	}

	fr, err := a.FileStorageObjects[file.Storage].ReadFile(ctx, srcPath)
	if err != nil {
		return err
	}

	defer fr.Close()

	src, err := imaging.Decode(fr, imaging.AutoOrientation(true))
	if err != nil {
		return err
	}

	thumb := imaging.Resize(src, 300, 0, imaging.Lanczos)

	var buf bytes.Buffer
	if err := imaging.Encode(&buf, thumb, imaging.JPEG, imaging.JPEGQuality(85)); err != nil {
		return err
	}

	thumbPath, err := a.BuildFilePath(file.Storage, "thumbnails/"+file.ID+".jpg", model.AppFiles)
	if err != nil {
		return err
	}

	return a.FileStorageObjects[file.Storage].WriteFile(ctx, thumbPath, &buf, int64(buf.Len()))
}

func (a *App) Rename(user model.User, id string, newName string) *model.AppError {
	file, appErr := a.HasPermission(id, user)
	if appErr != nil {
		return appErr
	}

	if file.IsLocked() {
		return model.NewAppError("file.locked", http.StatusForbidden)
	}

	file.OldName = file.DisplayName
	file.DisplayName = newName

	if err := a.Store.File.Rename(*file, newName, newName); err != nil {
		tlog.Errorw("Failed to rename file",
			"file_id", id,
			"error", err,
		)
		return model.NewAppError("file.rename_failed", http.StatusInternalServerError)
	}

	a.CreateFileNotification(*file, user, model.NOTIFICATION_FILE_RENAME, nil)

	a.RecordActivity(user.ID, model.AppFiles, model.ActivityFileRename, file.Parent, file.ID, map[string]any{
		"from": file.OldName,
		"to":   newName,
		"type": file.Type,
	})

	return nil
}

func (a *App) CreateZip(
	user model.User,
	fileID string,
	rw io.Writer,
) (*string, *model.AppError) {
	root, appErr := a.HasPermission(fileID, user)
	if appErr != nil {
		return nil, appErr
	}

	filePath := path.Join("tmp", model.NewID(), root.Name+".zip")
	ctx := context.Background()
	w := zip.NewWriter(rw)
	defer w.Close()

	rootFilePath, err := a.BuildFilePath(root.Storage, root.ID, model.AppFiles)
	if err != nil {
		tlog.Errorw("Failed to build file path",
			"file_id", fileID,
			"error", err,
		)
		return nil, model.NewAppError("file.archive_failed", http.StatusInternalServerError)
	}

	if !root.IsFolder {
		zf, err := w.Create(root.Name)
		if err != nil {
			tlog.Errorw("Failed to create zip entry",
				"file_id", fileID,
				"error", err,
			)
			return nil, model.NewAppError("file.archive_failed", http.StatusInternalServerError)
		}

		fr, err := a.FileStorageObjects[root.Storage].ReadFile(ctx, rootFilePath)
		if err != nil {
			tlog.Errorw("Failed to read file for archiving",
				"file_id", fileID,
				"error", err,
			)
			return nil, model.NewAppError("file.archive_failed", http.StatusInternalServerError)
		}

		defer fr.Close()

		if _, err = io.Copy(zf, fr); err != nil {
			tlog.Errorw("Failed to copy file content to zip",
				"file_id", fileID,
				"error", err,
			)
			return nil, model.NewAppError("file.archive_failed", http.StatusInternalServerError)
		}

		return &filePath, nil
	}

	if appErr := a.zipFolderRecursive(ctx, user.ID, *root, rootFilePath, w); appErr != nil {
		return nil, appErr
	}

	return &filePath, nil
}

func (a *App) zipFolderRecursive(
	ctx context.Context,
	userID string,
	parent model.File,
	currentPath string,
	w *zip.Writer,
) *model.AppError {
	children, err := a.listChildren(ctx, userID, parent)
	if err != nil {
		tlog.Errorw("Failed to retrieve folder children",
			"folder_id", parent.ID,
			"error", err,
		)
		return model.NewAppError("file.archive_failed", http.StatusInternalServerError)
	}

	for _, f := range children {
		zipPath := path.Join(currentPath, f.Name)

		if f.IsFolder {
			if err := a.zipFolderRecursive(ctx, userID, f, zipPath, w); err != nil {
				return err
			}

			continue
		}

		zf, err := w.Create(zipPath)
		if err != nil {
			tlog.Errorw("Failed to create zip entry",
				"file_id", f.ID,
				"error", err,
			)
			return model.NewAppError("file.archive_failed", http.StatusInternalServerError)
		}

		fPath, err := a.BuildFilePath(f.Storage, f.ID, model.AppFiles)
		if err != nil {
			tlog.Errorw("Failed to build file path for archiving",
				"file_id", f.ID,
				"error", err,
			)
			return model.NewAppError("file.archive_failed", http.StatusInternalServerError)
		}

		fr, err := a.FileStorageObjects[f.Storage].ReadFile(ctx, fPath)
		if err != nil {
			tlog.Errorw("Failed to read file for archiving",
				"file_id", f.ID,
				"error", err,
			)
			return model.NewAppError("file.archive_failed", http.StatusInternalServerError)
		}

		_, err = io.Copy(zf, fr)
		fr.Close()
		if err != nil {
			tlog.Errorw("Failed to copy file content to zip",
				"file_id", f.ID,
				"error", err,
			)
			return model.NewAppError("file.archive_failed", http.StatusInternalServerError)
		}
	}

	return nil
}

func (a *App) LoadThumbnail(user model.User, id string) (io.ReadCloser, *model.AppError) {
	file, appErr := a.HasPermission(id, user)
	if appErr != nil {
		return nil, appErr
	}

	fr, err := a.openThumbnail(context.Background(), *file)
	if err != nil {
		tlog.Errorw("Failed to read thumbnail file",
			"file_id", id,
			"error", err,
		)
		return nil, model.NewAppError("file.not_found", http.StatusNotFound)
	}

	return fr, nil
}

func (a *App) openThumbnail(ctx context.Context, file model.File) (io.ReadCloser, error) {
	thumbPath, err := a.BuildFilePath(file.Storage, "thumbnails/"+file.ID+".jpg", model.AppFiles)
	if err != nil {
		return nil, err
	}

	backend := a.FileStorageObjects[file.Storage]

	fr, err := backend.ReadFile(ctx, thumbPath)
	if err == nil || !isThumbnailType(file.Type) {
		return fr, err
	}

	if err = a.createImageThumbnail(ctx, file); err != nil {
		return nil, err
	}

	return backend.ReadFile(ctx, thumbPath)
}

func (a *App) isFreeSpace(user model.User, fileSize int64) (bool, error) {
	// SaaS tenants share one bucket, so the licence caps the instance as a whole
	// on top of whatever per-user limit applies. Self-hosted licences carry no
	// ceiling because the customer supplies the storage.
	if limit := a.Server.License.StorageLimitBytes(); limit > 0 {
		total, err := a.Store.File.GetTotalStorage()
		if err != nil {
			return false, err
		}

		if total+fileSize > limit {
			return false, nil
		}
	}

	if user.StorageLimit <= 0 {
		return true, nil
	}

	used, err := a.Store.File.GetStorageUsed(user.ID)
	if err != nil {
		return false, err
	}

	return used+fileSize <= user.StorageLimit, nil
}

// newFileOwner is who a file created under parent belongs to, and so whose
// storage limit it counts against: the folder's owner, even when someone they
// shared it with uploads, or the uploader at the root of their own drive.
func newFileOwner(parent model.File, uploader model.User) string {
	if parent.Type == "cloud#drive" {
		return uploader.ID
	}

	return parent.Owner
}

func (a *App) ownerHasSpace(ownerID string, uploader model.User, size int64) (bool, error) {
	owner := uploader
	if ownerID != uploader.ID {
		u, err := a.Store.User.Get(ownerID)
		if err != nil {
			return false, err
		}

		if u == nil {
			return false, nil
		}

		owner = *u
	}

	return a.isFreeSpace(owner, size)
}
