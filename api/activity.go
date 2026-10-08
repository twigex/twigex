// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package api

import (
	"net/http"

	"github.com/gorilla/mux"
)

func (a *API) initActivity() {
	a.BaseRoutes.Activity.HandleFunc("/file/{id}", a.getActivity).Methods("GET")
	a.BaseRoutes.Activity.Use(a.RequireSession)
	a.BaseRoutes.Activity.Use(a.RequireCSRF)
}

func (a *API) getActivity(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	activity, appErr := a.app.GetActivity(*user, params["id"])
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, activity)
}
