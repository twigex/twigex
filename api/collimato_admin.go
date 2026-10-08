// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/twigex/twigex/model"
)

func (a *API) initCollimatoAdmin() {
	a.BaseRoutes.Collimato.HandleFunc("/manage/workspaces", a.listAllWorkspaces).Methods("GET")
	a.BaseRoutes.Collimato.HandleFunc("/manage/workspaces/{workspace}", a.getWorkspaceDetails).Methods("GET")
	a.BaseRoutes.Collimato.HandleFunc("/manage/workspaces/{workspace}", a.updateManagedWorkspace).Methods("PUT")
	a.BaseRoutes.Collimato.HandleFunc("/manage/workspaces/{workspace}", a.deleteManagedWorkspace).Methods("DELETE")
	a.BaseRoutes.Collimato.HandleFunc("/manage/workspaces/{workspace}/join", a.joinManagedWorkspace).Methods("POST")
	a.BaseRoutes.Collimato.HandleFunc("/manage/workspaces/{workspace}/users", a.addManagedWorkspaceUsers).Methods("POST")
	a.BaseRoutes.Collimato.HandleFunc("/manage/workspaces/{workspace}/users/{user}", a.removeManagedWorkspaceUser).Methods("DELETE")
	a.BaseRoutes.Collimato.HandleFunc("/manage/workspaces/{workspace}/users/{user}/roles", a.updateManagedWorkspaceUserRoles).Methods("POST")
	a.BaseRoutes.Collimato.HandleFunc("/manage/workspaces/{workspace}/groups", a.addManagedWorkspaceGroups).Methods("POST")
	a.BaseRoutes.Collimato.HandleFunc("/manage/workspaces/{workspace}/groups/{group}", a.removeManagedWorkspaceGroup).Methods("DELETE")
	a.BaseRoutes.Collimato.HandleFunc("/manage/workspaces/{workspace}/groups/{group}/roles", a.updateManagedWorkspaceGroupRoles).Methods("POST")
}

func (a *API) listAllWorkspaces(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	workspaces, appErr := a.app.ListAllWorkspaces(r.Context(), *user)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, workspaces)
}

func (a *API) getWorkspaceDetails(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	details, appErr := a.app.GetWorkspaceDetails(r.Context(), *user, mux.Vars(r)["workspace"])
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, details)
}

func (a *API) updateManagedWorkspace(w http.ResponseWriter, r *http.Request) {
	var patch model.CollimatoWorkspacePatch
	if !decodeBody(w, r, &patch) {
		return
	}

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	workspace, appErr := a.app.UpdateManagedWorkspace(*user, mux.Vars(r)["workspace"], patch)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, workspace)
}

func (a *API) deleteManagedWorkspace(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	if appErr := a.app.DeleteManagedWorkspace(*user, mux.Vars(r)["workspace"]); appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *API) joinManagedWorkspace(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	if appErr := a.app.JoinManagedWorkspace(*user, mux.Vars(r)["workspace"]); appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *API) addManagedWorkspaceUsers(w http.ResponseWriter, r *http.Request) {
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

	if appErr := a.app.AddManagedWorkspaceUsers(*user, mux.Vars(r)["workspace"], s.Users); appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *API) removeManagedWorkspaceUser(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	if appErr := a.app.RemoveManagedWorkspaceUser(*user, params["workspace"], params["user"]); appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *API) updateManagedWorkspaceUserRoles(w http.ResponseWriter, r *http.Request) {
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

	if appErr := a.app.UpdateManagedWorkspaceUserRoles(*user, params["workspace"], params["user"], s.Roles); appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *API) addManagedWorkspaceGroups(w http.ResponseWriter, r *http.Request) {
	var req model.AddWorkspaceGroupsRequest
	if !decodeBody(w, r, &req) {
		return
	}

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	added, appErr := a.app.AddManagedWorkspaceGroups(*user, mux.Vars(r)["workspace"], req.GroupIDs, req.Roles)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, added)
}

func (a *API) removeManagedWorkspaceGroup(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	if appErr := a.app.RemoveManagedWorkspaceGroup(*user, params["workspace"], params["group"]); appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *API) updateManagedWorkspaceGroupRoles(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	var req model.UpdateWorkspaceGroupRolesRequest
	if !decodeBody(w, r, &req) {
		return
	}

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	if appErr := a.app.UpdateManagedWorkspaceGroupRoles(*user, params["workspace"], params["group"], req.Roles); appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	w.WriteHeader(http.StatusOK)
}
