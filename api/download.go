// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package api

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/tlog"
)

func (a *API) viewFile(w http.ResponseWriter, r *http.Request) {
	fileID := mux.Vars(r)["id"]

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	if !a.requirePermission(w, r, *user, model.FilePermissions.PermissionDownloadFiles) {
		return
	}

	file, appErr := a.app.GetFileToDownload(*user, fileID)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	if file.IsFolder {
		respondAppError(w, r, model.NewAppError("file.unknown_type", http.StatusBadRequest))
		return
	}

	setUploadHeaders(w, file.Type, file.Name)
	w.Header().Set("Accept-Ranges", "bytes")

	filePath, err := a.app.BuildFilePath(file.Storage, file.ID, model.AppFiles)
	if err != nil {
		tlog.Errorw("Failed to build file path",
			"file_id", fileID,
			"error", err,
		)
		return
	}

	if err := a.app.FileStorageObjects[file.Storage].ServeFile(filePath, w, r); err != nil {
		tlog.Errorw("Failed to serve file",
			"file_id", fileID,
			"error", err,
		)
		return
	}

	go a.app.RecordActivity(user.ID, model.AppFiles, model.ActivityFileView, file.Parent, file.ID, map[string]any{
		"name": file.DisplayName,
		"type": file.Type,
		"size": strconv.FormatInt(file.Size, 10),
	})
}

func (a *API) downloadFile(w http.ResponseWriter, r *http.Request) {
	fileID := mux.Vars(r)["id"]

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	if !a.requirePermission(w, r, *user, model.FilePermissions.PermissionDownloadFiles) {
		return
	}

	file, appErr := a.app.GetFileToDownload(*user, fileID)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	if file.IsFolder {
		name := file.Name + ".zip"
		safe := sanitizeHeaderFilename(name)
		w.Header().Set("Content-Type", "application/zip")
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"; filename*=UTF-8''%s`, safe, url.PathEscape(name)))
		w.Header().Set("X-Content-Type-Options", "nosniff")

		if _, appErr = a.app.CreateZip(*user, file.ID, w); appErr != nil {
			tlog.Errorw("Failed to create zip",
				"file_id", fileID,
				"error", appErr.Message,
			)
		}
	} else {
		safe := sanitizeHeaderFilename(file.Name)
		ctype := file.Type
		if ctype == "" {
			ctype = "application/octet-stream"
		}

		w.Header().Set("Content-Type", ctype)
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"; filename*=UTF-8''%s`, safe, url.PathEscape(file.Name)))
		w.Header().Set("X-Content-Type-Options", "nosniff")

		filePath, err := a.app.BuildFilePath(file.Storage, file.ID, model.AppFiles)
		if err != nil {
			tlog.Errorw("Failed to build file path",
				"file_id", fileID,
				"error", err,
			)
			return
		}

		if err := a.app.FileStorageObjects[file.Storage].ServeFile(filePath, w, r); err != nil {
			tlog.Errorw("Failed to serve file",
				"file_id", fileID,
				"error", err,
			)
			return
		}
	}

	go a.app.CreateFileDownloadNotification(*file, *user, model.NOTIFICATION_FILE_DOWNLOAD, nil)
	go a.app.RecordActivity(user.ID, model.AppFiles, model.ActivityFileDownload, file.Parent, file.ID, map[string]any{
		"name": file.DisplayName,
		"type": file.Type,
		"size": strconv.FormatInt(file.Size, 10),
	})
}

func sanitizeHeaderFilename(name string) string {
	name = strings.ReplaceAll(name, "\r", "")
	name = strings.ReplaceAll(name, "\n", "")
	name = strings.ReplaceAll(name, `"`, "")
	if name == "" {
		name = "file"
	}

	return name
}
