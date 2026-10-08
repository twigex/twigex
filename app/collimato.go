// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/twigex/twigex/collimato"
	"github.com/twigex/twigex/crypto"
	"github.com/twigex/twigex/internal/catalog"
	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/tlog"
	yaml "go.yaml.in/yaml/v3"
)

var cubeHTTPClient = &http.Client{Timeout: 60 * time.Second}

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

// catalogConnection is the one place a stored connection becomes something the
// catalog can open, so verification and introspection cannot disagree about how
// to reach a database.
func catalogConnection(conn model.Connection, password string) catalog.Connection {
	return catalog.Connection{
		Type:     conn.Type,
		Host:     conn.Host,
		Port:     conn.Port,
		User:     conn.Username,
		Password: password,
		Database: conn.Database,
		SSLMode:  conn.SSLMode,
		Config:   conn.Config,
	}
}

func (a *App) verifyConnection(conn model.Connection) error {
	ctx, cancel := a.dbCtx(context.Background())
	defer cancel()

	cat, err := catalog.Open(ctx, catalogConnection(conn, conn.Password))
	if err != nil {
		return err
	}

	return cat.Close()
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
		"reason", catalog.Failure(err),
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
		delete(conn.Config, catalog.CACertKey)
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
			"reason", catalog.Failure(err),
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
		delete(database.Config, catalog.CACertKey)
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
				"reason", catalog.Failure(err),
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

	// Deleting the default connection would leave the workspace with no default,
	// so GetCubeConnection("default") 404s and every query fails. Promote a
	// survivor to default (best-effort).
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
	names, err := a.GetCubeNames(workspace)
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

func (a *App) GetWorkspaceMeta(ctx context.Context, user model.User, workspaceID string) (*collimato.CubeCollection, *model.AppError) {
	if !a.canQueryWorkspaceData(user, workspaceID) {
		return nil, model.NewAppError("collimato.forbidden", http.StatusForbidden)
	}

	return a.filteredWorkspaceMeta(ctx, user, workspaceID)
}

func (a *App) filteredWorkspaceMeta(ctx context.Context, user model.User, workspaceID string) (*collimato.CubeCollection, *model.AppError) {
	cubeCollection, appErr := a.fetchWorkspaceMeta(ctx, workspaceID)
	if appErr != nil {
		return nil, appErr
	}

	// A hand-written model can put anything in its meta block, so only data
	// model readers get it.
	if !a.CollimatoWorkspaceHasPermission(user, workspaceID, model.CollimatoPermissions.PermissionViewDataModels) {
		for i := range cubeCollection.Cubes {
			cubeCollection.Cubes[i].Meta = nil
		}
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

		return cubeCollection, nil
	}

	return a.CollimatoRoles.FilterMeta(roles, *cubeCollection), nil
}

// GetWorkspaceRoleMeta gives whoever creates or edits roles every cube and
// field unfiltered, because a role cannot restrict a field its editor cannot
// see, and they could grant themselves full access anyway. Someone who only
// views roles gets no more than their own roles allow.
func (a *App) GetWorkspaceRoleMeta(ctx context.Context, user model.User, workspaceID string) (*collimato.CubeCollection, *model.AppError) {
	canCreate := a.CollimatoWorkspaceHasPermission(user, workspaceID, model.CollimatoPermissions.PermissionCreateRoles)
	canEdit := a.CollimatoWorkspaceHasPermission(user, workspaceID, model.CollimatoPermissions.PermissionEditRoles)

	if canCreate || canEdit {
		return a.fetchWorkspaceMeta(ctx, workspaceID)
	}

	if !a.CollimatoWorkspaceHasPermission(user, workspaceID, model.CollimatoPermissions.PermissionViewRoles) {
		return nil, model.NewAppError("collimato.forbidden", http.StatusForbidden)
	}

	return a.filteredWorkspaceMeta(ctx, user, workspaceID)
}

// canQueryWorkspaceData lets dashboard viewers load the data behind the charts
// they are shown. Table, column and row rules still decide what they get.
func (a *App) canQueryWorkspaceData(user model.User, workspaceID string) bool {
	if a.CollimatoWorkspaceHasPermission(user, workspaceID, model.CollimatoPermissions.PermissionViewCharts) {
		return true
	}

	return a.CollimatoWorkspaceHasPermission(user, workspaceID, model.CollimatoPermissions.PermissionViewDashboards)
}

func (a *App) fetchWorkspaceMeta(ctx context.Context, workspaceID string) (*collimato.CubeCollection, *model.AppError) {
	workspace, err := a.Store.Collimato.GetWorkspaceByID(workspaceID)
	if err != nil {
		tlog.Errorw("Failed to retrieve workspace",
			"workspace_id", workspaceID,
			"error", err,
		)
		return nil, model.NewAppError("collimato.retrieval_failed", http.StatusInternalServerError)
	}

	cubeCollection, err := a.QueryEngine.Meta(ctx, workspace.ID)
	if err != nil {
		if ctx.Err() != nil {
			return nil, model.NewAppError("collimato.request_failed", http.StatusInternalServerError)
		}

		tlog.Errorw("Failed to fetch CubeJS meta",
			"workspace_id", workspaceID,
			"error", err,
		)
		return nil, model.NewAppError("collimato.request_failed", http.StatusInternalServerError)
	}

	return cubeCollection, nil
}

func (a *App) GetWorkspaceSQL(ctx context.Context, user model.User, workspaceID string, query model.DataQuery) (map[string]any, *model.AppError) {
	if !a.CollimatoWorkspaceHasPermission(user, workspaceID, model.CollimatoPermissions.PermissionViewCharts) {
		return nil, model.NewAppError("collimato.forbidden", http.StatusForbidden)
	}

	// The generated SQL names source tables and column expressions, which are
	// part of the data model.
	if !a.CollimatoWorkspaceHasPermission(user, workspaceID, model.CollimatoPermissions.PermissionViewDataModels) {
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

	body, err := a.QueryEngine.SQL(ctx, wsp.ID, query)
	if err != nil {
		tlog.Errorw("Failed to run CubeJS SQL",
			"workspace_id", workspaceID,
			"error", err,
		)
		return nil, model.NewAppError("collimato.query_failed", http.StatusInternalServerError)
	}

	return body, nil
}

func (a *App) QueryData(ctx context.Context, user model.User, workspaceID string, query model.DataQuery) (map[string]any, *model.AppError) {
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

	body, err := a.QueryEngine.Query(ctx, wsp.ID, query)
	if err != nil {
		// Client disconnected or cancelled (expected, not a server error).
		if ctx.Err() != nil {
			return nil, model.NewAppError("collimato.request_failed", http.StatusInternalServerError)
		}

		tlog.Errorw("Failed to query CubeJS data",
			"workspace_id", workspaceID,
			"error", err,
		)
		return nil, model.NewAppError("collimato.query_failed", http.StatusInternalServerError)
	}

	return body, nil
}

// GetCubeSchemaFiles returns a workspace's cube and view definition files in the
// layout the shared Cube.js server expects from its repositoryFactory. It is
// served over the internal Cube callback endpoint (shared-secret authenticated),
// so it takes no user and performs no per-user permission check.
func (a *App) GetCubeSchemaFiles(workspaceID string) ([]model.CubeSchemaFile, *model.AppError) {
	out := make([]model.CubeSchemaFile, 0)

	for _, ft := range []struct {
		fileType string
		dir      string
	}{
		{model.CollimatoFileTypeCube, "cubes"},
		{model.CollimatoFileTypeView, "views"},
	} {
		files, err := a.Store.Collimato.GetWorkspaceFiles(workspaceID, ft.fileType)
		if err != nil {
			tlog.Errorw("Failed to retrieve workspace schema files",
				"workspace_id", workspaceID,
				"file_type", ft.fileType,
				"error", err,
			)
			return nil, model.NewAppError("collimato.retrieval_failed", http.StatusInternalServerError)
		}

		for _, f := range files {
			out = append(out, model.CubeSchemaFile{
				FileName: path.Join(ft.dir, f.Name),
				Content:  f.Content,
			})
		}
	}

	return out, nil
}

// GetCubeSchemaVersion returns the latest modification time across a workspace's
// cube and view files. The shared Cube.js server polls this via schemaVersion and
// hot-reloads the data model whenever the value changes, so no restart is needed.
func (a *App) GetCubeSchemaVersion(workspaceID string) (int64, *model.AppError) {
	var version int64

	for _, fileType := range []string{model.CollimatoFileTypeCube, model.CollimatoFileTypeView} {
		files, err := a.Store.Collimato.GetWorkspaceFiles(workspaceID, fileType)
		if err != nil {
			tlog.Errorw("Failed to retrieve workspace schema files",
				"workspace_id", workspaceID,
				"file_type", fileType,
				"error", err,
			)
			return 0, model.NewAppError("collimato.retrieval_failed", http.StatusInternalServerError)
		}

		for _, f := range files {
			if f.UpdatedAt > version {
				version = f.UpdatedAt
			}
		}
	}

	return version, nil
}

// GetCubeConnection resolves the database connection a Cube.js dataSource maps to.
// A dataSource of "" or "default" selects the workspace's default connection; any
// other value matches a connection by its lowercased database name, the same
// naming used when building the CUBEJS_DATASOURCES list. Served over the internal
// Cube callback endpoint, so it takes no user and performs no per-user check.
func (a *App) GetCubeConnection(workspaceID, dataSource string) (*model.CubeConnection, *model.AppError) {
	conns, err := a.Store.Collimato.GetConnections(workspaceID)
	if err != nil {
		tlog.Errorw("Failed to retrieve connections",
			"workspace_id", workspaceID,
			"error", err,
		)
		return nil, model.NewAppError("connection.retrieval_failed", http.StatusInternalServerError)
	}

	for _, c := range conns {
		match := false
		if dataSource == "" || dataSource == "default" {
			match = c.Default
		} else {
			match = strings.ToLower(strings.ReplaceAll(c.Database, " ", "")) == dataSource
		}

		if match {
			password, err := a.decryptConnectionPassword(c.Password)
			if err != nil {
				tlog.Errorw("Failed to decrypt connection password",
					"workspace_id", workspaceID,
					"connection_id", c.ID,
					"error", err,
				)
				return nil, model.NewAppError("connection.retrieval_failed", http.StatusInternalServerError)
			}

			return &model.CubeConnection{
				Type:     c.Type,
				Host:     c.Host,
				Port:     c.Port,
				Database: c.Database,
				User:     c.Username,
				Password: password,
				SSLMode:  c.SSLMode,
				SSLCA:    c.Config[catalog.CACertKey],
			}, nil
		}
	}

	return nil, model.NewAppError("connection.not_found", http.StatusNotFound)
}

// CollimatoConfigured reports whether the shared Cube.js backend has been
// configured (an API URL is set). The frontend uses this to decide whether to
// surface Collimato analytics features.
func (a *App) CollimatoConfigured() bool {
	return *a.ConfigStore.Config.CollimatoSettings.CubeAPIURL != ""
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

			cat, err := catalog.Open(ctx, catalogConnection(v, password))
			if err != nil {
				tlog.Error("Failed to open catalog",
					"workspace_id", workspaceID,
					"connection_id", v.ID,
					"error", err,
				)

				return
			}

			catTables, err := cat.Tables(ctx)
			if err != nil {
				tlog.Error("Failed to retrieve catalog data",
					"workspace_id", workspaceID,
					"connection_id", v.ID,
					"error", err,
				)

				return
			}

			dbTable := model.DatabaseTable{
				Name:     v.Database,
				Database: v.ID,
				Tables:   toModelTables(catTables),
			}

			err = cat.Close()
			if err != nil {
				tlog.Error("Failed to close catalog",
					"workspace_id", workspaceID,
					"connection_id", v.ID,
					"error", err,
				)
			}

			mu.Lock()
			tables = append(tables, &dbTable)
			mu.Unlock()
		}(v)
	}

	wg.Wait()

	return tables, nil
}

func toModelTables(catalogTables []catalog.TableName) []model.TableName {
	tableNames := make([]model.TableName, 0, len(catalogTables))

	for _, v := range catalogTables {
		columns := make([]model.ColumnDetail, 0, len(v.Columns))
		for _, c := range v.Columns {
			columns = append(columns, model.ColumnDetail{
				Name: c.Name,
				Type: c.Type,
			})
		}

		tableNames = append(tableNames, model.TableName{
			Name:    v.Name,
			Columns: columns,
		})
	}

	return tableNames
}

func (a *App) NewCubeFile(user model.User, workspaceID, fileType, filename string) (*model.CubeFile, *model.AppError) {
	if !a.CollimatoWorkspaceHasPermission(user, workspaceID, model.CollimatoPermissions.PermissionCreateDataModels) {
		return nil, model.NewAppError("datamodel.forbidden", http.StatusForbidden)
	}

	if !isValidIdentifier(filename) {
		return nil, model.NewAppError("datamodel.invalid_model", http.StatusBadRequest)
	}

	if _, err := a.Store.Collimato.GetWorkspaceByID(workspaceID); err != nil {
		tlog.Errorw("Failed to retrieve workspace",
			"workspace_id", workspaceID,
			"error", err,
		)
		return nil, model.NewAppError("collimato.retrieval_failed", http.StatusInternalServerError)
	}

	taken, err := a.nameIsTaken(workspaceID, filename, fileType, false)
	if err != nil {
		tlog.Errorw("Failed to check data model name",
			"workspace_id", workspaceID,
			"name", filename,
			"error", err,
		)
		return nil, model.NewAppError("datamodel.file_create_failed", http.StatusInternalServerError)
	}

	if taken {
		return nil, model.NewAppError("datamodel.name_taken", http.StatusConflict)
	}

	subPath := "model/cubes"
	if fileType == "view" {
		subPath = "model/views"
	}

	filenameWithExt := filename + ".js"

	if _, err := a.Store.Collimato.UpdateWorkspaceFile(model.CollimatoWorkspaceFile{
		WorkspaceID: workspaceID,
		Name:        filenameWithExt,
		FileType:    fileType,
		Content:     "",
	}); err != nil {
		tlog.Errorw("Failed to create cube file",
			"workspace_id", workspaceID,
			"file", filenameWithExt,
			"error", err,
		)
		return nil, model.NewAppError("datamodel.file_create_failed", http.StatusInternalServerError)
	}

	return &model.CubeFile{
		Name:    filenameWithExt,
		Content: "",
		Path:    path.Join(subPath, filenameWithExt),
		Type:    fileType,
	}, nil
}

func (a *App) GetCubeFiles(user model.User, workspaceID string) ([]*model.CubeFile, *model.AppError) {
	if !a.CollimatoWorkspaceHasPermission(user, workspaceID, model.CollimatoPermissions.PermissionViewDataModels) {
		return nil, model.NewAppError("datamodel.forbidden", http.StatusForbidden)
	}

	workspace, err := a.Store.Collimato.GetWorkspaceByID(workspaceID)
	if err != nil {
		tlog.Errorw("Failed to retrieve workspace",
			"workspace_id", workspaceID,
			"error", err,
		)
		return nil, model.NewAppError("collimato.retrieval_failed", http.StatusInternalServerError)
	}

	cubeFiles := make([]*model.CubeFile, 0)

	for _, ft := range []struct{ fileType, subPath string }{
		{model.CollimatoFileTypeCube, "model/cubes"},
		{model.CollimatoFileTypeView, "model/views"},
	} {
		dbFiles, err := a.Store.Collimato.GetWorkspaceFiles(workspace.ID, ft.fileType)
		if err != nil {
			tlog.Errorw("Failed to retrieve workspace files",
				"workspace_id", workspaceID,
				"file_type", ft.fileType,
				"error", err,
			)
			return nil, model.NewAppError("datamodel.file_read_failed", http.StatusInternalServerError)
		}

		for _, f := range dbFiles {
			cubeFiles = append(cubeFiles, &model.CubeFile{
				Name:         f.Name,
				Content:      f.Content,
				Path:         "/" + ft.subPath + "/" + f.Name,
				Type:         ft.fileType,
				BuilderModel: f.BuilderModel,
			})
		}
	}

	return cubeFiles, nil
}

func (a *App) SaveCubeFile(user model.User, name, content, fileType, workspaceID string) (*model.CubeFile, *model.AppError) {
	if !a.CollimatoWorkspaceHasPermission(user, workspaceID, model.CollimatoPermissions.PermissionEditDataModels) {
		return nil, model.NewAppError("datamodel.forbidden", http.StatusForbidden)
	}

	workspace, err := a.Store.Collimato.GetWorkspaceByID(workspaceID)
	if err != nil {
		tlog.Errorw("Failed to retrieve workspace",
			"workspace_id", workspaceID,
			"error", err,
		)
		return nil, model.NewAppError("collimato.retrieval_failed", http.StatusInternalServerError)
	}

	subPath := "model/cubes"
	if fileType == "view" {
		subPath = "model/views"
	}

	if _, err = a.Store.Collimato.UpdateWorkspaceFile(model.CollimatoWorkspaceFile{
		WorkspaceID: workspaceID,
		Name:        name,
		FileType:    fileType,
		Content:     content,
	}); err != nil {
		tlog.Errorw("Failed to save cube file",
			"workspace_id", workspaceID,
			"file", name,
			"error", err,
		)
		return nil, model.NewAppError("datamodel.file_write_failed", http.StatusInternalServerError)
	}

	if err = a.UpdateWorkspaceRolePermissions(*workspace); err != nil {
		tlog.Errorw("Failed to update workspace role permissions",
			"workspace_id", workspaceID,
			"user_id", user.ID,
			"error", err,
		)
		return nil, model.NewAppError("collimato.roles_update_failed", http.StatusInternalServerError)
	}

	return &model.CubeFile{
		Name:    name,
		Content: content,
		Path:    path.Join(subPath, name),
		Type:    fileType,
	}, nil
}

// isValidIdentifier reports whether s is a safe Cube member/cube name
// (^[A-Za-z_][A-Za-z0-9_]*$). Used to reject malformed builder input.
func isValidIdentifier(s string) bool {
	if s == "" {
		return false
	}

	for i, r := range s {
		switch {
		case r == '_' || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z'):
		case i > 0 && r >= '0' && r <= '9':
		default:
			return false
		}
	}

	return true
}

// Cube's YAML data model spells measure types in snake_case. The visual builder
// sends the JavaScript spelling for count_distinct, so both are accepted and
// rendered canonically. An unknown type breaks schema compilation for the whole
// workspace, not just this cube, so it is rejected before it can be written.
var cubeDimensionTypes = map[string]bool{
	"string":  true,
	"number":  true,
	"time":    true,
	"boolean": true,
}

var cubeMeasureTypes = map[string]string{
	"count":                 "count",
	"count_distinct":        "count_distinct",
	"countDistinct":         "count_distinct",
	"count_distinct_approx": "count_distinct_approx",
	"sum":                   "sum",
	"avg":                   "avg",
	"min":                   "min",
	"max":                   "max",
	"number":                "number",
}

// nameIsTaken reports whether the workspace already has a cube or view file
// under this name. Cube shares one namespace across cubes and views, so a new
// model must not collide with either. overwrite is set when the builder is
// editing the file of that exact name and type, which is the one case where
// replacing it is intended.
func (a *App) checkDataModelWrite(user model.User, workspaceID, name, fileType string, overwrite bool) *model.AppError {
	permission := model.CollimatoPermissions.PermissionCreateDataModels

	if overwrite {
		existing, err := a.Store.Collimato.GetWorkspaceFile(workspaceID, name+".yml", fileType)
		if err != nil {
			tlog.Errorw("Failed to check data model",
				"workspace_id", workspaceID,
				"name", name,
				"error", err,
			)
			return model.NewAppError("datamodel.file_write_failed", http.StatusInternalServerError)
		}

		if existing != nil {
			permission = model.CollimatoPermissions.PermissionEditDataModels
		}
	}

	if !a.CollimatoWorkspaceHasPermission(user, workspaceID, permission) {
		return model.NewAppError("datamodel.forbidden", http.StatusForbidden)
	}

	return nil
}

func (a *App) nameIsTaken(workspaceID, name, fileType string, overwrite bool) (bool, error) {
	for _, t := range []string{model.CollimatoFileTypeCube, model.CollimatoFileTypeView} {
		for _, ext := range []string{".yml", ".js"} {
			existing, err := a.Store.Collimato.GetWorkspaceFile(workspaceID, name+ext, t)
			if err != nil {
				return false, err
			}

			if existing == nil {
				continue
			}

			if overwrite && t == fileType && ext == ".yml" {
				continue
			}

			return true, nil
		}
	}

	return false, nil
}

// BuildCubeFile renders a structured Dataset to a Cube.js YAML file and persists
// both the rendered YAML (what Cube consumes) and the structured model (so the
// visual builder can re-open it without parsing). fileType is "cube" or "view".
func (a *App) BuildCubeFile(user model.User, workspaceID, fileType string, m model.Dataset, overwrite bool) (*model.CubeFile, *model.AppError) {
	if appErr := a.checkDataModelWrite(user, workspaceID, m.Name, fileType, overwrite); appErr != nil {
		return nil, appErr
	}

	if fileType != model.CollimatoFileTypeCube && fileType != model.CollimatoFileTypeView {
		return nil, model.NewAppError("datamodel.invalid_model", http.StatusBadRequest)
	}

	if !isValidIdentifier(m.Name) || m.SqlTable == "" || len(m.Dimensions) == 0 {
		return nil, model.NewAppError("datamodel.invalid_model", http.StatusBadRequest)
	}

	for _, d := range m.Dimensions {
		if !isValidIdentifier(d.Name) || d.SQL == "" || !cubeDimensionTypes[d.Type] {
			return nil, model.NewAppError("datamodel.invalid_model", http.StatusBadRequest)
		}
	}

	for _, ms := range m.Measures {
		if !isValidIdentifier(ms.Name) || cubeMeasureTypes[ms.Type] == "" {
			return nil, model.NewAppError("datamodel.invalid_model", http.StatusBadRequest)
		}
	}

	for _, j := range m.Joins {
		validRel := j.Relationship == "many_to_one" || j.Relationship == "one_to_many" || j.Relationship == "one_to_one"
		hasCondition := j.SQL != "" || (isValidIdentifier(j.ThisColumn) && isValidIdentifier(j.OtherColumn))
		if !isValidIdentifier(j.Name) || !validRel || !hasCondition {
			return nil, model.NewAppError("datamodel.invalid_model", http.StatusBadRequest)
		}
	}

	taken, err := a.nameIsTaken(workspaceID, m.Name, fileType, overwrite)
	if err != nil {
		tlog.Errorw("Failed to check data model name",
			"workspace_id", workspaceID,
			"name", m.Name,
			"error", err,
		)
		return nil, model.NewAppError("datamodel.file_write_failed", http.StatusInternalServerError)
	}

	if taken {
		return nil, model.NewAppError("datamodel.name_taken", http.StatusConflict)
	}

	workspace, err := a.Store.Collimato.GetWorkspaceByID(workspaceID)
	if err != nil {
		tlog.Errorw("Failed to retrieve workspace",
			"workspace_id", workspaceID,
			"error", err,
		)
		return nil, model.NewAppError("collimato.retrieval_failed", http.StatusInternalServerError)
	}

	conn, err := a.Store.Collimato.GetConnectionByID(m.ConnectionID)
	if err != nil {
		tlog.Errorw("Failed to retrieve connection",
			"workspace_id", workspaceID,
			"connection_id", m.ConnectionID,
			"error", err,
		)
		return nil, model.NewAppError("connection.retrieval_failed", http.StatusInternalServerError)
	}

	if conn.WorkspaceID != workspaceID {
		return nil, model.NewAppError("connection.forbidden", http.StatusForbidden)
	}

	content, err := renderCubeYAML(m, conn)
	if err != nil {
		tlog.Errorw("Failed to render cube YAML",
			"workspace_id", workspaceID,
			"error", err,
		)
		return nil, model.NewAppError("datamodel.file_write_failed", http.StatusInternalServerError)
	}

	modelJSON, err := json.Marshal(m)
	if err != nil {
		tlog.Errorw("Failed to marshal builder model",
			"workspace_id", workspaceID,
			"error", err,
		)
		return nil, model.NewAppError("datamodel.file_write_failed", http.StatusInternalServerError)
	}

	modelStr := string(modelJSON)

	subPath := "model/cubes"
	if fileType == model.CollimatoFileTypeView {
		subPath = "model/views"
	}

	name := m.Name + ".yml"

	if _, err = a.Store.Collimato.UpdateWorkspaceFile(model.CollimatoWorkspaceFile{
		WorkspaceID:  workspaceID,
		Name:         name,
		FileType:     fileType,
		Content:      content,
		BuilderModel: &modelStr,
	}); err != nil {
		tlog.Errorw("Failed to save cube file",
			"workspace_id", workspaceID,
			"file", name,
			"error", err,
		)
		return nil, model.NewAppError("datamodel.file_write_failed", http.StatusInternalServerError)
	}

	if err = a.UpdateWorkspaceRolePermissions(*workspace); err != nil {
		tlog.Errorw("Failed to update workspace role permissions",
			"workspace_id", workspaceID,
			"user_id", user.ID,
			"error", err,
		)
		return nil, model.NewAppError("collimato.roles_update_failed", http.StatusInternalServerError)
	}

	return &model.CubeFile{
		Name:         name,
		Content:      content,
		Path:         path.Join(subPath, name),
		Type:         fileType,
		BuilderModel: &modelStr,
	}, nil
}

func (a *App) BuildViewFile(user model.User, workspaceID string, m model.View, overwrite bool) (*model.CubeFile, *model.AppError) {
	if appErr := a.checkDataModelWrite(user, workspaceID, m.Name, model.CollimatoFileTypeView, overwrite); appErr != nil {
		return nil, appErr
	}

	if !isValidIdentifier(m.Name) || len(m.Cubes) == 0 {
		return nil, model.NewAppError("datamodel.invalid_model", http.StatusBadRequest)
	}

	for _, c := range m.Cubes {
		if c.JoinPath == "" || len(c.Includes) == 0 {
			return nil, model.NewAppError("datamodel.invalid_model", http.StatusBadRequest)
		}

		for _, seg := range strings.Split(c.JoinPath, ".") {
			if !isValidIdentifier(seg) {
				return nil, model.NewAppError("datamodel.invalid_model", http.StatusBadRequest)
			}
		}

		for _, inc := range c.Includes {
			if inc != "*" && !isValidIdentifier(inc) {
				return nil, model.NewAppError("datamodel.invalid_model", http.StatusBadRequest)
			}
		}
	}

	taken, err := a.nameIsTaken(workspaceID, m.Name, model.CollimatoFileTypeView, overwrite)
	if err != nil {
		tlog.Errorw("Failed to check data model name",
			"workspace_id", workspaceID,
			"name", m.Name,
			"error", err,
		)
		return nil, model.NewAppError("datamodel.file_write_failed", http.StatusInternalServerError)
	}

	if taken {
		return nil, model.NewAppError("datamodel.name_taken", http.StatusConflict)
	}

	workspace, err := a.Store.Collimato.GetWorkspaceByID(workspaceID)
	if err != nil {
		tlog.Errorw("Failed to retrieve workspace",
			"workspace_id", workspaceID,
			"error", err,
		)
		return nil, model.NewAppError("collimato.retrieval_failed", http.StatusInternalServerError)
	}

	content, err := renderViewYAML(m)
	if err != nil {
		tlog.Errorw("Failed to render view YAML",
			"workspace_id", workspaceID,
			"error", err,
		)
		return nil, model.NewAppError("datamodel.file_write_failed", http.StatusInternalServerError)
	}

	modelJSON, err := json.Marshal(m)
	if err != nil {
		tlog.Errorw("Failed to marshal builder model",
			"workspace_id", workspaceID,
			"error", err,
		)
		return nil, model.NewAppError("datamodel.file_write_failed", http.StatusInternalServerError)
	}

	modelStr := string(modelJSON)

	name := m.Name + ".yml"
	if _, err = a.Store.Collimato.UpdateWorkspaceFile(model.CollimatoWorkspaceFile{
		WorkspaceID:  workspaceID,
		Name:         name,
		FileType:     model.CollimatoFileTypeView,
		Content:      content,
		BuilderModel: &modelStr,
	}); err != nil {
		tlog.Errorw("Failed to save view file",
			"workspace_id", workspaceID,
			"file", name,
			"error", err,
		)
		return nil, model.NewAppError("datamodel.file_write_failed", http.StatusInternalServerError)
	}

	if err = a.UpdateWorkspaceRolePermissions(*workspace); err != nil {
		tlog.Errorw("Failed to update workspace role permissions",
			"workspace_id", workspaceID,
			"user_id", user.ID,
			"error", err,
		)
		return nil, model.NewAppError("collimato.roles_update_failed", http.StatusInternalServerError)
	}

	return &model.CubeFile{
		Name:         name,
		Content:      content,
		Path:         path.Join("model/views", name),
		Type:         model.CollimatoFileTypeView,
		BuilderModel: &modelStr,
	}, nil
}

// cubeDependents returns the file names of builder-made cubes and views that
// reference cubeName (a cube join target, or a view cube-ref join path), so a
// delete that would break Cube compilation can be refused.
func (a *App) cubeDependents(workspaceID, cubeName string) ([]string, error) {
	var deps []string

	cubes, err := a.Store.Collimato.GetWorkspaceFiles(workspaceID, model.CollimatoFileTypeCube)
	if err != nil {
		return nil, err
	}

	for _, f := range cubes {
		if f.BuilderModel == nil || strings.TrimSuffix(f.Name, path.Ext(f.Name)) == cubeName {
			continue
		}

		var cm model.Dataset
		if json.Unmarshal([]byte(*f.BuilderModel), &cm) != nil {
			continue
		}

		for _, j := range cm.Joins {
			if j.Name == cubeName {
				deps = append(deps, f.Name)
				break
			}
		}
	}

	views, err := a.Store.Collimato.GetWorkspaceFiles(workspaceID, model.CollimatoFileTypeView)
	if err != nil {
		return nil, err
	}

	for _, f := range views {
		if f.BuilderModel == nil {
			continue
		}

		var vm model.View
		if json.Unmarshal([]byte(*f.BuilderModel), &vm) != nil {
			continue
		}

		for _, c := range vm.Cubes {
			if slices.Contains(strings.Split(c.JoinPath, "."), cubeName) {
				deps = append(deps, f.Name)
				break
			}
		}
	}

	return deps, nil
}

func (a *App) DeleteCubeFile(user model.User, name, filetype, workspaceID string) *model.AppError {
	if !a.CollimatoWorkspaceHasPermission(user, workspaceID, model.CollimatoPermissions.PermissionDeleteDataModels) {
		return model.NewAppError("datamodel.forbidden", http.StatusForbidden)
	}

	workspace, err := a.Store.Collimato.GetWorkspaceByID(workspaceID)
	if err != nil {
		tlog.Errorw("Failed to retrieve workspace",
			"workspace_id", workspaceID,
			"error", err,
		)
		return model.NewAppError("collimato.retrieval_failed", http.StatusInternalServerError)
	}

	// Deleting a cube that other cubes (joins) or views (cube refs) still depend
	// on would break Cube's schema compilation for the whole workspace, so block
	// it until the dependents are updated or removed.
	if filetype == model.CollimatoFileTypeCube {
		deps, derr := a.cubeDependents(workspaceID, strings.TrimSuffix(name, path.Ext(name)))
		if derr != nil {
			tlog.Errorw("Failed to resolve cube dependents",
				"workspace_id", workspaceID,
				"cube", name,
				"error", derr,
			)
			return model.NewAppError("collimato.retrieval_failed", http.StatusInternalServerError)
		}

		if len(deps) > 0 {
			return model.NewAppError("datamodel.delete_has_dependents", http.StatusConflict)
		}
	}

	if err = a.Store.Collimato.DeleteWorkspaceFile(workspaceID, name, filetype); err != nil {
		tlog.Errorw("Failed to delete cube file",
			"workspace_id", workspaceID,
			"file", name,
			"error", err,
		)
		return model.NewAppError("datamodel.file_delete_failed", http.StatusInternalServerError)
	}

	if err = a.UpdateWorkspaceRolePermissions(*workspace); err != nil {
		tlog.Errorw("Failed to update workspace role permissions",
			"workspace_id", workspaceID,
			"user_id", user.ID,
			"error", err,
		)
		return model.NewAppError("collimato.roles_update_failed", http.StatusInternalServerError)
	}

	return nil
}

// cubeDataSource normalizes a connection to the data_source string that
// GetCubeConnection resolves against: "default" for the default connection,
// otherwise the lowercased, space-stripped database name.
func cubeDataSource(conn *model.Connection) string {
	if conn.Default {
		return "default"
	}

	return strings.ToLower(strings.ReplaceAll(conn.Database, " ", ""))
}

// yamlCube* are render-only structs whose field order controls the emitted YAML
// shape (the marshaler preserves struct field order).
type yamlCubeFile struct {
	Cubes []yamlCube `yaml:"cubes"`
}

type yamlCube struct {
	Name       string          `yaml:"name"`
	SqlTable   string          `yaml:"sql_table"`
	DataSource string          `yaml:"data_source"`
	Joins      []yamlJoin      `yaml:"joins,omitempty"`
	Dimensions []yamlDimension `yaml:"dimensions"`
	Measures   []yamlMeasure   `yaml:"measures"`
}

type yamlJoin struct {
	Name         string `yaml:"name"`
	Relationship string `yaml:"relationship"`
	SQL          string `yaml:"sql"`
}

type yamlDimension struct {
	Name       string `yaml:"name"`
	SQL        string `yaml:"sql"`
	Type       string `yaml:"type"`
	PrimaryKey bool   `yaml:"primary_key,omitempty"`
	Title      string `yaml:"title,omitempty"`
}

type yamlMeasure struct {
	Name   string `yaml:"name"`
	Type   string `yaml:"type"`
	SQL    string `yaml:"sql,omitempty"`
	Title  string `yaml:"title,omitempty"`
	Format string `yaml:"format,omitempty"`
}

type yamlViewFile struct {
	Views []yamlView `yaml:"views"`
}

type yamlView struct {
	Name  string         `yaml:"name"`
	Cubes []yamlViewCube `yaml:"cubes"`
}

type yamlViewCube struct {
	JoinPath string `yaml:"join_path"`
	// Includes is either the "*" splat string or a []string member list, so the
	// emitted YAML matches Cube's two accepted shapes.
	Includes interface{} `yaml:"includes"`
	Prefix   bool        `yaml:"prefix,omitempty"`
}

// renderViewYAML renders a View to a Cube.js view YAML file.
func renderViewYAML(m model.View) (string, error) {
	view := yamlView{Name: m.Name}
	for _, c := range m.Cubes {
		var inc interface{} = c.Includes
		if len(c.Includes) == 1 && c.Includes[0] == "*" {
			inc = "*"
		}

		view.Cubes = append(view.Cubes, yamlViewCube{
			JoinPath: c.JoinPath,
			Includes: inc,
			Prefix:   c.Prefix,
		})
	}

	var b strings.Builder
	enc := yaml.NewEncoder(&b)
	enc.SetIndent(2)
	if err := enc.Encode(yamlViewFile{Views: []yamlView{view}}); err != nil {
		return "", err
	}

	if err := enc.Close(); err != nil {
		return "", err
	}

	return b.String(), nil
}

// joinSQL builds a Cube join condition. A raw SQL override wins; otherwise it is
// composed from the structured columns as "{CUBE}.<this> = {<target>}.<other>".
func joinSQL(j model.CubeJoin) string {
	if j.SQL != "" {
		return cleanSQL(j.SQL)
	}

	return fmt.Sprintf("{CUBE}.%s = {%s}.%s", j.ThisColumn, j.Name, j.OtherColumn)
}

// cleanSQL strips trailing whitespace per line so the marshaler renders multi-line
// SQL (e.g. a CASE expression) as a clean literal block scalar rather than falling
// back to quoted style.
func cleanSQL(sql string) string {
	if !strings.Contains(sql, "\n") {
		return sql
	}

	lines := strings.Split(sql, "\n")
	for i, l := range lines {
		lines[i] = strings.TrimRight(l, " \t")
	}

	return strings.Join(lines, "\n")
}

// renderCubeYAML renders a Dataset to a Cube.js YAML data-model file using
// go.yaml.in/yaml/v3 (the maintained successor to gopkg.in/yaml.v3).
func renderCubeYAML(m model.Dataset, conn *model.Connection) (string, error) {
	cube := yamlCube{
		Name:       m.Name,
		SqlTable:   m.SqlTable,
		DataSource: cubeDataSource(conn),
	}

	for _, j := range m.Joins {
		cube.Joins = append(cube.Joins, yamlJoin{
			Name:         j.Name,
			Relationship: j.Relationship,
			SQL:          joinSQL(j),
		})
	}

	for _, d := range m.Dimensions {
		cube.Dimensions = append(cube.Dimensions, yamlDimension{
			Name:       d.Name,
			SQL:        cleanSQL(d.SQL),
			Type:       d.Type,
			PrimaryKey: d.PrimaryKey,
			Title:      d.Title,
		})
	}

	for _, ms := range m.Measures {
		cube.Measures = append(cube.Measures, yamlMeasure{
			Name:   ms.Name,
			Type:   cubeMeasureTypes[ms.Type],
			SQL:    cleanSQL(ms.SQL),
			Title:  ms.Title,
			Format: ms.Format,
		})
	}

	var b strings.Builder
	enc := yaml.NewEncoder(&b)
	enc.SetIndent(2)
	if err := enc.Encode(yamlCubeFile{Cubes: []yamlCube{cube}}); err != nil {
		return "", err
	}

	if err := enc.Close(); err != nil {
		return "", err
	}

	return b.String(), nil
}

func (a *App) GenerateCubeFiles(user model.User, m map[string]interface{}, workspaceID string) (*model.GenerateModelsResult, *model.AppError) {
	if !a.CollimatoWorkspaceHasPermission(user, workspaceID, model.CollimatoPermissions.PermissionCreateDataModels) {
		return nil, model.NewAppError("datamodel.forbidden", http.StatusForbidden)
	}

	workspace, err := a.Store.Collimato.GetWorkspaceByID(workspaceID)
	if err != nil {
		tlog.Errorw("Failed to retrieve workspace",
			"workspace_id", workspaceID,
			"error", err,
		)
		return nil, model.NewAppError("collimato.retrieval_failed", http.StatusInternalServerError)
	}

	result := &model.GenerateModelsResult{
		Created:        make([]string, 0),
		Skipped:        make([]string, 0),
		SkippedColumns: make(map[string][]string),
	}

	for connID, v := range m {
		conn, err := a.Store.Collimato.GetConnectionByID(connID)
		if err != nil {
			tlog.Errorw("Failed to retrieve connection",
				"workspace_id", workspaceID,
				"connection_id", connID,
				"error", err,
			)
			return nil, model.NewAppError("connection.retrieval_failed", http.StatusInternalServerError)
		}

		if conn.WorkspaceID != workspaceID {
			return nil, model.NewAppError("connection.forbidden", http.StatusForbidden)
		}

		tables, ok := v.(map[string]interface{})
		if !ok {
			return nil, model.NewAppError("datamodel.invalid_model", http.StatusBadRequest)
		}

		for table, cols := range tables {
			columns, ok := columnDetails(cols)
			if !ok || !isValidIdentifier(table) {
				result.Skipped = append(result.Skipped, table)
				continue
			}

			taken, err := a.nameIsTaken(workspaceID, table, model.CollimatoFileTypeCube, false)
			if err != nil {
				tlog.Errorw("Failed to check data model name",
					"workspace_id", workspaceID,
					"name", table,
					"error", err,
				)
				return nil, model.NewAppError("datamodel.generate_failed", http.StatusInternalServerError)
			}

			if taken {
				result.Skipped = append(result.Skipped, table)
				continue
			}

			dataset, skippedColumns := datasetFromTable(conn, table, columns)
			if len(dataset.Dimensions) == 0 {
				result.Skipped = append(result.Skipped, table)
				continue
			}

			if len(skippedColumns) > 0 {
				result.SkippedColumns[table] = skippedColumns
			}

			content, err := renderCubeYAML(dataset, conn)
			if err != nil {
				tlog.Errorw("Failed to render cube YAML",
					"workspace_id", workspaceID,
					"name", dataset.Name,
					"error", err,
				)
				return nil, model.NewAppError("datamodel.generate_failed", http.StatusInternalServerError)
			}

			modelJSON, err := json.Marshal(dataset)
			if err != nil {
				tlog.Errorw("Failed to marshal builder model",
					"workspace_id", workspaceID,
					"name", dataset.Name,
					"error", err,
				)
				return nil, model.NewAppError("datamodel.generate_failed", http.StatusInternalServerError)
			}

			modelStr := string(modelJSON)

			if _, err = a.Store.Collimato.UpdateWorkspaceFile(model.CollimatoWorkspaceFile{
				WorkspaceID:  workspaceID,
				Name:         dataset.Name + ".yml",
				FileType:     model.CollimatoFileTypeCube,
				Content:      content,
				BuilderModel: &modelStr,
			}); err != nil {
				tlog.Errorw("Failed to save generated cube file",
					"workspace_id", workspaceID,
					"file", dataset.Name,
					"error", err,
				)
				return nil, model.NewAppError("datamodel.generate_failed", http.StatusInternalServerError)
			}

			result.Created = append(result.Created, dataset.Name)
		}
	}

	roles, err := a.Store.Collimato.GetWorkspaceRoles(workspace.ID)
	if err != nil {
		tlog.Errorw("Failed to retrieve workspace roles",
			"workspace_id", workspaceID,
			"error", err,
		)
		return nil, model.NewAppError("role.retrieval_failed", http.StatusInternalServerError)
	}

	for _, role := range roles {
		if !role.AutoUpdate {
			continue
		}

		role.TablePermissions = append(role.TablePermissions, result.Created...)

		if err = a.Store.Collimato.UpdateWorkspaceRole(workspaceID, role); err != nil {
			tlog.Errorw("Failed to update workspace role",
				"workspace_id", workspaceID,
				"role_id", role.ID,
				"error", err,
			)
			return nil, model.NewAppError("role.update_failed", http.StatusInternalServerError)
		}
	}

	if workspace.Status == model.CollimatoWorkspaceStatusFinished {
		if err = a.UpdateWorkspaceRolePermissions(*workspace); err != nil {
			tlog.Errorw("Failed to update workspace role permissions",
				"workspace_id", workspaceID,
				"user_id", user.ID,
				"error", err,
			)
			return nil, model.NewAppError("collimato.roles_update_failed", http.StatusInternalServerError)
		}
	}

	return result, nil
}

func columnDetails(cols interface{}) ([]model.ColumnDetail, bool) {
	raw, ok := cols.([]interface{})
	if !ok {
		return nil, false
	}

	columns := make([]model.ColumnDetail, 0, len(raw))
	for _, c := range raw {
		entry, ok := c.(map[string]interface{})
		if !ok {
			return nil, false
		}

		name, nameOK := entry["name"].(string)
		columnType, typeOK := entry["type"].(string)
		if !nameOK || !typeOK {
			return nil, false
		}

		columns = append(columns, model.ColumnDetail{Name: name, Type: columnType})
	}

	return columns, true
}

func datasetFromTable(conn *model.Connection, table string, columns []model.ColumnDetail) (model.Dataset, []string) {
	dataset := model.Dataset{
		Name:         table,
		SqlTable:     table,
		ConnectionID: conn.ID,
		Dimensions:   make([]model.CubeDimension, 0, len(columns)),
		Measures:     []model.CubeMeasure{{Name: "count", Type: "count"}},
	}

	skipped := make([]string, 0)

	for _, c := range columns {
		t := columnType(c.Type)
		if !cubeDimensionTypes[t] || !isValidIdentifier(c.Name) {
			skipped = append(skipped, c.Name)
			continue
		}

		dataset.Dimensions = append(dataset.Dimensions, model.CubeDimension{
			Name: c.Name,
			SQL:  c.Name,
			Type: t,
		})
	}

	return dataset, skipped
}

func columnType(columnType string) string {
	lowerType := strings.ToLower(columnType)
	switch {
	case strings.HasPrefix(lowerType, "tinyint(1)"):
		return "boolean"

	case strings.HasPrefix(lowerType, "varchar"),
		strings.HasPrefix(lowerType, "text"),
		strings.HasPrefix(lowerType, "char"),
		strings.HasPrefix(lowerType, "tinytext"),
		strings.HasPrefix(lowerType, "mediumtext"),
		strings.HasPrefix(lowerType, "longtext"):
		return "string"

	case strings.HasPrefix(lowerType, "int"),
		strings.HasPrefix(lowerType, "tinyint"),
		strings.HasPrefix(lowerType, "smallint"),
		strings.HasPrefix(lowerType, "mediumint"),
		strings.HasPrefix(lowerType, "bigint"),
		strings.HasPrefix(lowerType, "float"),
		strings.HasPrefix(lowerType, "double"),
		strings.HasPrefix(lowerType, "numeric"),
		strings.HasPrefix(lowerType, "decimal"):
		return "number"

	case strings.HasPrefix(lowerType, "timestamp"),
		strings.HasPrefix(lowerType, "datetime"),
		strings.HasPrefix(lowerType, "date"),
		strings.HasPrefix(lowerType, "time"):
		return "time"

	case strings.HasPrefix(lowerType, "bool"),
		strings.HasPrefix(lowerType, "boolean"):
		return "boolean"

	default:
		return "unknown"
	}
}

func (a *App) GetCubeNames(workspace model.CollimatoWorkspace) ([]string, error) {
	collection, err := a.QueryEngine.Meta(context.Background(), workspace.ID)
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(collection.Cubes))
	for _, cube := range collection.Cubes {
		names = append(names, cube.Name)
	}

	return names, nil
}

// GenerateToken mints the short-lived JWT Twigex sends to the shared Cube.js
// server. It is signed with the single shared secret (CubeAPISecret) and carries
// the workspace id, which Cube.js exposes as securityContext.workspaceId to its
// contextToAppId, driverFactory, and repositoryFactory. Authorization and field
// filtering are already enforced by Twigex before this token is issued.
//
// It also carries connVersion, the latest connection change time for the
// workspace. Cube.js folds this into contextToOrchestratorId so that editing a
// connection produces a new orchestrator (and thus re-runs driverFactory with the
// updated credentials) instead of reusing the cached connection pool.
func (a *App) GenerateToken(workspaceID string) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"workspaceId": workspaceID,
		"connVersion": a.connectionVersion(workspaceID),
		"exp":         time.Now().Add(time.Hour * 2).Unix(),
	})

	tokenString, err := token.SignedString([]byte(*a.ConfigStore.Config.CollimatoSettings.CubeAPISecret))
	if err != nil {
		return ""
	}

	return tokenString
}

// connectionVersion returns the latest change time across a workspace's database
// connections. It is best-effort: on error it returns 0 (Cube simply keeps its
// current connection pool).
func (a *App) connectionVersion(workspaceID string) int64 {
	conns, err := a.Store.Collimato.GetConnections(workspaceID)
	if err != nil {
		tlog.Warnw("Failed to compute connection version",
			"workspace_id", workspaceID,
			"error", err,
		)
		return 0
	}

	var version int64
	for _, c := range conns {
		if c.UpdatedAt > version {
			version = c.UpdatedAt
		}

		if c.DeletedAt > version {
			version = c.DeletedAt
		}
	}

	return version
}
