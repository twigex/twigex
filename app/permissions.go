// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"context"
	"net/http"
	"time"

	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/tlog"
)

func (a *App) HasPermission(fileID string, user model.User) (*model.File, *model.AppError) {
	file, err := a.Store.File.Get(fileID)
	if err != nil {
		tlog.Errorw("Failed to retrieve file",
			"file_id", fileID,
			"error", err,
		)
		return nil, model.NewAppError("file.retrieval_failed", http.StatusInternalServerError)
	}

	if file == nil {
		return nil, model.NewAppError("file.not_found", http.StatusNotFound)
	}

	if file.Type == "cloud#drive" {
		return file, nil
	}

	if file.Owner == user.ID {
		file.AccessLevel = model.AccessManager
		return file, nil
	}

	// No request context to inherit cancellation from, so the ancestor up-walk
	// is capped by the configured query timeout instead.
	ctx, cancel := a.dbCtx(context.Background())
	defer cancel()

	permissions, err := a.Store.File.ResolveEffectiveShare(ctx, user.ID, fileID)
	if err != nil {
		tlog.Errorw("Failed to resolve effective share",
			"file_id", fileID,
			"user_id", user.ID,
			"error", err,
		)
		return nil, model.NewAppError("file.retrieval_failed", http.StatusInternalServerError)
	}

	if len(permissions) == 0 {
		return nil, model.NewAppError("file.forbidden", http.StatusForbidden)
	}

	for _, v := range permissions {
		if v.Expiration > file.Expiration {
			file.Expiration = v.Expiration
		}

		if v.AccessLevel > file.AccessLevel {
			file.AccessLevel = v.AccessLevel
		}
	}

	if file.Expiration != 0 && time.Now().Unix() > file.Expiration {
		return nil, model.NewAppError("file.forbidden", http.StatusForbidden)
	}

	return file, nil
}

func (a *App) CollimatoHasPermissionToFields(user model.User, workspaceID string, query *model.DataQuery) (bool, *model.AppError) {
	roleNames, err := a.Store.Collimato.GetEffectiveRolesForUser(context.Background(), workspaceID, user.ID)
	if err != nil {
		tlog.Errorw("Failed to resolve effective workspace roles",
			"user_id", user.ID,
			"workspace_id", workspaceID,
			"error", err,
		)
		return false, model.NewAppError("collimato.workspace_user_failed", http.StatusInternalServerError)
	}

	if len(roleNames) == 0 {
		return false, model.NewAppError("collimato.forbidden", http.StatusForbidden)
	}

	workspaceRoles, err := a.Store.Collimato.GetRolesByName(roleNames, workspaceID)
	if err != nil {
		tlog.Errorw("Failed to retrieve workspace roles",
			"user_id", user.ID,
			"workspace_id", workspaceID,
			"error", err,
		)
		return false, model.NewAppError("collimato.roles_failed", http.StatusInternalServerError)
	}

	if a.CollimatoRoles == nil {
		if rolesCarryDataSecurity(workspaceRoles) {
			return false, model.NewAppError("collimato.data_security_unavailable", http.StatusPaymentRequired)
		}

		return true, nil
	}

	ok, err := a.CollimatoRoles.HasPermissionToFields(workspaceRoles, query)
	if err != nil {
		return false, model.NewAppError("collimato.permission_denied", http.StatusForbidden)
	}

	return ok, nil
}

// rolesCarryDataSecurity reports whether any role holds row or column rules.
// Without the enterprise implementation registered those rules cannot be
// applied, so the request must be refused rather than served unrestricted.
func rolesCarryDataSecurity(roles []model.CollimatoRole) bool {
	for _, r := range roles {
		if len(r.RowPermissions) > 0 || len(r.ColumnPermissions) > 0 {
			return true
		}
	}

	return false
}

func (a *App) hasPermissionToChannel(user model.User, channelID string) (bool, *model.AppError) {
	channel, err := a.Store.Channels.Get(channelID)
	if err != nil {
		tlog.Errorw("Failed to retrieve channel",
			"channel_id", channelID,
			"error", err,
		)
		return false, model.NewAppError("channel.not_found", http.StatusInternalServerError)
	}

	if channel == nil {
		return false, model.NewAppError("channel.not_found", http.StatusNotFound)
	}

	if a.SessionHasPermission(user, model.AdminPermissions.PermissionManageChannels) {
		return true, nil
	}

	ok, err := a.Store.Channels.IsMember(channel.ID, user.ID)
	if err != nil {
		tlog.Errorw("Failed to check channel membership",
			"channel_id", channelID,
			"user_id", user.ID,
			"error", err,
		)
		return false, model.NewAppError("channel.not_found", http.StatusInternalServerError)
	}

	if !ok {
		return false, model.NewAppError("channel.forbidden", http.StatusForbidden)
	}

	for _, v := range channel.ChannelMembers {
		if v.UserID == user.ID && v.Role == model.ChannelRoleAdmin {
			return true, nil
		}
	}

	return false, model.NewAppError("channel.forbidden", http.StatusForbidden)
}

func (a *App) RequireChannelPermission(user model.User, channelID string) *model.AppError {
	ok, appErr := a.hasPermissionToChannel(user, channelID)
	if appErr != nil {
		return appErr
	}

	if !ok {
		return model.NewAppError("channel.forbidden", http.StatusForbidden)
	}

	return nil
}
