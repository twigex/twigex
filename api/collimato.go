// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/twigex/twigex/app"
	"github.com/twigex/twigex/model"
)

func (a *API) initCollimato() {
	a.BaseRoutes.Collimato.HandleFunc("/workspaces/{workspace}/connections", a.getConnections).Methods("GET")
	a.BaseRoutes.Collimato.HandleFunc("/workspaces/{workspace}/connections", a.addConnection).Methods("POST")
	a.BaseRoutes.Collimato.HandleFunc("/workspaces/{workspace}/connections/test", a.testConnection).Methods("POST")
	a.BaseRoutes.Collimato.HandleFunc("/workspaces/{workspace}/connections/{id}/test", a.testConnection).Methods("POST")
	a.BaseRoutes.Collimato.HandleFunc("/workspaces/{workspace}/tables", a.getTables).Methods("GET")
	a.BaseRoutes.Collimato.HandleFunc("/workspaces/{workspace}/connections/{id}", a.getConnectionById).Methods("GET")
	a.BaseRoutes.Collimato.HandleFunc("/workspaces/{workspace}/connections/{id}", a.deleteConnection).Methods("DELETE")
	a.BaseRoutes.Collimato.HandleFunc("/workspaces/{workspace}/connections/{id}", a.updateConnection).Methods("PUT")
	a.BaseRoutes.Collimato.HandleFunc("/workspaces/{workspace}/users", a.addWorkspaceUsers).Methods("POST")
	a.BaseRoutes.Collimato.HandleFunc("/workspaces/{workspace}/users", a.getWorkspaceUsers).Methods("GET")
	a.BaseRoutes.Collimato.HandleFunc("/workspaces/{workspace}/users/me", a.getCurrentWorkspaceUser).Methods("GET")
	a.BaseRoutes.Collimato.HandleFunc("/workspaces/{workspace}/users/{id}", a.deleteWorkspaceUser).Methods("DELETE")
	a.BaseRoutes.Collimato.HandleFunc("/workspaces/{workspace}/users/{id}/roles", a.updateUserRoles).Methods("POST")
	a.BaseRoutes.Collimato.HandleFunc("/workspaces/{workspace}/groups", a.addWorkspaceGroups).Methods("POST")
	a.BaseRoutes.Collimato.HandleFunc("/workspaces/{workspace}/groups", a.getWorkspaceGroups).Methods("GET")
	a.BaseRoutes.Collimato.HandleFunc("/workspaces/{workspace}/groups/{group}", a.removeWorkspaceGroup).Methods("DELETE")
	a.BaseRoutes.Collimato.HandleFunc("/workspaces/{workspace}/groups/{group}/roles", a.updateWorkspaceGroupRoles).Methods("POST")
	a.BaseRoutes.Collimato.HandleFunc("/workspaces/{workspace}/charts", a.getCharts).Methods("GET")
	a.BaseRoutes.Collimato.HandleFunc("/workspaces/{workspace}/charts", a.createChart).Methods("POST")
	a.BaseRoutes.Collimato.HandleFunc("/workspaces/{workspace}/charts/{id}", a.getChartByID).Methods("GET")
	a.BaseRoutes.Collimato.HandleFunc("/workspaces/{workspace}/charts/{id}", a.updateChart).Methods("PUT")
	a.BaseRoutes.Collimato.HandleFunc("/workspaces/{workspace}/charts/{id}", a.deleteChart).Methods("DELETE")
	a.BaseRoutes.Collimato.HandleFunc("/workspaces/{workspace}/dashboards", a.getDashboards).Methods("GET")
	a.BaseRoutes.Collimato.HandleFunc("/workspaces/{workspace}/dashboards", a.createDashboard).Methods("POST")
	a.BaseRoutes.Collimato.HandleFunc("/workspaces/{workspace}/dashboards/{id}", a.getDashboardById).Methods("GET")
	a.BaseRoutes.Collimato.HandleFunc("/workspaces/{workspace}/dashboards/{id}", a.updateDashboard).Methods("PUT")
	a.BaseRoutes.Collimato.HandleFunc("/workspaces/{workspace}/dashboards/{id}", a.deleteDashboard).Methods("DELETE")
	a.BaseRoutes.Collimato.HandleFunc("/workspaces/{workspace}/dashboards/{id}/charts", a.updateDashboardCharts).Methods("PUT")
	a.BaseRoutes.Collimato.HandleFunc("/workspaces/{workspace}/dashboards/{id}/filters/add", a.addDashboardFilter).Methods("POST")
	a.BaseRoutes.Collimato.HandleFunc("/workspaces/{workspace}/dashboards/{id}/filters/{filter}", a.updateDashboardFilter).Methods("PUT")
	a.BaseRoutes.Collimato.HandleFunc("/workspaces/{workspace}/dashboards/{id}/filters/{filter}", a.deleteDashboardFilter).Methods("DELETE")
	a.BaseRoutes.Collimato.HandleFunc("/workspaces/{workspace}/data", a.queryData).Methods("POST")
	a.BaseRoutes.Collimato.HandleFunc("/workspaces/{workspace}", a.deleteWorkspace).Methods("DELETE")
	a.BaseRoutes.Collimato.HandleFunc("/workspaces", a.getWorkspaces).Methods("GET")
	a.BaseRoutes.Collimato.HandleFunc("/workspaces", a.createWorkspace).Methods("POST")
	a.BaseRoutes.Collimato.HandleFunc("/workspaces/{workspace}", a.getWorkspace).Methods("GET")
	a.BaseRoutes.Collimato.HandleFunc("/workspaces/{workspace}/finish", a.finishWorkspaceCreation).Methods("POST")
	a.BaseRoutes.Collimato.HandleFunc("/workspaces/{workspace}/files", a.newCubeFile).Methods("POST")
	a.BaseRoutes.Collimato.HandleFunc("/workspaces/{workspace}/files", a.getCubeFiles).Methods("GET")
	a.BaseRoutes.Collimato.HandleFunc("/workspaces/{workspace}/files", a.saveCubeFile).Methods("PUT")
	a.BaseRoutes.Collimato.HandleFunc("/workspaces/{workspace}/files/delete", a.deleteCubeFile).Methods("POST")
	a.BaseRoutes.Collimato.HandleFunc("/workspaces/{workspace}/files/generate", a.generateCubeFiles).Methods("POST")
	a.BaseRoutes.Collimato.HandleFunc("/workspaces/{workspace}/files/build", a.buildCubeFile).Methods("POST")
	a.BaseRoutes.Collimato.HandleFunc("/workspaces/{workspace}/files/build-view", a.buildViewFile).Methods("POST")
	a.BaseRoutes.Collimato.HandleFunc("/workspaces/{workspace}/roles", a.getWorkspaceRoles).Methods("GET")
	a.BaseRoutes.Collimato.HandleFunc("/workspaces/{workspace}/roles/{id}", a.getWorkspaceRole).Methods("GET")
	a.BaseRoutes.Collimato.HandleFunc("/workspaces/{workspace}/roles/{id}/update", a.updateWorkspaceRole).Methods("POST")
	a.BaseRoutes.Collimato.HandleFunc("/workspaces/{workspace}/roles/{id}", a.deleteWorkspaceRole).Methods("DELETE")
	a.BaseRoutes.Collimato.HandleFunc("/workspaces/{workspace}/roles", a.createWorkspaceRole).Methods("POST")
	a.BaseRoutes.Collimato.HandleFunc("/workspaces/{workspace}/meta", a.getMeta).Methods("GET")
	a.BaseRoutes.Collimato.HandleFunc("/workspaces/{workspace}/meta/roles", a.getRoleMeta).Methods("GET")
	a.BaseRoutes.Collimato.HandleFunc("/workspaces/{workspace}/sql", a.getSql).Methods("POST")
	a.BaseRoutes.Collimato.HandleFunc("/status", a.collimatoStatus).Methods("GET")

	a.BaseRoutes.Collimato.Use(a.RequireSession)
	a.BaseRoutes.Collimato.Use(a.RequireCSRF)
}

func (a *API) addConnection(w http.ResponseWriter, r *http.Request) {
	workspaceID := mux.Vars(r)["workspace"]

	var s model.NewConnection
	if !decodeBody(w, r, &s) {
		return
	}

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	database, appErr := a.app.AddConnection(*user, workspaceID, s)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, database)
}

func (a *API) testConnection(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	var s model.NewConnection
	if !decodeBody(w, r, &s) {
		return
	}

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	if appErr = a.app.TestConnection(*user, vars["workspace"], vars["id"], s); appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (a *API) getConnections(w http.ResponseWriter, r *http.Request) {
	workspaceID := mux.Vars(r)["workspace"]

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	databases, appErr := a.app.GetConnections(*user, workspaceID)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, databases)
}

func (a *API) getConnectionById(w http.ResponseWriter, r *http.Request) {
	workspaceID := mux.Vars(r)["workspace"]
	connectionID := mux.Vars(r)["id"]

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	database, appErr := a.app.GetConnectionByID(*user, workspaceID, connectionID)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, database)
}

func (a *API) updateConnection(w http.ResponseWriter, r *http.Request) {
	workspaceID := mux.Vars(r)["workspace"]
	connectionID := mux.Vars(r)["id"]

	var s model.ConnectionPatch
	if !decodeBody(w, r, &s) {
		return
	}

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	if _, appErr = a.app.UpdateConnection(*user, workspaceID, connectionID, s); appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *API) deleteConnection(w http.ResponseWriter, r *http.Request) {
	workspaceID := mux.Vars(r)["workspace"]
	connectionID := mux.Vars(r)["id"]

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	appErr = a.app.DeleteConnection(*user, workspaceID, connectionID)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *API) addWorkspaceUsers(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	s := struct {
		Users []string `json:"users"`
	}{}

	if !decodeBody(w, r, &s) {
		return
	}

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	appErr = a.app.AddWorkspaceUsers(*user, params["workspace"], s.Users, []string{
		model.CollimatoWorkspaceUserRoleId,
	})
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *API) getWorkspaceUsers(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	users, appErr := a.app.GetWorkspaceUsers(*user, params["workspace"])
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, users)
}

func (a *API) deleteWorkspaceUser(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	appErr = a.app.RemoveWorkspaceUser(*user, params["workspace"], params["id"])
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *API) updateUserRoles(w http.ResponseWriter, r *http.Request) {
	workspaceID := mux.Vars(r)["workspace"]
	userID := mux.Vars(r)["id"]

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

	if appErr = a.app.UpdateUserRoles(*user, workspaceID, userID, s.Roles); appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *API) addWorkspaceGroups(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	var req model.AddWorkspaceGroupsRequest
	if !decodeBody(w, r, &req) {
		return
	}

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	added, appErr := a.app.AddGroupsToWorkspace(*user, params["workspace"], req.GroupIDs, req.Roles)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, added)
}

func (a *API) getWorkspaceGroups(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	groups, appErr := a.app.ListWorkspaceGroups(*user, params["workspace"])
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, groups)
}

func (a *API) removeWorkspaceGroup(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	if appErr := a.app.RemoveGroupFromWorkspace(*user, params["workspace"], params["group"]); appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *API) updateWorkspaceGroupRoles(w http.ResponseWriter, r *http.Request) {
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

	if appErr := a.app.UpdateWorkspaceGroupRoles(*user, params["workspace"], params["group"], req.Roles); appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *API) getCurrentWorkspaceUser(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	permissions, appErr := a.app.GetCurrentWorkspaceUser(*user, params["workspace"])
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	permissions.UserInfo = *user

	respondJSON(w, http.StatusOK, permissions)
}

func (a *API) createChart(w http.ResponseWriter, r *http.Request) {
	workspaceID := mux.Vars(r)["workspace"]

	var s model.NewChart
	if !decodeBody(w, r, &s) {
		return
	}

	if err := s.Validate(); err != nil {
		respondAppError(w, r, model.NewAppError("request.invalid", http.StatusBadRequest))
		return
	}

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	chart, appErr := a.app.CreateChart(*user, workspaceID, s)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, chart)
}

func (a *API) getCharts(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	charts, appErr := a.app.GetCharts(*user, params["workspace"])
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, charts)
}

func (a *API) getChartByID(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	chart, appErr := a.app.GetChartByID(*user, params["workspace"], params["id"])
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, chart)
}

func (a *API) updateChart(w http.ResponseWriter, r *http.Request) {
	workspaceID := mux.Vars(r)["workspace"]
	chartID := mux.Vars(r)["id"]

	var s model.ChartPatch
	if !decodeBody(w, r, &s) {
		return
	}

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	if _, appErr = a.app.UpdateChart(*user, workspaceID, chartID, s); appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *API) deleteChart(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	appErr = a.app.DeleteChart(*user, params["workspace"], params["id"])
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *API) getDashboards(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	dashboards, appErr := a.app.GetDashboards(*user, params["workspace"])
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, dashboards)
}

func (a *API) getDashboardById(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	dashboard, appErr := a.app.GetDashboardByID(*user, params["workspace"], params["id"])
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, dashboard)
}

func (a *API) updateDashboard(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	s := model.DashboardPatch{}
	if !decodeBody(w, r, &s) {
		return
	}

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	_, appErr = a.app.UpdateDashboard(*user, params["workspace"], params["id"], s)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *API) deleteDashboard(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	appErr = a.app.DeleteDashboard(*user, params["workspace"], params["id"])
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *API) updateDashboardCharts(w http.ResponseWriter, r *http.Request) {
	workspaceID := mux.Vars(r)["workspace"]
	dashboardID := mux.Vars(r)["id"]

	var s struct {
		Charts []model.DashboardChartPatch `json:"charts"`
	}
	if !decodeBody(w, r, &s) {
		return
	}

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	if appErr = a.app.UpdateDashboardCharts(*user, workspaceID, dashboardID, s.Charts); appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *API) createDashboard(w http.ResponseWriter, r *http.Request) {
	workspaceID := mux.Vars(r)["workspace"]

	var s model.Dashboard
	if !decodeBody(w, r, &s) {
		return
	}

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	dashboard, appErr := a.app.CreateDashboard(*user, workspaceID, s)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, dashboard)
}

func (a *API) addDashboardFilter(w http.ResponseWriter, r *http.Request) {
	workspaceID := mux.Vars(r)["workspace"]
	dashboardID := mux.Vars(r)["id"]

	var s model.DashboardFilter
	if !decodeBody(w, r, &s) {
		return
	}

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	filter, appErr := a.app.AddDashboardFilter(*user, workspaceID, dashboardID, s)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, filter)
}

func (a *API) updateDashboardFilter(w http.ResponseWriter, r *http.Request) {
	workspaceID := mux.Vars(r)["workspace"]
	dashboardID := mux.Vars(r)["id"]
	filterID := mux.Vars(r)["filter"]

	var s model.DashboardFilterPatch
	if !decodeBody(w, r, &s) {
		return
	}

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	filter, appErr := a.app.UpdateDashboardFilter(*user, workspaceID, dashboardID, filterID, s)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, filter)
}

func (a *API) deleteDashboardFilter(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	appErr = a.app.DeleteDashboardFilter(*user, params["workspace"], params["id"], params["filter"])
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *API) queryData(w http.ResponseWriter, r *http.Request) {
	workspaceID := mux.Vars(r)["workspace"]

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	var s model.DataQuery
	if !decodeBody(w, r, &s) {
		return
	}

	body, appErr := a.app.QueryData(r.Context(), *user, workspaceID, s)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, translateCubeError(r, body))
}

// translateCubeError replaces a driver error Cube passed through with wording a
// workspace admin can act on. Cube reports query errors inside a 200 body, and
// the original is already in the log via logCubeError.
func translateCubeError(r *http.Request, body map[string]any) map[string]any {
	raw, ok := body["error"].(string)
	if !ok {
		return body
	}

	id := app.CubeErrorID(raw)
	if id == "" {
		return body
	}

	body["error"] = model.NewAppError(id, http.StatusOK).Translate(getLocale(r))

	return body
}

func (a *API) getWorkspaces(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	workspaces, appErr := a.app.GetWorkspaces(r.Context(), *user)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, workspaces)
}

func (a *API) createWorkspace(w http.ResponseWriter, r *http.Request) {
	var s struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if !decodeBody(w, r, &s) {
		return
	}

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	workspace, appErr := a.app.CreateWorkspace(*user, s.Name, s.Description)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, workspace)
}

func (a *API) getWorkspace(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	workspace, appErr := a.app.GetWorkspaceByID(*user, params["workspace"])
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, workspace)
}

func (a *API) deleteWorkspace(w http.ResponseWriter, r *http.Request) {
	workspaceID := mux.Vars(r)["workspace"]

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	if appErr = a.app.DeleteWorkspace(*user, workspaceID); appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *API) finishWorkspaceCreation(w http.ResponseWriter, r *http.Request) {
	workspaceID := mux.Vars(r)["workspace"]

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	workspace, appErr := a.app.FinishWorkspaceCreation(*user, workspaceID, model.CollimatoWorkspaceStatusActive)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, workspace)
}

func (a *API) getMeta(w http.ResponseWriter, r *http.Request) {
	workspaceID := mux.Vars(r)["workspace"]

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	meta, appErr := a.app.GetWorkspaceMeta(r.Context(), *user, workspaceID)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, meta)
}

func (a *API) getRoleMeta(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	meta, appErr := a.app.GetWorkspaceRoleMeta(r.Context(), *user, mux.Vars(r)["workspace"])
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, meta)
}

func (a *API) getSql(w http.ResponseWriter, r *http.Request) {
	workspaceID := mux.Vars(r)["workspace"]

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	var s model.DataQuery
	if !decodeBody(w, r, &s) {
		return
	}

	body, appErr := a.app.GetWorkspaceSQL(r.Context(), *user, workspaceID, s)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, translateCubeError(r, body))
}

func (a *API) newCubeFile(w http.ResponseWriter, r *http.Request) {
	workspaceID := mux.Vars(r)["workspace"]

	var s struct {
		Filetype string
		Filename string
	}
	if !decodeBody(w, r, &s) {
		return
	}

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	newFile, appErr := a.app.NewCubeFile(*user, workspaceID, s.Filetype, s.Filename)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, newFile)
}

func (a *API) getCubeFiles(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	cubeFiles, appErr := a.app.GetCubeFiles(*user, params["workspace"])
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, cubeFiles)
}

func (a *API) buildCubeFile(w http.ResponseWriter, r *http.Request) {
	workspaceID := mux.Vars(r)["workspace"]

	var s struct {
		Type      string
		Model     model.Dataset
		Overwrite bool
	}
	if !decodeBody(w, r, &s) {
		return
	}

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	file, appErr := a.app.BuildCubeFile(*user, workspaceID, s.Type, s.Model, s.Overwrite)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, file)
}

func (a *API) buildViewFile(w http.ResponseWriter, r *http.Request) {
	workspaceID := mux.Vars(r)["workspace"]

	var s struct {
		Model     model.View
		Overwrite bool
	}
	if !decodeBody(w, r, &s) {
		return
	}

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	file, appErr := a.app.BuildViewFile(*user, workspaceID, s.Model, s.Overwrite)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, file)
}

func (a *API) saveCubeFile(w http.ResponseWriter, r *http.Request) {
	workspaceID := mux.Vars(r)["workspace"]

	var s struct {
		Name    string
		Content string
		Type    string
	}
	if !decodeBody(w, r, &s) {
		return
	}

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	newFile, appErr := a.app.SaveCubeFile(*user, s.Name, s.Content, s.Type, workspaceID)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, newFile)
}

func (a *API) deleteCubeFile(w http.ResponseWriter, r *http.Request) {
	workspaceID := mux.Vars(r)["workspace"]

	var s struct {
		Name     string
		Filetype string
	}
	if !decodeBody(w, r, &s) {
		return
	}

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	if appErr = a.app.DeleteCubeFile(*user, s.Name, s.Filetype, workspaceID); appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *API) generateCubeFiles(w http.ResponseWriter, r *http.Request) {
	workspaceID := mux.Vars(r)["workspace"]

	var m map[string]any
	if !decodeBody(w, r, &m) {
		return
	}

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	result, appErr := a.app.GenerateCubeFiles(*user, m, workspaceID)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, result)
}

func (a *API) getWorkspaceRoles(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	roles, appErr := a.app.GetWorkspaceRoles(*user, params["workspace"])
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, roles)
}

func (a *API) getWorkspaceRole(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	role, appErr := a.app.GetWorkspaceRole(*user, params["workspace"], params["id"])
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, role)
}

func (a *API) createWorkspaceRole(w http.ResponseWriter, r *http.Request) {
	workspaceID := mux.Vars(r)["workspace"]

	var s model.CollimatoRole
	if !decodeBody(w, r, &s) {
		return
	}

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	role, appErr := a.app.CreateWorkspaceRole(*user, workspaceID, s)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, role)
}

func (a *API) updateWorkspaceRole(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	s := model.CollimatoRolePatch{}
	if !decodeBody(w, r, &s) {
		return
	}

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	role, appErr := a.app.UpdateWorkspaceRole(*user, params["workspace"], params["id"], s)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, role)
}

func (a *API) deleteWorkspaceRole(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	appErr = a.app.DeleteWorkspaceRole(*user, params["workspace"], params["id"])
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *API) collimatoStatus(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, a.app.CollimatoConfigured())
}

func (a *API) getTables(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	tables, appErr := a.app.GetTables(*user, params["workspace"])
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, tables)
}
