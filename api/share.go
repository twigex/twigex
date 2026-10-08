// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/twigex/twigex/model"
)

func (a *API) share(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	req := model.FileShare{}
	if !decodeBody(w, r, &req) {
		return
	}

	if len(req.Users) == 0 && len(req.Groups) == 0 {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	js, appErr := a.app.Share(*user, req)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, js)
}

func (a *API) unshare(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	req := struct {
		User string
	}{}
	if !decodeBody(w, r, &req) {
		return
	}

	appErr = a.app.Unshare(*user, params["id"], req.User)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *API) unshareGroup(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	req := struct {
		Group string
	}{}
	if !decodeBody(w, r, &req) {
		return
	}

	if appErr := a.app.UnshareGroup(r.Context(), *user, params["id"], req.Group); appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *API) updateSharePermissions(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	s := model.FileSharePatch{}
	if !decodeBody(w, r, &s) {
		return
	}

	appErr = a.app.UpdateSharePermissions(r.Context(), *user, s)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	w.WriteHeader(http.StatusOK)
}
