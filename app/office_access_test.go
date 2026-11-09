// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"context"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/twigex/twigex/config"
	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/store"
)

const officeTestSecret = "office-test-secret"

type fakeOfficeFileStore struct {
	store.FileStore
	files  map[string]*model.File
	shares map[string][]model.File
}

func (f *fakeOfficeFileStore) Get(id string) (*model.File, error) {
	file, ok := f.files[id]
	if !ok {
		return nil, nil
	}

	copied := *file
	return &copied, nil
}

func (f *fakeOfficeFileStore) ResolveEffectiveShare(_ context.Context, userID, fileID string) ([]model.File, error) {
	return f.shares[userID+"/"+fileID], nil
}

type fakeOfficePostStore struct {
	store.PostStore
	attachments map[string]*model.PostFileAttachment
}

func (f *fakeOfficePostStore) GetAttachmentByID(id string) (*model.PostFileAttachment, error) {
	return f.attachments[id], nil
}

type fakeOfficeChannelStore struct {
	store.ChannelStore
	members map[string]bool
}

func (f *fakeOfficeChannelStore) IsMember(channelID, userID string) (bool, error) {
	return f.members[channelID+"/"+userID], nil
}

type fakeOfficeWorkspaceStore struct {
	store.WorkspaceStore
	attachments map[string]*model.WorkspaceAttachment
	members     map[string]bool
}

func (f *fakeOfficeWorkspaceStore) GetUserAttachment(id string) (*model.WorkspaceAttachment, error) {
	return f.attachments[id], nil
}

func (f *fakeOfficeWorkspaceStore) IsMember(workspaceID, userID string) (bool, error) {
	return f.members[workspaceID+"/"+userID], nil
}

type officeTestStores struct {
	files       *fakeOfficeFileStore
	attachments *fakeOfficePostStore
	channels    *fakeOfficeChannelStore
	workspaces  *fakeOfficeWorkspaceStore
}

func newOfficeAccessApp() (*App, officeTestStores) {
	serverConfig := model.ServerConfig{}
	serverConfig.SetDefaults()
	serverConfig.OfficeSettings.Secret = model.NewString(officeTestSecret)

	stores := officeTestStores{
		files: &fakeOfficeFileStore{
			files: map[string]*model.File{
				"doc-1": {ID: "doc-1", Owner: "owner-1"},
				"doc-2": {ID: "doc-2", Owner: "owner-1"},
			},
			shares: map[string][]model.File{},
		},
		attachments: &fakeOfficePostStore{
			attachments: map[string]*model.PostFileAttachment{
				"att-1": {ID: "att-1", ChannelID: "chan-1", Name: "notes.docx"},
				"att-2": {ID: "att-2", ChannelID: "chan-1", Name: "plan.docx"},
			},
		},
		channels: &fakeOfficeChannelStore{
			members: map[string]bool{"chan-1/user-1": true},
		},
		workspaces: &fakeOfficeWorkspaceStore{
			attachments: map[string]*model.WorkspaceAttachment{
				"task-att-1": {ID: "task-att-1", WorkspaceID: "ws-1", Name: "brief.docx"},
			},
			members: map[string]bool{"ws-1/user-1": true},
		},
	}

	a := &App{
		ConfigStore: config.ConfigStore{Config: &serverConfig},
		Store: store.Store{
			File:      stores.files,
			Posts:     stores.attachments,
			Channels:  stores.channels,
			Workspace: stores.workspaces,
			User:      &fakePublicUserStore{user: &model.User{ID: "user-1"}},
		},
	}

	return a, stores
}

func signOfficeTestToken(t *testing.T, claims jwt.MapClaims) string {
	t.Helper()

	claims["exp"] = time.Now().Add(time.Hour).Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signed, err := token.SignedString([]byte(officeTestSecret))
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	return signed
}

func userOfficeToken(t *testing.T, wopiApp, fileID string) string {
	t.Helper()

	return signOfficeTestToken(t, jwt.MapClaims{"id": "user-1", "app": wopiApp, "file": fileID})
}

func TestServeOfficeChatAttachment_RejectsMissingToken(t *testing.T) {
	a, _ := newOfficeAccessApp()

	appErr := a.ServeOfficeChatAttachment("att-1", "", io.Discard)
	requireAppErr(t, appErr, "office.token_invalid", http.StatusUnauthorized)
}

func TestServeOfficeChatAttachment_RejectsTokenWithoutDocumentBinding(t *testing.T) {
	a, _ := newOfficeAccessApp()
	token := signOfficeTestToken(t, jwt.MapClaims{"id": "user-1"})

	appErr := a.ServeOfficeChatAttachment("att-1", token, io.Discard)
	requireAppErr(t, appErr, "office.token_invalid", http.StatusForbidden)
}

func TestServeOfficeChatAttachment_RejectsTokenForAnotherAttachment(t *testing.T) {
	a, _ := newOfficeAccessApp()
	token := userOfficeToken(t, "chat", "att-1")

	appErr := a.ServeOfficeChatAttachment("att-2", token, io.Discard)
	requireAppErr(t, appErr, "office.token_invalid", http.StatusForbidden)
}

func TestServeOfficeChatAttachment_RejectsFilesTokenOnChatRoute(t *testing.T) {
	a, _ := newOfficeAccessApp()
	token := userOfficeToken(t, "files", "att-1")

	appErr := a.ServeOfficeChatAttachment("att-1", token, io.Discard)
	requireAppErr(t, appErr, "office.token_invalid", http.StatusForbidden)
}

func TestServeOfficeChatAttachment_RejectsRemovedMemberWithValidToken(t *testing.T) {
	a, stores := newOfficeAccessApp()
	token := userOfficeToken(t, "chat", "att-1")

	delete(stores.channels.members, "chan-1/user-1")

	appErr := a.ServeOfficeChatAttachment("att-1", token, io.Discard)
	requireAppErr(t, appErr, "channel.forbidden", http.StatusForbidden)
}

func TestServeOfficeChatAttachment_RejectsGuestShareToken(t *testing.T) {
	a, _ := newOfficeAccessApp()
	token := signOfficeTestToken(t, jwt.MapClaims{"share": "tok", "file": "att-1"})

	appErr := a.ServeOfficeChatAttachment("att-1", token, io.Discard)
	requireAppErr(t, appErr, "office.token_invalid", http.StatusForbidden)
}

func TestOfficeChatAttachment_AllowsChannelMember(t *testing.T) {
	a, _ := newOfficeAccessApp()

	attachment, appErr := a.officeChatAttachment("att-1", model.User{ID: "user-1"})
	if appErr != nil {
		t.Fatalf("expected member access, got %v", appErr)
	}

	if attachment.ID != "att-1" {
		t.Fatalf("expected att-1, got %q", attachment.ID)
	}
}

func TestOpenOfficeFile_RejectsChatAttachmentForNonMember(t *testing.T) {
	a, _ := newOfficeAccessApp()

	_, appErr := a.OpenOfficeFile("att-1", model.User{ID: "outsider"}, "chat")
	requireAppErr(t, appErr, "channel.forbidden", http.StatusForbidden)
}

func TestCheckFileInfo_RejectsRemovedChatMember(t *testing.T) {
	a, stores := newOfficeAccessApp()
	token := userOfficeToken(t, "chat", "att-1")

	delete(stores.channels.members, "chan-1/user-1")

	_, appErr := a.CheckFileInfo("att-1", token, "chat")
	requireAppErr(t, appErr, "channel.forbidden", http.StatusForbidden)
}

func TestGetOfficeFile_RejectsTokenForAnotherFile(t *testing.T) {
	a, _ := newOfficeAccessApp()
	token := userOfficeToken(t, "files", "doc-1")

	_, appErr := a.GetOfficeFile("doc-2", token)
	requireAppErr(t, appErr, "office.token_invalid", http.StatusForbidden)
}

func TestGetOfficeFile_RejectsUserWithoutShare(t *testing.T) {
	a, _ := newOfficeAccessApp()
	token := userOfficeToken(t, "files", "doc-1")

	_, appErr := a.GetOfficeFile("doc-1", token)
	requireAppErr(t, appErr, "file.forbidden", http.StatusForbidden)
}

func TestGetOfficeFile_AllowsSharedViewer(t *testing.T) {
	a, stores := newOfficeAccessApp()
	token := userOfficeToken(t, "files", "doc-1")

	stores.files.shares["user-1/doc-1"] = []model.File{{AccessLevel: model.AccessViewer}}

	file, appErr := a.GetOfficeFile("doc-1", token)
	if appErr != nil {
		t.Fatalf("expected viewer access, got %v", appErr)
	}

	if file.ID != "doc-1" {
		t.Fatalf("expected doc-1, got %q", file.ID)
	}
}

func TestOfficeSaveFile_RejectsViewer(t *testing.T) {
	a, stores := newOfficeAccessApp()
	token := userOfficeToken(t, "files", "doc-1")

	stores.files.shares["user-1/doc-1"] = []model.File{{AccessLevel: model.AccessViewer}}

	appErr := a.OfficeSaveFile("doc-1", token, io.NopCloser(nil))
	requireAppErr(t, appErr, "office.readonly", http.StatusForbidden)
}

func TestOfficeSaveFile_RejectsUserWithoutShare(t *testing.T) {
	a, _ := newOfficeAccessApp()
	token := userOfficeToken(t, "files", "doc-1")

	appErr := a.OfficeSaveFile("doc-1", token, io.NopCloser(nil))
	requireAppErr(t, appErr, "file.forbidden", http.StatusForbidden)
}

func TestOfficeEditableFile_AllowsEditor(t *testing.T) {
	a, stores := newOfficeAccessApp()
	token, appErr := a.VerifyOfficeToken(userOfficeToken(t, "files", "doc-1"))
	if appErr != nil {
		t.Fatalf("failed to verify token: %v", appErr)
	}

	stores.files.shares["user-1/doc-1"] = []model.File{{AccessLevel: model.AccessEditor}}

	file, appErr := a.officeEditableFile(token, "doc-1")
	if appErr != nil {
		t.Fatalf("expected editor access, got %v", appErr)
	}

	if file.ID != "doc-1" {
		t.Fatalf("expected doc-1, got %q", file.ID)
	}
}

func TestServeOfficeWorkspaceAttachment_RejectsMissingToken(t *testing.T) {
	a, _ := newOfficeAccessApp()

	appErr := a.ServeOfficeWorkspaceAttachment("task-att-1", "", io.Discard)
	requireAppErr(t, appErr, "office.token_invalid", http.StatusUnauthorized)
}

func TestServeOfficeWorkspaceAttachment_RejectsChatTokenOnProjectsRoute(t *testing.T) {
	a, _ := newOfficeAccessApp()
	token := userOfficeToken(t, "chat", "task-att-1")

	appErr := a.ServeOfficeWorkspaceAttachment("task-att-1", token, io.Discard)
	requireAppErr(t, appErr, "office.token_invalid", http.StatusForbidden)
}

func TestServeOfficeWorkspaceAttachment_RejectsRemovedMemberWithValidToken(t *testing.T) {
	a, stores := newOfficeAccessApp()
	token := userOfficeToken(t, "projects", "task-att-1")

	delete(stores.workspaces.members, "ws-1/user-1")

	appErr := a.ServeOfficeWorkspaceAttachment("task-att-1", token, io.Discard)
	requireAppErr(t, appErr, "workspace.forbidden", http.StatusForbidden)
}
