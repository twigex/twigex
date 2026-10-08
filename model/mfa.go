// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package model

type Mfa struct {
	Secret string `json:"secret"`
	QRCode string `json:"qr_code"`
}
