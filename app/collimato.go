// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"net/http"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/twigex/twigex/collimato"
	"github.com/twigex/twigex/crypto"
	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/queryengine"
	"github.com/twigex/twigex/tlog"
)

const collimatoUnlicensedWorkspaces = 1

// The workspace cards show this many member avatars.
const workspaceMemberPreviewSize = 4

func (a *App) CreateWorkspace(user model.User, name, description string) (*model.CollimatoWorkspace, *model.AppError) {
	if !a.SessionHasPermission(user, model.CollimatoSectionPermissions.PermissionCreateCollimatoWorkspace) {
		return nil, model.NewAppError("permission.forbidden", http.StatusForbidden)
	}

	if !a.Server.License.HasCollimatoUnlimitedWorkspaces() {
		count, err := a.Store.Collimato.CountWorkspaces(context.Background())
		if err != nil {
			tlog.Errorw("Failed to count workspaces",
				"error", err,
			)
			return nil, model.NewAppError("collimato.retrieval_failed", http.StatusInternalServerError)
		}

		if count >= collimatoUnlicensedWorkspaces {
			return nil, model.NewAppError("collimato.workspace_limit", http.StatusPaymentRequired)
		}
	}

	wsp, err := a.Store.Collimato.CreateWorkspace(user.ID, name, description, model.CollimatoWorkspaceStatusDraft, rand.Text())
	if err != nil {
		tlog.Errorw("Failed to create workspace",
			"user_id", user.ID,
			"error", err,
		)
		return nil, model.NewAppError("collimato.create_failed", http.StatusInternalServerError)
	}

	err = a.Store.Collimato.AddWorkspaceUsers(wsp.ID, []string{user.ID}, []string{
		model.CollimatoWorkspaceAdminRoleId,
	})
	if err != nil {
		tlog.Errorw("Failed to add user to workspace",
			"workspace_id", wsp.ID,
			"user_id", user.ID,
			"error", err,
		)
		return nil, model.NewAppError("collimato.member_add_failed", http.StatusInternalServerError)
	}

	return wsp, nil
}

func (a *App) GetWorkspaceByID(user model.User, workspaceID string) (*model.CollimatoWorkspace, *model.AppError) {
	workspace, err := a.Store.Collimato.GetWorkspaceByID(workspaceID)
	if err != nil {
		tlog.Errorw("Failed to retrieve workspace",
			"workspace_id", workspaceID,
			"error", err,
		)
		return nil, model.NewAppError("collimato.retrieval_failed", http.StatusInternalServerError)
	}

	roleNames, err := a.Store.Collimato.GetEffectiveRolesForUser(context.Background(), workspaceID, user.ID)
	if err != nil {
		tlog.Errorw("Failed to resolve effective workspace roles",
			"workspace_id", workspaceID,
			"user_id", user.ID,
			"error", err,
		)
		return nil, model.NewAppError("collimato.retrieval_failed", http.StatusInternalServerError)
	}

	if len(roleNames) == 0 {
		return nil, model.NewAppError("collimato.forbidden", http.StatusForbidden)
	}

	return workspace, nil
}

func (a *App) DeleteWorkspace(user model.User, workspaceID string) *model.AppError {
	workspace, err := a.Store.Collimato.GetWorkspaceByID(workspaceID)
	if err != nil {
		tlog.Errorw("Failed to retrieve workspace",
			"workspace_id", workspaceID,
			"error", err,
		)
		return model.NewAppError("collimato.retrieval_failed", http.StatusInternalServerError)
	}

	if !a.holdsCollimatoWorkspaceAdmin(user, workspace.ID) {
		return model.NewAppError("permission.forbidden", http.StatusForbidden)
	}

	err = a.Store.Collimato.DeleteWorkspace(workspaceID)
	if err != nil {
		tlog.Errorw("Failed to delete workspace",
			"workspace_id", workspaceID,
			"error", err,
		)
		return model.NewAppError("collimato.delete_failed", http.StatusInternalServerError)
	}

	return nil
}

func (a *App) FinishWorkspaceCreation(user model.User, workspaceID string, status string) (*model.CollimatoWorkspace, *model.AppError) {
	workspace, err := a.Store.Collimato.GetWorkspaceByID(workspaceID)
	if err != nil {
		tlog.Errorw("Failed to retrieve workspace",
			"workspace_id", workspaceID,
			"error", err,
		)
		return nil, model.NewAppError("collimato.retrieval_failed", http.StatusInternalServerError)
	}

	if !a.holdsCollimatoWorkspaceAdmin(user, workspace.ID) {
		return nil, model.NewAppError("collimato.forbidden", http.StatusForbidden)
	}

	connections, err := a.Store.Collimato.GetConnections(workspace.ID)
	if err != nil {
		tlog.Errorw("Failed to retrieve workspace connections",
			"workspace_id", workspaceID,
			"error", err,
		)
		return nil, model.NewAppError("collimato.finish_failed", http.StatusInternalServerError)
	}

	if len(connections) < 1 {
		tlog.Warnw("No connections found for workspace",
			"workspace_id", workspaceID,
		)
		return nil, model.NewAppError("collimato.finish_failed", http.StatusInternalServerError)
	}

	workspace.Status = status
	updatedWorkspace, err := a.Store.Collimato.UpdateWorkspace(*workspace)
	if err != nil {
		tlog.Errorw("Failed to update workspace status",
			"workspace_id", workspaceID,
			"status", status,
			"error", err,
		)
		return nil, model.NewAppError("collimato.finish_failed", http.StatusInternalServerError)
	}

	return updatedWorkspace, nil
}

func (a *App) GetWorkspaces(ctx context.Context, user model.User) ([]model.CollimatoWorkspace, *model.AppError) {
	if !a.SessionHasPermission(user, model.CollimatoSectionPermissions.PermissionViewCollimato) {
		return nil, model.NewAppError("permission.forbidden", http.StatusForbidden)
	}

	ctx, cancel := a.dbCtx(ctx)
	defer cancel()

	workspaces, err := a.Store.Collimato.GetWorkspacesForUser(user.ID)
	if err != nil {
		tlog.Errorw("Failed to retrieve user workspaces",
			"user_id", user.ID,
			"error", err,
		)
		return nil, model.NewAppError("collimato.retrieval_failed", http.StatusInternalServerError)
	}

	ids := make([]string, len(workspaces))
	for i, workspace := range workspaces {
		ids[i] = workspace.ID
	}

	// Every listed workspace is one whose roster this user may already read
	// (GetWorkspaceUsers), so the preview shows nothing new.
	members, err := a.Store.Collimato.GetWorkspaceMemberIDs(ctx, ids)
	if err != nil {
		tlog.Errorw("Failed to retrieve workspace members",
			"user_id", user.ID,
			"error", err,
		)
		return nil, model.NewAppError("collimato.retrieval_failed", http.StatusInternalServerError)
	}

	for i := range workspaces {
		memberIDs := members[workspaces[i].ID]

		workspaces[i].MemberCount = len(memberIDs)
		workspaces[i].MemberIDs = memberIDs[:min(len(memberIDs), workspaceMemberPreviewSize)]
		workspaces[i].CanDelete = a.holdsCollimatoWorkspaceAdmin(user, workspaces[i].ID)
	}

	return workspaces, nil
}

// engineConnector opens every connection to a customer database, so
// verification, listing tables and charts cannot disagree about how to reach
// one. Which addresses may be reached is not limited yet: that is Twigex's
// choice, made through ConnectorOptions.Dial. Queries stop after 10 minutes.
// Up to 8 connections per database: twice the engine's default, and a small
// share of the 151 a database usually allows.
var engineConnector = queryengine.NewConnector(queryengine.ConnectorOptions{
	QueryTimeLimit:     10 * time.Minute,
	MaxOpenConnections: 8,
})

func (a *App) verifyConnection(conn model.Connection) error {
	ctx, cancel := a.dbCtx(context.Background())
	defer cancel()

	settings := collimato.EngineConnectionSettings(conn, conn.Password)

	return engineConnector.Check(ctx, settings)
}

func (a *App) encryptConnectionPassword(plaintext string) (string, error) {
	return crypto.Encrypt(*a.ConfigStore.Config.ServerSettings.AtRestEncryptKey, plaintext)
}

func (a *App) decryptConnectionPassword(ciphertext string) (string, error) {
	return crypto.Decrypt(*a.ConfigStore.Config.ServerSettings.AtRestEncryptKey, ciphertext)
}

func (a *App) TestConnection(user model.User, workspaceID, id string, conn model.NewConnection) *model.AppError {
	permission := model.CollimatoPermissions.PermissionCreateConnections
	if id != "" {
		permission = model.CollimatoPermissions.PermissionEditConnections
	}

	if !a.CollimatoWorkspaceHasPermission(user, workspaceID, permission) {
		return model.NewAppError("connection.forbidden", http.StatusForbidden)
	}

	if !model.SSLModeSupported(conn.SSLMode) {
		return model.NewAppError("connection.ssl_mode_unsupported", http.StatusBadRequest)
	}

	password := conn.Password

	if password == "" && id != "" {
		stored, err := a.Store.Collimato.GetConnectionByID(id)
		if err != nil {
			tlog.Errorw("Failed to retrieve connection",
				"workspace_id", workspaceID,
				"connection_id", id,
				"error", err,
			)
			return model.NewAppError("connection.retrieval_failed", http.StatusInternalServerError)
		}

		if stored.WorkspaceID != workspaceID {
			return model.NewAppError("connection.forbidden", http.StatusForbidden)
		}

		if password, err = a.decryptConnectionPassword(stored.Password); err != nil {
			tlog.Errorw("Failed to decrypt connection password",
				"workspace_id", workspaceID,
				"connection_id", id,
				"error", err,
			)
			return model.NewAppError("connection.retrieval_failed", http.StatusInternalServerError)
		}
	}

	err := a.verifyConnection(model.Connection{
		Type:     conn.Type,
		Host:     conn.Host,
		Port:     conn.Port,
		Username: conn.Username,
		Password: password,
		Database: conn.Database,
		SSLMode:  conn.SSLMode,
		Config:   conn.Config,
	})
	if err == nil {
		return nil
	}

	tlog.Warnw("Connection test failed",
		"workspace_id", workspaceID,
		"connection_id", id,
		"reason", queryengine.ConnectionFailureOf(err),
		"error", err,
	)

	return verificationError(err)
}

func (a *App) AddConnection(user model.User, workspaceID string, conn model.NewConnection) (*model.Connection, *model.AppError) {
	if !a.CollimatoWorkspaceHasPermission(user, workspaceID, model.CollimatoPermissions.PermissionCreateConnections) {
		return nil, model.NewAppError("collimato.forbidden", http.StatusForbidden)
	}

	workspace, err := a.Store.Collimato.GetWorkspaceByID(workspaceID)
	if err != nil {
		tlog.Errorw("Failed to retrieve workspace",
			"workspace_id", workspaceID,
			"error", err,
		)
		return nil, model.NewAppError("collimato.retrieval_failed", http.StatusInternalServerError)
	}

	databases, err := a.Store.Collimato.GetConnections(workspace.ID)
	if err != nil {
		tlog.Errorw("Failed to retrieve connections",
			"workspace_id", workspace.ID,
			"error", err,
		)
		return nil, model.NewAppError("connection.retrieval_failed", http.StatusInternalServerError)
	}

	conn.Default = len(databases) <= 0
	conn.WorkspaceID = workspace.ID

	if !model.SSLModeSupported(conn.SSLMode) {
		return nil, model.NewAppError("connection.ssl_mode_unsupported", http.StatusBadRequest)
	}

	if !conn.UsesCACertificate() {
		delete(conn.Config, model.ConnectionCACertKey)
	}

	if err = a.verifyConnection(model.Connection{
		Type:     conn.Type,
		Host:     conn.Host,
		Port:     conn.Port,
		Username: conn.Username,
		Password: conn.Password,
		Database: conn.Database,
		SSLMode:  conn.SSLMode,
		Config:   conn.Config,
	}); err != nil {
		tlog.Warnw("Connection verification failed",
			"workspace_id", workspace.ID,
			"reason", queryengine.ConnectionFailureOf(err),
			"error", err,
		)
		return nil, verificationError(err)
	}

	encrypted, err := a.encryptConnectionPassword(conn.Password)
	if err != nil {
		tlog.Errorw("Failed to encrypt connection password",
			"workspace_id", workspace.ID,
			"error", err,
		)
		return nil, model.NewAppError("connection.create_failed", http.StatusInternalServerError)
	}

	conn.Password = encrypted

	database, err := a.Store.Collimato.CreateConnection(conn)
	if err != nil {
		tlog.Errorw("Failed to create connection",
			"workspace_id", workspace.ID,
			"error", err,
		)
		return nil, model.NewAppError("connection.create_failed", http.StatusInternalServerError)
	}

	return database, nil
}

func (a *App) GetConnections(user model.User, workspaceID string) ([]model.Connection, *model.AppError) {
	if !a.CollimatoWorkspaceHasPermission(user, workspaceID, model.CollimatoPermissions.PermissionViewConnections) {
		return nil, model.NewAppError("connection.forbidden", http.StatusForbidden)
	}

	databases, err := a.Store.Collimato.GetConnections(workspaceID)
	if err != nil {
		tlog.Errorw("Failed to retrieve connections",
			"workspace_id", workspaceID,
			"error", err,
		)
		return nil, model.NewAppError("connection.retrieval_failed", http.StatusInternalServerError)
	}

	return databases, nil
}

func (a *App) GetConnectionByID(user model.User, workspaceID, id string) (*model.Connection, *model.AppError) {
	if !a.CollimatoWorkspaceHasPermission(user, workspaceID, model.CollimatoPermissions.PermissionViewConnections) {
		return nil, model.NewAppError("connection.forbidden", http.StatusForbidden)
	}

	database, err := a.Store.Collimato.GetConnectionByID(id)
	if err != nil {
		tlog.Errorw("Failed to retrieve connection",
			"workspace_id", workspaceID,
			"connection_id", id,
			"error", err,
		)
		return nil, model.NewAppError("connection.retrieval_failed", http.StatusInternalServerError)
	}

	if database.WorkspaceID != workspaceID {
		return nil, model.NewAppError("connection.forbidden", http.StatusForbidden)
	}

	return database, nil
}

func (a *App) UpdateConnection(user model.User, workspaceID, id string, conn model.ConnectionPatch) (*model.Connection, *model.AppError) {
	if !a.CollimatoWorkspaceHasPermission(user, workspaceID, model.CollimatoPermissions.PermissionEditConnections) {
		return nil, model.NewAppError("connection.forbidden", http.StatusForbidden)
	}

	database, err := a.Store.Collimato.GetConnectionByID(id)
	if err != nil {
		tlog.Errorw("Failed to retrieve connection",
			"connection_id", id,
			"error", err,
		)
		return nil, model.NewAppError("connection.retrieval_failed", http.StatusInternalServerError)
	}

	workspace, err := a.Store.Collimato.GetWorkspaceByID(workspaceID)
	if err != nil {
		tlog.Errorw("Failed to retrieve workspace",
			"workspace_id", workspaceID,
			"error", err,
		)
		return nil, model.NewAppError("collimato.retrieval_failed", http.StatusInternalServerError)
	}

	if database.WorkspaceID != workspace.ID {
		return nil, model.NewAppError("connection.forbidden", http.StatusForbidden)
	}

	if conn.SSLMode != nil && !model.SSLModeSupported(*conn.SSLMode) {
		return nil, model.NewAppError("connection.ssl_mode_unsupported", http.StatusBadRequest)
	}

	stored := database.Password
	needsVerification := conn.NeedsVerification(*database)

	if err = database.Patch(conn); err != nil {
		tlog.Errorw("Failed to patch connection",
			"connection_id", id,
			"error", err,
		)
		return nil, model.NewAppError("connection.update_failed", http.StatusInternalServerError)
	}

	if !database.UsesCACertificate() {
		delete(database.Config, model.ConnectionCACertKey)
	}

	if needsVerification {
		password := database.Password
		if conn.Password == nil || *conn.Password == "" {
			if password, err = a.decryptConnectionPassword(stored); err != nil {
				tlog.Errorw("Failed to decrypt connection password",
					"connection_id", id,
					"workspace_id", workspaceID,
					"error", err,
				)
				return nil, model.NewAppError("connection.update_failed", http.StatusInternalServerError)
			}
		}

		verify := *database
		verify.Password = password

		if err = a.verifyConnection(verify); err != nil {
			tlog.Warnw("Connection verification failed",
				"connection_id", id,
				"workspace_id", workspaceID,
				"reason", queryengine.ConnectionFailureOf(err),
				"error", err,
			)
			return nil, verificationError(err)
		}
	}

	if conn.Password != nil && *conn.Password != "" {
		encrypted, err := a.encryptConnectionPassword(database.Password)
		if err != nil {
			tlog.Errorw("Failed to encrypt connection password",
				"connection_id", id,
				"workspace_id", workspaceID,
				"error", err,
			)
			return nil, model.NewAppError("connection.update_failed", http.StatusInternalServerError)
		}

		database.Password = encrypted
	}

	if err = a.Store.Collimato.UpdateConnection(*database); err != nil {
		tlog.Errorw("Failed to update connection",
			"connection_id", id,
			"workspace_id", workspaceID,
			"error", err,
		)
		return nil, model.NewAppError("connection.update_failed", http.StatusInternalServerError)
	}

	return database, nil
}

func (a *App) DeleteConnection(user model.User, workspaceID, id string) *model.AppError {
	if !a.CollimatoWorkspaceHasPermission(user, workspaceID, model.CollimatoPermissions.PermissionDeleteConnections) {
		return model.NewAppError("connection.forbidden", http.StatusForbidden)
	}

	database, err := a.Store.Collimato.GetConnectionByID(id)
	if err != nil {
		tlog.Errorw("Failed to retrieve connection",
			"connection_id", id,
			"workspace_id", workspaceID,
			"error", err,
		)
		return model.NewAppError("connection.retrieval_failed", http.StatusInternalServerError)
	}

	if database.WorkspaceID != workspaceID {
		return model.NewAppError("connection.forbidden", http.StatusForbidden)
	}

	if err := a.Store.Collimato.DeleteConnection(id); err != nil {
		tlog.Errorw("Failed to delete connection",
			"connection_id", id,
			"workspace_id", workspaceID,
			"error", err,
		)
		return model.NewAppError("connection.delete_failed", http.StatusInternalServerError)
	}

	// Deleting the default connection would leave the workspace with no
	// default, so a survivor becomes the default (best-effort).
	if database.Default {
		remaining, err := a.Store.Collimato.GetConnections(workspaceID)
		if err != nil {
			tlog.Warnw("Failed to load connections to promote a new default",
				"workspace_id", workspaceID,
				"error", err,
			)
		} else if len(remaining) > 0 {
			remaining[0].Default = true
			if err := a.Store.Collimato.UpdateConnection(remaining[0]); err != nil {
				tlog.Warnw("Failed to promote a new default connection",
					"workspace_id", workspaceID,
					"connection_id", remaining[0].ID,
					"error", err,
				)
			}
		}
	}

	return nil
}

func (a *App) AddWorkspaceUsers(user model.User, workspaceID string, users []string, roles []string) *model.AppError {
	if !a.CollimatoWorkspaceHasPermission(user, workspaceID, model.CollimatoPermissions.PermissionAddUsers) {
		return model.NewAppError("collimato.forbidden", http.StatusForbidden)
	}

	if !a.Server.License.HasCollimatoRoles() {
		roles = []string{model.CollimatoWorkspaceUserRoleId}
	}

	if appErr := a.guardCollimatoNewMembers(user, workspaceID, users, roles); appErr != nil {
		return appErr
	}

	if err := a.Store.Collimato.AddWorkspaceUsers(workspaceID, users, roles); err != nil {
		tlog.Errorw("Failed to add users to workspace",
			"workspace_id", workspaceID,
			"error", err,
		)
		return model.NewAppError("collimato.member_add_failed", http.StatusInternalServerError)
	}

	return nil
}

func (a *App) RemoveWorkspaceUser(user model.User, workspaceID, userID string) *model.AppError {
	if !a.CollimatoWorkspaceHasPermission(user, workspaceID, model.CollimatoPermissions.PermissionDeleteUsers) {
		return model.NewAppError("collimato.forbidden", http.StatusForbidden)
	}

	if err := a.Store.Collimato.RemoveWorkspaceUser(workspaceID, userID); err != nil {
		tlog.Errorw("Failed to remove user from workspace",
			"workspace_id", workspaceID,
			"user_id", userID,
			"error", err,
		)
		return model.NewAppError("collimato.member_remove_failed", http.StatusInternalServerError)
	}

	return nil
}

func (a *App) GetWorkspaceUsers(user model.User, workspaceID string) ([]model.CollimatoWorkspaceUser, *model.AppError) {
	ctx := context.Background()

	// Visibility gate: any member (direct or via group) can see the roster.
	roleNames, err := a.Store.Collimato.GetEffectiveRolesForUser(ctx, workspaceID, user.ID)
	if err != nil {
		tlog.Errorw("Failed to resolve effective workspace roles",
			"workspace_id", workspaceID,
			"user_id", user.ID,
			"error", err,
		)
		return nil, model.NewAppError("collimato.users_retrieval_failed", http.StatusInternalServerError)
	}

	if len(roleNames) == 0 {
		return nil, model.NewAppError("collimato.forbidden", http.StatusForbidden)
	}

	// Direct members.
	users, err := a.Store.Collimato.GetWorkspaceUsers(workspaceID)
	if err != nil {
		tlog.Errorw("Failed to retrieve workspace users",
			"workspace_id", workspaceID,
			"error", err,
		)
		return nil, model.NewAppError("collimato.users_retrieval_failed", http.StatusInternalServerError)
	}

	// Members reachable only through an attached group (flagged ViaGroup).
	viaGroup, err := a.Store.Collimato.GetWorkspaceGroupMembers(ctx, workspaceID)
	if err != nil {
		tlog.Errorw("Failed to retrieve workspace group members",
			"workspace_id", workspaceID,
			"error", err,
		)
		return nil, model.NewAppError("collimato.users_retrieval_failed", http.StatusInternalServerError)
	}

	return append(users, viaGroup...), nil
}

func (a *App) AddGroupsToWorkspace(user model.User, workspaceID string, groupIDs []string, roles []string) ([]model.CollimatoWorkspaceGroup, *model.AppError) {
	if !a.Server.License.HasGroups() {
		return nil, model.NewAppError("groups.license_required", http.StatusPaymentRequired)
	}

	if !a.CollimatoWorkspaceHasPermission(user, workspaceID, model.CollimatoPermissions.PermissionAddUsers) {
		return nil, model.NewAppError("collimato.forbidden", http.StatusForbidden)
	}

	if appErr := a.guardCollimatoGroupRoles(user, workspaceID, groupIDs, roles); appErr != nil {
		return nil, appErr
	}

	return a.attachWorkspaceGroups(user, workspaceID, groupIDs, roles)
}

func (a *App) attachWorkspaceGroups(user model.User, workspaceID string, groupIDs []string, roles []string) ([]model.CollimatoWorkspaceGroup, *model.AppError) {
	if len(groupIDs) == 0 {
		return nil, model.NewAppError("collimato.no_groups", http.StatusBadRequest)
	}

	if len(roles) == 0 {
		// A group with no role grants nothing, so require at least one.
		return nil, model.NewAppError("collimato.no_roles", http.StatusBadRequest)
	}

	ctx := context.Background()
	groupIDs = DedupeIDs(groupIDs)
	roles = DedupeIDs(roles)

	// Every requested group must exist (and not be soft-deleted).
	groups, err := a.Store.Groups.GetByIDs(ctx, groupIDs)
	if err != nil {
		tlog.Errorw("Failed to retrieve groups",
			"workspace_id", workspaceID,
			"error", err,
		)
		return nil, model.NewAppError("group.retrieval_failed", http.StatusInternalServerError)
	}

	if len(groups) != len(groupIDs) {
		return nil, model.NewAppError("group.not_found", http.StatusNotFound)
	}

	// Every role name must be a real role in this workspace.
	wsRoles, err := a.Store.Collimato.GetRolesByName(roles, workspaceID)
	if err != nil {
		tlog.Errorw("Failed to retrieve workspace roles",
			"workspace_id", workspaceID,
			"error", err,
		)
		return nil, model.NewAppError("collimato.roles_failed", http.StatusInternalServerError)
	}

	if len(wsRoles) != len(roles) {
		return nil, model.NewAppError("collimato.role_not_found", http.StatusBadRequest)
	}

	if err := a.Store.Collimato.AddWorkspaceGroups(ctx, workspaceID, groupIDs, roles, user.ID); err != nil {
		tlog.Errorw("Failed to attach groups to workspace",
			"workspace_id", workspaceID,
			"error", err,
		)
		return nil, model.NewAppError("collimato.group_add_failed", http.StatusInternalServerError)
	}

	// Re-read so the response carries accurate ids, names, and member counts.
	all, err := a.Store.Collimato.GetWorkspaceGroups(ctx, workspaceID)
	if err != nil {
		tlog.Errorw("Failed to list workspace groups after add",
			"workspace_id", workspaceID,
			"error", err,
		)
		return nil, model.NewAppError("collimato.group_add_failed", http.StatusInternalServerError)
	}

	requested := make(map[string]struct{}, len(groupIDs))
	for _, gid := range groupIDs {
		requested[gid] = struct{}{}
	}

	added := make([]model.CollimatoWorkspaceGroup, 0, len(groupIDs))
	for _, g := range all {
		if _, ok := requested[g.GroupID]; ok {
			added = append(added, g)
		}
	}

	return added, nil
}

func (a *App) ListWorkspaceGroups(user model.User, workspaceID string) ([]model.CollimatoWorkspaceGroup, *model.AppError) {
	if !a.CollimatoWorkspaceHasPermission(user, workspaceID, model.CollimatoPermissions.PermissionAddUsers) {
		return nil, model.NewAppError("collimato.forbidden", http.StatusForbidden)
	}

	groups, err := a.Store.Collimato.GetWorkspaceGroups(context.Background(), workspaceID)
	if err != nil {
		tlog.Errorw("Failed to list workspace groups",
			"workspace_id", workspaceID,
			"error", err,
		)
		return nil, model.NewAppError("collimato.groups_retrieval_failed", http.StatusInternalServerError)
	}

	return groups, nil
}

func (a *App) RemoveGroupFromWorkspace(user model.User, workspaceID, groupID string) *model.AppError {
	if !a.CollimatoWorkspaceHasPermission(user, workspaceID, model.CollimatoPermissions.PermissionDeleteUsers) {
		return model.NewAppError("collimato.forbidden", http.StatusForbidden)
	}

	if err := a.Store.Collimato.RemoveWorkspaceGroup(context.Background(), workspaceID, groupID); err != nil {
		tlog.Errorw("Failed to remove group from workspace",
			"workspace_id", workspaceID,
			"group_id", groupID,
			"error", err,
		)
		return model.NewAppError("collimato.group_remove_failed", http.StatusInternalServerError)
	}

	return nil
}

func (a *App) GetCurrentWorkspaceUser(user model.User, workspaceID string) (*model.CollimatoWorkspaceUser, *model.AppError) {
	workspaceUser, err := a.Store.Collimato.GetWorkspaceUserByUserID(user.ID, workspaceID)
	if err != nil {
		tlog.Errorw("Failed to retrieve workspace member",
			"workspace_id", workspaceID,
			"user_id", user.ID,
			"error", err,
		)
		return nil, model.NewAppError("collimato.member_not_found", http.StatusInternalServerError)
	}

	roleNames, err := a.Store.Collimato.GetEffectiveRolesForUser(context.Background(), workspaceID, user.ID)
	if err != nil {
		tlog.Errorw("Failed to resolve effective workspace roles",
			"workspace_id", workspaceID,
			"user_id", user.ID,
			"error", err,
		)
		return nil, model.NewAppError("collimato.roles_failed", http.StatusInternalServerError)
	}

	// Not a direct member and not reachable through any attached group.
	if workspaceUser == nil && len(roleNames) == 0 {
		return nil, model.NewAppError("collimato.forbidden", http.StatusForbidden)
	}

	// Group-only member: no direct collimato_workspace_users row exists, so
	// synthesize one from the effective roles for a consistent response shape.
	if workspaceUser == nil {
		workspaceUser = &model.CollimatoWorkspaceUser{
			WorkspaceID: workspaceID,
			UserID:      user.ID,
		}
	}

	workspaceUser.Role = strings.Join(roleNames, ",")

	roles, err := a.Store.Collimato.GetRolesByName(roleNames, workspaceID)
	if err != nil {
		tlog.Errorw("Failed to retrieve workspace roles",
			"workspace_id", workspaceID,
			"user_id", user.ID,
			"error", err,
		)
		return nil, model.NewAppError("collimato.roles_failed", http.StatusInternalServerError)
	}

	for _, role := range roles {
		workspaceUser.Permissions = append(workspaceUser.Permissions, role.Permissions...)
	}

	return workspaceUser, nil
}

func (a *App) CreateChart(user model.User, workspaceID string, chart model.NewChart) (*model.Chart, *model.AppError) {
	if !a.CollimatoWorkspaceHasPermission(user, workspaceID, model.CollimatoPermissions.PermissionCreateCharts) {
		return nil, model.NewAppError("chart.forbidden", http.StatusForbidden)
	}

	workspace, err := a.Store.Collimato.GetWorkspaceByID(workspaceID)
	if err != nil {
		tlog.Errorw("Failed to retrieve workspace",
			"workspace_id", workspaceID,
			"error", err,
		)
		return nil, model.NewAppError("collimato.retrieval_failed", http.StatusInternalServerError)
	}

	chart.OwnerID = user.ID
	chart.WorkspaceID = workspace.ID

	c, err := a.Store.Collimato.CreateChart(chart)
	if err != nil {
		tlog.Errorw("Failed to create chart",
			"workspace_id", workspaceID,
			"user_id", user.ID,
			"error", err,
		)
		return nil, model.NewAppError("chart.create_failed", http.StatusInternalServerError)
	}

	a.graphChartSaved(*c)

	return c, nil
}

func (a *App) GetCharts(user model.User, workspaceID string) ([]model.Chart, *model.AppError) {
	if !a.CollimatoWorkspaceHasPermission(user, workspaceID, model.CollimatoPermissions.PermissionViewCharts) {
		return nil, model.NewAppError("chart.forbidden", http.StatusForbidden)
	}

	charts, err := a.Store.Collimato.GetCharts(workspaceID)
	if err != nil {
		tlog.Errorw("Failed to retrieve charts",
			"workspace_id", workspaceID,
			"error", err,
		)
		return nil, model.NewAppError("chart.retrieval_failed", http.StatusInternalServerError)
	}

	verifiedCharts := make([]model.Chart, 0, len(charts))
	for _, chart := range charts {
		raw, err := json.Marshal(chart.Data["query"])
		if err != nil {
			tlog.Errorw("Failed to marshal chart query",
				"workspace_id", workspaceID,
				"chart_id", chart.ID,
				"error", err,
			)
			return nil, model.NewAppError("chart.retrieval_failed", http.StatusInternalServerError)
		}

		query := model.DataQuery{}
		if err = json.Unmarshal(raw, &query); err != nil {
			tlog.Errorw("Failed to unmarshal chart query",
				"workspace_id", workspaceID,
				"chart_id", chart.ID,
				"error", err,
			)
			return nil, model.NewAppError("chart.retrieval_failed", http.StatusInternalServerError)
		}

		ok, appErr := a.CollimatoHasPermissionToFields(user, workspaceID, &query)
		if appErr != nil {
			tlog.Warnw("Failed to verify chart field permissions",
				"workspace_id", workspaceID,
				"chart_id", chart.ID,
				"error", appErr.Message,
			)
		}

		if ok {
			verifiedCharts = append(verifiedCharts, chart)
		}
	}

	return verifiedCharts, nil
}

func (a *App) GetChartByID(user model.User, workspaceID, chartID string) (*model.Chart, *model.AppError) {
	if !a.CollimatoWorkspaceHasPermission(user, workspaceID, model.CollimatoPermissions.PermissionViewCharts) {
		return nil, model.NewAppError("chart.forbidden", http.StatusForbidden)
	}

	chart, err := a.Store.Collimato.GetChartByID(chartID)
	if err != nil {
		tlog.Errorw("Failed to retrieve chart",
			"workspace_id", workspaceID,
			"chart_id", chartID,
			"error", err,
		)
		return nil, model.NewAppError("chart.retrieval_failed", http.StatusInternalServerError)
	}

	if chart.WorkspaceID != workspaceID {
		return nil, model.NewAppError("chart.forbidden", http.StatusForbidden)
	}

	raw, err := json.Marshal(chart.Data["query"])
	if err != nil {
		tlog.Errorw("Failed to marshal chart query",
			"workspace_id", workspaceID,
			"chart_id", chartID,
			"error", err,
		)
		return nil, model.NewAppError("chart.retrieval_failed", http.StatusInternalServerError)
	}

	query := model.DataQuery{}
	if err = json.Unmarshal(raw, &query); err != nil {
		tlog.Errorw("Failed to unmarshal chart query",
			"workspace_id", workspaceID,
			"chart_id", chartID,
			"error", err,
		)
		return nil, model.NewAppError("chart.retrieval_failed", http.StatusInternalServerError)
	}

	ok, appErr := a.CollimatoHasPermissionToFields(user, workspaceID, &query)
	if appErr != nil {
		tlog.Errorw("Failed to verify chart field permissions",
			"workspace_id", workspaceID,
			"chart_id", chartID,
			"error", appErr.Message,
		)
		return nil, appErr
	}

	if !ok {
		return nil, model.NewAppError("chart.forbidden", http.StatusForbidden)
	}

	return chart, nil
}

func (a *App) UpdateChart(user model.User, workspaceID, id string, chartPatch model.ChartPatch) (*model.Chart, *model.AppError) {
	if !a.CollimatoWorkspaceHasPermission(user, workspaceID, model.CollimatoPermissions.PermissionEditCharts) {
		return nil, model.NewAppError("chart.forbidden", http.StatusForbidden)
	}

	chart, err := a.Store.Collimato.GetChartByID(id)
	if err != nil {
		tlog.Errorw("Failed to retrieve chart",
			"workspace_id", workspaceID,
			"chart_id", id,
			"error", err,
		)
		return nil, model.NewAppError("chart.retrieval_failed", http.StatusInternalServerError)
	}

	if chart.WorkspaceID != workspaceID {
		return nil, model.NewAppError("chart.forbidden", http.StatusForbidden)
	}

	chart.Patch(chartPatch)

	updatedChart, err := a.Store.Collimato.UpdateChart(id, *chart)
	if err != nil {
		tlog.Errorw("Failed to update chart",
			"workspace_id", workspaceID,
			"chart_id", id,
			"error", err,
		)
		return nil, model.NewAppError("chart.update_failed", http.StatusInternalServerError)
	}

	a.graphChartSaved(*updatedChart)

	return updatedChart, nil
}

func (a *App) DeleteChart(user model.User, workspaceID, id string) *model.AppError {
	if !a.CollimatoWorkspaceHasPermission(user, workspaceID, model.CollimatoPermissions.PermissionDeleteCharts) {
		return model.NewAppError("chart.forbidden", http.StatusForbidden)
	}

	chart, err := a.Store.Collimato.GetChartByID(id)
	if err != nil {
		tlog.Errorw("Failed to retrieve chart",
			"workspace_id", workspaceID,
			"chart_id", id,
			"error", err,
		)
		return model.NewAppError("chart.retrieval_failed", http.StatusInternalServerError)
	}

	if chart.WorkspaceID != workspaceID {
		return model.NewAppError("chart.forbidden", http.StatusForbidden)
	}

	if err := a.Store.Collimato.DeleteChart(id); err != nil {
		tlog.Errorw("Failed to delete chart",
			"workspace_id", workspaceID,
			"chart_id", id,
			"error", err,
		)
		return model.NewAppError("chart.delete_failed", http.StatusInternalServerError)
	}

	a.graphChartDeleted(id)

	return nil
}

func (a *App) GetDashboards(user model.User, workspaceID string) ([]model.Dashboard, *model.AppError) {
	if !a.CollimatoWorkspaceHasPermission(user, workspaceID, model.CollimatoPermissions.PermissionViewDashboards) {
		return nil, model.NewAppError("dashboard.forbidden", http.StatusForbidden)
	}

	dashboards, err := a.Store.Collimato.GetDashboards(workspaceID)
	if err != nil {
		tlog.Errorw("Failed to retrieve dashboards",
			"workspace_id", workspaceID,
			"error", err,
		)
		return nil, model.NewAppError("dashboard.retrieval_failed", http.StatusInternalServerError)
	}

	return dashboards, nil
}

func (a *App) GetDashboardByID(user model.User, workspaceID, id string) (*model.Dashboard, *model.AppError) {
	if !a.CollimatoWorkspaceHasPermission(user, workspaceID, model.CollimatoPermissions.PermissionViewDashboards) {
		return nil, model.NewAppError("dashboard.forbidden", http.StatusForbidden)
	}

	dashboard, err := a.Store.Collimato.GetDashboardByID(id)
	if err != nil {
		tlog.Errorw("Failed to retrieve dashboard",
			"workspace_id", workspaceID,
			"dashboard_id", id,
			"error", err,
		)
		return nil, model.NewAppError("dashboard.retrieval_failed", http.StatusInternalServerError)
	}

	if dashboard.WorkspaceID != workspaceID {
		return nil, model.NewAppError("dashboard.forbidden", http.StatusForbidden)
	}

	if !a.CollimatoWorkspaceHasPermission(user, workspaceID, model.CollimatoPermissions.PermissionViewDashboardFilters) {
		dashboard.Filters = make([]model.DashboardFilter, 0)
	}

	return dashboard, nil
}

func (a *App) CreateDashboard(user model.User, workspaceID string, dashboard model.Dashboard) (*model.Dashboard, *model.AppError) {
	if !a.CollimatoWorkspaceHasPermission(user, workspaceID, model.CollimatoPermissions.PermissionCreateDashboards) {
		return nil, model.NewAppError("dashboard.forbidden", http.StatusForbidden)
	}

	workspace, err := a.Store.Collimato.GetWorkspaceByID(workspaceID)
	if err != nil {
		tlog.Errorw("Failed to retrieve workspace",
			"workspace_id", workspaceID,
			"error", err,
		)
		return nil, model.NewAppError("collimato.retrieval_failed", http.StatusInternalServerError)
	}

	dashboard.OwnerID = user.ID
	dashboard.WorkspaceID = workspace.ID

	d, err := a.Store.Collimato.CreateDashboard(dashboard)
	if err != nil {
		tlog.Errorw("Failed to create dashboard",
			"workspace_id", workspaceID,
			"user_id", user.ID,
			"error", err,
		)
		return nil, model.NewAppError("dashboard.create_failed", http.StatusInternalServerError)
	}

	return d, nil
}

func (a *App) AddDashboardFilter(user model.User, workspaceID, dashboardID string, filter model.DashboardFilter) (*model.DashboardFilter, *model.AppError) {
	if !a.CollimatoWorkspaceHasPermission(user, workspaceID, model.CollimatoPermissions.PermissionCreateDashboardFilters) {
		return nil, model.NewAppError("dashboard.forbidden", http.StatusForbidden)
	}

	workspace, err := a.Store.Collimato.GetWorkspaceByID(workspaceID)
	if err != nil {
		tlog.Errorw("Failed to retrieve workspace",
			"workspace_id", workspaceID,
			"error", err,
		)
		return nil, model.NewAppError("collimato.retrieval_failed", http.StatusInternalServerError)
	}

	dashboard, err := a.Store.Collimato.GetDashboardByID(dashboardID)
	if err != nil {
		tlog.Errorw("Failed to retrieve dashboard",
			"workspace_id", workspaceID,
			"dashboard_id", dashboardID,
			"error", err,
		)
		return nil, model.NewAppError("dashboard.retrieval_failed", http.StatusInternalServerError)
	}

	if workspace.ID != dashboard.WorkspaceID {
		return nil, model.NewAppError("dashboard.workspace_mismatch", http.StatusForbidden)
	}

	newFilter, err := a.Store.Collimato.AddDashboardFilter(dashboard.ID, filter)
	if err != nil {
		tlog.Errorw("Failed to add dashboard filter",
			"workspace_id", workspaceID,
			"dashboard_id", dashboardID,
			"error", err,
		)
		return nil, model.NewAppError("dashboard.filter_add_failed", http.StatusInternalServerError)
	}

	return newFilter, nil
}

func (a *App) UpdateDashboardFilter(user model.User, workspaceID, dashboardID, filterID string, patch model.DashboardFilterPatch) (*model.DashboardFilter, *model.AppError) {
	if !a.CollimatoWorkspaceHasPermission(user, workspaceID, model.CollimatoPermissions.PermissionEditDashboardFilters) {
		return nil, model.NewAppError("dashboard.forbidden", http.StatusForbidden)
	}

	workspace, err := a.Store.Collimato.GetWorkspaceByID(workspaceID)
	if err != nil {
		tlog.Errorw("Failed to retrieve workspace",
			"workspace_id", workspaceID,
			"error", err,
		)
		return nil, model.NewAppError("collimato.retrieval_failed", http.StatusInternalServerError)
	}

	dashboard, err := a.Store.Collimato.GetDashboardByID(dashboardID)
	if err != nil {
		tlog.Errorw("Failed to retrieve dashboard",
			"workspace_id", workspaceID,
			"dashboard_id", dashboardID,
			"error", err,
		)
		return nil, model.NewAppError("dashboard.retrieval_failed", http.StatusInternalServerError)
	}

	if workspace.ID != dashboard.WorkspaceID {
		return nil, model.NewAppError("dashboard.workspace_mismatch", http.StatusForbidden)
	}

	filter, err := a.Store.Collimato.GetDashboardFilterByID(dashboard.ID, filterID)
	if err != nil {
		tlog.Errorw("Failed to retrieve dashboard filter",
			"workspace_id", workspaceID,
			"dashboard_id", dashboardID,
			"filter_id", filterID,
			"error", err,
		)
		return nil, model.NewAppError("dashboard.filter_not_found", http.StatusInternalServerError)
	}

	filter.Patch(patch)

	if err = a.Store.Collimato.UpdateDashboardFilter(*filter); err != nil {
		tlog.Errorw("Failed to update dashboard filter",
			"workspace_id", workspaceID,
			"dashboard_id", dashboardID,
			"filter_id", filterID,
			"error", err,
		)
		return nil, model.NewAppError("dashboard.filter_update_failed", http.StatusInternalServerError)
	}

	return filter, nil
}

func (a *App) DeleteDashboardFilter(user model.User, workspaceID, dashboardID, filterID string) *model.AppError {
	if !a.CollimatoWorkspaceHasPermission(user, workspaceID, model.CollimatoPermissions.PermissionDeleteDashboardFilters) {
		return model.NewAppError("dashboard.forbidden", http.StatusForbidden)
	}

	workspace, err := a.Store.Collimato.GetWorkspaceByID(workspaceID)
	if err != nil {
		tlog.Errorw("Failed to retrieve workspace",
			"workspace_id", workspaceID,
			"error", err,
		)
		return model.NewAppError("collimato.retrieval_failed", http.StatusInternalServerError)
	}

	dashboard, err := a.Store.Collimato.GetDashboardByID(dashboardID)
	if err != nil {
		tlog.Errorw("Failed to retrieve dashboard",
			"workspace_id", workspaceID,
			"dashboard_id", dashboardID,
			"error", err,
		)
		return model.NewAppError("dashboard.retrieval_failed", http.StatusInternalServerError)
	}

	if workspace.ID != dashboard.WorkspaceID {
		return model.NewAppError("dashboard.workspace_mismatch", http.StatusForbidden)
	}

	if err = a.Store.Collimato.DeleteDashboardFilter(dashboardID, filterID); err != nil {
		tlog.Errorw("Failed to delete dashboard filter",
			"workspace_id", workspaceID,
			"dashboard_id", dashboardID,
			"filter_id", filterID,
			"error", err,
		)
		return model.NewAppError("dashboard.filter_delete_failed", http.StatusInternalServerError)
	}

	return nil
}

func (a *App) UpdateDashboardCharts(user model.User, workspaceID, id string, patch []model.DashboardChartPatch) *model.AppError {
	if !a.CollimatoWorkspaceHasPermission(user, workspaceID, model.CollimatoPermissions.PermissionEditDashboards) {
		return model.NewAppError("dashboard.forbidden", http.StatusForbidden)
	}

	dashboard, err := a.Store.Collimato.GetDashboardByID(id)
	if err != nil {
		tlog.Errorw("Failed to retrieve dashboard",
			"workspace_id", workspaceID,
			"dashboard_id", id,
			"error", err,
		)
		return model.NewAppError("dashboard.retrieval_failed", http.StatusInternalServerError)
	}

	if dashboard.WorkspaceID != workspaceID {
		return model.NewAppError("dashboard.forbidden", http.StatusForbidden)
	}

	charts, err := a.Store.Collimato.GetCharts(workspaceID)
	if err != nil {
		tlog.Errorw("Failed to retrieve charts",
			"workspace_id", workspaceID,
			"error", err,
		)
		return model.NewAppError("dashboard.chart_update_failed", http.StatusInternalServerError)
	}

	for _, item := range patch {
		inWorkspace := slices.ContainsFunc(charts, func(c model.Chart) bool { return c.ID == item.ID })
		if !inWorkspace {
			return model.NewAppError("dashboard.forbidden", http.StatusForbidden)
		}
	}

	if err = a.Store.Collimato.AddDashboardCharts(user.ID, dashboard.ID, patch); err != nil {
		tlog.Errorw("Failed to update dashboard charts",
			"workspace_id", workspaceID,
			"dashboard_id", dashboard.ID,
			"user_id", user.ID,
			"error", err,
		)
		return model.NewAppError("dashboard.chart_update_failed", http.StatusInternalServerError)
	}

	return nil
}

func (a *App) UpdateDashboard(user model.User, workspaceID, dashboardID string, patch model.DashboardPatch) (*model.Dashboard, *model.AppError) {
	if !a.CollimatoWorkspaceHasPermission(user, workspaceID, model.CollimatoPermissions.PermissionEditDashboards) {
		return nil, model.NewAppError("dashboard.forbidden", http.StatusForbidden)
	}

	dashboard, err := a.Store.Collimato.GetDashboardByID(dashboardID)
	if err != nil {
		tlog.Errorw("Failed to retrieve dashboard",
			"workspace_id", workspaceID,
			"dashboard_id", dashboardID,
			"error", err,
		)
		return nil, model.NewAppError("dashboard.retrieval_failed", http.StatusInternalServerError)
	}

	if dashboard.WorkspaceID != workspaceID {
		return nil, model.NewAppError("dashboard.forbidden", http.StatusForbidden)
	}

	dashboard.Patch(patch)

	updatedDashboard, err := a.Store.Collimato.UpdateDashboard(*dashboard)
	if err != nil {
		tlog.Errorw("Failed to update dashboard",
			"workspace_id", workspaceID,
			"dashboard_id", dashboardID,
			"error", err,
		)
		return nil, model.NewAppError("dashboard.update_failed", http.StatusInternalServerError)
	}

	return updatedDashboard, nil
}

func (a *App) DeleteDashboard(user model.User, workspaceID, id string) *model.AppError {
	if !a.CollimatoWorkspaceHasPermission(user, workspaceID, model.CollimatoPermissions.PermissionDeleteDashboards) {
		return model.NewAppError("dashboard.forbidden", http.StatusForbidden)
	}

	dashboard, err := a.Store.Collimato.GetDashboardByID(id)
	if err != nil {
		tlog.Errorw("Failed to retrieve dashboard",
			"workspace_id", workspaceID,
			"dashboard_id", id,
			"error", err,
		)
		return model.NewAppError("dashboard.retrieval_failed", http.StatusInternalServerError)
	}

	if dashboard.WorkspaceID != workspaceID {
		return model.NewAppError("dashboard.forbidden", http.StatusForbidden)
	}

	if err = a.Store.Collimato.DeleteDashboard(dashboard.ID); err != nil {
		tlog.Errorw("Failed to delete dashboard",
			"workspace_id", workspaceID,
			"dashboard_id", dashboard.ID,
			"error", err,
		)
		return model.NewAppError("dashboard.delete_failed", http.StatusInternalServerError)
	}

	return nil
}

// GetWorkspaceRoles also serves users who may only assign roles, since they
// need the names to pick from, but leaves out what each role grants.
func (a *App) GetWorkspaceRoles(user model.User, workspaceID string) ([]model.CollimatoRole, *model.AppError) {
	canView := a.CollimatoWorkspaceHasPermission(user, workspaceID, model.CollimatoPermissions.PermissionViewRoles)
	canAssign := a.CollimatoWorkspaceHasPermission(user, workspaceID, model.CollimatoPermissions.PermissionAssignRoles)

	if !canView && !canAssign {
		return nil, model.NewAppError("role.forbidden", http.StatusForbidden)
	}

	roles, err := a.Store.Collimato.GetWorkspaceRoles(workspaceID)
	if err != nil {
		tlog.Errorw("Failed to retrieve workspace roles",
			"workspace_id", workspaceID,
			"error", err,
		)
		return nil, model.NewAppError("role.retrieval_failed", http.StatusInternalServerError)
	}

	if canView {
		return roles, nil
	}

	names := make([]model.CollimatoRole, len(roles))
	for i, role := range roles {
		names[i] = model.CollimatoRole{
			ID:          role.ID,
			WorkspaceID: role.WorkspaceID,
			Name:        role.Name,
			DisplayName: role.DisplayName,
			Description: role.Description,
		}
	}

	return names, nil
}

func (a *App) GetWorkspaceRole(user model.User, workspaceID, roleID string) (*model.CollimatoRole, *model.AppError) {
	if !a.CollimatoWorkspaceHasPermission(user, workspaceID, model.CollimatoPermissions.PermissionViewRoles) {
		return nil, model.NewAppError("role.forbidden", http.StatusForbidden)
	}

	role, err := a.Store.Collimato.GetWorkspaceRoleByID(workspaceID, roleID)
	if err != nil {
		tlog.Errorw("Failed to retrieve workspace role",
			"workspace_id", workspaceID,
			"role_id", roleID,
			"error", err,
		)
		return nil, model.NewAppError("role.retrieval_failed", http.StatusInternalServerError)
	}

	return role, nil
}

func (a *App) UpdateWorkspaceRolePermissions(workspace model.CollimatoWorkspace) error {
	names, err := a.GetDatasetNames(workspace)
	if err != nil {
		return err
	}

	roles, err := a.Store.Collimato.GetWorkspaceRoles(workspace.ID)
	if err != nil {
		return err
	}

	for _, role := range roles {
		if role.AutoUpdate {
			role.TablePermissions = names

			err = a.Store.Collimato.UpdateWorkspaceRole(workspace.ID, role)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (a *App) GetWorkspaceFieldList(ctx context.Context, user model.User, workspaceID string) (*collimato.FieldList, *model.AppError) {
	if !a.canQueryWorkspaceData(user, workspaceID) {
		return nil, model.NewAppError("collimato.forbidden", http.StatusForbidden)
	}

	return a.filteredWorkspaceFieldList(ctx, user, workspaceID)
}

func (a *App) filteredWorkspaceFieldList(ctx context.Context, user model.User, workspaceID string) (*collimato.FieldList, *model.AppError) {
	fieldList, appErr := a.fetchWorkspaceFieldList(ctx, workspaceID)
	if appErr != nil {
		return nil, appErr
	}

	// Which models cannot be used, and why, is part of the data model.
	if !a.CollimatoWorkspaceHasPermission(user, workspaceID, model.CollimatoPermissions.PermissionViewDataModels) {
		fieldList.ModelErrors = nil
	}

	roleNames, err := a.Store.Collimato.GetEffectiveRolesForUser(context.Background(), workspaceID, user.ID)
	if err != nil {
		tlog.Errorw("Failed to resolve effective workspace roles",
			"workspace_id", workspaceID,
			"user_id", user.ID,
			"error", err,
		)
		return nil, model.NewAppError("collimato.workspace_user_failed", http.StatusInternalServerError)
	}

	if len(roleNames) == 0 {
		return nil, model.NewAppError("collimato.forbidden", http.StatusForbidden)
	}

	roles, err := a.Store.Collimato.GetRolesByName(roleNames, workspaceID)
	if err != nil {
		tlog.Errorw("Failed to retrieve workspace roles",
			"workspace_id", workspaceID,
			"user_id", user.ID,
			"error", err,
		)
		return nil, model.NewAppError("collimato.roles_failed", http.StatusInternalServerError)
	}

	if a.CollimatoRoles == nil {
		if rolesCarryDataSecurity(roles) {
			return nil, model.NewAppError("collimato.data_security_unavailable", http.StatusPaymentRequired)
		}

		return fieldList, nil
	}

	return a.CollimatoRoles.FilterFieldList(roles, *fieldList), nil
}

// GetWorkspaceRoleFieldList gives whoever creates or edits roles every
// dataset and field unfiltered, because a role cannot restrict a field its
// editor cannot see, and they could grant themselves full access anyway.
// Someone who only views roles gets no more than their own roles allow.
func (a *App) GetWorkspaceRoleFieldList(ctx context.Context, user model.User, workspaceID string) (*collimato.FieldList, *model.AppError) {
	canCreate := a.CollimatoWorkspaceHasPermission(user, workspaceID, model.CollimatoPermissions.PermissionCreateRoles)
	canEdit := a.CollimatoWorkspaceHasPermission(user, workspaceID, model.CollimatoPermissions.PermissionEditRoles)

	if canCreate || canEdit {
		fieldList, appErr := a.fetchWorkspaceFieldList(ctx, workspaceID)
		if appErr != nil {
			return nil, appErr
		}

		// Which models cannot be used, and why, is part of the data model.
		if !a.CollimatoWorkspaceHasPermission(user, workspaceID, model.CollimatoPermissions.PermissionViewDataModels) {
			fieldList.ModelErrors = nil
		}

		return fieldList, nil
	}

	if !a.CollimatoWorkspaceHasPermission(user, workspaceID, model.CollimatoPermissions.PermissionViewRoles) {
		return nil, model.NewAppError("collimato.forbidden", http.StatusForbidden)
	}

	return a.filteredWorkspaceFieldList(ctx, user, workspaceID)
}

// canQueryWorkspaceData lets dashboard viewers load the data behind the charts
// they are shown. Table, column and row rules still decide what they get.
func (a *App) canQueryWorkspaceData(user model.User, workspaceID string) bool {
	if a.CollimatoWorkspaceHasPermission(user, workspaceID, model.CollimatoPermissions.PermissionViewCharts) {
		return true
	}

	return a.CollimatoWorkspaceHasPermission(user, workspaceID, model.CollimatoPermissions.PermissionViewDashboards)
}

func (a *App) fetchWorkspaceFieldList(ctx context.Context, workspaceID string) (*collimato.FieldList, *model.AppError) {
	workspace, err := a.Store.Collimato.GetWorkspaceByID(workspaceID)
	if err != nil {
		tlog.Errorw("Failed to retrieve workspace",
			"workspace_id", workspaceID,
			"error", err,
		)
		return nil, model.NewAppError("collimato.retrieval_failed", http.StatusInternalServerError)
	}

	fieldList, err := a.QueryEngine.FieldList(ctx, workspace.ID)
	if err != nil {
		tlog.Errorw("Failed to list a workspace's fields",
			"workspace_id", workspaceID,
			"error", err,
		)
		return nil, model.NewAppError("collimato.request_failed", http.StatusInternalServerError)
	}

	return fieldList, nil
}

func (a *App) GetWorkspaceSQL(ctx context.Context, user model.User, workspaceID string, query model.DataQuery) (string, *model.AppError) {
	if !a.CollimatoWorkspaceHasPermission(user, workspaceID, model.CollimatoPermissions.PermissionViewCharts) {
		return "", model.NewAppError("collimato.forbidden", http.StatusForbidden)
	}

	// The generated SQL names source tables and column expressions, which are
	// part of the data model.
	if !a.CollimatoWorkspaceHasPermission(user, workspaceID, model.CollimatoPermissions.PermissionViewDataModels) {
		return "", model.NewAppError("collimato.forbidden", http.StatusForbidden)
	}

	wsp, appErr := a.GetWorkspaceByID(user, workspaceID)
	if appErr != nil {
		return "", appErr
	}

	ok, appErr := a.CollimatoHasPermissionToFields(user, wsp.ID, &query)
	if appErr != nil {
		return "", appErr
	}

	if !ok {
		return "", model.NewAppError("collimato.permission_denied", http.StatusForbidden)
	}

	query.SetLimit()

	sqlText, err := a.QueryEngine.SQL(ctx, wsp.ID, query)
	if err != nil {
		return "", a.queryAppError(user, workspaceID, err)
	}

	return sqlText, nil
}

// QueryData runs a query sent by the browser. Ungrouped is always switched
// off here: with it, a measure such as an average salary would come back as
// each person's own value, row by row. Only Twigex's own row reads use it.
func (a *App) QueryData(ctx context.Context, user model.User, workspaceID string, query model.DataQuery) (*collimato.Answer, *model.AppError) {
	query.Ungrouped = false

	return a.queryData(ctx, user, workspaceID, query)
}

func (a *App) queryData(ctx context.Context, user model.User, workspaceID string, query model.DataQuery) (*collimato.Answer, *model.AppError) {
	if !a.canQueryWorkspaceData(user, workspaceID) {
		return nil, model.NewAppError("collimato.forbidden", http.StatusForbidden)
	}

	wsp, appErr := a.GetWorkspaceByID(user, workspaceID)
	if appErr != nil {
		return nil, appErr
	}

	ok, appErr := a.CollimatoHasPermissionToFields(user, wsp.ID, &query)
	if appErr != nil {
		return nil, appErr
	}

	if !ok {
		return nil, model.NewAppError("collimato.permission_denied", http.StatusForbidden)
	}

	query.SetLimit()

	answer, err := a.QueryEngine.Query(ctx, wsp.ID, query)
	if err != nil {
		return nil, a.queryAppError(user, workspaceID, err)
	}

	return answer, nil
}

// queryAppError is the error a reader gets when a chart cannot be answered:
// the engine's reason when it has one, otherwise a failed query. The
// database's own message names tables and columns, which are part of the data
// model, so only people who can read data models get it.
func (a *App) queryAppError(user model.User, workspaceID string, err error) *model.AppError {
	var queryError *QueryError
	if errors.As(err, &queryError) {
		isDatabaseMessage := queryError.ID == ""
		canReadModels := a.CollimatoWorkspaceHasPermission(user, workspaceID, model.CollimatoPermissions.PermissionViewDataModels)

		if isDatabaseMessage && !canReadModels {
			return model.NewAppError("collimato.query_failed", http.StatusBadGateway)
		}

		return queryError.appError()
	}

	// The reader left, or the request ran out of time: expected, not worth
	// logging.
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return model.NewAppError("collimato.request_failed", http.StatusInternalServerError)
	}

	tlog.Errorw("Failed to answer a chart",
		"workspace_id", workspaceID,
		"error", err,
	)

	return model.NewAppError("collimato.query_failed", http.StatusInternalServerError)
}

func (a *App) GetTables(user model.User, workspaceID string) ([]*model.DatabaseTable, *model.AppError) {
	canCreate := a.CollimatoWorkspaceHasPermission(user, workspaceID, model.CollimatoPermissions.PermissionCreateDataModels)
	canEdit := a.CollimatoWorkspaceHasPermission(user, workspaceID, model.CollimatoPermissions.PermissionEditDataModels)

	if !canCreate && !canEdit {
		return nil, model.NewAppError("datamodel.forbidden", http.StatusForbidden)
	}

	databases, err := a.Store.Collimato.GetConnections(workspaceID)
	if err != nil {
		tlog.Errorw("Failed to retrieve connections",
			"workspace_id", workspaceID,
			"error", err,
		)
		return nil, model.NewAppError("connection.retrieval_failed", http.StatusInternalServerError)
	}

	var (
		mu     sync.Mutex
		wg     sync.WaitGroup
		tables = make([]*model.DatabaseTable, 0, len(databases))
	)

	wg.Add(len(databases))
	for _, v := range databases {
		go func(v model.Connection) {
			defer wg.Done()
			password, err := a.decryptConnectionPassword(v.Password)
			if err != nil {
				tlog.Warnw("Failed to decrypt connection password",
					"workspace_id", workspaceID,
					"connection_id", v.ID,
					"error", err,
				)
				return
			}

			v.Password = password

			ctx, cancel := a.dbCtx(context.Background())
			defer cancel()

			settings := collimato.EngineConnectionSettings(v, password)

			connection, err := engineConnector.Open(settings)
			if err != nil {
				tlog.Error("Failed to open connection",
					"workspace_id", workspaceID,
					"connection_id", v.ID,
					"error", err,
				)

				return
			}
			defer connection.Close()

			engineTables, err := queryengine.ListTables(ctx, connection, settings.DatabaseType)
			if err != nil {
				tlog.Error("Failed to list tables",
					"workspace_id", workspaceID,
					"connection_id", v.ID,
					"reason", queryengine.ConnectionFailureOf(err),
					"error", err,
				)

				return
			}

			dbTable := model.DatabaseTable{
				Name:     v.Database,
				Database: v.ID,
				Type:     v.Type,
				Tables:   toModelTables(v.Type, engineTables),
			}

			mu.Lock()
			tables = append(tables, &dbTable)
			mu.Unlock()
		}(v)
	}

	wg.Wait()

	return tables, nil
}

func toModelTables(connectionType string, engineTables []queryengine.Table) []model.TableName {
	tableNames := make([]model.TableName, 0, len(engineTables))

	for _, table := range engineTables {
		columns := make([]model.ColumnDetail, 0, len(table.Columns))
		for _, column := range table.Columns {
			columns = append(columns, model.ColumnDetail{
				Name:          column.Name,
				Type:          column.ColumnType,
				SuggestedType: collimato.SuggestBuilderType(connectionType, column.ColumnType),
			})
		}

		tableNames = append(tableNames, model.TableName{
			Name:    collimato.BuilderTableName(connectionType, table),
			Columns: columns,
		})
	}

	return tableNames
}
