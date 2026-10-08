// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"io"
	"maps"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/twigex/twigex/interfaces"
	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/tlog"
)

func (a *App) GetUserWorkspaceAttachment(fileID string, user model.User) (*model.WorkspaceAttachment, *model.AppError) {
	attachment, err := a.Store.Workspace.GetUserAttachment(fileID)
	if err != nil {
		tlog.Errorw("Failed to retrieve workspace attachment",
			"file_id", fileID,
			"user_id", user.ID,
			"error", err,
		)
		return nil, model.NewAppError("workspace.attachment_retrieval_failed", http.StatusInternalServerError)
	}

	if attachment == nil {
		return nil, model.NewAppError("workspace.attachment_not_found", http.StatusNotFound)
	}

	isMember, err := a.Store.Workspace.IsMember(attachment.WorkspaceID, user.ID)
	if err != nil {
		tlog.Errorw("Failed to check workspace membership",
			"workspace_id", attachment.WorkspaceID,
			"user_id", user.ID,
			"error", err,
		)
		return nil, model.NewAppError("workspace.member_retrieval_failed", http.StatusInternalServerError)
	}

	if !isMember {
		return nil, model.NewAppError("workspace.forbidden", http.StatusForbidden)
	}

	if appErr := a.requireTaskAccess(user, attachment.WorkspaceID, attachment.TableID, attachment.TaskID); appErr != nil {
		return nil, appErr
	}

	return attachment, nil
}

func (a *App) GetProjectWorkspaceRoles(workspaceID string) ([]model.ProjectWorkspaceRole, *model.AppError) {
	roles, err := a.Store.Workspace.GetRoles(workspaceID)
	if err != nil {
		tlog.Errorw("Failed to retrieve workspace roles",
			"workspace_id", workspaceID,
			"error", err,
		)
		return nil, model.NewAppError("workspace.role_retrieval_failed", http.StatusInternalServerError)
	}

	return roles, nil
}

func (a *App) UpdateProjectWorkspaceRole(user model.User, workspaceID, roleID string, patch model.ProjectWorkspaceRolePatch) (*model.ProjectWorkspaceRole, *model.AppError) {
	if !a.ProjectWorkspaceHasPermission(user, workspaceID, model.PermissionUpdateRoles) {
		return nil, model.NewAppError("workspace.forbidden", http.StatusForbidden)
	}

	if appErr := a.guardProjectRoleEdit(user, workspaceID, roleID); appErr != nil {
		return nil, appErr
	}

	member, err := a.Store.Workspace.GetUserByUserID(user.ID, workspaceID)
	if err != nil {
		tlog.Errorw("Failed to retrieve workspace member",
			"workspace_id", workspaceID,
			"user_id", user.ID,
			"error", err,
		)
		return nil, model.NewAppError("workspace.member_retrieval_failed", http.StatusInternalServerError)
	}

	if member == nil {
		return nil, model.NewAppError("workspace.member_not_found", http.StatusForbidden)
	}

	role, err := a.Store.Workspace.GetRoleByID(workspaceID, roleID)
	if err != nil {
		tlog.Errorw("Failed to retrieve workspace role",
			"workspace_id", workspaceID,
			"role_id", roleID,
			"error", err,
		)
		return nil, model.NewAppError("workspace.role_retrieval_failed", http.StatusInternalServerError)
	}

	role.Patch(&patch)

	if err = a.Store.Workspace.UpdateRole(workspaceID, role); err != nil {
		tlog.Errorw("Failed to update workspace role",
			"workspace_id", workspaceID,
			"role_id", roleID,
			"error", err,
		)
		return nil, model.NewAppError("workspace.role_update_failed", http.StatusInternalServerError)
	}

	return role, nil
}

func (a *App) CreateProjectWorkspaceRole(user model.User, workspaceID string, role model.ProjectWorkspaceRole) (*model.ProjectWorkspaceRole, *model.AppError) {
	if !a.ProjectWorkspaceHasPermission(user, workspaceID, model.PermissionCreateRoles) {
		return nil, model.NewAppError("workspace.forbidden", http.StatusForbidden)
	}

	role.ID = model.NewID()
	role.WorkspaceID = workspaceID
	role.Name = strings.ReplaceAll(role.Name, " ", "_")

	if appErr := a.checkProjectRoleNameFree(workspaceID, role.Name); appErr != nil {
		return nil, appErr
	}

	r, err := a.Store.Workspace.CreateRole(role)
	if err != nil {
		tlog.Errorw("Failed to create workspace role",
			"workspace_id", workspaceID,
			"user_id", user.ID,
			"error", err,
		)
		return nil, model.NewAppError("workspace.role_create_failed", http.StatusInternalServerError)
	}

	return r, nil
}

func (a *App) DeleteProjectWorkspaceRole(ctx context.Context, user model.User, workspaceID string, roleID string) (bool, *model.AppError) {
	if !a.ProjectWorkspaceHasPermission(user, workspaceID, model.PermissionDeleteRoles) {
		return false, model.NewAppError("workspace.forbidden", http.StatusForbidden)
	}

	member, err := a.Store.Workspace.GetUserByUserID(user.ID, workspaceID)
	if err != nil {
		tlog.Errorw("Failed to retrieve workspace member",
			"workspace_id", workspaceID,
			"user_id", user.ID,
			"error", err,
		)
		return false, model.NewAppError("workspace.member_retrieval_failed", http.StatusInternalServerError)
	}

	if member == nil {
		return false, model.NewAppError("workspace.member_not_found", http.StatusForbidden)
	}

	role, err := a.Store.Workspace.GetRoleByID(workspaceID, roleID)
	if err != nil {
		tlog.Errorw("Failed to retrieve workspace role",
			"workspace_id", workspaceID,
			"role_id", roleID,
			"error", err,
		)
		return false, model.NewAppError("workspace.role_retrieval_failed", http.StatusInternalServerError)
	}

	if role.Name == model.ProjectWorkspaceAdminRoleID || role.Name == model.ProjectWorkspaceUserRoleID {
		return false, model.NewAppError("workspace.role_delete_default_forbidden", http.StatusConflict)
	}

	ctx, cancel := a.dbCtx(ctx)
	defer cancel()

	err = a.Store.Workspace.DeleteRole(ctx, workspaceID, roleID, role.Name, model.ProjectWorkspaceUserRoleID)
	if err != nil {
		tlog.Errorw("Failed to delete workspace role",
			"workspace_id", workspaceID,
			"role_id", roleID,
			"error", err,
		)
		return false, model.NewAppError("workspace.role_delete_failed", http.StatusInternalServerError)
	}

	return true, nil
}

func (a *App) UpdateWorkspaceMemberRole(workspaceID string, memberID string, role string, user model.User) (bool, *model.AppError) {
	if !a.ProjectWorkspaceHasPermission(user, workspaceID, model.PermissionUpdateRoles) {
		return false, model.NewAppError("workspace.forbidden", http.StatusForbidden)
	}

	if appErr := a.guardProjectMemberRole(user, workspaceID, memberID, role); appErr != nil {
		return false, appErr
	}

	return a.setProjectWorkspaceMemberRole(workspaceID, memberID, role, user)
}

func (a *App) setProjectWorkspaceMemberRole(workspaceID, memberID, role string, user model.User) (bool, *model.AppError) {
	// Allow clearing a role (role == "") so admins can remove restrictions after expiry.
	if role != "" && !a.Server.License.HasWorkspaceRoles() {
		return false, model.NewAppError("workspace.roles_license_required", http.StatusPaymentRequired)
	}

	roles, appErr := a.validProjectWorkspaceRoles(workspaceID, strings.Fields(role))
	if appErr != nil {
		return false, appErr
	}

	updated, err := a.Store.Workspace.UpdateMemberRole(workspaceID, memberID, strings.Join(roles, " "), user)
	if err != nil {
		tlog.Errorw("Failed to update workspace member role",
			"workspace_id", workspaceID,
			"member_id", memberID,
			"error", err,
		)
		return false, model.NewAppError("workspace.member_role_update_failed", http.StatusInternalServerError)
	}

	return updated, nil
}

func (a *App) UpdateWorkspaceFolder(ctx context.Context, workspaceID string, folderID string, name string, user model.User) (bool, *model.AppError) {
	if !a.ProjectWorkspaceHasPermission(user, workspaceID, model.PermissionUpdateWorkspaceFolder) {
		return false, model.NewAppError("workspace.forbidden", http.StatusForbidden)
	}

	ok, err := a.Store.Workspace.UpdateFolder(workspaceID, folderID, name)
	if err != nil {
		tlog.Errorw("Failed to update workspace folder",
			"workspace_id", workspaceID,
			"folder_id", folderID,
			"error", err,
		)
		return false, model.NewAppError("workspace.folder_update_failed", http.StatusInternalServerError)
	}

	a.publishProjectChange(ctx, projectChange{kind: "UPDATE_FOLDER_NAME", workspaceID: workspaceID,
		data: map[string]any{"folder_id": folderID, "data": map[string]any{"id": folderID, "name": name}}})
	return ok, nil
}

func (a *App) DeleteWorkspaceFolder(ctx context.Context, workspaceID string, folderID string, user model.User) (bool, *model.AppError) {
	if !a.ProjectWorkspaceHasPermission(user, workspaceID, model.PermissionDeleteWorkspaceFolder) {
		return false, model.NewAppError("workspace.forbidden", http.StatusForbidden)
	}

	withTables := a.ProjectWorkspaceHasPermission(user, workspaceID, model.PermissionDeleteWorkspaceTable)

	dbCtx, cancel := a.dbCtx(ctx)
	defer cancel()

	tableIDs, linkedFields, err := a.Store.Workspace.DeleteFolder(dbCtx, workspaceID, folderID, withTables)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return false, model.NewAppError("workspace.folder_not_found", http.StatusNotFound)
	case errors.Is(err, model.ErrFolderHasTables):
		return false, model.NewAppError("workspace.folder_has_tables", http.StatusBadRequest)
	case err != nil:
		tlog.Errorw("Failed to delete workspace folder",
			"workspace_id", workspaceID,
			"folder_id", folderID,
			"error", err,
		)
		return false, model.NewAppError("workspace.folder_delete_failed", http.StatusInternalServerError)
	}

	for _, tableID := range tableIDs {
		a.publishProjectChange(ctx, projectChange{kind: "DELETE_TABLE", workspaceID: workspaceID, tableID: tableID,
			data: map[string]any{"delete_id": tableID}})
	}

	a.publishLinkedFieldsGone(ctx, workspaceID, linkedFields)
	a.publishProjectChange(ctx, projectChange{kind: "DELETE_FOLDER", workspaceID: workspaceID,
		data: map[string]any{"folder_id": folderID}})
	return true, nil
}

// MoveWorkspaceItem puts a table or a folder in targetFolderID, or at the
// workspace root when targetFolderID is empty.
func (a *App) MoveWorkspaceItem(ctx context.Context, workspaceID, itemType, itemID, targetFolderID string, user model.User) *model.AppError {
	permission := model.PermissionUpdateWorkspaceTable
	switch itemType {
	case "table":
	case "folder":
		permission = model.PermissionUpdateWorkspaceFolder
	default:
		return model.NewAppError("workspace.move_item_invalid", http.StatusBadRequest)
	}

	if !a.ProjectWorkspaceHasPermission(user, workspaceID, permission) {
		return model.NewAppError("workspace.forbidden", http.StatusForbidden)
	}

	dbCtx, cancel := a.dbCtx(ctx)
	defer cancel()

	var err error
	if itemType == "table" {
		err = a.Store.Workspace.MoveTable(dbCtx, workspaceID, itemID, targetFolderID)
	} else {
		err = a.Store.Workspace.MoveFolder(dbCtx, workspaceID, itemID, targetFolderID)
	}

	switch {
	case errors.Is(err, sql.ErrNoRows):
		return model.NewAppError("workspace.move_item_not_found", http.StatusNotFound)
	case errors.Is(err, model.ErrFolderCycle):
		return model.NewAppError("workspace.folder_move_into_itself", http.StatusBadRequest)
	case err != nil:
		tlog.Errorw("Failed to move workspace item",
			"workspace_id", workspaceID,
			"item_type", itemType,
			"item_id", itemID,
			"target_folder_id", targetFolderID,
			"error", err,
		)
		return model.NewAppError("workspace.move_item_failed", http.StatusInternalServerError)
	}

	change := projectChange{
		kind:        "MOVE_ITEM",
		workspaceID: workspaceID,
		data: map[string]any{
			"item_id":          itemID,
			"item_type":        itemType,
			"target_folder_id": targetFolderID,
		},
	}
	if itemType == "table" {
		change.tableID = itemID
	}

	a.publishProjectChange(ctx, change)
	return nil
}

func (a *App) DeleteWorkspaceMember(workspaceID string, memberID string, user model.User) (bool, *model.AppError) {
	if !a.ProjectWorkspaceHasPermission(user, workspaceID, model.PermissionDeleteWorkspaceMember) {
		return false, model.NewAppError("workspace.forbidden", http.StatusForbidden)
	}

	return a.removeProjectWorkspaceMember(workspaceID, memberID)
}

func (a *App) removeProjectWorkspaceMember(workspaceID, memberID string) (bool, *model.AppError) {
	// If the user is still in a group attached to this workspace, keep their assignments
	if hasGroupAccess, _ := a.Store.Workspace.UserHasAnyGroupAccess(workspaceID, memberID); hasGroupAccess {
		deleted, storeErr := a.Store.Workspace.DeleteMember(workspaceID, memberID)
		if storeErr != nil {
			tlog.Errorw("Failed to delete workspace member",
				"workspace_id", workspaceID,
				"member_id", memberID,
				"error", storeErr,
			)
			return false, model.NewAppError("workspace.member_delete_failed", http.StatusInternalServerError)
		}

		return deleted, nil
	}

	rawTables, err := a.Store.Workspace.GetAllTablesBasic(workspaceID)
	if err != nil {
		tlog.Errorw("Failed to get workspace tables for member deletion",
			"workspace_id", workspaceID,
			"member_id", memberID,
			"error", err,
		)
		return false, model.NewAppError("workspace.tables_retrieval_failed", http.StatusInternalServerError)
	}

	for _, rt := range rawTables {
		a.clearAllPersonFields(rt, memberID)
	}

	deleted, storeErr := a.Store.Workspace.DeleteMember(workspaceID, memberID)
	if storeErr != nil {
		tlog.Errorw("Failed to delete workspace member",
			"workspace_id", workspaceID,
			"member_id", memberID,
			"error", storeErr,
		)
		return false, model.NewAppError("workspace.member_delete_failed", http.StatusInternalServerError)
	}

	if !deleted {
		return false, model.NewAppError("workspace.member_not_found", http.StatusNotFound)
	}

	return true, nil
}

func filterByAssignee(items []map[string]interface{}, userID string) []map[string]interface{} {
	out := items[:0]
	for _, item := range items {
		if assignee, ok := item["assignee"].(string); ok && assignee == userID {
			out = append(out, item)
		}
	}

	return out
}

func parseInt64(val interface{}) int64 {
	switch v := val.(type) {
	case int64:
		return v
	case int:
		return int64(v)
	case float64:
		return int64(v)
	case string:
		if parsed, err := strconv.ParseInt(v, 10, 64); err == nil {
			return parsed
		}
	}

	return 0
}

// runItemWorkers fans out fn over items using up to maxWorkers goroutines.
func runItemWorkers(items []map[string]interface{}, maxWorkers int, mu *sync.Mutex, fn func(map[string]interface{})) {
	if len(items) == 0 {
		return
	}

	type job struct{ idx int }
	jobs := make(chan job, len(items))
	workers := maxWorkers
	if workers > len(items) {
		workers = len(items)
	}

	if workers < 1 {
		workers = 1
	}

	var wg sync.WaitGroup
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs {
				item := items[j.idx]
				mu.Lock()
				fn(item)
				mu.Unlock()
			}
		}()
	}

	for i := range items {
		jobs <- job{idx: i}
	}

	close(jobs)
	wg.Wait()
}

func sortItemsByDueDate(items []map[string]interface{}) {
	sort.SliceStable(items, func(i, j int) bool {
		di := parseInt64(items[i]["due_date"])
		dj := parseInt64(items[j]["due_date"])
		if di == 0 {
			di = parseInt64(items[i]["created_at"])
		}

		if dj == 0 {
			dj = parseInt64(items[j]["created_at"])
		}

		return di < dj
	})
}

func normalizeStatusFilters(filters model.FilterPayload) model.FilterPayload {
	for i, f := range filters.FlatFilters {
		if f.Field == "status" && f.Value != "" && len(f.Values) == 0 {
			var ids []string
			if err := json.Unmarshal([]byte(f.Value), &ids); err == nil {
				filters.FlatFilters[i].Values = ids
			}
		}
	}

	for gi, group := range filters.Groups {
		for fi, f := range group.Filters {
			if f.Field == "status" && f.Value != "" && len(f.Values) == 0 {
				var ids []string
				if err := json.Unmarshal([]byte(f.Value), &ids); err == nil {
					filters.Groups[gi].Filters[fi].Values = ids
				}
			}
		}
	}

	return filters
}

// assignedOnlyFilter limits rows to the user's own tasks when their roles make
// them assigned-only for the table. The column is unqualified because some
// callers query the table without the "main" alias.
func (a *App) assignedOnlyFilter(user model.User, workspaceID, tableID string) model.SQLFilter {
	if !a.resolveAccess(a.buildAccessInput(user, workspaceID)).AssignedOnly(tableID) {
		return model.SQLFilter{}
	}
	// A workspace-wide rule names every table, but an option table, such as
	// the statuses a task picks from, has no one assigned to its rows.
	if isOptions, err := a.Store.Workspace.IsTableSingleSelect(tableID); err == nil && isOptions {
		return model.SQLFilter{}
	}

	return model.SQLFilter{SQL: "`assignee` = ?", Args: []any{user.ID}}
}

// taskFilterSQL turns a filter payload from the filter builder into SQL for
// the table, or into an empty filter when it holds no conditions.
func (a *App) taskFilterSQL(tableID, tableName string, filters model.FilterPayload, user model.User) (model.SQLFilter, error) {
	if len(filters.Groups) == 0 && len(filters.FlatFilters) == 0 {
		return model.SQLFilter{}, nil
	}

	colTypes, err := a.Store.Workspace.GetTableColumnTypes(tableName)
	if err != nil {
		return model.SQLFilter{}, err
	}

	headers := a.buildTableHeaders(tableID, colTypes, false, true)

	return a.filterSQLForHeaders(tableID, headers, filters, a.filterLocation(filters, user)), nil
}

// filterLocation is the time zone a filter's dates are read in: the one the
// request names, or else the user's.
func (a *App) filterLocation(filters model.FilterPayload, user model.User) *time.Location {
	timezone := filters.Timezone
	if timezone == "" {
		timezone = a.resolveUserTimezone(user.ID)
	}

	loc, err := time.LoadLocation(timezone)
	if err != nil || timezone == "" {
		return time.UTC
	}

	return loc
}

// filterSQLForHeaders turns a filter payload into SQL for a table whose
// headers are already loaded, resolving the tables its link fields point to.
func (a *App) filterSQLForHeaders(tableID string, headers []model.WorkspaceHeaders, filters model.FilterPayload, loc *time.Location) model.SQLFilter {
	if len(filters.Groups) == 0 && len(filters.FlatFilters) == 0 {
		return model.SQLFilter{}
	}

	filters = normalizeStatusFilters(filters)

	tableNames, err := a.Store.Workspace.GetTableNamesByIDs(collectLinkedTableIDs(headers))
	if err != nil {
		tlog.Warnw("Failed to resolve linked table names for filter", "table_id", tableID, "error", err)
	}

	return a.Store.Workspace.BuildFilterSQLInline(filters, headers, loc, tableNames)
}

func (a *App) openTasksFilter(tableID string) model.SQLFilter {
	statusMap, _ := a.Store.Workspace.GetStatusTypeMap(tableID)
	var doneIDs []any
	for id, st := range statusMap {
		if st == "Done" || st == "Closed" {
			doneIDs = append(doneIDs, id)
		}
	}

	if len(doneIDs) == 0 {
		return model.SQLFilter{}
	}

	return model.SQLFilter{
		SQL:  "(main.status IS NULL OR main.status NOT IN (" + sqlPlaceholderList(len(doneIDs)) + "))",
		Args: doneIDs,
	}
}

func (a *App) buildTableHeaders(tableID string, colTypes []*sql.ColumnType, withLinkedName, withBothDirections bool) []model.WorkspaceHeaders {
	headers := make([]model.WorkspaceHeaders, 0, len(colTypes))
	if len(colTypes) == 0 {
		return headers
	}

	meta, err := a.Store.Workspace.GetTableHeaderMeta(tableID)
	if err != nil {
		tlog.Errorw("Failed to load table headers", "table_id", tableID, "error", err)
		return headers
	}

	for _, ct := range colTypes {
		field := model.SystemFieldData(tableID, ct.Name())
		if field == nil {
			f, ok := meta.Fields[ct.Name()]
			if !ok {
				continue
			}

			field = &f
		}

		linkedID := meta.Links[ct.Name()]
		linked := meta.LinkedTables[linkedID]
		h := model.WorkspaceHeaders{
			ID:            field.ID,
			Name:          ct.Name(),
			HeaderType:    ct.DatabaseTypeName(),
			LinkedID:      linkedID,
			SingleSelect:  linked.SingleSelect,
			HeaderUsage:   field.FieldType,
			ParentTableID: linked.ParentTableID,
			DisplayName:   field.DisplayName,
			Formula:       field.Formula,
			StatusType:    field.StatusType,
		}
		if withBothDirections {
			h.LinkBothDirections = linked.BothDirections
		}

		if withLinkedName {
			h.LinkedName = meta.LinkedFieldNames[linked.ParentTableID+"/"+field.ID]
		}

		headers = append(headers, h)
	}

	return headers
}

// hydrateLinkedAndSingleSelect enriches items in-place with linked IDs and
// single-select resolved values, reading each field once for all the items.
func (a *App) hydrateLinkedAndSingleSelect(ctx context.Context, items []map[string]interface{}, headers []model.WorkspaceHeaders) {
	if len(items) == 0 {
		return
	}

	ids := make([]string, 0, len(items))
	for _, item := range items {
		if id, _ := item["id"].(string); id != "" {
			ids = append(ids, id)
		}
	}

	rawValue := func(item map[string]interface{}, name string) string {
		switch v := item[name].(type) {
		case string:
			return v
		case []byte:
			return string(v)
		}

		return ""
	}

	for _, header := range headers {
		switch {
		case header.SingleSelect:
			seen := map[string]bool{}
			var raws []string
			for _, item := range items {
				if raw := rawValue(item, header.Name); raw != "" && !seen[raw] {
					seen[raw] = true
					raws = append(raws, raw)
				}
			}

			if len(raws) == 0 {
				continue
			}

			options := map[string]*model.TaskOrderField{}
			var err error
			for chunk := range slices.Chunk(raws, 1000) {
				var part map[string]*model.TaskOrderField
				if part, err = a.Store.Workspace.GetSingleSelectOptions(ctx, header.LinkedID, chunk); err != nil {
					break
				}

				maps.Copy(options, part)
			}

			if err != nil {
				tlog.Warnw("Failed to load single-select options", "linked_id", header.LinkedID, "error", err)
				continue
			}

			for _, item := range items {
				if raw := rawValue(item, header.Name); raw != "" {
					item[header.Name] = options[raw]
				}
			}
		case header.LinkedID != "" && len(ids) > 0:
			links := map[string][]model.LinkedItem{}
			var err error
			for chunk := range slices.Chunk(ids, 1000) {
				var part map[string][]model.LinkedItem
				if part, err = a.Store.Workspace.GetLinkedIDsBatch(ctx, header.LinkedID, chunk); err != nil {
					break
				}

				maps.Copy(links, part)
			}

			if err != nil {
				tlog.Warnw("Failed to load linked items", "linked_id", header.LinkedID, "error", err)
				continue
			}

			for _, item := range items {
				if id, _ := item["id"].(string); id != "" {
					item[header.Name] = links[id]
				}
			}
		}
	}
}

// hydrateFileAttachments loads file attachments for all file-type headers into items.
func (a *App) hydrateFileAttachments(ctx context.Context, items []map[string]interface{}, headers []model.WorkspaceHeaders, tableID string) {
	taskIDs := make([]string, 0, len(items))
	for _, item := range items {
		if id, ok := item["id"].(string); ok {
			taskIDs = append(taskIDs, id)
		}
	}

	for _, h := range headers {
		if h.HeaderUsage != "file" {
			continue
		}

		files, err := a.Store.Workspace.GetAttachmentsForTasks(ctx, tableID, h.ID, taskIDs)
		if err != nil {
			tlog.Warnw("Failed to load file attachments", "table_id", tableID, "field_id", h.ID, "error", err)
		}

		for _, item := range items {
			id, ok := item["id"].(string)
			if !ok {
				continue
			}

			if f := files[id]; f != nil {
				item[h.Name] = f
			} else {
				item[h.Name] = []map[string]interface{}{}
			}
		}
	}
}

func (a *App) getTableItems(tableName, tableID string) ([]map[string]interface{}, []model.WorkspaceHeaders, error) {
	items, err := a.Store.Workspace.GetTableRows(tableName)
	if err != nil {
		return nil, nil, err
	}

	colTypes, err := a.Store.Workspace.GetTableColumnTypes(tableName)
	if err != nil {
		return nil, nil, err
	}

	headers := a.buildTableHeaders(tableID, colTypes, true, true)

	var mu sync.Mutex

	for _, header := range headers {
		if header.SingleSelect || header.LinkedID == "" {
			continue
		}

		h := header
		runItemWorkers(items, 16, &mu, func(item map[string]interface{}) {
			id, _ := item["id"].(string)
			if ids, err := a.Store.Workspace.GetLinkedIDs(h.LinkedID, id); err == nil {
				item[h.Name] = ids
			}
		})
	}

	for _, header := range headers {
		if !header.SingleSelect {
			continue
		}

		h := header
		runItemWorkers(items, 16, &mu, func(item map[string]interface{}) {
			nameVal, ok := item[h.Name].(string)
			if !ok {
				return
			}

			if linked, err := a.Store.Workspace.GetSingleSelectValues(h.LinkedID, nameVal); err == nil {
				item[h.Name] = linked
			}
		})
	}

	a.hydrateFileAttachments(context.Background(), items, headers, tableID)

	sort.SliceStable(items, func(i, j int) bool {
		var di, dj int64
		if v, ok := items[i]["due_date"]; ok && v != nil {
			di = parseInt64(v)
		}

		if v, ok := items[j]["due_date"]; ok && v != nil {
			dj = parseInt64(v)
		}

		if di == 0 {
			if v, ok := items[i]["created_at"]; ok && v != nil {
				di = parseInt64(v)
			}
		}

		if dj == 0 {
			if v, ok := items[j]["created_at"]; ok && v != nil {
				dj = parseInt64(v)
			}
		}

		return di < dj
	})

	return items, headers, nil
}

func (a *App) resolveUserTimezone(userID string) string {
	tz, err := a.Store.Workspace.GetUserTimezone(userID)
	if err != nil || tz == nil {
		return "UTC"
	}

	if tz.UseAutomaticTimezone {
		return tz.AutomaticTimezone
	}

	return tz.ManualTimezone
}

func (a *App) GetAllStatusIndex(workspaceIDs []string) (*model.StatusIndex, *model.AppError) {
	tables, err := a.Store.Workspace.GetTableMetas(workspaceIDs)
	if err != nil {
		tlog.Errorw("Failed to retrieve workspace tables for status index", "error", err)
		return nil, model.NewAppError("workspace.status_index_retrieval_failed", http.StatusInternalServerError)
	}

	if len(tables) == 0 {
		return &model.StatusIndex{
			NameToIDs: map[string][]string{},
			FlatList:  []model.StatusField{},
		}, nil
	}

	var linkedIDSet sync.Map
	var wg sync.WaitGroup

	for _, table := range tables {
		wg.Add(1)
		go func(tbl model.WorkspaceTable) {
			defer wg.Done()
			// Only need column metadata. Do NOT call getTableItems (it loads all rows).
			colTypes, err := a.Store.Workspace.GetTableColumnTypes(tbl.Name)
			if err != nil {
				return
			}

			headers := a.buildTableHeaders(tbl.ID, colTypes, true, true)
			for _, header := range headers {
				if strings.EqualFold(header.Name, "status") && header.LinkedID != "" {
					linkedIDSet.Store(header.LinkedID, struct{}{})
				}
			}
		}(table)
	}

	wg.Wait()

	var linkedIDs []string
	linkedIDSet.Range(func(key, value any) bool {
		linkedIDs = append(linkedIDs, key.(string))
		return true
	})

	if len(linkedIDs) == 0 {
		return &model.StatusIndex{
			NameToIDs: map[string][]string{},
			FlatList:  []model.StatusField{},
		}, nil
	}

	linkedTables, err := a.Store.Workspace.GetTablesByIDs(linkedIDs)
	if err != nil {
		tlog.Errorw("Failed to retrieve linked tables for status index", "error", err)
		return nil, model.NewAppError("workspace.status_index_retrieval_failed", http.StatusInternalServerError)
	}

	var nameToIDs sync.Map
	var flatListMu sync.Mutex
	var flatList []model.StatusField
	wg = sync.WaitGroup{}

	for _, t := range linkedTables {
		wg.Add(1)
		go func(tbl model.WorkspaceTable) {
			defer wg.Done()
			items, _, err := a.getTableItems(tbl.Name, tbl.ID)
			if err != nil {
				return
			}

			var localList []model.StatusField
			for _, item := range items {
				idRaw, ok := item["id"]
				if !ok {
					continue
				}

				nameRaw, ok := item["name"]
				if !ok {
					continue
				}

				idStr, ok1 := idRaw.(string)
				nameStr, ok2 := nameRaw.(string)
				if !ok1 || !ok2 {
					continue
				}

				statusType := ""
				if typeRaw, ok := item["status_type"]; ok {
					if typeStr, ok := typeRaw.(string); ok {
						statusType = typeStr
					}
				}

				status := model.StatusField{ID: idStr, Name: nameStr, StatusType: statusType}
				localList = append(localList, status)

				existing, _ := nameToIDs.Load(nameStr)
				if existing != nil {
					list := existing.([]string)
					nameToIDs.Store(nameStr, append(list, idStr))
				} else {
					nameToIDs.Store(nameStr, []string{idStr})
				}
			}

			flatListMu.Lock()
			flatList = append(flatList, localList...)
			flatListMu.Unlock()
		}(t)
	}

	wg.Wait()

	finalMap := make(map[string][]string)
	nameToIDs.Range(func(key, value any) bool {
		finalMap[key.(string)] = value.([]string)
		return true
	})

	return &model.StatusIndex{NameToIDs: finalMap, FlatList: flatList}, nil
}

// GetAllWorkspaceTasks reads tasks across every workspace named, or every
// workspace when none is, without checking membership, so it is limited to
// system admins.
func (a *App) GetAllWorkspaceTasks(ctx context.Context, workspaceIDs []string, user model.User, page, limit int, count bool, filters model.FilterPayload) (*model.AllWorkspaceTasksPage, *model.AppError) {
	if user.Role != model.SystemAdminRoleId {
		return nil, model.NewAppError("workspace.forbidden", http.StatusForbidden)
	}

	ctx, cancel := a.dbCtx(ctx)
	defer cancel()
	if page < 1 {
		page = 1
	}

	if limit <= 0 {
		limit = 100
	}

	tables, err := a.Store.Workspace.GetTableMetas(workspaceIDs)
	if err != nil {
		tlog.Errorw("Failed to retrieve workspace tables", "user_id", user.ID, "error", err)
		return nil, model.NewAppError("workspace.tasks_retrieval_failed", http.StatusInternalServerError)
	}

	allFilters := a.taskReportFilters(user.ID)
	if len(tables) == 0 {
		return &model.AllWorkspaceTasksPage{
			DataBase:  []map[string]interface{}{},
			Total:     model.NewInt(0),
			RootTotal: model.NewInt(0),
			Filters:   allFilters,
		}, nil
	}

	filters, appErr := a.widenStatusFilters(filters, workspaceIDs)
	if appErr != nil {
		return nil, appErr
	}

	failed := func(msg string, err error) (*model.AllWorkspaceTasksPage, *model.AppError) {
		tlog.Errorw("GetAllWorkspaceTasks: "+msg, "user_id", user.ID, "error", err)
		return nil, model.NewAppError("workspace.tasks_retrieval_failed", http.StatusInternalServerError)
	}

	type reportTable struct {
		table   model.WorkspaceTable
		headers []model.WorkspaceHeaders
		filter  model.SQLFilter
	}

	byID := make(map[string]reportTable, len(tables))
	sources := make([]model.TaskSource, 0, len(tables))
	var allHeaders []model.WorkspaceHeaders
	seenHeaders := map[string]bool{}
	loc := a.filterLocation(filters, user)

	for _, tbl := range tables {
		colTypes, err := a.Store.Workspace.GetTableColumnTypes(tbl.Name)
		if err != nil {
			tlog.Warnw("Skipping a table in the task report", "table_id", tbl.ID, "error", err)
			continue
		}

		headers := a.buildTableHeaders(tbl.ID, colTypes, true, true)
		filter := a.filterSQLForHeaders(tbl.ID, headers, filters, loc)

		columns := make(map[string]bool, len(colTypes))
		for _, ct := range colTypes {
			columns[ct.Name()] = true
		}

		byID[tbl.ID] = reportTable{table: tbl, headers: headers, filter: filter}
		sources = append(sources, model.TaskSource{TableID: tbl.ID, TableName: tbl.Name, Filter: filter, Columns: columns})
		for _, h := range headers {
			if !seenHeaders[h.Name] {
				seenHeaders[h.Name] = true
				allHeaders = append(allHeaders, h)
			}
		}
	}

	result := &model.AllWorkspaceTasksPage{
		DataBase: []map[string]interface{}{},
		Headers:  allHeaders,
		Filters:  allFilters,
	}

	offset := (page - 1) * limit

	if count {
		counts, err := a.Store.Workspace.CountTasksAcrossTables(ctx, sources, []model.SQLFilter{
			{SQL: model.TopLevelTaskSQL},
			{SQL: "1 = 1"},
		})
		if err != nil {
			return failed("failed to count tasks", err)
		}

		result.Total = model.NewInt(counts[0] + counts[1])
		result.RootTotal = model.NewInt(counts[0])
		if offset >= *result.Total {
			return result, nil
		}
	}

	keys, err := a.Store.Workspace.GetTaskPageKeys(ctx, sources, filters.Sort, limit, offset)
	if err != nil {
		return failed("failed to page tasks", err)
	}

	idsByTable := map[string][]any{}
	for _, k := range keys {
		idsByTable[k.TableID] = append(idsByTable[k.TableID], k.ID)
	}

	workspaceByTable := make(map[string]string, len(idsByTable))
	for id := range idsByTable {
		workspaceByTable[id] = byID[id].table.WorkspaceID
	}

	mainViewIDs, _ := a.Store.Workspace.GetMainViewIDs(workspaceByTable)

	rowsByKey := make(map[model.TaskKey]map[string]interface{}, len(keys))
	for tableID, ids := range idsByTable {
		rt := byID[tableID]
		rows, err := a.Store.Workspace.GetTableRowsFiltered(ctx, rt.table.Name,
			model.SQLFilter{SQL: "main.id IN (" + sqlPlaceholderList(len(ids)) + ")", Args: ids})
		if err != nil {
			return failed("failed to load tasks", err)
		}

		a.hydrateLinkedAndSingleSelect(ctx, rows, rt.headers)
		linkToTable := map[string]string{"workspace_id": rt.table.WorkspaceID, "table_id": tableID, "view_id": mainViewIDs[tableID]}
		for _, row := range rows {
			row["workspace_name"] = rt.table.WorkspaceName
			row["table_name"] = exportTableName(rt.table)
			row["link_to_table"] = linkToTable
			id, _ := row["id"].(string)
			rowsByKey[model.TaskKey{TableID: tableID, ID: id}] = row
		}
	}

	for _, k := range keys {
		if row, ok := rowsByKey[k]; ok {
			result.DataBase = append(result.DataBase, row)
		}
	}

	return result, nil
}

// widenStatusFilters makes each status filter match every status that shares
// a name with the ones it names, across the given workspaces. Each table has
// its own status options, so a status picked in one would otherwise match
// that table alone.
func (a *App) widenStatusFilters(filters model.FilterPayload, workspaceIDs []string) (model.FilterPayload, *model.AppError) {
	hasStatus := false
	for _, f := range filters.FlatFilters {
		hasStatus = hasStatus || f.Field == "status"
	}

	for _, g := range filters.Groups {
		for _, f := range g.Filters {
			hasStatus = hasStatus || f.Field == "status"
		}
	}

	if !hasStatus {
		return filters, nil
	}

	index, appErr := a.GetAllStatusIndex(workspaceIDs)
	if appErr != nil {
		return filters, appErr
	}

	nameByID := make(map[string]string, len(index.FlatList))
	for _, s := range index.FlatList {
		nameByID[s.ID] = s.Name
	}

	widen := func(f model.Filter) model.Filter {
		if f.Field != "status" || (f.Operator != "is" && f.Operator != "is_not") {
			return f
		}

		ids := f.Values
		if len(ids) == 0 && f.Value != "" {
			ids = []string{f.Value}
		}

		seen := map[string]bool{}
		var widened []string
		for _, id := range ids {
			matches := []string{id}
			if name, ok := nameByID[id]; ok {
				matches = index.NameToIDs[name]
			}

			for _, m := range matches {
				if !seen[m] {
					seen[m] = true
					widened = append(widened, m)
				}
			}
		}

		f.Values = widened
		return f
	}

	out := filters
	out.FlatFilters = make([]model.Filter, len(filters.FlatFilters))
	for i, f := range filters.FlatFilters {
		out.FlatFilters[i] = widen(f)
	}

	out.Groups = make([]model.FilterGroup, len(filters.Groups))
	for i, g := range filters.Groups {
		g.Filters = slices.Clone(g.Filters)
		for j, f := range g.Filters {
			g.Filters[j] = widen(f)
		}

		out.Groups[i] = g
	}

	return out, nil
}

func sqlPlaceholderList(n int) string {
	return strings.TrimSuffix(strings.Repeat("?,", n), ",")
}

// GetTaskReportFilters returns the user's saved task report filters with no
// tasks, for a report that loads its tasks once it knows its filter.
func (a *App) GetTaskReportFilters(user model.User) (*model.AllWorkspaceTasksPage, *model.AppError) {
	if user.Role != model.SystemAdminRoleId {
		return nil, model.NewAppError("workspace.forbidden", http.StatusForbidden)
	}

	return &model.AllWorkspaceTasksPage{
		DataBase: []map[string]interface{}{},
		Filters:  a.taskReportFilters(user.ID),
	}, nil
}

func (a *App) taskReportFilters(userID string) []model.SavedFilterWithPayload {
	saved, _ := a.Store.Workspace.GetTableFilters("all", "detailed-task-report", userID)
	var filters []model.SavedFilterWithPayload
	seen := map[string]bool{}
	for _, f := range saved {
		if seen[f.FilterID] {
			continue
		}

		seen[f.FilterID] = true
		var payload model.FilterPayload
		if json.Unmarshal([]byte(f.FilterSettings), &payload) == nil {
			f.Filters = &payload
		}

		filters = append(filters, f)
	}

	return filters
}

// assignedSection is a due-date group on the assigned tasks page.
type assignedSection struct {
	key    string
	filter model.SQLFilter
}

// assignedSections splits tasks by due date in the user's time zone, in the
// order the page shows them.
func assignedSections(now time.Time) []assignedSection {
	y, m, d := now.Date()
	todayStart := time.Date(y, m, d, 0, 0, 0, 0, now.Location()).Unix()
	tomorrow := time.Date(y, m, d+1, 0, 0, 0, 0, now.Location()).Unix()
	weekEnd := time.Date(y, m, d+8, 0, 0, 0, 0, now.Location()).Unix()
	return []assignedSection{
		{"overdue", model.SQLFilter{SQL: "(main.due_date > 0 AND main.due_date < ?)", Args: []any{todayStart}}},
		{"today", model.SQLFilter{SQL: "(main.due_date >= ? AND main.due_date < ?)", Args: []any{todayStart, tomorrow}}},
		{"thisWeek", model.SQLFilter{SQL: "(main.due_date >= ? AND main.due_date < ?)", Args: []any{tomorrow, weekEnd}}},
		{"upcoming", model.SQLFilter{SQL: "main.due_date >= ?", Args: []any{weekEnd}}},
		{"noDueDate", model.SQLFilter{SQL: "(main.due_date IS NULL OR main.due_date = 0)"}},
	}
}

// GetAssignedToMe returns the open tasks assigned to the user across their
// workspaces, grouped by due date: each group's size and first page of tasks,
// or, when section is set, that group's size and the page after the cursor.
func (a *App) GetAssignedToMe(ctx context.Context, userID, timezone, section, cursor string, limit int) (*model.AssignedToMePage, *model.AppError) {
	var after *taskCursor
	if section != "" {
		var appErr *model.AppError
		if after, appErr = decodeTaskCursor(cursor); appErr != nil {
			return nil, appErr
		}
	}

	ctx, cancel := a.dbCtx(ctx)
	defer cancel()
	page := &model.AssignedToMePage{Sections: []model.AssignedSection{}, Tables: map[string]model.AssignedTable{}}

	tables, err := a.Store.Workspace.GetTableMetasForUser(userID)
	if err != nil {
		tlog.Errorw("Failed to retrieve workspace tables for assigned tasks", "user_id", userID, "error", err)
		return nil, model.NewAppError("workspace.assigned_tasks_retrieval_failed", http.StatusInternalServerError)
	}

	if timezone == "" {
		timezone = a.resolveUserTimezone(userID)
	}

	loc, lerr := time.LoadLocation(timezone)
	if lerr != nil || timezone == "" {
		loc = time.UTC
	}

	sections := assignedSections(time.Now().In(loc))
	if section != "" {
		sections = slices.DeleteFunc(sections, func(s assignedSection) bool { return s.key != section })
	}

	sources := make([]model.TaskSource, 0, len(tables))
	tableByID := make(map[string]model.WorkspaceTable, len(tables))
	for _, t := range tables {
		filter := model.SQLFilter{SQL: "main.assignee = ?", Args: []any{userID}}.And(a.openTasksFilter(t.ID))
		sources = append(sources, model.TaskSource{TableID: t.ID, TableName: t.Name, Filter: filter})
		tableByID[t.ID] = t
	}

	filters := make([]model.SQLFilter, len(sections))
	for i, s := range sections {
		filters[i] = s.filter
	}

	counts, err := a.Store.Workspace.CountTasksAcrossTables(ctx, sources, filters)
	if err != nil {
		if !clientLeft(ctx) {
			tlog.Errorw("Failed to count assigned tasks", "user_id", userID, "error", err)
		}

		return nil, model.NewAppError("workspace.assigned_tasks_retrieval_failed", http.StatusInternalServerError)
	}

	var shown []map[string]interface{}
	for i, s := range sections {
		result := model.AssignedSection{Key: s.key, Total: counts[i], Tasks: []map[string]interface{}{}}

		if counts[i] > 0 {
			inSection := make([]model.TaskSource, len(sources))
			for j, src := range sources {
				src.Filter = src.Filter.And(s.filter)
				if after != nil {
					src.Filter = src.Filter.And(after.after("main.due_date"))
				}

				inSection[j] = src
			}

			rows, err := a.Store.Workspace.GetTasksAcrossTables(ctx, inSection, limit+1)
			if err != nil {
				if !clientLeft(ctx) {
					tlog.Errorw("Failed to retrieve assigned tasks", "user_id", userID, "section", s.key, "error", err)
				}

				return nil, model.NewAppError("workspace.assigned_tasks_retrieval_failed", http.StatusInternalServerError)
			}

			if len(rows) > limit {
				rows = rows[:limit]
				last := rows[limit-1]
				result.Next = taskCursor{ID: fmt.Sprint(last["id"]), Value: cursorValue(last["due_date"])}.encode()
			}

			result.Tasks = append(result.Tasks, rows...)
		}

		shown = append(shown, result.Tasks...)
		page.Sections = append(page.Sections, result)
	}

	a.hydrateAssignedStatuses(shown)

	workspaceByTable := map[string]string{}
	for _, row := range shown {
		if id, _ := row["table_id"].(string); id != "" {
			workspaceByTable[id] = tableByID[id].WorkspaceID
		}
	}

	viewIDs, err := a.Store.Workspace.GetMainViewIDs(workspaceByTable)
	if err != nil {
		tlog.Warnw("Failed to load main views for assigned tasks", "user_id", userID, "error", err)
	}

	for id := range workspaceByTable {
		t := tableByID[id]
		name := t.DisplayName.String
		if !t.DisplayName.Valid || name == "" {
			name = strings.TrimPrefix(t.Name, t.Prefix)
		}

		page.Tables[id] = model.AssignedTable{WorkspaceID: t.WorkspaceID, WorkspaceName: t.WorkspaceName, TableName: name, ViewID: viewIDs[id]}
	}

	return page, nil
}

// hydrateAssignedStatuses replaces each task's status id with the status
// option, reading each option once however many tasks share it.
func (a *App) hydrateAssignedStatuses(tasks []map[string]interface{}) {
	statusTables := map[string]string{}
	options := map[string]interface{}{}
	for _, task := range tasks {
		tableID, _ := task["table_id"].(string)
		raw, _ := task["status"].(string)
		if tableID == "" || raw == "" {
			continue
		}

		statusTable, ok := statusTables[tableID]
		if !ok {
			statusTable = a.Store.Workspace.GetLinkedTableIDByName(tableID, "status")
			statusTables[tableID] = statusTable
		}

		if statusTable == "" {
			continue
		}

		key := statusTable + "/" + raw
		option, ok := options[key]
		if !ok {
			if linked, err := a.Store.Workspace.GetSingleSelectValues(statusTable, raw); err == nil {
				option = linked
			}

			options[key] = option
		}

		if option != nil {
			task["status"] = option
		}
	}
}

func (a *App) CreateView(ctx context.Context, workspaceID string, tableID string, order string, name string, viewType string, isPublic bool, user model.User, groupField string, parentTableID string) (*model.WorkspaceView, *model.AppError) {
	if !a.ProjectWorkspaceHasPermission(user, workspaceID, model.PermissionCreateWorkspaceView) &&
		!a.ProjectWorkspaceTableHasPermission(user, workspaceID, tableID, "manage_views") {
		return nil, model.NewAppError("workspace.forbidden", http.StatusForbidden)
	}

	exists, err := a.Store.Workspace.ViewExistsByName(tableID, name)
	if err != nil {
		tlog.Errorw("Failed to check view name",
			"workspace_id", workspaceID,
			"table_id", tableID,
			"error", err,
		)
		return nil, model.NewAppError("workspace.view_create_failed", http.StatusInternalServerError)
	}

	if exists {
		return nil, model.NewAppError("workspace.view_already_exists", http.StatusConflict)
	}

	switch viewType {
	case "kanban":
		builtOrder, appErr := a.buildKanbanViewOrder(ctx, tableID, groupField)
		if appErr != nil {
			return nil, appErr
		}

		order = builtOrder
	case "grid":
		builtOrder, appErr := a.buildGridViewOrder(workspaceID, tableID)
		if appErr != nil {
			return nil, appErr
		}

		order = builtOrder
	}

	view, err := a.Store.Workspace.CreateView(workspaceID, tableID, order, name, viewType, user.ID, parentTableID, model.NewID())
	if err != nil {
		tlog.Errorw("Failed to create view",
			"workspace_id", workspaceID,
			"table_id", tableID,
			"error", err,
		)
		return nil, model.NewAppError("workspace.view_create_failed", http.StatusInternalServerError)
	}

	if !isPublic {
		if setErr := a.Store.Workspace.UpdateViewPublic(view.ID, workspaceID, tableID, false); setErr != nil {
			tlog.Errorw("Failed to set view visibility", "view_id", view.ID, "error", setErr)
		}
	}

	view.IsPublic = isPublic
	kind := "VIEW_CREATED"
	if viewType == "grid" {
		kind = "CREATE_GRID_VIEW"
	}

	a.publishViewChange(ctx, view.ID, projectChange{kind: kind, workspaceID: workspaceID, tableID: tableID,
		data: map[string]any{"data": view}})

	return view, nil
}

// resolveTaskFieldValue turns what a request sent into what the column
// stores. Only a checkbox's true and false need turning; the database reads
// numbers from text itself, and text is kept as typed, so 007 stays 007.
func resolveTaskFieldValue(value string, isRelationship bool, fieldType string) interface{} {
	isNull := value == "" || value == "null" || value == "NaN" || value == "NULL"
	if isNull {
		if isRelationship && value == "" {
			return ""
		}

		return nil
	}

	if isRelationship {
		return value
	}

	if fieldType == "bool" && (value == "true" || value == "false") {
		return value == "true"
	}

	return value
}

// safeLink reports whether a link may be stored and shown: http, https and
// mailto, or one without a scheme. Any other scheme, javascript: above all,
// would run in the browser of whoever clicks it.
func safeLink(link string) bool {
	s := strings.ToLower(strings.TrimSpace(link))
	for _, scheme := range []string{"http://", "https://", "mailto:"} {
		if strings.HasPrefix(s, scheme) {
			return true
		}
	}

	i := strings.IndexAny(s, ":/?#")
	return i == -1 || s[i] != ':'
}

// Hydration is skipped intentionally: the frontend resolves single-select options
// from the nav data_base, avoiding 40k+ N+1 queries for large tables.
func (a *App) GetTasksByDateRange(ctx context.Context, workspaceID, tableID string, from, to int64, filters model.FilterPayload, user model.User) (*model.WorkspaceTable, *model.AppError) {
	if !a.ProjectWorkspaceHasPermission(user, workspaceID, model.PermissionViewGantt) {
		return nil, model.NewAppError("workspace.forbidden", http.StatusForbidden)
	}

	ctx, cancel := a.dbCtx(ctx)
	defer cancel()

	tableName, err := a.Store.Workspace.GetTableName(tableID)
	if err != nil {
		tlog.Errorw("Failed to retrieve table name",
			"workspace_id", workspaceID,
			"table_id", tableID,
			"error", err,
		)
		return nil, model.NewAppError("workspace.date_range_failed", http.StatusInternalServerError)
	}

	filter, err := a.taskFilterSQL(tableID, tableName, filters, user)
	if err != nil {
		tlog.Errorw("Failed to retrieve table column types for gantt filter",
			"table_id", tableID, "error", err,
		)
		return nil, model.NewAppError("workspace.date_range_failed", http.StatusInternalServerError)
	}

	filter = filter.And(a.assignedOnlyFilter(user, workspaceID, tableID))
	items, total, err := a.Store.Workspace.GetTasksByDateRange(ctx, tableName, from, to, filter, GanttTaskLimit)
	if err != nil {
		tlog.Errorw("Failed to retrieve tasks by date range",
			"workspace_id", workspaceID,
			"table_id", tableID,
			"error", err,
		)
		return nil, model.NewAppError("workspace.date_range_failed", http.StatusInternalServerError)
	}

	headers := []model.WorkspaceHeaders{
		{Name: "id"},
		{Name: "name"},
		{Name: "start_date"},
		{Name: "due_date"},
	}

	return &model.WorkspaceTable{
		ID:       tableID,
		DataBase: items,
		Headers:  headers,
		Total:    total,
	}, nil
}

// GanttTaskLimit is how many tasks, nearest today, one Gantt request returns.
// The chart shows no more than this, and says so when there are more.
const GanttTaskLimit = 500

func (a *App) GetTasksByCalendarRange(ctx context.Context, workspaceID, tableID string, from, to int64, filters model.FilterPayload, user model.User) (*model.WorkspaceTable, *model.AppError) {
	if !a.ProjectWorkspaceHasPermission(user, workspaceID, model.PermissionViewCalendar) {
		return nil, model.NewAppError("workspace.forbidden", http.StatusForbidden)
	}

	ctx, cancel := a.dbCtx(ctx)
	defer cancel()

	tableName, err := a.Store.Workspace.GetTableName(tableID)
	if err != nil {
		tlog.Errorw("Failed to retrieve table name", "workspace_id", workspaceID, "table_id", tableID, "error", err)
		return nil, model.NewAppError("workspace.calendar_range_failed", http.StatusInternalServerError)
	}

	filter, err := a.taskFilterSQL(tableID, tableName, filters, user)
	if err != nil {
		tlog.Errorw("Failed to retrieve table column types for calendar filter", "table_id", tableID, "error", err)
		return nil, model.NewAppError("workspace.calendar_range_failed", http.StatusInternalServerError)
	}

	filter = filter.And(a.assignedOnlyFilter(user, workspaceID, tableID))
	items, err := a.Store.Workspace.GetTasksByCalendarRange(ctx, tableName, from, to, filter)
	if err != nil {
		if !clientLeft(ctx) {
			tlog.Errorw("Failed to retrieve tasks by calendar range", "workspace_id", workspaceID, "table_id", tableID, "error", err)
		}

		return nil, model.NewAppError("workspace.calendar_range_failed", http.StatusInternalServerError)
	}

	headers := []model.WorkspaceHeaders{
		{Name: "id"},
		{Name: "name"},
		{Name: "start_date"},
		{Name: "due_date"},
	}

	return &model.WorkspaceTable{ID: tableID, DataBase: items, Headers: headers}, nil
}

// maxLinkedRecordIDs caps how many rows one request may look up by id.
const maxLinkedRecordIDs = 200

// GetLinkedRecordsLite returns one page of a linked table's rows as id and
// name, for a picker: those matching search, or the rows with the given ids,
// so a picker can show the name of a value it has not loaded.
func (a *App) GetLinkedRecordsLite(ctx context.Context, workspaceID, tableID, search string, ids []string, limit, offset int, user model.User) ([]map[string]interface{}, *model.AppError) {
	if !a.ProjectWorkspaceHasPermission(user, workspaceID, model.PermissionViewGrid) {
		return nil, model.NewAppError("workspace.forbidden", http.StatusForbidden)
	}

	if len(ids) > maxLinkedRecordIDs {
		return nil, model.NewAppError("request.invalid", http.StatusBadRequest)
	}

	ctx, cancel := a.dbCtx(ctx)
	defer cancel()
	tableName, err := a.Store.Workspace.GetTableName(tableID)
	if err != nil {
		tlog.Errorw("GetLinkedRecordsLite: failed to get table name", "table_id", tableID, "error", err)
		return nil, model.NewAppError("workspace.linked_records_failed", http.StatusInternalServerError)
	}

	items, err := a.Store.Workspace.GetLinkedRecordsLite(ctx, tableName, search, ids, limit, offset, a.assignedOnlyFilter(user, workspaceID, tableID))
	if err != nil {
		tlog.Errorw("GetLinkedRecordsLite: failed to get records", "table_id", tableID, "error", err)
		return nil, model.NewAppError("workspace.linked_records_failed", http.StatusInternalServerError)
	}

	return items, nil
}

// kanbanSavedOrderLimit caps how many cards of a column's saved order decide
// the column's order; the rest follow by when they were created.
const kanbanSavedOrderLimit = 2000

// kanbanViewOrder is the part of a Kanban view's item_order the board is
// built from: the field it groups by and its columns, each with the cards
// in the order they were last arranged.
type kanbanViewOrder struct {
	Section string `json:"section"`
	Fields  []struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Order []struct {
			ID any `json:"id"`
		} `json:"order"`
	} `json:"fields"`
}

// GetKanbanData returns a page of a Kanban view's board: each column's size
// and first cards or, when column is set, that column's size and the cards
// after the cursor, with the subtasks of those cards.
func (a *App) GetKanbanData(ctx context.Context, workspaceID, tableID, viewID, column, cursor string, limit int, filters model.FilterPayload, user model.User) (*model.KanbanPage, *model.AppError) {
	if appErr := a.requireViewAccess(ctx, user.ID, viewID); appErr != nil {
		return nil, appErr
	}

	if !a.canReadTable(ctx, user, workspaceID, tableID) || !a.ProjectWorkspaceHasPermission(user, workspaceID, model.PermissionUpdateTask) {
		return nil, model.NewAppError("workspace.forbidden", http.StatusForbidden)
	}

	var after *taskCursor
	if column != "" {
		var appErr *model.AppError
		if after, appErr = decodeTaskCursor(cursor); appErr != nil {
			return nil, appErr
		}
	}

	ctx, cancel := a.dbCtx(ctx)
	defer cancel()
	failed := func(msg string, err error) (*model.KanbanPage, *model.AppError) {
		if !clientLeft(ctx) {
			tlog.Errorw("GetKanbanData: "+msg, "table_id", tableID, "view_id", viewID, "error", err)
		}

		return nil, model.NewAppError("workspace.kanban_data_failed", http.StatusInternalServerError)
	}

	view, err := a.Store.Workspace.GetView(ctx, viewID)
	if err != nil || view.TableID != tableID {
		return nil, model.NewAppError("workspace.view_or_field_not_found", http.StatusNotFound)
	}

	var order kanbanViewOrder
	if err := json.Unmarshal([]byte(view.TaskOrder), &order); err != nil {
		return failed("failed to read the view's order", err)
	}

	tableName, err := a.Store.Workspace.GetTableName(tableID)
	if err != nil {
		return failed("failed to get table name", err)
	}

	if hasSection, err := a.Store.Workspace.TableHasColumn(tableName, order.Section); err != nil {
		return failed("failed to check the grouped field", err)
	} else if !hasSection {
		return nil, model.NewAppError("workspace.view_or_field_not_found", http.StatusNotFound)
	}

	colTypes, err := a.Store.Workspace.GetTableColumnTypes(tableName)
	if err != nil {
		return failed("failed to get columns", err)
	}

	filter, err := a.taskFilterSQL(tableID, tableName, filters, user)
	if err != nil {
		return failed("failed to build filter", err)
	}

	filter = filter.And(a.assignedOnlyFilter(user, workspaceID, tableID))
	cards := filter.And(model.SQLFilter{SQL: kanbanCardsSQL})

	allGroups := kanbanColumnFilters(order)
	var columns []int
	var groups []model.SQLFilter
	for i, field := range order.Fields {
		if column == "" || column == field.ID {
			columns = append(columns, i)
			groups = append(groups, allGroups[i])
		}
	}

	counts, err := a.Store.Workspace.CountTasksAcrossTables(ctx, []model.TaskSource{{TableID: tableID, TableName: tableName, Filter: cards}}, groups)
	if err != nil {
		return failed("failed to count cards", err)
	}

	page := &model.KanbanPage{Columns: []model.KanbanColumnPage{}, Subtasks: []map[string]interface{}{}}
	var shown []map[string]interface{}
	for n, i := range columns {
		field := order.Fields[i]
		result := model.KanbanColumnPage{ID: field.ID, Total: counts[n], Tasks: []map[string]interface{}{}}

		if counts[n] > 0 {
			saved := make([]string, 0, min(len(field.Order), kanbanSavedOrderLimit))
			for _, o := range field.Order {
				if len(saved) == kanbanSavedOrderLimit {
					break
				}

				if id := fmt.Sprint(o.ID); o.ID != nil && id != "" {
					saved = append(saved, id)
				}
			}

			result.Tasks, result.Next, err = a.kanbanColumnPage(ctx, tableName, cards.And(groups[n]), saved, after, limit)
			if err != nil {
				return failed("failed to get cards", err)
			}
		}

		shown = append(shown, result.Tasks...)
		page.Columns = append(page.Columns, result)
	}

	if len(shown) > 0 {
		parents := make([]any, 0, len(shown))
		for _, task := range shown {
			parents = append(parents, task["id"])
		}

		subtasks, err := a.Store.Workspace.GetTableRowsFiltered(ctx, tableName,
			filter.And(model.SQLFilter{SQL: "main.parent_task_id IN (" + sqlPlaceholderList(len(parents)) + ")", Args: parents}))
		if err != nil {
			return failed("failed to get subtasks", err)
		}

		if subtasks != nil {
			page.Subtasks = subtasks
		}
	}

	rows := append(shown, page.Subtasks...)
	a.hydrateKanbanRows(ctx, tableID, colTypes, rows)
	a.restrictLinkedItems(ctx, rows, a.buildTableHeaders(tableID, colTypes, true, true), user, workspaceID)
	return page, nil
}

// kanbanColumnPage returns the cards of a column that follow the cursor, at
// most limit of them, and the cursor for the page after them. The cards in
// the column's saved order come first, in that order, then the rest by when
// they were created.
func (a *App) kanbanColumnPage(ctx context.Context, tableName string, cards model.SQLFilter, saved []string, after *taskCursor, limit int) ([]map[string]interface{}, string, error) {
	first := saved
	switch {
	case after != nil && after.Saved != nil:
		// A card dragged out of the saved order since the page before is
		// found by where it was.
		pos := slices.Index(saved, after.ID)
		if pos < 0 {
			pos = max(-1, min(*after.Saved, len(saved)-1))
		}

		first = saved[pos+1:]
		cards = cards.And(notInFilter("main.id", saved[:pos+1]))

	case after != nil:
		first = nil
		cards = cards.And(notInFilter("main.id", saved)).And(after.after("main.created_at"))
	}

	rows, err := a.Store.Workspace.GetTableRowsPage(ctx, tableName, cards, first, limit+1)
	if err != nil || len(rows) <= limit {
		return rows, "", err
	}

	rows = rows[:limit]
	last := rows[limit-1]
	next := taskCursor{ID: fmt.Sprint(last["id"])}
	if pos := slices.Index(saved, next.ID); pos >= 0 {
		next.Saved = &pos
	} else {
		next.Value = cursorValue(last["created_at"])
	}

	return rows, next.encode(), nil
}

// kanbanColumnFilters returns the condition for each of a board's columns.
// The Unassigned column, id "0" or named Unassigned, holds cards with no
// value in the grouped field or one that is none of the other columns.
func kanbanColumnFilters(order kanbanViewOrder) []model.SQLFilter {
	section := "main.`" + strings.ReplaceAll(order.Section, "`", "``") + "`"
	unassigned := func(i int) bool { return order.Fields[i].ID == "0" || order.Fields[i].Name == "Unassigned" }
	var others []any
	for i, f := range order.Fields {
		if !unassigned(i) {
			others = append(others, f.ID)
		}
	}

	filters := make([]model.SQLFilter, len(order.Fields))
	for i, f := range order.Fields {
		switch {
		case !unassigned(i):
			filters[i] = model.SQLFilter{SQL: section + " = ?", Args: []any{f.ID}}
		case len(others) == 0:
			filters[i] = model.SQLFilter{SQL: "1 = 1"}
		default:
			filters[i] = model.SQLFilter{
				SQL:  "(" + section + " IS NULL OR " + section + " = '' OR " + section + " NOT IN (" + sqlPlaceholderList(len(others)) + "))",
				Args: others,
			}
		}
	}

	return filters
}

// hydrateKanbanRows fills in the linked fields Kanban cards show and drops
// the description, which cards do not show and which can be large.
// Single-select options come from the workspace navigation instead.
func (a *App) hydrateKanbanRows(ctx context.Context, tableID string, colTypes []*sql.ColumnType, items []map[string]interface{}) {
	if len(items) > 0 {
		headers := a.buildTableHeaders(tableID, colTypes, true, true)

		taskIDs := make([]string, 0, len(items))
		for _, item := range items {
			if id, _ := item["id"].(string); id != "" {
				taskIDs = append(taskIDs, id)
			}
		}

		type linkedResult struct {
			name      string
			linkedMap map[string][]model.LinkedItem
		}

		results := make(chan linkedResult, len(headers))
		var wg sync.WaitGroup

		for _, h := range headers {
			if h.LinkedID == "" || h.SingleSelect {
				continue
			}

			if h.HeaderUsage == "assignee" || h.HeaderUsage == "default_assignee" {
				continue
			}

			wg.Add(1)
			go func(h model.WorkspaceHeaders) {
				defer wg.Done()
				lm, err := a.Store.Workspace.GetLinkedIDsBatch(ctx, h.LinkedID, taskIDs)
				if err == nil && lm != nil {
					results <- linkedResult{name: h.Name, linkedMap: lm}
				}
			}(h)
		}

		go func() { wg.Wait(); close(results) }()

		for r := range results {
			for i, item := range items {
				taskID, _ := item["id"].(string)
				if links, ok := r.linkedMap[taskID]; ok {
					items[i][r.name] = links
				}
			}
		}
	}

	for i := range items {
		delete(items[i], "description")
	}
}

func (a *App) buildKanbanViewOrder(ctx context.Context, tableID, groupField string) (string, *model.AppError) {
	// Column structure only, no tasks; tasks load lazily when the kanban view opens.
	section := groupField

	linkedTableID, err := a.Store.Workspace.GetLinkedTableID(tableID, section)
	if err != nil {
		tlog.Errorw("Failed to get linked table for kanban view",
			"table_id", tableID,
			"section", section,
			"error", err,
		)
		return "", model.NewAppError("workspace.view_create_failed", http.StatusInternalServerError)
	}

	linkedTableName, err := a.Store.Workspace.GetTableName(linkedTableID)
	if err != nil {
		tlog.Errorw("Failed to get linked table name for kanban view",
			"linked_table_id", linkedTableID,
			"error", err,
		)
		return "", model.NewAppError("workspace.view_create_failed", http.StatusInternalServerError)
	}

	statusOptions, err := a.Store.Workspace.GetKanbanStatusOptions(linkedTableName)
	if err != nil {
		tlog.Errorw("Failed to get kanban status options",
			"linked_table_name", linkedTableName,
			"error", err,
		)
		return "", model.NewAppError("workspace.view_create_failed", http.StatusInternalServerError)
	}

	fieldMap := make(map[string]map[string]interface{})
	var unassignedField map[string]interface{}

	for _, opt := range statusOptions {
		field := map[string]interface{}{
			"id":           opt.ID,
			"name":         opt.Name,
			"display_name": "",
			"width":        "300",
			"color":        opt.Color,
			"order":        []map[string]interface{}{},
			"visible":      true,
		}
		if opt.ID == "0" || opt.Name == "Unassigned" {
			unassignedField = field
			continue
		}

		fieldMap[opt.Name] = field
	}

	statusOrder := []string{"Completed", "Cancelled"}
	var fieldArray []map[string]interface{}
	for _, s := range statusOrder {
		if field, ok := fieldMap[s]; ok {
			fieldArray = append(fieldArray, field)
			delete(fieldMap, s)
		}
	}

	for _, field := range fieldMap {
		fieldArray = append(fieldArray, field)
	}

	if unassignedField == nil {
		unassignedField = map[string]interface{}{
			"id":      "0",
			"name":    "Unassigned",
			"width":   "300",
			"color":   "#000000",
			"order":   []map[string]interface{}{},
			"visible": true,
		}
	}

	fieldArray = append(fieldArray, unassignedField)

	viewOrder := map[string]interface{}{
		"section": section,
		"fields":  fieldArray,
	}
	viewOrderJSON, err := json.Marshal(viewOrder)
	if err != nil {
		tlog.Errorw("Failed to marshal kanban view order", "table_id", tableID, "error", err)
		return "", model.NewAppError("workspace.view_create_failed", http.StatusInternalServerError)
	}

	return string(viewOrderJSON), nil
}

func (a *App) UpdateView(ctx context.Context, itemID string, workspaceID string, tableID string, order string, name string, user model.User) (*model.WorkspaceView, *model.AppError) {
	if appErr := a.requireViewAccess(ctx, user.ID, itemID); appErr != nil {
		return nil, appErr
	}

	if !a.ProjectWorkspaceHasPermission(user, workspaceID, model.PermissionUpdateWorkspaceView) {
		return nil, model.NewAppError("workspace.forbidden", http.StatusForbidden)
	}

	current, err := a.Store.Workspace.GetView(ctx, itemID)
	if err != nil || current.TableID != tableID || current.WorkspaceID != workspaceID {
		return nil, model.NewAppError("workspace.view_or_field_not_found", http.StatusNotFound)
	}
	// The view's own type decides how its order is saved, not the caller.
	viewType := current.ViewType
	if name == "" {
		name = current.Name
	}

	switch {
	case order == "":
		// A rename: the view keeps its columns, filters and card order.
		err = a.Store.Workspace.UpdateViewName(itemID, workspaceID, tableID, name)
		order = current.TaskOrder
	case viewType == "kanban":
		if !json.Valid([]byte(order)) {
			return nil, model.NewAppError("request.invalid", http.StatusBadRequest)
		}

		order, err = a.Store.Workspace.ChangeViewOrder(ctx, itemID, workspaceID, tableID, func(stored string) (string, error) {
			return mergeKanbanColumns(stored, order)
		})
		if err == nil {
			err = a.Store.Workspace.UpdateViewName(itemID, workspaceID, tableID, name)
		}
	default:
		_, err = a.Store.Workspace.UpdateView(itemID, workspaceID, tableID, order, name, viewType)
	}

	if err != nil {
		tlog.Errorw("Failed to update view",
			"workspace_id", workspaceID,
			"table_id", tableID,
			"view_id", itemID,
			"error", err,
		)
		return nil, model.NewAppError("workspace.view_update_failed", http.StatusInternalServerError)
	}

	view := &model.WorkspaceView{
		ID:          itemID,
		WorkspaceID: workspaceID,
		TableID:     tableID,
		ViewType:    viewType,
		Name:        name,
		TaskOrder:   order,
	}
	if viewType == "kanban" {
		view.TaskOrder = withoutPlacedCards(order)
		a.publishKanbanColumns(ctx, workspaceID, tableID, itemID, view.TaskOrder)
	}

	return view, nil
}

// publishKanbanColumns tells a Kanban view's viewers its columns changed:
// their names, colours, widths, visibility or order. Where cards sit is
// published card by card as they move.
func (a *App) publishKanbanColumns(ctx context.Context, workspaceID, tableID, viewID, order string) {
	a.publishViewChange(ctx, viewID, projectChange{kind: "KANBAN_COLUMNS", workspaceID: workspaceID, tableID: tableID,
		data: map[string]any{"view_id": viewID, "order": order}})
}

func (a *App) SetViewPublic(ctx context.Context, user model.User, workspaceID, tableID, viewID string, isPublic bool) *model.AppError {
	if appErr := a.requireViewAccess(ctx, user.ID, viewID); appErr != nil {
		return appErr
	}

	if !a.ProjectWorkspaceHasPermission(user, workspaceID, model.PermissionUpdateWorkspaceView) {
		return model.NewAppError("workspace.forbidden", http.StatusForbidden)
	}

	before, beforeErr := a.viewAudience(ctx, viewID)
	if err := a.Store.Workspace.UpdateViewPublic(viewID, workspaceID, tableID, isPublic); err != nil {
		tlog.Errorw("Failed to set view visibility", "view_id", viewID, "error", err)
		return model.NewAppError("workspace.view_update_failed", http.StatusInternalServerError)
	}

	if beforeErr == nil {
		a.publishViewAccessChange(ctx, workspaceID, tableID, viewID, before)
	}

	return nil
}

func (a *App) GetViewShares(user model.User, workspaceID, viewID string) ([]string, *model.AppError) {
	if appErr := a.requireViewAccess(context.Background(), user.ID, viewID); appErr != nil {
		return nil, appErr
	}

	if !a.ProjectWorkspaceHasPermission(user, workspaceID, model.PermissionUpdateWorkspaceView) {
		return nil, model.NewAppError("workspace.forbidden", http.StatusForbidden)
	}

	ids, err := a.Store.Workspace.GetViewSharedUserIDs(viewID)
	if err != nil {
		tlog.Errorw("Failed to get view shares", "view_id", viewID, "error", err)
		return nil, model.NewAppError("workspace.view_shares_get_failed", http.StatusInternalServerError)
	}

	return ids, nil
}

func (a *App) SetViewShares(ctx context.Context, user model.User, workspaceID, tableID, viewID string, userIDs []string) *model.AppError {
	if appErr := a.requireViewAccess(ctx, user.ID, viewID); appErr != nil {
		return appErr
	}

	if !a.ProjectWorkspaceHasPermission(user, workspaceID, model.PermissionUpdateWorkspaceView) {
		return model.NewAppError("workspace.forbidden", http.StatusForbidden)
	}

	before, beforeErr := a.viewAudience(ctx, viewID)
	// Setting shares also marks the view as private automatically.
	if err := a.Store.Workspace.UpdateViewPublic(viewID, workspaceID, tableID, false); err != nil {
		tlog.Errorw("Failed to mark view private", "view_id", viewID, "error", err)
		return model.NewAppError("workspace.view_update_failed", http.StatusInternalServerError)
	}

	if err := a.Store.Workspace.UpdateViewShares(viewID, workspaceID, userIDs); err != nil {
		tlog.Errorw("Failed to set view shares", "view_id", viewID, "error", err)
		return model.NewAppError("workspace.view_shares_set_failed", http.StatusInternalServerError)
	}

	if beforeErr == nil {
		a.publishViewAccessChange(ctx, workspaceID, tableID, viewID, before)
	}

	return nil
}

func (a *App) UpdateWorkspaceSingleSelectName(ctx context.Context, workspaceID string, tableID string, taskID string, field string, value string, user model.User, LinkedTableID string) (*map[string]interface{}, *model.AppError) {
	if !a.CanPerformRowAction(user, workspaceID, tableID, model.PermissionUpdateTask, model.PermissionUpdateTask) {
		return nil, model.NewAppError("workspace.forbidden", http.StatusForbidden)
	}

	if appErr := a.CheckTableInWorkspace(ctx, workspaceID, LinkedTableID); appErr != nil {
		return nil, appErr
	}

	if isOptions, err := a.Store.Workspace.IsTableSingleSelect(LinkedTableID); err != nil || !isOptions {
		return nil, model.NewAppError("request.invalid", http.StatusBadRequest)
	}

	linkedTableName, err := a.Store.Workspace.GetTableName(LinkedTableID)
	if err != nil {
		tlog.Errorw("Failed to get linked table name for rename",
			"workspace_id", workspaceID,
			"table_id", tableID,
			"linked_table_id", LinkedTableID,
			"error", err,
		)
		return nil, model.NewAppError("workspace.single_select_update_failed", http.StatusInternalServerError)
	}

	oldName, err := a.Store.Workspace.GetTaskFieldValue(linkedTableName, taskID, "name")
	if err != nil {
		tlog.Errorw("Failed to get current option name for rename",
			"workspace_id", workspaceID,
			"table_id", tableID,
			"task_id", taskID,
			"error", err,
		)
		return nil, model.NewAppError("workspace.single_select_update_failed", http.StatusInternalServerError)
	}

	if err := a.Store.Workspace.RenameTaskAndKanbanSectionTx(workspaceID, tableID, taskID, field, value, oldName, linkedTableName, LinkedTableID); err != nil {
		tlog.Errorw("Failed to rename single select option",
			"workspace_id", workspaceID,
			"table_id", tableID,
			"task_id", taskID,
			"error", err,
		)
		return nil, model.NewAppError("workspace.single_select_update_failed", http.StatusInternalServerError)
	}

	result := map[string]interface{}{
		"task_id":  taskID,
		"old_name": oldName,
		"new_name": value,
	}

	a.publishProjectChange(ctx, projectChange{kind: "UPDATE_SINGLE_SELECT_HEADER_NAME", workspaceID: workspaceID, tableID: tableID,
		data: map[string]any{"data": result, "payload": map[string]any{
			"workspace_id": workspaceID, "table_id": tableID, "task_id": taskID,
			"field": field, "value": value, "linked_table_id": LinkedTableID,
		}}})
	return &result, nil
}

func (a *App) UpdateWorkspaceTask(ctx context.Context, workspaceID string, tableID string, taskID string, field string, value string, user model.User, fieldID string) (*map[string]interface{}, *model.AppError) {
	if !a.CanPerformRowAction(user, workspaceID, tableID, model.PermissionUpdateTask, model.PermissionUpdateTask) {
		isSingleSelect, _ := a.Store.Workspace.IsTableSingleSelect(tableID)
		parentID := a.Store.Workspace.GetTableParentID(tableID)
		if !isSingleSelect || parentID == "" || !a.CanPerformRowAction(user, workspaceID, parentID, model.PermissionEditFields, model.PermissionEditFields) {
			return nil, model.NewAppError("workspace.forbidden", http.StatusForbidden)
		}
	}

	if appErr := a.requireTaskAccess(user, workspaceID, tableID, taskID); appErr != nil {
		return nil, appErr
	}

	if field == "name" && value == "" {
		return nil, model.NewAppError("workspace.task_name_empty", http.StatusBadRequest)
	}

	tableName, err := a.Store.Workspace.GetTableName(tableID)
	if err != nil {
		tlog.Errorw("Failed to get table name for task update",
			"workspace_id", workspaceID,
			"table_id", tableID,
			"error", err,
		)
		return nil, model.NewAppError("workspace.task_update_failed", http.StatusInternalServerError)
	}

	if !a.writableColumn(ctx, tableName, field) {
		return nil, model.NewAppError("request.invalid", http.StatusBadRequest)
	}

	if field == "name" {
		currentName, err := a.Store.Workspace.GetTaskFieldValue(tableName, taskID, "name")
		if err != nil {
			tlog.Errorw("Failed to get current task name",
				"workspace_id", workspaceID,
				"table_id", tableID,
				"task_id", taskID,
				"error", err,
			)
			return nil, model.NewAppError("workspace.task_update_failed", http.StatusInternalServerError)
		}

		if currentName == "Completed" || currentName == "Cancelled" {
			return nil, model.NewAppError("workspace.task_update_protected", http.StatusBadRequest)
		}
	}

	linkedTableID, err := a.Store.Workspace.GetLinkedTableID(tableID, field)
	if err != nil {
		tlog.Errorw("Failed to check field relationship",
			"workspace_id", workspaceID,
			"table_id", tableID,
			"field", field,
			"error", err,
		)
		return nil, model.NewAppError("workspace.task_update_failed", http.StatusInternalServerError)
	}

	isRelationship := linkedTableID != ""

	fieldType, err := a.Store.Workspace.GetFieldType(ctx, tableID, field)
	if err != nil {
		tlog.Errorw("Failed to check the field's type", "workspace_id", workspaceID, "table_id", tableID, "field", field, "error", err)
		return nil, model.NewAppError("workspace.task_update_failed", http.StatusInternalServerError)
	}

	if fieldType == "url" && value != "" && !safeLink(value) {
		return nil, model.NewAppError("request.invalid", http.StatusBadRequest)
	}

	typedValue := resolveTaskFieldValue(value, isRelationship, fieldType)

	if value != "" {
		isPerson, err := a.isPersonField(tableID, field)
		if err != nil {
			tlog.Errorw("Failed to check the field's type", "workspace_id", workspaceID, "table_id", tableID, "field", field, "error", err)
			return nil, model.NewAppError("workspace.task_update_failed", http.StatusInternalServerError)
		}

		if isPerson {
			members, err := a.Store.Workspace.GetMemberUserIDs(ctx, workspaceID, []string{value})
			if err != nil {
				tlog.Errorw("Failed to check the assignee's membership", "workspace_id", workspaceID, "user_id", value, "error", err)
				return nil, model.NewAppError("workspace.task_update_failed", http.StatusInternalServerError)
			}

			if len(members) == 0 {
				return nil, model.NewAppError("request.invalid", http.StatusBadRequest)
			}
		}
	}

	isKanbanField := false
	if linkedTableID != "" {
		isKanbanField, err = a.Store.Workspace.IsTableSingleSelect(linkedTableID)
		if err != nil {
			tlog.Errorw("Failed to check kanban field type",
				"workspace_id", workspaceID,
				"table_id", tableID,
				"field", field,
				"error", err,
			)
			return nil, model.NewAppError("workspace.task_update_failed", http.StatusInternalServerError)
		}
	}

	if isKanbanField && value != "" && value != "0" {
		options, err := a.Store.Workspace.GetSingleSelectOptions(ctx, linkedTableID, []string{value})
		if err != nil {
			tlog.Errorw("Failed to check the chosen option", "workspace_id", workspaceID, "table_id", tableID, "field", field, "error", err)
			return nil, model.NewAppError("workspace.task_update_failed", http.StatusInternalServerError)
		}

		if options[value] == nil {
			return nil, model.NewAppError("request.invalid", http.StatusBadRequest)
		}
	}

	hasUpdatedAt, err := a.Store.Workspace.TableHasColumn(tableName, "updated_at")
	if err != nil {
		tlog.Errorw("Failed to check table schema",
			"workspace_id", workspaceID,
			"table_id", tableID,
			"error", err,
		)
		return nil, model.NewAppError("workspace.task_update_failed", http.StatusInternalServerError)
	}

	syncColors := field == "color" && typedValue != nil

	var previousAssignee string
	if field == "assignee" {
		previousAssignee, _ = a.Store.Workspace.GetTaskAssigneeID(tableName, taskID)
	}

	workspaceTask, err := a.Store.Workspace.UpdateTaskTx(workspaceID, tableID, taskID, field, value, typedValue, hasUpdatedAt, isKanbanField, syncColors)
	if err != nil {
		tlog.Errorw("Failed to update task", "workspace_id", workspaceID, "table_id", tableID, "task_id", taskID, "error", err)
		return nil, model.NewAppError("workspace.task_update_failed", http.StatusInternalServerError)
	}

	if field == "assignee" {
		a.publishTaskAssignment(ctx, workspaceID, tableID, taskID, previousAssignee)
	} else {
		a.publishTaskRow(ctx, "task_updated", workspaceID, tableID, taskID)
	}

	if field == "status" {
		var assignee *model.WorkspaceMember
		if tableName, err := a.Store.Workspace.GetTableName(tableID); err == nil {
			if assigneeUserID, err := a.Store.Workspace.GetTaskAssigneeID(tableName, taskID); err == nil && assigneeUserID != "" {
				assignee, _ = a.Store.Workspace.GetMemberByUserID(workspaceID, assigneeUserID)
			}
		}

		workspaceTaskName, err := a.Store.Workspace.GetTaskNameByID(workspaceID, tableID, taskID)
		if err != nil {
			tlog.Errorw("Failed to get task name for notification",
				"workspace_id", workspaceID,
				"table_id", tableID,
				"task_id", taskID,
				"error", err,
			)
			return nil, model.NewAppError("workspace.task_name_failed", http.StatusInternalServerError)
		}

		var statusName string
		if value != "" && value != "0" {
			linkedTableID, err := a.Store.Workspace.GetLinkedTableID(tableID, "status")
			if err != nil {
				tlog.Errorw("Failed to get status linked table for notification",
					"workspace_id", workspaceID,
					"table_id", tableID,
					"error", err,
				)
				return nil, model.NewAppError("workspace.task_status_name_failed", http.StatusInternalServerError)
			}

			statusTableName, err := a.Store.Workspace.GetTableName(linkedTableID)
			if err != nil {
				tlog.Errorw("Failed to get status table name for notification",
					"workspace_id", workspaceID,
					"table_id", tableID,
					"error", err,
				)
				return nil, model.NewAppError("workspace.task_status_name_failed", http.StatusInternalServerError)
			}

			statusName, err = a.Store.Workspace.GetStatusNameByID(statusTableName, value)
			if err != nil {
				tlog.Errorw("Failed to get status name for notification",
					"workspace_id", workspaceID,
					"table_id", tableID,
					"error", err,
				)
				return nil, model.NewAppError("workspace.task_status_name_failed", http.StatusInternalServerError)
			}
		}

		mainViewID, err := a.Store.Workspace.GetMainViewByTableID(workspaceID, tableID)
		if err != nil {
			tlog.Errorw("Failed to get main view for notification",
				"workspace_id", workspaceID,
				"table_id", tableID,
				"error", err,
			)
			return nil, model.NewAppError("workspace.task_main_view_failed", http.StatusInternalServerError)
		}

		if assignee != nil && assignee.UserID != "" {
			a.createProjectNotification(workspaceID, user, model.NOTIFICATION_PROJECT_TASK_STATUS_CHANGED, map[string]interface{}{
				"userID":     assignee.UserID,
				"taskName":   *workspaceTaskName,
				"value":      value,
				"statusName": statusName,
				"tableID":    tableID,
				"taskID":     taskID,
				"viewID":     mainViewID,
			})

			a.RecordActivity(user.ID, model.AppProjects, model.ActivityTaskStatusChanged, tableID, taskID, map[string]any{
				"workspaceID": workspaceID,
				"AddedUser":   assignee.UserID,
				"taskName":    *workspaceTaskName,
				"newStatus":   statusName,
				"tableID":     tableID,
				"taskID":      taskID,
				"viewID":      mainViewID,
			})
		}
	}

	return workspaceTask, nil
}

func (a *App) AddFieldValue(ctx context.Context, workspaceID string, tableID string, name string, field string, user model.User) (*model.TaskOrderField, *model.AppError) {
	linkedTableID, err := a.Store.Workspace.GetLinkedTableID(tableID, field)
	if err != nil {
		tlog.Errorw("Failed to get linked table ID",
			"table_id", tableID,
			"field", field,
			"error", err,
		)
		return nil, model.NewAppError("workspace.field_value_add_failed", http.StatusInternalServerError)
	}

	if linkedTableID == "" {
		return nil, model.NewAppError("workspace.relationship_not_found", http.StatusNotFound)
	}

	// tableID is the parent table that owns the single-select field, check edit_fields on it.
	if !a.CanPerformRowAction(user, workspaceID, tableID, model.PermissionEditFields, model.PermissionEditFields) {
		return nil, model.NewAppError("workspace.forbidden", http.StatusForbidden)
	}

	tableName, err := a.Store.Workspace.GetTableName(linkedTableID)
	if err != nil {
		tlog.Errorw("Failed to get linked table name",
			"linked_table_id", linkedTableID,
			"error", err,
		)
		return nil, model.NewAppError("workspace.linked_table_not_found", http.StatusNotFound)
	}

	options, err := a.Store.Workspace.GetKanbanStatusOptions(tableName)
	if err != nil {
		tlog.Errorw("Failed to get existing field values",
			"table_name", tableName,
			"error", err,
		)
		return nil, model.NewAppError("workspace.field_value_add_failed", http.StatusInternalServerError)
	}

	for _, opt := range options {
		if opt.Name == name {
			return nil, model.NewAppError("workspace.field_value_already_exists", http.StatusConflict)
		}
	}

	value, err := a.Store.Workspace.CreateFieldValue(workspaceID, tableID, linkedTableID, tableName, name, field, model.NewID(), "#16a34a")
	if err != nil {
		tlog.Errorw("Failed to add field value",
			"workspace_id", workspaceID,
			"table_id", tableID,
			"error", err,
		)
		return nil, model.NewAppError("workspace.field_value_add_failed", http.StatusInternalServerError)
	}

	a.publishProjectChange(ctx, projectChange{kind: "NEW_KANBAN_SECTION_ADDED", workspaceID: workspaceID, tableID: tableID,
		data: map[string]any{"data": value, "field": field, "parent_table_id": linkedTableID}})

	return value, nil
}

// GetWorkspaceLinkedTaskData returns the option a task's single-select field
// holds, read from the option table the field on the task's table points to.
func (a *App) GetWorkspaceLinkedTaskData(tableID, field, optionID string) (*model.TaskOrderField, *model.AppError) {
	optionTableID, err := a.Store.Workspace.GetLinkedTableID(tableID, field)
	if err != nil || optionTableID == "" {
		return nil, model.NewAppError("workspace.view_or_field_not_found", http.StatusNotFound)
	}

	workspaceTask, err := a.Store.Workspace.GetLinkedTaskData(optionTableID, optionID)
	if err != nil {
		tlog.Errorw("Failed to retrieve linked task data",
			"table_id", tableID,
			"field", field,
			"option_id", optionID,
			"error", err,
		)
		return nil, model.NewAppError("workspace.linked_task_data_failed", http.StatusInternalServerError)
	}

	return workspaceTask, nil
}

// GetTaskPageNumber returns the 1-based page of the grid that shows a task,
// which is the page of the branch it is in.
func (a *App) GetTaskPageNumber(ctx context.Context, workspaceID, tableID, taskID string, limit int, filters model.FilterPayload, user model.User) (int, *model.AppError) {
	if !a.canReadTable(ctx, user, workspaceID, tableID) {
		return 1, model.NewAppError("workspace.forbidden", http.StatusForbidden)
	}

	tableName, err := a.Store.Workspace.GetTableName(tableID)
	if err != nil {
		return 1, model.NewAppError("workspace.task_page_retrieval_failed", http.StatusInternalServerError)
	}
	// Build the filter using the same approach as GetFilteredTableData so positions match fetchPage.
	var filter model.SQLFilter
	colTypes, cErr := a.Store.Workspace.GetTableColumnTypes(tableName)
	if cErr == nil && (len(filters.FlatFilters) > 0 || len(filters.Groups) > 0) {
		headers := a.buildTableHeaders(tableID, colTypes, false, true)
		filter = a.filterSQLForHeaders(tableID, headers, filters, a.filterLocation(filters, user))
	}

	sortSQL := gridSortSQL(filters.Sort, columnSet(colTypes))
	position, err := a.Store.Workspace.GetBranchPosition(ctx, tableName, taskID, a.assignedOnlyFilter(user, workspaceID, tableID), filter, sortSQL)
	if err != nil || position <= 0 {
		return 1, nil
	}

	if limit < 1 {
		limit = 100
	}

	return (position-1)/limit + 1, nil
}

// GetTableStatusTypes returns a table's status options and the ids of those
// that count as Done or Closed, to a user who may read the table.
func (a *App) GetTableStatusTypes(ctx context.Context, workspaceID, tableID string, user model.User) ([]string, []map[string]string, *model.AppError) {
	if !a.canReadTable(ctx, user, workspaceID, tableID) {
		return nil, nil, model.NewAppError("workspace.forbidden", http.StatusForbidden)
	}

	failed := model.NewAppError("workspace.status_types_retrieval_failed", http.StatusInternalServerError)

	doneIDs, err := a.doneStatusIDs(tableID)
	if err != nil {
		tlog.Errorw("Failed to get status types", "table_id", tableID, "error", err)
		return nil, nil, failed
	}

	ctx, cancel := a.dbCtx(ctx)
	defer cancel()
	options, err := a.Store.Workspace.GetStatusOptions(ctx, tableID)
	if err != nil {
		tlog.Errorw("Failed to get status options", "table_id", tableID, "error", err)
		return nil, nil, failed
	}

	if options == nil {
		options = []map[string]string{}
	}

	return doneIDs, options, nil
}

func (a *App) GetItemForTableByID(ctx context.Context, workspaceID, tableID, itemID string, user model.User) (map[string]interface{}, []model.WorkspaceHeaders, *model.AppError) {
	if !a.canReadTable(ctx, user, workspaceID, tableID) {
		return nil, nil, model.NewAppError("workspace.forbidden", http.StatusForbidden)
	}

	if appErr := a.requireTaskAccess(user, workspaceID, tableID, itemID); appErr != nil {
		return nil, nil, appErr
	}

	tableName, err := a.Store.Workspace.GetTableName(tableID)
	if err != nil {
		return nil, nil, model.NewAppError("workspace.item_retrieval_failed", http.StatusInternalServerError)
	}

	item, err := a.Store.Workspace.GetTableRowByID(tableName, itemID)
	if err != nil {
		return nil, nil, model.NewAppError("workspace.item_retrieval_failed", http.StatusInternalServerError)
	}

	if item == nil {
		return nil, nil, model.NewAppError("workspace.item_not_found", http.StatusNotFound)
	}

	colTypes, err := a.Store.Workspace.GetTableColumnTypes(tableName)
	if err != nil {
		return nil, nil, model.NewAppError("workspace.item_retrieval_failed", http.StatusInternalServerError)
	}

	headers := a.buildTableHeaders(tableID, colTypes, false, true)

	// hydrate single item
	id, _ := item["id"].(string)
	for _, header := range headers {
		if !header.SingleSelect && header.LinkedID != "" && id != "" {
			if linkedIDs, err := a.Store.Workspace.GetLinkedIDs(header.LinkedID, id); err == nil {
				item[header.Name] = linkedIDs
			}
		}
	}

	for _, header := range headers {
		if header.SingleSelect {
			var itemName string
			if name, ok := item[header.Name].(string); ok {
				itemName = name
			} else {
				continue
			}

			if linkedName, err := a.Store.Workspace.GetSingleSelectValues(header.LinkedID, itemName); err == nil {
				item[header.Name] = linkedName
			}
		}
	}

	a.restrictLinkedItems(ctx, []map[string]interface{}{item}, headers, user, workspaceID)

	return item, headers, nil
}

// GetSubtasks returns the subtasks of a task that the user may see, shaped
// like the grid's rows. The task panel reads them here because each view
// loads a different part of the table.
func (a *App) GetSubtasks(ctx context.Context, workspaceID, tableID, parentID string, user model.User) ([]map[string]interface{}, *model.AppError) {
	if !a.canReadTable(ctx, user, workspaceID, tableID) {
		return nil, model.NewAppError("workspace.forbidden", http.StatusForbidden)
	}

	if appErr := a.requireTaskAccess(user, workspaceID, tableID, parentID); appErr != nil {
		return nil, appErr
	}

	tableName, err := a.Store.Workspace.GetTableName(tableID)
	if err != nil {
		return nil, model.NewAppError("workspace.item_retrieval_failed", http.StatusInternalServerError)
	}

	colTypes, err := a.Store.Workspace.GetTableColumnTypes(tableName)
	if err != nil {
		return nil, model.NewAppError("workspace.item_retrieval_failed", http.StatusInternalServerError)
	}

	filter := model.SQLFilter{
		SQL:  "main.parent_task_id = ? AND main.id <> main.parent_task_id",
		Args: []any{parentID},
	}.And(a.assignedOnlyFilter(user, workspaceID, tableID))

	subtasks, err := a.Store.Workspace.GetTableRowsFiltered(ctx, tableName, filter)
	if err != nil {
		tlog.Errorw("Failed to get subtasks",
			"workspace_id", workspaceID,
			"table_id", tableID,
			"error", err,
		)
		return nil, model.NewAppError("workspace.item_retrieval_failed", http.StatusInternalServerError)
	}

	headers := a.buildTableHeaders(tableID, colTypes, false, true)
	a.hydrateLinkedAndSingleSelect(ctx, subtasks, headers)
	a.restrictLinkedItems(ctx, subtasks, headers, user, workspaceID)

	return subtasks, nil
}

// GetTaskCompletion returns how many subtasks below a task are still open
// and, when the task has a parent the user may see, how the parent stands,
// so that closing a task can offer to close what belongs with it.
func (a *App) GetTaskCompletion(ctx context.Context, workspaceID, tableID, taskID string, user model.User) (*model.TaskCompletion, *model.AppError) {
	if !a.canReadTable(ctx, user, workspaceID, tableID) {
		return nil, model.NewAppError("workspace.forbidden", http.StatusForbidden)
	}

	if appErr := a.requireTaskAccess(user, workspaceID, tableID, taskID); appErr != nil {
		return nil, appErr
	}

	failed := model.NewAppError("workspace.item_retrieval_failed", http.StatusInternalServerError)
	completion := &model.TaskCompletion{}

	tableName, err := a.Store.Workspace.GetTableName(tableID)
	if err != nil {
		return nil, failed
	}

	hasStatus, err := a.Store.Workspace.TableHasColumn(tableName, "status")
	if err != nil {
		return nil, failed
	}

	if !hasStatus {
		return completion, nil
	}

	done, err := a.doneStatusIDs(tableID)
	if err != nil {
		return nil, failed
	}

	access := a.assignedOnlyFilter(user, workspaceID, tableID)

	open, err := a.Store.Workspace.GetSubtaskIDs(ctx, tableName, taskID, done, access)
	if err != nil {
		tlog.Errorw("Failed to count open subtasks", "workspace_id", workspaceID, "table_id", tableID, "error", err)
		return nil, failed
	}

	completion.OpenSubtasks = len(open)

	parentID, err := a.Store.Workspace.GetTaskParentID(tableName, taskID)
	if errors.Is(err, sql.ErrNoRows) || (err == nil && (parentID == "" || parentID == taskID)) {
		return completion, nil
	}

	if err != nil {
		return nil, failed
	}

	parents, err := a.Store.Workspace.GetTableRowsFiltered(ctx, tableName, model.SQLFilter{
		SQL:  "main.id = ?",
		Args: []any{parentID},
	}.And(access))
	if err != nil {
		return nil, failed
	}

	if len(parents) == 0 {
		return completion, nil
	}

	parentOpen, err := a.Store.Workspace.GetSubtaskIDs(ctx, tableName, parentID, done, access)
	if err != nil {
		return nil, failed
	}

	status, _ := parents[0]["status"].(string)
	name, _ := parents[0]["name"].(string)
	completion.Parent = &model.TaskCompletionParent{
		ID:           parentID,
		Name:         name,
		Open:         !slices.Contains(done, status),
		OpenSubtasks: len(parentOpen),
	}

	return completion, nil
}

// CompleteSubtasks gives every open subtask below a task status, which must
// be a Done or Closed one, each changed as if by hand, and returns how many
// it changed.
func (a *App) CompleteSubtasks(ctx context.Context, workspaceID, tableID, taskID, status string, user model.User) (int, *model.AppError) {
	if !a.canReadTable(ctx, user, workspaceID, tableID) {
		return 0, model.NewAppError("workspace.forbidden", http.StatusForbidden)
	}

	if appErr := a.requireTaskAccess(user, workspaceID, tableID, taskID); appErr != nil {
		return 0, appErr
	}

	failed := model.NewAppError("workspace.task_update_failed", http.StatusInternalServerError)

	tableName, err := a.Store.Workspace.GetTableName(tableID)
	if err != nil {
		return 0, failed
	}

	done, err := a.doneStatusIDs(tableID)
	if err != nil {
		return 0, failed
	}

	if !slices.Contains(done, status) {
		return 0, model.NewAppError("request.invalid", http.StatusBadRequest)
	}

	open, err := a.Store.Workspace.GetSubtaskIDs(ctx, tableName, taskID, done, a.assignedOnlyFilter(user, workspaceID, tableID))
	if err != nil {
		tlog.Errorw("Failed to find open subtasks", "workspace_id", workspaceID, "table_id", tableID, "error", err)
		return 0, failed
	}

	completed := 0
	for _, id := range open {
		if _, appErr := a.UpdateWorkspaceTask(ctx, workspaceID, tableID, id, "status", status, user, ""); appErr != nil {
			return completed, appErr
		}

		completed++
	}

	return completed, nil
}

// doneStatusIDs returns the ids of a table's status options that count as
// Done or Closed.
func (a *App) doneStatusIDs(tableID string) ([]string, error) {
	statusMap, err := a.Store.Workspace.GetStatusTypeMap(tableID)
	if err != nil {
		return nil, err
	}

	ids := []string{}
	for id, statusType := range statusMap {
		if statusType == "Done" || statusType == "Closed" {
			ids = append(ids, id)
		}
	}

	return ids, nil
}

func (a *App) UpdateWorkspaceTable(ctx context.Context, workspaceID, tableID, name string, user model.User) (*string, *model.AppError) {
	if !a.ProjectWorkspaceHasPermission(user, workspaceID, model.PermissionUpdateWorkspaceTable) &&
		!a.CanPerformRowAction(user, workspaceID, tableID, model.PermissionCreateFields, model.PermissionCreateFields) {
		return nil, model.NewAppError("workspace.forbidden", http.StatusForbidden)
	}

	if strings.TrimSpace(name) == "" {
		return nil, model.NewAppError("workspace.table_name_empty", http.StatusBadRequest)
	}

	data, err := a.Store.Workspace.UpdateTable(workspaceID, tableID, name)
	if err != nil {
		tlog.Errorw("Failed to update workspace table",
			"workspace_id", workspaceID,
			"table_id", tableID,
			"error", err,
		)
		return nil, model.NewAppError("workspace.table_update_failed", http.StatusInternalServerError)
	}

	a.publishProjectChange(ctx, projectChange{kind: "UPDATE_TABLE_NAME", workspaceID: workspaceID, tableID: tableID,
		data: map[string]any{"data": map[string]any{"id": tableID, "name": name}}})
	return data, nil
}

// UpdateWorkspace renames a workspace and, when description is not nil,
// replaces its description.
func (a *App) UpdateWorkspace(workspaceID, name string, description *string, user model.User) (*string, *model.AppError) {
	if !a.ProjectWorkspaceHasPermission(user, workspaceID, model.PermissionUpdateWorkspace) {
		return nil, model.NewAppError("workspace.forbidden", http.StatusForbidden)
	}

	name = strings.TrimSpace(name)
	if name == "" {
		return nil, model.NewAppError("workspace.workspace_name_empty", http.StatusBadRequest)
	}

	data, err := a.Store.Workspace.Update(workspaceID, name, description)
	if err != nil {
		tlog.Errorw("Failed to update workspace",
			"workspace_id", workspaceID,
			"error", err,
		)
		return nil, model.NewAppError("workspace.workspace_update_failed", http.StatusInternalServerError)
	}

	return data, nil
}

func (a *App) UpdateWorkspaceFilter(ctx context.Context, workspaceID, tableID, viewID, filterID, savedFilterID, name string, filters model.FilterPayload, isPrivate bool, user model.User, isActive bool) *model.AppError {
	if appErr := a.requireViewAccess(ctx, user.ID, viewID); appErr != nil {
		return appErr
	}

	if filterID == "" || savedFilterID == "" {
		return model.NewAppError("request.invalid", http.StatusBadRequest)
	}

	ownerPrivate, createdBy, err := a.Store.Workspace.GetSavedFilterOwner(savedFilterID, workspaceID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.NewAppError("workspace.filter_not_found", http.StatusNotFound)
		}

		tlog.Errorw("Failed to get saved filter owner",
			"workspace_id", workspaceID,
			"saved_filter_id", savedFilterID,
			"error", err,
		)
		return model.NewAppError("workspace.filter_update_failed", http.StatusInternalServerError)
	}

	if ownerPrivate {
		if createdBy != user.ID {
			return model.NewAppError("workspace.forbidden", http.StatusForbidden)
		}
	} else {
		if !a.ProjectWorkspaceHasPermission(user, workspaceID, model.PermissionUpdateWorkspaceView) {
			return model.NewAppError("workspace.forbidden", http.StatusForbidden)
		}
	}

	if err := a.Store.Workspace.UpdateFilter(filterID, workspaceID, tableID, filters, name, isPrivate, user.ID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.NewAppError("workspace.filter_not_found", http.StatusNotFound)
		}

		tlog.Errorw("Failed to update filter",
			"workspace_id", workspaceID,
			"table_id", tableID,
			"error", err,
		)
		return model.NewAppError("workspace.filter_update_failed", http.StatusInternalServerError)
	}

	if isActive {
		if err := a.Store.Workspace.DeactivateFiltersForView(workspaceID, tableID, viewID); err != nil {
			tlog.Errorw("Failed to deactivate filters for view",
				"workspace_id", workspaceID,
				"table_id", tableID,
				"view_id", viewID,
				"error", err,
			)
			return model.NewAppError("workspace.filter_update_failed", http.StatusInternalServerError)
		}
	}

	found, err := a.Store.Workspace.UpdateSavedFilter(savedFilterID, workspaceID, tableID, viewID, isActive)
	if err != nil {
		tlog.Errorw("Failed to update saved filter",
			"workspace_id", workspaceID,
			"saved_filter_id", savedFilterID,
			"error", err,
		)
		return model.NewAppError("workspace.filter_update_failed", http.StatusInternalServerError)
	}

	if !found {
		return model.NewAppError("workspace.filter_not_found", http.StatusNotFound)
	}

	a.publishSavedFilter(ctx, "FILTER_UPDATED", workspaceID, tableID, viewID, savedFilterID, filterID, name, filters, isPrivate, isActive, createdBy)

	return nil
}

func (a *App) DeleteSavedFilter(user model.User, workspaceID, savedFilterID string) *model.AppError {
	if savedFilterID == "" {
		return model.NewAppError("request.invalid", http.StatusBadRequest)
	}

	isPrivate, createdBy, err := a.Store.Workspace.GetSavedFilterOwner(savedFilterID, workspaceID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.NewAppError("workspace.filter_not_found", http.StatusNotFound)
		}

		tlog.Errorw("Failed to get saved filter owner",
			"workspace_id", workspaceID,
			"saved_filter_id", savedFilterID,
			"error", err,
		)
		return model.NewAppError("workspace.filter_delete_failed", http.StatusInternalServerError)
	}

	if isPrivate {
		if createdBy != user.ID {
			return model.NewAppError("workspace.forbidden", http.StatusForbidden)
		}
	} else {
		if !a.ProjectWorkspaceHasPermission(user, workspaceID, model.PermissionUpdateWorkspaceView) {
			return model.NewAppError("workspace.forbidden", http.StatusForbidden)
		}
	}

	if err := a.Store.Workspace.DeleteSavedFilter(savedFilterID, workspaceID); err != nil {
		tlog.Errorw("Failed to delete saved filter",
			"workspace_id", workspaceID,
			"saved_filter_id", savedFilterID,
			"error", err,
		)
		return model.NewAppError("workspace.filter_delete_failed", http.StatusInternalServerError)
	}

	return nil
}

func (a *App) SaveWorkspaceFilter(ctx context.Context, workspaceID, tableID, viewID, name, filterType string, filters model.FilterPayload, isPrivate bool, user model.User) (savedFilterID string, filterID string, appErr *model.AppError) {
	if appErr := a.requireViewAccess(ctx, user.ID, viewID); appErr != nil {
		return "", "", appErr
	}

	if !a.ProjectWorkspaceHasPermission(user, workspaceID, model.PermissionUpdateWorkspaceView) {
		return "", "", model.NewAppError("workspace.forbidden", http.StatusForbidden)
	}

	if workspaceID == "" || tableID == "" {
		return "", "", model.NewAppError("request.invalid", http.StatusBadRequest)
	}

	filterID, err := a.Store.Workspace.CreateFilter(workspaceID, tableID, filters, name, isPrivate, user.ID, model.NewID())
	if err != nil {
		tlog.Errorw("Failed to save filter",
			"workspace_id", workspaceID,
			"table_id", tableID,
			"error", err,
		)
		return "", "", model.NewAppError("workspace.filter_save_failed", http.StatusInternalServerError)
	}

	savedFilterID, err = a.Store.Workspace.CreateSavedFilter(workspaceID, tableID, viewID, filterID, user.ID, model.NewID())
	if err != nil {
		tlog.Errorw("Failed to save filter association",
			"workspace_id", workspaceID,
			"table_id", tableID,
			"filter_id", filterID,
			"error", err,
		)
		return "", "", model.NewAppError("workspace.filter_save_failed", http.StatusInternalServerError)
	}

	a.publishSavedFilter(ctx, "FILTER_CREATED", workspaceID, tableID, viewID, savedFilterID, filterID, name, filters, isPrivate, false, user.ID)

	return savedFilterID, filterID, nil
}

// publishSavedFilter publishes a saved filter as the view's filter list holds
// it. A private filter goes only to its owner, and a filter on a private view
// only to those who may see the view.
func (a *App) publishSavedFilter(ctx context.Context, kind, workspaceID, tableID, viewID, savedFilterID, filterID, name string, filters model.FilterPayload, isPrivate, isActive bool, ownerID string) {
	change := projectChange{kind: kind, workspaceID: workspaceID, tableID: tableID, data: map[string]any{
		"view_id": viewID,
		"filter": map[string]any{
			"id":         savedFilterID,
			"filter_id":  filterID,
			"name":       name,
			"is_private": isPrivate,
			"table_id":   tableID,
			"view_id":    viewID,
			"filters":    filters,
			"is_active":  isActive,
		},
	}}
	if isPrivate {
		change.onlyUsers = map[string]bool{ownerID: true}
	} else if audience, err := a.viewAudience(ctx, viewID); err == nil {
		change.onlyUsers = audience
	}

	a.publishProjectChange(ctx, change)
}

func (a *App) UpdateWorkspaceFilterActive(workspaceID, tableID, viewID, filterID string, isActive bool, user model.User) *model.AppError {
	if appErr := a.requireViewAccess(context.Background(), user.ID, viewID); appErr != nil {
		return appErr
	}

	if filterID == "" {
		return model.NewAppError("request.invalid", http.StatusBadRequest)
	}

	if isActive {
		if err := a.Store.Workspace.DeactivateFiltersForView(workspaceID, tableID, viewID); err != nil {
			tlog.Errorw("Failed to deactivate filters for view",
				"workspace_id", workspaceID,
				"table_id", tableID,
				"view_id", viewID,
				"error", err,
			)
			return model.NewAppError("workspace.filter_active_update_failed", http.StatusInternalServerError)
		}
	}

	found, err := a.Store.Workspace.UpdateFilterActiveStatus(workspaceID, tableID, viewID, filterID, isActive)
	if err != nil {
		tlog.Errorw("Failed to update filter active status",
			"workspace_id", workspaceID,
			"table_id", tableID,
			"view_id", viewID,
			"error", err,
		)
		return model.NewAppError("workspace.filter_active_update_failed", http.StatusInternalServerError)
	}

	if !found {
		return model.NewAppError("workspace.filter_not_found", http.StatusNotFound)
	}

	return nil
}

// TaskShownByFilter reports whether the grid, filtered by filters, shows the
// task, so a client can drop a row that an edit took out of the filter.
func (a *App) TaskShownByFilter(ctx context.Context, workspaceID, tableID, taskID string, filters model.FilterPayload, user model.User) (bool, *model.AppError) {
	if !a.canReadTable(ctx, user, workspaceID, tableID) {
		return false, model.NewAppError("workspace.forbidden", http.StatusForbidden)
	}

	ctx, cancel := a.dbCtx(ctx)
	defer cancel()
	tableName, err := a.Store.Workspace.GetTableName(tableID)
	if err != nil {
		tlog.Errorw("Failed to retrieve table name", "workspace_id", workspaceID, "table_id", tableID, "error", err)
		return false, model.NewAppError("workspace.filtered_table_data_failed", http.StatusInternalServerError)
	}

	filter, err := a.taskFilterSQL(tableID, tableName, filters, user)
	if err != nil {
		tlog.Errorw("Failed to build filter", "workspace_id", workspaceID, "table_id", tableID, "error", err)
		return false, model.NewAppError("workspace.filtered_table_data_failed", http.StatusInternalServerError)
	}

	shown, err := a.Store.Workspace.TaskOrParentMatches(ctx, tableName, taskID, filter.And(a.assignedOnlyFilter(user, workspaceID, tableID)))
	if err != nil {
		tlog.Errorw("Failed to check task against filter", "workspace_id", workspaceID, "table_id", tableID, "task_id", taskID, "error", err)
		return false, model.NewAppError("workspace.filtered_table_data_failed", http.StatusInternalServerError)
	}

	return shown, nil
}

// gridQuery is what a grid's page and its count are read with.
type gridQuery struct {
	tableName string
	headers   []model.WorkspaceHeaders
	access    model.SQLFilter
	filter    model.SQLFilter
}

// GetFilteredTableData returns one page of a grid, and how many branches the
// filter shows when withCount is set.
func (a *App) GetFilteredTableData(ctx context.Context, workspaceID, tableID, viewID string, filters model.FilterPayload, user model.User, withCount bool) (*model.WorkspaceTable, *model.AppError) {
	q, appErr := a.prepareGridQuery(ctx, workspaceID, tableID, viewID, filters, user)
	if appErr != nil {
		return nil, appErr
	}

	// A page without a limit would read every task of the table.
	const defaultGridPage, maxGridPage = 100, 500
	limit := filters.Limit
	if limit <= 0 {
		limit = defaultGridPage
	}

	limit = min(limit, maxGridPage)
	page := filters.Page
	if page < 1 {
		page = 1
	}

	offset := (page - 1) * limit

	sortSQL := gridSortSQL(filters.Sort, headerSet(q.headers))

	dbCtx, cancel := a.dbCtx(ctx)
	defer cancel()

	items, total, rootTotal, err := a.Store.Workspace.GetRootTasksPagedWithSubtasks(dbCtx, q.tableName, q.access, q.filter, sortSQL, limit, offset, withCount)
	if err != nil {
		if !clientLeft(ctx) {
			tlog.Errorw("Failed to retrieve filtered table data",
				"workspace_id", workspaceID,
				"table_id", tableID,
				"error", err,
			)
		}

		return nil, model.NewAppError("workspace.filtered_table_data_failed", http.StatusInternalServerError)
	}

	// The page is already limited to the user's tasks, but the subtasks
	// appended beneath them are not.
	if q.access.SQL != "" {
		items = filterByAssignee(items, user.ID)
	}

	a.hydrateLinkedAndSingleSelect(dbCtx, items, q.headers)
	a.restrictLinkedItems(dbCtx, items, q.headers, user, workspaceID)
	a.hydrateFileAttachments(dbCtx, items, q.headers, tableID)

	data := &model.WorkspaceTable{
		ID:        tableID,
		DataBase:  items,
		Headers:   q.headers,
		Total:     total,
		RootTotal: rootTotal,
	}

	return data, nil
}

// GetFilteredTableCount returns how many branches a grid's filter shows. The
// grid reads it beside its page, so the rows need not wait for it.
func (a *App) GetFilteredTableCount(ctx context.Context, workspaceID, tableID, viewID string, filters model.FilterPayload, user model.User) (int, *model.AppError) {
	q, appErr := a.prepareGridQuery(ctx, workspaceID, tableID, viewID, filters, user)
	if appErr != nil {
		return 0, appErr
	}

	dbCtx, cancel := a.dbCtx(ctx)
	defer cancel()

	count, err := a.Store.Workspace.CountBranches(dbCtx, q.tableName, q.access, q.filter)
	if err != nil {
		if !clientLeft(ctx) {
			tlog.Errorw("Failed to count filtered table data",
				"workspace_id", workspaceID,
				"table_id", tableID,
				"error", err,
			)
		}

		return 0, model.NewAppError("workspace.filtered_table_data_failed", http.StatusInternalServerError)
	}

	return count, nil
}

// prepareGridQuery checks the user may read the grid and builds its filter.
func (a *App) prepareGridQuery(ctx context.Context, workspaceID, tableID, viewID string, filters model.FilterPayload, user model.User) (*gridQuery, *model.AppError) {
	if appErr := a.requireViewAccess(ctx, user.ID, viewID); appErr != nil {
		return nil, appErr
	}

	if !a.canReadTable(ctx, user, workspaceID, tableID) {
		return nil, model.NewAppError("workspace.forbidden", http.StatusForbidden)
	}

	filters = normalizeStatusFilters(filters)

	timezoneStr := filters.Timezone
	if timezoneStr == "" {
		timezoneStr = a.resolveUserTimezone(user.ID)
	}

	loc, err := time.LoadLocation(timezoneStr)
	if err != nil || timezoneStr == "" {
		loc = time.UTC
	}

	tableName, err := a.Store.Workspace.GetTableName(tableID)
	if err != nil {
		tlog.Errorw("Failed to retrieve table name",
			"workspace_id", workspaceID,
			"table_id", tableID,
			"error", err,
		)
		return nil, model.NewAppError("workspace.filtered_table_data_failed", http.StatusInternalServerError)
	}

	colTypes, err := a.Store.Workspace.GetTableColumnTypes(tableName)
	if err != nil {
		tlog.Errorw("Failed to retrieve table column types",
			"workspace_id", workspaceID,
			"table_id", tableID,
			"error", err,
		)
		return nil, model.NewAppError("workspace.filtered_table_data_failed", http.StatusInternalServerError)
	}

	headers := a.buildTableHeaders(tableID, colTypes, false, true)
	tableIDs := collectLinkedTableIDs(headers)
	tableNames, err := a.Store.Workspace.GetTableNamesByIDs(tableIDs)
	if err != nil {
		tlog.Errorw("Failed to resolve linked table names for filter",
			"workspace_id", workspaceID,
			"table_id", tableID,
			"error", err,
		)
		return nil, model.NewAppError("workspace.filtered_table_data_failed", http.StatusInternalServerError)
	}

	return &gridQuery{
		tableName: tableName,
		headers:   headers,
		access:    a.assignedOnlyFilter(user, workspaceID, tableID),
		filter:    a.Store.Workspace.BuildFilterSQLInline(filters, headers, loc, tableNames),
	}, nil
}

// gridSortSQL orders a grid by the sort asked for or, without one, newest
// first, so that a task just made is on the first page.
func gridSortSQL(sort []model.SortParam, known func(string) bool) string {
	sortSQL := buildSortSQL(sort, known)
	if sortSQL == buildSortSQL(nil, nil) && known("created_at") {
		return buildSortSQL([]model.SortParam{{Field: "created_at", Direction: "desc"}}, known)
	}

	return sortSQL
}

// buildSortSQL orders by the fields a request asks for that are among the
// table's columns, as known tells; any other is left out.
func buildSortSQL(sort []model.SortParam, known func(string) bool) string {
	validField := regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	parts := make([]string, 0, len(sort))
	dir := "ASC"
	for _, s := range sort {
		if !validField.MatchString(s.Field) || !known(s.Field) {
			continue
		}

		dir = "ASC"
		if strings.EqualFold(s.Direction, "desc") {
			dir = "DESC"
		}

		parts = append(parts, fmt.Sprintf("main.`%s` %s", s.Field, dir))
	}

	if len(parts) == 0 {
		return "ORDER BY main.id ASC"
	}
	// The id breaks ties so that pages do not overlap. It runs the way the
	// last field does: an index on that field holds the id too, and is read
	// in order only when both run the same way.
	return "ORDER BY " + strings.Join(parts, ", ") + ", main.id " + dir
}

func columnSet(colTypes []*sql.ColumnType) func(string) bool {
	names := make(map[string]bool, len(colTypes))
	for _, ct := range colTypes {
		names[ct.Name()] = true
	}

	return func(name string) bool { return names[name] }
}

func headerSet(headers []model.WorkspaceHeaders) func(string) bool {
	names := make(map[string]bool, len(headers))
	for _, h := range headers {
		names[h.Name] = true
	}

	return func(name string) bool { return names[name] }
}

// maxLinkChanges caps how many links one request may add and remove.
const maxLinkChanges = 1000

var linkIDPattern = regexp.MustCompile(`^[A-Za-z0-9-]{1,36}$`)

func validLinkID(id string) bool {
	return linkIDPattern.MatchString(id)
}

// ChangeTaskLinks adds and removes links in a task's link field. Links the
// request does not name are never touched, so a client that saw only part of
// them cannot remove the rest.
func (a *App) ChangeTaskLinks(ctx context.Context, workspaceID, tableID, taskID, field string, add, remove []string, user model.User) (*map[string]interface{}, *model.AppError) {
	if !a.CanPerformRowAction(user, workspaceID, tableID, model.PermissionUpdateTask, model.PermissionUpdateTask) {
		return nil, model.NewAppError("workspace.forbidden", http.StatusForbidden)
	}

	if appErr := a.requireTaskAccess(user, workspaceID, tableID, taskID); appErr != nil {
		return nil, appErr
	}

	invalid := model.NewAppError("workspace.task_link_invalid_id", http.StatusBadRequest)
	if !validLinkID(taskID) || field == "" || len(add)+len(remove) > maxLinkChanges {
		return nil, invalid
	}

	adding := make(map[string]bool, len(add))
	for _, id := range add {
		if !validLinkID(id) {
			return nil, invalid
		}

		adding[id] = true
	}

	for _, id := range remove {
		if !validLinkID(id) || adding[id] {
			return nil, invalid
		}
	}

	ctx, cancel := a.dbCtx(ctx)
	defer cancel()

	workspaceTask, err := a.Store.Workspace.ChangeTaskLinks(ctx, workspaceID, tableID, taskID, field, add, remove)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, model.NewAppError("workspace.view_or_field_not_found", http.StatusNotFound)
	case errors.Is(err, model.ErrLinkTargetMissing):
		return nil, invalid
	case err != nil:
		tlog.Errorw("Failed to update task link",
			"workspace_id", workspaceID,
			"table_id", tableID,
			"task_id", taskID,
			"field", field,
			"error", err,
		)
		return nil, model.NewAppError("workspace.task_link_update_failed", http.StatusInternalServerError)
	}

	a.publishTaskRow(ctx, "task_updated", workspaceID, tableID, taskID)
	if workspaceTask != nil {
		published := map[string]bool{}
		for _, key := range []string{"secondLinkedItemsToAdd", "secondLinkedItemsToDelete"} {
			items, _ := (*workspaceTask)[key].([]map[string]string)
			for _, item := range items {
				otherTable, otherTask := item["secondTableID"], item["linkedItemID"]
				if otherTable == "" || otherTask == "" || published[otherTable+"/"+otherTask] {
					continue
				}

				published[otherTable+"/"+otherTask] = true
				a.publishTaskRow(ctx, "task_updated", workspaceID, otherTable, otherTask)
			}
		}
	}

	return workspaceTask, nil
}

func (a *App) UpdateDisplayName(ctx context.Context, workspaceID string, tableID string, viewID string, displayName string, tableFieldName string, visible bool, user model.User) (*map[string]interface{}, *model.AppError) {
	if appErr := a.requireViewAccess(ctx, user.ID, viewID); appErr != nil {
		return nil, appErr
	}

	if !a.ProjectWorkspaceHasPermission(user, workspaceID, model.PermissionUpdateWorkspaceTable) &&
		!a.CanPerformRowAction(user, workspaceID, tableID, model.PermissionEditFields, model.PermissionEditFields) {
		return nil, model.NewAppError("workspace.forbidden", http.StatusForbidden)
	}

	itemOrderJSON, err := a.Store.Workspace.GetGridViewItemOrder(workspaceID, tableID, viewID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, model.NewAppError("workspace.view_or_field_not_found", http.StatusNotFound)
		}

		tlog.Errorw("Failed to get view item order",
			"workspace_id", workspaceID,
			"table_id", tableID,
			"view_id", viewID,
			"error", err,
		)
		return nil, model.NewAppError("workspace.display_name_update_failed", http.StatusInternalServerError)
	}

	var items []map[string]interface{}
	if err := json.Unmarshal(itemOrderJSON, &items); err != nil {
		tlog.Errorw("Failed to parse view item order",
			"workspace_id", workspaceID,
			"view_id", viewID,
			"error", err,
		)
		return nil, model.NewAppError("workspace.display_name_update_failed", http.StatusInternalServerError)
	}

	updated := false
	for _, item := range items {
		if name, ok := item["name"].(string); ok && name == tableFieldName {
			item["display_name"] = displayName
			item["visible"] = visible
			updated = true
			break
		}
	}

	if !updated {
		return nil, model.NewAppError("workspace.view_or_field_not_found", http.StatusNotFound)
	}

	updatedJSON, err := json.Marshal(items)
	if err != nil {
		tlog.Errorw("Failed to marshal view item order",
			"workspace_id", workspaceID,
			"view_id", viewID,
			"error", err,
		)
		return nil, model.NewAppError("workspace.display_name_update_failed", http.StatusInternalServerError)
	}

	if err := a.Store.Workspace.UpdateViewItemOrder(workspaceID, tableID, viewID, updatedJSON); err != nil {
		tlog.Errorw("Failed to update view item order",
			"workspace_id", workspaceID,
			"view_id", viewID,
			"error", err,
		)
		return nil, model.NewAppError("workspace.display_name_update_failed", http.StatusInternalServerError)
	}

	if err := a.Store.Workspace.UpdateFieldDisplayName(workspaceID, tableID, tableFieldName, displayName); err != nil {
		tlog.Errorw("Failed to update field display name",
			"workspace_id", workspaceID,
			"table_id", tableID,
			"field_name", tableFieldName,
			"error", err,
		)
		return nil, model.NewAppError("workspace.display_name_update_failed", http.StatusInternalServerError)
	}

	result := map[string]interface{}{
		"workspace_id":  workspaceID,
		"table_id":      tableID,
		"updated_field": tableFieldName,
		"display_name":  displayName,
		"visible":       visible,
		"view_id":       viewID,
	}

	a.publishViewChange(ctx, viewID, projectChange{kind: "DISPLAY_NAME_UPDATE", workspaceID: workspaceID, tableID: tableID,
		data: map[string]any{"data": result, "view_id": viewID}})
	return &result, nil
}

// maxTaskDescription is what a TEXT column holds.
const maxTaskDescription = 65535

// CreateWorkspaceTask creates a task named name. singleSelect puts it in the
// option of the section field a Kanban column stands for; fields sets the
// rest of what the task starts with.
func (a *App) CreateWorkspaceTask(ctx context.Context, workspaceID string, tableID string, name string, singleSelect string, section string, fields model.NewTaskFields, user model.User, parentTaskID string) (*map[string]interface{}, *model.AppError) {
	if !a.CanPerformRowAction(user, workspaceID, tableID, model.PermissionCreateTask, model.PermissionCreateTask) {
		return nil, model.NewAppError("workspace.forbidden", http.StatusForbidden)
	}

	name = strings.TrimSpace(name)
	if name == "" {
		return nil, model.NewAppError("workspace.task_name_empty", http.StatusBadRequest)
	}

	tableName, err := a.Store.Workspace.GetTableName(tableID)
	if err != nil {
		return nil, model.NewAppError("workspace.view_or_field_not_found", http.StatusNotFound)
	}

	if parentTaskID != "" {
		if _, err := a.Store.Workspace.GetTaskFieldValue(tableName, parentTaskID, "id"); err != nil {
			return nil, model.NewAppError("request.invalid", http.StatusBadRequest)
		}

		if appErr := a.requireTaskAccess(user, workspaceID, tableID, parentTaskID); appErr != nil {
			return nil, appErr
		}
	}

	selects := map[string]string{}
	if singleSelect != "" {
		if appErr := a.checkSingleSelect(ctx, tableID, tableName, section, singleSelect); appErr != nil {
			return nil, appErr
		}

		selects[section] = singleSelect
	}

	if fields.Status != "" && fields.Status != "0" {
		if chosen, ok := selects["status"]; ok && chosen != fields.Status {
			return nil, model.NewAppError("request.invalid", http.StatusBadRequest)
		}

		if appErr := a.checkSingleSelect(ctx, tableID, tableName, "status", fields.Status); appErr != nil {
			return nil, appErr
		}

		selects["status"] = fields.Status
	}

	values, appErr := a.newTaskValues(ctx, workspaceID, tableName, fields)
	if appErr != nil {
		return nil, appErr
	}

	for field, option := range selects {
		values[field] = option
	}

	taskID := model.NewID()
	workspaceTask, err := a.Store.Workspace.CreateTask(workspaceID, tableID, name, user.ID, parentTaskID, taskID, time.Now().Unix(), values)
	if err != nil {
		tlog.Errorw("Failed to create task",
			"workspace_id", workspaceID,
			"table_id", tableID,
			"error", err,
		)
		return nil, model.NewAppError("workspace.task_create_failed", http.StatusInternalServerError)
	}

	if parentTaskID == "" {
		if err := a.Store.Workspace.SyncTaskKanban(workspaceID, tableID, taskID, selects, *workspaceTask); err != nil {
			tlog.Errorw("Failed to sync task kanban",
				"workspace_id", workspaceID,
				"table_id", tableID,
				"task_id", taskID,
				"error", err,
			)
			return nil, model.NewAppError("workspace.task_create_failed", http.StatusInternalServerError)
		}
	}

	a.publishTaskRow(ctx, "task_created", workspaceID, tableID, taskID)

	mainViewID, err := a.Store.Workspace.GetMainViewByTableID(workspaceID, tableID)
	if err != nil {
		tlog.Errorw("Failed to get main view for notification",
			"workspace_id", workspaceID,
			"table_id", tableID,
			"error", err,
		)
		return nil, model.NewAppError("workspace.task_main_view_failed", http.StatusInternalServerError)
	}

	a.createProjectNotification(workspaceID, user, model.NOTIFICATION_PROJECT_TASK_CREATED, map[string]interface{}{
		"taskName": name,
		"tableID":  tableID,
		"taskID":   taskID,
		"viewID":   mainViewID,
	})

	a.RecordActivity(user.ID, model.AppProjects, model.ActivityTaskCreated, tableID, taskID, map[string]any{
		"name":        name,
		"workspaceID": workspaceID,
		"tableID":     tableID,
		"taskID":      taskID,
		"viewID":      mainViewID,
	})

	if fields.Assignee != "" {
		a.createProjectNotification(workspaceID, user, model.NOTIFICATION_PROJECT_TASK_ASSIGNED, map[string]interface{}{
			"userID":   fields.Assignee,
			"taskName": name,
			"tableID":  tableID,
			"taskID":   taskID,
			"viewID":   mainViewID,
		})

		a.RecordActivity(user.ID, model.AppProjects, model.ActivityTaskAssigned, tableID, taskID, map[string]any{
			"workspaceID": workspaceID,
			"AddedUser":   fields.Assignee,
			"taskName":    name,
			"tableID":     tableID,
			"taskID":      taskID,
			"viewID":      mainViewID,
		})
	}

	return workspaceTask, nil
}

// checkSingleSelect refuses a value for field unless field is one of the
// table's single-select columns and value one of its options, or "0" for
// none.
func (a *App) checkSingleSelect(ctx context.Context, tableID, tableName, field, value string) *model.AppError {
	invalid := model.NewAppError("request.invalid", http.StatusBadRequest)
	if !a.writableColumn(ctx, tableName, field) {
		return invalid
	}

	optionTableID, err := a.Store.Workspace.GetLinkedTableID(tableID, field)
	if err != nil || optionTableID == "" {
		return invalid
	}

	if value == "0" {
		return nil
	}

	options, err := a.Store.Workspace.GetSingleSelectOptions(ctx, optionTableID, []string{value})
	if err != nil || options[value] == nil {
		return invalid
	}

	return nil
}

// newTaskValues checks the plain fields a task is created with and returns
// them by column.
func (a *App) newTaskValues(ctx context.Context, workspaceID, tableName string, fields model.NewTaskFields) (map[string]any, *model.AppError) {
	invalid := model.NewAppError("request.invalid", http.StatusBadRequest)
	values := map[string]any{}

	set := func(column string, value any) bool {
		if !a.writableColumn(ctx, tableName, column) {
			return false
		}

		values[column] = value
		return true
	}

	if fields.Assignee != "" {
		members, err := a.Store.Workspace.GetMemberUserIDs(ctx, workspaceID, []string{fields.Assignee})
		if err != nil {
			tlog.Errorw("Failed to check the assignee's membership", "workspace_id", workspaceID, "user_id", fields.Assignee, "error", err)
			return nil, model.NewAppError("workspace.task_create_failed", http.StatusInternalServerError)
		}

		if len(members) == 0 || !set("assignee", fields.Assignee) {
			return nil, invalid
		}
	}

	if fields.StartDate < 0 || fields.DueDate < 0 {
		return nil, invalid
	}

	if fields.StartDate > 0 && fields.DueDate > 0 && fields.StartDate > fields.DueDate {
		return nil, model.NewAppError("workspace.task_dates_reversed", http.StatusBadRequest)
	}

	if fields.StartDate > 0 && !set("start_date", fields.StartDate) {
		return nil, invalid
	}

	if fields.DueDate > 0 && !set("due_date", fields.DueDate) {
		return nil, invalid
	}

	if fields.Description != "" {
		if len(fields.Description) > maxTaskDescription || !set("description", fields.Description) {
			return nil, invalid
		}
	}

	return values, nil
}

func (a *App) buildGridViewOrder(workspaceID, tableID string) (string, *model.AppError) {
	defaultColumns := []model.TaskOrderField{
		{Name: "id", DisplayName: "ID", Width: "150", Visible: false},
		{Name: "name", DisplayName: "Name", Width: "150", Visible: true},
		{Name: "start_date", DisplayName: "Start Date", Width: "150", Visible: false},
		{Name: "due_date", DisplayName: "Due Date", Width: "150", Visible: true},
		{Name: "assignee", DisplayName: "Assignee", Width: "150", Visible: true},
		{Name: "status", DisplayName: "Status", Width: "150", Visible: true},
		{Name: "description", DisplayName: "Description", Width: "150", Visible: true},
		{Name: "updated_at", DisplayName: "Updated At", Width: "150", Visible: true},
		{Name: "created_at", DisplayName: "Created At", Width: "150", Visible: false},
		{Name: "deleted_at", DisplayName: "Deleted At", Width: "150", Visible: true},
		{Name: "created_by", DisplayName: "Created By", Width: "150", Visible: false},
	}

	customFields, err := a.Store.Workspace.GetCustomFields(workspaceID, tableID)
	if err != nil {
		tlog.Errorw("Failed to fetch custom fields for grid view",
			"workspace_id", workspaceID,
			"table_id", tableID,
			"error", err,
		)
		return "", model.NewAppError("workspace.grid_view_create_failed", http.StatusInternalServerError)
	}

	existing := make(map[string]bool)
	for _, f := range defaultColumns {
		existing[f.Name] = true
	}

	for _, field := range customFields {
		if existing[field.Name] {
			continue
		}

		defaultColumns = append(defaultColumns, model.TaskOrderField{
			Name:        field.Name,
			DisplayName: field.DisplayName,
			Width:       "150",
			Visible:     true,
		})
	}

	taskOrderJSON, _ := json.Marshal(defaultColumns)

	return string(taskOrderJSON), nil
}

func (a *App) UpdateColumnWidth(ctx context.Context, workspaceID, tableID, viewID, columnName string, width int, user model.User) (*model.TaskOrderField, *model.AppError) {
	if !a.ProjectWorkspaceHasPermission(user, workspaceID, model.PermissionUpdateWorkspaceView) &&
		!a.ProjectWorkspaceHasPermission(user, workspaceID, model.PermissionUpdateWorkspaceTable) {
		return nil, model.NewAppError("workspace.forbidden", http.StatusForbidden)
	}

	err := a.Store.Workspace.UpdateColumnWidth(workspaceID, tableID, viewID, columnName, width)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.NewAppError("workspace.column_not_found", http.StatusNotFound)
		}

		tlog.Errorw("Failed to update column width",
			"workspace_id", workspaceID,
			"table_id", tableID,
			"error", err,
		)
		return nil, model.NewAppError("workspace.column_width_update_failed", http.StatusInternalServerError)
	}

	a.publishViewChange(ctx, viewID, projectChange{kind: "GRID_HEADER_UPDATE", workspaceID: workspaceID, tableID: tableID,
		data: map[string]any{"column_name": columnName, "width": width, "view_id": viewID}})

	return nil, nil
}

func (a *App) UpdateGridHeaders(ctx context.Context, workspaceID string, tableID string, headers []model.WorkspaceHeaders, user model.User, viewID string) ([]model.WorkspaceHeaders, *model.AppError) {
	if !a.ProjectWorkspaceHasPermission(user, workspaceID, model.PermissionUpdateWorkspaceView) {
		return nil, model.NewAppError("workspace.forbidden", http.StatusForbidden)
	}

	itemOrderRaw, err := a.Store.Workspace.GetViewItemOrder(workspaceID, tableID, viewID)
	if err != nil {
		tlog.Errorw("Failed to get view item order",
			"workspace_id", workspaceID,
			"table_id", tableID,
			"view_id", viewID,
			"error", err,
		)
		return nil, model.NewAppError("workspace.grid_headers_update_failed", http.StatusInternalServerError)
	}

	var itemOrder []map[string]interface{}
	if err := json.Unmarshal(itemOrderRaw, &itemOrder); err != nil {
		tlog.Errorw("Failed to parse view item order",
			"workspace_id", workspaceID,
			"view_id", viewID,
			"error", err,
		)
		return nil, model.NewAppError("workspace.grid_headers_update_failed", http.StatusInternalServerError)
	}

	var headerObjects []map[string]interface{}
	var headerOrderEntry map[string]interface{}
	var otherItems []map[string]interface{}

	for _, item := range itemOrder {
		if name, ok := item["name"].(string); ok {
			if name == "headerOrder" {
				headerOrderEntry = item
			} else {
				headerObjects = append(headerObjects, item)
			}
		} else {
			otherItems = append(otherItems, item)
		}
	}

	headerMap := make(map[string]map[string]interface{})
	for _, h := range headerObjects {
		if name, ok := h["name"].(string); ok {
			headerMap[name] = h
		}
	}

	var newHeaderObjects []map[string]interface{}
	var newHeaderOrderNames []string

	for _, h := range headers {
		name := h.Name
		newHeaderOrderNames = append(newHeaderOrderNames, name)
		if existing, ok := headerMap[name]; ok {
			newHeaderObjects = append(newHeaderObjects, existing)
		} else {
			newHeaderObjects = append(newHeaderObjects, map[string]interface{}{
				"name":         name,
				"display_name": "",
				"width":        "150",
				"visible":      true,
			})
		}
	}

	if headerOrderEntry != nil {
		headerOrderEntry["value"] = newHeaderOrderNames
	} else {
		headerOrderEntry = map[string]interface{}{
			"name":  "headerOrder",
			"value": newHeaderOrderNames,
		}
	}

	var newItemOrder []map[string]interface{}
	newItemOrder = append(newItemOrder, newHeaderObjects...)
	newItemOrder = append(newItemOrder, headerOrderEntry)
	newItemOrder = append(newItemOrder, otherItems...)

	updatedItemOrder, err := json.Marshal(newItemOrder)
	if err != nil {
		tlog.Errorw("Failed to marshal updated item order",
			"workspace_id", workspaceID,
			"view_id", viewID,
			"error", err,
		)
		return nil, model.NewAppError("workspace.grid_headers_update_failed", http.StatusInternalServerError)
	}

	if err := a.Store.Workspace.UpdateViewItemOrder(workspaceID, tableID, viewID, updatedItemOrder); err != nil {
		tlog.Errorw("Failed to save updated grid headers",
			"workspace_id", workspaceID,
			"table_id", tableID,
			"view_id", viewID,
			"error", err,
		)
		return nil, model.NewAppError("workspace.grid_headers_update_failed", http.StatusInternalServerError)
	}

	a.publishViewChange(ctx, viewID, projectChange{kind: "UPDATE_GRD_HEADER_ORDER", workspaceID: workspaceID, tableID: tableID,
		data: map[string]any{"value": headers, "view_id": viewID}})

	return headers, nil
}

func (a *App) UpdateCalculationsField(ctx context.Context, workspaceID, tableID, fieldName string, formula *model.FormulaSpec, user model.User) *model.AppError {
	// A formula is part of the field, so changing it is editing the field,
	// not a task.
	if !a.CanPerformRowAction(user, workspaceID, tableID, model.PermissionEditFields, model.PermissionEditFields) {
		return model.NewAppError("workspace.forbidden", http.StatusForbidden)
	}

	ctx, cancel := a.dbCtx(ctx)
	defer cancel()

	if err := a.Store.Workspace.UpdateFieldFormula(ctx, workspaceID, tableID, fieldName, formula); err != nil {
		tlog.Errorw("Failed to update field formula", "workspace_id", workspaceID, "table_id", tableID, "field", fieldName, "error", err)
		return model.NewAppError("workspace.field_formula_update_failed", http.StatusInternalServerError)
	}

	a.publishProjectChange(ctx, projectChange{kind: "field_formula_updated", workspaceID: workspaceID, tableID: tableID,
		data: map[string]any{"field_name": fieldName, "formula": formula}})

	return nil
}

// isPersonField reports whether a table's column holds a user.
func (a *App) isPersonField(tableID, field string) (bool, error) {
	if field == "assignee" {
		return true, nil
	}

	names, err := a.Store.Workspace.GetPersonFieldNamesForTable(tableID)
	return slices.Contains(names, field), err
}

// writableColumn reports whether a request may set the column: it has to be
// one of the table's own columns, spelled exactly as the table spells it,
// since MySQL would also accept ID for id, and not one the server keeps.
func (a *App) writableColumn(ctx context.Context, tableName, column string) bool {
	if column == "" || model.IsServerKeptColumn(column) {
		return false
	}

	names, err := a.Store.Workspace.GetColumnNames(ctx, tableName)
	return err == nil && slices.Contains(names, column)
}

func (a *App) AddMemberToTask(ctx context.Context, workspaceID string, tableID string, taskID string, userID string, fieldName string, user model.User) (*model.User, *model.AppError) {
	if !a.CanPerformRowAction(user, workspaceID, tableID, model.PermissionUpdateTask, model.PermissionUpdateTask) {
		return nil, model.NewAppError("workspace.forbidden", http.StatusForbidden)
	}

	if appErr := a.requireTaskAccess(user, workspaceID, tableID, taskID); appErr != nil {
		return nil, appErr
	}

	tableName, err := a.Store.Workspace.GetTableName(tableID)
	if err != nil {
		return nil, model.NewAppError("workspace.view_or_field_not_found", http.StatusNotFound)
	}

	if !a.writableColumn(ctx, tableName, fieldName) {
		return nil, model.NewAppError("request.invalid", http.StatusBadRequest)
	}

	if userID != "" {
		members, err := a.Store.Workspace.GetMemberUserIDs(ctx, workspaceID, []string{userID})
		if err != nil {
			tlog.Errorw("Failed to check the assignee's membership", "workspace_id", workspaceID, "user_id", userID, "error", err)
			return nil, model.NewAppError("workspace.member_to_task_add_failed", http.StatusInternalServerError)
		}

		if len(members) == 0 {
			return nil, model.NewAppError("request.invalid", http.StatusBadRequest)
		}
	}

	var previousAssignee string
	if fieldName == "assignee" {
		previousAssignee, _ = a.Store.Workspace.GetTaskAssigneeID(tableName, taskID)
	}

	if err := a.Store.Workspace.CreateMemberToTask(workspaceID, tableID, taskID, userID, fieldName); err != nil {
		tlog.Errorw("Failed to add member to task",
			"workspace_id", workspaceID,
			"table_id", tableID,
			"task_id", taskID,
			"error", err,
		)
		return nil, model.NewAppError("workspace.member_to_task_add_failed", http.StatusInternalServerError)
	}

	if fieldName == "assignee" {
		a.publishTaskAssignment(ctx, workspaceID, tableID, taskID, previousAssignee)
	} else {
		a.publishTaskRow(ctx, "task_updated", workspaceID, tableID, taskID)
	}

	var workspaceTaskMember *model.User
	if userID != "" {
		var err error
		workspaceTaskMember, err = a.Store.User.Get(userID)
		if err != nil {
			tlog.Errorw("Failed to fetch assigned user",
				"user_id", userID,
				"error", err,
			)
			return nil, model.NewAppError("workspace.member_to_task_add_failed", http.StatusInternalServerError)
		}
	}

	workspaceTaskName, err := a.Store.Workspace.GetTaskNameByID(workspaceID, tableID, taskID)
	if err != nil {
		tlog.Errorw("Failed to get task name for notification",
			"workspace_id", workspaceID,
			"table_id", tableID,
			"task_id", taskID,
			"error", err,
		)
		return nil, model.NewAppError("workspace.task_name_failed", http.StatusInternalServerError)
	}

	mainViewID, err := a.Store.Workspace.GetMainViewByTableID(workspaceID, tableID)
	if err != nil {
		tlog.Errorw("Failed to get main view for notification",
			"workspace_id", workspaceID,
			"table_id", tableID,
			"error", err,
		)
		return nil, model.NewAppError("workspace.task_main_view_failed", http.StatusInternalServerError)
	}

	a.createProjectNotification(workspaceID, user, model.NOTIFICATION_PROJECT_TASK_ASSIGNED, map[string]interface{}{
		"userID":   userID,
		"taskName": *workspaceTaskName,
		"tableID":  tableID,
		"taskID":   taskID,
		"viewID":   mainViewID,
	})

	a.RecordActivity(user.ID, model.AppProjects, model.ActivityTaskAssigned, tableID, taskID, map[string]any{
		"workspaceID": workspaceID,
		"AddedUser":   userID,
		"taskName":    *workspaceTaskName,
		"tableID":     tableID,
		"taskID":      taskID,
		"viewID":      mainViewID,
	})

	return workspaceTaskMember, nil
}

func (a *App) AddWorkspaceFolder(ctx context.Context, workspaceID string, name string, parentFolderID string, user model.User) (*model.WorkspaceFolder, *model.AppError) {
	if !a.ProjectWorkspaceHasPermission(user, workspaceID, model.PermissionUpdateWorkspace) {
		return nil, model.NewAppError("workspace.forbidden", http.StatusForbidden)
	}

	dbCtx, cancel := a.dbCtx(ctx)
	defer cancel()

	workspaceFolder, err := a.Store.Workspace.CreateFolder(dbCtx, workspaceID, name, parentFolderID, model.NewID())
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, model.NewAppError("workspace.folder_not_found", http.StatusNotFound)
	case err != nil:
		tlog.Errorw("Failed to add workspace folder",
			"workspace_id", workspaceID,
			"error", err,
		)
		return nil, model.NewAppError("workspace.folder_add_failed", http.StatusInternalServerError)
	}

	parent := workspaceFolder.ParentFolderID
	if parent == "" {
		parent = workspaceID
	}

	a.publishProjectChange(ctx, projectChange{kind: "CREATE_FOLDER", workspaceID: workspaceID,
		data: map[string]any{"data": map[string]any{"id": workspaceFolder.ID, "name": workspaceFolder.Name, "parent_folder_id": parent, "workspace_id": workspaceID}}})
	return workspaceFolder, nil
}

func (a *App) DeleteWorkspaceTableSingleField(ctx context.Context, workspaceID string, tableID string, fieldID string, linkedTableID string, user model.User) (bool, *model.AppError) {
	if !a.ProjectWorkspaceHasPermission(user, workspaceID, model.PermissionUpdateWorkspaceTable) &&
		!a.CanPerformRowAction(user, workspaceID, tableID, model.PermissionCreateFields, model.PermissionCreateFields) {
		return false, model.NewAppError("workspace.forbidden", http.StatusForbidden)
	}

	// The option table comes from the request, and the row is deleted from it
	// outright, so it must be one of this table's.
	dbCtx, cancel := a.dbCtx(ctx)
	defer cancel()
	isOption, err := a.Store.Workspace.IsOptionTableOf(dbCtx, tableID, linkedTableID)
	if err != nil {
		tlog.Errorw("Failed to check the option table", "table_id", tableID, "linked_table_id", linkedTableID, "error", err)
		return false, model.NewAppError("workspace.single_field_delete_failed", http.StatusInternalServerError)
	}

	if !isOption {
		return false, model.NewAppError("workspace.forbidden", http.StatusForbidden)
	}

	linkedTableName, err := a.Store.Workspace.GetTableName(linkedTableID)
	if err != nil {
		tlog.Errorw("Failed to get linked table name",
			"linked_table_id", linkedTableID,
			"error", err,
		)
		return false, model.NewAppError("workspace.single_field_delete_failed", http.StatusInternalServerError)
	}

	ok, err := a.Store.Workspace.DeleteTableSingleField(workspaceID, tableID, fieldID, linkedTableID, linkedTableName)
	if err != nil {
		tlog.Errorw("Failed to delete single select field",
			"workspace_id", workspaceID,
			"table_id", tableID,
			"field_id", fieldID,
			"error", err,
		)
		return false, model.NewAppError("workspace.single_field_delete_failed", http.StatusInternalServerError)
	}

	a.publishProjectChange(ctx, projectChange{kind: "DELETE_SINGLE_FIELD", workspaceID: workspaceID, tableID: tableID,
		data: map[string]any{"delete_id": fieldID, "linked_table_id": linkedTableID}})

	return ok, nil
}

func (a *App) DeleteWorkspaceTableField(ctx context.Context, workspaceID string, tableID string, fieldID string, linkBothDirections bool, linkedID string, user model.User) (map[string]interface{}, *model.AppError) {
	if !a.ProjectWorkspaceHasPermission(user, workspaceID, model.PermissionUpdateWorkspaceTable) &&
		!a.CanPerformRowAction(user, workspaceID, tableID, model.PermissionCreateFields, model.PermissionCreateFields) {
		return nil, model.NewAppError("workspace.forbidden", http.StatusForbidden)
	}

	var parentFieldID, parentTableID string
	if linkBothDirections {
		var err error
		parentFieldID, parentTableID, err = a.Store.Workspace.GetFieldParentInfo(fieldID)
		if err != nil {
			tlog.Errorw("Failed to get field parent info",
				"workspace_id", workspaceID,
				"field_id", fieldID,
				"error", err,
			)
			return nil, model.NewAppError("workspace.table_field_delete_failed", http.StatusInternalServerError)
		}
	}

	if err := a.Store.Workspace.DeleteTableField(workspaceID, tableID, fieldID, parentFieldID); err != nil {
		tlog.Errorw("Failed to delete table field",
			"workspace_id", workspaceID,
			"table_id", tableID,
			"field_id", fieldID,
			"error", err,
		)
		return nil, model.NewAppError("workspace.table_field_delete_failed", http.StatusInternalServerError)
	}

	a.publishProjectChange(ctx, projectChange{kind: "DELETE_FIELD", workspaceID: workspaceID, tableID: tableID,
		data: map[string]any{"field_id": fieldID, "linked_data": map[string]any{"parent_field_id": parentFieldID, "table_id": parentTableID}}})

	return map[string]interface{}{
		"parent_field_id": parentFieldID,
		"table_id":        parentTableID,
	}, nil
}

func (a *App) DeleteWorkspaceTableView(ctx context.Context, workspaceID string, tableID string, viewID string, user model.User) (bool, *model.AppError) {
	if appErr := a.requireViewAccess(ctx, user.ID, viewID); appErr != nil {
		return false, appErr
	}

	if !a.ProjectWorkspaceHasPermission(user, workspaceID, model.PermissionDeleteTableView) &&
		!a.ProjectWorkspaceTableHasPermission(user, workspaceID, tableID, "manage_views") {
		return false, model.NewAppError("workspace.forbidden", http.StatusForbidden)
	}

	deleted := projectChange{kind: "DELETE_VIEW", workspaceID: workspaceID, tableID: tableID, data: map[string]any{"delete_id": viewID}}
	if audience, err := a.viewAudience(ctx, viewID); err == nil {
		deleted.onlyUsers = audience
	}

	ok, err := a.Store.Workspace.DeleteTableView(workspaceID, tableID, viewID)
	if err != nil {
		tlog.Errorw("Failed to delete table view",
			"workspace_id", workspaceID,
			"table_id", tableID,
			"view_id", viewID,
			"error", err,
		)
		return false, model.NewAppError("workspace.table_view_delete_failed", http.StatusInternalServerError)
	}

	a.publishProjectChange(ctx, deleted)

	return ok, nil
}

func (a *App) DeleteWorkspaceTask(ctx context.Context, workspaceID string, tableID string, taskID string, user model.User) (bool, *model.AppError) {
	if !a.CanPerformRowAction(user, workspaceID, tableID, model.PermissionDeleteTask, model.PermissionDeleteTask) {
		isSingleSelect, _ := a.Store.Workspace.IsTableSingleSelect(tableID)
		parentID := a.Store.Workspace.GetTableParentID(tableID)
		if !isSingleSelect || parentID == "" || !a.CanPerformRowAction(user, workspaceID, parentID, model.PermissionEditFields, model.PermissionEditFields) {
			return false, model.NewAppError("workspace.forbidden", http.StatusForbidden)
		}
	}

	if appErr := a.requireTaskAccess(user, workspaceID, tableID, taskID); appErr != nil {
		return false, appErr
	}

	tableName, isSingleSelect, metaErr := a.Store.Workspace.GetTableMeta(tableID)
	if metaErr != nil {
		tlog.Errorw("Failed to get table meta for delete",
			"workspace_id", workspaceID, "table_id", tableID, "error", metaErr)
		return false, model.NewAppError("workspace.task_delete_failed", http.StatusInternalServerError)
	}

	var cascades []model.DeleteCascade
	if isSingleSelect {
		parentTable, fieldName, e := a.Store.Workspace.GetSingleSelectCascadeInfo(tableID)
		if e != nil {
			tlog.Warnw("Failed to get single-select cascade info", "table_id", tableID, "error", e)
		} else if parentTable != "" && fieldName != "" {
			cascades = append(cascades, model.DeleteCascade{TableName: parentTable, FieldName: fieldName})
		}
	} else {
		// As parent side (parent_table_item_id): junction tables where this table's items are linked TO
		parentJunctions, e := a.Store.Workspace.GetJunctionTablesForLinkedTable(tableID)
		if e != nil {
			tlog.Warnw("Failed to get parent-side junction tables", "table_id", tableID, "error", e)
		}

		for _, jt := range parentJunctions {
			cascades = append(cascades, model.DeleteCascade{TableName: jt, JunctionColumn: "parent_table_item_id"})
		}
		// As main side (table_item_id): junction tables where this table's items are the linker
		mainJunctions, e2 := a.Store.Workspace.GetJunctionTablesForMainTable(tableID)
		if e2 != nil {
			tlog.Warnw("Failed to get main-side junction tables", "table_id", tableID, "error", e2)
		}

		for _, jt := range mainJunctions {
			cascades = append(cascades, model.DeleteCascade{TableName: jt, JunctionColumn: "table_item_id"})
		}
	}

	deleted := projectChange{kind: "task_deleted", workspaceID: workspaceID, tableID: tableID}
	if isSingleSelect {
		deleted.data = map[string]any{"row_id": taskID}
	} else {
		deleted.task = a.taskRef(tableID, taskID)
	}

	// A subtask is part of its parent's work, so it goes with it.
	taskIDs := []string{taskID}
	var subtasksDeleted []projectChange
	if !isSingleSelect {
		subtasks, err := a.Store.Workspace.GetSubtaskIDs(ctx, tableName, taskID, nil, model.SQLFilter{})
		if err != nil {
			tlog.Errorw("Failed to find the subtasks of a deleted task", "workspace_id", workspaceID, "table_id", tableID, "error", err)
			return false, model.NewAppError("workspace.task_delete_failed", http.StatusInternalServerError)
		}

		// A user who sees only their own tasks may not delete, with theirs,
		// subtasks assigned to others that they cannot see.
		if access := a.assignedOnlyFilter(user, workspaceID, tableID); access.SQL != "" {
			visible, err := a.Store.Workspace.GetSubtaskIDs(ctx, tableName, taskID, nil, access)
			if err != nil {
				tlog.Errorw("Failed to find the subtasks of a deleted task", "workspace_id", workspaceID, "table_id", tableID, "error", err)
				return false, model.NewAppError("workspace.task_delete_failed", http.StatusInternalServerError)
			}

			if len(visible) != len(subtasks) {
				return false, model.NewAppError("workspace.forbidden", http.StatusForbidden)
			}
		}

		for _, id := range subtasks {
			taskIDs = append(taskIDs, id)
			subtasksDeleted = append(subtasksDeleted, projectChange{
				kind:        "task_deleted",
				workspaceID: workspaceID,
				tableID:     tableID,
				task:        a.taskRef(tableID, id),
			})
		}
	}

	ok, err := a.Store.Workspace.DeleteTaskTx(tableName, isSingleSelect, taskIDs, cascades)
	if err != nil {
		tlog.Errorw("Failed to delete task",
			"workspace_id", workspaceID,
			"table_id", tableID,
			"task_id", taskID,
			"error", err,
		)
		return false, model.NewAppError("workspace.task_delete_failed", http.StatusInternalServerError)
	}

	a.publishProjectChange(ctx, deleted)
	for _, change := range subtasksDeleted {
		a.publishProjectChange(ctx, change)
	}

	workspaceTaskName, err := a.Store.Workspace.GetTaskNameByID(workspaceID, tableID, taskID)
	if err != nil {
		tlog.Errorw("Failed to get task name for notification",
			"workspace_id", workspaceID,
			"table_id", tableID,
			"task_id", taskID,
			"error", err,
		)
		return ok, nil
	}

	mainViewID, err := a.Store.Workspace.GetMainViewByTableID(workspaceID, tableID)
	if err != nil {
		tlog.Errorw("Failed to get main view for notification",
			"workspace_id", workspaceID,
			"table_id", tableID,
			"error", err,
		)
		return ok, nil
	}

	a.createProjectNotification(workspaceID, user, model.NOTIFICATION_PROJECT_TASK_DELETED, map[string]interface{}{
		"taskName": *workspaceTaskName,
		"tableID":  tableID,
		"taskID":   taskID,
		"viewID":   mainViewID,
	})

	a.RecordActivity(user.ID, model.AppProjects, model.ActivityTaskDeleted, tableID, taskID, map[string]any{
		"workspaceID": workspaceID,
		"taskName":    *workspaceTaskName,
		"tableID":     tableID,
		"taskID":      taskID,
		"viewID":      mainViewID,
	})

	return ok, nil
}

func (a *App) DeleteTable(ctx context.Context, workspaceID string, tableID string, user model.User) ([]map[string]interface{}, *model.AppError) {
	if !a.ProjectWorkspaceHasPermission(user, workspaceID, model.PermissionDeleteWorkspaceTable) {
		return nil, model.NewAppError("workspace.forbidden", http.StatusForbidden)
	}

	table, err := a.Store.Workspace.DeleteTable(workspaceID, tableID, time.Now().Unix())
	if err != nil {
		tlog.Errorw("Failed to delete table",
			"workspace_id", workspaceID,
			"table_id", tableID,
			"error", err,
		)
		return nil, model.NewAppError("workspace.table_delete_failed", http.StatusInternalServerError)
	}

	a.publishProjectChange(ctx, projectChange{kind: "DELETE_TABLE", workspaceID: workspaceID, tableID: tableID,
		data: map[string]any{"delete_id": tableID}})
	a.publishLinkedFieldsGone(ctx, workspaceID, table)
	return table, nil
}

// GetLinkingTables returns the tables whose link columns would go with a
// table, or with the tables in a folder when folderID is set, so the delete
// can warn about them first.
func (a *App) GetLinkingTables(ctx context.Context, workspaceID, tableID, folderID string, user model.User) ([]model.LinkingTable, *model.AppError) {
	permission := model.PermissionDeleteWorkspaceTable
	if folderID != "" {
		permission = model.PermissionDeleteWorkspaceFolder
	}

	if !a.ProjectWorkspaceHasPermission(user, workspaceID, permission) {
		return nil, model.NewAppError("workspace.forbidden", http.StatusForbidden)
	}

	dbCtx, cancel := a.dbCtx(ctx)
	defer cancel()

	tableIDs := []string{tableID}
	if folderID != "" {
		var err error
		tableIDs, err = a.Store.Workspace.GetFolderTableIDs(dbCtx, workspaceID, folderID)
		if err != nil {
			tlog.Errorw("Failed to get folder tables", "workspace_id", workspaceID, "folder_id", folderID, "error", err)
			return nil, model.NewAppError("workspace.linking_tables_failed", http.StatusInternalServerError)
		}
	}

	tables, err := a.Store.Workspace.GetLinkingTables(dbCtx, workspaceID, tableIDs)
	if err != nil {
		tlog.Errorw("Failed to get linking tables", "workspace_id", workspaceID, "table_id", tableID, "folder_id", folderID, "error", err)
		return nil, model.NewAppError("workspace.linking_tables_failed", http.StatusInternalServerError)
	}

	return tables, nil
}

// publishLinkedFieldsGone tells whoever has another table open that its link
// columns to a deleted table went with it. The tab that deleted it is told
// too, as nothing else takes those columns off its grid.
func (a *App) publishLinkedFieldsGone(ctx context.Context, workspaceID string, linkedFields []map[string]interface{}) {
	ctx = WithOriginClient(ctx, "")
	for _, group := range linkedFields {
		tableID, _ := group["table_id"].(string)
		fieldIDs, _ := group["field_ids"].([]string)
		for _, fieldID := range fieldIDs {
			a.publishProjectChange(ctx, projectChange{kind: "DELETE_FIELD", workspaceID: workspaceID, tableID: tableID,
				data: map[string]any{"field_id": fieldID}})
		}
	}
}

func (a *App) GetWorkspaceCurrentUser(user model.User, workspaceID string) (*model.WorkspaceMember, *model.AppError) {
	workspaceUser, err := a.Store.Workspace.GetUserByUserID(user.ID, workspaceID)
	if err != nil {
		tlog.Errorw("Failed to get workspace user", "workspace_id", workspaceID, "user_id", user.ID, "error", err)
		return nil, model.NewAppError("workspace.current_user_retrieval_failed", http.StatusInternalServerError)
	}

	// Group-only users have no workspace_members row, so build a synthetic one from group roles.
	if workspaceUser == nil {
		groupRoles, _ := a.Store.Workspace.GetGroupRolesForUser(user.ID, workspaceID)
		if len(groupRoles) == 0 {
			return nil, model.NewAppError("workspace.forbidden", http.StatusNotFound)
		}

		workspaceUser = &model.WorkspaceMember{
			UserID:      user.ID,
			WorkspaceID: workspaceID,
			Role:        strings.Join(groupRoles, " "),
		}
	}

	roleNames, _ := a.effectiveWorkspaceRoleNames(user.ID, workspaceID)
	workspaceUser.Role = strings.Join(roleNames, " ")

	roles, err := a.Store.Workspace.GetRolesByName(roleNames, workspaceID)
	if err != nil {
		tlog.Errorw("Failed to get workspace roles", "workspace_id", workspaceID, "user_id", user.ID, "error", err)
		return nil, model.NewAppError("workspace.current_user_retrieval_failed", http.StatusInternalServerError)
	}

	var roleIDs []string
	hasWorkspaceLevelRole := false
	hasPerTableRole := false
	for _, role := range roles {
		workspaceUser.Permissions = append(workspaceUser.Permissions, role.Permissions...)
		roleIDs = append(roleIDs, role.ID)
		if role.PerTableMode {
			hasPerTableRole = true
		} else {
			hasWorkspaceLevelRole = true
		}
	}
	// Per-table gating only applies when user has no workspace-level role.
	// A workspace-level role grants full access regardless of any per-table roles.
	workspaceUser.PerTableMode = hasPerTableRole && !hasWorkspaceLevelRole

	if workspaceUser.PerTableMode {
		tablePerms, err := a.Store.Workspace.GetTablePermissionsForRoles(roleIDs, workspaceID)
		if err != nil {
			tlog.Errorw("Failed to get table permissions", "workspace_id", workspaceID, "user_id", user.ID, "error", err)
			return nil, model.NewAppError("workspace.current_user_retrieval_failed", http.StatusInternalServerError)
		}

		workspaceUser.TablePermissions = tablePerms
	}

	return workspaceUser, nil
}

func (a *App) GetRoleTablePermissions(user model.User, workspaceID, roleID string) ([]model.TablePermission, *model.AppError) {
	if !a.ProjectWorkspaceHasPermission(user, workspaceID, model.PermissionUpdateRoles) {
		return nil, model.NewAppError("workspace.forbidden", http.StatusForbidden)
	}

	if _, err := a.Store.Workspace.GetRoleByID(workspaceID, roleID); err != nil {
		return nil, model.NewAppError("workspace.forbidden", http.StatusForbidden)
	}

	perms, err := a.Store.Workspace.GetTablePermissionsForRole(roleID, workspaceID)
	if err != nil {
		tlog.Errorw("Failed to get table permissions for role", "workspace_id", workspaceID, "role_id", roleID, "error", err)
		return nil, model.NewAppError("workspace.table_permissions_retrieval_failed", http.StatusInternalServerError)
	}

	return perms, nil
}

func (a *App) UpdateRoleTablePermissions(user model.User, workspaceID, roleID string, perms []model.TablePermission) *model.AppError {
	if !a.ProjectWorkspaceHasPermission(user, workspaceID, model.PermissionUpdateRoles) {
		return model.NewAppError("workspace.forbidden", http.StatusForbidden)
	}

	if _, err := a.Store.Workspace.GetRoleByID(workspaceID, roleID); err != nil {
		return model.NewAppError("workspace.forbidden", http.StatusForbidden)
	}

	if err := a.Store.Workspace.UpdateTablePermissionsForRole(roleID, workspaceID, perms); err != nil {
		tlog.Errorw("Failed to set table permissions for role", "workspace_id", workspaceID, "role_id", roleID, "error", err)
		return model.NewAppError("workspace.table_permissions_update_failed", http.StatusInternalServerError)
	}

	return nil
}

// GetWorkspaceTablesForRoleEditor returns all non-linked, non-single-select tables and
// folders for a workspace without any visibility filtering. Requires update_roles permission.
func (a *App) GetWorkspaceTablesForRoleEditor(workspaceID string, user model.User) ([]model.WorkspaceTable, []model.WorkspaceFolder, *model.AppError) {
	if !a.ProjectWorkspaceHasPermission(user, workspaceID, model.PermissionUpdateRoles) {
		return nil, nil, model.NewAppError("workspace.forbidden", http.StatusForbidden)
	}

	all, err := a.Store.Workspace.GetAllTablesBasic(workspaceID)
	if err != nil {
		tlog.Errorw("Failed to get workspace tables for role editor", "workspace_id", workspaceID, "error", err)
		return nil, nil, model.NewAppError("workspace.tables_retrieval_failed", http.StatusInternalServerError)
	}

	tables := make([]model.WorkspaceTable, 0, len(all))
	for _, t := range all {
		if !t.Linked && !t.SingleSelect {
			tables = append(tables, t)
		}
	}

	rawFolders, err := a.Store.Workspace.GetFolders(workspaceID)
	if err != nil {
		tlog.Errorw("Failed to get workspace folders for role editor", "workspace_id", workspaceID, "error", err)
		return nil, nil, model.NewAppError("workspace.details_retrieval_failed", http.StatusInternalServerError)
	}

	folders := buildFolderStructure(rawFolders, tables)
	return tables, folders, nil
}

func (a *App) DeleteWorkspaceAttachment(ctx context.Context, fileID, workspaceID string, user model.User) *model.AppError {
	attachment, appErr := a.GetWorkspaceAttachment(fileID, workspaceID, user)
	if appErr != nil {
		return appErr
	}

	if !a.CanPerformRowAction(user, workspaceID, attachment.TableID, model.PermissionUpdateTask, model.PermissionUpdateTask) {
		return model.NewAppError("workspace.forbidden", http.StatusForbidden)
	}

	fileExt := filepath.Ext(attachment.Name)
	if attachment.StorageID != "" {
		subPath := attachment.WorkspaceID + "/" + attachment.TableID + "/" + attachment.TaskID + "/" + attachment.ID + fileExt
		filePath, err := a.BuildFilePath(attachment.StorageID, subPath, model.AppProjects)
		if err != nil {
			tlog.Errorw("Failed to build project file path for delete", "file_id", fileID, "error", err)
			return model.NewAppError("workspace.attachment_delete_failed", http.StatusInternalServerError)
		}

		backend, exists := a.FileStorageObjects[attachment.StorageID]
		if !exists {
			tlog.Errorw("Storage backend not available for delete", "storage_id", attachment.StorageID)
			return model.NewAppError("storage.not_available", http.StatusInternalServerError)
		}

		if err := backend.RemoveFile(ctx, filePath); err != nil {
			tlog.Errorw("Failed to delete file from storage", "filepath", filePath, "error", err)
			return model.NewAppError("workspace.attachment_delete_failed", http.StatusInternalServerError)
		}
	} else {
		filePath := path.Join("data", "projects",
			attachment.WorkspaceID,
			attachment.TableID,
			attachment.TaskID,
			attachment.ID+fileExt)
		if _, err := os.Stat(filePath); err == nil {
			if err := os.Remove(filePath); err != nil {
				tlog.Errorw("Failed to delete file from storage", "filepath", filePath, "error", err)
				return model.NewAppError("workspace.attachment_delete_failed", http.StatusInternalServerError)
			}
		} else if os.IsNotExist(err) {
			tlog.Warnw("File not found in storage, deleting DB record", "filepath", filePath)
		} else {
			tlog.Errorw("Failed to check file existence", "filepath", filePath, "error", err)
			return model.NewAppError("workspace.attachment_delete_failed", http.StatusInternalServerError)
		}
	}

	err := a.Store.Workspace.DeleteAttachment(fileID, workspaceID)
	if err != nil {
		tlog.Errorw("Failed to delete workspace attachment", "file_id", fileID, "workspace_id", workspaceID, "error", err)
		return model.NewAppError("workspace.attachment_delete_failed", http.StatusInternalServerError)
	}

	if attachment.TableID != "" && attachment.TaskID != "" {
		a.publishTaskRow(ctx, "task_updated", workspaceID, attachment.TableID, attachment.TaskID)
	}

	return nil
}

func (a *App) GetWorkspaceAttachment(fileID, workspaceID string, user model.User) (*model.WorkspaceAttachment, *model.AppError) {
	isMember, appErr := a.IsWorkspaceMember(workspaceID, user.ID)
	if appErr != nil {
		return nil, appErr
	}

	if !isMember {
		return nil, model.NewAppError("workspace.forbidden", http.StatusForbidden)
	}

	attachment, err := a.Store.Workspace.GetAttachment(fileID, workspaceID)
	if err != nil {
		tlog.Errorw("Failed to get workspace attachment", "file_id", fileID, "workspace_id", workspaceID, "error", err)
		return nil, model.NewAppError("workspace.attachment_retrieval_failed", http.StatusInternalServerError)
	}

	if attachment == nil {
		return nil, model.NewAppError("workspace.attachment_not_found", http.StatusNotFound)
	}

	if appErr := a.requireTaskAccess(user, workspaceID, attachment.TableID, attachment.TaskID); appErr != nil {
		return nil, appErr
	}

	return attachment, nil
}

func (a *App) SaveGridSort(workspace_id, table_id, view_id, sortData string, user model.User) (bool, *model.AppError) {
	if appErr := a.requireViewAccess(context.Background(), user.ID, view_id); appErr != nil {
		return false, appErr
	}

	if !a.ProjectWorkspaceHasPermission(user, workspace_id, model.PermissionUpdateWorkspaceView) {
		return false, model.NewAppError("workspace.forbidden", http.StatusForbidden)
	}

	success, err := a.Store.Workspace.UpdateGridSort(workspace_id, table_id, view_id, sortData, model.NewID())
	if err != nil {
		tlog.Errorw("Failed to save grid sort", "workspace_id", workspace_id, "table_id", table_id, "view_id", view_id, "error", err)
		return false, model.NewAppError("workspace.grid_sort_save_failed", http.StatusInternalServerError)
	}

	return success, nil
}

func (a *App) DeleteProjectWorkspace(ctx context.Context, workspaceID string, user model.User) (bool, *model.AppError) {
	if !a.holdsProjectWorkspaceAdmin(user, workspaceID) {
		return false, model.NewAppError("workspace.forbidden", http.StatusForbidden)
	}

	return a.deleteProjectWorkspace(ctx, workspaceID, user)
}

func (a *App) deleteProjectWorkspace(ctx context.Context, workspaceID string, user model.User) (bool, *model.AppError) {
	workspace, err := a.Store.Workspace.Delete(workspaceID)
	if err != nil {
		tlog.Errorw("Failed to delete workspace", "workspace_id", workspaceID, "error", err)
		return false, model.NewAppError("workspace.delete_failed", http.StatusInternalServerError)
	}

	a.publishProjectChange(ctx, projectChange{kind: "DELETE_WORKSPACE", workspaceID: workspaceID})

	workspaceName, err := a.Store.Workspace.GetNameByID(workspaceID)
	if err != nil {
		tlog.Errorw("Failed to get workspace name for notification", "workspace_id", workspaceID, "error", err)
		return workspace, nil
	}

	a.createProjectNotification(workspaceID, user, model.NOTIFICATION_PROJECT_DELETED, map[string]interface{}{
		"workspaceName": *workspaceName,
		"workspaceID":   workspaceID,
	})

	a.RecordActivity(user.ID, model.AppProjects, model.ActivityProjectDeleted, "workspace", workspaceID, map[string]any{
		"workspaceID":   workspaceID,
		"workspaceName": *workspaceName,
	})

	return workspace, nil
}

func (a *App) IsWorkspaceMember(workspaceID, userID string) (bool, *model.AppError) {
	isMember, err := a.Store.Workspace.IsMember(workspaceID, userID)
	if err != nil {
		tlog.Errorw("Failed to check workspace membership", "workspace_id", workspaceID, "user_id", userID, "error", err)
		return false, model.NewAppError("workspace.membership_check_failed", http.StatusInternalServerError)
	}

	return isMember, nil
}

// CreateWorkspaceTableField adds a field. The names given are only what it is
// called; its column, and the linked table's for a two-way link, are named
// from the field's own id.
func (a *App) CreateWorkspaceTableField(ctx context.Context, workspaceID string, tableID string, fieldType string, linkedTableID string, user model.User, selectedType string, linkBothDirections bool, fieldNameDisplay string, fieldNameInSecondTableDisplay string, headerUsage string, formula *model.FormulaSpec) (*model.WorkspaceHeaders, *model.AppError) {
	if !a.ProjectWorkspaceHasPermission(user, workspaceID, model.PermissionUpdateWorkspaceTable) &&
		!a.CanPerformRowAction(user, workspaceID, tableID, model.PermissionCreateFields, model.PermissionCreateFields) {
		return nil, model.NewAppError("workspace.forbidden", http.StatusForbidden)
	}

	fieldNameDisplay = strings.TrimSpace(fieldNameDisplay)
	if fieldNameDisplay == "" {
		return nil, model.NewAppError("workspace.field_name_empty", http.StatusBadRequest)
	}

	var sqlFieldType string
	switch fieldType {
	case "single select", "text", "assignee", "link", "calculations", "master link", "url", "file":
		sqlFieldType = "TEXT"
	case "bool":
		sqlFieldType = "BOOLEAN"
	case "number":
		sqlFieldType = "NUMERIC"
	case "date":
		sqlFieldType = "DATE"
	case "decimal":
		sqlFieldType = "DECIMAL(20,2)"
	default:
		return nil, model.NewAppError("workspace.field_type_invalid", http.StatusBadRequest)
	}

	switch selectedType {
	case "", "single select", "text", "assignee", "link", "calculations", "master link", "url", "file", "bool", "number", "date", "decimal":
	default:
		return nil, model.NewAppError("workspace.field_type_invalid", http.StatusBadRequest)
	}

	if fieldType == "link" && linkedTableID == "" {
		return nil, model.NewAppError("workspace.link_table_id_empty", http.StatusBadRequest)
	}

	// Only a link points at another table. Given to a single select, the
	// table would be saved as its options' parent.
	if fieldType != "link" {
		linkedTableID = ""
	}

	if linkedTableID != "" {
		if appErr := a.CheckTableInWorkspace(ctx, workspaceID, linkedTableID); appErr != nil {
			return nil, appErr
		}
	}

	fieldNameInSecondTableDisplay = strings.TrimSpace(fieldNameInSecondTableDisplay)
	if linkBothDirections && fieldNameInSecondTableDisplay == "" {
		return nil, model.NewAppError("workspace.field_name_second_invalid", http.StatusBadRequest)
	}

	mainFieldID := model.NewID()
	secondFieldID := model.NewID()
	fieldName := physicalName("c", mainFieldID)
	fieldNameInSecondTable := physicalName("c", secondFieldID)

	workspaceTableField, err := a.Store.Workspace.CreateTableField(model.CreateFieldParams{
		WorkspaceID:                   workspaceID,
		TableID:                       tableID,
		FieldName:                     fieldName,
		FieldType:                     fieldType,
		SQLFieldType:                  sqlFieldType,
		LinkedTableID:                 linkedTableID,
		UserID:                        user.ID,
		SelectedType:                  selectedType,
		LinkBothDirections:            linkBothDirections,
		FieldNameInSecondTable:        fieldNameInSecondTable,
		FieldNameDisplay:              fieldNameDisplay,
		FieldNameInSecondTableDisplay: fieldNameInSecondTableDisplay,
		HeaderUsage:                   headerUsage,
		Formula:                       formula,
		MainFieldID:                   mainFieldID,
		SecondFieldID:                 secondFieldID,
		LinkTableID:                   model.NewID(),
		SecondLinkID:                  model.NewID(),
		LinkTablePhysicalName:         "t" + strings.ReplaceAll(model.NewID(), "-", ""),
		SecondLinkTablePhysicalName:   "t" + strings.ReplaceAll(model.NewID(), "-", ""),
		SingleSelectTableID:           model.NewID(),
		SingleSelectPhysicalName:      "t" + strings.ReplaceAll(model.NewID(), "-", ""),
		LinkViewID:                    model.NewID(),
		SecondLinkViewID:              model.NewID(),
	})
	if err != nil {
		tlog.Errorw("Failed to create workspace table field", "workspace_id", workspaceID, "table_id", tableID, "error", err)
		return nil, model.NewAppError("workspace.table_field_create_failed", http.StatusInternalServerError)
	}

	if err := a.Store.Workspace.UpdateTaskOrderGridField(workspaceID, tableID, *workspaceTableField); err != nil {
		tlog.Errorw("Failed to update grid field order", "workspace_id", workspaceID, "table_id", tableID, "error", err)
		return nil, model.NewAppError("workspace.table_field_create_failed", http.StatusInternalServerError)
	}

	a.publishProjectChange(ctx, projectChange{kind: "NEW_FIELD", workspaceID: workspaceID, tableID: tableID,
		data: map[string]any{
			"fieldData":              workspaceTableField,
			"fieldName":              fieldName,
			"fieldType":              fieldType,
			"linkedTableID":          linkedTableID,
			"fieldNameInSecondTable": fieldNameInSecondTable,
			"linkBothDirections":     linkBothDirections,
			"formula":                formula,
		}})

	return workspaceTableField, nil
}

// assignedOnlyTables maps tableID → true for tables that should show only rows assigned to the user.
// va holds view-access state: inPerTableMode, per-table view flags, and the union of all
// explicitly-allowed view IDs (role view:{id} perms + direct user shares).
// navOnly skips the rows of task tables; use for navigation/sidebar where only
// table structure (headers, views, options, filters) is needed. Otherwise each
// task table sends one page of its tasks. Option tables are small and
// always sent whole; link tables never send their rows.
// viewVisibleTo reports whether a user is shown a view: a public one, or a
// private one they made or that was shared with them. A private view is
// theirs alone, so roles that manage views or open its type do not show it
// to anyone else.
func viewVisibleTo(v model.WorkspaceView, userID string, sharedViews map[string]bool) bool {
	return v.IsPublic || v.CreatedBy == userID || sharedViews[v.ID]
}

// requireViewAccess answers for a private view the user is not shown as
// though it did not exist, so its id cannot open what the listing hides.
func (a *App) requireViewAccess(ctx context.Context, userID, viewID string) *model.AppError {
	if viewID == "" {
		return nil
	}

	isPublic, createdBy, err := a.Store.Workspace.GetViewVisibility(ctx, viewID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}

	if err != nil {
		tlog.Errorw("Failed to check the view's visibility", "view_id", viewID, "error", err)
		return model.NewAppError("workspace.view_access_failed", http.StatusInternalServerError)
	}

	if isPublic || createdBy == userID {
		return nil
	}

	shared, err := a.Store.Workspace.GetViewSharedUserIDs(viewID)
	if err != nil {
		tlog.Errorw("Failed to check who the view is shared with", "view_id", viewID, "error", err)
		return model.NewAppError("workspace.view_access_failed", http.StatusInternalServerError)
	}

	if slices.Contains(shared, userID) {
		return nil
	}

	return model.NewAppError("workspace.view_or_field_not_found", http.StatusNotFound)
}

func (a *App) hydrateWorkspaceTables(workspaceID, userID string, assignedOnlyTables map[string]bool, va interfaces.WorkspaceAccess, navOnly bool, pageNum, pageLimit int) ([]model.WorkspaceTable, error) {
	rawTables, err := a.Store.Workspace.GetTableList(workspaceID, userID)
	if err != nil {
		return nil, err
	}

	tableIDs := make([]string, len(rawTables))
	for i, rt := range rawTables {
		tableIDs[i] = rt.ID
	}

	metadata, err := a.Store.Workspace.GetTableViewMetadata(workspaceID, userID, tableIDs)
	if err != nil {
		return nil, err
	}

	sharedIDs, err := a.Store.Workspace.GetSharedViewIDsForUser(userID, workspaceID)
	if err != nil {
		return nil, err
	}

	sharedViews := make(map[string]bool, len(sharedIDs))
	for _, id := range sharedIDs {
		sharedViews[id] = true
	}

	tables := make([]model.WorkspaceTable, len(rawTables))
	var wg sync.WaitGroup
	errChan := make(chan error, len(rawTables))
	sem := make(chan struct{}, 5)

	for i, rt := range rawTables {
		wg.Add(1)
		go func(i int, rt model.WorkspaceTable) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			table := rt
			rawName := rt.Name
			if idx := strings.Index(rawName, "_"); idx >= 0 {
				table.Name = rawName[idx+1:]
			}

			if (navOnly || rt.ParentTableID != "") && !rt.SingleSelect {
				table.DataBase = []map[string]any{}
				if colTypes, err := a.Store.Workspace.GetTableColumnTypes(rawName); err == nil {
					table.Headers = a.buildTableHeaders(rt.ID, colTypes, true, true)
				} else {
					table.Headers = []model.WorkspaceHeaders{}
				}
			} else if !rt.SingleSelect {
				// Paginated path: fetch only needed rows from DB using SQL-level filter.
				colTypes, err := a.Store.Workspace.GetTableColumnTypes(rawName)
				if err != nil {
					errChan <- err
					return
				}

				headers := a.buildTableHeaders(rt.ID, colTypes, true, true)

				sortSQL := "ORDER BY main.id ASC"
				offset := (pageNum - 1) * pageLimit
				items, totalCount, rootCount, err := a.Store.Workspace.GetRootTasksPagedWithSubtasks(context.Background(), rawName, model.SQLFilter{}, model.SQLFilter{}, sortSQL, pageLimit, offset, true)
				if err != nil {
					errChan <- err
					return
				}

				if va.AssignedOnly(table.ID) {
					items = filterByAssignee(items, userID)
				}

				a.hydrateLinkedAndSingleSelect(context.Background(), items, headers)
				a.hideUnseenLinks(context.Background(), items, headers, userID, va.AssignedOnly)
				if items == nil {
					items = []map[string]any{}
				}

				table.Total = totalCount
				table.RootTotal = rootCount
				table.DataBase = items
				table.Headers = headers
			} else {
				// An option table's rows, such as the statuses a task picks
				// from, are assigned to no one, so assigned-only leaves them all.
				items, headers, err := a.getTableItems(rawName, rt.ID)
				if err != nil {
					errChan <- err
					return
				}

				sortItemsByDueDate(items)
				if items == nil {
					items = []map[string]any{}
				}

				if headers == nil {
					headers = []model.WorkspaceHeaders{}
				}

				table.Total = len(items)
				table.DataBase = items
				table.Headers = headers
			}

			meta := metadata[rt.ID]
			table.Views = meta.Views
			table.Options = meta.GridSettings

			// Table visibility is enforced by filterTablesByPerTableMode after hydration.
			{
				filtered := table.Views[:0]
				for _, v := range table.Views {
					if viewVisibleTo(v, userID, sharedViews) {
						if v.ViewType == "kanban" {
							v.TaskOrder = withoutPlacedCards(v.TaskOrder)
						}

						filtered = append(filtered, v)
					}
				}

				table.Views = filtered
			}

			filters := meta.SavedFilters
			for i := range filters {
				var payload model.FilterPayload
				if jsonErr := json.Unmarshal([]byte(filters[i].FilterSettings), &payload); jsonErr == nil {
					filters[i].Filters = &payload
				} else {
					filters[i].Filters = &model.FilterPayload{}
				}
			}

			table.Filters = filters

			tables[i] = table
		}(i, rt)
	}

	wg.Wait()
	close(errChan)
	for err := range errChan {
		return nil, err
	}

	return tables, nil
}

// A "*" key means workspace-level assigned-only (applies to all tables).

func (a *App) buildAccessInput(user model.User, workspaceID string) interfaces.AccessInput {
	roleNames, ok := a.effectiveWorkspaceRoleNames(user.ID, workspaceID)
	if !ok {
		return interfaces.AccessInput{}
	}

	roles, err := a.Store.Workspace.GetRolesByName(roleNames, workspaceID)
	if err != nil {
		return interfaces.AccessInput{}
	}

	var roleIDs []string
	anyPerTableMode := false
	for _, r := range roles {
		roleIDs = append(roleIDs, r.ID)
		if r.PerTableMode {
			anyPerTableMode = true
		}
	}

	var tablePerms []model.TablePermission
	if anyPerTableMode {
		tablePerms, _ = a.Store.Workspace.GetTablePermissionsForRoles(roleIDs, workspaceID)
	}

	return interfaces.AccessInput{Roles: roles, TablePerms: tablePerms}
}

// buildAccessInputs builds what buildAccessInput builds for each of the
// users, reading every user's roles, table permissions and shared views
// together, so the number of queries does not grow with the users.
func (a *App) buildAccessInputs(ctx context.Context, users []model.User, workspaceID string) map[string]interfaces.AccessInput {
	inputs := make(map[string]interfaces.AccessInput, len(users))
	var ids []string
	for _, u := range users {
		inputs[u.ID] = interfaces.AccessInput{}
		ids = append(ids, u.ID)
	}

	if len(ids) == 0 {
		return inputs
	}

	roleNames, err := a.Store.Workspace.GetRoleNamesForUsers(ctx, workspaceID, ids)
	if err != nil {
		tlog.Errorw("Failed to load workspace roles", "workspace_id", workspaceID, "error", err)
		return inputs
	}

	seen := map[string]bool{}
	var allNames []string
	for _, names := range roleNames {
		for _, n := range names {
			if !seen[n] {
				seen[n] = true
				allNames = append(allNames, n)
			}
		}
	}

	roles, err := a.Store.Workspace.GetRolesByName(allNames, workspaceID)
	if err != nil {
		tlog.Errorw("Failed to load workspace roles", "workspace_id", workspaceID, "error", err)
		return inputs
	}

	userRoles := make(map[string][]model.ProjectWorkspaceRole, len(roleNames))
	perTableRoleIDs := map[string]bool{}
	for userID, names := range roleNames {
		has := make(map[string]bool, len(names))
		for _, n := range names {
			has[n] = true
		}

		anyPerTableMode := false
		for _, r := range roles {
			if has[r.Name] {
				userRoles[userID] = append(userRoles[userID], r)
				anyPerTableMode = anyPerTableMode || r.PerTableMode
			}
		}

		if anyPerTableMode {
			for _, r := range userRoles[userID] {
				perTableRoleIDs[r.ID] = true
			}
		}
	}

	var tablePerms []model.TablePermission
	if len(perTableRoleIDs) > 0 {
		tablePerms, _ = a.Store.Workspace.GetTablePermissionsForRoles(slices.Collect(maps.Keys(perTableRoleIDs)), workspaceID)
	}

	for userID := range roleNames {
		rs := userRoles[userID]
		in := interfaces.AccessInput{Roles: rs}

		perTable := false
		roleIDs := make(map[string]bool, len(rs))
		for _, r := range rs {
			roleIDs[r.ID] = true
			perTable = perTable || r.PerTableMode
		}

		if perTable {
			for _, p := range tablePerms {
				if roleIDs[p.RoleID] {
					in.TablePerms = append(in.TablePerms, p)
				}
			}
		}

		inputs[userID] = in
	}

	return inputs
}

type openAccess struct{}

func (openAccess) VisibleTables(tables []model.WorkspaceTable) []model.WorkspaceTable { return tables }
func (openAccess) AssignedOnlyAll() bool                                              { return false }
func (openAccess) AssignedOnly(_ string) bool                                         { return false }
func (openAccess) CanPerformRowAction(_, _, _ string) bool                            { return true }
func (openAccess) TableHasPermission(_, _ string) bool                                { return true }
func (a *App) resolveAccess(in interfaces.AccessInput) interfaces.WorkspaceAccess {
	if a.WorkspaceRoles != nil {
		return a.WorkspaceRoles.Resolve(in)
	}

	return openAccess{}
}

func (a *App) buildAssignedOnlyMapFromInput(_ model.User, _ string, in interfaces.AccessInput) map[string]bool {
	access := a.resolveAccess(in)
	result := map[string]bool{}
	if access.AssignedOnlyAll() {
		result["*"] = true
	}
	// Per-table assigned-only is queried via access.AssignedOnly() at each table.
	return result
}

func (a *App) buildViewAccessStateFromInput(_ model.User, in interfaces.AccessInput) interfaces.WorkspaceAccess {
	return a.resolveAccess(in)
}

// filterTablesByPerTableModeFromInput removes tables the user cannot see when
// any of their roles has per_table_mode=true. No-op for system admins or when
// all roles are workspace-level (backward compat: all tables remain visible).
func (a *App) filterTablesByPerTableModeFromInput(in interfaces.AccessInput, tables []model.WorkspaceTable) []model.WorkspaceTable {
	return a.resolveAccess(in).VisibleTables(tables)
}

func (a *App) GetWorkspaceTables(workspaceID string, user model.User, data string, pageNum, pageLimit int) ([]model.WorkspaceTable, *model.AppError) {
	// Only the navigation form or one page of each table: the rows of a whole
	// workspace are too large to send.
	const maxWorkspaceTablesPage = 500
	navOnly := data == "nav"
	if !navOnly && pageLimit <= 0 {
		return nil, model.NewAppError("request.invalid", http.StatusBadRequest)
	}

	pageLimit = min(pageLimit, maxWorkspaceTablesPage)
	pageNum = max(pageNum, 1)

	in := a.buildAccessInput(user, workspaceID)
	assignedOnlyTables := a.buildAssignedOnlyMapFromInput(user, workspaceID, in)
	va := a.buildViewAccessStateFromInput(user, in)

	tables, err := a.hydrateWorkspaceTables(workspaceID, user.ID, assignedOnlyTables, va, navOnly, pageNum, pageLimit)
	if err != nil {
		tlog.Errorw("Failed to get workspace tables", "workspace_id", workspaceID, "error", err)
		return nil, model.NewAppError("workspace.tables_retrieval_failed", http.StatusInternalServerError)
	}

	tables = a.filterTablesByPerTableModeFromInput(in, tables)
	return tables, nil
}

func (a *App) CreateProjectWorkspace(userID, name, description string) (*model.Workspace, *model.AppError) {
	if strings.TrimSpace(name) == "" {
		return nil, model.NewAppError("workspace.workspace_name_empty", http.StatusBadRequest)
	}

	prefix, err := a.Store.Workspace.GenerateUniquePrefix()
	if err != nil {
		tlog.Errorw("Failed to generate workspace prefix", "user_id", userID, "error", err)
		return nil, model.NewAppError("workspace.create_failed", http.StatusInternalServerError)
	}

	workspaceID := model.NewID()

	defaultRoles := model.MakeDefaultProjectWorkspaceRoles()

	adminRole := *defaultRoles[model.ProjectWorkspaceAdminRoleID]
	adminRole.ID = model.NewID()
	adminRole.WorkspaceID = workspaceID
	adminRole.DisplayName = "Admin"
	adminRole.Description = "Administrator with full access to manage workspace, members, and all features"

	userRole := *defaultRoles[model.ProjectWorkspaceUserRoleID]
	userRole.ID = model.NewID()
	userRole.WorkspaceID = workspaceID
	userRole.DisplayName = "User"
	userRole.Description = "Regular user with basic permissions to create and update tasks"

	member := model.WorkspaceMember{
		ID:          model.NewID(),
		UserID:      userID,
		WorkspaceID: workspaceID,
		Role:        model.ProjectWorkspaceAdminRoleID,
		DateJoined:  time.Now().Unix(),
	}

	if err := a.Store.Workspace.CreateTx(workspaceID, userID, name, description, prefix, adminRole, userRole, member); err != nil {
		tlog.Errorw("Failed to create project workspace", "user_id", userID, "error", err)
		return nil, model.NewAppError("workspace.create_failed", http.StatusInternalServerError)
	}

	workspace, err := a.Store.Workspace.GetRow(workspaceID)
	if err != nil {
		tlog.Errorw("Failed to fetch created workspace", "workspace_id", workspaceID, "error", err)
		return nil, model.NewAppError("workspace.create_failed", http.StatusInternalServerError)
	}

	workspace.UserPermissions = adminRole.Permissions
	workspace.UserRoles = []string{adminRole.DisplayName}
	workspace.CanDelete = true
	if members, err := a.Store.Workspace.GetMembers(workspaceID); err == nil {
		workspace.Members = members
	}

	return workspace, nil
}

func (a *App) GetTaskComments(ctx context.Context, workspaceID string, tableID string, taskID string, user model.User) ([]model.TaskComment, *model.AppError) {
	if !a.canReadTable(ctx, user, workspaceID, tableID) {
		return nil, model.NewAppError("workspace.forbidden", http.StatusForbidden)
	}

	if appErr := a.requireTaskAccess(user, workspaceID, tableID, taskID); appErr != nil {
		return nil, appErr
	}
	// Comments are stored by task id alone, so the task must be shown to be in
	// the table the caller may read.
	tableName, err := a.Store.Workspace.GetTableName(tableID)
	if err != nil {
		tlog.Errorw("Failed to get table name for task comments", "table_id", tableID, "error", err)
		return nil, model.NewAppError("workspace.comments_retrieval_failed", http.StatusInternalServerError)
	}

	task, err := a.Store.Workspace.GetTableRowByID(tableName, taskID)
	if err != nil {
		tlog.Errorw("Failed to get task for comments", "table_id", tableID, "task_id", taskID, "error", err)
		return nil, model.NewAppError("workspace.comments_retrieval_failed", http.StatusInternalServerError)
	}

	if task == nil {
		return nil, model.NewAppError("workspace.forbidden", http.StatusForbidden)
	}

	comments, err := a.Store.Workspace.GetComments(taskID)
	if err != nil {
		tlog.Errorw("Failed to get task comments", "workspace_id", workspaceID, "table_id", tableID, "task_id", taskID, "error", err)
		return nil, model.NewAppError("workspace.comments_retrieval_failed", http.StatusInternalServerError)
	}

	activities, err := a.Store.Workspace.GetTaskActivities(taskID)
	if err != nil {
		tlog.Errorw("Failed to get task activities", "workspace_id", workspaceID, "table_id", tableID, "task_id", taskID, "error", err)
		return nil, model.NewAppError("workspace.comments_retrieval_failed", http.StatusInternalServerError)
	}

	for i := range activities {
		if err := json.Unmarshal([]byte(activities[i].ParametersJSON), &activities[i].Parameters); err != nil {
			activities[i].Parameters = make(map[string]interface{})
		}
	}

	for _, activity := range activities {
		var addedUserID string
		if addedUser, ok := activity.Parameters["AddedUser"].(string); ok {
			addedUserID = addedUser
		}

		relevantUserID := activity.UserID
		displayName := activity.UserName

		if activity.Type == model.ActivityTaskAssigned || activity.Type == model.ActivityTaskStatusChanged || activity.Type == model.ActivityProjectInvite {
			if addedUserID != "" {
				relevantUserID = addedUserID
				displayName = "Unknown User"
			}
		}

		comments = append(comments, model.TaskComment{
			ID:           activity.ID,
			ItemID:       activity.ItemID,
			UserID:       activity.UserID,
			AffectedUser: relevantUserID,
			CreatedAt:    activity.CreatedAt,
			ActivityType: activity.Type,
			Content:      fmt.Sprintf("Activity: %s by %s (App: %s)", activity.Type, displayName, activity.App),
		})
	}

	return comments, nil
}

func (a *App) AddTaskComment(ctx context.Context, workspaceID string, tableID string, taskID string, comment string, user model.User, mentionedUsers []string) (*model.TaskComment, *model.AppError) {
	if !a.CanPerformRowAction(user, workspaceID, tableID, model.PermissionUpdateTask, model.PermissionUpdateTask) {
		return nil, model.NewAppError("workspace.forbidden", http.StatusForbidden)
	}

	if appErr := a.requireTaskAccess(user, workspaceID, tableID, taskID); appErr != nil {
		return nil, appErr
	}

	if runes := utf8.RuneCountInString(comment); runes > model.MaxTaskCommentRunes {
		tlog.Warnw("Rejected oversized task comment",
			"user_id", user.ID,
			"workspace_id", workspaceID,
			"task_id", taskID,
			"rune_count", runes,
		)
		return nil, model.NewAppError("workspace.comment_too_long", http.StatusBadRequest)
	}

	taskComment, err := a.Store.Workspace.CreateTaskComment(workspaceID, tableID, taskID, comment, user.ID, model.NewID())
	if err != nil {
		tlog.Errorw("Failed to add task comment", "workspace_id", workspaceID, "table_id", tableID, "task_id", taskID, "error", err)
		return nil, model.NewAppError("workspace.comment_add_failed", http.StatusInternalServerError)
	}

	a.publishProjectChange(ctx, projectChange{
		kind:        "comment_added",
		workspaceID: workspaceID,
		tableID:     tableID,
		task:        a.taskRef(tableID, taskID),
		data:        map[string]any{"task_id": taskID, "comment": taskComment},
	})

	mainViewID, err := a.Store.Workspace.GetMainViewByTableID(workspaceID, tableID)
	if err != nil {
		tlog.Errorw("Failed to get main view for notification", "workspace_id", workspaceID, "table_id", tableID, "error", err)
		return taskComment, nil
	}

	workspaceTaskName, err := a.Store.Workspace.GetTaskNameByID(workspaceID, tableID, taskID)
	if err != nil {
		tlog.Errorw("Failed to get task name for notification", "workspace_id", workspaceID, "table_id", tableID, "task_id", taskID, "error", err)
		return taskComment, nil
	}

	if len(mentionedUsers) > 0 {
		for _, mentionedUserID := range mentionedUsers {
			if mentionedUserID == user.ID {
				continue
			}

			isMember, _ := a.Store.Workspace.IsMember(workspaceID, mentionedUserID)
			if !isMember {
				continue
			}

			a.createProjectNotification(workspaceID, user, model.NOTIFICATION_TASK_COMMENT, map[string]interface{}{
				"userID":      mentionedUserID,
				"tableID":     tableID,
				"taskID":      taskID,
				"viewID":      mainViewID,
				"workspaceID": workspaceID,
				"taskName":    *workspaceTaskName,
				"isMention":   true,
				"comment":     comment,
			})
		}
	}

	return taskComment, nil
}

// buildFolderStructure nests the folders and puts each table in its folder.
// A folder whose parent is gone goes to the top, so it stays reachable.
func buildFolderStructure(rawFolders []model.WorkspaceFolder, tables []model.WorkspaceTable) []model.WorkspaceFolder {
	live := make(map[string]bool, len(rawFolders))
	for _, f := range rawFolders {
		live[f.ID] = true
	}

	tablesByFolder := make(map[string][]model.WorkspaceTable)
	for _, table := range tables {
		if live[table.FolderID] {
			tablesByFolder[table.FolderID] = append(tablesByFolder[table.FolderID], table)
		}
	}

	byParent := make(map[string][]model.WorkspaceFolder)
	for _, f := range rawFolders {
		parent := f.ParentFolderID
		if !live[parent] {
			parent = ""
		}

		byParent[parent] = append(byParent[parent], f)
	}

	var build func(parent string) []model.WorkspaceFolder
	build = func(parent string) []model.WorkspaceFolder {
		folders := make([]model.WorkspaceFolder, 0, len(byParent[parent]))
		for _, f := range byParent[parent] {
			f.Tables = append(f.Tables, tablesByFolder[f.ID]...)
			f.Children = build(f.ID)
			folders = append(folders, f)
		}

		return folders
	}

	return build("")
}

func (a *App) GetProjectWorkspaceDetails(user model.User, workspaceID string) (*model.Workspace, *model.AppError) {
	workspace, err := a.Store.Workspace.GetRow(workspaceID)
	if err != nil {
		tlog.Errorw("Failed to get workspace row", "workspace_id", workspaceID, "error", err)
		return nil, model.NewAppError("workspace.details_retrieval_failed", http.StatusInternalServerError)
	}

	in := a.buildAccessInput(user, workspaceID)
	assignedOnlyTables := a.buildAssignedOnlyMapFromInput(user, workspaceID, in)
	va := a.buildViewAccessStateFromInput(user, in)

	var (
		members               []model.WorkspaceMember
		tables                []model.WorkspaceTable
		membersErr, tablesErr error
		wg                    sync.WaitGroup
	)

	// Group members are NOT included here. Merging thousands of group members
	// into workspace.Members caused performance issues (6000+ reactive objects,
	// connection pool saturation). The assignee picker uses /members/search instead.
	wg.Add(2)
	go func() {
		defer wg.Done()
		m, e := a.Store.Workspace.GetMembers(workspaceID)
		if e != nil {
			membersErr = e
			return
		}

		members = m
	}()
	go func() {
		defer wg.Done()
		tables, tablesErr = a.hydrateWorkspaceTables(workspaceID, user.ID, assignedOnlyTables, va, true, 0, 0)
	}()
	wg.Wait()

	if membersErr != nil {
		tlog.Errorw("Failed to get workspace members", "workspace_id", workspaceID, "error", membersErr)
		return nil, model.NewAppError("workspace.details_retrieval_failed", http.StatusInternalServerError)
	}

	if tablesErr != nil {
		tlog.Errorw("Failed to get workspace tables", "workspace_id", workspaceID, "error", tablesErr)
		return nil, model.NewAppError("workspace.details_retrieval_failed", http.StatusInternalServerError)
	}

	// Filter tables by per-table view permission (uses merged direct+group roles).
	tables = a.filterTablesByPerTableModeFromInput(in, tables)
	{
	}

	workspace.Members = members
	workspace.Tables = tables

	rawFolders, err := a.Store.Workspace.GetFolders(workspaceID)
	if err != nil {
		tlog.Errorw("Failed to get workspace folders", "workspace_id", workspaceID, "error", err)
		return nil, model.NewAppError("workspace.details_retrieval_failed", http.StatusInternalServerError)
	}

	folders := buildFolderStructure(rawFolders, tables)

	liveFolders := make(map[string]bool, len(rawFolders))
	for _, f := range rawFolders {
		liveFolders[f.ID] = true
	}

	workspaceFolder := model.WorkspaceFolder{
		ID:       workspace.ID,
		Name:     workspace.Title,
		IsFolder: false,
		Children: folders,
		Tables:   []model.WorkspaceTable{},
	}
	for _, table := range tables {
		if liveFolders[table.FolderID] {
			continue
		}

		workspaceFolder.Tables = append(workspaceFolder.Tables, table)
	}

	workspace.Folders = []model.WorkspaceFolder{workspaceFolder}

	return workspace, nil
}

// GetProjectWorkspaces lists the user's workspaces with what their cards and
// the sidebar show, in the same few queries however many workspaces there are.
func (a *App) GetProjectWorkspaces(ctx context.Context, user model.User) ([]model.Workspace, *model.AppError) {
	failed := model.NewAppError("workspace.workspaces_retrieval_failed", http.StatusInternalServerError)

	workspaces, err := a.Store.Workspace.GetAllForUser(user.ID)
	if err != nil {
		tlog.Errorw("Failed to get project workspaces", "user_id", user.ID, "error", err)
		return nil, failed
	}

	if len(workspaces) == 0 {
		return workspaces, nil
	}

	ctx, cancel := a.dbCtx(ctx)
	defer cancel()

	ids := make([]string, len(workspaces))
	for i, ws := range workspaces {
		ids[i] = ws.ID
	}

	members, err := a.Store.Workspace.GetMembersForWorkspaces(ctx, ids)
	if err != nil {
		tlog.Errorw("Failed to get workspace members", "user_id", user.ID, "error", err)
		return nil, failed
	}

	groupCounts, err := a.Store.Workspace.GetGroupCounts(ctx, ids)
	if err != nil {
		tlog.Errorw("Failed to get workspace group counts", "user_id", user.ID, "error", err)
		return nil, failed
	}

	tables, err := a.Store.Workspace.GetTableMetas(ids)
	if err != nil {
		tlog.Errorw("Failed to get workspace tables", "user_id", user.ID, "error", err)
		return nil, failed
	}

	inputs, err := a.workspaceAccessInputs(ctx, user, ids)
	if err != nil {
		tlog.Errorw("Failed to get workspace roles", "user_id", user.ID, "error", err)
		return nil, failed
	}

	tablesByWorkspace := make(map[string][]model.WorkspaceTable, len(ids))
	for _, t := range tables {
		tablesByWorkspace[t.WorkspaceID] = append(tablesByWorkspace[t.WorkspaceID], t)
	}

	for i := range workspaces {
		ws := &workspaces[i]
		in := inputs[ws.ID]

		ws.Members = members[ws.ID]
		if ws.Members == nil {
			ws.Members = []model.WorkspaceMember{}
		}

		ws.GroupCount = groupCounts[ws.ID]
		ws.TableCount = len(a.resolveAccess(in).VisibleTables(tablesByWorkspace[ws.ID]))

		for _, role := range in.Roles {
			ws.UserPermissions = append(ws.UserPermissions, role.Permissions...)
			ws.UserRoles = append(ws.UserRoles, role.DisplayName)

			if role.Name == model.ProjectWorkspaceAdminRoleID {
				ws.CanDelete = true
			}
		}
	}

	return workspaces, nil
}

// workspaceAccessInputs gives, for each workspace, the roles and table
// permissions that decide what the user sees there.
func (a *App) workspaceAccessInputs(ctx context.Context, user model.User, workspaceIDs []string) (map[string]interfaces.AccessInput, error) {
	inputs := make(map[string]interfaces.AccessInput, len(workspaceIDs))
	roleNames, err := a.Store.Workspace.GetUserRoleNamesByWorkspace(ctx, user.ID, workspaceIDs)
	if err != nil {
		return nil, err
	}

	roles, err := a.Store.Workspace.GetRolesForWorkspaces(ctx, workspaceIDs)
	if err != nil {
		return nil, err
	}

	var perTableRoleIDs []string
	for _, id := range workspaceIDs {
		var held []model.ProjectWorkspaceRole
		for _, role := range roles[id] {
			if slices.Contains(roleNames[id], role.Name) {
				held = append(held, role)
				if role.PerTableMode {
					perTableRoleIDs = append(perTableRoleIDs, role.ID)
				}
			}
		}

		inputs[id] = interfaces.AccessInput{Roles: held}
	}

	perms, err := a.Store.Workspace.GetTablePermissionsForRoleIDs(ctx, perTableRoleIDs)
	if err != nil {
		return nil, err
	}

	for _, p := range perms {
		in := inputs[p.WorkspaceID]
		in.TablePerms = append(in.TablePerms, p)
		inputs[p.WorkspaceID] = in
	}

	return inputs, nil
}

func (a *App) UploadWorkspaceFiles(workspaceID, tableID, taskID, field, fieldID string, user model.User, r *http.Request) ([]model.ProjectFileMetadata, *model.AppError) {
	isMember, appErr := a.IsWorkspaceMember(workspaceID, user.ID)
	if appErr != nil {
		return nil, appErr
	}

	if !isMember {
		return nil, model.NewAppError("workspace.forbidden", http.StatusForbidden)
	}

	if appErr := a.CheckTableInWorkspace(r.Context(), workspaceID, tableID); appErr != nil {
		return nil, appErr
	}

	if !a.CanPerformRowAction(user, workspaceID, tableID, model.PermissionUpdateTask, model.PermissionUpdateTask) {
		return nil, model.NewAppError("workspace.forbidden", http.StatusForbidden)
	}

	if appErr := a.requireTaskAccess(user, workspaceID, tableID, taskID); appErr != nil {
		return nil, appErr
	}

	form := r.MultipartForm
	files := form.File["files"]

	if len(files) == 0 {
		return nil, model.NewAppError("workspace.file_no_files", http.StatusBadRequest)
	}

	primary, err := a.Store.Storage.GetPrimary()
	if err != nil {
		tlog.Errorw("Failed to get primary storage", "workspace_id", workspaceID, "error", err)
		return nil, model.NewAppError("workspace.file_upload_failed", http.StatusInternalServerError)
	}

	uploadedFiles := make([]model.ProjectFileMetadata, 0)
	currentTime := time.Now().Unix()

	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			tlog.Errorw("Failed to open uploaded file", "filename", fileHeader.Filename, "error", err)
			continue
		}

		defer file.Close()

		fileID := model.NewID()
		fileExt := filepath.Ext(fileHeader.Filename)
		storageID := ""

		var writeErr error
		if primary != nil {
			backend, exists := a.FileStorageObjects[primary.ID]
			if !exists {
				tlog.Errorw("Primary storage backend not available", "workspace_id", workspaceID, "storage_id", primary.ID)
				continue
			}

			subPath := workspaceID + "/" + tableID + "/" + taskID + "/" + fileID + fileExt
			filePath, pathErr := a.BuildFilePath(primary.ID, subPath, model.AppProjects)
			if pathErr != nil {
				tlog.Errorw("Failed to build project file path", "workspace_id", workspaceID, "error", pathErr)
				continue
			}

			writeErr = backend.WriteFile(context.Background(), filePath, file, fileHeader.Size)
			if writeErr == nil {
				storageID = primary.ID
			}
		} else {
			targetDir := path.Join("data", "projects", workspaceID, tableID, taskID)
			if mkErr := os.MkdirAll(targetDir, 0o700); mkErr != nil {
				tlog.Errorw("Failed to create directory for workspace files", "workspace_id", workspaceID, "error", mkErr)
				continue
			}

			filePath := path.Join(targetDir, fileID+fileExt)
			dst, createErr := os.Create(filePath)
			if createErr != nil {
				tlog.Errorw("Failed to create file on disk", "filepath", filePath, "error", createErr)
				continue
			}

			_, writeErr = io.Copy(dst, file)
			dst.Close()
			if writeErr != nil {
				os.Remove(filePath)
			}
		}

		if writeErr != nil {
			tlog.Errorw("Failed to write uploaded file", "file_id", fileID, "workspace_id", workspaceID, "error", writeErr)
			continue
		}

		width := 0
		height := 0
		mimeType := fileHeader.Header.Get("Content-Type")

		if strings.HasPrefix(mimeType, "image/") && storageID != "" {
			subPath := workspaceID + "/" + tableID + "/" + taskID + "/" + fileID + fileExt
			if imgPath, pathErr := a.BuildFilePath(storageID, subPath, model.AppProjects); pathErr == nil {
				if imgFile, openErr := os.Open(imgPath); openErr == nil {
					defer imgFile.Close()
					if config, _, decErr := image.DecodeConfig(imgFile); decErr == nil {
						width = config.Width
						height = config.Height
					}
				}
			}
		} else if strings.HasPrefix(mimeType, "image/") {
			legacyPath := path.Join("data", "projects", workspaceID, tableID, taskID, fileID+fileExt)
			if imgFile, openErr := os.Open(legacyPath); openErr == nil {
				defer imgFile.Close()
				if config, _, decErr := image.DecodeConfig(imgFile); decErr == nil {
					width = config.Width
					height = config.Height
				}
			}
		}

		attachment := &model.WorkspaceAttachment{
			ID:          fileID,
			WorkspaceID: workspaceID,
			TableID:     tableID,
			FieldID:     fieldID,
			TaskID:      taskID,
			UserID:      user.ID,
			Name:        fileHeader.Filename,
			Size:        fileHeader.Size,
			MimeType:    mimeType,
			Width:       width,
			Height:      height,
			StorageID:   storageID,
			CreatedAt:   currentTime,
			UpdatedAt:   currentTime,
			DeletedAt:   0,
		}

		if err := a.Store.Workspace.UpdateAttachment(attachment); err != nil {
			tlog.Errorw("Failed to save file attachment to database", "file_id", fileID, "workspace_id", workspaceID, "error", err)
			continue
		}

		fileMetadata := model.ProjectFileMetadata{
			ID:   fileID,
			Name: fileHeader.Filename,
			Size: fileHeader.Size,
			Type: mimeType,
		}

		uploadedFiles = append(uploadedFiles, fileMetadata)
	}

	if len(uploadedFiles) == 0 {
		return nil, model.NewAppError("workspace.file_upload_failed", http.StatusBadRequest)
	}

	a.publishTaskRow(r.Context(), "task_updated", workspaceID, tableID, taskID)

	return uploadedFiles, nil
}

func (a *App) CreateWorkspaceTable(ctx context.Context, workspaceID string, name string, linked bool, parent_table_id string, user model.User, single_select bool, linkBothDirections bool, secondTableID string, folderID string, create_subtasks bool) (*model.WorkspaceTable, *model.AppError) {
	if !a.ProjectWorkspaceHasPermission(user, workspaceID, model.PermissionCreateTable) {
		return nil, model.NewAppError("workspace.forbidden", http.StatusForbidden)
	}

	if strings.TrimSpace(name) == "" {
		return nil, model.NewAppError("workspace.table_name_empty", http.StatusBadRequest)
	}

	defaultTaskOrder := []model.TaskOrderField{
		{Name: "id", DisplayName: "ID", Width: "150", Visible: false},
		{Name: "name", DisplayName: "Name", Width: "150", Visible: true},
		{Name: "start_date", DisplayName: "Start Date", Width: "150", Visible: false},
		{Name: "due_date", DisplayName: "Due Date", Width: "150", Visible: true},
		{Name: "assignee", DisplayName: "Assignee", Width: "150", Visible: true},
		{Name: "status", DisplayName: "Status", Width: "150", Visible: true},
		{Name: "description", DisplayName: "Description", Width: "150", Visible: true},
		{Name: "updated_at", DisplayName: "Updated At", Width: "150", Visible: true},
		{Name: "created_at", DisplayName: "Created At", Width: "150", Visible: false},
		{Name: "deleted_at", DisplayName: "Deleted At", Width: "150", Visible: true},
		{Name: "created_by", DisplayName: "Created By", Width: "150", Visible: false},
	}
	taskOrderJSON, jsonErr := json.Marshal(defaultTaskOrder)
	if jsonErr != nil {
		return nil, model.NewAppError("workspace.table_create_failed", http.StatusInternalServerError)
	}

	defaultStatuses := []model.KanbanStatusOption{
		{ID: model.NewID(), Name: "Cancelled", Color: "#4b5563", StatusType: "Closed"},
		{ID: model.NewID(), Name: "Completed", Color: "#16a34a", StatusType: "Done"},
	}

	tableID := model.NewID()
	safeName := physicalName("t", tableID)
	statusTableID := model.NewID()
	statusTablePhysicalName := "t" + strings.ReplaceAll(model.NewID(), "-", "")

	workspaceTable, err := a.Store.Workspace.CreateTable(tableID, workspaceID, name, safeName, linked, parent_table_id, user.ID, single_select, linkBothDirections, secondTableID, folderID, string(taskOrderJSON), defaultStatuses, statusTableID, statusTablePhysicalName, model.NewID())
	if err != nil {
		tlog.Errorw("Failed to create workspace table", "workspace_id", workspaceID, "error", err)
		return nil, model.NewAppError("workspace.table_create_failed", http.StatusInternalServerError)
	}

	if create_subtasks {
		subtaskName := name + "_subtasks"
		subtaskTableID := model.NewID()
		safeSubtaskName := physicalName("t", subtaskTableID)
		subtaskStatusTableID := model.NewID()
		subtaskStatusPhysicalName := "t" + strings.ReplaceAll(model.NewID(), "-", "")
		subtaskStatuses := []model.KanbanStatusOption{
			{ID: model.NewID(), Name: "Cancelled", Color: "#4b5563", StatusType: "Closed"},
			{ID: model.NewID(), Name: "Completed", Color: "#16a34a", StatusType: "Done"},
		}

		subtaskTable, err := a.Store.Workspace.CreateTable(subtaskTableID, workspaceID, subtaskName, safeSubtaskName, false, "", user.ID, false, true, "", folderID, string(taskOrderJSON), subtaskStatuses, subtaskStatusTableID, subtaskStatusPhysicalName, model.NewID())
		if err != nil {
			tlog.Errorw("Failed to create subtask table", "workspace_id", workspaceID, "error", err)
			return nil, model.NewAppError("workspace.table_create_failed", http.StatusInternalServerError)
		}

		_, err = a.Store.Workspace.CreateTableField(model.CreateFieldParams{
			WorkspaceID:                   workspaceID,
			TableID:                       tableID,
			FieldName:                     "subtasks",
			FieldType:                     "link",
			SQLFieldType:                  "TEXT",
			LinkedTableID:                 subtaskTableID,
			UserID:                        user.ID,
			SelectedType:                  "link",
			LinkBothDirections:            true,
			FieldNameInSecondTable:        "parent_task",
			FieldNameDisplay:              "Subtasks",
			FieldNameInSecondTableDisplay: "Tasks",
			MainFieldID:                   model.NewID(),
			SecondFieldID:                 model.NewID(),
			LinkTableID:                   model.NewID(),
			SecondLinkID:                  model.NewID(),
			LinkTablePhysicalName:         "t" + strings.ReplaceAll(model.NewID(), "-", ""),
			SecondLinkTablePhysicalName:   "t" + strings.ReplaceAll(model.NewID(), "-", ""),
			SingleSelectTableID:           model.NewID(),
			SingleSelectPhysicalName:      "t" + strings.ReplaceAll(model.NewID(), "-", ""),
			LinkViewID:                    model.NewID(),
			SecondLinkViewID:              model.NewID(),
		})
		if err != nil {
			tlog.Errorw("Failed to create subtask link field", "workspace_id", workspaceID, "error", err)
			return nil, model.NewAppError("workspace.table_create_failed", http.StatusInternalServerError)
		}

		workspaceTable.SubtaskProject = subtaskTable
	}

	a.publishProjectChange(ctx, projectChange{kind: "UPDATE_TABLES", workspaceID: workspaceID, tableID: workspaceTable.ID,
		data: map[string]any{"data": workspaceTable}})
	if sub := workspaceTable.SubtaskProject; sub != nil {
		a.publishProjectChange(ctx, projectChange{kind: "UPDATE_TABLES", workspaceID: workspaceID, tableID: sub.ID,
			data: map[string]any{"data": sub}})
	}

	return workspaceTable, nil
}

func (a *App) AddMemberToWorkspace(workspaceID string, userIDs []string, user model.User) ([]model.WorkspaceMember, *model.AppError) {
	if !a.ProjectWorkspaceHasPermission(user, workspaceID, model.PermissionAddMemberToWorkspace) {
		return nil, model.NewAppError("workspace.forbidden", http.StatusForbidden)
	}

	return a.addProjectWorkspaceMembers(workspaceID, userIDs, user)
}

func (a *App) addProjectWorkspaceMembers(workspaceID string, userIDs []string, user model.User) ([]model.WorkspaceMember, *model.AppError) {
	now := time.Now().Unix()
	members := make([]model.WorkspaceMember, 0, len(userIDs))
	for _, uid := range userIDs {
		members = append(members, model.WorkspaceMember{
			ID:          model.NewID(),
			UserID:      uid,
			WorkspaceID: workspaceID,
			Role:        "user",
			DateJoined:  now,
		})
	}

	workspaceMembers, err := a.Store.Workspace.AddMember(members)
	if err != nil {
		tlog.Errorw("Failed to add member to workspace", "workspace_id", workspaceID, "error", err)
		return nil, model.NewAppError("workspace.member_add_failed", http.StatusInternalServerError)
	}

	workspaceName, err := a.Store.Workspace.GetNameByID(workspaceID)
	if err != nil {
		tlog.Errorw("Failed to get workspace name for notification", "workspace_id", workspaceID, "error", err)
		return workspaceMembers, nil
	}

	newIDs := make(map[string]bool, len(members))
	for _, m := range members {
		newIDs[m.ID] = true
	}

	for _, added := range workspaceMembers {
		if !newIDs[added.ID] {
			continue
		}

		userID := added.UserID

		a.createProjectNotification(workspaceID, user, model.NOTIFICATION_PROJECT_INVITE, map[string]interface{}{
			"userID":        userID,
			"workspaceName": *workspaceName,
			"workspaceID":   workspaceID,
		})

		a.RecordActivity(user.ID, model.AppProjects, model.ActivityProjectInvite, "workspace", workspaceID, map[string]any{
			"workspaceID":   workspaceID,
			"AddedUser":     userID,
			"WhoAdded":      user.ID,
			"workspaceName": *workspaceName,
		})
	}

	return workspaceMembers, nil
}

// physicalName is the SQL name of a table or column the server creates: its
// kind's letter and the record's own id. What anyone typed is never part of
// it, so it needs no escaping and cannot clash.
func physicalName(kind, id string) string {
	return kind + strings.ReplaceAll(id, "-", "")
}

// collectLinkedTableIDs returns the unique set of LinkedID and ParentTableID values
// from headers that have linked-table relationships. Used to pre-fetch table names
// before building filter SQL.
func collectLinkedTableIDs(headers []model.WorkspaceHeaders) []string {
	seen := make(map[string]struct{})
	for _, h := range headers {
		if h.ParentTableID != "" && !h.SingleSelect {
			if h.LinkedID != "" {
				seen[h.LinkedID] = struct{}{}
			}

			seen[h.ParentTableID] = struct{}{}
		}
	}

	ids := make([]string, 0, len(seen))
	for id := range seen {
		ids = append(ids, id)
	}

	return ids
}
