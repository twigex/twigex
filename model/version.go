// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package model

var (
	Version        = "dev"
	BuildDate      = "unknown"
	BuildHash      = "dev"
	BuildTimestamp = "0"

	// An enterprise build keeps Version at the release it is built from and
	// numbers its own fixes as revisions, so upgrade checks never see them.
	Edition            = "community"
	EnterpriseRevision = ""
	EnterpriseHash     = ""

	// MinClientVersion should be raised via MIN_CLIENT_VERSION.
	MinClientVersion = "0.0.0"

	// MinUpgradableVersion should be raised via MIN_UPGRADABLE_VERSION.
	MinUpgradableVersion = "0.0.0"
)

// LastUngatedVersion is what an installation carrying no recorded version is
// taken to be running, 0.14.0 being the release that starts recording one. A
// fresh install that dies between migrating and recording leaves one too, so
// this outlives 0.14.0. Upgrade stops must stay above it.
const LastUngatedVersion = "0.13.0"
