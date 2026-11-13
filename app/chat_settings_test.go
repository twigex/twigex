// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"net/http"
	"testing"

	"github.com/twigex/twigex/config"
	"github.com/twigex/twigex/internal/klipy"
	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/store"
)

type fakeChatSettingsStore struct {
	store.SystemSettingsStore
	saved map[string]string
}

func (f *fakeChatSettingsStore) UpdateBatch(settings map[string]string) error {
	f.saved = settings

	return nil
}

func chatSettingsApp(locked config.EnvLockedSettings) (*App, *fakeChatSettingsStore) {
	settingsStore := &fakeChatSettingsStore{}

	a := &App{
		Store: store.Store{SystemSettings: settingsStore},
		GIF:   klipy.New("saved-gif-key"),
		ConfigStore: config.ConfigStore{
			EnvLocked: locked,
			Config: &model.ServerConfig{
				ServerSettings: model.ServerSettings{AtRestEncryptKey: model.NewString("0123456789abcdef0123456789abcdef")},
				ChannelSettings: model.ChannelSettings{
					Enabled:      model.NewBool(true),
					Host:         model.NewString("wss://video.example"),
					Key:          model.NewString("video-key"),
					Secret:       model.NewString("video-secret"),
					LinkPreviews: model.NewBool(false),
					KlipyGIF: model.KlipyGIFSettings{
						Enabled: model.NewBool(false),
						APIKey:  model.NewString("saved-gif-key"),
					},
				},
			},
		},
	}

	return a, settingsStore
}

func chatSettingsForm() model.ChatSettings {
	settings := model.ChatSettings{}
	settings.VideoSettings.Enabled = false
	settings.VideoSettings.Host = "wss://changed.example"
	settings.GIFSettings.Enabled = true
	settings.LinkPreviewSettings.Enabled = true

	return settings
}

func TestChatSettingsSaveTheGroupsNotSetByEnvironment(t *testing.T) {
	a, settingsStore := chatSettingsApp(config.EnvLockedSettings{Video: true})

	appErr := a.UpdateChatSettings(model.User{Role: model.SystemAdminRoleId}, chatSettingsForm())
	if appErr != nil {
		t.Fatalf("want the save to work with only video locked, got %v", appErr)
	}

	if _, saved := settingsStore.saved["channel.host"]; saved {
		t.Error("want the video settings left alone, they come from the environment")
	}

	if settingsStore.saved["channel.gif_enabled"] != "true" || settingsStore.saved["channel.link_previews"] != "true" {
		t.Errorf("want GIF and link previews saved, got %v", settingsStore.saved)
	}

	if *a.ConfigStore.Config.ChannelSettings.Host != "wss://video.example" {
		t.Error("want the video host in memory unchanged")
	}
}

func TestChatSettingsKeepTheGIFKeyWhenNotRetyped(t *testing.T) {
	a, _ := chatSettingsApp(config.EnvLockedSettings{})

	appErr := a.UpdateChatSettings(model.User{Role: model.SystemAdminRoleId}, chatSettingsForm())
	if appErr != nil {
		t.Fatalf("UpdateChatSettings: %v", appErr)
	}

	if a.GIF.APIKey != "saved-gif-key" {
		t.Errorf("want the saved GIF key kept, got %q", a.GIF.APIKey)
	}
}

func TestChatSettingsAllSetByEnvironmentIsRefused(t *testing.T) {
	a, _ := chatSettingsApp(config.EnvLockedSettings{Video: true, GIF: true, LinkPreview: true})

	appErr := a.UpdateChatSettings(model.User{Role: model.SystemAdminRoleId}, chatSettingsForm())
	if appErr == nil || appErr.Status != http.StatusForbidden {
		t.Fatalf("want 403, got %v", appErr)
	}
}
