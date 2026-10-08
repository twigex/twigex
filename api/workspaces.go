// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package api

import (
	"encoding/json"
	"fmt"

	// "log"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/gorilla/mux"
	"github.com/twigex/twigex/internal/parse"
	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/tlog"
)

func (a *API) initWorkspaces() {
	a.BaseRoutes.Workspaces.HandleFunc("/create", a.createProjectWorkspace)
	a.BaseRoutes.Workspaces.HandleFunc("/get", a.getProjectWorkspaces)
	a.BaseRoutes.Workspaces.HandleFunc("/get/{id}", a.getProjectWorkspaceDetails)
	a.BaseRoutes.Workspaces.HandleFunc("/{id}/table/create", a.createWorkspaceTable)
	a.BaseRoutes.Workspaces.HandleFunc("/{id}/tables/get/{data}", a.getWorkspaceTables)
	a.BaseRoutes.Workspaces.HandleFunc("/{id}/table/{tid}/create", a.createWorkspaceTableField)
	a.BaseRoutes.Workspaces.HandleFunc("/{id}/table/{tid}/task/create", a.createWorkspaceTask)
	a.BaseRoutes.Workspaces.HandleFunc("/{id}/table/{tid}/task/{taskid}/update", a.updateWorkspaceTask)
	a.BaseRoutes.Workspaces.HandleFunc("/{id}/table/{tid}/single-select-name/{taskid}/update", a.updateWorkspaceSingleSelectName)
	a.BaseRoutes.Workspaces.HandleFunc("/{id}/table/{tid}/task/{taskid}/update-link", a.updateWorkspaceTaskLink)
	a.BaseRoutes.Workspaces.HandleFunc("/{id}/member/add", a.addMemberToWorkspace)
	a.BaseRoutes.Workspaces.HandleFunc("/{id}/members/search", a.searchWorkspaceMembers).Methods("GET")
	a.BaseRoutes.Workspaces.HandleFunc("/{id}/groups", a.addProjectWorkspaceGroups).Methods("POST")
	a.BaseRoutes.Workspaces.HandleFunc("/{id}/groups", a.getProjectWorkspaceGroups).Methods("GET")
	a.BaseRoutes.Workspaces.HandleFunc("/{id}/groups/{group_id}", a.removeProjectWorkspaceGroup).Methods("DELETE")
	a.BaseRoutes.Workspaces.HandleFunc("/{id}/groups/{group_id}/roles", a.updateProjectWorkspaceGroupRoles).Methods("POST")
	a.BaseRoutes.Workspaces.HandleFunc("/{id}/table/{tid}/view/create", a.createNewView)
	a.BaseRoutes.Workspaces.HandleFunc("/{id}/table/{tid}/view/{view_id}/update", a.updateView)
	a.BaseRoutes.Workspaces.HandleFunc("/{id}/table/{tid}/view/{view_id}/kanban/move", a.moveKanbanCard).Methods("POST")
	a.BaseRoutes.Workspaces.HandleFunc("/{id}/table/{tid}/field-value/add", a.addFieldValue)
	a.BaseRoutes.Workspaces.HandleFunc("/{id}/table/{tid}/field-update", a.updateWorkspaceTableField)
	a.BaseRoutes.Workspaces.HandleFunc("/{id}/table/{tid}/field-edit-formula", a.editFieldFormula)
	a.BaseRoutes.Workspaces.HandleFunc("/{id}/table/{tid}/kanban-data", a.getKanbanData).Methods("POST")
	a.BaseRoutes.Workspaces.HandleFunc("/{id}/table/{tid}/tasks/date-range", a.getTasksByDateRange).Methods("POST")
	a.BaseRoutes.Workspaces.HandleFunc("/{id}/table/{tid}/tasks/calendar-range", a.getTasksByCalendarRange).Methods("POST")
	a.BaseRoutes.Workspaces.HandleFunc("/{id}/table/{tid}/rows-lite", a.getLinkedRecordsLite).Methods("GET")
	a.BaseRoutes.Workspaces.HandleFunc("/{id}/table/{tid}/view/{fid}/display-name-update", a.updateDisplayName)
	a.BaseRoutes.Workspaces.HandleFunc("/{id}/table/{tid}/task/{taskid}/assigne", a.addMemberToTask)
	a.BaseRoutes.Workspaces.HandleFunc("/{id}/delete", a.deleteProjectWorkspace)
	a.BaseRoutes.Workspaces.HandleFunc("/{id}/member/{member_id}/delete", a.deleteWorkspaceMember)
	a.BaseRoutes.Workspaces.HandleFunc("/{id}/table/{tid}/delete", a.deleteTable)
	a.BaseRoutes.Workspaces.HandleFunc("/{id}/table/{tid}/task/{taskid}/delete", a.deleteWorkspaceTask)
	a.BaseRoutes.Workspaces.HandleFunc("/{id}/table/{tid}/view/{viewid}/delete", a.deleteWorkspaceTableView)
	a.BaseRoutes.Workspaces.HandleFunc("/{id}/table/{tid}/field/{fieldid}/delete", a.deleteWorkspaceTableField)
	a.BaseRoutes.Workspaces.HandleFunc("/{id}/table/{tid}/single/{fieldid}/delete", a.deleteWorkspaceTableSingleField)
	a.BaseRoutes.Workspaces.HandleFunc("/assigned-to-me", a.assignedToMe)
	a.BaseRoutes.Workspaces.HandleFunc("/all-workspace-tasks", a.getAllWorkspaceTasks)
	a.BaseRoutes.Workspaces.HandleFunc("/all-workspace-tasks/export", a.exportAllWorkspaceTasks).Methods("GET")
	a.BaseRoutes.Workspaces.HandleFunc("/all-workspace-tasks/members", a.searchTaskReportMembers).Methods("GET")
	a.BaseRoutes.Workspaces.HandleFunc("/{id}/table/{tid}/task/{taskid}/page", a.getTaskPage).Methods("POST")
	a.BaseRoutes.Workspaces.HandleFunc("/{id}/table/{tid}/task/{taskid}/matches", a.taskShownByFilter).Methods("POST")
	a.BaseRoutes.Workspaces.HandleFunc("/{id}/table/{tid}/status-types", a.getTableStatusTypes).Methods("GET")
	a.BaseRoutes.Workspaces.HandleFunc("/{id}/table/{tid}/task/{taskid}/add-comment", a.addTaskComment)
	a.BaseRoutes.Workspaces.HandleFunc("/{id}/table/{tid}/task/{taskid}/get-comments", a.getTaskComments)
	a.BaseRoutes.Workspaces.HandleFunc("/{id}/folder/create", a.addWorkspaceFolder)
	a.BaseRoutes.Workspaces.HandleFunc("/{id}/table/{tid}/item/{itemid}", a.getItemForTableByID)
	a.BaseRoutes.Workspaces.HandleFunc("/{id}/table/{tid}/item/{itemid}/subtasks", a.getSubtasks).Methods("GET")
	a.BaseRoutes.Workspaces.HandleFunc("/{id}/table/{tid}/item/{itemid}/completion", a.getTaskCompletion).Methods("GET")
	a.BaseRoutes.Workspaces.HandleFunc("/{id}/table/{tid}/item/{itemid}/complete-subtasks", a.completeSubtasks).Methods("POST")
	a.BaseRoutes.Workspaces.HandleFunc("/{id}/table/{tid}/view/{view_id}/filter", a.getFilteredTableData)
	a.BaseRoutes.Workspaces.HandleFunc("/{id}/update", a.updateWorkspace)
	a.BaseRoutes.Workspaces.HandleFunc("/{id}/table/{tid}/update", a.updateWorkspaceTable)
	a.BaseRoutes.Workspaces.HandleFunc("/{id}/folder/{folder_id}/delete", a.deleteWorkspaceFolder)
	a.BaseRoutes.Workspaces.HandleFunc("/{id}/folder/{folder_id}/linking-tables", a.getLinkingTables).Methods("GET")
	a.BaseRoutes.Workspaces.HandleFunc("/{id}/table/{tid}/linking-tables", a.getLinkingTables).Methods("GET")
	a.BaseRoutes.Workspaces.HandleFunc("/{id}/folder/{folder_id}/update", a.updateWorkspaceFolder)
	a.BaseRoutes.Workspaces.HandleFunc("/{id}/move-item", a.moveWorkspaceItem).Methods("POST")
	a.BaseRoutes.Workspaces.HandleFunc("/tasks/status-list", a.getAllStatusIndex)
	a.BaseRoutes.Workspaces.HandleFunc("/{id}/table/{tid}/header/update", a.updateGridHeaders)
	a.BaseRoutes.Workspaces.HandleFunc("/{id}/member/{user_id}/update-role", a.updateWorkspaceMemberRole)
	a.BaseRoutes.Workspaces.HandleFunc("/{id}/table/{tid}/view/{view_id}/save-sort", a.saveGridSort)
	a.BaseRoutes.Workspaces.HandleFunc("/{id}/member/role/create", a.createProjectWorkspaceRole)
	a.BaseRoutes.Workspaces.HandleFunc("/{id}/member/role/{rid}/update", a.updateProjectWorkspaceRole)
	a.BaseRoutes.Workspaces.HandleFunc("/{id}/member/roles/get", a.getProjectWorkspaceRoles)
	a.BaseRoutes.Workspaces.HandleFunc("/{id}/member/role/{rid}/delete", a.deleteProjectWorkspaceRole)
	a.BaseRoutes.Workspaces.HandleFunc("/{id}/tables/for-role-editor", a.getWorkspaceTablesForRoleEditor).Methods("GET")
	a.BaseRoutes.Workspaces.HandleFunc("/{id}/member/role/{rid}/table-permissions", a.getRoleTablePermissions).Methods("GET")
	a.BaseRoutes.Workspaces.HandleFunc("/{id}/member/role/{rid}/table-permissions", a.updateRoleTablePermissions).Methods("PUT")
	a.BaseRoutes.Workspaces.HandleFunc("/{id}/users/me", a.getWorkspaceCurrentUser)
	a.BaseRoutes.Workspaces.HandleFunc("/{id}/files", a.uploadWorkspaceFile).Methods("POST")
	a.BaseRoutes.Workspaces.HandleFunc("/{id}/file/{file_id}", a.getWorkspaceFile).Methods("GET")
	a.BaseRoutes.Workspaces.HandleFunc("/{workspace_id}/file/{file_id}", a.deleteWorkspaceFile).Methods("DELETE")
	a.BaseRoutes.Workspaces.HandleFunc("/{id}/table/{tid}/view/{fid}/save-filter", a.saveWorkspaceFilter).Methods("POST")
	a.BaseRoutes.Workspaces.HandleFunc("/{id}/table/{tid}/view/{fid}/update-filter", a.updateWorkspaceFilter).Methods("PUT")
	a.BaseRoutes.Workspaces.HandleFunc("/{id}/table/{tid}/view/{fid}/update-filter-active", a.updateWorkspaceFilterActive).Methods("PUT")
	a.BaseRoutes.Workspaces.HandleFunc("/{id}/table/{tid}/view/{fid}/saved-filter/{filter_id}", a.deleteWorkspaceFilter).Methods("DELETE")
	a.BaseRoutes.Workspaces.HandleFunc("/{id}/table/{tid}/view/{view_id}/set-visibility", a.setViewVisibility).Methods("POST")
	a.BaseRoutes.Workspaces.HandleFunc("/{id}/table/{tid}/view/{view_id}/shares", a.getViewShares).Methods("GET")
	a.BaseRoutes.Workspaces.HandleFunc("/{id}/table/{tid}/view/{view_id}/shares", a.setViewShares).Methods("POST")
	a.BaseRoutes.Workspaces.Use(a.RequireSession)
	a.BaseRoutes.Workspaces.Use(a.RequireCSRF)
	a.BaseRoutes.Workspaces.Use(a.RequireTableInWorkspace)
	a.BaseRoutes.Workspaces.Use(a.RecordOriginClient)
}

func (a *API) updateWorkspaceFilterActive(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	params := mux.Vars(r)
	workspaceID := params["id"]
	tableID := params["tid"]
	viewID := params["fid"]

	var requestData struct {
		FilterID string `json:"filter_id"`
		IsActive bool   `json:"is_active"`
	}

	if !decodeBody(w, r, &requestData) {
		return
	}

	if requestData.FilterID == "" {
		respondAppError(w, r, model.NewAppError("request.invalid", http.StatusBadRequest))
		return
	}

	appErr = a.app.UpdateWorkspaceFilterActive(
		workspaceID,
		tableID,
		viewID,
		requestData.FilterID,
		requestData.IsActive,
		*user,
	)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *API) updateWorkspaceFilter(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	params := mux.Vars(r)
	workspaceID := params["id"]
	tableID := params["tid"]
	viewID := params["fid"]

	var requestData struct {
		FilterID      string              `json:"filter_id"`
		SavedFilterID string              `json:"saved_filter_id"`
		Name          string              `json:"name"`
		Type          string              `json:"type"`
		Filters       model.FilterPayload `json:"filters"`
		IsPrivate     bool                `json:"is_private"`
		IsActive      bool                `json:"is_active"`
	}

	if !decodeBody(w, r, &requestData) {
		return
	}

	appErr = a.app.UpdateWorkspaceFilter(
		r.Context(),
		workspaceID,
		tableID,
		viewID,
		requestData.FilterID,
		requestData.SavedFilterID,
		requestData.Name,
		requestData.Filters,
		requestData.IsPrivate,
		*user,
		requestData.IsActive,
	)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *API) deleteWorkspaceFilter(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	params := mux.Vars(r)
	workspaceID := params["id"]
	filterID := params["filter_id"]

	appErr = a.app.DeleteSavedFilter(*user, workspaceID, filterID)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (a *API) saveWorkspaceFilter(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	params := mux.Vars(r)
	workspaceID := params["id"]
	tableID := params["tid"]
	viewID := params["fid"]

	var requestData struct {
		Name      string              `json:"name"`
		Type      string              `json:"type"`
		Filters   model.FilterPayload `json:"filters"`
		IsPrivate bool                `json:"is_private"`
	}

	if !decodeBody(w, r, &requestData) {
		return
	}

	savedFilterID, filterID, appErr := a.app.SaveWorkspaceFilter(
		r.Context(),
		workspaceID,
		tableID,
		viewID,
		requestData.Name,
		requestData.Type,
		requestData.Filters,
		requestData.IsPrivate,
		*user,
	)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, struct {
		ID       string `json:"id"`
		FilterID string `json:"filter_id"`
	}{ID: savedFilterID, FilterID: filterID})
}

func (a *API) deleteWorkspaceFile(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	workspaceID := params["workspace_id"]
	fileID := params["file_id"]

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	if workspaceID == "" || fileID == "" {
		respondAppError(w, r, model.NewAppError("request.invalid", http.StatusBadRequest))
		return
	}

	appErr = a.app.DeleteWorkspaceAttachment(r.Context(), fileID, workspaceID, *user)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *API) getWorkspaceFile(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	workspaceID := params["id"]
	fileID := params["file_id"]

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	if workspaceID == "" || fileID == "" {
		respondAppError(w, r, model.NewAppError("request.invalid", http.StatusBadRequest))
		return
	}

	attachment, appErr := a.app.GetWorkspaceAttachment(fileID, workspaceID, *user)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	if attachment.WorkspaceID != workspaceID {
		respondAppError(w, r, model.NewAppError("workspace.forbidden", http.StatusForbidden))
		return
	}

	setUploadHeaders(w, attachment.MimeType, attachment.Name)
	w.Header().Set("Cache-Control", "public, max-age=31536000")

	fileExt2 := filepath.Ext(attachment.Name)
	if attachment.StorageID != "" {
		subPath := attachment.WorkspaceID + "/" + attachment.TableID + "/" + attachment.TaskID + "/" + attachment.ID + fileExt2
		filePath, err := a.app.BuildFilePath(attachment.StorageID, subPath, model.AppProjects)
		if err != nil {
			tlog.Errorw("Failed to build project file path", "file_id", fileID, "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		backend, exists := a.app.FileStorageObjects[attachment.StorageID]
		if !exists {
			tlog.Errorw("Storage backend not available", "storage_id", attachment.StorageID)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if err := backend.ServeFile(filePath, w, r); err != nil {
			tlog.Errorw("Failed to serve project file", "file_id", fileID, "error", err)
		}

		return
	}

	filePath := path.Join("data", "projects",
		attachment.WorkspaceID,
		attachment.TableID,
		attachment.TaskID,
		attachment.ID+fileExt2)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		respondAppError(w, r, model.NewAppError("workspace.attachment_retrieval_failed", http.StatusNotFound))
		return
	}

	http.ServeFile(w, r, filePath)
}

func (a *API) uploadWorkspaceFile(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	workspaceID := params["id"]

	r.Body = http.MaxBytesReader(w, r.Body, 100<<20) // 100 MB

	err := r.ParseMultipartForm(100 << 20)
	if err != nil {
		tlog.Errorw("Failed to parse multipart form", "workspace_id", workspaceID, "error", err)
		if err.Error() == "http: request body too large" {
			respondAppError(w, r, model.NewAppError("workspace.file_upload_failed", http.StatusRequestEntityTooLarge))
		} else {
			respondAppError(w, r, model.NewAppError("request.invalid", http.StatusBadRequest))
		}

		return
	}

	tableID := r.FormValue("table_id")
	taskID := r.FormValue("task_id")
	field := r.FormValue("field")
	fieldID := r.FormValue("field_id")

	if workspaceID == "" || tableID == "" || taskID == "" {
		respondAppError(w, r, model.NewAppError("request.invalid", http.StatusBadRequest))
		return
	}

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	uploadedFiles, appErr := a.app.UploadWorkspaceFiles(workspaceID, tableID, taskID, field, fieldID, *user, r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, uploadedFiles)
}

func (a *API) getWorkspaceCurrentUser(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	permissions, appErr := a.app.GetWorkspaceCurrentUser(*user, params["id"])
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	permissions.UserInfo = *user

	respondJSON(w, http.StatusOK, permissions)
}

func (a *API) deleteProjectWorkspaceRole(w http.ResponseWriter, r *http.Request) {
	if !a.app.Server.License.HasWorkspaceRoles() {
		respondAppError(w, r, model.NewAppError("workspace.roles_license_required", http.StatusPaymentRequired))
		return
	}

	params := mux.Vars(r)

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	deleteRole, appErr := a.app.DeleteProjectWorkspaceRole(r.Context(), *user, params["id"], params["rid"])
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, deleteRole)
}

func (a *API) getRoleTablePermissions(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	perms, appErr := a.app.GetRoleTablePermissions(*user, params["id"], params["rid"])
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, perms)
}

func (a *API) getWorkspaceTablesForRoleEditor(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	params := mux.Vars(r)

	tables, folders, appErr := a.app.GetWorkspaceTablesForRoleEditor(params["id"], *user)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"tables":  tables,
		"folders": folders,
	})
}

func (a *API) updateRoleTablePermissions(w http.ResponseWriter, r *http.Request) {
	if !a.app.Server.License.HasWorkspaceRoles() {
		respondAppError(w, r, model.NewAppError("workspace.roles_license_required", http.StatusPaymentRequired))
		return
	}

	params := mux.Vars(r)

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	var items []struct {
		TableID string `json:"table_id"`
		Action  string `json:"action"`
	}
	if !decodeBody(w, r, &items) {
		return
	}

	perms := make([]model.TablePermission, len(items))
	for i, item := range items {
		perms[i] = model.TablePermission{
			ID:      model.NewID(),
			TableID: item.TableID,
			Action:  item.Action,
		}
	}

	appErr = a.app.UpdateRoleTablePermissions(*user, params["id"], params["rid"], perms)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *API) getProjectWorkspaceRoles(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	roles, appErr := a.app.GetProjectWorkspaceRoles(params["id"])
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, roles)
}

func (a *API) updateProjectWorkspaceRole(w http.ResponseWriter, r *http.Request) {
	if !a.app.Server.License.HasWorkspaceRoles() {
		respondAppError(w, r, model.NewAppError("workspace.roles_license_required", http.StatusPaymentRequired))
		return
	}

	params := mux.Vars(r)

	var s model.ProjectWorkspaceRolePatch
	if !decodeBody(w, r, &s) {
		return
	}

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	role, appErr := a.app.UpdateProjectWorkspaceRole(*user, params["id"], params["rid"], s)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, role)
}

func (a *API) createProjectWorkspaceRole(w http.ResponseWriter, r *http.Request) {
	if !a.app.Server.License.HasWorkspaceRoles() {
		respondAppError(w, r, model.NewAppError("workspace.roles_license_required", http.StatusPaymentRequired))
		return
	}

	params := mux.Vars(r)

	var s model.ProjectWorkspaceRole
	if !decodeBody(w, r, &s) {
		return
	}

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	role, appErr := a.app.CreateProjectWorkspaceRole(*user, params["id"], s)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, role)
}

func (a *API) saveGridSort(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	params := mux.Vars(r)

	s := struct {
		Sort string `json:"sort"`
	}{}

	if ok := decodeBody(w, r, &s); !ok {
		return
	}

	_, appErr = a.app.SaveGridSort(params["id"], params["tid"], params["view_id"], s.Sort, *user)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *API) updateWorkspaceMemberRole(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	var s struct {
		Role string `json:"role"`
	}
	if !decodeBody(w, r, &s) {
		return
	}

	params := mux.Vars(r)

	updateRole, appErr := a.app.UpdateWorkspaceMemberRole(params["id"], params["user_id"], s.Role, *user)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, updateRole)
}

func (a *API) updateGridHeaders(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	s := struct {
		Headers []model.WorkspaceHeaders `json:"headers"`
		ViewID  string                   `json:"view_id"`
	}{}

	if ok := decodeBody(w, r, &s); !ok {
		return
	}

	params := mux.Vars(r)

	updateHeaders, appErr := a.app.UpdateGridHeaders(r.Context(), params["id"], params["tid"], s.Headers, *user, s.ViewID)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, updateHeaders)
}

func (a *API) getAllStatusIndex(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	if user.Role != "system_admin" {
		respondAppError(w, r, model.NewAppError("workspace.forbidden", http.StatusForbidden))
		return
	}

	s := struct {
		WorkspaceIds []string `json:"workspace_ids"`
	}{}

	if r.ContentLength > 0 {
		if !decodeBody(w, r, &s) {
			return
		}
	}

	statusIndex, appErr := a.app.GetAllStatusIndex(s.WorkspaceIds)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, statusIndex)
}

func (a *API) updateWorkspaceFolder(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	var s struct {
		Name string `json:"name"`
	}
	if !decodeBody(w, r, &s) {
		return
	}

	params := mux.Vars(r)

	ok, appErr := a.app.UpdateWorkspaceFolder(r.Context(), params["id"], params["folder_id"], s.Name, *user)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, ok)
}

func (a *API) deleteWorkspaceFolder(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	params := mux.Vars(r)

	ok, appErr := a.app.DeleteWorkspaceFolder(r.Context(), params["id"], params["folder_id"], *user)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, ok)
}

func (a *API) moveWorkspaceItem(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	var s struct {
		ItemID         string `json:"item_id"`
		ItemType       string `json:"item_type"`
		TargetFolderID string `json:"target_folder_id"`
	}
	if !decodeBody(w, r, &s) {
		return
	}

	params := mux.Vars(r)

	if appErr := a.app.MoveWorkspaceItem(r.Context(), params["id"], s.ItemType, s.ItemID, s.TargetFolderID, *user); appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, true)
}

func (a *API) deleteWorkspaceMember(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	params := mux.Vars(r)

	ok, appErr := a.app.DeleteWorkspaceMember(params["id"], params["member_id"], *user)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, ok)
}

func (a *API) updateWorkspaceTable(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	var s struct {
		Name string `json:"name"`
	}
	if !decodeBody(w, r, &s) {
		return
	}

	params := mux.Vars(r)

	data, appErr := a.app.UpdateWorkspaceTable(r.Context(), params["id"], params["tid"], s.Name, *user)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, data)
}

func (a *API) updateWorkspace(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	var s struct {
		Name        string  `json:"name"`
		Description *string `json:"description"`
	}
	if !decodeBody(w, r, &s) {
		return
	}

	params := mux.Vars(r)

	data, appErr := a.app.UpdateWorkspace(params["id"], s.Name, s.Description, *user)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, data)
}

func (a *API) addWorkspaceFolder(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	var s struct {
		Name           string `json:"name"`
		ParentFolderID string `json:"parent_folder_id,omitempty"`
	}
	if !decodeBody(w, r, &s) {
		return
	}

	params := mux.Vars(r)

	addFolder, appErr := a.app.AddWorkspaceFolder(r.Context(), params["id"], s.Name, s.ParentFolderID, *user)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, addFolder)
}

func (a *API) getTaskComments(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	params := mux.Vars(r)

	comments, appErr := a.app.GetTaskComments(r.Context(), params["id"], params["tid"], params["taskid"], *user)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, comments)
}

func (a *API) addTaskComment(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	s := struct {
		Comment        string   `json:"comment"`
		MentionedUsers []string `json:"mentioned_users"`
	}{}

	if ok := decodeBody(w, r, &s); !ok {
		return
	}

	params := mux.Vars(r)

	addComment, appErr := a.app.AddTaskComment(r.Context(), params["id"], params["tid"], params["taskid"], s.Comment, *user, s.MentionedUsers)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, addComment)
}

func (a *API) getAllWorkspaceTasks(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	if !a.app.SessionHasPermission(*user, model.ProjectSectionPermissions.PermissionViewProjects) {
		respondAppError(w, r, model.NewAppError("permission.forbidden", http.StatusForbidden))
		return
	}

	s := struct {
		WorkspaceIds []string            `json:"workspace_ids"`
		Page         int                 `json:"page"`
		Limit        int                 `json:"limit"`
		SkipCount    bool                `json:"skip_count"`
		FiltersOnly  bool                `json:"filters_only"`
		Filters      model.FilterPayload `json:"filters"`
	}{}

	if r.ContentLength > 0 {
		if !decodeBody(w, r, &s) {
			return
		}
	}

	if s.Page < 1 {
		s.Page = 1
	}

	if s.Limit <= 0 || s.Limit > 500 {
		s.Limit = 100
	}

	if s.FiltersOnly {
		page, appErr := a.app.GetTaskReportFilters(*user)
		if appErr != nil {
			respondAppError(w, r, appErr)
			return
		}

		respondJSON(w, http.StatusOK, page)
		return
	}

	tasks, appErr := a.app.GetAllWorkspaceTasks(r.Context(), s.WorkspaceIds, *user, s.Page, s.Limit, !s.SkipCount, s.Filters)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, tasks)
}

// exportAllWorkspaceTasks streams the task report as a CSV file. It is a GET
// so the browser can download it straight to disk from a link.
func (a *API) exportAllWorkspaceTasks(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	if !a.app.SessionHasPermission(*user, model.ProjectSectionPermissions.PermissionViewProjects) {
		respondAppError(w, r, model.NewAppError("permission.forbidden", http.StatusForbidden))
		return
	}

	q := r.URL.Query()
	var filters model.FilterPayload
	if raw := q.Get("filters"); raw != "" {
		if err := json.Unmarshal([]byte(raw), &filters); err != nil {
			respondAppError(w, r, model.NewAppError("request.invalid", http.StatusBadRequest))
			return
		}
	}

	var labels map[string]string
	if raw := q.Get("labels"); raw != "" {
		if err := json.Unmarshal([]byte(raw), &labels); err != nil {
			respondAppError(w, r, model.NewAppError("request.invalid", http.StatusBadRequest))
			return
		}
	}

	var workspaceIDs []string
	if raw := q.Get("workspace_ids"); raw != "" {
		workspaceIDs = strings.Split(raw, ",")
	}

	export, appErr := a.app.PrepareTaskExport(r.Context(), workspaceIDs, *user, filters, q.Get("tz"), q.Get("clock"), labels)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	name := export.FileName()
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"; filename*=UTF-8''%s`, sanitizeHeaderFilename(name), url.PathEscape(name)))
	w.Header().Set("X-Content-Type-Options", "nosniff")
	flusher, _ := w.(http.Flusher)
	flush := func() {
		if flusher != nil {
			flusher.Flush()
		}
	}

	if err := export.WriteCSV(r.Context(), w, flush); err != nil {
		tlog.Errorw("Failed to export tasks", "user_id", user.ID, "error", err)
	}
}

func (a *API) assignedToMe(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	if !a.app.SessionHasPermission(*user, model.ProjectSectionPermissions.PermissionViewProjects) {
		respondAppError(w, r, model.NewAppError("permission.forbidden", http.StatusForbidden))
		return
	}

	q := r.URL.Query()
	limit := parse.Int(q.Get("limit"), 50)
	if limit <= 0 || limit > 200 {
		limit = 50
	}

	assignedToMe, appErr := a.app.GetAssignedToMe(r.Context(), user.ID, q.Get("tz"), q.Get("section"), q.Get("after"), limit)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, assignedToMe)
}

func (a *API) deleteWorkspaceTableSingleField(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	s := struct {
		LinkedID string `json:"linked_table_id"`
	}{}

	if ok := decodeBody(w, r, &s); !ok {
		return
	}

	params := mux.Vars(r)

	deleteView, appErr := a.app.DeleteWorkspaceTableSingleField(r.Context(), params["id"], params["tid"], params["fieldid"], s.LinkedID, *user)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, deleteView)
}

func (a *API) deleteWorkspaceTableField(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	s := struct {
		LinkBothDirections bool   `json:"link_both_directions"`
		LinkedID           string `json:"linked_id"`
	}{}

	if ok := decodeBody(w, r, &s); !ok {
		return
	}

	params := mux.Vars(r)

	deleteView, appErr := a.app.DeleteWorkspaceTableField(r.Context(), params["id"], params["tid"], params["fieldid"], s.LinkBothDirections, s.LinkedID, *user)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, deleteView)
}

func (a *API) deleteWorkspaceTableView(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	params := mux.Vars(r)

	deleteView, appErr := a.app.DeleteWorkspaceTableView(r.Context(), params["id"], params["tid"], params["viewid"], *user)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, deleteView)
}

func (a *API) deleteWorkspaceTask(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	params := mux.Vars(r)

	deleteTask, appErr := a.app.DeleteWorkspaceTask(r.Context(), params["id"], params["tid"], params["taskid"], *user)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, deleteTask)
}

func (a *API) deleteTable(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	params := mux.Vars(r)

	deleteTable, appErr := a.app.DeleteTable(r.Context(), params["id"], params["tid"], *user)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, deleteTable)
}

func (a *API) getLinkingTables(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	params := mux.Vars(r)

	tables, appErr := a.app.GetLinkingTables(r.Context(), params["id"], params["tid"], params["folder_id"], *user)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, tables)
}

func (a *API) deleteProjectWorkspace(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	params := mux.Vars(r)

	deleteWorkspace, appErr := a.app.DeleteProjectWorkspace(r.Context(), params["id"], *user)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, deleteWorkspace)
}

func (a *API) addMemberToTask(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	s := struct {
		UserID    string `json:"user_id"`
		FieldName string `json:"field_name"`
	}{}

	if ok := decodeBody(w, r, &s); !ok {
		return
	}

	params := mux.Vars(r)

	addMemberToTask, appErr := a.app.AddMemberToTask(r.Context(), params["id"], params["tid"], params["taskid"], s.UserID, s.FieldName, *user)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, addMemberToTask)
}

func (a *API) updateDisplayName(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	var s struct {
		DisplayName    string `json:"display_name"`
		TableFieldName string `json:"header_name"`
		Visible        bool   `json:"visible"`
	}
	if !decodeBody(w, r, &s) {
		return
	}

	params := mux.Vars(r)

	val, appErr := a.app.UpdateDisplayName(r.Context(), params["id"], params["tid"], params["fid"], s.DisplayName, s.TableFieldName, s.Visible, *user)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, val)
}

func (a *API) updateWorkspaceTableField(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	s := struct {
		ColumnName string `json:"column_name"`
		Width      int    `json:"width"`
		ViewID     string `json:"view_id"`
	}{}

	if ok := decodeBody(w, r, &s); !ok {
		return
	}

	params := mux.Vars(r)
	workspaceID := params["id"]
	tableID := params["tid"]

	if s.ColumnName == "" {
		respondAppError(w, r, model.NewAppError("request.invalid", http.StatusBadRequest))
		return
	}

	if s.Width < 50 {
		respondAppError(w, r, model.NewAppError("request.invalid", http.StatusBadRequest))
		return
	}

	val, appErr := a.app.UpdateColumnWidth(r.Context(), workspaceID, tableID, s.ViewID, s.ColumnName, s.Width, *user)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, val)
}

func (a *API) editFieldFormula(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	s := struct {
		FieldName string             `json:"field_name"`
		Formula   *model.FormulaSpec `json:"formula"`
	}{}

	if ok := decodeBody(w, r, &s); !ok {
		return
	}

	params := mux.Vars(r)
	workspaceID := params["id"]
	tableID := params["tid"]

	if s.FieldName == "" || s.Formula == nil {
		respondAppError(w, r, model.NewAppError("request.invalid", http.StatusBadRequest))
		return
	}

	if appErr := a.app.UpdateCalculationsField(r.Context(), workspaceID, tableID, s.FieldName, s.Formula, *user); appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (a *API) addFieldValue(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	var s struct {
		Name  string `json:"name"`
		Field string `json:"field"`
	}
	if !decodeBody(w, r, &s) {
		return
	}

	params := mux.Vars(r)

	val, appErr := a.app.AddFieldValue(r.Context(), params["id"], params["tid"], s.Name, s.Field, *user)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, val)
}

func (a *API) getKanbanData(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	params := mux.Vars(r)

	var body struct {
		ViewID  string               `json:"view_id"`
		Column  string               `json:"column"`
		After   string               `json:"after"`
		Limit   int                  `json:"limit"`
		Filters *model.FilterPayload `json:"filters,omitempty"`
	}
	if !decodeBody(w, r, &body) {
		return
	}

	var filters model.FilterPayload
	if body.Filters != nil {
		filters = *body.Filters
	}

	if body.Limit <= 0 || body.Limit > 200 {
		body.Limit = 50
	}

	data, appErr := a.app.GetKanbanData(r.Context(), params["id"], params["tid"], body.ViewID, body.Column, body.After, body.Limit, filters, *user)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, data)
}

func (a *API) getTasksByDateRange(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	params := mux.Vars(r)

	var body struct {
		From    int64                `json:"from"`
		To      int64                `json:"to"`
		Filters *model.FilterPayload `json:"filters,omitempty"`
	}
	if !decodeBody(w, r, &body) {
		return
	}

	if body.From == 0 || body.To == 0 {
		respondAppError(w, r, model.NewAppError("workspace.date_range_failed", http.StatusBadRequest))
		return
	}

	var filters model.FilterPayload
	if body.Filters != nil {
		filters = *body.Filters
	}

	data, appErr := a.app.GetTasksByDateRange(r.Context(), params["id"], params["tid"], body.From, body.To, filters, *user)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, data)
}

func (a *API) getTasksByCalendarRange(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	params := mux.Vars(r)

	var body struct {
		From    int64                `json:"from"`
		To      int64                `json:"to"`
		Filters *model.FilterPayload `json:"filters,omitempty"`
	}
	if !decodeBody(w, r, &body) {
		return
	}

	if body.From == 0 || body.To == 0 {
		respondAppError(w, r, model.NewAppError("workspace.calendar_range_invalid", http.StatusBadRequest))
		return
	}

	var filters model.FilterPayload
	if body.Filters != nil {
		filters = *body.Filters
	}

	data, appErr := a.app.GetTasksByCalendarRange(r.Context(), params["id"], params["tid"], body.From, body.To, filters, *user)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, data)
}

func (a *API) getLinkedRecordsLite(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	params := mux.Vars(r)
	q := r.URL.Query()
	limit := parse.Int(q.Get("limit"), 50)
	if limit <= 0 || limit > 200 {
		limit = 50
	}

	offset := max(parse.Int(q.Get("offset"), 0), 0)
	var ids []string
	if raw := q.Get("ids"); raw != "" {
		ids = strings.Split(raw, ",")
		limit = max(limit, len(ids))
	}

	items, appErr := a.app.GetLinkedRecordsLite(r.Context(), params["id"], params["tid"], q.Get("q"), ids, limit, offset, *user)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, items)
}

func (a *API) getTaskPage(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	params := mux.Vars(r)

	var body struct {
		Limit       int                 `json:"limit"`
		Sort        []model.SortParam   `json:"sort"`
		FlatFilters []model.Filter      `json:"flat_filters"`
		Groups      []model.FilterGroup `json:"groups"`
	}
	body.Limit = 100
	if r.ContentLength > 0 {
		if ok := decodeBody(w, r, &body); !ok {
			return
		}
	}

	if body.Limit < 1 {
		body.Limit = 100
	}

	filters := model.FilterPayload{
		FlatFilters: body.FlatFilters,
		Groups:      body.Groups,
		Sort:        body.Sort,
		Limit:       body.Limit,
	}

	page, appErr := a.app.GetTaskPageNumber(r.Context(), params["id"], params["tid"], params["taskid"], body.Limit, filters, *user)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, map[string]int{"page": page})
}

func (a *API) taskShownByFilter(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	params := mux.Vars(r)

	var filters model.FilterPayload
	if !decodeBody(w, r, &filters) {
		return
	}

	shown, appErr := a.app.TaskShownByFilter(r.Context(), params["id"], params["tid"], params["taskid"], filters, *user)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, map[string]bool{"matches": shown})
}

func (a *API) getTableStatusTypes(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	params := mux.Vars(r)
	doneIDs, options, appErr := a.app.GetTableStatusTypes(r.Context(), params["id"], params["tid"], *user)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"done_ids": doneIDs,
		"options":  options,
	})
}

func (a *API) getItemForTableByID(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	params := mux.Vars(r)

	item, headers, appErr := a.app.GetItemForTableByID(r.Context(), params["id"], params["tid"], params["itemid"], *user)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	response := map[string]interface{}{
		"item":    item,
		"headers": headers,
	}

	respondJSON(w, http.StatusOK, response)
}

func (a *API) getSubtasks(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	params := mux.Vars(r)

	subtasks, appErr := a.app.GetSubtasks(r.Context(), params["id"], params["tid"], params["itemid"], *user)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, subtasks)
}

func (a *API) getTaskCompletion(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	params := mux.Vars(r)

	completion, appErr := a.app.GetTaskCompletion(r.Context(), params["id"], params["tid"], params["itemid"], *user)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, completion)
}

func (a *API) completeSubtasks(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	var s struct {
		Status string `json:"status"`
	}
	if !decodeBody(w, r, &s) {
		return
	}

	params := mux.Vars(r)

	completed, appErr := a.app.CompleteSubtasks(r.Context(), params["id"], params["tid"], params["itemid"], s.Status, *user)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, map[string]int{"completed": completed})
}

func (a *API) updateView(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	s := struct {
		Order string `json:"order"`
		Name  string `json:"name"`
	}{}

	if !decodeBody(w, r, &s) {
		return
	}

	params := mux.Vars(r)

	updateViewOrder, appErr := a.app.UpdateView(r.Context(), params["view_id"], params["id"], params["tid"], s.Order, s.Name, *user)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, updateViewOrder)
}

func (a *API) moveKanbanCard(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	var s struct {
		TaskID  string `json:"task_id"`
		Column  string `json:"column"`
		AfterID string `json:"after_id"`
	}
	if !decodeBody(w, r, &s) {
		return
	}

	params := mux.Vars(r)
	if appErr := a.app.MoveKanbanCard(r.Context(), params["id"], params["tid"], params["view_id"], s.TaskID, s.Column, s.AfterID, *user); appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (a *API) setViewVisibility(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	var body struct {
		IsPublic bool `json:"is_public"`
	}
	if !decodeBody(w, r, &body) {
		return
	}

	params := mux.Vars(r)
	if appErr := a.app.SetViewPublic(r.Context(), *user, params["id"], params["tid"], params["view_id"], body.IsPublic); appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{"is_public": body.IsPublic})
}

func (a *API) getViewShares(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	params := mux.Vars(r)
	userIDs, appErr := a.app.GetViewShares(*user, params["id"], params["view_id"])
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{"user_ids": userIDs})
}

func (a *API) setViewShares(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	var body struct {
		UserIDs []string `json:"user_ids"`
	}
	if !decodeBody(w, r, &body) {
		return
	}

	params := mux.Vars(r)
	if appErr := a.app.SetViewShares(r.Context(), *user, params["id"], params["tid"], params["view_id"], body.UserIDs); appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{"user_ids": body.UserIDs})
}

func (a *API) createNewView(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	var s struct {
		Name          string `json:"name"`
		ViewType      string `json:"view_type"`
		GroupField    string `json:"group_field"`
		ParentTableID string `json:"parent_table_id"`
		IsPublic      *bool  `json:"is_public"`
	}
	if !decodeBody(w, r, &s) {
		return
	}

	isPublic := true
	if s.IsPublic != nil {
		isPublic = *s.IsPublic
	}

	params := mux.Vars(r)

	createView, appErr := a.app.CreateView(r.Context(), params["id"], params["tid"], "", s.Name, s.ViewType, isPublic, *user, s.GroupField, s.ParentTableID)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, createView)
}

func (a *API) addMemberToWorkspace(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	s := struct {
		UserIDs []string `json:"user_ids"`
	}{}

	if ok := decodeBody(w, r, &s); !ok {
		return
	}

	params := mux.Vars(r)

	addMemberToWorkspace, appErr := a.app.AddMemberToWorkspace(params["id"], s.UserIDs, *user)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, addMemberToWorkspace)
}

func (a *API) addProjectWorkspaceGroups(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	var req model.AddProjectGroupsRequest
	if ok := decodeBody(w, r, &req); !ok {
		return
	}

	params := mux.Vars(r)
	groups, appErr := a.app.AddGroupsToProjectWorkspace(*user, params["id"], req.GroupIDs, req.Roles)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, groups)
}

func (a *API) getProjectWorkspaceGroups(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	params := mux.Vars(r)
	groups, appErr := a.app.ListProjectWorkspaceGroups(*user, params["id"])
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, groups)
}

func (a *API) removeProjectWorkspaceGroup(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	params := mux.Vars(r)
	if appErr := a.app.RemoveProjectWorkspaceGroup(*user, params["id"], params["group_id"]); appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (a *API) updateProjectWorkspaceGroupRoles(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	var req model.UpdateProjectGroupRolesRequest
	if ok := decodeBody(w, r, &req); !ok {
		return
	}

	params := mux.Vars(r)
	if appErr := a.app.UpdateProjectWorkspaceGroupRoles(*user, params["id"], params["group_id"], req.Roles); appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (a *API) searchWorkspaceMembers(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	params := mux.Vars(r)
	q := r.URL.Query().Get("q")
	limit := parse.Int(r.URL.Query().Get("limit"), 20)
	offset := parse.Int(r.URL.Query().Get("offset"), 0)

	users, appErr := a.app.SearchProjectWorkspaceMembers(*user, params["id"], q, limit, offset)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, users)
}

func (a *API) searchTaskReportMembers(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	var workspaceIDs []string
	for _, id := range strings.Split(r.URL.Query().Get("workspace_ids"), ",") {
		if id = strings.TrimSpace(id); id != "" {
			workspaceIDs = append(workspaceIDs, id)
		}
	}

	limit := parse.Int(r.URL.Query().Get("limit"), 50)

	users, appErr := a.app.SearchTaskReportMembers(r.Context(), *user, workspaceIDs, r.URL.Query().Get("q"), limit)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, users)
}

func (a *API) updateWorkspaceTaskLink(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	var s struct {
		LinkedIDs    []string `json:"linked_ids"`
		Add          []string `json:"add"`
		Remove       []string `json:"remove"`
		SingleSelect bool     `json:"single_select"`
		Field        string   `json:"field"`
		FieldID      string   `json:"field_id"`
	}
	if !decodeBody(w, r, &s) {
		return
	}

	params := mux.Vars(r)
	wsID := params["id"]
	tableID := params["tid"]
	taskID := params["taskid"]

	if !s.SingleSelect {
		updateTaskValue, appErr := a.app.ChangeTaskLinks(r.Context(), wsID, tableID, taskID, s.Field, s.Add, s.Remove, *user)
		if appErr != nil {
			respondAppError(w, r, appErr)
			return
		}

		respondJSON(w, http.StatusOK, updateTaskValue)
		return
	}

	var selected string
	if len(s.LinkedIDs) > 0 {
		selected = s.LinkedIDs[0]
	}

	updatedTask, appErr := a.app.UpdateWorkspaceTask(r.Context(), wsID, tableID, taskID, s.Field, selected, *user, s.FieldID)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	// if we cleared the field, don’t try to fetch linked task data
	if selected == "" {
		respondJSON(w, http.StatusOK, map[string]interface{}{
			"linked_task": nil,
			"cleared":     true,
		})
		return
	}

	// otherwise, resolve the current field value (string) and fetch linked data
	fieldValue, ok := (*updatedTask)[s.Field]
	if !ok {
		respondAppError(w, r, model.NewAppError("request.invalid", http.StatusBadRequest))
		return
	}

	var fieldString string
	switch v := fieldValue.(type) {
	case []byte:
		fieldString = string(v)
	case string:
		fieldString = v
	default:
		respondAppError(w, r, model.NewAppError("request.invalid", http.StatusBadRequest))
		return
	}

	getTaskValueData, appErr := a.app.GetWorkspaceLinkedTaskData(tableID, s.Field, fieldString)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, getTaskValueData)
}

func (a *API) updateWorkspaceSingleSelectName(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	var s struct {
		Field         string `json:"field"`
		Value         string `json:"value"`
		LinkedTableID string `json:"linked_table_id"`
	}
	if !decodeBody(w, r, &s) {
		return
	}

	params := mux.Vars(r)

	updateTaskValue, appErr := a.app.UpdateWorkspaceSingleSelectName(r.Context(), params["id"], params["tid"], params["taskid"], s.Field, s.Value, *user, s.LinkedTableID)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, updateTaskValue)
}

func (a *API) updateWorkspaceTask(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	var s struct {
		Field string `json:"field"`
		Value string `json:"value"`
	}
	if !decodeBody(w, r, &s) {
		return
	}

	params := mux.Vars(r)

	updateTaskValue, appErr := a.app.UpdateWorkspaceTask(r.Context(), params["id"], params["tid"], params["taskid"], s.Field, s.Value, *user, "")
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, updateTaskValue)
}

func (a *API) createWorkspaceTask(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	var s struct {
		Name         string              `json:"name"`
		SingleSelect string              `json:"single_select"`
		Section      string              `json:"section"`
		ParentTaskID string              `json:"parent_task_id,omitempty"`
		Fields       model.NewTaskFields `json:"fields"`
	}
	if !decodeBody(w, r, &s) {
		return
	}

	params := mux.Vars(r)

	newTask, appErr := a.app.CreateWorkspaceTask(r.Context(), params["id"], params["tid"], s.Name, s.SingleSelect, s.Section, s.Fields, *user, s.ParentTaskID)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, newTask)
}

func (a *API) createWorkspaceTableField(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	s := struct {
		FieldType                     string `json:"fieldType"`
		LinkedTableID                 string `json:"linkedTableID"`
		SelectedType                  string `json:"selectedType"`
		LinkBothDirections            bool   `json:"linkBothDirections"`
		FieldNameDisplay              string `json:"fieldNameDisplay"`
		FieldNameInSecondTableDisplay string `json:"fieldNameInSecondTableDisplay"`

		HeaderUsage string             `json:"header_usage,omitempty"`
		Formula     *model.FormulaSpec `json:"formula,omitempty"`
	}{}

	if ok := decodeBody(w, r, &s); !ok {
		return
	}

	params := mux.Vars(r)

	workspaceTableField, appErr := a.app.CreateWorkspaceTableField(
		r.Context(),
		params["id"],
		params["tid"],
		s.FieldType,
		s.LinkedTableID,
		*user,
		s.SelectedType,
		s.LinkBothDirections,
		s.FieldNameDisplay,
		s.FieldNameInSecondTableDisplay,
		s.HeaderUsage,
		s.Formula,
	)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, workspaceTableField)
}

func (a *API) getWorkspaceTables(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	params := mux.Vars(r)
	pageNum := parse.Int(r.URL.Query().Get("page"), 1)
	pageLimit := parse.Int(r.URL.Query().Get("limit"), 0)
	workspaceTables, appErr := a.app.GetWorkspaceTables(params["id"], *user, params["data"], pageNum, pageLimit)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, workspaceTables)
}

func (a *API) createWorkspaceTable(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	params := mux.Vars(r)

	s := struct {
		Name           string `json:"name"`
		FolderID       string `json:"folder_id,omitempty"`
		CreateSubtasks bool   `json:"create_subtasks"`
	}{}

	if ok := decodeBody(w, r, &s); !ok {
		return
	}

	createWorkspaceTable, appErr := a.app.CreateWorkspaceTable(r.Context(), params["id"], s.Name, false, "", *user, false, false, "", s.FolderID, s.CreateSubtasks)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, createWorkspaceTable)
}

func (a *API) createProjectWorkspace(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	if !a.app.SessionHasPermission(*user, model.ProjectSectionPermissions.PermissionCreateProject) {
		respondAppError(w, r, model.NewAppError("permission.forbidden", http.StatusForbidden))
		return
	}

	var s struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if !decodeBody(w, r, &s) {
		return
	}

	createWorkspace, appErr := a.app.CreateProjectWorkspace(user.ID, s.Name, s.Description)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, createWorkspace)
}

func (a *API) getFilteredTableData(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	params := mux.Vars(r)

	s := struct {
		Filters   model.FilterPayload `json:"filters"`
		SkipCount bool                `json:"skip_count"`
		CountOnly bool                `json:"count_only"`
	}{}

	if !decodeBody(w, r, &s) {
		return
	}

	if s.CountOnly {
		count, appErr := a.app.GetFilteredTableCount(r.Context(), params["id"], params["tid"], params["view_id"], s.Filters, *user)
		if appErr != nil {
			respondAppError(w, r, appErr)
			return
		}

		respondJSON(w, http.StatusOK, map[string]int{"total": count, "root_total": count})
		return
	}

	tableData, appErr := a.app.GetFilteredTableData(r.Context(), params["id"], params["tid"], params["view_id"], s.Filters, *user, !s.SkipCount)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, tableData)
}

func (a *API) getProjectWorkspaceDetails(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	params := mux.Vars(r)

	workspaces, appErr := a.app.GetProjectWorkspaceDetails(*user, params["id"])
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, workspaces)
}

func (a *API) getProjectWorkspaces(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	if !a.app.SessionHasPermission(*user, model.ProjectSectionPermissions.PermissionViewProjects) {
		respondAppError(w, r, model.NewAppError("permission.forbidden", http.StatusForbidden))
		return
	}

	workspaces, appErr := a.app.GetProjectWorkspaces(r.Context(), *user)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, workspaces)
}
