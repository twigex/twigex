// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/twigex/twigex/model"
)

func (a *API) initGraph() {
	a.BaseRoutes.Graph.HandleFunc("/rebuild", a.getGraphRebuildStatus).Methods("GET")
	a.BaseRoutes.Graph.HandleFunc("/rebuild", a.rebuildGraph).Methods("POST")
	a.BaseRoutes.Graph.HandleFunc("/entities/{type}/{id}/relations", a.getGraphEntityRelations).Methods("GET")
	a.BaseRoutes.Graph.HandleFunc("/types", a.getGraphTypes).Methods("GET")
	a.BaseRoutes.Graph.HandleFunc("/entity-types", a.createGraphEntityType).Methods("POST")
	a.BaseRoutes.Graph.HandleFunc("/entity-types/{id}", a.updateGraphEntityType).Methods("PUT")
	a.BaseRoutes.Graph.HandleFunc("/entity-types/{id}", a.deleteGraphEntityType).Methods("DELETE")
	a.BaseRoutes.Graph.HandleFunc("/relation-types", a.createGraphRelationType).Methods("POST")
	a.BaseRoutes.Graph.HandleFunc("/relation-types/{id}", a.updateGraphRelationType).Methods("PUT")
	a.BaseRoutes.Graph.HandleFunc("/relation-types/{id}", a.deleteGraphRelationType).Methods("DELETE")

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

func (a *API) getGraphEntityRelations(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	vars := mux.Vars(r)

	relations, appErr := a.app.GetGraphEntityRelations(r.Context(), *user, vars["type"], vars["id"])
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, relations)
}

func (a *API) getGraphTypes(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	types, appErr := a.app.GetGraphTypes(r.Context(), *user)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, types)
}

func (a *API) createGraphEntityType(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	var input model.GraphEntityTypeInput
	if !decodeBody(w, r, &input) {
		return
	}

	entityType, appErr := a.app.CreateGraphEntityType(r.Context(), *user, input)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, entityType)
}

func (a *API) updateGraphEntityType(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	var input model.GraphEntityTypeInput
	if !decodeBody(w, r, &input) {
		return
	}

	entityType, appErr := a.app.UpdateGraphEntityType(r.Context(), *user, mux.Vars(r)["id"], input)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, entityType)
}

func (a *API) deleteGraphEntityType(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	appErr = a.app.DeleteGraphEntityType(r.Context(), *user, mux.Vars(r)["id"])
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *API) createGraphRelationType(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	var input model.GraphRelationTypeInput
	if !decodeBody(w, r, &input) {
		return
	}

	relationType, appErr := a.app.CreateGraphRelationType(r.Context(), *user, input)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, relationType)
}

func (a *API) updateGraphRelationType(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	var input model.GraphRelationTypeInput
	if !decodeBody(w, r, &input) {
		return
	}

	relationType, appErr := a.app.UpdateGraphRelationType(r.Context(), *user, mux.Vars(r)["id"], input)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, relationType)
}

func (a *API) deleteGraphRelationType(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	appErr = a.app.DeleteGraphRelationType(r.Context(), *user, mux.Vars(r)["id"])
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	w.WriteHeader(http.StatusOK)
}
