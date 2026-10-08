// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package api

import (
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/tlog"
)

const shareAccessCookiePrefix = "share_access_"

func (a *API) initPublic() {
	a.BaseRoutes.PublicLink.HandleFunc("/share", a.linkShare).Methods("POST")
	a.BaseRoutes.PublicLink.HandleFunc("/update", a.updateLink).Methods("POST")
	a.BaseRoutes.PublicLink.HandleFunc("/delete", a.deleteLink).Methods("POST")

	a.BaseRoutes.PublicLink.Use(a.RequireSession)
	a.BaseRoutes.PublicLink.Use(a.RequireCSRF)
}

func (a *API) linkShare(w http.ResponseWriter, r *http.Request) {
	var req model.Request
	if !decodeBody(w, r, &req) {
		return
	}

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	link, appErr := a.app.CreateLink(r.Context(), *user, req)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, link)
}

func (a *API) updateLink(w http.ResponseWriter, r *http.Request) {
	var l model.LinkUpdate
	if !decodeBody(w, r, &l) {
		return
	}

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	if appErr := a.app.UpdateLink(r.Context(), *user, l); appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *API) deleteLink(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token string `json:"token"`
	}
	if !decodeBody(w, r, &req) {
		return
	}

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	if appErr := a.app.DeleteLink(r.Context(), *user, req.Token); appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *API) initPublicShare() {
	a.BaseRoutes.PublicSharedFiles.HandleFunc("/{token}", a.getPublicShare).Methods("GET")
	a.BaseRoutes.PublicSharedFiles.HandleFunc("/{token}/folder/{childID}", a.getPublicShare).Methods("GET")
	a.BaseRoutes.PublicSharedFiles.HandleFunc("/{token}/auth", a.authPublicShare).Methods("POST")
	a.BaseRoutes.PublicSharedFiles.HandleFunc("/{token}/download", a.downloadPublicFile).Methods("GET")
	a.BaseRoutes.PublicSharedFiles.HandleFunc("/{token}/download/{childID}", a.downloadPublicFile).Methods("GET")
	a.BaseRoutes.PublicSharedFiles.HandleFunc("/{token}/thumbnail/{childID}", a.publicThumbnail).Methods("GET")
	a.BaseRoutes.PublicSharedFiles.HandleFunc("/{token}/view", a.viewPublicFile).Methods("GET")
	a.BaseRoutes.PublicSharedFiles.HandleFunc("/{token}/view/{childID}", a.viewPublicFile).Methods("GET")
	a.BaseRoutes.PublicSharedFiles.HandleFunc("/{token}/office", a.openPublicOffice).Methods("GET")
	a.BaseRoutes.PublicSharedFiles.HandleFunc("/{token}/office/{childID}", a.openPublicOffice).Methods("GET")
	a.BaseRoutes.PublicSharedFiles.HandleFunc("/{token}/upload", a.publicUpload).Methods("POST")

	a.BaseRoutes.PublicSharedFiles.Use(a.publicShareHeaders)
}

func (a *API) publicShareHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Referrer-Policy", "no-referrer")
		next.ServeHTTP(w, r)
	})
}

func (a *API) getPublicShare(w http.ResponseWriter, r *http.Request) {
	token := mux.Vars(r)["token"]
	childID := mux.Vars(r)["childID"]

	file, appErr := a.app.GetPublicShare(r.Context(), token, childID, publicAccessToken(r, token))
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, file)
}

func (a *API) viewPublicFile(w http.ResponseWriter, r *http.Request) {
	token := mux.Vars(r)["token"]
	childID := mux.Vars(r)["childID"]

	file, appErr := a.app.PublicViewFile(r.Context(), token, childID, publicAccessToken(r, token))
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	setUploadHeaders(w, file.Type, file.DisplayName)

	filePath, err := a.app.BuildFilePath(file.Storage, file.ID, model.AppFiles)
	if err != nil {
		tlog.Errorw("Failed to build public view path", "file_id", file.ID, "error", err)
		return
	}

	if err := a.app.FileStorageObjects[file.Storage].ServeFile(filePath, w, r); err != nil {
		tlog.Errorw("Failed to serve public view file", "file_id", file.ID, "error", err)
		return
	}
}

func (a *API) publicThumbnail(w http.ResponseWriter, r *http.Request) {
	token := mux.Vars(r)["token"]
	childID := mux.Vars(r)["childID"]

	fr, appErr := a.app.PublicThumbnail(r.Context(), token, childID, publicAccessToken(r, token))
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	defer fr.Close()

	w.Header().Set("Content-Type", "image/jpeg")
	if _, err := io.Copy(w, fr); err != nil {
		tlog.Warnw("Failed to stream public thumbnail", "share_token", token)
	}
}

func (a *API) authPublicShare(w http.ResponseWriter, r *http.Request) {
	token := mux.Vars(r)["token"]

	var body struct {
		Password string `json:"password"`
	}
	if !decodeBody(w, r, &body) {
		return
	}

	accessToken, appErr := a.app.AuthenticatePublicShare(r.Context(), token, body.Password, a.app.ClientIP(r))
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	if accessToken != "" {
		http.SetCookie(w, &http.Cookie{
			Name:     shareAccessCookiePrefix + token,
			Value:    accessToken,
			Path:     "/api/public",
			MaxAge:   86400,
			HttpOnly: true,
			Secure:   r.TLS != nil,
			SameSite: http.SameSiteLaxMode,
		})
	}

	w.WriteHeader(http.StatusOK)
}

func (a *API) downloadPublicFile(w http.ResponseWriter, r *http.Request) {
	token := mux.Vars(r)["token"]
	childID := mux.Vars(r)["childID"]

	file, appErr := a.app.PublicDownloadTarget(r.Context(), token, childID, publicAccessToken(r, token))
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	if file.IsFolder {
		name := file.DisplayName + ".zip"
		safe := sanitizeHeaderFilename(name)
		w.Header().Set("Content-Type", "application/zip")
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"; filename*=UTF-8''%s`, safe, url.PathEscape(name)))
		w.Header().Set("X-Content-Type-Options", "nosniff")

		if appErr := a.app.ZipPublicFolderContents(r.Context(), *file, w); appErr != nil {
			tlog.Errorw("Failed to zip public folder",
				"file_id", file.ID,
				"error", appErr.Message,
			)
		}

		return
	}

	safe := sanitizeHeaderFilename(file.DisplayName)
	ctype := file.Type
	if ctype == "" {
		ctype = "application/octet-stream"
	}

	w.Header().Set("Content-Type", ctype)
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"; filename*=UTF-8''%s`, safe, url.PathEscape(file.DisplayName)))
	w.Header().Set("X-Content-Type-Options", "nosniff")

	filePath, err := a.app.BuildFilePath(file.Storage, file.ID, model.AppFiles)
	if err != nil {
		tlog.Errorw("Failed to build public file path",
			"file_id", file.ID,
			"error", err,
		)
		return
	}

	if err := a.app.FileStorageObjects[file.Storage].ServeFile(filePath, w, r); err != nil {
		tlog.Errorw("Failed to serve public file",
			"file_id", file.ID,
			"error", err,
		)
		return
	}
}

func (a *API) openPublicOffice(w http.ResponseWriter, r *http.Request) {
	token := mux.Vars(r)["token"]
	childID := mux.Vars(r)["childID"]
	nickname := r.URL.Query().Get("name")
	if len(nickname) > 64 {
		nickname = nickname[:64]
	}

	result, appErr := a.app.OpenPublicOfficeFile(r.Context(), token, childID, publicAccessToken(r, token), nickname)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, result)
}

func (a *API) publicUpload(w http.ResponseWriter, r *http.Request) {
	token := mux.Vars(r)["token"]
	accessToken := publicAccessToken(r, token)
	folder := r.URL.Query().Get("folder")

	switch r.URL.Query().Get("action") {
	case "start":
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			respondAppError(w, r, model.NewAppError("request.invalid", http.StatusBadRequest))
			return
		}

		session, appErr := a.app.StartPublicUploadSession(token, folder, accessToken, r)
		if appErr != nil {
			respondAppError(w, r, appErr)
			return
		}

		respondJSON(w, http.StatusOK, session)

	case "chunk":
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			respondAppError(w, r, model.NewAppError("upload.too_large", http.StatusRequestEntityTooLarge))
			return
		}

		if appErr := a.app.PublicUploadChunk(token, folder, accessToken, r); appErr != nil {
			respondAppError(w, r, appErr)
			return
		}

		w.WriteHeader(http.StatusOK)

	case "finish":
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			respondAppError(w, r, model.NewAppError("request.invalid", http.StatusBadRequest))
			return
		}

		newFile, appErr := a.app.FinishPublicUpload(token, folder, accessToken, r)
		if appErr != nil {
			respondAppError(w, r, appErr)
			return
		}

		respondJSON(w, http.StatusOK, struct {
			Name string `json:"name"`
		}{Name: newFile.DisplayName})

	case "abort":
		if appErr := a.app.AbortPublicUploadSession(token, folder, accessToken, r.URL.Query().Get("session_id")); appErr != nil {
			respondAppError(w, r, appErr)
			return
		}

		w.WriteHeader(http.StatusOK)

	default:
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			respondAppError(w, r, model.NewAppError("upload.too_large", http.StatusRequestEntityTooLarge))
			return
		}

		newFile, appErr := a.app.PublicUpload(token, folder, accessToken, r)
		if appErr != nil {
			respondAppError(w, r, appErr)
			return
		}

		respondJSON(w, http.StatusOK, struct {
			Name string `json:"name"`
		}{Name: newFile.DisplayName})
	}
}

func publicAccessToken(r *http.Request, token string) string {
	c, err := r.Cookie(shareAccessCookiePrefix + token)
	if err != nil {
		return ""
	}

	return c.Value
}
