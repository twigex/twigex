// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package model

const (
	PreferencesCategoryDisplay              = "display_settings"
	PreferencesCategoryDefaultNotifications = "default_notifications"
	PreferencesCategoryNotifications        = "notifications"
)

type Preference struct {
	UserID   string `json:"user_id"`
	Category string `json:"category"`
	Name     string `json:"name"`
	Value    string `json:"value"`
	Type     string `json:"type"`
}

type Preferences struct {
	DisplaySettings []Preference
	Notifications   []Preference
}
