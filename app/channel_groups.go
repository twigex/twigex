// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"context"
	"net/http"

	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/tlog"
)

func (a *App) ListChannelGroups(ctx context.Context, user model.User, channelID string) ([]model.ChannelGroup, *model.AppError) {
	if appErr := a.RequireChannelPermission(user, channelID); appErr != nil {
		return nil, appErr
	}

	groups, err := a.Store.Channels.GetGroups(ctx, channelID)
	if err != nil {
		tlog.Errorw("Failed to list channel groups",
			"channel_id", channelID,
			"error", err,
		)
		return nil, model.NewAppError("channel_group.list_failed", http.StatusInternalServerError)
	}

	return groups, nil
}

func (a *App) AddGroupsToChannel(ctx context.Context, user model.User, channelID string, req model.AddChannelGroupsRequest) ([]model.ChannelGroup, *model.AppError) {
	if appErr := a.RequireChannelPermission(user, channelID); appErr != nil {
		return nil, appErr
	}

	if !a.Server.License.HasGroups() {
		return nil, model.NewAppError("groups.license_required", http.StatusPaymentRequired)
	}

	if len(req.GroupIDs) == 0 {
		return nil, model.NewAppError("channel_group.no_groups", http.StatusBadRequest)
	}

	if _, err := a.Store.Channels.Get(channelID); err != nil {
		tlog.Errorw("Channel lookup failed",
			"channel_id", channelID,
			"error", err,
		)
		return nil, model.NewAppError("channel.not_found", http.StatusNotFound)
	}

	groupIDs := DedupeIDs(req.GroupIDs)

	groups, err := a.Store.Groups.GetByIDs(ctx, groupIDs)
	if err != nil {
		tlog.Errorw("Failed to retrieve groups",
			"channel_id", channelID,
			"error", err,
		)
		return nil, model.NewAppError("group.retrieval_failed", http.StatusInternalServerError)
	}

	if len(groups) != len(groupIDs) {
		return nil, model.NewAppError("group.not_found", http.StatusNotFound)
	}

	if err := a.Store.Channels.AddGroups(ctx, channelID, groupIDs, user.ID); err != nil {
		tlog.Errorw("Failed to add groups to channel",
			"channel_id", channelID,
			"error", err,
		)
		return nil, model.NewAppError("channel_group.add_failed", http.StatusInternalServerError)
	}

	if err := a.Store.Channels.MaterializeGroups(ctx, channelID, groupIDs); err != nil {
		tlog.Errorw("Failed to materialize group members into channel",
			"channel_id", channelID,
			"error", err,
		)
		return nil, model.NewAppError("channel_group.add_failed", http.StatusInternalServerError)
	}

	all, err := a.Store.Channels.GetGroups(ctx, channelID)
	if err != nil {
		tlog.Errorw("Failed to list channel groups after add",
			"channel_id", channelID,
			"error", err,
		)
		return nil, model.NewAppError("channel_group.add_failed", http.StatusInternalServerError)
	}

	requested := make(map[string]struct{}, len(groupIDs))
	for _, gid := range groupIDs {
		requested[gid] = struct{}{}
	}

	added := make([]model.ChannelGroup, 0, len(groupIDs))
	for _, cg := range all {
		if _, ok := requested[cg.GroupID]; ok {
			added = append(added, cg)
		}
	}

	return added, nil
}

func DedupeIDs(ids []string) []string {
	seen := make(map[string]struct{}, len(ids))
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; ok {
			continue
		}

		seen[id] = struct{}{}
		out = append(out, id)
	}

	return out
}

func (a *App) RemoveGroupFromChannel(ctx context.Context, user model.User, channelID, groupID string) *model.AppError {
	if appErr := a.RequireChannelPermission(user, channelID); appErr != nil {
		return appErr
	}

	// Delete the channel_groups row first so the cascade's "any other path"
	// check naturally excludes the group being removed.
	if err := a.Store.Channels.RemoveGroup(ctx, channelID, groupID); err != nil {
		tlog.Errorw("Failed to remove group from channel",
			"channel_id", channelID,
			"group_id", groupID,
			"error", err,
		)
		return model.NewAppError("channel_group.remove_failed", http.StatusInternalServerError)
	}

	if err := a.Store.Channels.DematerializeGroup(ctx, channelID, groupID); err != nil {
		tlog.Errorw("Failed to dematerialize group members from channel",
			"channel_id", channelID,
			"group_id", groupID,
			"error", err,
		)
		return model.NewAppError("channel_group.remove_failed", http.StatusInternalServerError)
	}

	return nil
}
