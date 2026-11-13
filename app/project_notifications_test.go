// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/store"
)

type fakeNotifyWorkspaceStore struct {
	store.WorkspaceStore
	members []model.WorkspaceMember
}

func (f fakeNotifyWorkspaceStore) GetMembers(string) ([]model.WorkspaceMember, error) {
	return f.members, nil
}

type fakeNotifyUserStore struct {
	store.UserStore
}

func (f fakeNotifyUserStore) GetByIDs(ids []string) ([]model.User, error) {
	users := []model.User{}

	for _, id := range ids {
		users = append(users, model.User{ID: id, Role: model.SystemAdminRoleId})
	}

	return users, nil
}

func (f fakeNotifyUserStore) Get(string) (*model.User, error) {
	return nil, errors.New("not needed in this test")
}

type fakeNotifyStore struct {
	store.NotificationStore
	mu    sync.Mutex
	saved []string
}

func (f *fakeNotifyStore) Create(_ model.User, receiver string, _ string, _ string, _ string, props map[string]interface{}) (*model.NotificationMessage, error) {
	encoded, err := json.Marshal(props)
	if err != nil {
		return nil, err
	}

	f.mu.Lock()
	f.saved = append(f.saved, string(encoded))
	f.mu.Unlock()

	message := &model.NotificationMessage{ID: "n-" + receiver, Receiver: receiver, Details: props}

	return message, nil
}

type fakeNotifyPreferenceStore struct {
	store.PreferencesStore
}

func (f fakeNotifyPreferenceStore) GetNotifications(string, string) ([]model.Preference, error) {
	return nil, nil
}

func TestProjectNotificationsDoNotShareOneMessage(t *testing.T) {
	members := []model.WorkspaceMember{}

	for index := range 20 {
		members = append(members, model.WorkspaceMember{UserID: fmt.Sprintf("member-%d", index)})
	}

	notifications := &fakeNotifyStore{}

	a := &App{
		Server: Server{NotificationHub: newHub()},
		Store: store.Store{
			Workspace:     fakeNotifyWorkspaceStore{members: members},
			User:          fakeNotifyUserStore{},
			Notifications: notifications,
			Preferences:   fakeNotifyPreferenceStore{},
		},
	}

	creator := model.User{ID: "creator"}
	props := map[string]any{"taskName": "Write the report"}

	a.createProjectNotification("workspace-1", creator, model.NOTIFICATION_PROJECT_TASK_CREATED, props)

	time.Sleep(200 * time.Millisecond)

	notifications.mu.Lock()
	defer notifications.mu.Unlock()

	if len(notifications.saved) != len(members) {
		t.Fatalf("want %d notifications, got %d", len(members), len(notifications.saved))
	}

	for _, saved := range notifications.saved {
		if strings.Contains(saved, `"message"`) {
			t.Fatalf("want each notification saved without another member's message, got %s", saved)
		}
	}
}
