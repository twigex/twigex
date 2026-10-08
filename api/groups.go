// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package api

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/twigex/twigex/model"
)

func (a *API) initGroups() {
	a.BaseRoutes.Groups.HandleFunc("", a.createGroup).Methods("POST")
	a.BaseRoutes.Groups.HandleFunc("/page", a.listGroupsPaged).Methods("GET")
	a.BaseRoutes.Groups.HandleFunc("/search", a.searchGroups).Methods("GET")
	a.BaseRoutes.Groups.HandleFunc("/{id}", a.getGroup).Methods("GET")
	a.BaseRoutes.Groups.HandleFunc("/{id}", a.updateGroup).Methods("PUT")
	a.BaseRoutes.Groups.HandleFunc("/{id}", a.deleteGroup).Methods("DELETE")
	a.BaseRoutes.Groups.HandleFunc("/{id}/members", a.listGroupMembers).Methods("GET")
	a.BaseRoutes.Groups.HandleFunc("/{id}/members", a.addGroupMembers).Methods("POST")
	a.BaseRoutes.Groups.HandleFunc("/{id}/members/{user_id}", a.removeGroupMember).Methods("DELETE")

	a.BaseRoutes.Groups.Use(a.RequireSession)
	a.BaseRoutes.Groups.Use(a.RequireCSRF)
}

func (a *API) listGroupsPaged(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	query := r.URL.Query().Get("q")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	groups, total, appErr := a.app.ListGroupsPaged(r.Context(), *user, query, sortFromQuery(r), limit, offset)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, model.GroupsPage{Items: groups, Total: total})
}

func (a *API) searchGroups(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	groups, appErr := a.app.SearchGroups(r.Context(), *user, r.URL.Query().Get("q"), limit)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, groups)
}

func (a *API) createGroup(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	var req model.CreateGroupRequest
	if !decodeBody(w, r, &req) {
		return
	}

	group, appErr := a.app.CreateGroup(r.Context(), *user, req)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, group)
}

func (a *API) getGroup(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	group, appErr := a.app.GetGroup(r.Context(), *user, mux.Vars(r)["id"])
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, group)
}

func (a *API) updateGroup(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	var req model.UpdateGroupRequest
	if !decodeBody(w, r, &req) {
		return
	}

	group, appErr := a.app.UpdateGroup(r.Context(), *user, mux.Vars(r)["id"], req)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, group)
}

func (a *API) deleteGroup(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	if appErr := a.app.DeleteGroup(r.Context(), *user, mux.Vars(r)["id"]); appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *API) listGroupMembers(w http.ResponseWriter, r *http.Request) {
	if _, appErr := a.app.GetCurrentUser(r); appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	query := r.URL.Query().Get("q")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	members, total, appErr := a.app.GetGroupMembers(r.Context(), mux.Vars(r)["id"], query, sortFromQuery(r), limit, offset)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, model.GroupMembersPage{Items: members, Total: total})
}

func (a *API) addGroupMembers(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	var req model.AddGroupMembersRequest
	if !decodeBody(w, r, &req) {
		return
	}

	if appErr := a.app.AddGroupMembers(r.Context(), *user, mux.Vars(r)["id"], req); appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *API) removeGroupMember(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	params := mux.Vars(r)
	if appErr := a.app.RemoveGroupMember(r.Context(), *user, params["id"], params["user_id"]); appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	w.WriteHeader(http.StatusOK)
}
