// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package model

var (
	StatusOnline  = "online"
	StatusOffline = "offline"
	StatusAway    = "away"
	StatusDnD     = "dnd"
)

type ChatSettings struct {
	VideoSettings        VideoSettings       `json:"video_settings"`
	GIFSettings          GIFSettings         `json:"gif_settings"`
	LinkPreviewSettings  LinkPreviewSettings `json:"link_preview_settings"`
	VideoEnvLocked       bool                `json:"video_env_locked"`
	GIFEnvLocked         bool                `json:"gif_env_locked"`
	LinkPreviewEnvLocked bool                `json:"link_preview_env_locked"`
}

type VideoSettings struct {
	Enabled bool   `json:"enabled"`
	Host    string `json:"host"`
	Key     string `json:"key"`
	Secret  string `json:"secret"`
}

type GIFSettings struct {
	Enabled bool   `json:"enabled"`
	APIKey  string `json:"api_key"`
}

type LinkPreviewSettings struct {
	Enabled bool `json:"enabled"`
}
