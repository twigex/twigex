// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/tlog"
)

func (a *API) initOffice() {
	a.BaseRoutes.Office.HandleFunc("/{app}/file/open/{id}", a.officeOpenFile).Methods("POST")
	a.BaseRoutes.Office.HandleFunc("/file/support", a.officeSupport).Methods("POST")
	a.BaseRoutes.Office.Use(a.RequireSession)
	a.BaseRoutes.Office.Use(a.RequireCSRF)

	Office := a.BaseRoutes.APIRoot.PathPrefix("/office").Subrouter()
	Office.HandleFunc("/wopi/{app}/files/{id}/contents", a.officeGetFile).Methods("GET")
	Office.HandleFunc("/wopi/{app}/files/{id}/contents", a.officeSaveFile).Methods("POST")
	Office.HandleFunc("/wopi/{app}/files/{id}", a.checkFileInfo).Methods("GET")
	Office.HandleFunc("/eurooffice/file/{app}/{id}", a.officeGetFile).Methods("GET")
	Office.HandleFunc("/eurooffice/callback/{app}/{id}", a.euroOfficeCallback).Methods("POST")
}

func (a *API) officeSupport(w http.ResponseWriter, r *http.Request) {
	s := struct {
		OfficeType string `json:"officeType"`
	}{}
	if !decodeBody(w, r, &s) {
		return
	}

	ok, appErr := a.app.OfficeSupportsFileType(s.OfficeType)
	if appErr != nil {
		respondJSON(w, http.StatusOK, false)
		return
	}

	if !*a.app.ConfigStore.Config.OfficeSettings.Enable {
		ok = false
	}

	respondJSON(w, http.StatusOK, ok)
}

func (a *API) officeOpenFile(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	id := params["id"]
	app := params["app"]

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		http.Error(w, appErr.Message, appErr.Status)
		return
	}

	result, appErr := a.app.OpenOfficeFile(id, *user, app)
	if appErr != nil {
		http.Error(w, appErr.Message, appErr.Status)
		return
	}

	respondJSON(w, http.StatusOK, result)
}

func (a *API) officeGetFile(w http.ResponseWriter, r *http.Request) {
	tokenString := r.URL.Query().Get("access_token")
	params := mux.Vars(r)
	id := params["id"]

	switch params["app"] {
	case "files":
		file, appErr := a.app.GetOfficeFile(id, tokenString)
		if appErr != nil {
			respondAppError(w, r, appErr)
			return
		}

		filePath, err := a.app.BuildFilePath(file.Storage, file.ID, model.AppFiles)
		if err != nil {
			tlog.Errorw("Failed to build office file path",
				"file_id", file.ID,
				"error", err,
			)
			respondAppError(w, r, model.NewAppError("office.open_failed", http.StatusInternalServerError))
			return
		}

		if err := a.app.FileStorageObjects[file.Storage].ServeFile(filePath, w, r); err != nil {
			tlog.Errorw("Failed to serve office file",
				"file_id", file.ID,
				"error", err,
			)
			respondAppError(w, r, model.NewAppError("office.open_failed", http.StatusInternalServerError))
		}

	case "chat":
		appErr := a.app.ServeOfficeChatAttachment(id, w)
		if appErr != nil {
			respondAppError(w, r, appErr)
		}

	case "projects":
		appErr := a.app.ServeOfficeWorkspaceAttachment(id, w)
		if appErr != nil {
			respondAppError(w, r, appErr)
		}

	default:
		respondAppError(w, r, model.NewAppError("office.open_failed", http.StatusNotFound))
	}
}

func (a *API) checkFileInfo(w http.ResponseWriter, r *http.Request) {
	tokenString := r.URL.Query().Get("access_token")

	params := mux.Vars(r)
	id := params["id"]
	app := params["app"]

	bytes, appErr := a.app.CheckFileInfo(id, tokenString, app)
	if appErr != nil {
		http.Error(w, appErr.Message, appErr.Status)
		return
	}

	if _, err := w.Write(*bytes); err != nil {
		tlog.Errorw("Failed to write response", "error", err)
	}
}

func (a *API) officeSaveFile(w http.ResponseWriter, r *http.Request) {
	tokenString := r.URL.Query().Get("access_token")

	params := mux.Vars(r)
	id := params["id"]

	appErr := a.app.OfficeSaveFile(id, tokenString, r.Body)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *API) euroOfficeCallback(w http.ResponseWriter, r *http.Request) {
	tokenString := r.URL.Query().Get("access_token")

	params := mux.Vars(r)
	id := params["id"]
	app := params["app"]

	appErr := a.app.EuroOfficeCallback(id, app, tokenString, r.Body)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write([]byte(`{"error":0}`)); err != nil {
		tlog.Errorw("Failed to write response", "error", err)
	}
}
