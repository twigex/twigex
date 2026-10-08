// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"archive/zip"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"path"
	"strconv"
	"time"

	"github.com/twigex/twigex/crypto"
	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/tlog"
)

func (a *App) maxPublicUploadBytes() int64 {
	mb := 2048
	if a.ConfigStore.Config.FileSettings.MaxPublicUploadSize != nil {
		mb = *a.ConfigStore.Config.FileSettings.MaxPublicUploadSize
	}

	return int64(mb) * 1024 * 1024
}

// resolveUploadShare validates that a share exists, the caller is authorized
// (password), and the target is a folder accepting uploads.
// resolveUploadShare resolves the upload target: the shared root, or a subtree
// subfolder when childID is set (so uploads land in the folder the recipient is
// browsing, not always the root). The target must be a folder on an upload link.
func (a *App) resolveUploadShare(ctx context.Context, shareToken, childID, accessToken string) (*model.ExternalShare, *model.File, *model.AppError) {
	link, target, appErr := a.publicTarget(ctx, shareToken, childID, accessToken)
	if appErr != nil {
		return nil, nil, appErr
	}

	if !link.AllowUpload || !target.IsFolder {
		return nil, nil, model.NewAppError("link.upload_forbidden", http.StatusForbidden)
	}

	return link, target, nil
}

// shareOwnerWithSpace resolves the folder owner and confirms they have room for
// size. An owner near their limit is surfaced to the anonymous uploader as a generic
// failure (never revealing the owner's storage state); the real reason is logged.
func (a *App) shareOwnerWithSpace(folder *model.File, size int64) (*model.User, *model.AppError) {
	owner, err := a.Store.User.Get(folder.Owner)
	if err != nil {
		tlog.Errorw("Failed to retrieve share owner for public upload", "file_id", folder.ID, "error", err)
		return nil, model.NewAppError("link.upload_failed", http.StatusInternalServerError)
	}

	if owner == nil {
		return nil, model.NewAppError("link.not_found", http.StatusNotFound)
	}

	free, err := a.isFreeSpace(*owner, size)
	if err != nil {
		tlog.Errorw("Failed to check owner storage limit for public upload", "owner_id", owner.ID, "error", err)
		return nil, model.NewAppError("link.upload_failed", http.StatusInternalServerError)
	}

	if !free {
		tlog.Warnw("Public upload rejected: owner storage running low",
			"owner_id", owner.ID,
			"file_id", folder.ID,
		)
		return nil, model.NewAppError("link.upload_failed", http.StatusInternalServerError)
	}

	return owner, nil
}

const (
	shareAccessPrefix = "share_access:"
	shareAccessTTL    = time.Hour
)

func (a *App) CreateLink(ctx context.Context, user model.User, req model.Request) (*model.SharedLinks, *model.AppError) {
	if !a.SessionHasPermission(user, model.FilePermissions.PermissionShareFiles) {
		return nil, model.NewAppError("permission.forbidden", http.StatusForbidden)
	}

	file, appErr := a.HasPermission(req.Item, user)
	if appErr != nil {
		return nil, appErr
	}

	if file.Owner != user.ID && !file.AccessLevel.CanShare() {
		return nil, model.NewAppError("file.share_forbidden", http.StatusForbidden)
	}

	if !req.AllowView && !req.AllowDownload && !req.AllowUpload {
		return nil, model.NewAppError("link.no_access", http.StatusBadRequest)
	}

	if req.AllowUpload && !file.IsFolder {
		return nil, model.NewAppError("link.upload_requires_folder", http.StatusBadRequest)
	}

	if req.PasswordProtected && req.Password != "" {
		hashed, err := crypto.HashPassword(req.Password)
		if err != nil {
			tlog.Errorw("Failed to hash share link password",
				"file_id", file.ID,
				"error", err,
			)
			return nil, model.NewAppError("link.create_failed", http.StatusInternalServerError)
		}

		req.Password = hashed
	} else {
		req.PasswordProtected = false
		req.Password = ""
	}

	token := crypto.RandomText()

	req.Folder = file.IsFolder
	req.Parent = file.Parent

	link, err := a.Store.Share.Create(ctx, user.ID, token, req)
	if err != nil {
		tlog.Errorw("Failed to create share link",
			"file_id", file.ID,
			"error", err,
		)
		return nil, model.NewAppError("link.create_failed", http.StatusInternalServerError)
	}

	link.FullName = user.Name + " " + user.LastName

	if err := a.Store.File.Update(*file); err != nil {
		tlog.Errorw("Failed to mark file shared after link creation",
			"file_id", file.ID,
			"error", err,
		)
		return nil, model.NewAppError("link.create_failed", http.StatusInternalServerError)
	}

	a.RecordActivity(user.ID, model.AppFiles, model.ActivityFileShare, file.Parent, file.ID, map[string]any{
		"name": file.DisplayName,
		"type": file.Type,
		"link": true,
	})

	return link, nil
}

func (a *App) UpdateLink(ctx context.Context, user model.User, l model.LinkUpdate) *model.AppError {
	link, err := a.Store.Share.GetByToken(ctx, l.Token)
	if err != nil {
		tlog.Errorw("Failed to retrieve share link for update",
			"error", err,
		)
		return model.NewAppError("link.update_failed", http.StatusInternalServerError)
	}

	if link == nil {
		return model.NewAppError("link.not_found", http.StatusNotFound)
	}
	// Governed by sharing capability on the file, not the link's creator.
	f, appErr := a.HasPermission(link.FileID, user)
	if appErr != nil {
		return appErr
	}

	if f.Owner != user.ID && !f.AccessLevel.CanShare() {
		return model.NewAppError("link.forbidden", http.StatusForbidden)
	}

	if !l.AllowView && !l.AllowDownload && !l.AllowUpload {
		return model.NewAppError("link.no_access", http.StatusBadRequest)
	}

	if l.PasswordProtected && l.Password != "" {
		hashed, err := crypto.HashPassword(l.Password)
		if err != nil {
			tlog.Errorw("Failed to hash share link password",
				"error", err,
			)
			return model.NewAppError("link.update_failed", http.StatusInternalServerError)
		}

		l.Password = hashed
	} else if l.PasswordProtected {
		l.Password = link.Password
	} else {
		l.Password = ""
	}

	if err := a.Store.Share.Update(ctx, link.Owner, link.Source, l); err != nil {
		tlog.Errorw("Failed to update share link",
			"error", err,
		)
		return model.NewAppError("link.update_failed", http.StatusInternalServerError)
	}

	return nil
}

func (a *App) DeleteLink(ctx context.Context, user model.User, token string) *model.AppError {
	link, err := a.Store.Share.GetByToken(ctx, token)
	if err != nil {
		tlog.Errorw("Failed to retrieve share link for delete",
			"error", err,
		)
		return model.NewAppError("link.delete_failed", http.StatusInternalServerError)
	}

	if link == nil {
		return model.NewAppError("link.not_found", http.StatusNotFound)
	}

	f, appErr := a.HasPermission(link.FileID, user)
	if appErr != nil {
		return appErr
	}

	if f.Owner != user.ID && !f.AccessLevel.CanShare() {
		return model.NewAppError("link.forbidden", http.StatusForbidden)
	}

	if err := a.Store.Share.Delete(ctx, link.Owner, link.Source); err != nil {
		tlog.Errorw("Failed to delete share link",
			"error", err,
		)
		return model.NewAppError("link.delete_failed", http.StatusInternalServerError)
	}

	return nil
}

func (a *App) resolvePublicShare(ctx context.Context, token string) (*model.ExternalShare, *model.File, *model.AppError) {
	link, err := a.Store.Share.GetByToken(ctx, token)
	if err != nil {
		tlog.Errorw("Failed to retrieve public share", "error", err)
		return nil, nil, model.NewAppError("link.retrieval_failed", http.StatusInternalServerError)
	}

	if link == nil {
		return nil, nil, model.NewAppError("link.not_found", http.StatusNotFound)
	}

	if link.Expiration != 0 && time.Now().Unix() > link.Expiration {
		return nil, nil, model.NewAppError("link.expired", http.StatusNotFound)
	}

	file, err := a.Store.File.Get(link.FileID)
	if err != nil {
		tlog.Errorw("Failed to retrieve public share file", "file_id", link.FileID, "error", err)
		return nil, nil, model.NewAppError("link.retrieval_failed", http.StatusInternalServerError)
	}

	if file == nil || file.DeletedAt != 0 {
		return nil, nil, model.NewAppError("link.not_found", http.StatusNotFound)
	}

	return link, file, nil
}

func (a *App) validatePublicAccess(link *model.ExternalShare, accessToken string) bool {
	if !link.PasswordProtected {
		return true
	}

	if accessToken == "" {
		return false
	}

	var sharedToken string
	if err := a.Store.Auth.Get(shareAccessPrefix+accessToken, &sharedToken); err != nil {
		return false
	}

	return sharedToken == link.ShareToken
}

func (a *App) AuthenticatePublicShare(ctx context.Context, token, password, clientIP string) (string, *model.AppError) {
	allowed, err := a.shareAuthLimiter().Allow(clientIP, token)
	if err != nil {
		tlog.Errorw("Failed to check share auth rate limit", "error", err)
		return "", model.NewAppError("link.auth_failed", http.StatusInternalServerError)
	}

	if !allowed {
		return "", model.NewAppError("link.rate_limited", http.StatusTooManyRequests)
	}

	link, _, appErr := a.resolvePublicShare(ctx, token)
	if appErr != nil {
		return "", appErr
	}

	if !link.PasswordProtected {
		return "", nil
	}

	if !crypto.PasswordMatches(password, link.Password) {
		return "", model.NewAppError("link.password_invalid", http.StatusForbidden)
	}

	accessToken := crypto.RandomText()
	if err := a.Store.Auth.Set(shareAccessPrefix+accessToken, shareAccessTTL, link.ShareToken); err != nil {
		tlog.Errorw("Failed to store share access token", "error", err)
		return "", model.NewAppError("link.auth_failed", http.StatusInternalServerError)
	}

	return accessToken, nil
}

// publicTarget resolves the share, checks the password gate, and returns the
// target file: the shared root, or a subtree descendant when childID is set
// (subtree-validated so a link cannot reach files outside its folder).
func (a *App) publicTarget(ctx context.Context, token, childID, accessToken string) (*model.ExternalShare, *model.File, *model.AppError) {
	link, root, appErr := a.resolvePublicShare(ctx, token)
	if appErr != nil {
		return nil, nil, appErr
	}

	if !a.validatePublicAccess(link, accessToken) {
		return nil, nil, model.NewAppError("link.password_required", http.StatusUnauthorized)
	}

	target := root
	if childID != "" && childID != root.ID {
		if !root.IsFolder {
			return nil, nil, model.NewAppError("link.not_found", http.StatusNotFound)
		}

		descendant, appErr := a.publicSubtreeFile(*root, childID)
		if appErr != nil {
			return nil, nil, appErr
		}

		target = descendant
	}

	return link, target, nil
}

func (a *App) GetPublicShare(ctx context.Context, token, childID, accessToken string) (*model.PublicFile, *model.AppError) {
	link, target, appErr := a.publicTarget(ctx, token, childID, accessToken)
	if appErr != nil {
		return nil, appErr
	}

	owner, err := a.Store.User.Get(link.Owner)
	if err != nil {
		tlog.Errorw("Failed to retrieve public share owner", "file_id", target.ID, "error", err)
		return nil, model.NewAppError("link.retrieval_failed", http.StatusInternalServerError)
	}

	if owner == nil {
		return nil, model.NewAppError("link.not_found", http.StatusNotFound)
	}

	id := link.ShareToken
	if childID != "" {
		id = target.ID
	}

	pf := &model.PublicFile{
		ID:            id,
		Name:          target.DisplayName,
		Size:          target.Size,
		Type:          target.Type,
		IsFolder:      target.IsFolder,
		Message:       link.Message,
		AllowView:     link.AllowView,
		AllowDownload: link.AllowDownload,
		AllowUpload:   link.AllowUpload,
		AllowEdit:     link.AllowEdit,
		User:          model.UserInfo{Name: owner.Name, LastName: owner.LastName},
	}

	if target.IsFolder && link.AllowView {
		children, appErr := a.publicFolderChildren(*target, link)
		if appErr != nil {
			return nil, appErr
		}

		pf.Children = children
	}

	if err := a.Store.Share.IncrementAccessed(ctx, link.ShareToken); err != nil {
		tlog.Warnw("Failed to increment share access count", "share_token", link.ShareToken)
	}

	return pf, nil
}

func (a *App) publicFolderChildren(folder model.File, link *model.ExternalShare) ([]model.PublicFile, *model.AppError) {
	subs, err := a.Store.File.GetSubFilesByFolder(folder.ID, folder.Storage)
	if err != nil {
		tlog.Errorw("Failed to retrieve public folder children", "file_id", folder.ID, "error", err)
		return nil, model.NewAppError("link.retrieval_failed", http.StatusInternalServerError)
	}

	children := make([]model.PublicFile, 0)
	for _, sub := range subs {
		if sub.ID == folder.ID || sub.Parent != folder.ID || sub.DeletedAt != 0 {
			continue
		}

		children = append(children, model.PublicFile{
			ID:            sub.ID,
			Name:          sub.DisplayName,
			Size:          sub.Size,
			Type:          sub.Type,
			IsFolder:      sub.IsFolder,
			AllowView:     link.AllowView,
			AllowDownload: link.AllowDownload,
		})
	}

	return children, nil
}

// PublicDownloadTarget validates a public download and returns the target, which
// may be a file (stream it) or a folder (zip it via ZipPublicFolderContents).
func (a *App) PublicDownloadTarget(ctx context.Context, token, childID, accessToken string) (*model.File, *model.AppError) {
	link, target, appErr := a.publicTarget(ctx, token, childID, accessToken)
	if appErr != nil {
		return nil, appErr
	}

	if !link.AllowDownload {
		return nil, model.NewAppError("link.download_forbidden", http.StatusForbidden)
	}

	if link.MaxDownloads > 0 && link.Downloaded >= link.MaxDownloads {
		return nil, model.NewAppError("link.download_limit", http.StatusForbidden)
	}

	if err := a.Store.Share.IncrementDownloaded(ctx, link.ShareToken); err != nil {
		tlog.Warnw("Failed to increment share download count", "share_token", link.ShareToken)
	}

	owner, err := a.Store.User.Get(link.Owner)
	if err == nil && owner != nil {
		go a.CreateFileNotification(*target, *owner, model.NOTIFICATION_FILE_PUBLIC_DOWNLOAD, nil)
	}

	return target, nil
}

// PublicViewFile returns a file for inline preview (image/video/pdf/office). It
// is gated on allow_view (not allow_download) and does not count as a download,
// so view-only shares can preview without consuming the download budget.
func (a *App) PublicViewFile(ctx context.Context, token, childID, accessToken string) (*model.File, *model.AppError) {
	link, target, appErr := a.publicTarget(ctx, token, childID, accessToken)
	if appErr != nil {
		return nil, appErr
	}

	if !link.AllowView {
		return nil, model.NewAppError("link.view_forbidden", http.StatusForbidden)
	}

	if target.IsFolder {
		return nil, model.NewAppError("file.unknown_type", http.StatusBadRequest)
	}

	return target, nil
}

// ZipPublicFolderContents streams the non-deleted files of a shared folder's
// subtree as a zip. The caller (API) has already validated access via
// PublicDownloadTarget; this only reads and archives.
func (a *App) ZipPublicFolderContents(ctx context.Context, folder model.File, rw io.Writer) *model.AppError {
	subs, err := a.Store.File.GetSubFilesByFolder(folder.ID, folder.Storage)
	if err != nil {
		tlog.Errorw("Failed to list folder for public zip", "file_id", folder.ID, "error", err)
		return model.NewAppError("file.archive_failed", http.StatusInternalServerError)
	}

	byID := make(map[string]model.File, len(subs))
	for _, f := range subs {
		byID[f.ID] = f
	}

	w := zip.NewWriter(rw)
	defer w.Close()

	for _, f := range subs {
		if f.ID == folder.ID || f.IsFolder || f.DeletedAt != 0 {
			continue
		}

		rel := publicZipPath(byID, folder.ID, f)
		if rel == "" {
			continue
		}

		fPath, err := a.BuildFilePath(f.Storage, f.ID, model.AppFiles)
		if err != nil {
			tlog.Errorw("Failed to build path for public zip entry", "file_id", f.ID, "error", err)
			return model.NewAppError("file.archive_failed", http.StatusInternalServerError)
		}

		zf, err := w.Create(rel)
		if err != nil {
			return model.NewAppError("file.archive_failed", http.StatusInternalServerError)
		}

		fr, err := a.FileStorageObjects[f.Storage].ReadFile(ctx, fPath)
		if err != nil {
			tlog.Errorw("Failed to read file for public zip", "file_id", f.ID, "error", err)
			return model.NewAppError("file.archive_failed", http.StatusInternalServerError)
		}

		_, copyErr := io.Copy(zf, fr)
		fr.Close()
		if copyErr != nil {
			return model.NewAppError("file.archive_failed", http.StatusInternalServerError)
		}
	}

	return nil
}

// publicZipPath builds a file's path relative to rootID by walking parent links.
// Returns "" if a parent is missing or deleted (excluding the file from the zip).
func publicZipPath(byID map[string]model.File, rootID string, f model.File) string {
	parts := []string{f.Name}
	cur := f
	for i := 0; cur.Parent != rootID; i++ {
		if i > 64 {
			return ""
		}

		parent, ok := byID[cur.Parent]
		if !ok || parent.DeletedAt != 0 {
			return ""
		}

		parts = append([]string{parent.Name}, parts...)
		cur = parent
	}

	return path.Join(parts...)
}

func (a *App) PublicThumbnail(ctx context.Context, token, childID, accessToken string) (io.ReadCloser, *model.AppError) {
	link, target, appErr := a.publicTarget(ctx, token, childID, accessToken)
	if appErr != nil {
		return nil, appErr
	}

	if !link.AllowView {
		return nil, model.NewAppError("link.view_forbidden", http.StatusForbidden)
	}

	if target.IsFolder {
		return nil, model.NewAppError("file.not_found", http.StatusNotFound)
	}

	thumbPath, err := a.BuildFilePath(target.Storage, "thumbnails/"+target.ID+".jpg", model.AppFiles)
	if err != nil {
		return nil, model.NewAppError("file.not_found", http.StatusNotFound)
	}

	fr, err := a.FileStorageObjects[target.Storage].ReadFile(ctx, thumbPath)
	if err != nil {
		return nil, model.NewAppError("file.not_found", http.StatusNotFound)
	}

	return fr, nil
}

func (a *App) publicSubtreeFile(folder model.File, childID string) (*model.File, *model.AppError) {
	subs, err := a.Store.File.GetSubFilesByFolder(folder.ID, folder.Storage)
	if err != nil {
		tlog.Errorw("Failed to resolve public subtree file", "file_id", folder.ID, "child_id", childID, "error", err)
		return nil, model.NewAppError("link.retrieval_failed", http.StatusInternalServerError)
	}

	byID := make(map[string]model.File, len(subs))
	for i := range subs {
		byID[subs[i].ID] = subs[i]
	}

	target, ok := byID[childID]
	if !ok || target.DeletedAt != 0 {
		return nil, model.NewAppError("link.not_found", http.StatusNotFound)
	}

	// Reject unless every folder up to the shared root is live,
	// not just the file itself. Backstops a partial cascade or future regression;
	// trash/move already reject while a descendant is locked.
	steps := 0
	for cur := target; cur.Parent != folder.ID; {
		steps++
		if steps > len(subs) {
			return nil, model.NewAppError("link.not_found", http.StatusNotFound)
		}

		parent, ok := byID[cur.Parent]
		if !ok || parent.DeletedAt != 0 {
			return nil, model.NewAppError("link.not_found", http.StatusNotFound)
		}

		cur = parent
	}

	return &target, nil
}

// PublicUpload accepts a single multipart file into a folder shared with upload
// permission. The file is owned by (and counts against the storage limit of) the share
// owner, never the anonymous uploader. Takes *http.Request to read the multipart
// body, matching the existing authenticated upload path.
func (a *App) PublicUpload(shareToken, childID, accessToken string, r *http.Request) (*model.File, *model.AppError) {
	ctx := r.Context()

	_, folder, appErr := a.resolveUploadShare(ctx, shareToken, childID, accessToken)
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
		name = header.Filename
	}

	name = path.Base(name)
	if name == "" || name == "." {
		return nil, model.NewAppError("upload.missing_name", http.StatusBadRequest)
	}

	if header.Size <= 0 {
		return nil, model.NewAppError("upload.invalid_size", http.StatusBadRequest)
	}

	if header.Size > a.maxPublicUploadBytes() {
		return nil, model.NewAppError("upload.too_large", http.StatusRequestEntityTooLarge)
	}

	owner, appErr := a.shareOwnerWithSpace(folder, header.Size)
	if appErr != nil {
		return nil, appErr
	}

	mime := r.FormValue("type")
	if mime == "" {
		mime = "application/octet-stream"
	}

	newID := model.NewID()
	newFilePath, err := a.BuildFilePath(folder.Storage, newID, model.AppFiles)
	if err != nil {
		tlog.Errorw("Failed to build public upload path", "file_id", newID, "error", err)
		return nil, model.NewAppError("file.create_failed", http.StatusInternalServerError)
	}

	if err := a.FileStorageObjects[folder.Storage].WriteFile(ctx, newFilePath, file, header.Size); err != nil {
		tlog.Errorw("Failed to write public upload to storage", "file_id", newID, "error", err)
		return nil, model.NewAppError("file.create_failed", http.StatusInternalServerError)
	}

	newFile := a.buildPublicUploadFile(newID, name, mime, header.Size, *folder)
	if appErr := a.finalizePublicUpload(ctx, newFile, *folder, *owner); appErr != nil {
		return nil, appErr
	}

	return &newFile, nil
}

func (a *App) buildPublicUploadFile(id, name, mime string, size int64, folder model.File) model.File {
	return model.File{
		ID:          id,
		Owner:       folder.Owner,
		Parent:      folder.ID,
		Storage:     folder.Storage,
		Size:        size,
		Name:        name,
		DisplayName: name,
		Type:        mime,
	}
}

func (a *App) finalizePublicUpload(ctx context.Context, newFile, folder model.File, owner model.User) *model.AppError {
	if _, err := a.Store.File.Create(newFile); err != nil {
		tlog.Errorw("Failed to add public upload file record", "file_id", newFile.ID, "error", err)
		return model.NewAppError("file.create_failed", http.StatusInternalServerError)
	}

	if err := a.createImageThumbnail(ctx, newFile); err != nil {
		tlog.Warnw("Failed to create thumbnail for public upload", "file_id", newFile.ID)
	}

	a.RecordActivity(owner.ID, model.AppFiles, model.ActivityFilePublicUpload, folder.ID, newFile.ID, map[string]any{
		"name": newFile.DisplayName,
		"type": newFile.Type,
		"size": strconv.FormatInt(newFile.Size, 10),
	})

	return nil
}

func (a *App) StartPublicUploadSession(shareToken, childID, accessToken string, r *http.Request) (*model.UploadSession, *model.AppError) {
	ctx := r.Context()

	_, folder, appErr := a.resolveUploadShare(ctx, shareToken, childID, accessToken)
	if appErr != nil {
		return nil, appErr
	}

	name := path.Base(r.FormValue("name"))
	if name == "" || name == "." {
		return nil, model.NewAppError("upload.missing_name", http.StatusBadRequest)
	}

	var totalSize int64
	if v := r.FormValue("full_size"); v != "" {
		if parsed, err := strconv.ParseInt(v, 10, 64); err == nil {
			totalSize = parsed
		}
	}

	if totalSize <= 0 {
		return nil, model.NewAppError("upload.invalid_size", http.StatusBadRequest)
	}

	if totalSize > a.maxPublicUploadBytes() {
		return nil, model.NewAppError("upload.too_large", http.StatusRequestEntityTooLarge)
	}

	if _, appErr := a.shareOwnerWithSpace(folder, totalSize); appErr != nil {
		return nil, appErr
	}

	mime := r.FormValue("type")
	if mime == "" {
		mime = "application/octet-stream"
	}

	now := time.Now().Unix()
	session := model.UploadSession{
		ID:          model.NewID(),
		UserID:      folder.Owner,
		Storage:     folder.Storage,
		ContextType: "files",
		ContextID:   folder.ID,
		FileName:    name,
		MimeType:    mime,
		TotalSize:   totalSize,
		Status:      model.UploadStatusPending,
		CreatedAt:   now,
		UpdatedAt:   now,
		ExpiresAt:   time.Now().Add(24 * time.Hour).Unix(),
	}

	sessionPath, err := a.BuildFilePath(session.Storage, session.ID, model.AppFiles)
	if err != nil {
		tlog.Errorw("Failed to build public upload session path", "session_id", session.ID, "error", err)
		return nil, model.NewAppError("upload.start_failed", http.StatusInternalServerError)
	}

	uploadID, err := a.FileStorageObjects[session.Storage].StartMultipartUpload(ctx, sessionPath)
	if err != nil {
		tlog.Errorw("Failed to start public multipart upload", "session_id", session.ID, "error", err)
		return nil, model.NewAppError("upload.start_failed", http.StatusInternalServerError)
	}

	session.UploadID = uploadID

	created, err := a.Store.UploadSession.Create(session)
	if err != nil {
		tlog.Errorw("Failed to create public upload session", "session_id", session.ID, "error", err)
		return nil, model.NewAppError("upload.start_failed", http.StatusInternalServerError)
	}

	return created, nil
}

// publicUploadSession loads a session and binds it to the share's folder, so a
// session token cannot be retargeted to a different share or folder.
func (a *App) publicUploadSession(ctx context.Context, shareToken, childID, accessToken, sessionID string) (*model.UploadSession, *model.File, *model.AppError) {
	_, folder, appErr := a.resolveUploadShare(ctx, shareToken, childID, accessToken)
	if appErr != nil {
		return nil, nil, appErr
	}

	if sessionID == "" {
		return nil, nil, model.NewAppError("upload.missing_destination", http.StatusBadRequest)
	}

	session, err := a.Store.UploadSession.Get(sessionID)
	if err != nil {
		tlog.Errorw("Failed to retrieve public upload session", "session_id", sessionID, "error", err)
		return nil, nil, model.NewAppError("upload.not_found", http.StatusInternalServerError)
	}

	if session == nil {
		return nil, nil, model.NewAppError("upload.not_found", http.StatusNotFound)
	}

	if session.ContextID != folder.ID || session.UserID != folder.Owner {
		return nil, nil, model.NewAppError("upload.forbidden", http.StatusForbidden)
	}

	return session, folder, nil
}

func (a *App) PublicUploadChunk(shareToken, childID, accessToken string, r *http.Request) *model.AppError {
	ctx := r.Context()

	session, _, appErr := a.publicUploadSession(ctx, shareToken, childID, accessToken, r.FormValue("upload_session_id"))
	if appErr != nil {
		return appErr
	}

	if time.Unix(session.ExpiresAt, 0).Before(time.Now()) {
		go a.cleanupExpiredSession(session)
		return model.NewAppError("upload.expired", http.StatusGone)
	}

	partNumber, err := strconv.Atoi(r.FormValue("part_number"))
	if err != nil || partNumber < 1 {
		return model.NewAppError("upload.invalid_size", http.StatusBadRequest)
	}

	if session.Status != model.UploadStatusPending && session.Status != model.UploadStatusUploading {
		return model.NewAppError("upload.conflict", http.StatusConflict)
	}

	if session.UploadID == "" {
		return model.NewAppError("upload.start_failed", http.StatusBadRequest)
	}

	existingParts, err := session.GetParts()
	if err != nil {
		existingParts = make(map[int]string)
	}

	if _, exists := existingParts[partNumber]; exists {
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

	sessionPath, err := a.BuildFilePath(session.Storage, session.ID, session.ContextType)
	if err != nil {
		tlog.Errorw("Failed to build public chunk path", "session_id", session.ID, "error", err)
		return model.NewAppError("upload.chunk_failed", http.StatusInternalServerError)
	}

	etag, err := a.FileStorageObjects[session.Storage].UploadPart(ctx, sessionPath, session.UploadID, partNumber, file, header.Size)
	if err != nil {
		tlog.Errorw("Failed to upload public chunk", "session_id", session.ID, "part_number", partNumber, "error", err)
		return model.NewAppError("upload.chunk_failed", http.StatusInternalServerError)
	}

	existingParts[partNumber] = etag
	jsonBytes, err := json.Marshal(existingParts)
	if err != nil {
		return model.NewAppError("upload.chunk_failed", http.StatusInternalServerError)
	}

	now := time.Now().Unix()
	if err := a.Store.UploadSession.UpdateParts(session.ID, string(jsonBytes), session.UploadedSize+header.Size, len(existingParts), now); err != nil {
		tlog.Errorw("Failed to update public upload session parts", "session_id", session.ID, "error", err)
		return model.NewAppError("upload.chunk_failed", http.StatusInternalServerError)
	}

	if session.Status == model.UploadStatusPending {
		if err := a.Store.UploadSession.UpdateStatus(session.ID, model.UploadStatusUploading, now); err != nil {
			tlog.Errorw("Failed to update public upload session status", "session_id", session.ID, "error", err)
		}
	}

	return nil
}

func (a *App) FinishPublicUpload(shareToken, childID, accessToken string, r *http.Request) (*model.File, *model.AppError) {
	ctx := r.Context()

	session, folder, appErr := a.publicUploadSession(ctx, shareToken, childID, accessToken, r.FormValue("upload_session_id"))
	if appErr != nil {
		return nil, appErr
	}

	if session.UploadedSize != session.TotalSize {
		return nil, model.NewAppError("upload.incomplete", http.StatusConflict)
	}

	parts, err := session.GetParts()
	if err != nil || len(parts) == 0 {
		return nil, model.NewAppError("upload.incomplete", http.StatusBadRequest)
	}

	owner, appErr := a.shareOwnerWithSpace(folder, session.TotalSize)
	if appErr != nil {
		return nil, appErr
	}

	completePath, err := a.BuildFilePath(session.Storage, session.ID, session.ContextType)
	if err != nil {
		tlog.Errorw("Failed to build public completion path", "session_id", session.ID, "error", err)
		return nil, model.NewAppError("upload.finalize_failed", http.StatusInternalServerError)
	}

	if err := a.FileStorageObjects[session.Storage].CompleteMultipartUpload(ctx, completePath, session.UploadID, parts); err != nil {
		tlog.Errorw("Failed to complete public multipart upload", "session_id", session.ID, "error", err)
		return nil, model.NewAppError("upload.finalize_failed", http.StatusInternalServerError)
	}

	newFile := a.buildPublicUploadFile(session.ID, session.FileName, session.MimeType, session.TotalSize, *folder)
	if appErr := a.finalizePublicUpload(ctx, newFile, *folder, *owner); appErr != nil {
		return nil, appErr
	}

	return &newFile, nil
}

func (a *App) AbortPublicUploadSession(shareToken, childID, accessToken, sessionID string) *model.AppError {
	session, _, appErr := a.publicUploadSession(context.Background(), shareToken, childID, accessToken, sessionID)
	if appErr != nil {
		return appErr
	}

	a.cleanupExpiredSession(session)
	return nil
}
