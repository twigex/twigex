// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"context"
	"net/http"

	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/tlog"
)

func (a *App) ListMeetingGroups(ctx context.Context, user model.User, channelID, meetingID string) ([]model.ChannelMeetingGroup, *model.AppError) {
	meeting, appErr := a.getMeetingForChannel(ctx, channelID, meetingID)
	if appErr != nil {
		return nil, appErr
	}

	if !a.canAccessMeeting(ctx, user, meeting) {
		return nil, model.NewAppError("channel.meeting_not_invited", http.StatusForbidden)
	}

	groups, err := a.Store.Channels.GetMeetingGroups(ctx, meetingID)
	if err != nil {
		tlog.Errorw("Failed to list meeting groups", "meeting_id", meetingID, "error", err)
		return nil, model.NewAppError("meeting_group.list_failed", http.StatusInternalServerError)
	}

	return groups, nil
}

func (a *App) AddGroupsToMeeting(ctx context.Context, user model.User, channelID, meetingID string, req model.AddMeetingGroupsRequest) ([]model.ChannelMeetingGroup, *model.AppError) {
	channel, err := a.Store.Channels.Get(channelID)
	if err != nil || channel == nil {
		return nil, model.NewAppError("channel.not_found", http.StatusNotFound)
	}

	if channel.Type == model.ChannelTypeDirect {
		return nil, model.NewAppError("channel.direct_invite_forbidden", http.StatusBadRequest)
	}

	meeting, appErr := a.getMeetingForChannel(ctx, channelID, meetingID)
	if appErr != nil {
		return nil, appErr
	}

	if appErr := a.requireMeetingHost(user, meeting); appErr != nil {
		return nil, appErr
	}

	if meeting.EndedAt != nil {
		return nil, model.NewAppError("channel.meeting_ended", http.StatusConflict)
	}

	if !a.Server.License.HasGroups() {
		return nil, model.NewAppError("groups.license_required", http.StatusPaymentRequired)
	}

	groupIDs := DedupeIDs(req.GroupIDs)
	if len(groupIDs) == 0 {
		return nil, model.NewAppError("meeting_group.no_groups", http.StatusBadRequest)
	}

	groups, err := a.Store.Groups.GetByIDs(ctx, groupIDs)
	if err != nil {
		tlog.Errorw("Failed to retrieve groups", "meeting_id", meetingID, "error", err)
		return nil, model.NewAppError("group.retrieval_failed", http.StatusInternalServerError)
	}

	if len(groups) != len(groupIDs) {
		return nil, model.NewAppError("group.not_found", http.StatusNotFound)
	}

	if err := a.Store.Channels.AddMeetingGroups(ctx, meetingID, groupIDs, user.ID); err != nil {
		tlog.Errorw("Failed to add groups to meeting", "meeting_id", meetingID, "error", err)
		return nil, model.NewAppError("meeting_group.add_failed", http.StatusInternalServerError)
	}

	all, err := a.Store.Channels.GetMeetingGroups(ctx, meetingID)
	if err != nil {
		tlog.Errorw("Failed to list meeting groups after add", "meeting_id", meetingID, "error", err)
		return nil, model.NewAppError("meeting_group.add_failed", http.StatusInternalServerError)
	}

	if err := a.syncMeetingInviteOnly(ctx, meeting); err != nil {
		tlog.Errorw("Failed to update meeting invite only", "meeting_id", meeting.ID, "error", err)
		return nil, model.NewAppError("meeting_group.add_failed", http.StatusInternalServerError)
	}

	go a.NotifyMeetingGroupsInvited(user, *channel, *meeting)
	go a.RefreshMeetingLists(*meeting, channelID, nil)

	return all, nil
}
