// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/twigex/twigex/model"
)

func (a *API) initProjectAdmin() {
	a.BaseRoutes.Workspaces.HandleFunc("/manage", a.listAllProjectWorkspaces).Methods("GET")
	a.BaseRoutes.Workspaces.HandleFunc("/manage/{workspace}", a.getManagedProjectWorkspace).Methods("GET")
	a.BaseRoutes.Workspaces.HandleFunc("/manage/{workspace}", a.updateManagedProjectWorkspace).Methods("PUT")
	a.BaseRoutes.Workspaces.HandleFunc("/manage/{workspace}", a.deleteManagedProjectWorkspace).Methods("DELETE")
	a.BaseRoutes.Workspaces.HandleFunc("/manage/{workspace}/join", a.joinManagedProjectWorkspace).Methods("POST")
	a.BaseRoutes.Workspaces.HandleFunc("/manage/{workspace}/members", a.addManagedProjectWorkspaceMembers).Methods("POST")
	a.BaseRoutes.Workspaces.HandleFunc("/manage/{workspace}/members/{user}", a.removeManagedProjectWorkspaceMember).Methods("DELETE")
	a.BaseRoutes.Workspaces.HandleFunc("/manage/{workspace}/members/{user}/role", a.updateManagedProjectWorkspaceMemberRole).Methods("POST")
	a.BaseRoutes.Workspaces.HandleFunc("/manage/{workspace}/groups", a.addManagedProjectWorkspaceGroups).Methods("POST")
	a.BaseRoutes.Workspaces.HandleFunc("/manage/{workspace}/groups/{group}", a.removeManagedProjectWorkspaceGroup).Methods("DELETE")
	a.BaseRoutes.Workspaces.HandleFunc("/manage/{workspace}/groups/{group}/roles", a.updateManagedProjectWorkspaceGroupRoles).Methods("POST")
}

func (a *API) listAllProjectWorkspaces(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	workspaces, appErr := a.app.ListAllProjectWorkspaces(r.Context(), *user)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, workspaces)
}

func (a *API) getManagedProjectWorkspace(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	details, appErr := a.app.GetManagedProjectWorkspace(r.Context(), *user, mux.Vars(r)["workspace"])
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, details)
}

func (a *API) updateManagedProjectWorkspace(w http.ResponseWriter, r *http.Request) {
	var patch model.ProjectWorkspacePatch
	if !decodeBody(w, r, &patch) {
		return
	}

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	if appErr := a.app.UpdateManagedProjectWorkspace(*user, mux.Vars(r)["workspace"], patch); appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *API) deleteManagedProjectWorkspace(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	if appErr := a.app.DeleteManagedProjectWorkspace(r.Context(), *user, mux.Vars(r)["workspace"]); appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *API) joinManagedProjectWorkspace(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	if appErr := a.app.JoinManagedProjectWorkspace(*user, mux.Vars(r)["workspace"]); appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *API) addManagedProjectWorkspaceMembers(w http.ResponseWriter, r *http.Request) {
	var s struct {
		Users []string `json:"users"`
	}
	if !decodeBody(w, r, &s) {
		return
	}

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	if appErr := a.app.AddManagedProjectWorkspaceMembers(*user, mux.Vars(r)["workspace"], s.Users); appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *API) removeManagedProjectWorkspaceMember(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	if appErr := a.app.RemoveManagedProjectWorkspaceMember(*user, params["workspace"], params["user"]); appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *API) updateManagedProjectWorkspaceMemberRole(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	var s struct {
		Role string `json:"role"`
	}
	if !decodeBody(w, r, &s) {
		return
	}

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	if appErr := a.app.UpdateManagedProjectWorkspaceMemberRole(*user, params["workspace"], params["user"], s.Role); appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *API) addManagedProjectWorkspaceGroups(w http.ResponseWriter, r *http.Request) {
	var req model.AddProjectGroupsRequest
	if !decodeBody(w, r, &req) {
		return
	}

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	added, appErr := a.app.AddManagedProjectWorkspaceGroups(*user, mux.Vars(r)["workspace"], req.GroupIDs, req.Roles)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, added)
}

func (a *API) removeManagedProjectWorkspaceGroup(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	if appErr := a.app.RemoveManagedProjectWorkspaceGroup(*user, params["workspace"], params["group"]); appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *API) updateManagedProjectWorkspaceGroupRoles(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	var s struct {
		Roles []string `json:"roles"`
	}
	if !decodeBody(w, r, &s) {
		return
	}

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	if appErr := a.app.UpdateManagedProjectWorkspaceGroupRoles(*user, params["workspace"], params["group"], s.Roles); appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	w.WriteHeader(http.StatusOK)
}
