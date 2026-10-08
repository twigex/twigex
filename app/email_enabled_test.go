// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"testing"

	"github.com/twigex/twigex/config"
	"github.com/twigex/twigex/model"
)

func emailApp(enable *bool, server string) *App {
	return &App{
		ConfigStore: config.ConfigStore{Config: &model.ServerConfig{
			EmailSettings: model.EmailSettings{
				EnableEmail: enable,
				SMTPServer:  model.NewString(server),
			},
		}},
	}
}

func TestEmailEnabled(t *testing.T) {
	on, off := true, false

	cases := []struct {
		name   string
		enable *bool
		server string
		want   bool
	}{
		{"never chosen, no server", nil, "", false},
		{"never chosen, server set", nil, "smtp.example.com", true},
		{"switched off, server set", &off, "smtp.example.com", false},
		{"switched on, no server", &on, "", true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := emailApp(tc.enable, tc.server).emailEnabled(); got != tc.want {
				t.Fatalf("want %v, got %v", tc.want, got)
			}
		})
	}
}

func TestPublicAuthSettingsHidePasswordResetWhenEmailIsOff(t *testing.T) {
	for _, enabled := range []bool{true, false} {
		cfg := &model.ServerConfig{}
		cfg.SetDefaults()
		cfg.EmailSettings.EnableEmail = &enabled

		a := &App{ConfigStore: config.ConfigStore{Config: cfg}}

		got, appErr := a.GetPublicAuthSettings()
		if appErr != nil {
			t.Fatalf("GetPublicAuthSettings: %v", appErr)
		}

		if got.PasswordResetEnabled != enabled {
			t.Errorf("email enabled %v: want passwordResetEnabled %v, got %v", enabled, enabled, got.PasswordResetEnabled)
		}
	}
}

func TestGenerateEmailSkipsWhenEmailIsOff(t *testing.T) {
	off := false
	a := emailApp(&off, "smtp.example.com")

	if err := a.generateEmail(model.NotificationMessage{Receiver: "u1"}); err != nil {
		t.Fatalf("expected a silent skip, got %v", err)
	}
}
