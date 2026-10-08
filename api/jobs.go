// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package api

import (
	"net/http"

	"github.com/gorilla/mux"
)

func (a *API) initJobs() {
	a.BaseRoutes.Jobs.HandleFunc("/running", a.getRunningJobs).Methods("GET")
	a.BaseRoutes.Jobs.HandleFunc("/cancel/{id}", a.cancelJob).Methods("POST")
	a.BaseRoutes.Jobs.HandleFunc("/acknowledge/{id}", a.acknowledgeJob).Methods("POST")

	a.BaseRoutes.Jobs.Use(a.RequireSession)
	a.BaseRoutes.Jobs.Use(a.RequireCSRF)
}

func (a *API) getRunningJobs(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		http.Error(w, appErr.Message, http.StatusUnauthorized)
		return
	}

	jobs, appErr := a.app.GetUserRunningJobs(*user)
	if appErr != nil {
		http.Error(w, appErr.Message, http.StatusBadRequest)
		return
	}

	respondJSON(w, http.StatusOK, jobs)
}

func (a *API) cancelJob(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	jobID := params["id"]

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		http.Error(w, appErr.Message, http.StatusUnauthorized)
		return
	}

	appErr = a.app.CancelUserJob(*user, jobID)
	if appErr != nil {
		http.Error(w, appErr.Message, http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *API) acknowledgeJob(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	jobID := params["id"]

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		http.Error(w, appErr.Message, http.StatusUnauthorized)
		return
	}

	appErr = a.app.AcknowledgeUserJob(*user, jobID)
	if appErr != nil {
		http.Error(w, appErr.Message, http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}
