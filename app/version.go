// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"strconv"
	"time"

	"github.com/twigex/twigex/internal/upgrade"
	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/tlog"
)

// RecordVersion must run after DoAppMigrations, so that a crash part way
// through an upgrade leaves the previous version in place and the next start
// retries.
func (a *App) RecordVersion() {
	if !upgrade.IsReal(model.Version) {
		return
	}

	recorded, err := a.Store.SystemSettings.Get(upgrade.VersionKey)
	if err != nil {
		tlog.Errorw("Failed to read the recorded app version", "error", err)
		return
	}

	if recorded == model.Version {
		return
	}

	err = a.Store.SystemSettings.UpdateBatch(map[string]string{
		upgrade.VersionKey:       model.Version,
		upgrade.MinUpgradableKey: model.MinUpgradableVersion,
		upgrade.UpgradedAtKey:    strconv.FormatInt(time.Now().Unix(), 10),
	})
	if err != nil {
		tlog.Errorw("Failed to record the app version", "error", err)
		return
	}

	tlog.Infow("Recorded app version", "from", recorded, "to", model.Version)
}
