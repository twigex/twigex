// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package model

import "time"

type LDAPDirectoryUser struct {
	ExternalID string
	Username   string
	Email      string
	FirstName  string
	LastName   string
}

type LDAPSyncSettings struct {
	Enabled              bool
	Interval             time.Duration
	MaxDeactivatePercent int
}
