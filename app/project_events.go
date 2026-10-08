// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"context"
	"encoding/json"
	"maps"

	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/tlog"
)

type originClientKey struct{}

// WithOriginClient records which browser tab made a request, so that tab can
// recognise the project changes it caused when they come back over the socket.
func WithOriginClient(ctx context.Context, clientID string) context.Context {
	return context.WithValue(ctx, originClientKey{}, clientID)
}

// eventCtx gives publishing its own timeout, apart from the request's: the
// change is already saved, and a client that moved away and cancelled its
// request must not keep everyone else from hearing about it.
func (a *App) eventCtx(ctx context.Context) (context.Context, context.CancelFunc) {
	return a.dbCtx(context.WithoutCancel(ctx))
}

func originClient(ctx context.Context) string {
	id, _ := ctx.Value(originClientKey{}).(string)
	return id
}

// projectChange is a saved change in a workspace. task is set for changes to
// a single task; users who are assigned-only for the table receive those only
// when the task is theirs. onlyUsers, when set, limits a change to those
// users, for things only they can see, such as a private view or saved filter.
type projectChange struct {
	kind        string
	workspaceID string
	tableID     string
	task        map[string]any
	data        map[string]any
	onlyUsers   map[string]bool
	// headers are the task's table's, read for the tasks its link cells name.
	headers []model.WorkspaceHeaders
}

// publishProjectChange pushes a change to the connected users who may see it:
// system admins, and members of the workspace, directly or through a group,
// whose roles show them the table.
func (a *App) publishProjectChange(ctx context.Context, change projectChange) {
	hub := a.Server.NotificationHub
	if hub == nil {
		return
	}

	ctx, cancel := a.eventCtx(ctx)
	defer cancel()

	recipients, ownTasksOnly := a.projectChangeRecipients(ctx, change)
	if len(recipients) == 0 {
		return
	}

	data := map[string]any{
		"type":             change.kind,
		"workspace_id":     change.workspaceID,
		"table_id":         change.tableID,
		"origin_client_id": originClient(ctx),
	}
	for k, v := range change.data {
		data[k] = v
	}

	if change.task != nil {
		data["task"] = change.task
	}

	message, err := json.Marshal(model.WebsocketEvent{Event: model.NOTIFICATION_PROJECT_CHANGE, App: model.AppProjects, Data: data})
	if err != nil {
		tlog.Errorw("Failed to encode project change", "type", change.kind, "workspace_id", change.workspaceID, "error", err)
		return
	}

	hub.send(message, func(c *Client) bool { return recipients[c.user] && ownTasksOnly[c.user] == nil })

	// A user who sees only their own tasks in a linked table gets a copy
	// whose links to others' tasks have no names.
	for userID, seesOwnOnly := range ownTasksOnly {
		task := maps.Clone(change.task)
		a.hideUnseenLinks(ctx, []map[string]interface{}{task}, change.headers, userID, seesOwnOnly)
		data["task"] = task

		message, err := json.Marshal(model.WebsocketEvent{Event: model.NOTIFICATION_PROJECT_CHANGE, App: model.AppProjects, Data: data})
		if err != nil {
			tlog.Errorw("Failed to encode project change", "type", change.kind, "workspace_id", change.workspaceID, "error", err)
			continue
		}

		hub.send(message, func(c *Client) bool { return c.user == userID })
	}
}

// projectChangeRecipients returns who may hear of a change and, for the
// recipients who see only their own tasks in a table the changed task links
// to, which tables those are, so their copy hides the links to others' tasks.
func (a *App) projectChangeRecipients(ctx context.Context, change projectChange) (map[string]bool, map[string]func(string) bool) {
	ownTasksOnly := map[string]func(string) bool{}
	connected := a.Server.NotificationHub.connectedUsers()
	if len(connected) == 0 {
		return nil, nil
	}

	users, err := a.Store.User.GetByIDs(connected)
	if err != nil {
		tlog.Errorw("Failed to load connected users for a project change", "workspace_id", change.workspaceID, "error", err)
		return nil, nil
	}

	memberIDs, err := a.Store.Workspace.GetMemberUserIDs(ctx, change.workspaceID, connected)
	if err != nil {
		tlog.Errorw("Failed to load workspace members for a project change", "workspace_id", change.workspaceID, "error", err)
		return nil, nil
	}

	members := make(map[string]bool, len(memberIDs))
	for _, id := range memberIDs {
		members[id] = true
	}

	recipients := make(map[string]bool)
	var candidates []model.User
	for _, user := range users {
		if change.onlyUsers != nil && !change.onlyUsers[user.ID] {
			continue
		}

		if members[user.ID] {
			candidates = append(candidates, user)
		}
	}

	if change.tableID == "" {
		for _, user := range candidates {
			recipients[user.ID] = true
		}

		return recipients, ownTasksOnly
	}

	assignee, _ := change.task["assignee"].(string)
	inputs := a.buildAccessInputs(ctx, candidates, change.workspaceID)
	for _, user := range candidates {
		access := a.resolveAccess(inputs[user.ID])
		if len(access.VisibleTables([]model.WorkspaceTable{{ID: change.tableID}})) == 0 {
			continue
		}

		if change.task != nil && access.AssignedOnly(change.tableID) && assignee != user.ID {
			continue
		}

		recipients[user.ID] = true
		if change.task != nil && linksToOwnTasksOnly(change.headers, access.AssignedOnly) {
			ownTasksOnly[user.ID] = access.AssignedOnly
		}
	}

	return recipients, ownTasksOnly
}

// linksToOwnTasksOnly reports whether a link field among headers leads to a
// table where the user sees only their own tasks.
func linksToOwnTasksOnly(headers []model.WorkspaceHeaders, ownTasksOnly func(tableID string) bool) bool {
	for _, header := range headers {
		if header.LinkedID != "" && !header.SingleSelect && header.ParentTableID != "" && ownTasksOnly(header.ParentTableID) {
			return true
		}
	}

	return false
}

// publishViewChange publishes a change to one view. A private view's changes
// go only to the users who may see it, and are not published when the view is
// unknown.
func (a *App) publishViewChange(ctx context.Context, viewID string, change projectChange) {
	ctx, cancel := a.eventCtx(ctx)
	defer cancel()

	audience, err := a.viewAudience(ctx, viewID)
	if err != nil {
		tlog.Errorw("Failed to load view for a project change", "view_id", viewID, "error", err)
		return
	}

	change.onlyUsers = audience
	a.publishProjectChange(ctx, change)
}

// publishViewAccessChange tells the users who gained or lost a view, when it
// was made public or private or shared differently, as though it had been
// created or deleted. before is the audience the view had before the change.
func (a *App) publishViewAccessChange(ctx context.Context, workspaceID, tableID, viewID string, before map[string]bool) {
	hub := a.Server.NotificationHub
	if hub == nil {
		return
	}

	ctx, cancel := a.eventCtx(ctx)
	defer cancel()

	after, err := a.viewAudience(ctx, viewID)
	if err != nil {
		tlog.Errorw("Failed to load view for a project change", "view_id", viewID, "error", err)
		return
	}

	sees := func(audience map[string]bool, userID string) bool { return audience == nil || audience[userID] }
	gained, lost, kept := map[string]bool{}, map[string]bool{}, map[string]bool{}
	for _, id := range hub.connectedUsers() {
		switch b, s := sees(before, id), sees(after, id); {
		case !b && s:
			gained[id] = true
		case b && !s:
			lost[id] = true
		case b && s:
			kept[id] = true
		}
	}

	if len(lost) > 0 {
		a.publishProjectChange(ctx, projectChange{kind: "DELETE_VIEW", workspaceID: workspaceID, tableID: tableID,
			data: map[string]any{"delete_id": viewID}, onlyUsers: lost})
	}
	// Those who still see the view are told whether it is now public, which
	// their view menus offer to change.
	if len(kept) > 0 {
		a.publishProjectChange(ctx, projectChange{kind: "VIEW_VISIBILITY", workspaceID: workspaceID, tableID: tableID,
			data: map[string]any{"view_id": viewID, "is_public": after == nil}, onlyUsers: kept})
	}

	if len(gained) == 0 {
		return
	}

	view, err := a.Store.Workspace.GetView(ctx, viewID)
	if err != nil {
		tlog.Errorw("Failed to load view for a project change", "view_id", viewID, "error", err)
		return
	}

	kind := "VIEW_CREATED"
	if view.ViewType == "grid" {
		kind = "CREATE_GRID_VIEW"
	}

	a.publishProjectChange(ctx, projectChange{kind: kind, workspaceID: workspaceID, tableID: tableID,
		data: map[string]any{"data": view}, onlyUsers: gained})
}

// viewAudience returns nil for a public view, and otherwise the creator of the
// view and the users it is shared with.
func (a *App) viewAudience(ctx context.Context, viewID string) (map[string]bool, error) {
	isPublic, createdBy, err := a.Store.Workspace.GetViewVisibility(ctx, viewID)
	if err != nil || isPublic {
		return nil, err
	}

	shared, err := a.Store.Workspace.GetViewSharedUserIDs(viewID)
	if err != nil {
		return nil, err
	}

	audience := map[string]bool{createdBy: true}
	for _, id := range shared {
		audience[id] = true
	}

	return audience, nil
}

// publishTaskRow publishes a task as a grid page returns it, so clients can
// put it in place of the row they hold. Rows of single-select tables are
// options, not tasks, so they go to everyone who can see the table.
func (a *App) publishTaskRow(ctx context.Context, kind, workspaceID, tableID, taskID string) {
	ctx, cancel := a.eventCtx(ctx)
	defer cancel()

	row, headers, isSingleSelect, ok := a.loadTaskRow(ctx, tableID, taskID)
	if !ok {
		return
	}

	change := projectChange{kind: kind, workspaceID: workspaceID, tableID: tableID}
	if isSingleSelect {
		change.data = map[string]any{"row": row}
	} else {
		change.task = row
		change.headers = headers
	}

	a.publishProjectChange(ctx, change)
}

// publishTaskAssignment publishes an update to a task whose assignee may have
// changed. Users who see only their own tasks never hear about a task that is
// not theirs, so the previous assignee is told to drop it and the new one is
// sent it as a task they have not seen.
func (a *App) publishTaskAssignment(ctx context.Context, workspaceID, tableID, taskID, previousAssignee string) {
	ctx, cancel := a.eventCtx(ctx)
	defer cancel()

	row, headers, isSingleSelect, ok := a.loadTaskRow(ctx, tableID, taskID)
	if !ok {
		return
	}

	if isSingleSelect {
		a.publishProjectChange(ctx, projectChange{kind: "task_updated", workspaceID: workspaceID, tableID: tableID, data: map[string]any{"row": row}})
		return
	}

	a.publishProjectChange(ctx, projectChange{kind: "task_updated", workspaceID: workspaceID, tableID: tableID, task: row, headers: headers})

	assignee, _ := row["assignee"].(string)
	if assignee == previousAssignee {
		return
	}

	if previousAssignee != "" && a.seesOnlyOwnTasks(previousAssignee, workspaceID, tableID) {
		a.publishProjectChange(ctx, projectChange{kind: "task_unassigned", workspaceID: workspaceID, tableID: tableID,
			data: map[string]any{"task_id": taskID}, onlyUsers: map[string]bool{previousAssignee: true}})
	}

	if assignee != "" && a.seesOnlyOwnTasks(assignee, workspaceID, tableID) {
		a.publishProjectChange(ctx, projectChange{kind: "task_created", workspaceID: workspaceID, tableID: tableID, task: row, headers: headers, onlyUsers: map[string]bool{assignee: true}})
	}
}

func (a *App) seesOnlyOwnTasks(userID, workspaceID, tableID string) bool {
	users, err := a.Store.User.GetByIDs([]string{userID})
	if err != nil || len(users) == 0 {
		return false
	}

	return a.resolveAccess(a.buildAccessInput(users[0], workspaceID)).AssignedOnly(tableID)
}

func (a *App) loadTaskRow(ctx context.Context, tableID, taskID string) (map[string]interface{}, []model.WorkspaceHeaders, bool, bool) {
	tableName, isSingleSelect, err := a.Store.Workspace.GetTableMeta(tableID)
	if err != nil {
		tlog.Errorw("Failed to load table for a project change", "table_id", tableID, "error", err)
		return nil, nil, false, false
	}

	row, err := a.Store.Workspace.GetTableRowByID(tableName, taskID)
	if err != nil || row == nil {
		tlog.Errorw("Failed to load task for a project change", "table_id", tableID, "task_id", taskID, "error", err)
		return nil, nil, false, false
	}

	colTypes, err := a.Store.Workspace.GetTableColumnTypes(tableName)
	if err != nil {
		tlog.Errorw("Failed to load columns for a project change", "table_id", tableID, "error", err)
		return nil, nil, false, false
	}

	headers := a.buildTableHeaders(tableID, colTypes, false, true)
	rows := []map[string]interface{}{row}
	a.hydrateLinkedAndSingleSelect(ctx, rows, headers)
	a.hydrateFileAttachments(ctx, rows, headers, tableID)
	return rows[0], headers, isSingleSelect, true
}

// taskRef identifies a task for a change that does not carry the task itself,
// with enough of it to decide who may hear about it.
func (a *App) taskRef(tableID, taskID string) map[string]any {
	ref := map[string]any{"id": taskID}
	if tableName, err := a.Store.Workspace.GetTableName(tableID); err == nil {
		if assignee, err := a.Store.Workspace.GetTaskAssigneeID(tableName, taskID); err == nil {
			ref["assignee"] = assignee
		}
	}

	return ref
}
