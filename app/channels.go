// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"image"
	"image/gif"
	"io"
	"net/http"
	"os"
	"path"
	"regexp"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/livekit/protocol/livekit"
	"github.com/twigex/twigex/crypto"
	"github.com/twigex/twigex/internal/imaging"
	"github.com/twigex/twigex/internal/klipy"
	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/safehttp"
	"github.com/twigex/twigex/tlog"
)

const GuestLinkValidity = 4 * time.Hour

var allowedGuestLinkValidities = map[int]time.Duration{
	1:   time.Hour,
	4:   4 * time.Hour,
	24:  24 * time.Hour,
	168: 7 * 24 * time.Hour,
}

func guestLinkValidity(expiresInHours int) time.Duration {
	if d, ok := allowedGuestLinkValidities[expiresInHours]; ok {
		return d
	}

	return GuestLinkValidity
}

func (a *App) GetChannelAndDMList(user model.User, urlParams string) ([]model.Channel, *model.AppError) {
	allChannels := make([]model.Channel, 0)

	if urlParams == "true" {
		openChannels, err := a.Store.Channels.GetAllByType(model.ChannelTypePublic)
		if err != nil {
			tlog.Errorw("Failed to retrieve channel list by type",
				"channel_type", model.ChannelTypePublic,
				"error", err)
			return nil, model.NewAppError("channel.retrieval_failed", http.StatusInternalServerError)
		}

		allChannels = append(allChannels, openChannels...)
	}

	channels, err := a.Store.Channels.GetAllForUser(user.ID)
	if err != nil {
		tlog.Errorw("Failed to retrieve user channel list", "user_id", user.ID, "error", err)
		return nil, model.NewAppError("channel.retrieval_failed", http.StatusBadRequest)
	}

	for _, v := range channels {
		if channelExistsInList(allChannels, v.ID) {
			continue // already exists in the list
		}

		allChannels = append(allChannels, v)
	}

	return allChannels, nil
}

func (a *App) GetChannelList(user model.User) ([]model.Channel, *model.AppError) {
	if !a.SessionHasPermission(user, model.AdminPermissions.PermissionManageChannels) {
		return nil, model.NewAppError("channel.forbidden", http.StatusForbidden)
	}

	channels, err := a.Store.Channels.GetAll()
	if err != nil {
		tlog.Errorw("Failed to retrieve channel list", "user_id", user.ID, "error", err)
		return nil, model.NewAppError("channel.retrieval_failed", http.StatusBadRequest)
	}

	return channels, nil
}

func (a *App) CreateChannel(user model.User, name, channelType, description, userID string) (*model.Channel, *model.AppError) {
	sanitized := strings.ToLower(strings.ReplaceAll(name, " ", ""))

	switch channelType {
	case "public":
		channelType = model.ChannelTypePublic
	case "private":
		channelType = model.ChannelTypePrivate
	case "direct":
		channelType = model.ChannelTypeDirect
	default:
		tlog.Warnw("Invalid channel type received", "channel_type", channelType)
		return nil, model.NewAppError("channel.invalid_type", http.StatusBadRequest)
	}

	t := time.Now().Unix()

	ch := model.Channel{
		ID:           model.NewID(),
		Type:         channelType,
		DisplayName:  name,
		Name:         sanitized, // need to sanitize this only alphabet and lower case
		Header:       "",
		Description:  description,
		LastPost:     0,
		MessageCount: 0,
		CreatedBy:    user.ID,
		CreatedAt:    t,
		UpdatedAt:    t,
		DeletedAt:    0,
	}

	if channelType == model.ChannelTypeDirect {
		// check if userID is valid
		dmUser, err := a.Store.User.Get(userID)
		if err != nil {
			tlog.Errorw("User lookup failed", "user_id", userID, "error", err)
			return nil, model.NewAppError("user.not_found", http.StatusNotFound)
		}

		if dmUser == nil {
			return nil, model.NewAppError("user.not_found", http.StatusNotFound)
		}

		chName := strings.Join([]string{user.ID, dmUser.ID}, "__")
		ch.Name = chName
		ch.DisplayName = ""

		// check if channel with two users
		channel, err := a.Store.Channels.GetDirectMessage(user.ID, userID)
		if err != nil {
			tlog.Errorw("Failed to look up direct channel", "error", err)
			return nil, model.NewAppError("channel.create_failed", http.StatusInternalServerError)
		}

		if channel != nil {
			return nil, model.NewAppError("channel.already_exists", http.StatusConflict)
		}
	}

	_, err := a.Store.Channels.Create(&ch)
	if err != nil {
		tlog.Errorw("Failed to create new channel", "error", err)
		return nil, model.NewAppError("channel.create_failed", http.StatusInternalServerError)
	}

	channelUsers := make([]model.User, 0)

	if channelType == model.ChannelTypeDirect {
		u, err := a.Store.User.Get(userID)
		if err != nil {
			tlog.Errorw("User lookup failed", "user_id", userID, "error", err)
			return nil, model.NewAppError("channel.create_failed", http.StatusInternalServerError)
		}

		channelUsers = append(channelUsers, user, *u)
	} else {
		channelUsers = append(channelUsers, user)
	}

	memberNow := time.Now().UnixMilli()
	channelMembers := make([]model.ChannelMember, 0)
	for i := range channelUsers {
		v := channelUsers[i]
		role := model.ChannelRoleMember

		if ch.CreatedBy == v.ID || channelType == model.ChannelTypeDirect {
			role = model.ChannelRoleAdmin
		}

		channelMembers = append(channelMembers, model.ChannelMember{
			ID:           model.NewID(),
			ChannelID:    ch.ID,
			UserID:       v.ID,
			Role:         role,
			NotifyProps:  model.ChannelNotifyProps{},
			MsgCount:     ch.MessageCount,
			MentionCount: 0,
			LastViewedAt: 0,
			UpdatedAt:    memberNow,
			DateJoined:   memberNow,
			IsDirect:     true,
			UserInfo:     &v,
		})
	}

	if err := a.Store.Channels.CreateOrPromoteMembers(context.Background(), channelMembers); err != nil {
		tlog.Errorw("Failed to insert channel members",
			"channel_id", ch.ID,
			"error", err,
		)
		return nil, model.NewAppError("channel.member_add_failed", http.StatusInternalServerError)
	}

	ch.ChannelMembers = channelMembers

	return &ch, nil
}

func (a *App) UpdateChannel(user model.User, channelID string, patch model.ChannelPatch) (*model.Channel, *model.AppError) {
	if appErr := a.RequireChannelPermission(user, channelID); appErr != nil {
		return nil, appErr
	}

	channel, appErr := a.writableChannel(channelID)
	if appErr != nil {
		return nil, appErr
	}

	if channel.Type == model.ChannelTypeDirect {
		return nil, model.NewAppError("channel.direct_update_forbidden", http.StatusBadRequest)
	}

	channel.Patch(patch)

	err := a.Store.Channels.Update(channel)
	if err != nil {
		tlog.Errorw("Channel update failed", "channel_id", channel.ID, "error", err)
		return nil, model.NewAppError("channel.update_failed", http.StatusInternalServerError)
	}

	// Send Channel update notification
	a.CreateChannelUpdateNotification(*channel, user, model.NOTIFICATION_CHANNEL_UPDATE)

	return channel, nil
}

func (a *App) AddUsersToChannelByID(user model.User, channelID string, userIDs []string) ([]model.ChannelMember, *model.AppError) {
	if !a.canAdministerChannel(user, channelID) {
		return nil, model.NewAppError("channel.forbidden", http.StatusForbidden)
	}

	channel, appErr := a.writableChannel(channelID)
	if appErr != nil {
		return nil, appErr
	}

	return a.AddUsersToChannel(userIDs, *channel)
}

// Adds the users as direct channel members in one statement that inserts new
// rows and updates existing ones together. Anyone already materialized via a
// group is updated in place to direct (not duplicated), and the whole batch
// applies together or not at all.
func (a *App) AddUsersToChannel(userIDs []string, channel model.Channel) ([]model.ChannelMember, *model.AppError) {
	existing, err := a.Store.User.GetByIDs(userIDs)
	if err != nil {
		tlog.Errorw("User lookup failed", "error", err)
		return nil, model.NewAppError("user.not_found", http.StatusNotFound)
	}

	now := time.Now().UnixMilli()
	members := make([]model.ChannelMember, 0, len(existing))
	for i := range existing {
		user := existing[i]
		members = append(members, model.ChannelMember{
			ID:           model.NewID(),
			ChannelID:    channel.ID,
			UserID:       user.ID,
			Role:         model.ChannelRoleMember,
			NotifyProps:  model.ChannelNotifyProps{},
			MsgCount:     channel.MessageCount,
			MentionCount: 0,
			LastViewedAt: 0,
			UpdatedAt:    now,
			DateJoined:   now,
			IsDirect:     true,
			UserInfo:     &user,
		})
	}

	if err := a.Store.Channels.CreateOrPromoteMembers(context.Background(), members); err != nil {
		tlog.Errorw("Failed to insert channel members",
			"channel_id", channel.ID,
			"error", err,
		)
		return nil, model.NewAppError("channel.member_add_failed", http.StatusInternalServerError)
	}

	return members, nil
}

func (a *App) JoinChannel(user model.User, channelID string) ([]model.ChannelMember, *model.AppError) {
	channel, appErr := a.writableChannel(channelID)
	if appErr != nil {
		return nil, appErr
	}

	ok, err := a.Store.Channels.IsMember(channel.ID, user.ID)
	if err != nil {
		tlog.Errorw("Failed to check channel membership",
			"channel_id", channel.ID,
			"user_id", user.ID,
			"error", err,
		)
		return nil, model.NewAppError("channel.not_member", http.StatusInternalServerError)
	}

	if ok {
		return nil, model.NewAppError("channel.already_member", http.StatusBadRequest)
	}

	if channel.Type == model.ChannelTypePublic {
		// Add user to public channel
		users, appErr := a.AddUsersToChannel([]string{user.ID}, *channel)
		if appErr != nil {
			tlog.Errorw("Failed to add user to channel", "channel_id", channel.ID, "error", err)
			return nil, model.NewAppError("channel.join_failed", http.StatusInternalServerError)
		}

		if len(users) > 0 {
			return users, nil
		}

		return nil, model.NewAppError("channel.join_failed", http.StatusInternalServerError)
	}

	return nil, model.NewAppError("channel.not_public", http.StatusForbidden)
}

func (a *App) RemoveUserFromChannel(user model.User, channelID, userID string) *model.AppError {
	if appErr := a.RequireChannelPermission(user, channelID); appErr != nil {
		return appErr
	}

	channel, appErr := a.writableChannel(channelID)
	if appErr != nil {
		return appErr
	}

	if !a.canAdministerChannel(user, channelID) {
		return model.NewAppError("channel.forbidden", http.StatusForbidden)
	}

	if channel.Type == model.ChannelTypeDirect {
		return model.NewAppError("channel.direct_remove_forbidden", http.StatusBadRequest)
	}

	err := a.Store.Channels.RemoveUser(channelID, userID)
	if err != nil {
		tlog.Errorw("Failed to remove user from channel", "channel_id", channel.ID, "user_id", userID, "error", err)
		return model.NewAppError("channel.member_remove_failed", http.StatusBadRequest)
	}

	return nil
}

func (a *App) LeaveFromChannel(user model.User, channelID string) *model.AppError {
	// Leaving requires only membership, not channel-admin.
	channel, appErr := a.requireChannelMember(user, channelID)
	if appErr != nil {
		return appErr
	}

	if channel.Type == model.ChannelTypeDirect {
		return model.NewAppError("channel.direct_remove_forbidden", http.StatusBadRequest)
	}

	if err := a.Store.Channels.RemoveUser(channelID, user.ID); err != nil {
		tlog.Errorw("Failed to leave from channel", "channel_id", channel.ID, "user_id", user.ID, "error", err)
		return model.NewAppError("channel.leave_failed", http.StatusBadRequest)
	}

	return nil
}

func (a *App) ChangeUserRole(user model.User, channelID, userID, role string) *model.AppError {
	if role != model.ChannelRoleAdmin && role != model.ChannelRoleMember {
		return model.NewAppError("channel.role_invalid", http.StatusBadRequest)
	}

	if !a.canAdministerChannel(user, channelID) {
		return model.NewAppError("channel.forbidden", http.StatusForbidden)
	}

	// A row materialized from a group goes when the group is detached, taking
	// the role with it.
	if role == model.ChannelRoleAdmin && !a.isDirectMember(channelID, userID) {
		return model.NewAppError("channel.role_requires_direct_member", http.StatusBadRequest)
	}

	err := a.Store.Channels.UpdateUserRole(channelID, userID, role)
	if err != nil {
		tlog.Errorw("Failed to update user role for channel",
			"channel_id", channelID,
			"user_id", userID,
			"error", err)
		return model.NewAppError("channel.role_change_failed", http.StatusBadRequest)
	}

	return nil
}

func (a *App) writableChannel(channelID string) (*model.Channel, *model.AppError) {
	channel, err := a.Store.Channels.Get(channelID)
	if err != nil {
		tlog.Errorw("Channel lookup failed", "channel_id", channelID, "error", err)
		return nil, model.NewAppError("channel.not_found", http.StatusNotFound)
	}

	if channel.DeletedAt != 0 {
		return nil, model.NewAppError("channel.archived", http.StatusConflict)
	}

	return channel, nil
}

func (a *App) ArchiveChannel(ctx context.Context, user model.User, channelID string) *model.AppError {
	channel, appErr := a.writableChannel(channelID)
	if appErr != nil {
		return appErr
	}

	if channel.Type != model.ChannelTypePublic {
		return model.NewAppError("channel.archive_type_invalid", http.StatusBadRequest)
	}

	if !a.IsChannelAdmin(channelID, user.ID) &&
		!a.SessionHasPermission(user, model.AdminPermissions.PermissionArchiveChannels) {
		return model.NewAppError("channel.forbidden", http.StatusForbidden)
	}

	dbCtx, cancel := a.dbCtx(ctx)
	defer cancel()

	if err := a.Store.Channels.UpdateArchived(dbCtx, channelID, time.Now().Unix()); err != nil {
		tlog.Errorw("Failed to archive channel", "channel_id", channelID, "error", err)
		return model.NewAppError("channel.archive_failed", http.StatusInternalServerError)
	}

	return nil
}

func (a *App) UnarchiveChannel(ctx context.Context, user model.User, channelID string) *model.AppError {
	channel, err := a.Store.Channels.Get(channelID)
	if err != nil {
		tlog.Errorw("Channel lookup failed", "channel_id", channelID, "error", err)
		return model.NewAppError("channel.not_found", http.StatusNotFound)
	}

	if channel.DeletedAt == 0 {
		return model.NewAppError("channel.not_archived", http.StatusConflict)
	}

	if !a.IsChannelAdmin(channelID, user.ID) &&
		!a.SessionHasPermission(user, model.AdminPermissions.PermissionArchiveChannels) {
		return model.NewAppError("channel.forbidden", http.StatusForbidden)
	}

	dbCtx, cancel := a.dbCtx(ctx)
	defer cancel()

	if err := a.Store.Channels.UpdateArchived(dbCtx, channelID, 0); err != nil {
		tlog.Errorw("Failed to unarchive channel", "channel_id", channelID, "error", err)
		return model.NewAppError("channel.unarchive_failed", http.StatusInternalServerError)
	}

	return nil
}

func (a *App) canAdministerChannel(user model.User, channelID string) bool {
	return a.IsChannelAdmin(channelID, user.ID) ||
		a.SessionHasPermission(user, model.AdminPermissions.PermissionManageChannels)
}

func (a *App) isDirectMember(channelID, userID string) bool {
	members, err := a.Store.Channels.GetMembers(channelID)
	if err != nil || members == nil {
		return false
	}

	for _, m := range members {
		if m.UserID == userID {
			return m.IsDirect
		}
	}

	return false
}

func (a *App) IsChannelAdmin(channelID, userID string) bool {
	members, err := a.Store.Channels.GetMembers(channelID)
	if err != nil || members == nil {
		return false
	}

	for _, m := range members {
		if m.UserID == userID && m.Role == model.ChannelRoleAdmin {
			return true
		}
	}

	return false
}

func (a *App) GetChannelPosts(ctx context.Context, user model.User, channelID, param, postID string) (*model.PostResponse, *model.AppError) {
	channel, err := a.Store.Channels.Get(channelID)
	if err != nil {
		tlog.Errorw("Channel lookup failed", "channel_id", channelID, "error", err)
		return nil, model.NewAppError("channel.not_found", http.StatusNotFound)
	}

	ok, err := a.Store.Channels.IsMember(channel.ID, user.ID)
	if err != nil {
		tlog.Errorw("Failed to check channel membership",
			"channel_id", channel.ID,
			"user_id", user.ID,
			"error", err,
		)
		return nil, model.NewAppError("channel.not_member", http.StatusInternalServerError)
	}

	if !ok {
		return nil, model.NewAppError("channel.not_member", http.StatusForbidden)
	}

	posts, err := a.Store.Posts.GetAllForChannel(ctx, channel.ID, postID, param)
	if err != nil {
		tlog.Errorw("Failed to retrieve channel posts", "channel_id", channelID, "error", err)
		// The posts read but their attachments or reactions did not. Serving
		// the messages without that metadata beats an empty channel.
		if posts == nil {
			return nil, model.NewAppError("post.retrieval_failed", http.StatusInternalServerError)
		}
	}

	// Parse posts to get url for OpenGraph and other metadata
	if *a.ConfigStore.Config.ChannelSettings.LinkPreviews {
		for i, post := range posts.Posts {
			urls := extractURLs(post.Message)
			if len(urls) > 0 {
				// posts
				linkHash := createLinkHash(urls[0])
				linkMetadata, err := a.Store.Posts.GetLinkMetadataByHash(linkHash)
				if err != nil {
					tlog.Errorw("Failed to get link metadata", "post_id", post.ID, "error", err)
				} else if linkMetadata != nil {
					hydrateLinkPreviewImage(linkMetadata)
					posts.Posts[i].Metadata.Links = append(posts.Posts[i].Metadata.Links, &linkMetadata.Data)
				}
			}
		}
	}

	a.attachResponseMentionUsers(posts)

	return posts, nil
}

func (a *App) UploadChatFile(channelID string, user model.User, file io.Reader, size int64) (fileID string, storageID string, appErr *model.AppError) {
	ok, err := a.Store.Channels.IsMember(channelID, user.ID)
	if err != nil {
		tlog.Errorw("Failed to check channel membership",
			"channel_id", channelID,
			"user_id", user.ID,
			"error", err,
		)
		return "", "", model.NewAppError("channel.not_found", http.StatusInternalServerError)
	}

	if !ok {
		return "", "", model.NewAppError("channel.forbidden", http.StatusForbidden)
	}

	id := model.NewID()

	primary, err := a.Store.Storage.GetPrimary()
	if err != nil {
		tlog.Errorw("Failed to get primary storage",
			"channel_id", channelID,
			"error", err,
		)
		return "", "", model.NewAppError("channel.file_upload_failed", http.StatusInternalServerError)
	}

	if primary == nil {
		// Legacy: no primary storage set, write to local temp path
		tempDir := path.Join("data/chat/tmp", channelID)
		if err = os.MkdirAll(tempDir, os.ModePerm); err != nil {
			tlog.Errorw("Failed to create temp directory",
				"channel_id", channelID,
				"error", err,
			)
			return "", "", model.NewAppError("channel.file_upload_failed", http.StatusInternalServerError)
		}

		tempFile, err := os.Create(path.Join(tempDir, id))
		if err != nil {
			tlog.Errorw("Failed to create temp file",
				"channel_id", channelID,
				"error", err,
			)
			return "", "", model.NewAppError("channel.file_upload_failed", http.StatusInternalServerError)
		}

		defer tempFile.Close()

		if _, err = io.Copy(tempFile, file); err != nil {
			tlog.Errorw("Failed to write uploaded file",
				"channel_id", channelID,
				"error", err,
			)
			return "", "", model.NewAppError("channel.file_upload_failed", http.StatusInternalServerError)
		}

		return id, "", nil
	}

	backend, exists := a.FileStorageObjects[primary.ID]
	if !exists {
		tlog.Errorw("Primary storage backend not available",
			"channel_id", channelID,
			"storage_id", primary.ID,
		)
		return "", "", model.NewAppError("storage.not_available", http.StatusInternalServerError)
	}

	tempPath, err := a.BuildTempFilePath(primary.ID, channelID+"/"+id, model.AppChat)
	if err != nil {
		tlog.Errorw("Failed to build temp file path",
			"channel_id", channelID,
			"error", err,
		)
		return "", "", model.NewAppError("channel.file_upload_failed", http.StatusInternalServerError)
	}

	if err := backend.WriteFile(context.Background(), tempPath, file, size); err != nil {
		tlog.Errorw("Failed to write chat file to storage",
			"channel_id", channelID,
			"storage_id", primary.ID,
			"error", err,
		)
		return "", "", model.NewAppError("channel.file_upload_failed", http.StatusInternalServerError)
	}

	return id, primary.ID, nil
}

func (a *App) GetChatFileAttachment(channelID string, fileID string, user model.User) (*model.PostFileAttachment, *model.AppError) {
	ok, err := a.Store.Channels.IsMember(channelID, user.ID)
	if err != nil {
		tlog.Errorw("Failed to check channel membership",
			"channel_id", channelID,
			"user_id", user.ID,
			"error", err,
		)
		return nil, model.NewAppError("channel.not_found", http.StatusInternalServerError)
	}

	if !ok {
		return nil, model.NewAppError("channel.forbidden", http.StatusForbidden)
	}

	attachment, err := a.Store.Posts.GetAttachmentByID(fileID)
	if err != nil {
		tlog.Errorw("Failed to retrieve post attachment",
			"channel_id", channelID,
			"file_id", fileID,
			"error", err,
		)
		return nil, model.NewAppError("channel.attachment_not_found", http.StatusInternalServerError)
	}

	if attachment == nil {
		return nil, model.NewAppError("channel.attachment_not_found", http.StatusNotFound)
	}

	if attachment.ChannelID != channelID {
		return nil, model.NewAppError("channel.attachment_forbidden", http.StatusForbidden)
	}

	return attachment, nil
}

func (a *App) handlePostFileAttachments(
	user model.User,
	channelID string,
	postID string,
	files []model.PostFileMetadata,
) ([]model.PostFileAttachment, error) {
	if len(files) == 0 {
		return nil, nil
	}

	attachments := make([]model.PostFileAttachment, 0, len(files))

	for _, file := range files {
		var width, height int64

		if file.StorageID != "" {
			backend, exists := a.FileStorageObjects[file.StorageID]
			if !exists {
				return nil, errors.New("storage backend not available: " + file.StorageID)
			}

			tempPath, err := a.BuildTempFilePath(file.StorageID, channelID+"/"+file.ID, model.AppChat)
			if err != nil {
				return nil, err
			}

			permanentPath, err := a.BuildFilePath(file.StorageID, channelID+"/"+file.ID, model.AppChat)
			if err != nil {
				return nil, err
			}

			if err := backend.MoveFile(context.Background(), tempPath, permanentPath); err != nil {
				return nil, err
			}

			if strings.HasPrefix(file.Type, "image/") {
				if reader, err := backend.ReadFile(context.Background(), permanentPath); err == nil {
					if config, _, err := image.DecodeConfig(reader); err == nil {
						width = int64(config.Width)
						height = int64(config.Height)
					}

					reader.Close()
				}
			}
		} else {
			// Legacy: no storage set, use local paths
			base := path.Join("data/chat")
			src := path.Join(base, "tmp", channelID, file.ID)
			dst := path.Join(base, channelID, file.ID)

			if err := os.MkdirAll(path.Join(base, channelID), 0o700); err != nil {
				return nil, err
			}

			if err := os.Rename(src, dst); err != nil {
				return nil, err
			}

			if f, err := os.Open(dst); err == nil {
				if config, _, err := image.DecodeConfig(f); err == nil {
					width = int64(config.Width)
					height = int64(config.Height)
				}

				f.Close()
			}
		}

		kind := "file"
		switch {
		case file.Type == "image/gif":
			kind = "gif"
		case strings.HasPrefix(file.Type, "image/"):
			kind = "image"
		case strings.HasPrefix(file.Type, "video/"):
			kind = "video"
		}

		attachments = append(attachments, model.PostFileAttachment{
			ID:        file.ID,
			ChannelID: channelID,
			PostID:    postID,
			UserID:    user.ID,
			Name:      file.Name,
			Size:      file.Size,
			Width:     width,
			Height:    height,
			MimeType:  file.Type,
			Kind:      kind,
			StorageID: file.StorageID,
		})
	}

	return attachments, nil
}

func (a *App) attachKlipyGIF(
	post *model.Post,
	user model.User,
	channelID string,
	gifID string,
) error {
	if gifID == "" || a.GIF == nil || !*a.ConfigStore.Config.ChannelSettings.KlipyGIF.Enabled {
		return nil
	}

	klipyClient := klipy.New(*a.ConfigStore.Config.ChannelSettings.KlipyGIF.APIKey)

	gifData, err := klipyClient.FindByID(gifID)
	if err != nil {
		return err
	}

	for _, v := range gifData.Results {
		attachment := model.PostFileAttachment{
			ID:        model.NewID(),
			ChannelID: channelID,
			PostID:    post.ID,
			UserID:    user.ID,
			Name:      v.Title,
			MimeType:  "video/mp4",
			URL:       v.MediaFormats["tinymp4"].URL,
			Width:     int64(v.MediaFormats["tinymp4"].Dims[0]),
			Height:    int64(v.MediaFormats["tinymp4"].Dims[1]),
			Provider:  "klipy",
			Kind:      "gifv",
		}

		files, err := a.Store.Posts.CreateAttachment(
			[]model.PostFileAttachment{attachment},
		)
		if err != nil {
			return err
		}

		post.Metadata.Files = append(post.Metadata.Files, files...)
	}

	return nil
}

func (a *App) attachMarkdownGIFs(
	post *model.Post,
	user model.User,
	channelID string,
	message string,
) error {
	re := regexp.MustCompile(`!\[([^\]]*)\]\(([^)]+)\)`)
	matches := re.FindAllStringSubmatch(message, -1)

	for _, m := range matches {
		alt, url := m[1], m[2]

		// Only fetch known GIF hosts; other image links stay as plain markdown.
		if imaging.NormalizeGIF(url).Provider == "unknown" {
			continue
		}

		gif, err := createGIFAttachments(url)
		if err != nil {
			// Enrichment is best-effort; a URL that is not a decodable GIF must not fail the post.
			tlog.Warnw("Failed to decode markdown GIF",
				"post_id", post.ID,
				"channel_id", channelID,
				"error", err,
			)
			continue
		}

		attachment := model.PostFileAttachment{
			ID:        model.NewID(),
			ChannelID: channelID,
			PostID:    post.ID,
			UserID:    user.ID,
			Name:      alt,
			MimeType:  "image/gif",
			URL:       gif.RawURL,
			Width:     int64(gif.Width),
			Height:    int64(gif.Height),
			Provider:  gif.Provider,
			Kind:      "gif",
		}

		files, err := a.Store.Posts.CreateAttachment(
			[]model.PostFileAttachment{attachment},
		)
		if err != nil {
			return err
		}

		post.Metadata.Files = append(post.Metadata.Files, files...)
	}

	return nil
}

func (a *App) CreatePost(
	ctx context.Context,
	user model.User,
	channelID, message, reply string,
	files []model.PostFileMetadata,
	gif string,
	pendingPostID string,
) (*model.PostResponse, *model.AppError) {
	if runes := utf8.RuneCountInString(message); runes > model.MaxPostMessageRunes {
		tlog.Warnw("Rejected oversized post",
			"user_id", user.ID,
			"channel_id", channelID,
			"rune_count", runes,
		)
		return nil, model.NewAppError("post.too_long", http.StatusBadRequest)
	}

	channel, appErr := a.writableChannel(channelID)
	if appErr != nil {
		return nil, appErr
	}

	ok, err := a.Store.Channels.IsMember(channel.ID, user.ID)
	if err != nil {
		tlog.Errorw("Failed to check channel membership",
			"channel_id", channel.ID,
			"user_id", user.ID,
			"error", err,
		)
		return nil, model.NewAppError("channel.not_member", http.StatusInternalServerError)
	}

	if !ok {
		return nil, model.NewAppError("channel.not_member", http.StatusForbidden)
	}

	postResponse := &model.PostResponse{
		Posts:      []model.Post{},
		ReplyPosts: []model.Post{},
	}

	replyPostID := ""
	if reply != "" {
		replyPost, err := a.Store.Posts.Get(ctx, reply)
		if err != nil {
			tlog.Errorw("Reply post lookup failed",
				"post_id", reply,
				"error", err,
			)
			return nil, model.NewAppError("post.invalid_reply", http.StatusBadRequest)
		}

		if !replyPost.HasSameChannel(channel.ID) {
			return nil, model.NewAppError("post.invalid_reply", http.StatusBadRequest)
		}

		replyPostID = replyPost.ID
		postResponse.ReplyPosts = append(postResponse.ReplyPosts, *replyPost)
	}

	now := time.Now().UnixMilli()

	post, err := a.Store.Posts.Create(model.Post{
		ID:        model.NewID(),
		UserID:    user.ID,
		ChannelID: channel.ID,
		Message:   message,
		ReplyID:   replyPostID,
		CreatedAt: now,
		UpdatedAt: now,
	})
	if err != nil {
		tlog.Errorw("Failed to create post",
			"channel_id", channel.ID,
			"user_id", user.ID,
			"error", err,
		)
		return nil, model.NewAppError("post.create_failed", http.StatusInternalServerError)
	}

	post.PendingPostID = pendingPostID

	if replyPostID != "" {
		replyPost, _ := a.Store.Posts.Get(ctx, replyPostID)
		post.Reply = replyPost
	}

	attachments, err := a.handlePostFileAttachments(user, channel.ID, post.ID, files)
	if err != nil {
		tlog.Errorw("Failed to handle post file attachments",
			"post_id", post.ID,
			"error", err,
		)
		return nil, model.NewAppError("post.attachment_failed", http.StatusInternalServerError)
	}

	if len(attachments) > 0 {
		dbFiles, err := a.Store.Posts.CreateAttachment(attachments)
		if err != nil {
			tlog.Errorw("Failed to store post file attachments",
				"post_id", post.ID,
				"error", err,
			)
			return nil, model.NewAppError("post.attachment_failed", http.StatusInternalServerError)
		}

		post.Metadata.Files = dbFiles
	}

	if err := a.attachKlipyGIF(post, user, channel.ID, gif); err != nil {
		tlog.Errorw("Failed to attach Klipy GIF",
			"post_id", post.ID,
			"gif_id", gif,
			"error", err,
		)
		return nil, model.NewAppError("post.attachment_failed", http.StatusInternalServerError)
	}

	if err := a.attachMarkdownGIFs(post, user, channel.ID, message); err != nil {
		tlog.Errorw("Failed to attach markdown GIFs",
			"post_id", post.ID,
			"error", err,
		)
		return nil, model.NewAppError("post.attachment_failed", http.StatusInternalServerError)
	}

	postResponse.Posts = append(postResponse.Posts, *post)
	a.attachResponseMentionUsers(postResponse)

	_ = a.Store.Posts.MarkAllRead(user.ID, channelID)
	mentions := a.resolvePostMentions(*post)
	a.CreatePostNotification(ctx, *post, user, model.NOTIFICATION_CHANNEL_POST, mentions)
	go a.ApplyMentions(*post, user, mentions)
	go a.AttachLinkPreview(post)

	return postResponse, nil
}

func (a *App) UpdatePost(
	ctx context.Context,
	user model.User,
	channelID, postID, message string,
	remove []string,
	files []model.PostFileMetadata,
	gif string,
) (*model.Post, *model.AppError) {
	if runes := utf8.RuneCountInString(message); runes > model.MaxPostMessageRunes {
		tlog.Warnw("Rejected oversized post",
			"user_id", user.ID,
			"channel_id", channelID,
			"rune_count", runes,
		)
		return nil, model.NewAppError("post.too_long", http.StatusBadRequest)
	}

	if _, appErr := a.writableChannel(channelID); appErr != nil {
		return nil, appErr
	}

	post, err := a.Store.Posts.Get(ctx, postID)
	if err != nil {
		tlog.Errorw("Post lookup failed",
			"post_id", postID,
			"error", err,
		)
		return nil, model.NewAppError("post.not_found", http.StatusNotFound)
	}

	ok, err := a.Store.Channels.IsMember(post.ChannelID, user.ID)
	if err != nil {
		tlog.Errorw("Failed to check channel membership",
			"channel_id", post.ChannelID,
			"user_id", user.ID,
			"error", err,
		)
		return nil, model.NewAppError("channel.not_member", http.StatusInternalServerError)
	}

	if !ok || post.UserID != user.ID {
		return nil, model.NewAppError("post.forbidden", http.StatusForbidden)
	}

	attachments, err := a.handlePostFileAttachments(user, post.ChannelID, post.ID, files)
	if err != nil {
		tlog.Errorw("Failed to handle post file attachments",
			"post_id", post.ID,
			"error", err,
		)
		return nil, model.NewAppError("post.attachment_failed", http.StatusInternalServerError)
	}

	if err := a.attachKlipyGIF(post, user, post.ChannelID, gif); err != nil {
		tlog.Errorw("Failed to attach Klipy GIF",
			"post_id", post.ID,
			"gif_id", gif,
			"error", err,
		)
		return nil, model.NewAppError("post.attachment_failed", http.StatusInternalServerError)
	}

	if err := a.attachMarkdownGIFs(post, user, post.ChannelID, message); err != nil {
		tlog.Errorw("Failed to attach markdown GIFs",
			"post_id", post.ID,
			"error", err,
		)
		return nil, model.NewAppError("post.attachment_failed", http.StatusInternalServerError)
	}

	_, err = a.Store.Posts.Update(post.ID, message, remove, attachments)
	if err != nil {
		tlog.Errorw("Failed to update post",
			"post_id", post.ID,
			"user_id", user.ID,
			"error", err,
		)
		return nil, model.NewAppError("post.update_failed", http.StatusInternalServerError)
	}

	newPost, err := a.Store.Posts.Get(ctx, post.ID)
	if err != nil {
		tlog.Errorw("Failed to retrieve post after update",
			"post_id", post.ID,
			"error", err,
		)
		return nil, model.NewAppError("post.not_found", http.StatusInternalServerError)
	}

	a.attachMentionUsers(newPost)

	added := a.resolveHandles(post.ChannelID, addedMentions(post.Message, newPost.Message))
	a.CreatePostNotification(ctx, *newPost, user, model.NOTIFICATION_CHANNEL_POST_UPDATE, added)
	go a.ApplyMentions(*newPost, user, added)
	go a.AttachLinkPreview(newPost)

	return newPost, nil
}

func (a *App) ForwardChannelPost(ctx context.Context, user model.User, channelID, postID string, channels []string, users []string) *model.AppError {
	ok, err := a.Store.Channels.IsMember(channelID, user.ID)
	if err != nil {
		tlog.Errorw("Failed to check channel membership",
			"channel_id", channelID,
			"user_id", user.ID,
			"error", err,
		)
		return model.NewAppError("channel.not_member", http.StatusInternalServerError)
	}

	if !ok {
		return model.NewAppError("channel.not_member", http.StatusForbidden)
	}

	post, err := a.Store.Posts.Get(ctx, postID)
	if err != nil {
		tlog.Errorw("Post lookup failed",
			"post_id", postID,
			"error", err,
		)
		return model.NewAppError("post.not_found", http.StatusNotFound)
	}

	for _, userID := range users {
		ch, err := a.Store.Channels.GetDirectMessage(user.ID, userID)
		if err != nil {
			tlog.Errorw("Failed to retrieve direct message channel",
				"user_id", user.ID,
				"target_id", userID,
				"error", err,
			)
			return model.NewAppError("post.forward_failed", http.StatusInternalServerError)
		}

		if ch == nil {
			newChannel, appErr := a.CreateChannel(user, "", "direct", "", userID)
			if appErr != nil {
				tlog.Errorw("Failed to create direct message channel",
					"user_id", user.ID,
					"target_id", userID,
					"error", appErr,
				)
				return model.NewAppError("post.forward_failed", http.StatusInternalServerError)
			}

			channels = append(channels, newChannel.ID)
		} else {
			channels = append(channels, ch.ID)
		}
	}

	for _, targetChannelID := range channels {
		ok, err := a.Store.Channels.IsMember(targetChannelID, user.ID)
		if err != nil {
			tlog.Errorw("Failed to check channel membership",
				"channel_id", targetChannelID,
				"user_id", user.ID,
				"error", err,
			)
			return model.NewAppError("channel.not_member", http.StatusInternalServerError)
		}

		if !ok {
			return model.NewAppError("channel.not_member", http.StatusForbidden)
		}

		forwardedPost, err := a.Store.Posts.Create(model.Post{
			ID:        model.NewID(),
			UserID:    user.ID,
			ChannelID: targetChannelID,
			Message:   post.Message,
			Type:      model.PostTypeForward,
			CreatedAt: time.Now().UnixMilli(),
			UpdatedAt: time.Now().UnixMilli(),
		})
		if err != nil {
			tlog.Errorw("Failed to create forwarded post",
				"channel_id", targetChannelID,
				"user_id", user.ID,
				"error", err,
			)
			return model.NewAppError("post.forward_failed", http.StatusInternalServerError)
		}

		if len(post.Metadata.Files) > 0 {
			attachments := make([]model.PostFileAttachment, 0, len(post.Metadata.Files))

			for _, f := range post.Metadata.Files {
				newID := model.NewID()

				if f.StorageID != "" {
					backend, exists := a.FileStorageObjects[f.StorageID]
					if !exists {
						tlog.Errorw("Storage backend not found for forward",
							"file_id", f.ID,
							"storage_id", f.StorageID,
						)
						continue
					}

					srcPath, err := a.BuildFilePath(f.StorageID, channelID+"/"+f.ID, model.AppChat)
					if err != nil {
						tlog.Errorw("Failed to build source path for forward",
							"file_id", f.ID,
							"error", err,
						)
						continue
					}

					dstPath, err := a.BuildFilePath(f.StorageID, targetChannelID+"/"+newID, model.AppChat)
					if err != nil {
						tlog.Errorw("Failed to build dest path for forward",
							"file_id", f.ID,
							"error", err,
						)
						continue
					}

					reader, err := backend.ReadFile(context.Background(), srcPath)
					if err != nil {
						tlog.Errorw("Failed to read source attachment for forward",
							"file_id", f.ID,
							"error", err,
						)
						continue
					}

					if err := backend.WriteFile(context.Background(), dstPath, reader, f.Size); err != nil {
						reader.Close()
						tlog.Errorw("Failed to write forwarded attachment",
							"file_id", f.ID,
							"error", err,
						)
						continue
					}

					reader.Close()

					attachments = append(attachments, model.PostFileAttachment{
						ID:        newID,
						ChannelID: targetChannelID,
						PostID:    forwardedPost.ID,
						UserID:    user.ID,
						Name:      f.Name,
						Size:      f.Size,
						Height:    f.Height,
						Width:     f.Width,
						MimeType:  f.MimeType,
						StorageID: f.StorageID,
						CreatedAt: time.Now().UnixMilli(),
						UpdatedAt: time.Now().UnixMilli(),
						Kind:      f.Kind,
					})
				} else {
					// Legacy: no storage, use local paths
					base := path.Join("data/chat")
					srcPath := path.Join(base, channelID, f.ID)
					dstPath := path.Join(base, targetChannelID, newID)

					if err := os.MkdirAll(path.Join(base, targetChannelID), 0o700); err != nil {
						tlog.Errorw("Failed to create attachment directory for forward",
							"channel_id", targetChannelID,
							"error", err,
						)
						continue
					}

					srcFile, err := os.Open(srcPath)
					if err != nil {
						tlog.Errorw("Failed to open source attachment",
							"file_id", f.ID,
							"error", err,
						)
						continue
					}

					dstFile, err := os.Create(dstPath)
					if err != nil {
						srcFile.Close()
						tlog.Errorw("Failed to create forwarded attachment file",
							"file_id", f.ID,
							"error", err,
						)
						continue
					}

					if _, err := io.Copy(dstFile, srcFile); err != nil {
						tlog.Errorw("Failed to copy attachment for forward",
							"file_id", f.ID,
							"error", err,
						)
					}

					srcFile.Close()
					dstFile.Close()

					attachments = append(attachments, model.PostFileAttachment{
						ID:        newID,
						ChannelID: targetChannelID,
						PostID:    forwardedPost.ID,
						UserID:    user.ID,
						Name:      f.Name,
						Size:      f.Size,
						Height:    f.Height,
						Width:     f.Width,
						MimeType:  f.MimeType,
						CreatedAt: time.Now().UnixMilli(),
						UpdatedAt: time.Now().UnixMilli(),
						Provider:  f.Provider,
						URL:       f.URL,
						Kind:      f.Kind,
					})
				}
			}

			savedAttachments, err := a.Store.Posts.CreateAttachment(attachments)
			if err != nil {
				tlog.Errorw("Failed to store forwarded post attachments",
					"post_id", forwardedPost.ID,
					"error", err,
				)
			} else {
				forwardedPost.Metadata.Files = savedAttachments
			}
		}

		mentions := a.resolvePostMentions(*forwardedPost)
		a.CreatePostNotification(ctx, *forwardedPost, user, model.NOTIFICATION_CHANNEL_POST, mentions)
		go a.ApplyMentions(*forwardedPost, user, mentions)
	}

	return nil
}

func (a *App) GetChannelByID(user model.User, channelID string) (*model.Channel, *model.AppError) {
	channel, err := a.Store.Channels.Get(channelID)
	if err != nil {
		tlog.Errorw("Channel lookup failed",
			"channel_id", channelID,
			"error", err,
		)
		return nil, model.NewAppError("channel.not_found", http.StatusNotFound)
	}

	ok, err := a.Store.Channels.IsMember(channel.ID, user.ID)
	if err != nil {
		tlog.Errorw("Failed to check channel membership",
			"channel_id", channel.ID,
			"user_id", user.ID,
			"error", err,
		)
		return nil, model.NewAppError("channel.not_member", http.StatusInternalServerError)
	}

	if !ok && !a.SessionHasPermission(user, model.AdminPermissions.PermissionManageChannels) {
		return nil, model.NewAppError("channel.not_member", http.StatusForbidden)
	}

	members, err := a.Store.Channels.GetDirectMembersForUser(channel.ID, user.ID)
	if err != nil {
		tlog.Errorw("Failed to retrieve channel members",
			"channel_id", channel.ID,
			"user_id", user.ID,
			"error", err,
		)
		return nil, model.NewAppError("channel.retrieval_failed", http.StatusInternalServerError)
	}

	channel.ChannelMembers = members

	return channel, nil
}

func (a *App) ReadChannelPosts(user model.User, channelID string) *model.AppError {
	err := a.Store.Posts.MarkAllRead(user.ID, channelID)
	if err != nil {
		tlog.Errorw("Failed to mark channel posts as read",
			"channel_id", channelID,
			"user_id", user.ID,
			"error", err,
		)
		return model.NewAppError("channel.read_failed", http.StatusInternalServerError)
	}

	return nil
}

func (a *App) CreatePostReaction(ctx context.Context, user model.User, channelID, postID, reaction string) (*model.PostReaction, *model.AppError) {
	if !model.IsValidReaction(reaction) {
		tlog.Warnw("Rejected malformed reaction",
			"user_id", user.ID,
			"channel_id", channelID,
		)
		return nil, model.NewAppError("post.invalid_reaction", http.StatusBadRequest)
	}

	if _, appErr := a.writableChannel(channelID); appErr != nil {
		return nil, appErr
	}

	ok, err := a.Store.Channels.IsMember(channelID, user.ID)
	if err != nil {
		tlog.Errorw("Failed to check channel membership",
			"channel_id", channelID,
			"user_id", user.ID,
			"error", err,
		)
		return nil, model.NewAppError("channel.not_member", http.StatusInternalServerError)
	}

	if !ok {
		return nil, model.NewAppError("channel.not_member", http.StatusForbidden)
	}

	postReaction, err := a.Store.Posts.CreateReaction(postID, channelID, user.ID, reaction)
	if err != nil {
		tlog.Errorw("Failed to create post reaction",
			"post_id", postID,
			"channel_id", channelID,
			"user_id", user.ID,
			"error", err,
		)
		return nil, model.NewAppError("post.reaction_create_failed", http.StatusInternalServerError)
	}

	a.CreatePostReactionNotification(ctx, postID, user, model.NOTIFICATION_CHANNEL_POST_REACTION)
	return postReaction, nil
}

func (a *App) DeletePost(ctx context.Context, user model.User, channelID, postID string) *model.AppError {
	if _, appErr := a.writableChannel(channelID); appErr != nil {
		return appErr
	}

	ok, err := a.Store.Channels.IsMember(channelID, user.ID)
	if err != nil {
		tlog.Errorw("Failed to check channel membership",
			"channel_id", channelID,
			"user_id", user.ID,
			"error", err,
		)
		return model.NewAppError("channel.not_member", http.StatusInternalServerError)
	}

	if !ok {
		return model.NewAppError("channel.not_member", http.StatusForbidden)
	}

	post, err := a.Store.Posts.Get(ctx, postID)
	if err != nil {
		tlog.Errorw("Post lookup failed",
			"post_id", postID,
			"error", err,
		)
		return model.NewAppError("post.not_found", http.StatusNotFound)
	}

	if user.ID != post.UserID {
		return model.NewAppError("post.forbidden", http.StatusForbidden)
	}

	// Read before the rows go, since they are the only record of what to remove.
	attachments, err := a.Store.Posts.GetAttachments(postID)
	if err != nil {
		tlog.Warnw("Could not list post attachments, their files will be left behind",
			"post_id", postID,
			"error", err,
		)
	}

	err = a.Store.Posts.Delete(postID)
	if err != nil {
		tlog.Errorw("Failed to delete post",
			"post_id", postID,
			"user_id", user.ID,
			"error", err,
		)
		return model.NewAppError("post.delete_failed", http.StatusInternalServerError)
	}

	a.removeChatAttachmentFiles(channelID, attachments)

	a.CreatePostNotification(ctx, *post, user, model.NOTIFICATION_CHANNEL_POST_DELETE, PostMentions{})
	return nil
}

// Best effort: the post is already gone, so a storage failure must not fail the
// request. It is logged because the row that pointed at the file is gone too,
// leaving nothing else to find the orphan by.
func (a *App) removeChatAttachmentFiles(channelID string, attachments []model.PostFileAttachment) {
	for _, attachment := range attachments {
		// Klipy GIFs and link previews are remote URLs with nothing stored here.
		if attachment.StorageID == "" {
			continue
		}

		backend, ok := a.FileStorageObjects[attachment.StorageID]
		if !ok {
			tlog.Warnw("No storage backend for a deleted attachment",
				"attachment_id", attachment.ID,
				"storage_id", attachment.StorageID,
			)
			continue
		}

		path, err := a.BuildFilePath(attachment.StorageID, channelID+"/"+attachment.ID, model.AppChat)
		if err != nil {
			tlog.Warnw("Could not build the path of a deleted attachment",
				"attachment_id", attachment.ID,
				"error", err,
			)
			continue
		}

		if err := backend.RemoveFile(context.Background(), path); err != nil {
			tlog.Warnw("Could not remove the file of a deleted attachment",
				"attachment_id", attachment.ID,
				"path", path,
				"error", err,
			)
		}
	}
}

func (a *App) DeletePostReaction(ctx context.Context, user model.User, channelID, postID, reaction string) *model.AppError {
	if _, appErr := a.writableChannel(channelID); appErr != nil {
		return appErr
	}

	ok, err := a.Store.Channels.IsMember(channelID, user.ID)
	if err != nil {
		tlog.Errorw("Failed to check channel membership",
			"channel_id", channelID,
			"user_id", user.ID,
			"error", err,
		)
		return model.NewAppError("channel.not_member", http.StatusInternalServerError)
	}

	if !ok {
		return model.NewAppError("channel.not_member", http.StatusForbidden)
	}

	err = a.Store.Posts.DeleteReaction(postID, channelID, user.ID, reaction)
	if err != nil {
		tlog.Errorw("Failed to delete post reaction",
			"post_id", postID,
			"channel_id", channelID,
			"user_id", user.ID,
			"error", err,
		)
		return model.NewAppError("post.reaction_delete_failed", http.StatusInternalServerError)
	}

	a.CreatePostReactionNotification(ctx, postID, user, model.NOTIFICATION_CHANNEL_POST_REACTION)
	return nil
}

func channelExistsInList(channels []model.Channel, channelID string) bool {
	for _, channel := range channels {
		if channel.ID == channelID {
			return true
		}
	}

	return false
}

// The GIF URL comes from user post content, so guard the fetch against SSRF.
var gifFetchClient = safehttp.NewClient(safehttp.Config{Timeout: 10 * time.Second})

func createGIFAttachments(url string) (*model.GifNormalized, error) {
	gifNormalized := imaging.NormalizeGIF(url)

	// find the width and height from the URL response

	req, err := http.NewRequest("GET", gifNormalized.RawURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Range", "bytes=0-50000")

	resp, err := gifFetchClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	// Read only what we got (should be partial gif data)
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Decode GIF header (does NOT require full GIF)
	cfg, err := gif.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	gifNormalized.Width = cfg.Width
	gifNormalized.Height = cfg.Height

	return &gifNormalized, nil
}

// Reads keep using requireChannelMember, so an archived channel stays readable.
func (a *App) requireWritableChannelMember(user model.User, channelID string) (*model.Channel, *model.AppError) {
	channel, appErr := a.requireChannelMember(user, channelID)
	if appErr != nil {
		return nil, appErr
	}

	if channel.DeletedAt != 0 {
		return nil, model.NewAppError("channel.archived", http.StatusConflict)
	}

	return channel, nil
}

const meetingEmptyGrace = 2 * time.Minute

func (a *App) requireChannelMember(user model.User, channelID string) (*model.Channel, *model.AppError) {
	channel, err := a.Store.Channels.Get(channelID)
	if err != nil {
		tlog.Errorw("Failed to retrieve channel", "channel_id", channelID, "error", err)
		return nil, model.NewAppError("channel.not_found", http.StatusInternalServerError)
	}

	if channel == nil {
		return nil, model.NewAppError("channel.not_found", http.StatusNotFound)
	}

	if user.Role == model.SystemAdminRoleId {
		return channel, nil
	}

	ok, err := a.Store.Channels.IsMember(channelID, user.ID)
	if err != nil {
		tlog.Errorw("Failed to check channel membership", "channel_id", channelID, "user_id", user.ID, "error", err)
		return nil, model.NewAppError("channel.not_member", http.StatusInternalServerError)
	}

	if !ok {
		return nil, model.NewAppError("channel.not_member", http.StatusForbidden)
	}

	return channel, nil
}

func (a *App) getMeetingForChannel(ctx context.Context, channelID, meetingID string) (*model.ChannelMeeting, *model.AppError) {
	meeting, err := a.Store.Channels.GetMeeting(ctx, meetingID)
	if err != nil {
		tlog.Errorw("Failed to retrieve meeting", "meeting_id", meetingID, "error", err)
		return nil, model.NewAppError("channel.meeting_failed", http.StatusInternalServerError)
	}

	if meeting == nil || meeting.ChannelID != channelID {
		return nil, model.NewAppError("channel.meeting_not_found", http.StatusNotFound)
	}

	return meeting, nil
}

func (a *App) requireMeetingHost(user model.User, meeting *model.ChannelMeeting) *model.AppError {
	if user.Role == model.SystemAdminRoleId || meeting.HostID == user.ID {
		return nil
	}

	return model.NewAppError("channel.meeting_host_only", http.StatusForbidden)
}

func defaultMeetingTitle(user model.User) string {
	name := strings.TrimSpace(user.Name + " " + user.LastName)
	if name == "" {
		return "Meeting"
	}

	return name + "'s meeting"
}

// Invitees need not be channel members, being invited grants access to this meeting only.
func (a *App) sanitizeInvitees(invitees []string, hostID string) []string {
	seen := map[string]bool{hostID: true}
	clean := []string{}
	for _, id := range invitees {
		id = strings.TrimSpace(id)
		if id == "" || seen[id] {
			continue
		}

		seen[id] = true
		user, err := a.Store.User.Get(id)
		if err != nil || user == nil {
			continue
		}

		clean = append(clean, id)
	}

	return clean
}

// The access checks read invite_only while the meeting lists read the invitee and
// group rows, so the flag is recomputed whenever those rows change.
func (a *App) syncMeetingInviteOnly(ctx context.Context, meeting *model.ChannelMeeting) error {
	groups, err := a.Store.Channels.GetMeetingGroups(ctx, meeting.ID)
	if err != nil {
		return err
	}

	inviteOnly := len(meeting.Invitees) > 0 || len(groups) > 0
	if inviteOnly == meeting.InviteOnly {
		return nil
	}

	if err := a.Store.Channels.UpdateMeetingInviteOnly(ctx, meeting.ID, inviteOnly); err != nil {
		return err
	}

	meeting.InviteOnly = inviteOnly
	return nil
}

// Assumes the caller already checked channel membership, so open meetings are visible to anyone here.
func (a *App) meetingAccessAllowed(ctx context.Context, user model.User, meeting *model.ChannelMeeting) bool {
	if user.Role == model.SystemAdminRoleId || meeting.HostID == user.ID {
		return true
	}

	for _, id := range meeting.Invitees {
		if id == user.ID {
			return true
		}
	}

	if !meeting.InviteOnly {
		return true
	}

	inGroup, err := a.Store.Channels.UserInMeetingGroup(ctx, meeting.ID, user.ID)
	if err != nil {
		tlog.Warnw("Failed to check meeting group access", "meeting_id", meeting.ID, "user_id", user.ID, "error", err)
		return false
	}

	return inGroup
}

// Unlike meetingAccessAllowed, this resolves channel membership itself for non-channel-scoped entry points.
func (a *App) canAccessMeeting(ctx context.Context, user model.User, meeting *model.ChannelMeeting) bool {
	if user.Role == model.SystemAdminRoleId || meeting.HostID == user.ID {
		return true
	}

	for _, id := range meeting.Invitees {
		if id == user.ID {
			return true
		}
	}

	inGroup, err := a.Store.Channels.UserInMeetingGroup(ctx, meeting.ID, user.ID)
	if err != nil {
		tlog.Warnw("Failed to check meeting group access", "meeting_id", meeting.ID, "user_id", user.ID, "error", err)
	} else if inGroup {
		return true
	}

	if !meeting.InviteOnly {
		ok, err := a.Store.Channels.IsMember(meeting.ChannelID, user.ID)
		return err == nil && ok
	}

	return false
}

func (a *App) GetChannelMemberUsers(user model.User, channelID string) ([]model.ChannelMemberUser, *model.AppError) {
	if _, appErr := a.requireChannelMember(user, channelID); appErr != nil {
		return nil, appErr
	}

	members, err := a.Store.Channels.GetMembers(channelID)
	if err != nil {
		tlog.Errorw("Failed to retrieve channel members", "channel_id", channelID, "error", err)
		return nil, model.NewAppError("channel.retrieval_failed", http.StatusInternalServerError)
	}

	ids := make([]string, 0, len(members))
	for _, m := range members {
		ids = append(ids, m.UserID)
	}

	users, err := a.Store.User.GetByIDs(ids)
	if err != nil {
		tlog.Errorw("Failed to retrieve member users", "channel_id", channelID, "error", err)
		return nil, model.NewAppError("channel.retrieval_failed", http.StatusInternalServerError)
	}

	result := make([]model.ChannelMemberUser, 0, len(users))
	for _, u := range users {
		result = append(result, model.ChannelMemberUser{
			ID:       u.ID,
			Name:     u.Name,
			LastName: u.LastName,
			Username: u.Username,
		})
	}

	return result, nil
}

func (a *App) SearchChannelMemberUsers(ctx context.Context, user model.User, channelID, query string, limit int) ([]model.User, *model.AppError) {
	if _, appErr := a.requireChannelMember(user, channelID); appErr != nil {
		return nil, appErr
	}

	if limit <= 0 || limit > 50 {
		limit = 25
	}

	ctx, cancel := a.dbCtx(ctx)
	defer cancel()

	users, err := a.Store.Channels.SearchMembers(ctx, channelID, query, limit)
	if err != nil {
		tlog.Errorw("Failed to search channel members", "channel_id", channelID, "error", err)
		return nil, model.NewAppError("channel.retrieval_failed", http.StatusInternalServerError)
	}

	return users, nil
}

func (a *App) CreateChannelMeeting(ctx context.Context, user model.User, channelID string, scheduledAt int64, title string, durationMinutes int, invitees []string, guestPassword, timezone string, groupIDs []string) (*model.ChannelMeeting, *model.AppError) {
	channel, appErr := a.requireWritableChannelMember(user, channelID)
	if appErr != nil {
		return nil, appErr
	}

	now := time.Now().Unix()
	title = strings.TrimSpace(title)
	if title == "" {
		title = defaultMeetingTitle(user)
	}

	if durationMinutes <= 0 {
		durationMinutes = 60
	}

	meeting := model.ChannelMeeting{
		ID:              model.NewID(),
		ChannelID:       channelID,
		HostID:          user.ID,
		CreatedBy:       user.ID,
		Title:           title,
		DurationMinutes: durationMinutes,
		Timezone:        strings.TrimSpace(timezone),
		CreatedAt:       now,
		Invitees:        a.sanitizeInvitees(invitees, user.ID),
	}

	cleanGroups := DedupeIDs(groupIDs)
	if len(cleanGroups) > 0 {
		groups, err := a.Store.Groups.GetByIDs(ctx, cleanGroups)
		if err != nil {
			tlog.Errorw("Failed to retrieve groups for meeting", "channel_id", channelID, "error", err)
			return nil, model.NewAppError("group.retrieval_failed", http.StatusInternalServerError)
		}

		if len(groups) != len(cleanGroups) {
			return nil, model.NewAppError("group.not_found", http.StatusNotFound)
		}
	}
	// Selecting members or groups at creation makes it invite-only, otherwise it stays open to the channel.
	meeting.InviteOnly = len(meeting.Invitees) > 0 || len(cleanGroups) > 0
	if scheduledAt > now {
		meeting.ScheduledAt = &scheduledAt
	} else {
		meeting.StartedAt = &now
	}

	guestPassword = strings.TrimSpace(guestPassword)
	if guestPassword != "" {
		hashed, err := crypto.HashPassword(guestPassword)
		if err != nil {
			tlog.Errorw("Failed to hash guest password", "channel_id", channelID, "error", err)
			return nil, model.NewAppError("channel.meeting_create_failed", http.StatusInternalServerError)
		}

		meeting.GuestPassword = hashed
	}

	if err := a.Store.Channels.CreateMeeting(ctx, meeting); err != nil {
		tlog.Errorw("Failed to create meeting", "channel_id", channelID, "error", err)
		return nil, model.NewAppError("channel.meeting_create_failed", http.StatusInternalServerError)
	}

	if len(cleanGroups) > 0 {
		if err := a.Store.Channels.AddMeetingGroups(ctx, meeting.ID, cleanGroups, user.ID); err != nil {
			tlog.Errorw("Failed to add groups to meeting", "meeting_id", meeting.ID, "error", err)
			return nil, model.NewAppError("meeting_group.add_failed", http.StatusInternalServerError)
		}
	}

	meeting.SetStatus()

	if meeting.ScheduledAt != nil {
		go a.NotifyMeetingScheduled(user, *channel, meeting)
		go a.RefreshMeetingLists(meeting, channelID, nil)
	} else if meeting.StartedAt != nil {
		go a.BroadcastMeetingStarted(meeting, channelID)
	}

	return &meeting, nil
}

func (a *App) GetActiveChannelMeetings(ctx context.Context, user model.User, channelID string) ([]model.ChannelMeeting, *model.AppError) {
	if _, appErr := a.requireChannelMember(user, channelID); appErr != nil {
		return nil, appErr
	}

	meetings, err := a.Store.Channels.GetActiveMeetings(ctx, channelID)
	if err != nil {
		tlog.Errorw("Failed to list active meetings", "channel_id", channelID, "error", err)
		return nil, model.NewAppError("channel.meeting_failed", http.StatusInternalServerError)
	}

	counts := a.roomParticipantCounts(ctx, meetings)
	now := time.Now().Unix()

	active := make([]model.ChannelMeeting, 0, len(meetings))
	for _, m := range meetings {
		count := counts[m.ID]
		var startedAt int64
		if m.StartedAt != nil {
			startedAt = *m.StartedAt
		}

		if count == 0 && now-startedAt > int64(meetingEmptyGrace.Seconds()) {
			a.endMeeting(ctx, m.ID)
			continue
		}

		if !a.meetingAccessAllowed(ctx, user, &m) {
			continue
		}

		m.Participants = count
		m.SetStatus()
		active = append(active, m)
	}

	return active, nil
}

func (a *App) GetScheduledChannelMeetings(ctx context.Context, user model.User, channelID string) ([]model.ChannelMeeting, *model.AppError) {
	if _, appErr := a.requireChannelMember(user, channelID); appErr != nil {
		return nil, appErr
	}

	meetings, err := a.Store.Channels.GetScheduledMeetings(ctx, channelID)
	if err != nil {
		tlog.Errorw("Failed to list scheduled meetings", "channel_id", channelID, "error", err)
		return nil, model.NewAppError("channel.meeting_failed", http.StatusInternalServerError)
	}

	visible := make([]model.ChannelMeeting, 0, len(meetings))
	for i := range meetings {
		if !a.meetingAccessAllowed(ctx, user, &meetings[i]) {
			continue
		}

		meetings[i].SetStatus()
		visible = append(visible, meetings[i])
	}

	return visible, nil
}

// Carries channel context because a non-member won't have the channel in their client store.
type MeetingWithChannel struct {
	model.ChannelMeeting
	ChannelName string `json:"channel_name"`
	ChannelType string `json:"channel_type"`
}

// No channel-membership gate, visibility comes from the query so an invited non-member still sees their meeting.
func (a *App) GetUserMeetings(ctx context.Context, user model.User) ([]MeetingWithChannel, *model.AppError) {
	meetings, err := a.Store.Channels.GetMeetingsForUser(ctx, user.ID)
	if err != nil {
		tlog.Errorw("Failed to list user meetings", "user_id", user.ID, "error", err)
		return nil, model.NewAppError("channel.meeting_failed", http.StatusInternalServerError)
	}

	counts := a.roomParticipantCounts(ctx, meetings)
	now := time.Now().Unix()

	channels := map[string]*model.Channel{}
	result := make([]MeetingWithChannel, 0, len(meetings))
	for _, m := range meetings {
		if m.StartedAt != nil {
			count := counts[m.ID]
			if count == 0 && now-*m.StartedAt > int64(meetingEmptyGrace.Seconds()) {
				a.endMeeting(ctx, m.ID)
				continue
			}

			m.Participants = count
		}

		m.SetStatus()

		ch, ok := channels[m.ChannelID]
		if !ok {
			ch, _ = a.Store.Channels.Get(m.ChannelID)
			channels[m.ChannelID] = ch
		}

		item := MeetingWithChannel{ChannelMeeting: m}
		if ch != nil {
			item.ChannelName = ch.DisplayName
			item.ChannelType = ch.Type
		}

		result = append(result, item)
	}

	return result, nil
}

func (a *App) GetMeeting(ctx context.Context, user model.User, meetingID string) (*MeetingWithChannel, *model.AppError) {
	meeting, err := a.Store.Channels.GetMeeting(ctx, meetingID)
	if err != nil {
		tlog.Errorw("Failed to retrieve meeting", "meeting_id", meetingID, "error", err)
		return nil, model.NewAppError("channel.meeting_failed", http.StatusInternalServerError)
	}

	if meeting == nil {
		return nil, model.NewAppError("channel.meeting_not_found", http.StatusNotFound)
	}

	if !a.canAccessMeeting(ctx, user, meeting) {
		return nil, model.NewAppError("channel.meeting_not_invited", http.StatusForbidden)
	}

	if meeting.StartedAt != nil && meeting.EndedAt == nil {
		counts := a.roomParticipantCounts(ctx, []model.ChannelMeeting{*meeting})
		meeting.Participants = counts[meeting.ID]
	}

	meeting.SetStatus()

	item := &MeetingWithChannel{ChannelMeeting: *meeting}
	if ch, _ := a.Store.Channels.Get(meeting.ChannelID); ch != nil {
		item.ChannelName = ch.DisplayName
		item.ChannelType = ch.Type
	}

	return item, nil
}

// Host only.
func (a *App) StartChannelMeeting(ctx context.Context, user model.User, channelID, meetingID string) (*model.ChannelMeeting, *model.AppError) {
	if _, appErr := a.requireWritableChannelMember(user, channelID); appErr != nil {
		return nil, appErr
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

	if meeting.StartedAt == nil {
		now := time.Now().Unix()
		if err := a.Store.Channels.StartMeeting(ctx, meetingID, now); err != nil {
			tlog.Errorw("Failed to start meeting", "meeting_id", meetingID, "error", err)
			return nil, model.NewAppError("channel.meeting_failed", http.StatusInternalServerError)
		}

		meeting.StartedAt = &now
		// Signal invitees so their lobby auto-joins and lists update.
		go a.BroadcastMeetingStarted(*meeting, channelID)
	}

	meeting.SetStatus()
	return meeting, nil
}

// Host only.
func (a *App) EndChannelMeeting(ctx context.Context, user model.User, channelID, meetingID string) *model.AppError {
	if _, appErr := a.requireChannelMember(user, channelID); appErr != nil {
		return appErr
	}

	meeting, appErr := a.getMeetingForChannel(ctx, channelID, meetingID)
	if appErr != nil {
		return appErr
	}

	if appErr := a.requireMeetingHost(user, meeting); appErr != nil {
		return appErr
	}

	if meeting.EndedAt != nil {
		return nil
	}

	a.endMeeting(ctx, meetingID)
	go a.BroadcastMeetingEnded(*meeting, channelID)
	return nil
}

const directCallRingTimeout = 45 * time.Second

// A 1:1 call rings before any meeting or room exists; those are created only on
// answer. The mutex is load-bearing, not just map safety: the atomic
// check-and-set is what stops two simultaneous callers from both ringing.
type directCallRegistry struct {
	mu    sync.Mutex
	calls map[string]ringingCall
}

type ringingCall struct {
	callerID string
	timer    *time.Timer
}

// If the other member is already ringing, ringOrAnswer consumes that ring and
// returns them so the press becomes an answer instead of a second ring.
func (a *App) ringOrAnswer(channelID, callerID string) (answeredCaller string, isAnswer bool) {
	a.directCalls.mu.Lock()
	defer a.directCalls.mu.Unlock()
	if a.directCalls.calls == nil {
		a.directCalls.calls = make(map[string]ringingCall)
	}

	if existing, ok := a.directCalls.calls[channelID]; ok {
		existing.timer.Stop()
		if existing.callerID != callerID {
			delete(a.directCalls.calls, channelID)
			return existing.callerID, true
		}
	}

	timer := time.AfterFunc(directCallRingTimeout, func() { a.expireDirectCall(channelID, callerID) })
	a.directCalls.calls[channelID] = ringingCall{callerID: callerID, timer: timer}
	return "", false
}

func (a *App) clearRinging(channelID string) (callerID string, existed bool) {
	a.directCalls.mu.Lock()
	defer a.directCalls.mu.Unlock()
	existing, ok := a.directCalls.calls[channelID]
	if !ok {
		return "", false
	}

	existing.timer.Stop()
	delete(a.directCalls.calls, channelID)
	return existing.callerID, true
}

// Ring-timeout callback. The caller check makes a late timer a no-op once the
// ring was answered or replaced, so it can't cancel a call that has gone live.
func (a *App) expireDirectCall(channelID, callerID string) {
	a.directCalls.mu.Lock()
	existing, ok := a.directCalls.calls[channelID]
	if !ok || existing.callerID != callerID {
		a.directCalls.mu.Unlock()
		return
	}

	delete(a.directCalls.calls, channelID)
	a.directCalls.mu.Unlock()
	a.NotifyCallCancelled(channelID, "")
}

// The two DM members are the access list, so there are no invitees, scheduling or email.
func (a *App) StartDirectCall(ctx context.Context, user model.User, channelID string) (*model.DirectCallResult, *model.AppError) {
	channel, appErr := a.requireChannelMember(user, channelID)
	if appErr != nil {
		return nil, appErr
	}

	if channel.Type != model.ChannelTypeDirect {
		return nil, model.NewAppError("channel.not_direct", http.StatusBadRequest)
	}

	// A 1:1 call has no rejoin path: it ends the instant either side drops, so
	// every press mints a fresh room rather than risk landing in a room LiveKit
	// is still tearing down.
	caller, isAnswer := a.ringOrAnswer(channelID, user.ID)
	if !isAnswer {
		go a.RingDirectCall(user, channelID)
		return &model.DirectCallResult{Status: model.DirectCallRinging}, nil
	}

	return a.connectDirectCall(ctx, channelID, caller)
}

// AnswerDirectCall never starts a new ring (unlike StartDirectCall), so a stale
// Accept (the caller already hung up) resolves to "ended" instead of ringing
// them back.
func (a *App) AnswerDirectCall(ctx context.Context, user model.User, channelID string) (*model.DirectCallResult, *model.AppError) {
	channel, appErr := a.requireChannelMember(user, channelID)
	if appErr != nil {
		return nil, appErr
	}

	if channel.Type != model.ChannelTypeDirect {
		return nil, model.NewAppError("channel.not_direct", http.StatusBadRequest)
	}

	caller, ok := a.clearRinging(channelID)
	if !ok || caller == user.ID {
		return &model.DirectCallResult{Status: model.DirectCallEnded}, nil
	}

	return a.connectDirectCall(ctx, channelID, caller)
}

// The meeting is hosted by the original caller, not the answerer who creates it.
func (a *App) connectDirectCall(ctx context.Context, channelID, callerID string) (*model.DirectCallResult, *model.AppError) {
	now := time.Now().Unix()
	meeting := model.ChannelMeeting{
		ID:              model.NewID(),
		ChannelID:       channelID,
		HostID:          callerID,
		CreatedBy:       callerID,
		Title:           "Call",
		DurationMinutes: 60,
		CreatedAt:       now,
		StartedAt:       &now,
	}

	if err := a.Store.Channels.CreateMeeting(ctx, meeting); err != nil {
		tlog.Errorw("Failed to start direct call", "channel_id", channelID, "error", err)
		return nil, model.NewAppError("channel.meeting_create_failed", http.StatusInternalServerError)
	}

	meeting.SetStatus()
	go a.NotifyCallConnected(channelID, meeting, callerID)
	go a.NotifyCallHandled(channelID, callerID)
	return &model.DirectCallResult{Status: model.DirectCallConnected, Meeting: &meeting}, nil
}

func (a *App) DeclineDirectCall(ctx context.Context, user model.User, channelID string) *model.AppError {
	channel, appErr := a.requireChannelMember(user, channelID)
	if appErr != nil {
		return appErr
	}

	if channel.Type != model.ChannelTypeDirect {
		return model.NewAppError("channel.not_direct", http.StatusBadRequest)
	}

	caller, ok := a.clearRinging(channelID)
	if !ok {
		return nil
	}

	go a.NotifyCallDeclined(user, channelID, caller)
	go a.NotifyCallHandled(channelID, caller)
	return nil
}

func (a *App) CancelDirectCall(ctx context.Context, user model.User, channelID string) *model.AppError {
	channel, appErr := a.requireChannelMember(user, channelID)
	if appErr != nil {
		return appErr
	}

	if channel.Type != model.ChannelTypeDirect {
		return model.NewAppError("channel.not_direct", http.StatusBadRequest)
	}

	if _, ok := a.clearRinging(channelID); !ok {
		return nil
	}

	go a.NotifyCallCancelled(channelID, user.ID)
	return nil
}

// EndDirectCall ends the call authoritatively (no host check, unlike
// EndChannelMeeting) because in a 1:1 either member leaving means it is over.
// Idempotent, so both sides leaving or a stray re-press is a harmless no-op.
func (a *App) EndDirectCall(ctx context.Context, user model.User, channelID string) *model.AppError {
	channel, appErr := a.requireChannelMember(user, channelID)
	if appErr != nil {
		return appErr
	}

	if channel.Type != model.ChannelTypeDirect {
		return model.NewAppError("channel.not_direct", http.StatusBadRequest)
	}

	active, err := a.Store.Channels.GetActiveMeetings(ctx, channelID)
	if err != nil {
		tlog.Errorw("Failed to list active calls", "channel_id", channelID, "error", err)
		return model.NewAppError("channel.meeting_failed", http.StatusInternalServerError)
	}

	for _, m := range active {
		a.endMeeting(ctx, m.ID)
	}

	return nil
}

// Host only.
func (a *App) CancelChannelMeeting(ctx context.Context, user model.User, channelID, meetingID string) *model.AppError {
	channel, appErr := a.requireWritableChannelMember(user, channelID)
	if appErr != nil {
		return appErr
	}

	meeting, appErr := a.getMeetingForChannel(ctx, channelID, meetingID)
	if appErr != nil {
		return appErr
	}

	if appErr := a.requireMeetingHost(user, meeting); appErr != nil {
		return appErr
	}

	if meeting.EndedAt != nil {
		return nil
	}

	if meeting.StartedAt != nil {
		return model.NewAppError("channel.meeting_already_started", http.StatusConflict)
	}

	// Collect guest emails before the links are revoked.
	guestEmails := []string{}
	if links, err := a.Store.Channels.GetActiveMeetingGuestLinks(ctx, meetingID, time.Now().Unix()); err == nil {
		for _, l := range links {
			guestEmails = append(guestEmails, l.InvitedEmail)
		}
	}

	a.endMeeting(ctx, meetingID)

	// Bump the sequence so the cancellation ICS supersedes the last invite.
	if seq, err := a.Store.Channels.IncrementMeetingSequence(ctx, meetingID); err != nil {
		tlog.Errorw("Failed to bump meeting sequence on cancel", "meeting_id", meetingID, "error", err)
	} else {
		meeting.Sequence = seq
	}

	go a.NotifyMeetingCancelled(user, *channel, *meeting, guestEmails)
	go a.RefreshMeetingLists(*meeting, channelID, nil)

	return nil
}

// Host only.
func (a *App) UpdateChannelMeeting(ctx context.Context, user model.User, channelID, meetingID string, scheduledAt int64, title string, durationMinutes int, invitees []string, guestPassword, timezone string, groupIDs []string) (*model.ChannelMeeting, *model.AppError) {
	channel, appErr := a.requireWritableChannelMember(user, channelID)
	if appErr != nil {
		return nil, appErr
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

	if meeting.StartedAt != nil {
		return nil, model.NewAppError("channel.meeting_already_started", http.StatusConflict)
	}

	now := time.Now().Unix()
	if scheduledAt <= now {
		return nil, model.NewAppError("channel.meeting_past_time", http.StatusBadRequest)
	}

	title = strings.TrimSpace(title)
	if title == "" {
		title = defaultMeetingTitle(user)
	}

	if durationMinutes <= 0 {
		durationMinutes = 60
	}

	// A blank guest password leaves the existing one untouched.
	guestPassword = strings.TrimSpace(guestPassword)
	if guestPassword != "" {
		hashed, err := crypto.HashPassword(guestPassword)
		if err != nil {
			tlog.Errorw("Failed to hash guest password", "channel_id", channelID, "error", err)
			return nil, model.NewAppError("channel.meeting_update_failed", http.StatusInternalServerError)
		}

		meeting.GuestPassword = hashed
	}

	var cleanGroups []string
	if groupIDs != nil {
		cleanGroups = DedupeIDs(groupIDs)
		if len(cleanGroups) > 0 {
			groups, err := a.Store.Groups.GetByIDs(ctx, cleanGroups)
			if err != nil {
				tlog.Errorw("Failed to retrieve groups for meeting update", "meeting_id", meetingID, "error", err)
				return nil, model.NewAppError("group.retrieval_failed", http.StatusInternalServerError)
			}

			if len(groups) != len(cleanGroups) {
				return nil, model.NewAppError("group.not_found", http.StatusNotFound)
			}
		}
	}

	previousInvitees := meeting.Invitees
	previousAudience := a.meetingAudience(*meeting, channelID)
	meeting.Title = title
	meeting.DurationMinutes = durationMinutes
	meeting.Timezone = strings.TrimSpace(timezone)
	meeting.ScheduledAt = &scheduledAt
	meeting.Invitees = a.sanitizeInvitees(invitees, meeting.HostID)

	seq, err := a.Store.Channels.UpdateMeeting(ctx, *meeting)
	if err != nil {
		tlog.Errorw("Failed to update meeting", "meeting_id", meetingID, "error", err)
		return nil, model.NewAppError("channel.meeting_update_failed", http.StatusInternalServerError)
	}

	meeting.Sequence = seq

	if groupIDs != nil {
		if err := a.Store.Channels.DeleteMeetingGroups(ctx, meetingID); err != nil {
			tlog.Errorw("Failed to clear meeting groups on update", "meeting_id", meetingID, "error", err)
			return nil, model.NewAppError("meeting_group.update_failed", http.StatusInternalServerError)
		}

		if len(cleanGroups) > 0 {
			if err := a.Store.Channels.AddMeetingGroups(ctx, meetingID, cleanGroups, user.ID); err != nil {
				tlog.Errorw("Failed to add groups to meeting on update", "meeting_id", meetingID, "error", err)
				return nil, model.NewAppError("meeting_group.update_failed", http.StatusInternalServerError)
			}
		}
	}

	if err := a.syncMeetingInviteOnly(ctx, meeting); err != nil {
		tlog.Errorw("Failed to update meeting invite only", "meeting_id", meeting.ID, "error", err)
		return nil, model.NewAppError("channel.meeting_update_failed", http.StatusInternalServerError)
	}

	meeting.SetStatus()

	removed := removedInvitees(previousInvitees, meeting.Invitees)
	go a.NotifyMeetingUpdated(user, *channel, *meeting, removed)
	go a.RefreshMeetingLists(*meeting, channelID, previousAudience)

	return meeting, nil
}

func removedInvitees(before, after []string) []string {
	kept := make(map[string]bool, len(after))
	for _, id := range after {
		kept[id] = true
	}

	removed := []string{}
	for _, id := range before {
		if !kept[id] {
			removed = append(removed, id)
		}
	}

	return removed
}

// An empty newHostID auto-promotes the longest-present member, used when a host leaves without choosing one.
func (a *App) SetChannelMeetingHost(ctx context.Context, user model.User, channelID, meetingID, newHostID string) (*model.ChannelMeeting, *model.AppError) {
	if _, appErr := a.requireChannelMember(user, channelID); appErr != nil {
		return nil, appErr
	}

	meeting, appErr := a.getMeetingForChannel(ctx, channelID, meetingID)
	if appErr != nil {
		return nil, appErr
	}

	if appErr := a.requireMeetingHost(user, meeting); appErr != nil {
		return nil, appErr
	}

	newHostID = strings.TrimSpace(newHostID)
	if newHostID == "" {
		newHostID = a.longestPresentMember(ctx, meetingID, user.ID)
	}

	if newHostID == "" {
		return meeting, nil // nobody to promote, host unchanged
	}

	ok, err := a.Store.Channels.IsMember(channelID, newHostID)
	if err != nil || !ok {
		return nil, model.NewAppError("channel.meeting_host_invalid", http.StatusBadRequest)
	}

	if err := a.Store.Channels.UpdateMeetingHost(ctx, meetingID, newHostID); err != nil {
		tlog.Errorw("Failed to set meeting host", "meeting_id", meetingID, "error", err)
		return nil, model.NewAppError("channel.meeting_failed", http.StatusInternalServerError)
	}

	meeting.HostID = newHostID
	meeting.SetStatus()
	a.broadcastMeetingHostChanged(ctx, meetingID, newHostID)
	return meeting, nil
}

func (a *App) broadcastMeetingHostChanged(ctx context.Context, meetingID, newHostID string) {
	payload, err := json.Marshal(map[string]string{
		"type":    "meeting_host_changed",
		"host_id": newHostID,
	})
	if err != nil {
		tlog.Errorw("Failed to marshal meeting host change", "meeting_id", meetingID, "error", err)
		return
	}

	topic := "meeting"
	if _, err := a.LiveKit.SendData(ctx, &livekit.SendDataRequest{
		Room:  meetingID,
		Data:  payload,
		Kind:  livekit.DataPacket_RELIABLE,
		Topic: &topic,
	}); err != nil {
		tlog.Errorw("Failed to broadcast meeting host change", "meeting_id", meetingID, "error", err)
	}
}

func (a *App) endMeeting(ctx context.Context, meetingID string) {
	now := time.Now().Unix()
	if err := a.Store.Channels.EndMeeting(ctx, meetingID, now); err != nil {
		tlog.Errorw("Failed to end meeting", "meeting_id", meetingID, "error", err)
	}

	if err := a.Store.Channels.RevokeMeetingGuestLinks(ctx, meetingID, now); err != nil {
		tlog.Errorw("Failed to revoke meeting guest links", "meeting_id", meetingID, "error", err)
	}

	if _, err := a.LiveKit.DeleteRoom(ctx, &livekit.DeleteRoomRequest{Room: meetingID}); err != nil {
		tlog.Errorw("Failed to delete room", "meeting_id", meetingID, "error", err)
	}
}

func (a *App) roomParticipantCounts(ctx context.Context, meetings []model.ChannelMeeting) map[string]int {
	counts := map[string]int{}
	if len(meetings) == 0 {
		return counts
	}

	names := make([]string, 0, len(meetings))
	for _, m := range meetings {
		names = append(names, m.ID)
	}

	resp, err := a.LiveKit.ListRooms(ctx, &livekit.ListRoomsRequest{Names: names})
	if err != nil {
		tlog.Errorw("Failed to list rooms", "error", err)
		return counts
	}

	for _, r := range resp.Rooms {
		counts[r.Name] = int(r.NumParticipants)
	}

	return counts
}

func (a *App) GetMeetingVideoToken(ctx context.Context, channelID, meetingID, userID string) (string, *model.AppError) {
	user, err := a.Store.User.Get(userID)
	if err != nil {
		tlog.Errorw("Failed to retrieve user for video token", "meeting_id", meetingID, "user_id", userID, "error", err)
		return "", model.NewAppError("channel.video_token_failed", http.StatusInternalServerError)
	}

	meeting, appErr := a.getMeetingForChannel(ctx, channelID, meetingID)
	if appErr != nil {
		return "", appErr
	}
	// Access is by the meeting's own rule, not strict channel membership, so an invited non-member can join.
	if !a.canAccessMeeting(ctx, *user, meeting) {
		return "", model.NewAppError("channel.meeting_not_invited", http.StatusForbidden)
	}

	if meeting.EndedAt != nil {
		return "", model.NewAppError("channel.meeting_ended", http.StatusConflict)
	}

	if meeting.StartedAt == nil {
		return "", model.NewAppError("channel.meeting_not_started", http.StatusConflict)
	}

	return a.mintVideoToken(meeting.ID, user.ID, user.Name+" "+user.LastName, time.Hour*24)
}

func (a *App) AnnounceMeetingPresence(ctx context.Context, user model.User, channelID, meetingID string) *model.AppError {
	meeting, appErr := a.getMeetingForChannel(ctx, channelID, meetingID)
	if appErr != nil {
		return appErr
	}

	if !a.canAccessMeeting(ctx, user, meeting) {
		return model.NewAppError("channel.meeting_not_invited", http.StatusForbidden)
	}

	count, appErr := a.videoParticipantCount(meetingID)
	if appErr != nil {
		return appErr
	}

	go a.BroadcastMeetingPresence(*meeting, channelID, count)
	return nil
}

func (a *App) InviteMeetingGuests(ctx context.Context, user model.User, channelID, meetingID string, emails []string, expiresInHours int) *model.AppError {
	channel, err := a.Store.Channels.Get(channelID)
	if err != nil || channel == nil {
		return model.NewAppError("channel.not_found", http.StatusNotFound)
	}

	if channel.Type == model.ChannelTypeDirect {
		return model.NewAppError("channel.direct_invite_forbidden", http.StatusBadRequest)
	}

	meeting, appErr := a.getMeetingForChannel(ctx, channelID, meetingID)
	if appErr != nil {
		return appErr
	}
	// Only the meeting host (or a system admin) can invite.
	if appErr := a.requireMeetingHost(user, meeting); appErr != nil {
		return appErr
	}

	if meeting.EndedAt != nil {
		return model.NewAppError("channel.meeting_ended", http.StatusConflict)
	}

	now := time.Now().Unix()
	anchor := now
	if meeting.ScheduledAt != nil && *meeting.ScheduledAt > now {
		anchor = *meeting.ScheduledAt
	}

	expires := anchor + int64(guestLinkValidity(expiresInHours).Seconds())

	guestLinks := make([]model.ChannelGuestLink, 0, len(emails))
	for _, email := range emails {
		email = strings.TrimSpace(email)
		if email == "" {
			continue
		}

		guestLinks = append(guestLinks, model.ChannelGuestLink{
			ID:           model.NewID(),
			ChannelID:    channelID,
			MeetingID:    meetingID,
			CreatedBy:    user.ID,
			InvitedEmail: email,
			ExpiresAt:    expires,
			CreatedAt:    now,
		})
	}

	if len(guestLinks) == 0 {
		return model.NewAppError("channel.guest_link_no_recipients", http.StatusBadRequest)
	}

	if err := a.Store.Channels.CreateVideoGuestLink(ctx, guestLinks); err != nil {
		tlog.Errorw("Failed to create guest links", "meeting_id", meetingID, "error", err)
		return model.NewAppError("channel.guest_link_failed", http.StatusInternalServerError)
	}

	for _, link := range guestLinks {
		go a.SendVideoGuestInviteEmail(link, *meeting, user, false)
	}

	return nil
}

func (a *App) InviteMeetingMembers(ctx context.Context, user model.User, channelID, meetingID string, userIDs []string) *model.AppError {
	channel, err := a.Store.Channels.Get(channelID)
	if err != nil || channel == nil {
		return model.NewAppError("channel.not_found", http.StatusNotFound)
	}

	if channel.Type == model.ChannelTypeDirect {
		return model.NewAppError("channel.direct_invite_forbidden", http.StatusBadRequest)
	}

	meeting, appErr := a.getMeetingForChannel(ctx, channelID, meetingID)
	if appErr != nil {
		return appErr
	}

	if appErr := a.requireMeetingHost(user, meeting); appErr != nil {
		return appErr
	}

	if meeting.EndedAt != nil {
		return model.NewAppError("channel.meeting_ended", http.StatusConflict)
	}

	invitees := a.sanitizeInvitees(userIDs, meeting.HostID)
	if len(invitees) == 0 {
		return model.NewAppError("channel.guest_link_no_recipients", http.StatusBadRequest)
	}

	if err := a.Store.Channels.AddMeetingInvitees(ctx, meetingID, invitees); err != nil {
		tlog.Errorw("Failed to add meeting invitees", "meeting_id", meetingID, "error", err)
		return model.NewAppError("channel.meeting_update_failed", http.StatusInternalServerError)
	}

	meeting.Invitees = append(meeting.Invitees, invitees...)

	if err := a.syncMeetingInviteOnly(ctx, meeting); err != nil {
		tlog.Errorw("Failed to update meeting invite only", "meeting_id", meeting.ID, "error", err)
		return model.NewAppError("channel.meeting_update_failed", http.StatusInternalServerError)
	}

	go a.NotifyMeetingInvited(user, *channel, *meeting, invitees)
	go a.RefreshMeetingLists(*meeting, channelID, nil)

	return nil
}

func (a *App) ValidateVideoGuestLink(ctx context.Context, linkID string) (*model.ChannelGuestLink, *model.ChannelMeeting, *model.AppError) {
	link, err := a.Store.Channels.GetVideoGuestLink(ctx, linkID)
	if err != nil {
		tlog.Errorw("Failed to retrieve guest link", "link_id", linkID, "error", err)
		return nil, nil, model.NewAppError("channel.guest_link_failed", http.StatusInternalServerError)
	}

	if link == nil {
		return nil, nil, model.NewAppError("channel.guest_link_invalid", http.StatusNotFound)
	}

	if link.RevokedAt != nil {
		return nil, nil, model.NewAppError("channel.guest_link_invalid", http.StatusForbidden)
	}

	if time.Now().Unix() > link.ExpiresAt {
		return nil, nil, model.NewAppError("channel.guest_link_expired", http.StatusForbidden)
	}

	meeting, err := a.Store.Channels.GetMeeting(ctx, link.MeetingID)
	if err != nil {
		tlog.Errorw("Failed to retrieve meeting for guest link", "link_id", linkID, "error", err)
		return nil, nil, model.NewAppError("channel.guest_link_failed", http.StatusInternalServerError)
	}

	if meeting == nil {
		return nil, nil, model.NewAppError("channel.guest_link_invalid", http.StatusNotFound)
	}

	if meeting.EndedAt != nil {
		return nil, nil, model.NewAppError("channel.meeting_ended", http.StatusForbidden)
	}

	meeting.SetStatus()

	return link, meeting, nil
}

func (a *App) GetActiveMeetingGuestLinks(ctx context.Context, user model.User, channelID, meetingID string) ([]model.ChannelGuestLink, *model.AppError) {
	meeting, appErr := a.getMeetingForChannel(ctx, channelID, meetingID)
	if appErr != nil {
		return nil, appErr
	}

	if appErr := a.requireMeetingHost(user, meeting); appErr != nil {
		return nil, appErr
	}

	links, err := a.Store.Channels.GetActiveMeetingGuestLinks(ctx, meetingID, time.Now().Unix())
	if err != nil {
		tlog.Errorw("Failed to list guest links", "meeting_id", meetingID, "error", err)
		return nil, model.NewAppError("channel.guest_link_failed", http.StatusInternalServerError)
	}

	return links, nil
}

// Blocks future joins only, it does not disconnect a guest already in the room.
func (a *App) RevokeVideoGuestLink(ctx context.Context, user model.User, channelID, meetingID, linkID string) *model.AppError {
	meeting, appErr := a.getMeetingForChannel(ctx, channelID, meetingID)
	if appErr != nil {
		return appErr
	}

	if appErr := a.requireMeetingHost(user, meeting); appErr != nil {
		return appErr
	}

	link, err := a.Store.Channels.GetVideoGuestLink(ctx, linkID)
	if err != nil {
		tlog.Errorw("Failed to retrieve guest link for revoke", "link_id", linkID, "error", err)
		return model.NewAppError("channel.guest_link_failed", http.StatusInternalServerError)
	}

	if link == nil || link.ChannelID != channelID || link.MeetingID != meetingID {
		return model.NewAppError("channel.guest_link_invalid", http.StatusNotFound)
	}

	if err := a.Store.Channels.RevokeVideoGuestLink(ctx, linkID, time.Now().Unix()); err != nil {
		tlog.Errorw("Failed to revoke guest link", "link_id", linkID, "error", err)
		return model.NewAppError("channel.guest_link_failed", http.StatusInternalServerError)
	}

	return nil
}

func (a *App) GetGuestVideoToken(ctx context.Context, linkID, guestName, password string) (string, *model.AppError) {
	link, meeting, appErr := a.ValidateVideoGuestLink(ctx, linkID)
	if appErr != nil {
		return "", appErr
	}

	if meeting.StartedAt == nil {
		return "", model.NewAppError("channel.meeting_not_started", http.StatusConflict)
	}

	if meeting.GuestPassword != "" && !crypto.PasswordMatches(strings.TrimSpace(password), meeting.GuestPassword) {
		return "", model.NewAppError("channel.meeting_password_invalid", http.StatusForbidden)
	}

	guestName = strings.TrimSpace(guestName)
	if guestName == "" {
		return "", model.NewAppError("channel.guest_name_required", http.StatusBadRequest)
	}

	token, appErr := a.mintVideoToken(meeting.ID, "guest-"+model.NewID(), guestName, time.Until(time.Unix(link.ExpiresAt, 0)))
	if appErr != nil {
		return "", appErr
	}

	if err := a.Store.Channels.TouchVideoGuestLink(ctx, link.ID, time.Now().Unix()); err != nil {
		tlog.Errorw("Failed to mark guest link as used", "link_id", link.ID, "error", err)
	}

	return token, nil
}

// No channel-membership check, for guest-facing endpoints.
func (a *App) GetGuestVideoParticipantCount(meetingID string) (int, *model.AppError) {
	return a.videoParticipantCount(meetingID)
}

// ResolveMeetingParticipant maps a LiveKit identity to the user it belongs to,
// going through the room's own participant list so that a guest-facing endpoint
// cannot reach someone who is not in the meeting.
func (a *App) ResolveMeetingParticipant(ctx context.Context, meetingID, identity string) (string, *model.AppError) {
	if identity == "" || strings.HasPrefix(identity, "guest-") {
		return "", model.NewAppError("user.not_found", http.StatusNotFound)
	}

	present, appErr := a.LiveKit.Present(ctx, meetingID, identity)
	if appErr != nil {
		return "", appErr
	}

	if !present {
		return "", model.NewAppError("user.not_found", http.StatusNotFound)
	}

	return identity, nil
}

const deleteChannelBatch = 500

// EraseChannel runs from a job rather than a request because a busy channel can
// hold hundreds of thousands of posts. Each batch commits on its own, so a
// failure leaves the channel archived with fewer posts and re-running continues
// from there.
func (a *App) EraseChannel(ctx context.Context, channelID string) error {
	for {
		if err := ctx.Err(); err != nil {
			return err
		}

		postIDs, err := a.Store.Posts.GetBatchForChannel(ctx, channelID, deleteChannelBatch)
		if err != nil {
			return err
		}

		if len(postIDs) == 0 {
			break
		}

		// Read before deleting: these rows are the only record of which objects
		// to remove from storage.
		attachments, err := a.Store.Posts.GetAttachmentsForMany(ctx, postIDs)
		if err != nil {
			return err
		}

		if err := a.Store.Posts.DeleteBatch(ctx, postIDs); err != nil {
			return err
		}

		a.removeChatAttachmentFiles(channelID, attachments)
	}

	return a.Store.Channels.DeletePermanently(ctx, channelID)
}

func (a *App) DeleteChannel(ctx context.Context, user model.User, channelID string) *model.AppError {
	channel, err := a.Store.Channels.Get(channelID)
	if err != nil {
		tlog.Errorw("Channel lookup failed", "channel_id", channelID, "error", err)
		return model.NewAppError("channel.not_found", http.StatusNotFound)
	}

	if channel.Type != model.ChannelTypePublic {
		return model.NewAppError("channel.delete_type_invalid", http.StatusBadRequest)
	}

	if !a.SessionHasPermission(user, model.AdminPermissions.PermissionDeleteChannels) {
		return model.NewAppError("permission.forbidden", http.StatusForbidden)
	}

	// Archived first so the channel leaves every list at once, while the job
	// clears what is behind it.
	if channel.DeletedAt == 0 {
		dbCtx, cancel := a.dbCtx(ctx)
		defer cancel()

		if err := a.Store.Channels.UpdateArchived(dbCtx, channelID, time.Now().Unix()); err != nil {
			tlog.Errorw("Failed to archive channel before deletion", "channel_id", channelID, "error", err)
			return model.NewAppError("channel.delete_failed", http.StatusInternalServerError)
		}
	}

	payload, err := json.Marshal(model.DeleteChannelPayload{ChannelID: channelID})
	if err != nil {
		tlog.Errorw("Failed to marshal delete channel payload", "channel_id", channelID, "error", err)
		return model.NewAppError("channel.delete_failed", http.StatusInternalServerError)
	}

	now := time.Now().Unix()
	if _, err := a.Store.Jobs.Create(model.Job{
		ID:        model.NewID(),
		Type:      model.JobTypeDeleteChannel,
		Status:    model.JobStatusPending,
		Payload:   payload,
		CreatedAt: now,
		UpdatedAt: now,
	}); err != nil {
		tlog.Errorw("Failed to enqueue channel deletion", "channel_id", channelID, "error", err)
		return model.NewAppError("channel.delete_failed", http.StatusInternalServerError)
	}

	return nil
}
