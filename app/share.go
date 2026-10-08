// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"context"
	"net/http"
	"strconv"

	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/tlog"
)

func (a *App) Share(user model.User, req model.FileShare) (*model.ShareResponse, *model.AppError) {
	if !a.SessionHasPermission(user, model.FilePermissions.PermissionShareFiles) {
		return nil, model.NewAppError("permission.forbidden", http.StatusForbidden)
	}

	f, appErr := a.HasPermission(req.FileID, user)
	if appErr != nil {
		return nil, appErr
	}

	if f.Owner != user.ID && !f.AccessLevel.CanShare() {
		return nil, model.NewAppError("file.share_forbidden", http.StatusForbidden)
	}

	users := make([]string, 0)
	sharedUsers := make([]model.SharedUsers, 0)

	for _, u := range req.Users {
		targetUser, err := a.Store.User.Get(u)
		if err != nil {
			tlog.Errorw("Failed to retrieve target user for share",
				"file_id", req.FileID,
				"user_id", u,
				"error", err,
			)
			return nil, model.NewAppError("user.not_found", http.StatusInternalServerError)
		}

		if targetUser == nil {
			return nil, model.NewAppError("user.not_found", http.StatusNotFound)
		}

		if targetUser.DeactivatedAt != 0 {
			return nil, model.NewAppError("file.share_user_deactivated", http.StatusForbidden)
		}

		// Access only ever rises down the tree, so refuse a grant below what the
		// user already inherits from an ancestor (it would be a no-op).
		inherited, err := a.Store.File.InheritedAccessLevel(context.Background(), f.ID, model.SHARE_TYPE_USER, targetUser.ID)
		if err != nil {
			tlog.Errorw("Failed to resolve inherited access level",
				"file_id", f.ID,
				"user_id", targetUser.ID,
				"error", err,
			)
			return nil, model.NewAppError("file.share_failed", http.StatusInternalServerError)
		}

		if req.AccessLevel < inherited {
			return nil, model.NewAppError("file.share_below_inherited", http.StatusBadRequest)
		}

		// Only the shared node gets a share row; descendants inherit access
		// when it is resolved.
		if _, err = a.Store.File.CreateShare(*f, model.SHARE_TYPE_USER, user.ID, targetUser.ID, req.Expiration, req.AccessLevel); err != nil {
			tlog.Errorw("Failed to create share",
				"file_id", f.ID,
				"user_id", targetUser.ID,
				"error", err,
			)
			return nil, model.NewAppError("file.share_failed", http.StatusInternalServerError)
		}

		sharedUsers = append(sharedUsers, model.SharedUsers{
			ID:          targetUser.ID,
			Owner:       f.IsOwner(targetUser.ID),
			Name:        targetUser.Name,
			LastName:    targetUser.LastName,
			Photo:       targetUser.Photo,
			Email:       targetUser.Email,
			Expiration:  req.Expiration,
			AccessLevel: req.AccessLevel,
		})
		users = append(users, targetUser.ID)
	}

	sharedGroups, appErr := a.shareWithGroups(context.Background(), user, f, req)
	if appErr != nil {
		return nil, appErr
	}

	groupIDs := make([]string, 0, len(sharedGroups))
	for _, g := range sharedGroups {
		groupIDs = append(groupIDs, g.ID)
	}

	if len(sharedUsers) > 0 || len(sharedGroups) > 0 {
		if err := a.Store.File.Update(*f); err != nil {
			tlog.Errorw("Failed to update file after share",
				"file_id", f.ID,
				"error", err,
			)
			return nil, model.NewAppError("file.share_failed", http.StatusInternalServerError)
		}

		// Notify only the newly granted users, not everyone with access.
		recipientIDs := append([]string{}, users...)
		for _, gid := range groupIDs {
			members, err := a.Store.Groups.GetMembers(context.Background(), gid)
			if err != nil {
				tlog.Errorw("Failed to list group members for share notification",
					"file_id", f.ID,
					"group_id", gid,
					"error", err,
				)
				continue
			}

			for _, mbr := range members {
				recipientIDs = append(recipientIDs, mbr.UserID)
			}
		}

		a.CreateFileShareNotification(*f, user, recipientIDs)
	}

	a.RecordActivity(user.ID, model.AppFiles, model.ActivityFileShare, f.Parent, f.ID, map[string]any{
		"name":   f.DisplayName,
		"type":   f.Type,
		"users":  users,
		"groups": groupIDs,
		"size":   strconv.FormatInt(f.Size, 10),
	})

	return &model.ShareResponse{SharedUsers: sharedUsers, SharedGroups: sharedGroups}, nil
}

// Notifications are deliberately not fanned out here. The caller (Share) fires
// one CreateFileNotification after all shares are created, covering both direct
// user shares and resolved group members in a single pass.
func (a *App) shareWithGroups(ctx context.Context, initiator model.User, f *model.File, req model.FileShare) ([]model.SharedGroup, *model.AppError) {
	if len(req.Groups) == 0 {
		return nil, nil
	}

	if !a.Server.License.HasGroups() {
		return nil, model.NewAppError("groups.license_required", http.StatusPaymentRequired)
	}

	sharedGroups := make([]model.SharedGroup, 0, len(req.Groups))
	for _, groupID := range req.Groups {
		group, err := a.Store.Groups.Get(ctx, groupID)
		if err != nil {
			tlog.Errorw("Failed to retrieve group for share",
				"file_id", req.FileID,
				"group_id", groupID,
				"error", err,
			)
			return nil, model.NewAppError("file.share_failed", http.StatusInternalServerError)
		}

		if group == nil {
			return nil, model.NewAppError("group.not_found", http.StatusNotFound)
		}

		// Access only ever rises down the tree, so refuse a grant below what the
		// group already inherits from an ancestor (it would be a no-op).
		inherited, err := a.Store.File.InheritedAccessLevel(ctx, f.ID, model.SHARE_TYPE_GROUP, group.ID)
		if err != nil {
			tlog.Errorw("Failed to resolve inherited group access level",
				"file_id", f.ID,
				"group_id", group.ID,
				"error", err,
			)
			return nil, model.NewAppError("file.share_failed", http.StatusInternalServerError)
		}

		if req.AccessLevel < inherited {
			return nil, model.NewAppError("file.share_below_inherited", http.StatusBadRequest)
		}

		// Only the shared node gets a share row; descendants inherit access
		// when it is resolved.
		if _, err = a.Store.File.CreateShare(*f, model.SHARE_TYPE_GROUP, initiator.ID, group.ID, req.Expiration, req.AccessLevel); err != nil {
			tlog.Errorw("Failed to create group share",
				"file_id", f.ID,
				"group_id", group.ID,
				"error", err,
			)
			return nil, model.NewAppError("file.share_failed", http.StatusInternalServerError)
		}

		sharedGroups = append(sharedGroups, model.SharedGroup{
			ID:          group.ID,
			Name:        group.Name,
			Description: group.Description,
			MemberCount: group.MemberCount,
			Expiration:  req.Expiration,
			AccessLevel: req.AccessLevel,
		})
	}

	return sharedGroups, nil
}

// UnshareGroup removes a group share from a file (and any descendants if the
// file is a folder).
func (a *App) UnshareGroup(ctx context.Context, user model.User, fileID, groupID string) *model.AppError {
	if !a.SessionHasPermission(user, model.FilePermissions.PermissionShareFiles) {
		return model.NewAppError("permission.forbidden", http.StatusForbidden)
	}

	file, appErr := a.HasPermission(fileID, user)
	if appErr != nil {
		return appErr
	}

	// Removing a group affects every member, so it is owner or manager only.
	if file.Owner != user.ID && !file.AccessLevel.CanShare() {
		return model.NewAppError("file.forbidden", http.StatusForbidden)
	}

	// Escalate to the ancestor that grants the group, same as user unshare.
	root, err := a.Store.File.GetGrantingRoot(ctx, fileID, model.SHARE_TYPE_GROUP, groupID)
	if err != nil {
		tlog.Errorw("Failed to resolve granting root",
			"file_id", fileID,
			"group_id", groupID,
			"error", err,
		)
		return model.NewAppError("file.unshare_failed", http.StatusInternalServerError)
	}

	if root == "" {
		return nil
	}

	if err := a.Store.File.RemoveGroupFromShare(ctx, root, groupID); err != nil {
		tlog.Errorw("Failed to remove group from shared file",
			"file_id", root,
			"group_id", groupID,
			"error", err,
		)
		return model.NewAppError("file.unshare_failed", http.StatusInternalServerError)
	}

	return nil
}

func (a *App) Unshare(user model.User, fileID string, ID string) *model.AppError {
	if !a.SessionHasPermission(user, model.FilePermissions.PermissionShareFiles) {
		return model.NewAppError("permission.forbidden", http.StatusForbidden)
	}

	file, appErr := a.HasPermission(fileID, user)
	if appErr != nil {
		return appErr
	}

	// Anyone may remove themselves (leave); removing others is owner or manager.
	if ID != user.ID && file.Owner != user.ID && !file.AccessLevel.CanShare() {
		return model.NewAppError("file.forbidden", http.StatusForbidden)
	}

	ctx, cancel := a.dbCtx(context.Background())
	defer cancel()

	// An inherited file has no share row, so a revoke escalates to the granting
	// ancestor and unshares that whole subtree.
	root, err := a.Store.File.GetGrantingRoot(ctx, fileID, model.SHARE_TYPE_USER, ID)
	if err != nil {
		tlog.Errorw("Failed to resolve granting root",
			"file_id", fileID,
			"user_id", ID,
			"error", err,
		)
		return model.NewAppError("file.unshare_failed", http.StatusInternalServerError)
	}

	if root == "" {
		return nil
	}

	if err := a.Store.File.RemoveUserFromShare(root, ID); err != nil {
		tlog.Errorw("Failed to remove user from shared file",
			"file_id", root,
			"user_id", ID,
			"error", err,
		)
		return model.NewAppError("file.unshare_failed", http.StatusInternalServerError)
	}

	return nil
}

func (a *App) UpdateSharePermissions(ctx context.Context, user model.User, s model.FileSharePatch) *model.AppError {
	if !a.SessionHasPermission(user, model.FilePermissions.PermissionShareFiles) {
		return model.NewAppError("permission.forbidden", http.StatusForbidden)
	}

	if s.ShareType == 0 {
		s.ShareType = model.SHARE_TYPE_USER
	}

	file, appErr := a.HasPermission(s.FileID, user)
	if appErr != nil {
		return appErr
	}

	if file.Owner != user.ID && !file.AccessLevel.CanShare() {
		return model.NewAppError("file.forbidden", http.StatusForbidden)
	}

	// Editing an inherited file escalates to the ancestor that grants the principal.
	root, err := a.Store.File.GetGrantingRoot(ctx, s.FileID, s.ShareType, s.ID)
	if err != nil {
		tlog.Errorw("Failed to resolve granting root",
			"file_id", s.FileID,
			"user_id", s.ID,
			"error", err,
		)
		return model.NewAppError("file.permissions_update_failed", http.StatusInternalServerError)
	}

	if root == "" {
		return nil
	}

	s.FileID = root

	if err := a.Store.File.UpdateUserPermission(s); err != nil {
		tlog.Errorw("Failed to update user share permissions",
			"file_id", s.FileID,
			"user_id", s.ID,
			"error", err,
		)
		return model.NewAppError("file.permissions_update_failed", http.StatusInternalServerError)
	}

	return nil
}
