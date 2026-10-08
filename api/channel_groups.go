// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/twigex/twigex/model"
)

func (a *API) listChannelGroups(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	groups, appErr := a.app.ListChannelGroups(r.Context(), *user, mux.Vars(r)["id"])
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, groups)
}

func (a *API) addChannelGroups(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	var req model.AddChannelGroupsRequest
	if !decodeBody(w, r, &req) {
		return
	}

	added, appErr := a.app.AddGroupsToChannel(r.Context(), *user, mux.Vars(r)["id"], req)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, added)
}

func (a *API) removeChannelGroup(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	params := mux.Vars(r)
	if appErr := a.app.RemoveGroupFromChannel(r.Context(), *user, params["id"], params["group"]); appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	w.WriteHeader(http.StatusOK)
}
