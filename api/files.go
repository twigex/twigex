// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package api

import (
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/twigex/twigex/model"
)

func (a *API) initFiles() {
	a.BaseRoutes.Files.HandleFunc("/share", a.share).Methods("POST")
	a.BaseRoutes.Files.HandleFunc("/share/update", a.updateSharePermissions).Methods("POST")
	a.BaseRoutes.Files.HandleFunc("/move", a.moveFile).Methods("POST")
	a.BaseRoutes.Files.HandleFunc("/{id}/unshare", a.unshare).Methods("POST")
	a.BaseRoutes.Files.HandleFunc("/{id}/unshare-group", a.unshareGroup).Methods("POST")
	a.BaseRoutes.Files.HandleFunc("/upload", a.uploadFile).Methods("POST")
	a.BaseRoutes.Files.HandleFunc("/delete", a.deleteFile).Methods("POST")
	a.BaseRoutes.Files.HandleFunc("/deleted", a.getDeleted).Methods("GET")
	a.BaseRoutes.Files.HandleFunc("/restore", a.restoreFile).Methods("POST")
	a.BaseRoutes.Files.HandleFunc("/trash", a.trashFile).Methods("POST")
	a.BaseRoutes.Files.HandleFunc("/rename/{id}", a.renameFile).Methods("POST")
	a.BaseRoutes.Files.HandleFunc("/recents", a.getRecents).Methods("GET")
	a.BaseRoutes.Files.HandleFunc("/favorites/{id}", a.addFavorite).Methods("POST")
	a.BaseRoutes.Files.HandleFunc("/favorites", a.getFavorites).Methods("GET")
	a.BaseRoutes.Files.HandleFunc("/get-drive", a.getDrive).Methods("GET")
	a.BaseRoutes.Files.HandleFunc("/shared", a.getShared).Methods("GET")
	a.BaseRoutes.Files.HandleFunc("/parent/{id}", a.getParent).Methods("GET")
	a.BaseRoutes.Files.HandleFunc("/folder/{id}", a.loadFiles).Methods("GET")
	a.BaseRoutes.Files.HandleFunc("/download/{id}", a.downloadFile).Methods("GET")
	a.BaseRoutes.Files.HandleFunc("/view/{id}", a.viewFile).Methods("GET")
	a.BaseRoutes.Files.HandleFunc("/create-folder", a.createFolder).Methods("POST")
	a.BaseRoutes.Files.HandleFunc("/thumbnails/{id}", a.loadThumbnail).Methods("GET")
	a.BaseRoutes.Files.HandleFunc("/create-document", a.createDocument).Methods("POST")
	a.BaseRoutes.Files.HandleFunc("/details/{id}", a.getDetails).Methods("GET")
	a.BaseRoutes.Files.HandleFunc("/metadata/entry/add", a.addMetadataEntryToFile).Methods("POST")     // adds metadata to file in file_metadata_entries
	a.BaseRoutes.Files.HandleFunc("/metadata/entry/remove", a.removeFileMetadataEntry).Methods("POST") //
	a.BaseRoutes.Files.HandleFunc("/metadata/entry/{id}", a.updateFileMetadataEntry).Methods("PUT")    // updates file metadata entry
	a.BaseRoutes.Files.HandleFunc("/file/{id}", a.getFileInfo).Methods("GET")

	a.BaseRoutes.Files.Use(a.RequireSession)
	a.BaseRoutes.Files.Use(a.RequireCSRF)
}

func (a *API) getFileInfo(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	file, appErr := a.app.HasPermission(id, *user)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, file)
}

func (a *API) addMetadataEntryToFile(w http.ResponseWriter, r *http.Request) {
	var m struct {
		ID     string
		Metaid string
		Value  string
	}
	if !decodeBody(w, r, &m) {
		return
	}

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	meta, appErr := a.app.CreateMetadataEntry(*user, m.ID, m.Metaid, m.Value)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, meta)
}

func (a *API) removeFileMetadataEntry(w http.ResponseWriter, r *http.Request) {
	var m struct {
		FileID string
		ID     string
	}
	if !decodeBody(w, r, &m) {
		return
	}

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	if appErr = a.app.DeleteMetadataEntry(m.FileID, m.ID, *user); appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *API) updateFileMetadataEntry(w http.ResponseWriter, r *http.Request) {
	entryID := mux.Vars(r)["id"]

	var m struct {
		Value any
	}
	if !decodeBody(w, r, &m) {
		return
	}

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	if appErr = a.app.UpdateMetadataEntry(*user, entryID, m.Value); appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// Loads files in the parent directory
func (a *API) loadFiles(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	if !a.requirePermission(w, r, *user, model.FilePermissions.PermissionViewFiles) {
		return
	}

	files, appErr := a.app.GetUserFiles(*user, params["id"])
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, files)
}

func (a *API) getDrive(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	if !a.requirePermission(w, r, *user, model.FilePermissions.PermissionViewFiles) {
		return
	}

	drives, appErr := a.app.GetUserDrives(*user)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, drives)
}

func (a *API) getShared(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	if !a.requirePermission(w, r, *user, model.FilePermissions.PermissionViewFiles) {
		return
	}

	sharedFiles, appErr := a.app.GetSharedFiles(user.ID)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, sharedFiles)
}

func (a *API) getRecents(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	if !a.requirePermission(w, r, *user, model.FilePermissions.PermissionViewFiles) {
		return
	}

	recents, appErr := a.app.GetRecents(*user)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, recents)
}

func (a *API) getParent(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	if !a.requirePermission(w, r, *user, model.FilePermissions.PermissionViewFiles) {
		return
	}

	files, appErr := a.app.GetParent(*user, params["id"])
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, files)
}

func (a *API) createDocument(w http.ResponseWriter, r *http.Request) {
	d := &model.FileCrateRequest{}
	if !decodeBody(w, r, &d) {
		return
	}

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	if !a.requirePermission(w, r, *user, model.FilePermissions.PermissionCreateFiles) {
		return
	}

	newFile, appErr := a.app.CreateDocument(*user, *d)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, newFile)
}

func (a *API) createFolder(w http.ResponseWriter, r *http.Request) {
	d := struct {
		Name string
		Id   string
	}{}

	if !decodeBody(w, r, &d) {
		return
	}

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	if !a.requirePermission(w, r, *user, model.FilePermissions.PermissionCreateFiles) {
		return
	}

	f, appErr := a.app.CreateFolder(*user, d.Name, d.Id)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, f)
}

func (a *API) moveFile(w http.ResponseWriter, r *http.Request) {
	move := struct {
		ID   []string
		Path string
	}{}

	if !decodeBody(w, r, &move) {
		return
	}

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	job, appErr := a.app.MoveFile(r.Context(), *user, move.ID, move.Path)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, job)
}

func (a *API) deleteFile(w http.ResponseWriter, r *http.Request) {
	d := struct {
		ID []string
	}{}

	if !decodeBody(w, r, &d) {
		return
	}

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	appErr = a.app.DeleteFile(*user, d.ID)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *API) getDeleted(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	if !a.requirePermission(w, r, *user, model.FilePermissions.PermissionViewFiles) {
		return
	}

	files, appErr := a.app.GetDeleted(*user)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, files)
}

func (a *API) restoreFile(w http.ResponseWriter, r *http.Request) {
	d := struct {
		ID []string
	}{}

	if !decodeBody(w, r, &d) {
		return
	}

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	appErr = a.app.Restore(*user, d.ID)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *API) trashFile(w http.ResponseWriter, r *http.Request) {
	d := struct {
		ID []string
	}{}

	if !decodeBody(w, r, &d) {
		return
	}

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	appErr = a.app.Trash(*user, d.ID)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *API) uploadFile(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	if !a.requirePermission(w, r, *user, model.FilePermissions.PermissionUploadFiles) {
		return
	}

	switch r.URL.Query().Get("action") {
	case "start":
		session, appErr := a.app.StartUploadSession(*user, r)
		if appErr != nil {
			respondAppError(w, r, appErr)
			return
		}

		respondJSON(w, http.StatusOK, session)

	case "abort":
		sessionID := r.URL.Query().Get("session_id")
		if sessionID == "" {
			respondAppError(w, r, model.NewAppError("request.invalid", http.StatusBadRequest))
			return
		}

		if appErr = a.app.AbortUploadSession(*user, sessionID); appErr != nil {
			respondAppError(w, r, appErr)
			return
		}

		w.WriteHeader(http.StatusOK)

	default:
		newFile, appErr := a.app.UploadFile(*user, r)
		if appErr != nil {
			respondAppError(w, r, appErr)
			return
		}

		if newFile != nil {
			respondJSON(w, http.StatusOK, newFile)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func (a *API) renameFile(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	f := struct {
		Name string
	}{}
	if !decodeBody(w, r, &f) {
		return
	}

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	appErr = a.app.Rename(*user, params["id"], f.Name)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *API) addFavorite(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	if !a.requirePermission(w, r, *user, model.FilePermissions.PermissionViewFiles) {
		return
	}

	appErr = a.app.AddFavorite(*user, params["id"])
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *API) getFavorites(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	if !a.requirePermission(w, r, *user, model.FilePermissions.PermissionViewFiles) {
		return
	}

	f, appErr := a.app.GetFavourites(user.ID)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, f)
}

func (a *API) getDetails(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	if !a.requirePermission(w, r, *user, model.FilePermissions.PermissionViewFiles) {
		return
	}

	details, appErr := a.app.GetDetails(r.Context(), *user, params["id"])
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, details)
}

func (a *API) loadThumbnail(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	if !a.requirePermission(w, r, *user, model.FilePermissions.PermissionViewFiles) {
		return
	}

	file, appErr := a.app.LoadThumbnail(*user, params["id"])
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	defer file.Close()

	w.Header().Set("Content-Type", "image/jpeg")
	if _, err := io.Copy(w, file); err != nil {
		return
	}
}
