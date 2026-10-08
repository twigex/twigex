// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package model

type WebsocketEvent struct {
	Event string         `json:"event"`
	App   string         `json:"app"`
	Data  map[string]any `json:"data"` // parlikt uz map[string]interface{}, jo posts nav NotificationMessage?
}
