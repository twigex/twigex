// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"context"
	"net/http"
	"testing"

	"github.com/twigex/twigex/config"
	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/store"
)

type fakeArchiveChannelStore struct {
	store.ChannelStore
	channel    model.Channel
	members    []model.ChannelMember
	notAMember bool
	archivedTo int64
	archiveSet bool
}

func (f *fakeArchiveChannelStore) Get(string) (*model.Channel, error) {
	c := f.channel
	c.ChannelMembers = f.members
	return &c, nil
}

func (f *fakeArchiveChannelStore) GetMembers(string) ([]model.ChannelMember, error) {
	return f.members, nil
}

func (f *fakeArchiveChannelStore) IsMember(string, string) (bool, error) {
	return !f.notAMember, nil
}

func (f *fakeArchiveChannelStore) UpdateUserRole(_, _, _ string) error {
	return nil
}

func (f *fakeArchiveChannelStore) UpdateArchived(_ context.Context, _ string, archivedAt int64) error {
	f.archivedTo = archivedAt
	f.archiveSet = true
	return nil
}

type fakeArchiveRoleStore struct {
	store.RoleStore
	permissions []string
}

func (f fakeArchiveRoleStore) GetByNames([]string) ([]model.Role, error) {
	roles := []model.Role{{Name: model.SystemUserRoleId, Permissions: f.permissions}}
	return roles, nil
}

func archiveApp(channel model.Channel, members []model.ChannelMember, perms ...string) (*App, *fakeArchiveChannelStore) {
	s := &fakeArchiveChannelStore{channel: channel, members: members}
	return &App{
		Store: store.Store{
			Channels: s,
			Roles:    fakeArchiveRoleStore{permissions: perms},
		},
		ConfigStore: config.ConfigStore{
			Config: &model.ServerConfig{
				SqlSettings: model.SqlSettings{QueryTimeout: model.NewInt(25)},
			},
		},
	}, s
}

const (
	archiveChannelID = "channel-1"
	channelAdminID   = "user-admin"
	plainMemberID    = "user-plain"
)

func openChannel() model.Channel {
	return model.Channel{ID: archiveChannelID, Type: model.ChannelTypePublic}
}

func adminMembers() []model.ChannelMember {
	return []model.ChannelMember{
		{UserID: channelAdminID, Role: model.ChannelRoleAdmin},
		{UserID: plainMemberID},
	}
}

func userWith(id string) model.User {
	return model.User{ID: id, Role: model.SystemUserRoleId}
}

func TestArchiveChannelRequiresChannelAdmin(t *testing.T) {
	a, s := archiveApp(openChannel(), adminMembers())

	appErr := a.ArchiveChannel(t.Context(), userWith(plainMemberID), archiveChannelID)
	if appErr == nil || appErr.Status != http.StatusForbidden {
		t.Fatalf("a plain member must not archive, got %v", appErr)
	}
	if s.archiveSet {
		t.Error("the store was written to despite the refusal")
	}
}

func TestArchiveChannelRejectsDirectMessages(t *testing.T) {
	dm := openChannel()
	dm.Type = model.ChannelTypeDirect

	a, s := archiveApp(dm, adminMembers())

	appErr := a.ArchiveChannel(t.Context(), model.User{ID: channelAdminID, Role: model.SystemAdminRoleId}, archiveChannelID)
	if appErr == nil || appErr.Status != http.StatusBadRequest {
		t.Fatalf("a direct message must not be archivable, got %v", appErr)
	}
	if s.archiveSet {
		t.Error("the store was written to despite the refusal")
	}
}

func TestArchiveChannelRefusesWhenAlreadyArchived(t *testing.T) {
	archived := openChannel()
	archived.DeletedAt = 1700000000

	a, _ := archiveApp(archived, adminMembers())

	appErr := a.ArchiveChannel(t.Context(), model.User{ID: channelAdminID, Role: model.SystemAdminRoleId}, archiveChannelID)
	if appErr == nil || appErr.Status != http.StatusConflict {
		t.Fatalf("an archived channel must refuse a second archive, got %v", appErr)
	}
}

func TestArchiveChannelPermissionWorksWithoutMembership(t *testing.T) {
	a, s := archiveApp(openChannel(), []model.ChannelMember{},
		model.AdminPermissions.PermissionArchiveChannels.Id)

	if appErr := a.ArchiveChannel(t.Context(), userWith("manager"), archiveChannelID); appErr != nil {
		t.Fatalf("the permission must archive a channel the user is not in, got %v", appErr)
	}
	if !s.archiveSet {
		t.Error("the channel was not marked archived")
	}
}

func TestSystemAdminCanArchiveWithoutMembership(t *testing.T) {
	a, s := archiveApp(openChannel(), []model.ChannelMember{})

	if appErr := a.ArchiveChannel(t.Context(), model.User{ID: "root", Role: model.SystemAdminRoleId}, archiveChannelID); appErr != nil {
		t.Fatalf("a system admin must be able to archive any channel, got %v", appErr)
	}
	if !s.archiveSet || s.archivedTo == 0 {
		t.Error("the channel was not marked archived")
	}
}

func TestUnarchiveClearsTheTimestamp(t *testing.T) {
	archived := openChannel()
	archived.DeletedAt = 1700000000

	a, s := archiveApp(archived, adminMembers())

	if appErr := a.UnarchiveChannel(t.Context(), model.User{ID: "root", Role: model.SystemAdminRoleId}, archiveChannelID); appErr != nil {
		t.Fatalf("unarchive failed: %v", appErr)
	}
	if !s.archiveSet || s.archivedTo != 0 {
		t.Errorf("deleted_at should be cleared to 0, got %d", s.archivedTo)
	}
}

// A write path that forgets the guard fails here.
func TestArchivedChannelRefusesWrites(t *testing.T) {
	archived := openChannel()
	archived.DeletedAt = 1700000000

	a, _ := archiveApp(archived, adminMembers())
	member := userWith(plainMemberID)

	writes := map[string]func() *model.AppError{
		"CreatePost": func() *model.AppError {
			_, appErr := a.CreatePost(t.Context(), member, archiveChannelID, "hello", "", nil, "", "")
			return appErr
		},
		"CreatePostReaction": func() *model.AppError {
			_, appErr := a.CreatePostReaction(t.Context(), member, archiveChannelID, "post-1", ":+1:")
			return appErr
		},
		"DeletePostReaction": func() *model.AppError {
			return a.DeletePostReaction(t.Context(), member, archiveChannelID, "post-1", ":+1:")
		},
		"DeletePost": func() *model.AppError {
			return a.DeletePost(t.Context(), member, archiveChannelID, "post-1")
		},
	}

	for name, write := range writes {
		t.Run(name, func(t *testing.T) {
			appErr := write()
			if appErr == nil || appErr.Status != http.StatusConflict {
				t.Errorf("must refuse on an archived channel, got %v", appErr)
			}
		})
	}
}

func TestChangeUserRoleRefusesAPlainMember(t *testing.T) {
	a, _ := archiveApp(openChannel(), adminMembers())

	appErr := a.ChangeUserRole(userWith(plainMemberID), archiveChannelID, channelAdminID, model.ChannelRoleMember)
	if appErr == nil || appErr.Status != http.StatusForbidden {
		t.Fatalf("a plain member must not change roles, got %v", appErr)
	}
}

func TestChangeUserRoleRejectsAnUnknownRole(t *testing.T) {
	a, _ := archiveApp(openChannel(), adminMembers())

	appErr := a.ChangeUserRole(userWith(channelAdminID), archiveChannelID, plainMemberID, "superuser")
	if appErr == nil || appErr.Status != http.StatusBadRequest {
		t.Fatalf("an unknown role must be refused, got %v", appErr)
	}
}

func TestUpdateChannelRefusesAPlainMember(t *testing.T) {
	a, _ := archiveApp(openChannel(), adminMembers())

	_, appErr := a.UpdateChannel(userWith(plainMemberID), archiveChannelID, model.ChannelPatch{})
	if appErr == nil || appErr.Status != http.StatusForbidden {
		t.Fatalf("a plain member must not rename a channel, got %v", appErr)
	}
}

func TestUpdateChannelRejectsDirectMessages(t *testing.T) {
	dm := openChannel()
	dm.Type = model.ChannelTypeDirect

	a, _ := archiveApp(dm, adminMembers())

	_, appErr := a.UpdateChannel(userWith(channelAdminID), archiveChannelID, model.ChannelPatch{})
	if appErr == nil || appErr.Status != http.StatusBadRequest {
		t.Fatalf("a direct message must not be renamable, got %v", appErr)
	}
}

func TestChangeUserRoleRefusesToPromoteAGroupMember(t *testing.T) {
	members := []model.ChannelMember{
		{UserID: channelAdminID, Role: model.ChannelRoleAdmin, IsDirect: true},
		{UserID: plainMemberID, Role: model.ChannelRoleMember},
	}
	a, _ := archiveApp(openChannel(), members)

	appErr := a.ChangeUserRole(userWith(channelAdminID), archiveChannelID, plainMemberID, model.ChannelRoleAdmin)
	if appErr == nil || appErr.Status != http.StatusBadRequest {
		t.Fatalf("a group-materialized member must not be promoted, got %v", appErr)
	}
}

func TestChangeUserRolePromotesADirectMember(t *testing.T) {
	members := []model.ChannelMember{
		{UserID: channelAdminID, Role: model.ChannelRoleAdmin, IsDirect: true},
		{UserID: plainMemberID, Role: model.ChannelRoleMember, IsDirect: true},
	}
	a, _ := archiveApp(openChannel(), members)

	if appErr := a.ChangeUserRole(userWith(channelAdminID), archiveChannelID, plainMemberID, model.ChannelRoleAdmin); appErr != nil {
		t.Fatalf("a direct member must be promotable, got %v", appErr)
	}
}
