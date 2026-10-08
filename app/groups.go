// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strings"

	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/tlog"
)

func (a *App) CreateGroup(ctx context.Context, user model.User, req model.CreateGroupRequest) (*model.Group, *model.AppError) {
	if user.Role != model.SystemAdminRoleId {
		return nil, model.NewAppError("group.forbidden", http.StatusForbidden)
	}

	if !a.Server.License.HasGroups() {
		return nil, model.NewAppError("groups.license_required", http.StatusPaymentRequired)
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, model.NewAppError("group.name_required", http.StatusBadRequest)
	}

	roles, appErr := a.validateGroupRoles(req.Roles)
	if appErr != nil {
		return nil, appErr
	}

	group, err := a.Store.Groups.Create(ctx, user.ID, name, strings.TrimSpace(req.Description), roles)
	if err != nil {
		tlog.Errorw("Failed to create group",
			"owner_id", user.ID,
			"error", err,
		)
		return nil, model.NewAppError("group.create_failed", http.StatusInternalServerError)
	}

	if len(req.MemberIDs) > 0 {
		if err := a.Store.Groups.AddMembers(ctx, group.ID, req.MemberIDs, model.GroupRoleMember); err != nil {
			tlog.Errorw("Failed to add initial members to group",
				"group_id", group.ID,
				"error", err,
			)
			return nil, model.NewAppError("group.add_members_failed", http.StatusInternalServerError)
		}

		count, err := a.Store.Groups.CountMembers(ctx, group.ID, "")
		if err != nil {
			tlog.Errorw("Failed to count group members",
				"group_id", group.ID,
				"error", err,
			)
			return nil, model.NewAppError("group.retrieval_failed", http.StatusInternalServerError)
		}

		group.MemberCount = count
	}

	return group, nil
}

func (a *App) GetGroup(ctx context.Context, user model.User, groupID string) (*model.Group, *model.AppError) {
	group, err := a.Store.Groups.Get(ctx, groupID)
	if err != nil {
		tlog.Errorw("Failed to retrieve group",
			"group_id", groupID,
			"error", err,
		)
		return nil, model.NewAppError("group.retrieval_failed", http.StatusInternalServerError)
	}

	if group == nil {
		return nil, model.NewAppError("group.not_found", http.StatusNotFound)
	}

	if user.Role != model.SystemAdminRoleId && group.OwnerID != user.ID {
		isMember, appErr := a.userIsGroupMember(ctx, user.ID, groupID)
		if appErr != nil {
			return nil, appErr
		}

		if !isMember {
			return nil, model.NewAppError("group.forbidden", http.StatusForbidden)
		}
	}

	return group, nil
}

const (
	groupMembersDefaultLimit = 30
	groupMembersMaxLimit     = 100

	groupsPageDefaultLimit = 20
	groupsPageMaxLimit     = 100

	groupSearchDefaultLimit = 20
	groupSearchMaxLimit     = 50
)

// Viewing group membership is intentionally open to any authenticated user (no
// permission gate); the member pickers are only opened where the group is
// already visible.
func (a *App) GetGroupMembers(ctx context.Context, groupID, query string, sort model.Sort, limit, offset int) ([]model.GroupMember, int, *model.AppError) {
	group, err := a.Store.Groups.Get(ctx, groupID)
	if err != nil {
		tlog.Errorw("Failed to retrieve group", "group_id", groupID, "error", err)
		return nil, 0, model.NewAppError("group.retrieval_failed", http.StatusInternalServerError)
	}

	if group == nil {
		return nil, 0, model.NewAppError("group.not_found", http.StatusNotFound)
	}

	if limit <= 0 || limit > groupMembersMaxLimit {
		limit = groupMembersDefaultLimit
	}

	if offset < 0 {
		offset = 0
	}

	query = strings.TrimSpace(query)

	members, err := a.Store.Groups.GetMembersPaged(ctx, groupID, query, sort, limit, offset)
	if err != nil {
		tlog.Errorw("Failed to list group members", "group_id", groupID, "error", err)
		return nil, 0, model.NewAppError("group.members_failed", http.StatusInternalServerError)
	}

	total, err := a.Store.Groups.CountMembers(ctx, groupID, query)
	if err != nil {
		tlog.Errorw("Failed to count group members", "group_id", groupID, "error", err)
		return nil, 0, model.NewAppError("group.members_failed", http.StatusInternalServerError)
	}

	return members, total, nil
}

func (a *App) ListGroupsPaged(ctx context.Context, user model.User, query string, sort model.Sort, limit, offset int) ([]model.Group, int, *model.AppError) {
	if user.Role != model.SystemAdminRoleId {
		return nil, 0, model.NewAppError("group.forbidden", http.StatusForbidden)
	}

	if limit <= 0 || limit > groupsPageMaxLimit {
		limit = groupsPageDefaultLimit
	}

	if offset < 0 {
		offset = 0
	}

	query = strings.TrimSpace(query)

	groups, err := a.Store.Groups.GetAllPaged(ctx, query, sort, limit, offset)
	if err != nil {
		tlog.Errorw("Failed to list groups",
			"error", err,
		)
		return nil, 0, model.NewAppError("group.list_failed", http.StatusInternalServerError)
	}

	total, err := a.Store.Groups.CountAll(ctx, query)
	if err != nil {
		tlog.Errorw("Failed to count groups",
			"error", err,
		)
		return nil, 0, model.NewAppError("group.list_failed", http.StatusInternalServerError)
	}

	return groups, total, nil
}

func (a *App) SearchGroups(ctx context.Context, user model.User, query string, limit int) ([]model.Group, *model.AppError) {
	if limit <= 0 || limit > groupSearchMaxLimit {
		limit = groupSearchDefaultLimit
	}

	query = strings.TrimSpace(query)

	var groups []model.Group
	var err error

	if user.Role == model.SystemAdminRoleId {
		groups, err = a.Store.Groups.GetAllPaged(ctx, query, model.Sort{}, limit, 0)
	} else {
		groups, err = a.Store.Groups.SearchForUser(ctx, user.ID, query, limit)
	}

	if err != nil {
		tlog.Errorw("Failed to search groups",
			"user_id", user.ID,
			"error", err,
		)
		return nil, model.NewAppError("group.list_failed", http.StatusInternalServerError)
	}

	return groups, nil
}

func (a *App) UpdateGroup(ctx context.Context, user model.User, groupID string, req model.UpdateGroupRequest) (*model.Group, *model.AppError) {
	if user.Role != model.SystemAdminRoleId {
		return nil, model.NewAppError("group.forbidden", http.StatusForbidden)
	}

	if !a.Server.License.HasGroups() {
		return nil, model.NewAppError("groups.license_required", http.StatusPaymentRequired)
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, model.NewAppError("group.name_required", http.StatusBadRequest)
	}

	roles, appErr := a.validateGroupRoles(req.Roles)
	if appErr != nil {
		return nil, appErr
	}

	if err := a.Store.Groups.Update(ctx, groupID, name, strings.TrimSpace(req.Description), roles); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.NewAppError("group.not_found", http.StatusNotFound)
		}

		tlog.Errorw("Failed to update group",
			"group_id", groupID,
			"error", err,
		)
		return nil, model.NewAppError("group.update_failed", http.StatusInternalServerError)
	}

	group, err := a.Store.Groups.Get(ctx, groupID)
	if err != nil {
		tlog.Errorw("Failed to retrieve group after update",
			"group_id", groupID,
			"error", err,
		)
		return nil, model.NewAppError("group.retrieval_failed", http.StatusInternalServerError)
	}

	if group == nil {
		return nil, model.NewAppError("group.not_found", http.StatusNotFound)
	}

	return group, nil
}

func (a *App) DeleteGroup(ctx context.Context, user model.User, groupID string) *model.AppError {
	if user.Role != model.SystemAdminRoleId {
		return model.NewAppError("group.forbidden", http.StatusForbidden)
	}

	// Fetch members before Groups.SoftDelete, which deletes group_members in the same transaction.
	members, _ := a.Store.Groups.GetMembers(ctx, groupID)

	if err := a.Store.Channels.RemoveGroupFromAllChannels(ctx, groupID); err != nil {
		tlog.Errorw("Failed to detach group from channels",
			"group_id", groupID,
			"error", err,
		)
		return model.NewAppError("group.delete_failed", http.StatusInternalServerError)
	}

	if err := a.Store.Groups.SoftDelete(ctx, groupID); err != nil {
		tlog.Errorw("Failed to delete group",
			"group_id", groupID,
			"error", err,
		)
		return model.NewAppError("group.delete_failed", http.StatusInternalServerError)
	}

	// Best-effort: clear assignee fields using pre-fetched members
	if len(members) > 0 {
		go a.clearGroupMembersFromWorkspaces(groupID, members)
	}

	return nil
}

func (a *App) AddGroupMembers(ctx context.Context, user model.User, groupID string, req model.AddGroupMembersRequest) *model.AppError {
	if user.Role != model.SystemAdminRoleId {
		return model.NewAppError("group.forbidden", http.StatusForbidden)
	}

	if !a.Server.License.HasGroups() {
		return model.NewAppError("groups.license_required", http.StatusPaymentRequired)
	}

	if len(req.UserIDs) == 0 {
		return model.NewAppError("group.no_members", http.StatusBadRequest)
	}

	group, err := a.Store.Groups.Get(ctx, groupID)
	if err != nil {
		tlog.Errorw("Failed to retrieve group",
			"group_id", groupID,
			"error", err,
		)
		return model.NewAppError("group.retrieval_failed", http.StatusInternalServerError)
	}

	if group == nil {
		return model.NewAppError("group.not_found", http.StatusNotFound)
	}

	if err := a.Store.Groups.AddMembers(ctx, groupID, req.UserIDs, model.GroupRoleMember); err != nil {
		tlog.Errorw("Failed to add group members",
			"group_id", groupID,
			"error", err,
		)
		return model.NewAppError("group.add_members_failed", http.StatusInternalServerError)
	}
	// One bulk INSERT across every channel this group is attached to, so there
	// is no per-user partial-success window.
	if err := a.Store.Channels.MaterializeUserInGroupChannels(ctx, req.UserIDs, groupID); err != nil {
		tlog.Errorw("Failed to materialize new group members into channels",
			"group_id", groupID,
			"error", err,
		)
		return model.NewAppError("group.add_members_failed", http.StatusInternalServerError)
	}

	return nil
}

func (a *App) RemoveGroupMember(ctx context.Context, user model.User, groupID, userID string) *model.AppError {
	if user.Role != model.SystemAdminRoleId {
		return model.NewAppError("group.forbidden", http.StatusForbidden)
	}

	// Remove the group_members row first so the cascade's "any other path"
	// check naturally excludes the leaving group.
	if err := a.Store.Groups.RemoveMember(ctx, groupID, userID); err != nil {
		tlog.Errorw("Failed to remove group member",
			"group_id", groupID,
			"user_id", userID,
			"error", err,
		)
		return model.NewAppError("group.remove_member_failed", http.StatusInternalServerError)
	}

	if err := a.Store.Channels.DematerializeUserFromGroupChannels(ctx, userID, groupID); err != nil {
		tlog.Errorw("Failed to dematerialize ex-group-member from channels",
			"group_id", groupID,
			"user_id", userID,
			"error", err,
		)
		return model.NewAppError("group.remove_member_failed", http.StatusInternalServerError)
	}

	// Best-effort: clear task assignments in any workspace where this user's
	// only access was via this group (now removed).
	go a.clearRemovedGroupMemberAssignees(groupID, userID)
	return nil
}

func (a *App) clearRemovedGroupMemberAssignees(groupID, userID string) {
	wsIDs, err := a.Store.Workspace.GetIDsForGroup(groupID)
	if err != nil || len(wsIDs) == 0 {
		return
	}

	for _, wsID := range wsIDs {
		direct, _ := a.Store.Workspace.GetMemberByUserID(wsID, userID)
		other, _ := a.Store.Workspace.UserHasOtherGroupAccess(wsID, groupID, userID)
		if direct != nil || other {
			continue
		}

		tables, err := a.Store.Workspace.GetAllTablesBasic(wsID)
		if err != nil {
			continue
		}

		for _, rt := range tables {
			a.clearAllPersonFields(rt, userID)
		}
	}
}

// Groups must never grant system_admin, which would be a full permission
// bypass, so that role is rejected; remaining names are normalized and verified
// against real system roles.
func (a *App) validateGroupRoles(roles []string) ([]string, *model.AppError) {
	seen := make(map[string]struct{}, len(roles))
	cleaned := make([]string, 0, len(roles))
	for _, r := range roles {
		r = strings.TrimSpace(r)
		if r == "" {
			continue
		}

		if r == model.SystemAdminRoleId {
			return nil, model.NewAppError("group.role_admin_forbidden", http.StatusBadRequest)
		}

		if _, ok := seen[r]; ok {
			continue
		}

		seen[r] = struct{}{}
		cleaned = append(cleaned, r)
	}

	if len(cleaned) == 0 {
		return cleaned, nil
	}

	found, err := a.Store.Roles.GetByNames(cleaned)
	if err != nil {
		tlog.Errorw("Failed to validate group roles", "error", err)
		return nil, model.NewAppError("group.role_validation_failed", http.StatusInternalServerError)
	}

	if found == nil || len(found) != len(cleaned) {
		return nil, model.NewAppError("group.invalid_role", http.StatusBadRequest)
	}

	return cleaned, nil
}

// Folds group-granted roles into the in-memory user.Role once, at user load, so
// every downstream SessionHasPermission stays a cheap in-memory check over the
// space-separated string and the hot permission path is untouched. Best-effort:
// a resolution failure is logged and leaves the user with only their direct
// roles rather than failing the request. system_admin already bypasses every
// check, so nothing is added for it.
func (a *App) applyGroupRoles(ctx context.Context, user *model.User) {
	if user == nil || user.Role == model.SystemAdminRoleId {
		return
	}

	groupRoles, err := a.Store.Groups.GetRolesForUser(ctx, user.ID)
	if err != nil {
		tlog.Errorw("Failed to resolve group roles for user",
			"user_id", user.ID,
			"error", err,
		)
		return
	}

	if len(groupRoles) == 0 {
		return
	}

	roles := strings.Fields(user.Role)
	seen := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		seen[r] = struct{}{}
	}

	for _, r := range groupRoles {
		if _, ok := seen[r]; ok {
			continue
		}

		seen[r] = struct{}{}
		roles = append(roles, r)
	}

	user.Role = strings.Join(roles, " ")
}

func (a *App) userIsGroupMember(ctx context.Context, userID, groupID string) (bool, *model.AppError) {
	ids, err := a.Store.Groups.GetIDsForUser(ctx, userID)
	if err != nil {
		tlog.Errorw("Failed to retrieve user group IDs",
			"user_id", userID,
			"error", err,
		)
		return false, model.NewAppError("group.retrieval_failed", http.StatusInternalServerError)
	}

	for _, id := range ids {
		if id == groupID {
			return true, nil
		}
	}

	return false, nil
}
