// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package api

import (
	"net/http"
)

func (a *API) initGraph() {
	a.BaseRoutes.Graph.HandleFunc("/rebuild", a.getGraphRebuildStatus).Methods("GET")
	a.BaseRoutes.Graph.HandleFunc("/rebuild", a.rebuildGraph).Methods("POST")

	a.BaseRoutes.Graph.Use(a.RequireSession)
	a.BaseRoutes.Graph.Use(a.RequireCSRF)
}

func (a *API) getGraphRebuildStatus(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	status, appErr := a.app.GetGraphRebuildStatus(*user)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, status)
}

func (a *API) rebuildGraph(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	job, appErr := a.app.RequestGraphRebuild(*user)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusAccepted, job)
}
