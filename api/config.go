// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package api

import (
	"net/http"
)

func (a *API) initConfig() {
	a.BaseRoutes.Config.HandleFunc("/client", a.getClientConfig).Methods("GET")
}

func (a *API) getClientConfig(w http.ResponseWriter, r *http.Request) {

	config := a.app.GetClientConfig(false)

	_, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondJSON(w, http.StatusOK, config)
		return
	}

	config = a.app.GetClientConfig(true)

	respondJSON(w, http.StatusOK, config)
}
